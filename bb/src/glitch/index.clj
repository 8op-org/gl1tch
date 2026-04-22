(ns glitch.index
  "Code intelligence extraction pipeline.
   Three-phase: extract symbols, resolve cross-file edges, index to ES."
  (:require [babashka.process :as proc]
            [cheshire.core :as json]
            [clojure.string :as str]
            [glitch.es :as es]
            [glitch.index.clojure :as clj-index]
            [glitch.index.languages :as lang])
  (:import [java.security MessageDigest]))

;; ---------------------------------------------------------------------------
;; ES index mappings
;; ---------------------------------------------------------------------------

(def symbols-mappings
  {:mappings
   {:properties
    {:id         {:type "keyword"}
     :file       {:type "keyword"}
     :kind       {:type "keyword"}
     :name       {:type "keyword"}
     :signature  {:type "text"}
     :language   {:type "keyword"}
     :start_line {:type "integer"}
     :end_line   {:type "integer"}
     :parent_id  {:type "keyword"}
     :docstring  {:type "text"}
     :file_hash  {:type "keyword"}
     :repo       {:type "keyword"}
     :indexed_at {:type "date"}}}})

(def edges-mappings
  {:mappings
   {:properties
    {:source {:type "keyword"}
     :target {:type "keyword"}
     :kind   {:type "keyword"}
     :file   {:type "keyword"}
     :repo   {:type "keyword"}}}})

;; ---------------------------------------------------------------------------
;; Helpers
;; ---------------------------------------------------------------------------

(defn- log [& args]
  (binding [*out* *err*]
    (apply println args)))

(defn- symbols-index [repo] (str "glitch-symbols-" repo))
(defn- edges-index [repo] (str "glitch-edges-" repo))

(defn- repo-name
  "Basename of a repo path."
  [repo-path]
  (last (str/split repo-path #"/")))

(defn- now-iso []
  (str (java.time.Instant/now)))

;; ---------------------------------------------------------------------------
;; File discovery
;; ---------------------------------------------------------------------------

(defn find-source-files
  "Walk repo directory, return seq of files matching supported language extensions.
   Respects .gitignore via `git ls-files`."
  [repo-path & {:keys [languages]}]
  (let [result (proc/shell {:out :string :err :string :continue true
                            :dir repo-path}
                           "git" "ls-files" "--cached" "--others" "--exclude-standard")
        files  (when (zero? (:exit result))
                 (str/split-lines (str/trim (:out result))))
        lang-set (when languages (set languages))]
    (->> (or files [])
         (filter (fn [f]
                   (let [detected (lang/detect-language f)]
                     (and detected
                          (or (nil? lang-set)
                              (contains? lang-set detected))))))
         (mapv (fn [f]
                 (if (str/starts-with? f "/")
                   f
                   (str repo-path "/" f)))))))

;; ---------------------------------------------------------------------------
;; SHA256 hashing
;; ---------------------------------------------------------------------------

(defn file-hash
  "SHA256 hash of file contents."
  [path]
  (let [md      (MessageDigest/getInstance "SHA-256")
        content (.getBytes (slurp path) "UTF-8")
        digest  (.digest md content)]
    (apply str (map #(format "%02x" %) digest))))

;; ---------------------------------------------------------------------------
;; ast-grep execution
;; ---------------------------------------------------------------------------

(defn run-sg
  "Run ast-grep on a file with a rule. Returns parsed JSON matches.
   Each match has: :text :range :file :ruleId :language :labels etc."
  [rule-file source-file]
  (let [result (proc/shell {:out :string :err :string :continue true}
                           "sg" "scan" "--rule" rule-file
                           "--json=stream" source-file)]
    (if (and (zero? (:exit result))
             (not (str/blank? (:out result))))
      (->> (str/split-lines (str/trim (:out result)))
           (mapv #(json/parse-string % true))
           ;; ast-grep >=0.38 emits matches directly as JSON objects;
           ;; older versions wrapped them in {"type":"match","data":{...}}.
           ;; Unwrap if needed, otherwise use as-is.
           (mapv (fn [m]
                   (if (= "match" (:type m))
                     (:data m)
                     m))))
      [])))

;; ---------------------------------------------------------------------------
;; Kind detection from ast-grep match text / ruleId
;; ---------------------------------------------------------------------------

(def ^:private kind-patterns
  "Ordered patterns to detect symbol kind from the matched text."
  [;; Methods MUST come before function patterns (Go receiver)
   [#"^func\s*\("                  "method"]     ;; Go method: func (r Recv) Name(
   ;; Go / Rust / C / Java functions
   [#"^func\s"                     "function"]
   [#"^fn\s"                       "function"]
   [#"^\w+\s+\w+\s*\("            "function"]   ;; C-style: int main(
   [#"^def\s"                      "function"]   ;; Python
   [#"^function\s"                 "function"]   ;; JS
   [#"^(export\s+)?(async\s+)?function\s" "function"]
   [#"^(const|let|var)\s+\w+\s*=" "const"]      ;; JS const/let/var
   ;; Types
   [#"^type\s+\w+\s+struct\b"     "struct"]
   [#"^type\s+\w+\s+interface\b"  "interface"]
   [#"^type\s"                     "type"]
   [#"^struct\s"                   "struct"]
   [#"^enum\s"                     "enum"]
   [#"^trait\s"                    "trait"]
   [#"^impl\s"                     "impl"]
   [#"^class\s"                    "class"]
   ;; Imports
   [#"^import\s"                   "import"]
   [#"^from\s"                     "import"]
   [#"^use\s"                      "import"]
   ;; Constants / vars
   [#"^const\s"                    "const"]
   [#"^static\s"                   "static"]
   [#"^var\s"                      "var"]
   [#"^pub\s+static\s"            "static"]
   [#"^pub\s+fn\s"                "function"]
   [#"^pub\s+struct\s"            "struct"]
   [#"^pub\s+enum\s"              "enum"]
   [#"^pub\s+trait\s"             "trait"]])

(defn- detect-kind
  "Detect the symbol kind from matched text."
  [text]
  (let [trimmed (str/trim text)]
    (or (some (fn [[pat kind]]
                (when (re-find pat trimmed) kind))
              kind-patterns)
        "unknown")))

;; ---------------------------------------------------------------------------
;; Name extraction from matched text
;; ---------------------------------------------------------------------------

(def ^:private name-patterns
  "Patterns to extract the symbol name from match text."
  [;; Go: func Name(  or  func (r Recv) Name(
   [#"^func\s+\([\w\s*]+\)\s+(\w+)" 1]
   [#"^func\s+(\w+)"                 1]
   ;; Go: type Name struct/interface/...
   [#"^type\s+(\w+)"                 1]
   ;; Rust: fn name / pub fn name
   [#"(?:pub\s+)?fn\s+(\w+)"         1]
   ;; Rust: struct/enum/trait/impl
   [#"(?:pub\s+)?(?:struct|enum|trait)\s+(\w+)" 1]
   [#"^impl(?:<[^>]*>)?\s+(\w+)"     1]
   ;; Python: def name / class name
   [#"^(?:async\s+)?def\s+(\w+)"     1]
   [#"^class\s+(\w+)"                1]
   ;; JavaScript/TypeScript: function name / class name / const name =
   [#"^(?:export\s+)?(?:async\s+)?function\s+(\w+)" 1]
   [#"^(?:export\s+)?class\s+(\w+)"  1]
   [#"^(?:export\s+)?(?:const|let|var)\s+(\w+)" 1]
   ;; Java: class/interface/enum
   [#"(?:public|private|protected)?\s*(?:static\s+)?(?:final\s+)?class\s+(\w+)" 1]
   [#"(?:public|private|protected)?\s*interface\s+(\w+)" 1]
   [#"(?:public|private|protected)?\s*enum\s+(\w+)" 1]
   ;; Java method: type name(
   [#"(?:public|private|protected)?\s*(?:static\s+)?(?:final\s+)?(?:synchronized\s+)?\w+(?:<[^>]*>)?\s+(\w+)\s*\(" 1]
   ;; C: type name( or struct name
   [#"^(?:static\s+)?(?:inline\s+)?\w+\s+(\w+)\s*\(" 1]
   [#"^struct\s+(\w+)"               1]
   [#"^enum\s+(\w+)"                 1]
   ;; Go const/var: const Name or var Name
   [#"^(?:const|var)\s+(\w+)"        1]
   ;; Rust: const/static
   [#"(?:pub\s+)?(?:const|static)\s+(\w+)" 1]])

(defn- extract-name
  "Extract symbol name from matched text."
  [text]
  (let [trimmed (str/trim text)]
    (or (some (fn [[pat grp]]
                (when-let [m (re-find pat trimmed)]
                  (if (string? m) m (nth m grp nil))))
              name-patterns)
        ;; Fallback: first word-like token after any keywords
        (second (re-find #"(?:func|fn|def|class|type|struct|enum|trait|impl|const|var|static|function|interface)\s+(\w+)" trimmed))
        ;; Last resort: first identifier
        (re-find #"\w+" trimmed))))

;; ---------------------------------------------------------------------------
;; Import name extraction
;; ---------------------------------------------------------------------------

(defn- extract-import-name
  "Extract the imported module/package name from import text."
  [text]
  (let [trimmed (str/trim text)]
    (or
     ;; Go: import "pkg" or import alias "pkg"
     (when-let [m (re-find #"\"([^\"]+)\"" trimmed)]
       (last (str/split (second m) #"/")))
     ;; Python: from X import Y  or  import X
     (when-let [m (re-find #"^from\s+([\w.]+)" trimmed)]
       (second m))
     (when-let [m (re-find #"^import\s+([\w.]+)" trimmed)]
       (second m))
     ;; JS: import ... from "mod"  or  import "mod"
     (when-let [m (re-find #"from\s+['\"]([^'\"]+)['\"]" trimmed)]
       (last (str/split (second m) #"/")))
     ;; Rust: use crate::module
     (when-let [m (re-find #"^use\s+([\w:]+)" trimmed)]
       (last (str/split (second m) #"::")))
     ;; Java: import pkg.Class
     (when-let [m (re-find #"^import\s+(?:static\s+)?([\w.]+)" trimmed)]
       (last (str/split (second m) #"\.")))
     ;; Fallback
     (re-find #"\w+" trimmed))))

;; ---------------------------------------------------------------------------
;; Call name extraction
;; ---------------------------------------------------------------------------

(defn- extract-call-name
  "Extract the function/method name from a call expression."
  [text]
  (let [trimmed (str/trim text)]
    (or
     ;; method.call(  or  pkg.Func(
     (when-let [m (re-find #"(\w+)\s*\(" trimmed)]
       (second m))
     (re-find #"\w+" trimmed))))

;; ---------------------------------------------------------------------------
;; Extends / Implements / Exports extraction
;; ---------------------------------------------------------------------------

(defn- extract-extends
  "Extract the superclass/parent name from a class/struct declaration.
   Returns nil if no inheritance found."
  [text]
  (or
   ;; Java/JS/TS: class Foo extends Bar
   (second (re-find #"extends\s+(\w+)" text))
   ;; Python: class Foo(Bar)
   (second (re-find #"class\s+\w+\s*\((\w+)\)" text))
   ;; Rust: impl Trait for Struct
   (when-let [m (re-find #"impl\s+(\w+)\s+for\s+" text)]
     nil)))  ;; impl X for Y → implements, not extends

(defn- extract-implements
  "Extract interface/trait names from a declaration.
   Returns seq of names."
  [text]
  (concat
   ;; Java: class Foo implements Bar, Baz
   (when-let [m (re-find #"implements\s+([\w\s,]+)" text)]
     (map str/trim (str/split (second m) #",")))
   ;; Rust: impl Trait for Struct
   (when-let [m (re-find #"impl\s+(\w+)\s+for\s+" text)]
     [(second m)])))

(defn- exported?
  "Check if a symbol is exported (public) based on language conventions."
  [name language]
  (case language
    "go"   (when name (Character/isUpperCase (first name)))
    ;; Python: no underscore prefix = public
    "python" (when name (not (str/starts-with? name "_")))
    ;; JS/TS: check for export keyword in signature (handled during extraction)
    ;; Java/Rust/C: public by default in most cases
    true))

;; ---------------------------------------------------------------------------
;; Parent-child nesting (contains edges)
;; ---------------------------------------------------------------------------

(defn- find-parent
  "Find the narrowest enclosing symbol for a given line range.
   symbols must be sorted by start_line."
  [symbols start-line end-line]
  (let [enclosing (filter (fn [s]
                            (and (< (:start_line s) start-line)
                                 (> (:end_line s) end-line)))
                          symbols)]
    (when (seq enclosing)
      ;; Narrowest = smallest span among enclosing
      (:id (apply min-key (fn [s] (- (:end_line s) (:start_line s))) enclosing)))))

;; ---------------------------------------------------------------------------
;; Symbol extraction (Phase 1)
;; ---------------------------------------------------------------------------

(defn extract-symbols
  "Extract symbols from a single file using ast-grep.
   Returns {:symbols [...] :raw-imports [...] :raw-calls [...] :raw-refs [...] :contains-edges [...]}"
  [repo-path file-path language]
  (let [repo     (repo-name repo-path)
        rel-file (if (str/starts-with? file-path repo-path)
                   (subs file-path (inc (count repo-path)))
                   file-path)
        hash     (file-hash file-path)
        ts       (now-iso)

        ;; Run ast-grep for each category
        sym-rule  (lang/rules-dir language "symbols")
        imp-rule  (lang/rules-dir language "imports")
        call-rule (lang/rules-dir language "calls")
        ref-rule  (lang/rules-dir language "references")

        raw-syms  (run-sg sym-rule file-path)
        raw-imps  (run-sg imp-rule file-path)
        raw-calls (run-sg call-rule file-path)
        raw-refs  (run-sg ref-rule file-path)

        ;; Build symbol docs
        symbols (->> raw-syms
                     (mapv (fn [m]
                             (let [text       (:text m)
                                   start-line (get-in m [:range :start :line])
                                   end-line   (get-in m [:range :end :line])
                                   sym-name   (extract-name text)
                                   kind       (detect-kind text)
                                   sig        (first (str/split-lines text))
                                   sym-id     (str repo ":" rel-file ":" sym-name ":" start-line)]
                               {:id         sym-id
                                :file       rel-file
                                :kind       kind
                                :name       sym-name
                                :signature  sig
                                :language   language
                                :start_line start-line
                                :end_line   end-line
                                :parent_id  nil
                                :docstring  nil
                                :file_hash  hash
                                :repo       repo
                                :indexed_at ts})))
                     ;; Filter out nil names
                     (filterv :name))

        ;; Assign parent_id by line-range containment
        symbols (mapv (fn [s]
                        (let [parent (find-parent symbols (:start_line s) (:end_line s))]
                          (if parent
                            (assoc s :parent_id parent)
                            s)))
                      symbols)

        ;; Build contains edges from parent-child
        contains-edges (->> symbols
                            (filter :parent_id)
                            (mapv (fn [s]
                                    {:source (:parent_id s)
                                     :target (:id s)
                                     :kind   "contains"
                                     :file   rel-file
                                     :repo   repo})))

        ;; Collect raw import records
        imports (->> raw-imps
                     (mapv (fn [m]
                             {:text       (:text m)
                              :name       (extract-import-name (:text m))
                              :file       rel-file
                              :start_line (get-in m [:range :start :line])
                              :repo       repo})))

        ;; Collect raw call records
        calls (->> raw-calls
                   (mapv (fn [m]
                           {:text       (:text m)
                            :name       (extract-call-name (:text m))
                            :file       rel-file
                            :start_line (get-in m [:range :start :line])
                            :repo       repo})))

        ;; Collect raw reference records
        refs (->> raw-refs
                  (mapv (fn [m]
                          {:text       (:text m)
                           :name       (extract-call-name (:text m))
                           :file       rel-file
                           :start_line (get-in m [:range :start :line])
                           :repo       repo})))

        ;; Extract extends edges from symbol declarations
        extends-edges (->> symbols
                           (keep (fn [s]
                                   (when-let [parent-name (extract-extends (:signature s))]
                                     {:raw-source (:id s)
                                      :name       parent-name
                                      :file       rel-file
                                      :repo       repo})))
                           vec)

        ;; Extract implements edges from symbol declarations
        implements-edges (->> symbols
                              (mapcat (fn [s]
                                        (map (fn [iface-name]
                                               {:raw-source (:id s)
                                                :name       iface-name
                                                :file       rel-file
                                                :repo       repo})
                                             (extract-implements (:signature s)))))
                              vec)

        ;; Extract export edges for public symbols
        export-edges (->> symbols
                          (filter (fn [s] (exported? (:name s) language)))
                          (mapv (fn [s]
                                  {:source (:id s)
                                   :target (:id s)
                                   :kind   "exports"
                                   :file   rel-file
                                   :repo   repo})))]

    {:symbols          symbols
     :raw-imports      imports
     :raw-calls        calls
     :raw-refs         refs
     :raw-extends      extends-edges
     :raw-implements   implements-edges
     :contains-edges   contains-edges
     :export-edges     export-edges}))

;; ---------------------------------------------------------------------------
;; Cross-file resolution (Phase 2 + 3)
;; ---------------------------------------------------------------------------

(defn resolve-imports
  "Cross-file: match raw imports to known symbols, create import edges.
   symbol-table is {name -> [symbol-id ...]} built from all extracted symbols."
  [raw-imports symbol-table repo]
  (->> raw-imports
       (mapcat (fn [{:keys [name file]}]
                 (when-let [target-ids (get symbol-table name)]
                   (map (fn [tid]
                          {:source (str repo ":" file ":import:" name)
                           :target tid
                           :kind   "imports"
                           :file   file
                           :repo   repo})
                        target-ids))))
       vec))

(defn resolve-calls
  "Cross-file: match raw calls to known symbols, create call edges."
  [raw-calls symbol-table repo]
  (->> raw-calls
       (mapcat (fn [{:keys [name file start_line]}]
                 (when-let [target-ids (get symbol-table name)]
                   (map (fn [tid]
                          {:source (str repo ":" file ":call:" name ":" start_line)
                           :target tid
                           :kind   "calls"
                           :file   file
                           :repo   repo})
                        target-ids))))
       vec))

(defn resolve-references
  "Cross-file: match raw references to known symbols, create reference edges."
  [raw-refs symbol-table repo]
  (->> raw-refs
       (mapcat (fn [{:keys [name file start_line]}]
                 (when-let [target-ids (get symbol-table name)]
                   (map (fn [tid]
                          {:source (str repo ":" file ":ref:" name ":" start_line)
                           :target tid
                           :kind   "references"
                           :file   file
                           :repo   repo})
                        target-ids))))
       vec))

(defn resolve-extends
  "Cross-file: match raw extends to known symbols, create extends edges."
  [raw-extends symbol-table _repo]
  (->> raw-extends
       (mapcat (fn [{:keys [raw-source name file repo]}]
                 (when-let [target-ids (get symbol-table name)]
                   (map (fn [tid]
                          {:source raw-source
                           :target tid
                           :kind   "extends"
                           :file   file
                           :repo   repo})
                        target-ids))))
       vec))

(defn resolve-implements
  "Cross-file: match raw implements to known symbols, create implements edges."
  [raw-implements symbol-table _repo]
  (->> raw-implements
       (mapcat (fn [{:keys [raw-source name file repo]}]
                 (when-let [target-ids (get symbol-table name)]
                   (map (fn [tid]
                          {:source raw-source
                           :target tid
                           :kind   "implements"
                           :file   file
                           :repo   repo})
                        target-ids))))
       vec))

;; ---------------------------------------------------------------------------
;; Incremental check
;; ---------------------------------------------------------------------------

(defn stale-files
  "Given file paths, check which ones have changed since last index.
   Uses SHA256 hash comparison against ES."
  [es-url repo files]
  (if (empty? files)
    []
    (let [sym-idx   (symbols-index (repo-name repo))
          ;; Compute current hashes
          current   (into {} (map (fn [f] [f (file-hash f)]) files))
          ;; Query ES for existing file_hash values keyed by file
          rel-files (mapv (fn [f]
                            (if (str/starts-with? f repo)
                              (subs f (inc (count repo)))
                              f))
                          files)
          ;; Build a map of rel-file -> indexed hash
          indexed   (when (es/index-exists? es-url sym-idx)
                      (->> (es/search es-url sym-idx
                                      {:size 10000
                                       :query {:terms {:file rel-files}}
                                       :_source ["file" "file_hash"]
                                       :collapse {:field "file"}})
                           (into {} (map (fn [doc] [(:file doc) (:file_hash doc)])))))
          indexed   (or indexed {})]
      ;; Return files whose hash differs or is missing
      (filterv (fn [f]
                 (let [rel (if (str/starts-with? f repo)
                             (subs f (inc (count repo)))
                             f)
                       cur-hash  (get current f)
                       prev-hash (get indexed rel)]
                   (not= cur-hash prev-hash)))
               files))))

;; ---------------------------------------------------------------------------
;; Delete stale documents
;; ---------------------------------------------------------------------------

(defn- delete-by-files
  "Delete all documents for given files from an index."
  [es-url index rel-files]
  (when (and (seq rel-files) (es/index-exists? es-url index))
    (es/delete-by-query es-url index {:query {:terms {:file rel-files}}})))

;; ---------------------------------------------------------------------------
;; Main pipeline
;; ---------------------------------------------------------------------------

(defn index-repo
  "Index a repository. Three-phase pipeline.
   Options: :es-url, :languages, :full (skip hash check), :stats-only"
  [repo-path & {:keys [es-url languages full stats-only]
                :or {es-url (or (System/getenv "GLITCH_ES_URL") "http://localhost:9200")}}]
  (let [repo (repo-name repo-path)]
    (if stats-only
      ;; Just print counts from ES and return
      (let [sym-idx  (symbols-index repo)
            edge-idx (edges-index repo)]
        (log "=== Code Intelligence Stats for" repo "===")
        (if (es/index-exists? es-url sym-idx)
          (do
            (log "  Symbols:" (es/doc-count es-url sym-idx))
            (log "  Edges:  " (es/doc-count es-url edge-idx)))
          (log "  No index found. Run `glitch index` first."))
        {:symbols 0 :edges 0})

      ;; Full indexing pipeline
      (do
        (log "=== Indexing" repo "===")

        ;; 1. Ensure indices exist
        (let [sym-idx  (symbols-index repo)
              edge-idx (edges-index repo)]
          (es/create-index es-url sym-idx symbols-mappings)
          (es/create-index es-url edge-idx edges-mappings)

          ;; 2. Find source files
          (let [files (find-source-files repo-path :languages languages)]
            (log "  Found" (count files) "source files")

            (if (empty? files)
              (do (log "  No source files found.")
                  {:symbols 0 :edges 0})

              ;; 3. Filter to stale files (unless :full)
              (let [target-files (if full
                                  files
                                  (stale-files es-url repo-path files))]
                (log "  Stale files:" (count target-files)
                     (if full "(full reindex)" "(incremental)"))

                (if (empty? target-files)
                  (do (log "  Everything up to date.")
                      {:symbols 0 :edges 0})

                  ;; 4. Phase 1: extract symbols for each file
                  (do
                    (log "  Phase 1: Extracting symbols...")
                    (let [results (mapv (fn [f]
                                          (let [language (lang/detect-language f)]
                                            (log "    " f)
                                            (if (lang/native-language? language)
                                              (clj-index/extract-symbols
                                                repo-path f language
                                                (repo-name repo-path) (file-hash f) (now-iso))
                                              (extract-symbols repo-path f language))))
                                        target-files)

                          all-symbols    (into [] (mapcat :symbols) results)
                          all-imports    (into [] (mapcat :raw-imports) results)
                          all-calls      (into [] (mapcat :raw-calls) results)
                          all-refs       (into [] (mapcat :raw-refs) results)
                          all-contains   (into [] (mapcat :contains-edges) results)
                          all-exports    (into [] (mapcat :export-edges) results)
                          all-raw-ext    (into [] (mapcat :raw-extends) results)
                          all-raw-impl   (into [] (mapcat :raw-implements) results)]

                      (log "    Extracted" (count all-symbols) "symbols,"
                           (count all-imports) "imports,"
                           (count all-calls) "calls,"
                           (count all-refs) "references")

                      ;; 5. Build symbol table {name -> [id1 id2 ...]}
                      ;; Multiple symbols can share the same name across files
                      (let [symbol-table (reduce (fn [m s]
                                                   (update m (:name s) (fnil conj []) (:id s)))
                                                 {} all-symbols)]

                        ;; 6. Phase 2: resolve imports + extends + implements
                        (log "  Phase 2: Resolving imports, extends, implements...")
                        (let [import-edges     (resolve-imports all-imports symbol-table repo)
                              extends-edges    (resolve-extends all-raw-ext symbol-table repo)
                              implements-edges (resolve-implements all-raw-impl symbol-table repo)]

                          ;; 7. Phase 3: resolve calls + references
                          (log "  Phase 3: Resolving calls & references...")
                          (let [call-edges (resolve-calls all-calls symbol-table repo)
                                ref-edges  (resolve-references all-refs symbol-table repo)
                                all-edges  (into [] cat [all-contains all-exports import-edges
                                                         extends-edges implements-edges
                                                         call-edges ref-edges])]

                            ;; 8. Delete stale symbols/edges for changed files
                            (let [rel-files (mapv (fn [f]
                                                    (if (str/starts-with? f repo-path)
                                                      (subs f (inc (count repo-path)))
                                                      f))
                                                  target-files)]
                              (log "  Cleaning stale data for" (count rel-files) "files...")
                              (delete-by-files es-url sym-idx rel-files)
                              (delete-by-files es-url edge-idx rel-files))

                            ;; 9. Bulk index symbols + edges
                            (when (seq all-symbols)
                              (log "  Indexing" (count all-symbols) "symbols...")
                              (doseq [batch (partition-all 500 all-symbols)]
                                (es/bulk-index es-url sym-idx batch)))

                            (when (seq all-edges)
                              (log "  Indexing" (count all-edges) "edges...")
                              (doseq [batch (partition-all 500 all-edges)]
                                (es/bulk-index es-url edge-idx batch)))

                            ;; 10. Print summary
                            (log "  Done. Symbols:" (count all-symbols) "Edges:" (count all-edges))
                            {:symbols (count all-symbols)
                             :edges   (count all-edges)}))))))))))))))
;; ---------------------------------------------------------------------------
;; Query functions
;; ---------------------------------------------------------------------------

(defn query-symbols
  "Search symbols index. Options: :name, :kind, :language, :file, :limit"
  [es-url repo opts]
  (let [idx     (symbols-index repo)
        clauses (cond-> [{:term {:repo repo}}]
                  (:kind opts)     (conj {:term {:kind (:kind opts)}})
                  (:language opts) (conj {:term {:language (:language opts)}})
                  (:file opts)     (conj {:term {:file (:file opts)}})
                  (:name opts)     (conj (if (str/includes? (str (:name opts)) "*")
                                           {:wildcard {:name (:name opts)}}
                                           {:term {:name (:name opts)}})))
        query   {:size  (or (:limit opts) 20)
                 :query {:bool {:must clauses}}}]
    (es/search es-url idx query)))

(defn query-edges
  "Search edges index. Options: :source, :target, :kind, :depth, :limit"
  [es-url repo opts]
  (let [idx   (edges-index repo)
        depth (or (:depth opts) 1)
        limit (or (:limit opts) 50)]
    (if (<= depth 1)
      ;; Simple single-hop query
      (let [clauses (cond-> [{:term {:repo repo}}]
                      (:source opts) (conj {:term {:source (:source opts)}})
                      (:target opts) (conj {:term {:target (:target opts)}})
                      (:kind opts)   (conj {:term {:kind (:kind opts)}}))
            query   {:size limit :query {:bool {:must clauses}}}]
        (es/search es-url idx query))

      ;; BFS for depth > 1
      (loop [d         1
             collected []
             frontier  (cond
                         (:source opts) #{(:source opts)}
                         (:target opts) #{(:target opts)}
                         :else          #{})]
        (if (or (> d depth) (empty? frontier))
          collected
          (let [field   (if (:source opts) :source :target)
                clauses [{:term {:repo repo}}
                         {:terms {field (vec frontier)}}]
                clauses (if (:kind opts)
                          (conj clauses {:term {:kind (:kind opts)}})
                          clauses)
                query   {:size limit :query {:bool {:must clauses}}}
                results (es/search es-url idx query)
                new-ids (->> results
                             (map (if (= field :source) :target :source))
                             (remove (set (map (if (= field :source) :target :source) collected)))
                             set)]
            (recur (inc d)
                   (into collected results)
                   new-ids)))))))

(defn query-context
  "Get full context for a symbol: definition + all edges."
  [es-url repo name]
  (let [sym-idx  (symbols-index repo)
        edge-idx (edges-index repo)
        ;; 1. Search for symbol by name
        symbols  (es/search es-url sym-idx
                            {:size 1
                             :query {:bool {:must [{:term {:repo repo}}
                                                   {:term {:name name}}]}}})
        symbol   (first symbols)]
    (when symbol
      (let [sym-id (:id symbol)
            ;; 2. Query edges where source or target matches symbol id
            outgoing (es/search es-url edge-idx
                                {:size 1000
                                 :query {:bool {:must [{:term {:repo repo}}
                                                       {:term {:source sym-id}}]}}})
            incoming (es/search es-url edge-idx
                                {:size 1000
                                 :query {:bool {:must [{:term {:repo repo}}
                                                       {:term {:target sym-id}}]}}})]
        {:symbol     symbol
         :callers    (filterv #(= "calls" (:kind %)) incoming)
         :callees    (filterv #(= "calls" (:kind %)) outgoing)
         :parent     (when (:parent_id symbol)
                       (first (es/search es-url sym-idx
                                         {:size 1
                                          :query {:term {:id (:parent_id symbol)}}})))
         :children   (filterv #(= "contains" (:kind %)) outgoing)
         :imports    (filterv #(= "imports" (:kind %)) outgoing)
         :references (filterv #(= "references" (:kind %)) incoming)}))))
