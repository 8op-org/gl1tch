(ns glitch.mcp.handlers
  (:require [babashka.process :as bp]
            [cheshire.core :as json]
            [clojure.string :as str]
            [glitch.core :as g]
            [glitch.index :as index]
            [glitch.session :as session]
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

(defn- make-sci-ctx []
  (sci/init {:namespaces
             {'user
              {'trace            g/trace
               'input            g/input
               'params           g/params
               'param            g/param
               'ref              g/ref
               'sh              g/sh
               'search           g/search
               'save             g/save
               'read-file        g/read-file
               'write-file       g/write-file
               'get-steps        g/get-steps
               'last-output      g/last-output
               'gate             g/gate
               'call-workflow    g/call-workflow
               'json-extract     g/json-extract
               'validate-schema  g/validate-schema
               'validate         g/validate
               'llm              g/llm
               'grounded?        g/grounded?
               'consensus        g/consensus
               'composite-score  g/composite-score
               'search-symbols   g/search-symbols
               'search-edges     g/search-edges
               'symbol-context   g/symbol-context}
              'clojure.string
              {'upper-case   clojure.string/upper-case
               'lower-case   clojure.string/lower-case
               'trim         clojure.string/trim
               'split        clojure.string/split
               'join         clojure.string/join
               'replace      clojure.string/replace
               'starts-with? clojure.string/starts-with?
               'ends-with?   clojure.string/ends-with?
               'includes?    clojure.string/includes?
               'blank?       clojure.string/blank?}}}))

(defn- handle-eval [sci-ctx arguments]
  (let [expression (get arguments "expression")
        result (sci/eval-string* sci-ctx expression)]
    (str result)))

(defn- handle-check [arguments]
  (let [file (get arguments "file")
        result (bp/shell {:out :string :err :string :continue true}
                         "glitch" "check" file)]
    (if (zero? (:exit result))
      (str/trim (:out result))
      (str "error: " (str/trim (:err result))))))

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

(defn- extract-description
  "Extract description from the first ;; comment line of a workflow file."
  [file]
  (try
    (let [lines (str/split-lines (slurp file))]
      (if-let [comment-line (first (filter #(str/starts-with? (str/trim %) ";;") lines))]
        (str/trim (subs (str/trim comment-line) 2))
        ""))
    (catch Exception _ "")))

(defn- handle-list-workflows [arguments]
  (let [path (or (get arguments "path") ".glitch/workflows")
        dir  (java.io.File. path)]
    (if (.isDirectory dir)
      (let [files (->> (or (.listFiles dir) [])
                       (filter #(or (str/ends-with? (.getName %) ".glitch")
                                    (str/ends-with? (.getName %) ".clj")))
                       sort)
            entries (mapv (fn [f]
                           {"name"        (str/replace (.getName f) #"\.(glitch|clj)$" "")
                            "file"        (.getAbsolutePath f)
                            "description" (extract-description f)})
                         files)]
        (json/generate-string entries {:pretty true}))
      (json/generate-string [] {:pretty true}))))

(defn- handle-recall [arguments]
  (let [query   (get arguments "query")
        results (session/recall-search query)]
    (if (seq results)
      (str/join "\n" (map (fn [r]
                            (str (:path r "")
                                 " — "
                                 (:description r "")))
                          results))
      "No matching workflows found.")))

(defn- handle-advise [arguments]
  (let [task    (get arguments "task")
        context (get arguments "context" "")
        input   (if (seq context)
                  (str task "\n\nContext: " context)
                  task)
        cmd     ["glitch" "run" ".glitch/workflows/advise.glitch" input]
        result  (apply bp/shell {:out :string :err :string :continue true} cmd)
        response (if (zero? (:exit result))
                   (let [raw (str/trim (:out result))]
                     (try
                       (let [extracted (-> raw
                                          (str/replace #"^```[a-z]*\n?" "")
                                          (str/replace #"\n?```\s*$" "")
                                          str/trim)
                             start (str/index-of extracted "{")
                             end   (when start (inc (str/last-index-of extracted "}")))
                             json-str (if (and start end)
                                        (subs extracted start end)
                                        extracted)
                             parsed (json/parse-string json-str)]
                         {"approach"           (get parsed "approach" "none")
                          "primitives"         (get parsed "primitives" [])
                          "reasoning"          (get parsed "reasoning" "")
                          "example"            (get parsed "example" "")
                          "existing_workflows" (get parsed "existing_workflows" [])})
                       (catch Exception _
                         {"approach" "none"
                          "primitives" []
                          "reasoning" (str "Advisory workflow returned unparseable response: " raw)
                          "example" ""
                          "existing_workflows" []})))
                   {"approach" "none"
                    "primitives" []
                    "reasoning" (str "Advisory workflow failed: " (str/trim (:err result)))
                    "example" ""
                    "existing_workflows" []})]
    ;; Record in session (silently ignore errors)
    (try
      (session/record-advise! {:task task :recommendation response})
      (catch Exception _ nil))
    (json/generate-string response {:pretty true})))

(defn make-handler
  [_context]
  (let [sci-ctx (make-sci-ctx)]
    (fn [tool-name arguments]
      (case tool-name
        "glitch_run"            (handle-run arguments)
        "glitch_eval"           (handle-eval sci-ctx arguments)
        "glitch_check"          (handle-check arguments)
        "glitch_search_symbols" (handle-search-symbols arguments)
        "glitch_search_edges"   (handle-search-edges arguments)
        "glitch_symbol_context" (handle-symbol-context arguments)
        "glitch_list_workflows" (handle-list-workflows arguments)
        "glitch_recall"         (handle-recall arguments)
        "glitch_advise"         (handle-advise arguments)
        (throw (ex-info (str "unknown tool: " tool-name) {}))))))
