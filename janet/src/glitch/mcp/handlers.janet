# MCP tool call handler implementations.
# Each tool dispatches to the appropriate glitch subsystem.

(import spork/json)
(import glitch/mcp/search :as search)
(import glitch/mcp/indexer :as idx)
(import glitch/mcp/embeddings :as emb)

# ============================================================
# Individual tool handlers
# ============================================================

(defn- handle-search [context arguments]
  (def db (context :search-db))
  (def query (get arguments "query"))
  (def repo (get arguments "repo"))
  (def embed-fn (context :embed-fn))
  (def limit (get arguments "limit" 10))
  (def results (search/hybrid-search db query repo
                 :embed-fn embed-fn :limit limit))
  (json/encode results))

(defn- handle-index [context arguments]
  (def db (context :search-db))
  (def repo (get arguments "repo"))
  (def embed-fn (context :embed-fn))
  (def reindex (get arguments "reindex" false))
  (def result (idx/index-repo db repo
                :embed-fn embed-fn :reindex reindex))
  (json/encode result))

(defn- handle-run [arguments]
  (def workflow (get arguments "workflow"))
  (def input (get arguments "input"))
  (def set-params (get arguments "set"))
  (def cmd @["glitch" "run" workflow])
  (when input (array/push cmd input))
  (when set-params
    (eachp [k v] set-params
      (array/push cmd "-s" (string k "=" v))))
  (def proc (os/spawn cmd :p {:out :pipe :err :pipe}))
  (def stdout (:read (proc :out) :all))
  (def stderr (:read (proc :err) :all))
  (def exit-code (os/proc-wait proc))
  (if (= exit-code 0)
    (string stdout)
    (error (string "workflow failed (exit " exit-code "): " stderr))))

(defn- handle-eval [arguments]
  (def expression (get arguments "expression"))
  (string (eval-string expression)))

(defn- handle-check [arguments]
  (def file (get arguments "file"))
  (def content (slurp file))
  (try
    (do (parse content) "ok")
    ([err] (string "error: " err))))

(defn- handle-grep [context arguments]
  (def pattern (get arguments "pattern"))
  (def path (get arguments "path" (context :workspace-path)))
  (def glob-pat (get arguments "glob"))
  (def cmd @["grep" "-rn" "--color=never" pattern path])
  (when glob-pat (array/push cmd "--include" glob-pat))
  (def proc (os/spawn cmd :p {:out :pipe :err :pipe}))
  (def stdout (:read (proc :out) :all))
  (os/proc-wait proc)
  (string stdout))

(defn- handle-symbols [context arguments]
  (def db (context :search-db))
  (def query (get arguments "query"))
  (def repo (get arguments "repo"))
  (def kind (get arguments "kind"))
  # Use FTS5 keyword search on chunks_fts, then filter by symbols
  (def kw-results (search/keyword-search db query :repo repo :limit 50))
  (def filtered @[])
  (each r kw-results
    (when (string/find query (or (r :symbols) ""))
      (array/push filtered
        {:path (r :path)
         :symbols (r :symbols)
         :score (r :score)})))
  (json/encode filtered))

(defn- handle-read-file [arguments]
  (def path (get arguments "path"))
  (unless (os/stat path)
    (error (string "file not found: " path)))
  (def content (slurp path))
  (def lines (string/split "\n" content))
  (def limit (min 200 (length lines)))
  (def selected (array/slice lines 0 limit))
  (string/join selected "\n"))

# ============================================================
# Public API
# ============================================================

(defn make-handler [context]
  "Create a tool dispatch function from a context table.
   context keys: :search-db :workspace-path :embed-fn
   Returns (fn [tool-name arguments] ...) -> string result."
  (fn [tool-name arguments]
    (case tool-name
      "glitch_search"    (handle-search context arguments)
      "glitch_index"     (handle-index context arguments)
      "glitch_run"       (handle-run arguments)
      "glitch_eval"      (handle-eval arguments)
      "glitch_check"     (handle-check arguments)
      "glitch_grep"      (handle-grep context arguments)
      "glitch_symbols"   (handle-symbols context arguments)
      "glitch_read_file" (handle-read-file arguments)
      (error (string "unknown tool: " tool-name)))))
