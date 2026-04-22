(ns glitch.mcp.protocol
  (:require [cheshire.core :as json]))

(defn parse-message
  "Parse a JSON string. Returns parsed map or {:error true :message ...}."
  [line]
  (try
    (json/parse-string line)
    (catch Exception e
      {:error true :message (.getMessage e)})))

(defn format-result
  "Build a JSON-RPC 2.0 success response string."
  [id result]
  (json/generate-string {"jsonrpc" "2.0" "id" id "result" result}))

(defn format-error
  "Build a JSON-RPC 2.0 error response string."
  [id code message]
  (json/generate-string {"jsonrpc" "2.0" "id" id
                         "error" {"code" code "message" message}}))

(defn format-tool-result
  "Build an MCP tool success response with content array."
  [id text]
  (format-result id {"content" [{"type" "text" "text" text}]}))

(defn format-tool-error
  "Build an MCP tool error response with isError flag."
  [id text]
  (format-result id {"isError" true
                     "content" [{"type" "text" "text" text}]}))

(defn- handle-initialize []
  {"protocolVersion" "2024-11-05"
   "serverInfo" {"name" "glitch" "version" "0.3.0"}
   "capabilities" {"tools" {}}})

(defn dispatch
  "Route a parsed JSON-RPC message. Returns response string or nil for notifications."
  [msg context]
  (let [id (get msg "id")
        method (get msg "method")]
    (when id
      (case method
        "initialize"
        (format-result id (handle-initialize))

        "tools/list"
        (format-result id {"tools" (or (:tools context) [])})

        "tools/call"
        (let [params (get msg "params" {})
              tool-name (get params "name")
              arguments (get params "arguments" {})
              handler (:tool-handler context)]
          (try
            (let [result (handler tool-name arguments)]
              (format-tool-result id result))
            (catch Exception e
              (format-tool-error id (.getMessage e)))))

        ;; default — unknown method
        (format-error id -32601 "Method not found")))))
