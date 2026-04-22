(ns glitch.index.clojure
  "Clojure/Babashka symbol extraction via bb-native s-expression parsing.
   ast-grep has no Clojure grammar, so we read forms directly."
  (:require [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Clojure-specific kind/name detection
;; ---------------------------------------------------------------------------

(def ^:private defn-forms
  #{'defn 'defn- 'defmulti 'defmethod})

(def ^:private def-forms
  #{'def 'defonce})

(def ^:private macro-forms
  #{'defmacro})

(def ^:private type-forms
  #{'defprotocol 'definterface})

(def ^:private record-forms
  #{'defrecord 'deftype})

(def ^:private ns-form
  #{'ns})

(defn- classify-form
  "Classify a top-level form. Returns {:kind :name :signature :docstring} or nil."
  [form]
  (when (and (list? form) (symbol? (first form)))
    (let [head (first form)
          args (rest form)]
      (cond
        (contains? defn-forms head)
        (let [nm   (first args)
              rest (next args)
              doc  (when (string? (first rest)) (first rest))
              sig  (str "(" head " " nm
                        (when-let [argv (first (filter vector? args))]
                          (str " " argv))
                        ")")]
          {:kind      "function"
           :name      (str nm)
           :signature sig
           :docstring doc
           :exported  (not= head 'defn-)})

        (contains? macro-forms head)
        (let [nm  (first args)
              rest (next args)
              doc (when (string? (first rest)) (first rest))]
          {:kind      "macro"
           :name      (str nm)
           :signature (str "(" head " " nm ")")
           :docstring doc
           :exported  true})

        (contains? def-forms head)
        (let [nm  (first args)
              rest (next args)
              doc (when (string? (first rest)) (first rest))]
          {:kind      "var"
           :name      (str nm)
           :signature (str "(" head " " nm ")")
           :docstring doc
           :exported  true})

        (contains? type-forms head)
        (let [nm  (first args)
              rest (next args)
              doc (when (string? (first rest)) (first rest))]
          {:kind      (if (= head 'defprotocol) "protocol" "interface")
           :name      (str nm)
           :signature (str "(" head " " nm ")")
           :docstring doc
           :exported  true})

        (contains? record-forms head)
        (let [nm  (first args)]
          {:kind      (if (= head 'defrecord) "record" "type")
           :name      (str nm)
           :signature (str "(" head " " nm
                           (when-let [fields (second args)]
                             (when (vector? fields) (str " " fields)))
                           ")")
           :docstring nil
           :exported  true})

        (contains? ns-form head)
        (let [nm (first args)]
          {:kind      "namespace"
           :name      (str nm)
           :signature (str "(ns " nm ")")
           :docstring nil
           :exported  true})

        :else nil))))

;; ---------------------------------------------------------------------------
;; Require/import extraction
;; ---------------------------------------------------------------------------

(defn- extract-ns-requires
  "Extract :require entries from an ns form. Returns seq of namespace strings."
  [form]
  (when (and (list? form) (= 'ns (first form)))
    (let [clauses (drop 2 form)] ;; skip ns-name and optional docstring
      (->> clauses
           (filter #(and (list? %) (= :require (first %))))
           (mapcat rest)
           (map (fn [r]
                  (cond
                    (vector? r) (str (first r))
                    (symbol? r) (str r)
                    :else nil)))
           (remove nil?)))))

;; ---------------------------------------------------------------------------
;; Call extraction — find function invocations in form bodies
;; ---------------------------------------------------------------------------

(defn- extract-calls-from-body
  "Walk a form tree collecting function call names (first element of lists)."
  [form]
  (cond
    (list? form)
    (let [head (first form)
          head-calls (if (symbol? head)
                       [(str head)]
                       (extract-calls-from-body head))
          rest-calls (mapcat extract-calls-from-body (rest form))]
      (concat head-calls rest-calls))

    (vector? form)
    (mapcat extract-calls-from-body form)

    (map? form)
    (mapcat extract-calls-from-body (concat (keys form) (vals form)))

    (set? form)
    (mapcat extract-calls-from-body form)

    :else []))

(defn- extract-defn-calls
  "Extract function calls from the body of a defn form."
  [form]
  (when (and (list? form) (symbol? (first form)))
    (let [head (first form)]
      (when (contains? (into defn-forms macro-forms) head)
        (let [;; Skip name, optional docstring, optional metadata, argvec
              body (->> (rest form)
                        (drop-while #(or (symbol? %) (string? %) (map? %)))
                        (drop-while vector?))]
          (->> (mapcat extract-calls-from-body body)
               ;; Filter out special forms and core forms we don't want to index
               (remove #(contains? #{"if" "let" "do" "when" "cond" "case" "fn"
                                      "loop" "recur" "try" "catch" "throw"
                                      "quote" "var" "." ".." "new" "set!"
                                      "monitor-enter" "monitor-exit"
                                      "def" "defn" "defn-" "defmacro"
                                      "str" "println" "prn" "pr-str"
                                      "first" "rest" "next" "cons" "conj"
                                      "map" "filter" "reduce" "mapv" "filterv"
                                      "into" "assoc" "dissoc" "get" "get-in"
                                      "update" "merge" "select-keys" "keys" "vals"
                                      "apply" "comp" "partial" "identity"
                                      "not" "nil?" "some?" "true?" "false?"
                                      "=" "==" "not=" "<" ">" "<=" ">="
                                      "+" "-" "*" "/" "mod" "rem" "inc" "dec"
                                      "and" "or" "some" "every?"
                                      "count" "empty?" "seq" "vec" "set"
                                      "contains?" "nth" "last" "butlast"
                                      "concat" "flatten" "distinct" "sort"
                                      "sort-by" "group-by" "partition"
                                      "partition-all" "partition-by"
                                      "take" "drop" "take-while" "drop-while"
                                      "subs" "re-find" "re-matches" "re-seq"
                                      "format" "name" "keyword" "symbol"
                                      "atom" "deref" "reset!" "swap!"
                                      "doseq" "dotimes" "for" "while"
                                      "binding" "with-open" "with-out-str"
                                      "->" "->>" "as->" "cond->" "cond->>"
                                      "some->" "some->>" "doto"
                                      "require" "import" "use" "refer"} %))
               distinct
               vec))))))

;; ---------------------------------------------------------------------------
;; Public API — same return shape as index/extract-symbols
;; ---------------------------------------------------------------------------

(defn extract-symbols
  "Extract symbols from a Clojure file using bb-native reading.
   Returns same shape as glitch.index/extract-symbols."
  [repo-path file-path _language repo-name file-hash now-iso]
  (let [content  (slurp file-path)
        rel-file (if (str/starts-with? file-path repo-path)
                   (subs file-path (inc (count repo-path)))
                   file-path)
        repo     repo-name
        hash     file-hash
        ts       now-iso

        ;; Read all top-level forms
        rdr   (java.io.PushbackReader. (java.io.StringReader. content))
        forms (loop [acc [] line-reader (java.io.LineNumberReader.
                                          (java.io.StringReader. content))]
                (let [form (try
                             (let [;; Use edamame-style reading via bb reader
                                   lnr (java.io.LineNumberReader.
                                         (java.io.StringReader. content))
                                   _ (dotimes [_ (count acc)]
                                       ;; re-read to advance past already-read forms
                                       nil)]
                               ;; Actually just read sequentially from the pushback reader
                               (let [f (read {:eof ::eof :read-cond :allow} rdr)]
                                 (when-not (= f ::eof) f)))
                             (catch Exception _e nil))]
                  (if (nil? form)
                    acc
                    (recur (conj acc form) line-reader))))

        ;; Match forms to line numbers by scanning content
        form-lines (loop [remaining forms
                          offset    0
                          result    []]
                     (if (empty? remaining)
                       result
                       (let [form     (first remaining)
                             form-str (pr-str form)
                             ;; Search for a recognizable prefix instead of exact match
                             prefix   (let [s (pr-str form)]
                                        (subs s 0 (min 40 (count s))))
                             idx      (str/index-of content prefix offset)
                             line     (if idx
                                        (count (filter #(= % \newline) (subs content 0 idx)))
                                        0)
                             ;; Estimate end line from form string
                             end-line (+ line (count (filter #(= % \newline) form-str)))]
                         (recur (rest remaining)
                                (if idx (+ idx (count prefix)) offset)
                                (conj result {:form form :start line :end end-line})))))

        ;; Classify forms into symbols
        symbols (->> form-lines
                     (keep (fn [{:keys [form start end]}]
                             (when-let [info (classify-form form)]
                               (let [sym-id (str repo ":" rel-file ":" (:name info) ":" start)]
                                 {:id         sym-id
                                  :file       rel-file
                                  :kind       (:kind info)
                                  :name       (:name info)
                                  :signature  (:signature info)
                                  :language   "clojure"
                                  :start_line start
                                  :end_line   end
                                  :parent_id  nil
                                  :docstring  (:docstring info)
                                  :file_hash  hash
                                  :repo       repo
                                  :indexed_at ts}))))
                     vec)

        ;; Extract requires from ns form
        ns-form (first (filter #(and (list? (:form %)) (= 'ns (first (:form %)))) form-lines))
        imports (if ns-form
                  (->> (extract-ns-requires (:form ns-form))
                       (mapv (fn [ns-name]
                               {:text       (str "(require " ns-name ")")
                                :name       (last (str/split ns-name #"\."))
                                :file       rel-file
                                :start_line (:start ns-form)
                                :repo       repo})))
                  [])

        ;; Extract calls from defn bodies
        calls (->> form-lines
                   (mapcat (fn [{:keys [form start]}]
                             (when-let [call-names (extract-defn-calls form)]
                               (map (fn [cn]
                                      {:text       cn
                                       :name       (last (str/split cn #"/"))
                                       :file       rel-file
                                       :start_line start
                                       :repo       repo})
                                    call-names))))
                   vec)

        ;; Export edges for public symbols
        export-edges (->> symbols
                          (filter (fn [s]
                                    (some (fn [{:keys [form]}]
                                            (when-let [info (classify-form form)]
                                              (and (= (:name info) (:name s))
                                                   (:exported info))))
                                          form-lines)))
                          (mapv (fn [s]
                                  {:source (:id s)
                                   :target (:id s)
                                   :kind   "exports"
                                   :file   rel-file
                                   :repo   repo})))]

    {:symbols        symbols
     :raw-imports    imports
     :raw-calls      calls
     :raw-refs       []
     :raw-extends    []
     :raw-implements []
     :contains-edges []
     :export-edges   export-edges}))
