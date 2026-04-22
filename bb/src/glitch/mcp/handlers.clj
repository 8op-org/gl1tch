(ns glitch.mcp.handlers
  (:require [babashka.process :as bp]
            [cheshire.core :as json]
            [clojure.string :as str]
            [glitch.index :as index]
            [sci.core :as sci]))

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
      "glitch_run"            (handle-run arguments)
      "glitch_eval"           (handle-eval arguments)
      "glitch_check"          (handle-check arguments)
      "glitch_search_symbols" (handle-search-symbols arguments)
      "glitch_search_edges"   (handle-search-edges arguments)
      "glitch_symbol_context" (handle-symbol-context arguments)
      (throw (ex-info (str "unknown tool: " tool-name) {})))))
