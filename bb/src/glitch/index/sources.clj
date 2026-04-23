(ns glitch.index.sources
  "Pluggable index-source registry and pipeline.
   Extends the code-symbol index pipeline to arbitrary data sources
   (GitHub issues, GitLab MRs, support tickets, etc.).

   Each source is a map registered via `defindexer`:
     :doc        - human description
     :detect     - (fn [repo-path]) -> truthy if source applies
     :index-name - (fn [repo-name]) -> ES index name string
     :mapping    - ES mapping {:properties {...}}
     :cli-opts   - [{:flag \"--since\" :doc \"...\"}] extra CLI flags
     :fetch      - (fn [opts]) -> seq of documents to index"
  (:require [clojure.string :as str]
            [glitch.es :as es]
            [babashka.process :as proc]))

;; ---------------------------------------------------------------------------
;; Logging — matches glitch.index convention
;; ---------------------------------------------------------------------------

(defn- log [& args]
  (binding [*out* *err*]
    (apply println args)))

;; ---------------------------------------------------------------------------
;; Registry
;; ---------------------------------------------------------------------------

(def registry
  "Atom mapping source name (string) to source map."
  (atom {}))

(defn- symbol->name
  "Convert a Clojure symbol to kebab-case string name."
  [sym]
  (-> (name sym)
      (str/replace #"_" "-")))

(defmacro defindexer
  "Register an index source. Creates a var and adds it to the registry.

   Usage:
     (defindexer github-issues
       {:doc        \"Index GitHub issues and PRs\"
        :detect     (fn [repo-path] ...)
        :index-name (fn [repo-name] ...)
        :mapping    {:properties {:title {:type \"text\"}}}
        :cli-opts   [{:flag \"--since\" :doc \"Only issues after date\"}]
        :fetch      (fn [opts] ...)})"
  [source-sym source-map]
  (let [src-name (symbol->name source-sym)]
    `(do
       (def ~source-sym (assoc ~source-map :name ~src-name))
       (swap! registry assoc ~src-name ~source-sym)
       ~source-sym)))

;; ---------------------------------------------------------------------------
;; Lookup
;; ---------------------------------------------------------------------------

(defn list-sources
  "Return all registered sources as [{:name :doc :detected}].
   Calls each source's :detect fn with repo-path."
  [repo-path]
  (->> @registry
       vals
       (mapv (fn [src]
               {:name     (:name src)
                :doc      (:doc src)
                :detected (boolean ((:detect src) repo-path))}))))

(defn get-source
  "Look up a source by name string. Returns source map or nil."
  [source-name]
  (get @registry source-name))

;; ---------------------------------------------------------------------------
;; CLI option parsing
;; ---------------------------------------------------------------------------

(defn parse-source-opts
  "Parse source-specific CLI flags from raw args.
   Given a source's :cli-opts spec and a seq of raw CLI arg strings,
   returns a map of parsed flag values (keyword keys, no leading dashes).

   Example:
     cli-opts: [{:flag \"--since\" :doc \"...\"}]
     args:     [\"--since\" \"2024-01-01\" \"--verbose\"]
     => {:since \"2024-01-01\"}"
  [cli-opts args]
  (let [known-flags (set (map :flag cli-opts))
        pairs       (partition-all 2 1 args)]
    (reduce (fn [m [flag val]]
              (if (contains? known-flags flag)
                (let [k (keyword (str/replace flag #"^--?" ""))]
                  (assoc m k val))
                m))
            {}
            pairs)))

;; ---------------------------------------------------------------------------
;; Pipeline
;; ---------------------------------------------------------------------------

(defn fetch-and-index
  "Main pipeline for a pluggable index source.
   1. Look up source from registry
   2. Detect applicability
   3. Fetch documents
   4. Create ES index with mapping
   5. Bulk-index in batches of 500

   Returns {:source source-name :documents count}."
  [es-url repo-path source-name opts]
  (let [src (get-source source-name)]
    (when-not src
      (log "ERROR: unknown source" (pr-str source-name))
      (throw (ex-info (str "Unknown index source: " source-name)
                      {:source source-name})))

    (let [repo-name (last (str/split repo-path #"/"))]
      (log "=== Source:" (:name src) "===")

      ;; Detect
      (if-not ((:detect src) repo-path)
        (do
          (log "  WARN: source" (:name src) "not detected for" repo-path)
          {:source source-name :documents 0})

        ;; Fetch
        (let [_         (log "  Fetching documents...")
              docs      ((:fetch src) (merge opts {:repo-path repo-path
                                                    :repo-name repo-name}))
              doc-count (count docs)
              idx-name  ((:index-name src) repo-name)
              mapping   {:mappings (:mapping src)}]

          (log "  Fetched" doc-count "documents")

          ;; Create index
          (es/create-index es-url idx-name mapping)

          ;; Bulk index in batches of 500
          (when (seq docs)
            (log "  Indexing" doc-count "documents into" idx-name "...")
            (doseq [batch (partition-all 500 docs)]
              (es/bulk-index es-url idx-name batch)
              (log "    batch:" (count batch))))

          (log "  Done.")
          {:source source-name :documents doc-count})))))
