(use spork/test)
(import spork/json)

(start-suite "mcp-integration")

# --- Helpers ---

(def project-root
  (do
    (def here (os/cwd))
    # Ensure we resolve to the janet project root
    (if (string/has-suffix? "/janet" here)
      here
      (string here "/janet"))))

(defn make-tmp-workspace []
  "Create a temp workspace dir with .glitch/ subdir. Returns the path."
  (def dir (string "/tmp/glitch-mcp-int-" (os/time) "-" (math/rng-int (math/rng))))
  (os/mkdir dir)
  (os/mkdir (string dir "/.glitch"))
  dir)

(defn cleanup [dir]
  "Remove a temp directory tree."
  (os/shell (string "rm -rf " dir)))

(defn spawn-mcp [workspace-path]
  "Spawn the MCP server subprocess. Returns the process handle."
  (def modpath (string project-root "/src/:all:.janet"))
  (def expr
    (string/format
      `(array/push module/paths [%q :source]) (import glitch/mcp :as mcp) (mcp/start @{:workspace-path %q})`
      modpath workspace-path))
  (os/spawn ["janet" "-e" expr]
    :p {:in :pipe :out :pipe :err :pipe}))

(defn read-line [stream]
  "Read a complete JSON-RPC response line from stream. Returns string or nil."
  (def buf @"")
  (var result nil)
  (while true
    (def chunk (ev/read stream 4096))
    (when (or (nil? chunk) (= (length chunk) 0))
      (when (> (length buf) 0)
        (set result (string buf)))
      (break))
    (buffer/push buf chunk)
    (when (string/find "\n" (string buf))
      (set result (string buf))
      (break)))
  result)

(defn send-recv [proc msg]
  "Send a JSON-RPC message and read the response line."
  (def line (string (json/encode msg) "\n"))
  (ev/write (proc :in) line)
  (def response (read-line (proc :out)))
  (when response (json/decode (string/trim response))))

(defn send-no-reply [proc msg]
  "Send a JSON-RPC notification (no response expected)."
  (def line (string (json/encode msg) "\n"))
  (ev/write (proc :in) line))

(defn do-initialize [proc]
  "Send the initialize handshake and return the response."
  (def resp (send-recv proc
    @{"jsonrpc" "2.0" "id" 1
      "method" "initialize"
      "params" @{"capabilities" @{}}}))
  (send-no-reply proc
    @{"jsonrpc" "2.0"
      "method" "notifications/initialized"})
  resp)

(defn stop-proc [proc]
  "Close stdin and wait for process to exit."
  (ev/close (proc :in))
  (os/proc-wait proc))

# --- Test 1: Initialize handshake ---

(do
  (def ws (make-tmp-workspace))
  (def proc (spawn-mcp ws))

  (def resp (send-recv proc
    @{"jsonrpc" "2.0" "id" 1
      "method" "initialize"
      "params" @{"capabilities" @{}}}))

  (assert resp "initialize: got a response")
  (assert (get resp "result") "initialize: response has result")
  (def server-info (get-in resp ["result" "serverInfo"]))
  (assert server-info "initialize: result has serverInfo")
  (assert (= (get server-info "name") "glitch")
          "initialize: serverInfo.name is glitch")

  # Send initialized notification (no response)
  (send-no-reply proc
    @{"jsonrpc" "2.0"
      "method" "notifications/initialized"})

  (stop-proc proc)
  (cleanup ws))

# --- Test 2: tools/list returns 8 tools ---

(do
  (def ws (make-tmp-workspace))
  (def proc (spawn-mcp ws))
  (do-initialize proc)

  (def resp (send-recv proc
    @{"jsonrpc" "2.0" "id" 2
      "method" "tools/list"
      "params" @{}}))

  (assert resp "tools/list: got a response")
  (def tools (get-in resp ["result" "tools"]))
  (assert tools "tools/list: result has tools")
  (assert (= (length tools) 8)
          (string "tools/list: expected 8 tools, got " (length tools)))

  # Check specific tool names
  (def names (map |(get $ "name") tools))
  (assert (find |(= $ "glitch_search") names)
          "tools/list: includes glitch_search")
  (assert (find |(= $ "glitch_read_file") names)
          "tools/list: includes glitch_read_file")
  (assert (find |(= $ "glitch_grep") names)
          "tools/list: includes glitch_grep")

  (stop-proc proc)
  (cleanup ws))

# --- Test 3: glitch_read_file tool call ---

(do
  (def ws (make-tmp-workspace))
  # Create a test file
  (def test-file (string ws "/hello.txt"))
  (spit test-file "Hello from integration test!\nLine two.\n")

  (def proc (spawn-mcp ws))
  (do-initialize proc)

  (def resp (send-recv proc
    @{"jsonrpc" "2.0" "id" 3
      "method" "tools/call"
      "params" @{"name" "glitch_read_file"
                 "arguments" @{"path" test-file}}}))

  (assert resp "read_file: got a response")
  (def content (get-in resp ["result" "content"]))
  (assert content "read_file: result has content array")
  (assert (> (length content) 0) "read_file: content is non-empty")
  (def text (get-in content [0 "text"]))
  (assert (string/find "Hello from integration test!" text)
          "read_file: response contains file content")
  (assert (string/find "Line two." text)
          "read_file: response contains second line")

  (stop-proc proc)
  (cleanup ws))

# --- Test 4: Unknown tool returns error ---

(do
  (def ws (make-tmp-workspace))
  (def proc (spawn-mcp ws))
  (do-initialize proc)

  (def resp (send-recv proc
    @{"jsonrpc" "2.0" "id" 4
      "method" "tools/call"
      "params" @{"name" "nonexistent"
                 "arguments" @{}}}))

  (assert resp "unknown tool: got a response")
  (def result (get resp "result"))
  (assert result "unknown tool: has result")
  (assert (= (get result "isError") true)
          "unknown tool: isError is true")
  (def text (get-in result ["content" 0 "text"]))
  (assert (string/find "unknown tool" text)
          "unknown tool: error message mentions unknown tool")

  (stop-proc proc)
  (cleanup ws))

# --- Test 5: Invalid JSON returns parse error ---

(do
  (def ws (make-tmp-workspace))
  (def proc (spawn-mcp ws))

  # Write raw invalid JSON
  (ev/write (proc :in) "this is not json at all\n")
  (def response (read-line (proc :out)))
  (assert response "invalid JSON: got a response")
  (def resp (json/decode (string/trim response)))
  (assert (get resp "error") "invalid JSON: response has error")
  (assert (= (get-in resp ["error" "code"]) -32700)
          "invalid JSON: error code is -32700 (Parse error)")

  (stop-proc proc)
  (cleanup ws))

(end-suite)
