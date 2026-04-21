(ns glitch.runner
  "Workflow runner — wires core + provider + store together and evaluates
   .glitch workflow files via SCI."
  (:require [glitch.core :as g]
            [glitch.graph :as graph]
            [glitch.store :as store]
            [glitch.provider :as prov]
            [glitch.tool_loop]
            [glitch.mcp.tools :as mcp-tools]
            [glitch.mcp.handlers :as mcp-handlers]
            [glitch.mcp.indexer :as indexer]
            [glitch.mcp.search :as search]
            [sci.core :as sci]
            [cheshire.core :as json]
            [clojure.string :as str]
            [clojure.java.io :as io]))

;; ---------------------------------------------------------------------------
;; SCI macro source — evaled into SCI context before each workflow run.
;; These mirror the macros in core.clj but are written as SCI-evaluable
;; source strings because host Clojure macros cannot be directly registered
;; as SCI macros in babashka.
;; ---------------------------------------------------------------------------

(def ^:private sci-macros
  "(defmacro workflow [name & body]
     (let [forms (loop [b (seq body) acc []]
                   (if (nil? b) acc
                     (if (keyword? (first b))
                       (recur (nnext b) acc)
                       (recur (next b) (conj acc (first b))))))]
       `(do
          (glitch.core/trace \">>\" ~name)
          (let [result# (do ~@forms)]
            {:name ~name
             :output (str result#)
             :steps (get-steps)}))))

   (defmacro par [& forms]
     (let [futs (mapv (fn [f] `(future-call (fn [] ~f))) forms)]
       `(mapv deref ~futs)))

   (defmacro retry [n & body]
     `(loop [attempts# ~n
             last-err# nil]
        (if (<= attempts# 0)
          (throw (ex-info (str \"retry exhausted after \" ~n \" attempts: \" last-err#) {}))
          (let [result# (try
                          {:ok (do ~@body)}
                          (catch Exception e#
                            {:err (.getMessage e#)}))]
            (if (:ok result#)
              (:ok result#)
              (recur (dec attempts#) (:err result#)))))))

   (defmacro with-timeout [seconds & body]
     `(let [f# (future-call (fn [] ~@body))
            result# (deref f# (* ~seconds 1000) ::timeout)]
        (when (= result# ::timeout)
          (future-cancel f#)
          (throw (ex-info (str \"timeout after \" ~seconds \"s\") {})))
        result#))

   (defmacro step [id body & opts]
     `(do
        (glitch.core/trace \"  >\" ~id)
        (let [t0# (System/currentTimeMillis)
              result# ~body
              ms# (- (System/currentTimeMillis) t0#)]
          (glitch.core/trace \"  ✓\" ~id (str \"(\" ms# \"ms)\"))
          (glitch.core/_step ~id result# ~@opts))))

   (defmacro phase [name & body]
     `(do
        (when-let [rec# @~(symbol \"glitch.core/*step-recorder*\")]
          (rec# {:step-id ~name :kind \"phase\" :output \"started\"}))
        (let [result# (do ~@body)]
          (when-let [rec# @~(symbol \"glitch.core/*step-recorder*\")]
            (rec# {:step-id ~name :kind \"phase\" :output (str result#)}))
          result#)))

   (defmacro investigate [goal-claim & body]
     `(let [g# (glitch.graph/make-graph)]
        (reset! ~(symbol \"glitch.core/*graph*\") g#)
        (glitch.graph/set-goal! g# ~goal-claim)
        (try
          (do ~@body)
          (finally
            (reset! ~(symbol \"glitch.core/*graph*\") nil)))
        (let [snapshot# (deref g#)]
          {:graph snapshot#
           :goal ~goal-claim
           :reachable (glitch.graph/reachable? g# \"goal\")
           :gap (glitch.graph/confidence-gap g# \"goal\")
           :stats (glitch.graph/graph-stats g#)
           :efficiency (glitch.graph/efficiency-metrics g#)})))")

;; ---------------------------------------------------------------------------
;; SCI context — pre-binds all core symbols so .glitch workflows can use
;; bare (step ...), (llm ...), etc.  SCI evaluates in the 'user namespace
;; by default, so bindings in 'user are available as bare symbols.
;; ---------------------------------------------------------------------------

;; Atom holding the current SCI ctx — needed so call-workflow can reuse it.
(def ^:private ^:dynamic *sci-ctx* nil)

;; Atom holding the search function — bound at run time when index is available.
(def ^:private *search-fn* (atom nil))

(defn- sci-call-workflow
  "SCI-compatible call-workflow: resolves the child workflow file and evals
   it with sci/eval-string* using the current SCI context.  Provides
   binding isolation identical to core/call-workflow."
  [name & {:keys [input set]}]
  (when (some #{name} @g/*call-stack*)
    (throw (ex-info (str "call-workflow cycle: " name " already on stack "
                         (str/join " -> " @g/*call-stack*))
                    {:name name :stack @g/*call-stack*})))
  (swap! g/*call-stack* conj name)
  (try
    (let [wf-dir      @g/*workflows-dir*
          path        (or (some #(when (.exists (io/file %)) %)
                            [(str wf-dir "/" name ".glitch")
                             (str wf-dir "/" name ".clj")])
                          (throw (ex-info (str "call-workflow: " name " not found in " wf-dir)
                                         {:name name :dir wf-dir})))
          saved-rec   @g/*step-recorder*
          saved-prov  @g/*provider-fn*]
      (binding [g/*steps*          (atom {})
                g/*step-order*     (atom [])
                g/*input*          (atom (or input ""))
                g/*params*         (atom (merge @g/*params* (or set {})))
                g/*step-recorder*  (atom saved-rec)
                g/*provider-fn*    (atom saved-prov)]
        (sci/eval-string* *sci-ctx* (slurp path))
        {:output (g/last-output) :steps @g/*steps*}))
    (finally
      (swap! g/*call-stack* (comp vec pop)))))

(defn make-sci-ctx
  "Build a SCI evaluation context with all glitch primitives pre-bound.
   The returned ctx is ready for sci/eval-string*."
  []
  (let [str-ns  (into {} (map (fn [[k v]] [(symbol (name k)) @v]))
                       (ns-publics 'clojure.string))
        user-ns {;; core primitives (step is a SCI macro, see sci-macros)
                 'ref            g/ref
                 'input          g/input
                 'params         g/params
                 'param          g/param
                 'sh             g/sh
                 'save           g/save
                 'read-file      g/read-file
                 'write-file     g/write-file
                 'get-steps      g/get-steps
                 'last-output    g/last-output
                 'llm            g/llm
                 'gate           g/gate
                 'call-workflow  sci-call-workflow
                 'json-extract   g/json-extract

                 ;; confidence framework
                 'validate        g/validate
                 'validate-schema g/validate-schema
                 'check-contract  g/check-contract
                 'grounded?       g/grounded?
                 'consensus       g/consensus
                 'composite-score g/composite-score

                 ;; investigation graph
                 'fact              graph/sci-fact
                 'fact-from         graph/sci-fact-from
                 'approve!          graph/sci-approve!
                 'corroborate-from  graph/sci-corroborate-from
                 'reachable?        graph/sci-reachable?
                 'graph-stats       graph/sci-graph-stats
                 'confidence-gap    graph/sci-confidence-gap
                 'suggest-next      graph/sci-suggest-next

                 ;; search (bound at run time via *search-fn*)
                 'search          (fn [query & {:keys [limit] :or {limit 10}}]
                                   (if-let [f @*search-fn*]
                                     (f query limit)
                                     (throw (ex-info "search not available — no index" {}))))

                 'mkdir-p        (fn [dir] (.mkdirs (io/file dir)))

                 ;; state setters (rarely needed in workflows, but available)
                 'set-input!        g/set-input!
                 'set-params!       g/set-params!
                 'set-provider-fn!  g/set-provider-fn!
                 'set-step-recorder! g/set-step-recorder!
                 'set-workflows-dir! g/set-workflows-dir!
                 'reset-steps!      g/reset!

                 ;; concurrency primitives (SCI doesn't expose these by default)
                 'future-call    future-call
                 'future-cancel  future-cancel
                 'deref          deref
                 'atom           atom
                 'swap!          swap!
                 'reset!         reset!

                 ;; common utils
                 'slurp          slurp
                 'spit           spit
                 'println        println
                 'prn            prn
                 'str            str
                 'pr-str         pr-str
                 'read-string    read-string
                 'ex-info        ex-info
                 'ex-message     ex-message
                 'ex-data        ex-data}
        ctx     (sci/init
                  {:namespaces
                   {'user user-ns
                    'str  str-ns
                    'json {'decode         json/parse-string
                           'encode         json/generate-string
                           'parse-string   json/parse-string
                           'generate-string json/generate-string}
                    'glitch.core
                    {'_step           g/_step
                     'step            g/_step
                     'trace           g/trace
                     '*step-recorder* g/*step-recorder*
                     '*provider-fn*   g/*provider-fn*
                     '*steps*         g/*steps*
                     '*step-order*    g/*step-order*
                     '*input*         g/*input*
                     '*params*        g/*params*
                     '*call-stack*    g/*call-stack*
                     '*workflows-dir* g/*workflows-dir*
                     '*graph*         g/*graph*}
                    'glitch.graph
                    {'make-graph         graph/make-graph
                     'set-goal!          graph/set-goal!
                     'add-fact!          graph/add-fact!
                     'add-edge!          graph/add-edge!
                     'contradict!        graph/contradict!
                     'corroborate!       graph/corroborate!
                     'approve!           graph/approve!
                     'shortest-path      graph/shortest-path
                     'confidence-gap     graph/confidence-gap
                     'reachable?         graph/reachable?
                     'suggest-next       graph/suggest-next
                     'graph-stats        graph/graph-stats
                     'efficiency-metrics graph/efficiency-metrics}}
                   :classes
                   {'System  System
                    'Math    Math
                    'Thread  Thread
                    'Exception Exception}})]
    ;; Eval macro definitions into the context
    (sci/eval-string* ctx sci-macros)
    ctx))

;; ---------------------------------------------------------------------------
;; run — the main entry point
;; ---------------------------------------------------------------------------

(defn run
  "Evaluate a .glitch workflow file.

   Required:
     workflow-path — path to the .glitch or .clj file

   Options (keyword args):
     :input          — string input available via (input)
     :db             — SQLite db path (from store/open); nil to skip recording
     :project        — project/workspace name for run record
     :model          — default model name for LLM calls
     :params         — map of parameters available via (params)/(param k)
     :seed-steps     — map of step-id->value to pre-populate before eval
     :workflows-dir  — directory for call-workflow resolution
     :parent-run-id  — parent run ID for sub-workflow recording
     :tiers          — provider tier list (overrides default-tiers)"
  [workflow-path & {:keys [input db project model params seed-steps
                           workflows-dir parent-run-id tiers]}]
  (let [input  (or input "")
        model  (or model "default")
        params (or params {})
        wf-dir (or workflows-dir
                   (let [f (io/file workflow-path)]
                     (if-let [parent (.getParent f)]
                       parent
                       ".")))]
    ;; Reset core state
    (g/reset!)
    (reset! *search-fn* nil)
    (g/set-input! input)
    (g/set-params! params)
    (g/set-workflows-dir! wf-dir)

    ;; Build tool handler for this workspace
    (let [workspace-path (System/getProperty "user.dir")
          _  (g/trace "  indexing" workspace-path)
          t0 (System/currentTimeMillis)
          search-db (try (indexer/open-search-db workspace-path) (catch Exception _ nil))
          _         (when search-db
                      (try (indexer/index-repo search-db workspace-path) (catch Exception _ nil))
                      (g/trace "  ✓ indexed" (str "(" (- (System/currentTimeMillis) t0) "ms)"))
                      (reset! *search-fn*
                        (fn [query limit]
                          (let [results (search/hybrid-search search-db query workspace-path :limit limit)]
                            (str/join "\n\n"
                              (map (fn [{:keys [path content score]}]
                                     (str "--- " path " (score: " (format "%.2f" (double score)) ") ---\n" content))
                                   results))))))
          tool-context {:workspace-path workspace-path
                        :search-db search-db
                        :embed-fn nil}
          tool-handler (mcp-handlers/make-handler tool-context)
          available-defs (if search-db
                           mcp-tools/tool-definitions
                           (let [search-tools #{"glitch_search" "glitch_index" "glitch_symbols"}]
                             (filterv #(not (contains? search-tools (get % "name")))
                                      mcp-tools/tool-definitions)))]

      ;; Wire provider dispatch
      (g/set-provider-fn!
        (fn [opts]
          (let [pname  (or (:provider opts) "lmstudio")
                ;; Inject tool-handler and default tool definitions unless :tools []
                tools-requested (get opts :tools :all)
                tool-defs (when-not (and (coll? tools-requested) (empty? tools-requested))
                            (if (= tools-requested :all)
                              available-defs
                              (filterv #(contains? (set tools-requested) (get % "name"))
                                       available-defs)))
                merged (assoc opts
                         :model (let [m (or (:model opts) model)]
                                  (when-not (= m "default") m))
                         :tool-defs tool-defs
                         :tool-handler tool-handler)]
            (if tiers
              (prov/call-tiered merged tiers)
              (if (:provider opts)
                ;; Explicit provider — fail loudly, no fallback
                (prov/call-provider pname merged)
                ;; No provider specified — try default, then tier escalation
                (try
                  (prov/call-provider pname merged)
                  (catch Exception _
                    (prov/call-tiered merged prov/default-tiers))))))))

      ;; Pre-seed steps
      (when seed-steps
        (doseq [[k v] seed-steps]
          (g/step k v)))

      ;; Record run in store (when db is provided)
      (let [run-id (when db
                     (let [rid (store/record-run db
                                 {:name          ""
                                  :input         input
                                  :workflow-file workflow-path
                                  :model         model
                                  :workspace     (or project "")
                                  :parent-run-id parent-run-id})]
                       ;; Wire step recorder to persist every step
                       (g/set-step-recorder!
                         (fn [rec]
                           (store/record-step db (assoc rec :run-id rid))))
                       rid))
            ;; Build SCI context and evaluate the workflow
            ctx (make-sci-ctx)]
        (binding [*sci-ctx* ctx]
          (try
            (sci/eval-string* ctx (slurp workflow-path))
            (let [result {:name    ""
                          :output  (g/last-output)
                          :steps   (g/get-steps)
                          :run-id  (or run-id 0)}]
              (when db
                (store/finish-run db run-id (:output result) 0))
              result)
            (catch Exception e
              (when db
                (store/finish-run db run-id (str "ERROR: " (.getMessage e)) 1))
              (throw e))))))))

;; ---------------------------------------------------------------------------
;; list-workflows — discover workflow files in a directory
;; ---------------------------------------------------------------------------

(defn list-workflows
  "Return a sorted vec of {:name :path} maps for .glitch and .clj files
   found in `dir`."
  [dir]
  (let [d (io/file dir)]
    (when (.exists d)
      (->> (.listFiles d)
           (filter #(let [n (.getName %)]
                      (or (str/ends-with? n ".glitch")
                          (str/ends-with? n ".clj"))))
           (map (fn [f]
                  (let [n       (.getName f)
                        ext-len (if (str/ends-with? n ".glitch") 7 4)]
                    {:name (subs n 0 (- (count n) ext-len))
                     :path (.getPath f)})))
           (sort-by :name)
           vec))))
