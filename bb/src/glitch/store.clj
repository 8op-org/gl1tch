(ns glitch.store
  "In-memory run/step/fact store backed by a single EDN file.
   No pods, no SQLite, no external dependencies."
  (:require [clojure.java.io :as io]
            [clojure.edn :as edn]))

;; ---------------------------------------------------------------------------
;; State shape
;;
;; {:counter  <int>            ;; monotonic id generator
;;  :runs     {id -> run-map}  ;; run/id -> run data
;;  :steps    [[run-id step-id] -> step-map]
;;  :facts    {fact-id -> fact-map}
;;  :edges    [edge-map ...]}
;; ---------------------------------------------------------------------------

(defn- empty-state []
  {:counter 0
   :runs    {}
   :steps   {}
   :facts   {}
   :edges   []})

(defn- flush-state! [{:keys [state path]}]
  (let [parent (.getParentFile (io/file path))]
    (when (and parent (not (.exists parent)))
      (.mkdirs parent)))
  (spit path (pr-str @state)))

(defn- load-state [path]
  (when (.exists (io/file path))
    (let [raw (slurp path)]
      (when (seq (clojure.string/trim raw))
        (edn/read-string raw)))))

;; ---------------------------------------------------------------------------
;; Open / Close
;; ---------------------------------------------------------------------------

(defn open
  "Open or create a glitch store at `path`.
   Default: ~/.local/share/glitch/glitch.edn.
   Returns a store map with :state atom and :path."
  [& [path]]
  (let [path (or path
                 (str (System/getProperty "user.home")
                      "/.local/share/glitch/glitch.edn"))
        parent (.getParentFile (io/file path))]
    (when (and parent (not (.exists parent)))
      (.mkdirs parent))
    (let [state (atom (or (load-state path) (empty-state)))]
      {:state state :conn state :path path})))

(defn close
  "Flush the store to disk."
  [store]
  (when store (flush-state! store)))

;; ---------------------------------------------------------------------------
;; Runs
;; ---------------------------------------------------------------------------

(defn record-run
  "Insert a run record. `rec` is a map with keys:
     :name :input :workflow-file :model :parent-run-id
   Returns the integer ID of the new run."
  [store rec]
  (let [now (quot (System/currentTimeMillis) 1000)
        id  (-> (swap! (:state store) update :counter inc)
                :counter)
        run (cond-> {:run/id         id
                     :run/name       (or (:name rec) "")
                     :run/started-at now}
              (:input rec)         (assoc :run/input (:input rec))
              (:workflow-file rec) (assoc :run/workflow (:workflow-file rec))
              (:model rec)         (assoc :run/model (:model rec))
              (:parent-run-id rec) (assoc :run/parent-id (:parent-run-id rec)))]
    (swap! (:state store) assoc-in [:runs id] run)
    id))

(defn finish-run
  "Update a run with final output, exit status, and optional totals map
   with keys :tokens-in :tokens-out :cost."
  [store id output exit-status & [totals]]
  (let [now (quot (System/currentTimeMillis) 1000)]
    (swap! (:state store) update-in [:runs id]
           (fn [run]
             (cond-> (assoc run
                            :run/output      output
                            :run/exit        exit-status
                            :run/finished-at now)
               (:tokens-in totals)  (assoc :run/tokens-in (:tokens-in totals))
               (:tokens-out totals) (assoc :run/tokens-out (:tokens-out totals))
               (:cost totals)       (assoc :run/cost (:cost totals)))))
    (flush-state! store)))

(defn get-run
  "Fetch a single run by ID. Returns a map or nil."
  [store id]
  (get-in @(:state store) [:runs id]))

(defn list-runs
  "List runs with optional filters.
   Options: :parent-id, :workflow, :limit (default 50)."
  [store & {:keys [parent-id workflow limit] :or {limit 50}}]
  (let [runs (vals (get @(:state store) :runs {}))]
    (->> runs
         (filter (cond
                   parent-id #(= parent-id (:run/parent-id %))
                   workflow   #(= workflow (:run/workflow %))
                   :else      (constantly true)))
         (sort-by :run/id #(compare %2 %1))
         (take limit)
         vec)))

;; ---------------------------------------------------------------------------
;; Steps
;; ---------------------------------------------------------------------------

(defn record-step
  "Insert or replace a step record. `rec` is a map with keys:
     :run-id :step-id :prompt :output :model :duration :kind
     :exit :tokens-in :tokens-out :gate :artifacts :confidence"
  [store rec]
  (let [key [(:run-id rec) (:step-id rec)]
        step (cond-> {:step/run-id (:run-id rec)
                      :step/id     (:step-id rec)}
               (:prompt rec)             (assoc :step/prompt (:prompt rec))
               (:output rec)             (assoc :step/output (:output rec))
               (:model rec)              (assoc :step/model (:model rec))
               (:duration rec)           (assoc :step/duration (:duration rec))
               (:kind rec)               (assoc :step/kind (:kind rec))
               (:exit rec)               (assoc :step/exit (:exit rec))
               (:tokens-in rec)          (assoc :step/tokens-in (:tokens-in rec))
               (:tokens-out rec)         (assoc :step/tokens-out (:tokens-out rec))
               (:gate rec)               (assoc :step/gate (:gate rec))
               (:artifacts rec)          (assoc :step/artifacts (:artifacts rec))
               (contains? rec :confidence) (assoc :step/confidence (:confidence rec)))]
    (swap! (:state store) assoc-in [:steps key] step)))

(defn get-steps
  "Fetch all steps for a run, ordered by insertion."
  [store run-id]
  (->> (get @(:state store) :steps {})
       (filter (fn [[[rid _] _]] (= rid run-id)))
       (sort-by (fn [[[_ sid] _]] sid))
       (mapv second)))

;; ---------------------------------------------------------------------------
;; Facts
;; ---------------------------------------------------------------------------

(defn record-fact
  "Insert or replace a fact record. `rec` is a map with keys:
     :id :run-id :claim :confidence :status :source-step :source-prov :tokens-cost"
  [store rec]
  (let [now (quot (System/currentTimeMillis) 1000)
        fact {:fact/id          (:id rec)
              :fact/run-id      (:run-id rec)
              :fact/claim       (:claim rec)
              :fact/confidence  (:confidence rec)
              :fact/status      (or (some-> (:status rec) name) "unapproved")
              :fact/source-step (:source-step rec)
              :fact/source-prov (:source-prov rec)
              :fact/tokens-cost (or (:tokens-cost rec) 0)
              :fact/created-at  now}]
    (swap! (:state store) assoc-in [:facts (:id rec)] fact)))

(defn record-fact-edge
  "Insert a fact edge record. `rec` is a map with keys:
     :run-id :from-id :to-id :rel :weight :source"
  [store rec]
  (let [edge {:edge/run-id  (:run-id rec)
              :edge/from-id (:from-id rec)
              :edge/to-id   (:to-id rec)
              :edge/rel     (or (some-> (:rel rec) name) (str (:rel rec)))
              :edge/weight  (or (:weight rec) 1.0)
              :edge/source  (:source rec)}]
    (swap! (:state store) update :edges (fnil conj []) edge)))

(defn get-facts
  "Fetch all facts for a run, ordered by created-at."
  [store run-id]
  (->> (vals (get @(:state store) :facts {}))
       (filter #(= run-id (:fact/run-id %)))
       (sort-by :fact/created-at)
       vec))

(defn get-fact-edges
  "Fetch all fact edges for a run."
  [store run-id]
  (->> (get @(:state store) :edges [])
       (filter #(= run-id (:edge/run-id %)))
       vec))
