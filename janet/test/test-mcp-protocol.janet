(use spork/test)
(import spork/json)

(array/push module/paths ["src/:all:.janet" :source])
(import glitch/mcp/protocol :as proto)

(start-suite "mcp-protocol")

# --- parse-message ---

(do
  (def msg (proto/parse-message `{"jsonrpc":"2.0","method":"initialize","id":1}`))
  (assert (not (msg :error)) "parse-message: valid JSON succeeds")
  (assert (= (get msg "method") "initialize")
          "parse-message: method parsed correctly")
  (assert (= (get msg "id") 1)
          "parse-message: id parsed correctly"))

(do
  (def msg (proto/parse-message "not json at all"))
  (assert (msg :error) "parse-message: invalid JSON returns error")
  (assert (msg :message) "parse-message: error has message"))

# --- format-result ---

(do
  (def out (proto/format-result 1 @{"name" "glitch"}))
  (def parsed (json/decode out))
  (assert (= (get parsed "jsonrpc") "2.0")
          "format-result: has jsonrpc 2.0")
  (assert (= (get parsed "id") 1)
          "format-result: has correct id")
  (assert (= (get (get parsed "result") "name") "glitch")
          "format-result: has result payload"))

# --- format-error ---

(do
  (def out (proto/format-error 1 -32601 "Method not found"))
  (def parsed (json/decode out))
  (assert (= (get parsed "jsonrpc") "2.0")
          "format-error: has jsonrpc 2.0")
  (assert (= (get parsed "id") 1)
          "format-error: has correct id")
  (def err (get parsed "error"))
  (assert (= (get err "code") -32601)
          "format-error: has error code")
  (assert (= (get err "message") "Method not found")
          "format-error: has error message"))

# --- format-tool-result ---

(do
  (def out (proto/format-tool-result 1 "hello world"))
  (def parsed (json/decode out))
  (def result (get parsed "result"))
  (def content (get result "content"))
  (assert (= (length content) 1)
          "format-tool-result: content has one item")
  (def item (get content 0))
  (assert (= (get item "type") "text")
          "format-tool-result: item type is text")
  (assert (= (get item "text") "hello world")
          "format-tool-result: item text matches"))

# --- format-tool-error ---

(do
  (def out (proto/format-tool-error 1 "something broke"))
  (def parsed (json/decode out))
  (def result (get parsed "result"))
  (assert (= (get result "isError") true)
          "format-tool-error: isError is true")
  (def content (get result "content"))
  (def item (get content 0))
  (assert (= (get item "type") "text")
          "format-tool-error: item type is text")
  (assert (= (get item "text") "something broke")
          "format-tool-error: item text matches"))

# --- handle-initialize ---

(do
  (def info (proto/handle-initialize))
  (def server (info :serverInfo))
  (assert (= (server :name) "glitch")
          "handle-initialize: server name is glitch")
  (assert (= (server :version) "0.1.0")
          "handle-initialize: version is 0.1.0")
  (assert (info :capabilities)
          "handle-initialize: has capabilities")
  (assert ((info :capabilities) :tools)
          "handle-initialize: has tools capability"))

# --- dispatch ---

# dispatch routes initialize
(do
  (def msg @{"jsonrpc" "2.0" "method" "initialize" "id" 1})
  (def out (proto/dispatch msg @{}))
  (assert out "dispatch: initialize returns a response")
  (def parsed (json/decode out))
  (def result (get parsed "result"))
  (assert (get result "serverInfo")
          "dispatch: initialize has serverInfo"))

# dispatch routes tools/list
(do
  (def tools @[@{"name" "greet" "description" "say hi"}])
  (def ctx @{:tools tools})
  (def msg @{"jsonrpc" "2.0" "method" "tools/list" "id" 2})
  (def out (proto/dispatch msg ctx))
  (def parsed (json/decode out))
  (def result (get parsed "result"))
  (assert (= (length (get result "tools")) 1)
          "dispatch: tools/list returns tools from context"))

# dispatch returns error for unknown method
(do
  (def msg @{"jsonrpc" "2.0" "method" "bogus/method" "id" 3})
  (def out (proto/dispatch msg @{}))
  (def parsed (json/decode out))
  (assert (get parsed "error")
          "dispatch: unknown method returns error")
  (assert (= (get (get parsed "error") "code") -32601)
          "dispatch: unknown method has -32601 code"))

# dispatch ignores notifications (no id) — returns nil
(do
  (def msg @{"jsonrpc" "2.0" "method" "notifications/initialized"})
  (def out (proto/dispatch msg @{}))
  (assert (nil? out)
          "dispatch: notification without id returns nil"))

# dispatch routes tools/call with mock handler
(do
  (def handler (fn [name args]
                 (string "Hello, " (get args "who"))))
  (def ctx @{:tools @[]
             :tool-handler handler})
  (def msg @{"jsonrpc" "2.0" "method" "tools/call" "id" 4
             "params" @{"name" "greet"
                        "arguments" @{"who" "world"}}})
  (def out (proto/dispatch msg ctx))
  (def parsed (json/decode out))
  (def result (get parsed "result"))
  (def content (get result "content"))
  (assert (= (get (get content 0) "text") "Hello, world")
          "dispatch: tools/call invokes handler and wraps result"))

# dispatch routes tools/call — handler error wraps with format-tool-error
(do
  (def handler (fn [name args] (error "kaboom")))
  (def ctx @{:tools @[] :tool-handler handler})
  (def msg @{"jsonrpc" "2.0" "method" "tools/call" "id" 5
             "params" @{"name" "explode" "arguments" @{}}})
  (def out (proto/dispatch msg ctx))
  (def parsed (json/decode out))
  (def result (get parsed "result"))
  (assert (= (get result "isError") true)
          "dispatch: tools/call error sets isError"))

(end-suite)
