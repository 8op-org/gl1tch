# MCP stdio server entry point.
# Reads JSON-RPC messages from stdin, dispatches via protocol, writes to stdout.
# All logging goes to stderr — stdout is the JSON-RPC channel.

(import sqlite3 :as sql)
(import glitch/mcp/protocol :as proto)
(import glitch/mcp/tools :as tools)
(import glitch/mcp/handlers :as handlers)
(import glitch/mcp/indexer :as idx)
(import glitch/mcp/embeddings :as emb)

(defn start [opts]
  "Start the MCP stdio server.
   opts: {:workspace-path :model :base-url}"
  (def workspace-path (or (opts :workspace-path) (os/cwd)))
  (def model (opts :model))
  (def base-url (opts :base-url))

  # Open search database
  (def search-db (idx/open-search-db workspace-path))

  # Build embed function
  (def embed-fn
    (fn [texts]
      (emb/embed texts :model model :base-url base-url)))

  # Build handler context
  (def context
    @{:search-db search-db
      :workspace-path workspace-path
      :embed-fn embed-fn})

  # Create tool handler
  (def handler (handlers/make-handler context))

  # Dispatch context for protocol
  (def dispatch-ctx
    @{:tools tools/tool-definitions
      :tool-handler handler})

  (eprint "[glitch-mcp] server started")

  # Stdio loop
  (defer (do
           (sql/close search-db)
           (eprint "[glitch-mcp] server stopped"))
    (while true
      (def line (file/read stdin :line))
      (unless line (break))
      (def trimmed (string/trim line))
      (when (not= trimmed "")
        (def msg (proto/parse-message trimmed))
        (if (get msg :error)
          # Parse error
          (do
            (def resp (proto/format-error nil -32700 "Parse error"))
            (file/write stdout resp)
            (file/write stdout "\n")
            (file/flush stdout))
          # Valid message — dispatch
          (when-let [resp (proto/dispatch msg dispatch-ctx)]
            (file/write stdout resp)
            (file/write stdout "\n")
            (file/flush stdout)))))))
