(ns glitch.mcp.handlers
  (:require [babashka.process :as bp]
            [cheshire.core :as json]
            [clojure.string :as str]
            [glitch.index :as index]
            [sci.core :as sci]))

(defn- handle-search [arguments]
  (let [query      (get arguments "query")
        path       (get arguments "path")
        glob       (get arguments "glob")
        fixed      (get arguments "fixed" false)
        multiline  (get arguments "multiline" false)
        pcre2      (get arguments "pcre2" false)
        context    (get arguments "context")
        limit      (get arguments "limit")
        smart-case (get arguments "smart_case" true)
        cmd (cond-> ["rg" "--json"]
              smart-case (conj "-S")
              fixed      (conj "-F")
              multiline  (conj "-U")
              pcre2      (conj "-P")
              context    (conj "-C" (str context))
              limit      (conj "-m" (str limit))
              glob       (conj "-g" glob)
              true       (conj "--" query path))
        result (bp/shell {:out :string :err :string :continue true} cmd)]
    (if (zero? (:exit result))
      (:out result)
      (if (= 1 (:exit result))
        "[]"
        (throw (ex-info (str "rg failed: " (:err result)) {}))))))

(def ^:private symbol-patterns
  {"clojure"    "^\\(def[n\\-]?\\s+%s"
   "go"         "^(func|type|var|const)\\s+.*%s"
   "python"     "^(def|class)\\s+%s"
   "javascript" "(function|const|let|var|class|export)\\s+%s"
   "typescript" "(function|const|let|var|class|export|interface|type)\\s+%s"
   "rust"       "^(fn|struct|enum|trait|type|const|static)\\s+%s"})

(defn- detect-language [path]
  (let [result (bp/shell {:out :string :err :string :continue true}
                 "rg" "--files" "--max-depth" "2" path)
        files (str/split-lines (or (:out result) ""))]
    (cond
      (some #(str/ends-with? % ".clj") files)  "clojure"
      (some #(str/ends-with? % ".go") files)   "go"
      (some #(str/ends-with? % ".py") files)   "python"
      (some #(str/ends-with? % ".ts") files)   "typescript"
      (some #(str/ends-with? % ".js") files)   "javascript"
      (some #(str/ends-with? % ".rs") files)   "rust"
      :else nil)))

(defn- handle-symbols [arguments]
  (let [query    (get arguments "query")
        path     (get arguments "path")
        language (or (get arguments "language")
                     (detect-language path))
        pattern  (if language
                   (format (get symbol-patterns language
                             "(def|fn|func|class|type|struct|const|let|var)\\s+%s")
                           query)
                   query)
        cmd      ["rg" "-n" "-S" "--" pattern path]
        result   (bp/shell {:out :string :err :string :continue true} cmd)]
    (if (<= (:exit result) 1)
      (or (:out result) "")
      (throw (ex-info (str "rg failed: " (:err result)) {})))))

(defn- handle-run [arguments]
  (let [file       (get arguments "file")
        input      (get arguments "input")
        set-params (get arguments "set")
        cmd (cond-> ["glitch" "run"]
              set-params (into (mapcat (fn [[k v]] ["-s" (str k "=" v)]) set-params))
              true       (conj file)
              input      (conj input))
        result (apply bp/shell {:out :string :err :string :continue true} cmd)]
    (if (zero? (:exit result))
      (:out result)
      (throw (ex-info (str "workflow failed (exit " (:exit result) "): " (:err result)) {})))))

(defn- handle-eval [arguments]
  (let [expression (get arguments "expression")
        ctx (sci/init {:namespaces {'user {}}})
        result (sci/eval-string* ctx expression)]
    (str result)))

(defn- handle-check [arguments]
  (let [file (get arguments "file")
        content (slurp file)]
    (try
      (read-string (str "[" content "]"))
      "ok"
      (catch Exception e
        (str "error: " (.getMessage e))))))

(defn- handle-read-file [arguments]
  (let [path (get arguments "path")
        f    (java.io.File. path)]
    (when-not (.exists f)
      (throw (ex-info (str "file not found: " path) {})))
    (let [lines (str/split-lines (slurp f))
          selected (take 200 lines)]
      (str/join "\n" selected))))

(defn- handle-search-symbols [arguments]
  (let [repo (or (get arguments "repo")
                 (last (str/split (System/getProperty "user.dir") #"/")))
        es-url (or (System/getenv "GLITCH_ES_URL") "http://localhost:9200")
        opts {:name     (get arguments "name")
              :kind     (get arguments "kind")
              :language (get arguments "language")
              :file     (get arguments "file")
              :limit    (or (get arguments "limit") 20)}
        results (index/query-symbols es-url repo opts)]
    (json/generate-string results {:pretty true})))

(defn- handle-search-edges [arguments]
  (let [repo (or (get arguments "repo")
                 (last (str/split (System/getProperty "user.dir") #"/")))
        es-url (or (System/getenv "GLITCH_ES_URL") "http://localhost:9200")
        opts {:source (get arguments "source")
              :target (get arguments "target")
              :kind   (get arguments "kind")
              :depth  (or (get arguments "depth") 1)
              :limit  (or (get arguments "limit") 50)}
        results (index/query-edges es-url repo opts)]
    (json/generate-string results {:pretty true})))

(defn- handle-symbol-context [arguments]
  (let [repo (or (get arguments "repo")
                 (last (str/split (System/getProperty "user.dir") #"/")))
        es-url (or (System/getenv "GLITCH_ES_URL") "http://localhost:9200")
        name (get arguments "name")
        result (index/query-context es-url repo name)]
    (json/generate-string (or result {:error "Symbol not found"}) {:pretty true})))

(defn make-handler
  [_context]
  (fn [tool-name arguments]
    (case tool-name
      "glitch_search"         (handle-search arguments)
      "glitch_symbols"        (handle-symbols arguments)
      "glitch_run"            (handle-run arguments)
      "glitch_eval"           (handle-eval arguments)
      "glitch_check"          (handle-check arguments)
      "glitch_read_file"      (handle-read-file arguments)
      "glitch_search_symbols" (handle-search-symbols arguments)
      "glitch_search_edges"   (handle-search-edges arguments)
      "glitch_symbol_context" (handle-symbol-context arguments)
      (throw (ex-info (str "unknown tool: " tool-name) {})))))
