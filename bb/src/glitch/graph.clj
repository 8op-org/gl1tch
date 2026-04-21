(ns glitch.graph
  "Investigation fact graph — live reasoning graph for structured
   investigation workflows. Tracks facts, contradictions, corroboration,
   and computes shortest paths to investigation goals."
  (:require [clojure.string :as str]
            [glitch.core :as core]))

;; ---------------------------------------------------------------------------
;; Helpers — contradiction detection
;; ---------------------------------------------------------------------------

(def ^:private negation-words
  #{"not" "no" "never" "neither" "nor" "none" "nothing"
    "isn't" "aren't" "wasn't" "weren't" "won't" "don't"
    "doesn't" "didn't" "can't" "cannot" "shouldn't" "wouldn't"
    "couldn't" "hasn't" "haven't" "hadn't" "mustn't"
    "false" "unsafe" "fail" "wrong" "incorrect" "missing" "absent" "broken"})

(defn- tokenize
  "Simple whitespace + punctuation tokenizer."
  [s]
  (when s
    (-> (str/lower-case (str s))
        (str/replace #"[^\w\s']" " ")
        (str/split #"\s+")
        (->> (remove str/blank?)
             vec))))

(defn- keyword-overlap
  "Return the set of shared non-stopword tokens between two strings."
  [a b]
  (let [stop #{"the" "a" "an" "is" "are" "was" "were" "be" "been"
               "being" "have" "has" "had" "do" "does" "did" "will"
               "would" "could" "should" "may" "might" "shall" "can"
               "this" "that" "these" "those" "it" "its" "of" "in"
               "on" "at" "to" "for" "with" "by" "from" "as"}
        ta (set (remove stop (tokenize a)))
        tb (set (remove stop (tokenize b)))]
    (clojure.set/intersection ta tb)))

(defn- has-negation?
  "Check if a string contains negation words."
  [s]
  (let [tokens (set (tokenize s))]
    (some negation-words tokens)))

(defn detect-contradictions
  "Detect if new-fact contradicts any existing facts.
   Returns a seq of existing fact IDs that contradict, or empty seq.
   Uses keyword overlap + polarity flip detection.
   Falls back to keyword-based detection (no embedding required)."
  [new-fact existing-facts]
  (let [new-claim (:claim new-fact)]
    (->> existing-facts
         (filter (fn [[_id existing]]
                   (let [existing-claim (:claim existing)
                         overlap (keyword-overlap new-claim existing-claim)]
                     ;; Need significant overlap (same topic) + polarity flip
                     (and (>= (count overlap) 3)
                          (not= (boolean (has-negation? new-claim))
                                (boolean (has-negation? existing-claim)))))))
         (map first))))

;; ---------------------------------------------------------------------------
;; Forward declarations
;; ---------------------------------------------------------------------------

(declare contradict!)

;; ---------------------------------------------------------------------------
;; Graph construction
;; ---------------------------------------------------------------------------

(defn make-graph
  "Create an empty investigation graph atom."
  []
  (atom {:facts {}
         :edges []
         :goal  nil
         :stats {:tokens-spent 0
                 :time-spent-ms 0
                 :facts-added 0
                 :contradictions-found 0
                 :corroborations 0}}))

(defn set-goal!
  "Set the investigation goal. Creates a goal fact node."
  [graph goal-claim]
  (let [goal-fact {:id "goal"
                   :claim goal-claim
                   :confidence 0.0
                   :status :goal
                   :source-step nil
                   :source-prov nil
                   :tokens-cost 0
                   :created-at (System/currentTimeMillis)}]
    (swap! graph assoc-in [:facts "goal"] goal-fact)
    (swap! graph assoc :goal goal-claim)))

(defn add-fact!
  "Insert a fact into the graph. Checks for contradictions and enforces
   the 0.70 single-source confidence cap. Applies geometric decay (0.7^n)
   BEFORE insertion when contradictions are detected.
   fact map: {:id :claim :confidence :source-step :source-prov :tokens-cost}"
  [graph fact]
  (let [capped-conf (min (:confidence fact 0.5) 0.70)
        existing (:facts @graph)
        base-fact (assoc fact
                    :confidence capped-conf
                    :status (or (:status fact) :unapproved)
                    :created-at (or (:created-at fact) (System/currentTimeMillis)))
        contradicting (detect-contradictions base-fact existing)
        ;; Apply geometric decay BEFORE insertion
        decayed-conf (* capped-conf (Math/pow 0.7 (count contradicting)))
        final-fact (assoc base-fact :confidence decayed-conf)]
    ;; Add fact with already-decayed confidence
    (swap! graph assoc-in [:facts (:id final-fact)] final-fact)
    (swap! graph update-in [:stats :facts-added] inc)
    ;; Handle contradictions (reduces existing facts + adds edges + sets status)
    (doseq [contra-id contradicting]
      (contradict! graph (:id final-fact) contra-id))
    final-fact))

(defn add-edge!
  "Link two facts with a relationship edge.
   edge map: {:from :to :rel :weight :source}"
  [graph edge]
  (let [edge (assoc edge :weight (or (:weight edge) 1.0))]
    (swap! graph update :edges conj edge)
    edge))

(defn contradict!
  "Mark a contradiction between two facts. Reduces confidence on both by 0.7x.
   Sets status to :contradicted. Adds a :contradicts edge."
  [graph id-a id-b]
  (swap! graph (fn [g]
                 (-> g
                     (update-in [:facts id-a :confidence] * 0.7)
                     (update-in [:facts id-b :confidence] * 0.7)
                     (assoc-in [:facts id-a :status] :contradicted)
                     (assoc-in [:facts id-b :status] :contradicted)
                     (update :edges conj {:from id-a :to id-b
                                          :rel :contradicts :weight 0.0
                                          :source "contradiction-detector"})
                     (update-in [:stats :contradictions-found] inc)))))

(defn corroborate!
  "Corroborate a fact with evidence from a new source.
   Uses Bayesian accumulation to break the 0.70 single-source cap.
   Formula: new-conf = 1 - (1 - old-conf) * (1 - src-auth * 0.85)"
  [graph fact-id new-source]
  (let [fact (get-in @graph [:facts fact-id])
        old-conf (:confidence fact)
        src-auth (or (:authority new-source) 0.85)
        new-conf (- 1.0 (* (- 1.0 old-conf) (- 1.0 (* src-auth 0.85))))]
    (swap! graph assoc-in [:facts fact-id :confidence] new-conf)
    (swap! graph update :edges conj
           {:from fact-id :to fact-id :rel :corroborates
            :weight new-conf :source (:step new-source)})
    (swap! graph update-in [:stats :corroborations] inc)))

(defn approve!
  "Mark a fact as :approved with confidence 1.0."
  [graph fact-id]
  (swap! graph (fn [g]
                 (-> g
                     (assoc-in [:facts fact-id :status] :approved)
                     (assoc-in [:facts fact-id :confidence] 1.0)))))

;; ---------------------------------------------------------------------------
;; Pathfinding — Dijkstra
;; ---------------------------------------------------------------------------

(defn- build-adjacency
  "Build adjacency map from edges, excluding :contradicts edges."
  [edges]
  (reduce (fn [adj {:keys [from to rel] :as edge}]
            (if (= rel :contradicts)
              adj
              (update adj from (fnil conj []) edge)))
          {}
          edges))

(defn shortest-path
  "Dijkstra from any :approved fact to goal-id.
   Edge cost = 1 - confidence of target node.
   Returns {:path [...ids] :cost total-cost} or nil if unreachable."
  [graph goal-id]
  (let [g @graph
        facts (:facts g)
        edges (:edges g)
        adj (build-adjacency edges)
        ;; Start from all approved facts
        start-ids (->> facts
                       (filter (fn [[_ f]] (= :approved (:status f))))
                       (map first)
                       set)]
    (when (and (seq start-ids) (contains? facts goal-id))
      ;; Dijkstra — also track the edge used to reach each node
      (loop [dist (into {} (map (fn [id] [id 0.0]) start-ids))
             prev (into {} (map (fn [id] [id nil]) start-ids))
             prev-edge (into {} (map (fn [id] [id nil]) start-ids))
             queue (into (sorted-set-by
                           (fn [a b]
                             (let [c (compare (first a) (first b))]
                               (if (zero? c) (compare (second a) (second b)) c))))
                         (map (fn [id] [0.0 id]) start-ids))
             visited #{}]
        (if (empty? queue)
          ;; Check if we reached the goal
          (when (contains? dist goal-id)
            (let [path (loop [id goal-id acc []]
                         (if (nil? id)
                           (reverse acc)
                           (recur (get prev id) (conj acc id))))
                  path-vec (vec path)
                  ;; Collect edge weights along the path
                  edge-weights (loop [id goal-id ws []]
                                 (if-let [ew (get prev-edge id)]
                                   (recur (get prev id) (conj ws ew))
                                   ws))
                  min-weight (if (seq edge-weights)
                               (apply min edge-weights)
                               nil)]
              {:path path-vec
               :cost (get dist goal-id)
               :min-edge min-weight}))
          (let [[cost-u u] (first queue)
                queue (disj queue (first queue))]
            (if (contains? visited u)
              (recur dist prev prev-edge queue visited)
              (let [visited (conj visited u)
                    neighbors (get adj u [])
                    [dist' prev' prev-edge' queue']
                    (reduce
                      (fn [[d p pe q] {:keys [to weight] :as edge}]
                        (if (contains? visited to)
                          [d p pe q]
                          (let [w (max (or weight 0.01) 0.01)
                                edge-cost (- 1.0 w)
                                alt (+ cost-u edge-cost)]
                            (if (< alt (get d to Double/MAX_VALUE))
                              [(assoc d to alt)
                               (assoc p to u)
                               (assoc pe to w)
                               (conj q [alt to])]
                              [d p pe q]))))
                      [dist prev prev-edge queue]
                      neighbors)]
                (recur dist' prev' prev-edge' queue' visited)))))))))

(defn confidence-gap
  "Find the weakest link on the best path from approved facts to goal.
   Returns {:fact-id :confidence :claim :path-cost} or nil if no path."
  [graph goal-id]
  (when-let [{:keys [path cost]} (shortest-path graph goal-id)]
    (let [facts (:facts @graph)
          path-facts (->> path
                          (map (fn [id] (assoc (get facts id) :id id)))
                          (remove #(= :approved (:status %)))
                          (remove #(= :goal (:status %))))]
      (when (seq path-facts)
        (let [weakest (apply min-key :confidence path-facts)]
          {:fact-id (:id weakest)
           :confidence (:confidence weakest)
           :claim (:claim weakest)
           :path-cost cost})))))

(defn reachable?
  "Boolean: can the goal be reached from any approved fact?"
  [graph goal-id]
  (some? (shortest-path graph goal-id)))

(defn suggest-next
  "Suggest what to investigate next. Priority:
   1. Resolve contradictions (highest priority)
   2. Strengthen weakest link on critical path
   3. Explore disconnected facts"
  [graph goal-id]
  (let [g @graph
        facts (:facts g)
        edges (:edges g)
        ;; 1. Find contradicted facts
        contradicted-facts (->> (vals facts)
                                (filter #(= :contradicted (:status %)))
                                vec)]
    (if (seq contradicted-facts)
      {:action :resolve-contradiction
       :facts (mapv #(select-keys % [:id :claim :confidence]) contradicted-facts)}
      ;; 2. Weakest link
      (if-let [gap (confidence-gap graph goal-id)]
        (let [weakest gap]
          {:action :strengthen
           :fact (select-keys weakest [:fact-id :claim :confidence])
           :suggestion (str "Corroborate: \"" (:claim weakest) "\" — current confidence "
                            (format "%.2f" (double (:confidence weakest))))})
        ;; 3. Explore — no path to goal
        {:action :explore
         :suggestion "No path to goal exists. Discover connecting facts."}))))

;; ---------------------------------------------------------------------------
;; Stats and metrics
;; ---------------------------------------------------------------------------

(defn graph-stats
  "Return graph statistics."
  [graph]
  (let [g @graph
        facts (:facts g)
        edges (:edges g)]
    {:node-count (count facts)
     :edge-count (count edges)
     :approved-count (count (filter (fn [[_ f]] (= :approved (:status f))) facts))
     :unapproved-count (count (filter (fn [[_ f]] (= :unapproved (:status f))) facts))
     :contradiction-count (count (filter #(= :contradicts (:rel %)) edges))
     :goal (:goal g)
     :reachable (when (:goal g) (reachable? graph "goal"))
     :stats (:stats g)}))

(defn efficiency-metrics
  "Compute confidence-per-token and confidence-per-time ratios."
  [graph]
  (let [g @graph
        stats (:stats g)
        facts (:facts g)
        total-conf (reduce + 0.0 (map :confidence (vals facts)))
        tokens (:tokens-spent stats 0)
        time-ms (:time-spent-ms stats 0)]
    {:total-facts       (count facts)
     :avg-confidence    (if (pos? (count facts)) (/ total-conf (count facts)) 0.0)
     :tokens-spent      tokens
     :time-spent-ms     time-ms
     :conf-per-token    (if (pos? tokens) (/ total-conf tokens) 0.0)
     :conf-per-time-ms  (if (pos? time-ms) (/ total-conf time-ms) 0.0)
     :facts-per-1k-tok  (if (pos? tokens) (* 1000.0 (/ (double (count facts)) tokens)) 0.0)}))

;; ---------------------------------------------------------------------------
;; SCI helper functions — operate on bound *graph* from core
;; ---------------------------------------------------------------------------

(defn- current-graph
  "Get the current investigation graph from glitch.core/*graph*.
   *graph* is (atom <graph-atom-or-nil>), so we deref to get the inner value."
  []
  @core/*graph*)

(defn sci-fact
  "Create a fact and add to the current graph.
   Usage: (fact \"claim text\" :confidence 0.65 :source-step \"step-id\")"
  [claim & {:keys [confidence source-step source-prov tokens-cost id]
            :or {confidence 0.65}}]
  (when-let [g (current-graph)]
    (let [fact-id (or id (str "fact-" (count (:facts @g))))
          fact {:id fact-id
                :claim claim
                :confidence confidence
                :source-step source-step
                :source-prov source-prov
                :tokens-cost (or tokens-cost 0)}]
      (add-fact! g fact)
      fact-id)))

(defn sci-fact-from
  "Extract a fact from a step's output.
   Usage: (fact-from \"step-id\" :confidence 0.7)"
  [step-id & {:keys [confidence source-prov id]
              :or {confidence 0.65}}]
  (let [claim (core/ref step-id)]
    (when claim
      (sci-fact claim
                :confidence confidence
                :source-step step-id
                :source-prov source-prov
                :id id))))

(defn sci-approve!
  "Approve a fact in the current graph."
  [fact-id]
  (when-let [g (current-graph)]
    (approve! g fact-id)))

(defn sci-corroborate-from
  "Corroborate an existing fact with evidence from a step.
   Usage: (corroborate-from \"step-id\" :fact-id \"f1\" :authority 0.9)"
  [step-id & {:keys [fact-id authority]
              :or {authority 0.85}}]
  (when-let [g (current-graph)]
    (corroborate! g fact-id {:step step-id :authority authority})))

(defn sci-reachable?
  "Check if the goal is reachable in the current graph."
  []
  (when-let [g (current-graph)]
    (reachable? g "goal")))

(defn sci-graph-stats
  "Get stats for the current graph."
  []
  (when-let [g (current-graph)]
    (graph-stats g)))

(defn sci-confidence-gap
  "Get confidence gap analysis for the current graph."
  []
  (when-let [g (current-graph)]
    (confidence-gap g "goal")))

(defn sci-suggest-next
  "Get next investigation suggestion for the current graph."
  []
  (when-let [g (current-graph)]
    (suggest-next g "goal")))
