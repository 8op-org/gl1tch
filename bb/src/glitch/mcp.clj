(ns glitch.mcp
  "MCP stdio server entry point.
   Reads JSON-RPC messages from stdin, dispatches via protocol, writes to stdout."
  (:require [glitch.mcp.protocol :as proto]
            [glitch.mcp.tools :as tools]
            [glitch.mcp.handlers :as handlers]
            [clojure.string :as str]))

(defn start [_opts]
  (let [handler (handlers/make-handler {})
        dispatch-ctx {:tools tools/tool-definitions
                      :tool-handler handler}]
    (binding [*out* *err*]
      (println "[glitch-mcp] server started"))
    (try
      (loop []
        (when-let [line (read-line)]
          (let [trimmed (str/trim line)]
            (when (seq trimmed)
              (let [msg (proto/parse-message trimmed)]
                (if (:error msg)
                  (do
                    (println (proto/format-error nil -32700 "Parse error"))
                    (flush))
                  (when-let [resp (proto/dispatch msg dispatch-ctx)]
                    (println resp)
                    (flush))))))
          (recur)))
      (finally
        (binding [*out* *err*]
          (println "[glitch-mcp] server stopped"))))))

(defn -main [& _args]
  (start {}))
