(ns glitch.mcp.handlers
  (:require [glitch.mcp.search :as search]
            [glitch.mcp.indexer :as idx]
            [babashka.process :as bp]
            [cheshire.core :as json]
            [clojure.string :as str]
            [sci.core :as sci]))

(defn- confined-path
  "Resolve path and verify it's under workspace-path. Throws if not."
  [path workspace-path]
  (let [resolved (.getCanonicalPath (java.io.File. path))
        base (.getCanonicalPath (java.io.File. workspace-path))]
    (when-not (str/starts-with? resolved base)
      (throw (ex-info (str "path outside workspace: " path) {})))
    resolved))

(defn- handle-search [context arguments]
  (let [db (:search-db context)
        query (get arguments "query")
        repo (get arguments "repo")
        embed-fn (:embed-fn context)
        limit (get arguments "limit" 10)]
    (json/generate-string
      (search/hybrid-search db query repo :embed-fn embed-fn :limit limit))))

(defn- handle-index [context arguments]
  (let [db (:search-db context)
        repo (get arguments "repo")
        embed-fn (:embed-fn context)
        reindex (get arguments "reindex" false)]
    (json/generate-string
      (idx/index-repo db repo :embed-fn embed-fn :reindex reindex))))

(defn- handle-run [arguments]
  (let [workflow (get arguments "workflow")
        input (get arguments "input")
        set-params (get arguments "set")
        cmd (cond-> ["glitch" "run" workflow]
              input (conj input)
              set-params (into (mapcat (fn [[k v]] ["-s" (str k "=" v)]) set-params)))
        result (apply bp/shell {:out :string :err :string :continue true} cmd)]
    (if (zero? (:exit result))
      (:out result)
      (throw (ex-info (str "workflow failed (exit " (:exit result) "): " (:err result)) {})))))

(defn- handle-eval [arguments]
  (let [expression (get arguments "expression")
        ctx (sci/init {:namespaces {'user {}}})
        result (sci/eval-string* ctx expression)]
    (str result)))

(defn- handle-check [context arguments]
  (let [file (get arguments "file")
        _ (confined-path file (:workspace-path context))
        content (slurp file)]
    (try
      (read-string (str "[" content "]"))
      "ok"
      (catch Exception e
        (str "error: " (.getMessage e))))))

(defn- handle-grep [context arguments]
  (let [pattern (get arguments "pattern")
        path (get arguments "path" (:workspace-path context))
        _ (confined-path path (:workspace-path context))
        glob-pat (get arguments "glob")
        cmd (cond-> ["grep" "-rn" "--color=never" "--" pattern path]
              glob-pat (conj "--include" glob-pat))
        result (apply bp/shell {:out :string :err :string :continue true} cmd)]
    (or (:out result) "")))

(defn- handle-symbols [context arguments]
  (let [db (:search-db context)
        query (get arguments "query")
        repo (get arguments "repo")
        kw-results (search/keyword-search db query :repo repo :limit 50)
        filtered (filter #(str/includes? (or (:symbols %) "") query) kw-results)]
    (json/generate-string
      (mapv #(select-keys % [:path :symbols :score]) filtered))))

(defn- handle-read-file [context arguments]
  (let [path (get arguments "path")
        _ (confined-path path (:workspace-path context))
        f (java.io.File. path)]
    (when-not (.exists f)
      (throw (ex-info (str "file not found: " path) {})))
    (let [lines (str/split-lines (slurp f))
          selected (take 200 lines)]
      (str/join "\n" selected))))

(defn make-handler [context]
  (fn [tool-name arguments]
    (case tool-name
      "glitch_search"    (handle-search context arguments)
      "glitch_index"     (handle-index context arguments)
      "glitch_run"       (handle-run arguments)
      "glitch_eval"      (handle-eval arguments)
      "glitch_check"     (handle-check context arguments)
      "glitch_grep"      (handle-grep context arguments)
      "glitch_symbols"   (handle-symbols context arguments)
      "glitch_read_file" (handle-read-file context arguments)
      (throw (ex-info (str "unknown tool: " tool-name) {})))))
