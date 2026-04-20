(ns glitch.mcp
  "MCP stdio server entry point.
   Reads JSON-RPC messages from stdin, dispatches via protocol, writes to stdout."
  (:require [glitch.mcp.protocol :as proto]
            [glitch.mcp.tools :as tools]
            [glitch.mcp.handlers :as handlers]
            [glitch.mcp.indexer :as idx]
            [glitch.mcp.embeddings :as emb]
            [clojure.string :as str]))

(defn start [{:keys [workspace-path model base-url]}]
  (let [workspace-path (or workspace-path (System/getProperty "user.dir"))
        search-db (idx/open-search-db workspace-path)
        embed-fn (fn [texts]
                   (emb/embed texts
                     :model (or model "nomic-embed-text")
                     :base-url (or base-url "http://localhost:1234")))
        context {:search-db search-db
                 :workspace-path workspace-path
                 :embed-fn embed-fn}
        handler (handlers/make-handler context)
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
        (idx/close-search-db search-db)
        (binding [*out* *err*]
          (println "[glitch-mcp] server stopped"))))))

(defn -main [& _args]
  (start {:workspace-path (System/getProperty "user.dir")}))
