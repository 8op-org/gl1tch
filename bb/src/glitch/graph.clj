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
    "couldn't" "hasn't" "haven't" "hadn't" "mustn't"})

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
                     (and (>= (count overlap) 2)
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
   the 0.70 single-source confidence cap.
   fact map: {:id :claim :confidence :source-step :source-prov :tokens-cost}"
  [graph fact]
  (let [capped-conf (min (:confidence fact 0.5) 0.70)
        fact (assoc fact
               :confidence capped-conf
               :status (or (:status fact) :unapproved)
               :created-at (or (:created-at fact) (System/currentTimeMillis)))
        existing (:facts @graph)
        contradicting (detect-contradictions fact existing)]
    ;; Add fact
    (swap! graph assoc-in [:facts (:id fact)] fact)
    (swap! graph update-in [:stats :facts-added] inc)
    ;; Handle contradictions
    (doseq [contra-id contradicting]
      (contradict! graph (:id fact) contra-id))
    fact))

(defn add-edge!
  "Link two facts with a relationship edge.
   edge map: {:from :to :rel :weight :source}"
  [graph edge]
  (let [edge (assoc edge :weight (or (:weight edge) 1.0))]
    (swap! graph update :edges conj edge)
    edge))

(defn contradict!
  "Mark a contradiction between two facts. Reduces confidence on both by 0.7x.
   Adds a :contradicts edge."
  [graph id-a id-b]
  (swap! graph (fn [g]
                 (-> g
                     (update-in [:facts id-a :confidence] * 0.7)
                     (update-in [:facts id-b :confidence] * 0.7)
                     (update :edges conj {:from id-a :to id-b
                                          :rel :contradicts :weight 0.0
                                          :source "contradiction-detector"})
                     (update-in [:stats :contradictions-found] inc)))))

(defn corroborate!
  "Corroborate a fact with evidence from a new source.
   Uses Bayesian accumulation to break the 0.70 single-source cap.
   If glitch.confidence is available, uses bayesian-combine; otherwise
   uses a simple formula: 1 - (1-current)(1-authority*base)."
  [graph fact-id new-source & {:keys [authority base-conf]
                                :or {authority 0.85 base-conf 0.65}}]
  (swap! graph (fn [g]
                 (let [current-conf (get-in g [:facts fact-id :confidence] 0.5)
                       ;; Simple Bayesian accumulation: 1 - (1-a)(1-b)
                       new-conf (- 1.0 (* (- 1.0 current-conf)
                                          (- 1.0 (* authority base-conf))))]
                     (-> g
                         (assoc-in [:facts fact-id :confidence] (min 1.0 new-conf))
                         (assoc-in [:facts fact-id :sources]
                                   (conj (get-in g [:facts fact-id :sources] []) new-source))
                         (update-in [:stats :corroborations] inc))))))

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
              (-> adj
                  (update from (fnil conj []) edge)
                  (update to (fnil conj []) (assoc edge :from to :to from)))))
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
      ;; Dijkstra
      (loop [dist (into {} (map (fn [id] [id 0.0]) start-ids))
             prev (into {} (map (fn [id] [id nil]) start-ids))
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
                           (recur (get prev id) (conj acc id))))]
              {:path (vec path) :cost (get dist goal-id)}))
          (let [[cost-u u] (first queue)
                queue (disj queue (first queue))]
            (if (contains? visited u)
              (recur dist prev queue visited)
              (let [visited (conj visited u)
                    neighbors (get adj u [])
                    [dist' prev' queue']
                    (reduce
                      (fn [[d p q] {:keys [to]}]
                        (if (contains? visited to)
                          [d p q]
                          (let [target-conf (get-in facts [to :confidence] 0.0)
                                edge-cost (- 1.0 target-conf)
                                alt (+ cost-u edge-cost)]
                            (if (< alt (get d to Double/MAX_VALUE))
                              [(assoc d to alt)
                               (assoc p to u)
                               (conj q [alt to])]
                              [d p q]))))
                      [dist prev queue]
                      neighbors)]
                (recur dist' prev' queue' visited)))))))))

(defn confidence-gap
  "Find the weakest link on the best path from approved facts to goal.
   Returns {:fact-id :confidence :claim} or nil if no path."
  [graph goal-id]
  (when-let [{:keys [path]} (shortest-path graph goal-id)]
    (let [facts (:facts @graph)
          path-facts (->> path
                          (map (fn [id] (assoc (get facts id) :id id)))
                          (remove #(= :approved (:status %)))
                          (remove #(= :goal (:status %))))]
      (when (seq path-facts)
        (let [weakest (apply min-key :confidence path-facts)]
          {:fact-id (:id weakest)
           :confidence (:confidence weakest)
           :claim (:claim weakest)})))))

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
        ;; 1. Find contradictions
        contradictions (->> edges
                           (filter #(= :contradicts (:rel %)))
                           (mapv (fn [e]
                                   {:action :resolve-contradiction
                                    :fact-a (:from e)
                                    :fact-b (:to e)
                                    :claim-a (get-in facts [(:from e) :claim])
                                    :claim-b (get-in facts [(:to e) :claim])})))]
    (if (seq contradictions)
      (first contradictions)
      ;; 2. Weakest link
      (if-let [gap (confidence-gap graph goal-id)]
        {:action :strengthen
         :fact-id (:fact-id gap)
         :confidence (:confidence gap)
         :claim (:claim gap)}
        ;; 3. Explore — find unapproved facts not on any path
        (let [unapproved (->> facts
                              (filter (fn [[_ f]] (= :unapproved (:status f))))
                              (sort-by (fn [[_ f]] (:confidence f)))
                              first)]
          (when unapproved
            {:action :explore
             :fact-id (first unapproved)
             :confidence (:confidence (second unapproved))
             :claim (:claim (second unapproved))}))))))

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
    {:total-confidence total-conf
     :tokens-spent tokens
     :time-spent-ms time-ms
     :confidence-per-token (if (pos? tokens)
                             (/ total-conf tokens)
                             0.0)
     :confidence-per-ms (if (pos? time-ms)
                          (/ total-conf time-ms)
                          0.0)}))

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
    (corroborate! g fact-id step-id :authority authority)))

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
