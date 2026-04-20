(import spork/json)

(defn parse-message
  "Parse a JSON string into a table. Returns {:error true :message ...} on invalid JSON."
  [line]
  (try
    (json/decode line)
    ([err]
      @{:error true :message (string err)})))

(defn format-result
  "Build a JSON-RPC 2.0 success response string."
  [id result]
  (json/encode @{"jsonrpc" "2.0" "id" id "result" result}))

(defn format-error
  "Build a JSON-RPC 2.0 error response string."
  [id code message]
  (json/encode @{"jsonrpc" "2.0" "id" id
                 "error" @{"code" code "message" message}}))

(defn format-tool-result
  "Build an MCP tool success response with content array."
  [id text]
  (format-result id @{"content" @[@{"type" "text" "text" text}]}))

(defn format-tool-error
  "Build an MCP tool error response with isError flag."
  [id text]
  (format-result id @{"isError" true
                      "content" @[@{"type" "text" "text" text}]}))

(defn handle-initialize
  "Return server info and capabilities for the initialize handshake."
  []
  @{:serverInfo @{:name "glitch" :version "0.1.0"}
    :capabilities @{:tools @{}}})

(defn dispatch
  "Route a parsed JSON-RPC message. Returns response string or nil for notifications."
  [msg context]
  (def id (get msg "id"))
  (def method (get msg "method"))
  # notifications have no id — ignore
  (unless id (break nil))
  (case method
    "initialize"
    (format-result id (handle-initialize))

    "tools/list"
    (format-result id @{"tools" (or (context :tools) @[])})

    "tools/call"
    (do
      (def params (get msg "params" @{}))
      (def tool-name (get params "name"))
      (def arguments (get params "arguments" @{}))
      (def handler (context :tool-handler))
      (try
        (let [result (handler tool-name arguments)]
          (format-tool-result id result))
        ([err]
          (format-tool-error id (string err)))))

    # default — unknown method
    (format-error id -32601 "Method not found")))
