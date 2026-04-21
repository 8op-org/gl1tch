(ns glitch.graph-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.graph :as graph]
            [glitch.core :as core]
            [glitch.runner :as runner]
            [glitch.provider :as prov]
            [clojure.java.io :as io]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Fixtures
;; ---------------------------------------------------------------------------

(use-fixtures :each
  (fn [f]
    (core/reset!)
    (core/set-input! "")
    (core/set-params! {})
    (core/set-step-recorder! nil)
    (core/set-provider-fn! nil)
    (reset! core/*call-stack* [])
    (reset! core/*graph* nil)
    (core/set-workflows-dir! ".")
    (f)))

;; ---------------------------------------------------------------------------
;; make-graph
;; ---------------------------------------------------------------------------

(deftest make-graph-test
  (testing "make-graph returns atom with correct empty structure"
    (let [g (graph/make-graph)]
      (is (instance? clojure.lang.Atom g))
      (is (= {} (:facts @g)))
      (is (= [] (:edges @g)))
      (is (nil? (:goal @g)))
      (is (= 0 (get-in @g [:stats :tokens-spent])))
      (is (= 0 (get-in @g [:stats :time-spent-ms])))
      (is (= 0 (get-in @g [:stats :facts-added]))))))

;; ---------------------------------------------------------------------------
;; add-fact!
;; ---------------------------------------------------------------------------

(deftest add-fact-basic-test
  (testing "add-fact! inserts a fact and caps confidence at 0.70"
    (let [g (graph/make-graph)
          fact {:id "f1" :claim "Service X is down" :confidence 0.85}]
      (graph/add-fact! g fact)
      (is (contains? (:facts @g) "f1"))
      (is (<= (get-in @g [:facts "f1" :confidence]) 0.70))
      (is (= :unapproved (get-in @g [:facts "f1" :status])))
      (is (= 1 (get-in @g [:stats :facts-added]))))))

(deftest add-fact-preserves-low-confidence-test
  (testing "add-fact! preserves confidence below 0.70"
    (let [g (graph/make-graph)
          fact {:id "f1" :claim "CPU is high" :confidence 0.50}]
      (graph/add-fact! g fact)
      (is (= 0.50 (get-in @g [:facts "f1" :confidence]))))))

(deftest add-fact-contradiction-test
  (testing "add-fact! detects contradiction with 3+ keyword overlap and reduces both facts"
    (let [g (graph/make-graph)]
      ;; Need 3+ non-stopword overlap tokens + polarity flip
      (graph/add-fact! g {:id "f1" :claim "service database query response healthy"
                          :confidence 0.65})
      (graph/add-fact! g {:id "f2" :claim "service database query response not healthy"
                          :confidence 0.65})
      ;; f2 decayed before insertion: 0.65 * 0.7^1 = 0.455, then contradict! hits both
      (is (< (get-in @g [:facts "f1" :confidence]) 0.65))
      (is (< (get-in @g [:facts "f2" :confidence]) 0.455))
      (is (= :contradicted (get-in @g [:facts "f1" :status])))
      (is (= :contradicted (get-in @g [:facts "f2" :status])))
      (is (= 1 (get-in @g [:stats :contradictions-found]))))))

;; ---------------------------------------------------------------------------
;; detect-contradictions
;; ---------------------------------------------------------------------------

(deftest detect-contradictions-polarity-flip-test
  (testing "detects polarity flip with 3+ keyword overlap"
    (let [facts {"f1" {:claim "database responding queries slowly today"}}
          new-fact {:claim "database responding queries not slowly today"}]
      (is (seq (graph/detect-contradictions new-fact facts))))))

(deftest detect-contradictions-no-false-positive-test
  (testing "does not flag unrelated claims as contradictions"
    (let [facts {"f1" {:claim "The database is healthy"}}
          new-fact {:claim "The network latency is high"}]
      (is (empty? (graph/detect-contradictions new-fact facts))))))

(deftest detect-contradictions-requires-overlap-test
  (testing "requires significant keyword overlap for contradiction"
    (let [facts {"f1" {:claim "Service A is running"}}
          new-fact {:claim "The weather is not good"}]
      (is (empty? (graph/detect-contradictions new-fact facts))))))

;; ---------------------------------------------------------------------------
;; contradict!
;; ---------------------------------------------------------------------------

(deftest contradict-test
  (testing "contradict! adds edge, reduces confidence, and sets :contradicted status"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "f1"] {:id "f1" :claim "A" :confidence 0.60 :status :unapproved})
      (swap! g assoc-in [:facts "f2"] {:id "f2" :claim "B" :confidence 0.60 :status :unapproved})
      (graph/contradict! g "f1" "f2")
      (is (== 0.42 (get-in @g [:facts "f1" :confidence])))
      (is (== 0.42 (get-in @g [:facts "f2" :confidence])))
      (is (= :contradicted (get-in @g [:facts "f1" :status])))
      (is (= :contradicted (get-in @g [:facts "f2" :status])))
      (is (= 1 (count (:edges @g))))
      (is (= :contradicts (:rel (first (:edges @g))))))))

;; ---------------------------------------------------------------------------
;; corroborate!
;; ---------------------------------------------------------------------------

(deftest corroborate-breaks-cap-test
  (testing "corroborate! uses spec formula: 1 - (1-old)(1-auth*0.85)"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "f1"] {:id "f1" :claim "Logs confirm OOM"
                                        :confidence 0.65 :status :unapproved})
      (graph/corroborate! g "f1" {:authority 0.90 :step "provider-b"})
      (let [new-conf (get-in @g [:facts "f1" :confidence])]
        ;; 1 - (1-0.65)(1-0.90*0.85) = 1 - 0.35 * 0.235 = 1 - 0.08225 = 0.91775
        (is (> new-conf 0.91) "Should be ~0.918")
        (is (< new-conf 0.92))
        (is (> new-conf 0.70) "Should break the 0.70 single-source cap")))))

(deftest corroborate-adds-edge-test
  (testing "corroborate! adds a :corroborates edge"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "f1"] {:id "f1" :claim "test" :confidence 0.5
                                        :status :unapproved})
      (graph/corroborate! g "f1" {:authority 0.85 :step "source-a"})
      (let [edges (:edges @g)]
        (is (= 1 (count edges)))
        (is (= :corroborates (:rel (first edges))))
        (is (= "source-a" (:source (first edges))))))))

;; ---------------------------------------------------------------------------
;; approve!
;; ---------------------------------------------------------------------------

(deftest approve-test
  (testing "approve! sets status to :approved and confidence to 1.0"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "f1"] {:id "f1" :claim "Confirmed" :confidence 0.6
                                        :status :unapproved})
      (graph/approve! g "f1")
      (is (= :approved (get-in @g [:facts "f1" :status])))
      (is (= 1.0 (get-in @g [:facts "f1" :confidence]))))))

;; ---------------------------------------------------------------------------
;; shortest-path
;; ---------------------------------------------------------------------------

(deftest shortest-path-test
  (testing "finds shortest path using edge weights (not node confidence)"
    (let [g (graph/make-graph)]
      ;; Build: approved -> f1 -> f2 -> goal
      (swap! g assoc-in [:facts "start"] {:id "start" :status :approved :confidence 1.0
                                           :claim "Known"})
      (swap! g assoc-in [:facts "f1"] {:id "f1" :status :unapproved :confidence 0.8
                                        :claim "Intermediate"})
      (swap! g assoc-in [:facts "f2"] {:id "f2" :status :unapproved :confidence 0.6
                                        :claim "Near goal"})
      (swap! g assoc-in [:facts "goal"] {:id "goal" :status :goal :confidence 0.0
                                          :claim "Investigation goal"})
      (graph/add-edge! g {:from "start" :to "f1" :rel :supports :weight 0.9})
      (graph/add-edge! g {:from "f1" :to "f2" :rel :supports :weight 0.7})
      (graph/add-edge! g {:from "f2" :to "goal" :rel :supports :weight 0.5})
      (let [result (graph/shortest-path g "goal")]
        (is (some? result))
        (is (= ["start" "f1" "f2" "goal"] (:path result)))
        ;; Cost = (1-0.9) + (1-0.7) + (1-0.5) = 0.1 + 0.3 + 0.5 = 0.9
        (is (< (Math/abs (- (:cost result) 0.9)) 0.001))
        ;; min-edge = min(0.9, 0.7, 0.5) = 0.5
        (is (= 0.5 (:min-edge result)))))))

(deftest shortest-path-directional-test
  (testing "edges are directional — reverse direction not traversable"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "start"] {:id "start" :status :approved :confidence 1.0
                                           :claim "Known"})
      (swap! g assoc-in [:facts "mid"] {:id "mid" :status :unapproved :confidence 0.5
                                         :claim "Mid"})
      (swap! g assoc-in [:facts "goal"] {:id "goal" :status :goal :confidence 0.0
                                          :claim "Goal"})
      ;; Edge goes goal -> mid -> start (wrong direction)
      (graph/add-edge! g {:from "goal" :to "mid" :rel :supports :weight 0.8})
      (graph/add-edge! g {:from "mid" :to "start" :rel :supports :weight 0.8})
      ;; Should NOT find a path since edges point the wrong way
      (is (nil? (graph/shortest-path g "goal"))))))

(deftest shortest-path-unreachable-test
  (testing "returns nil for disconnected graph"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "start"] {:id "start" :status :approved :confidence 1.0
                                           :claim "Known"})
      (swap! g assoc-in [:facts "goal"] {:id "goal" :status :goal :confidence 0.0
                                          :claim "Unreachable"})
      ;; No edges connecting them
      (is (nil? (graph/shortest-path g "goal"))))))

(deftest shortest-path-no-approved-test
  (testing "returns nil when no approved facts exist"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "f1"] {:id "f1" :status :unapproved :confidence 0.5
                                        :claim "A"})
      (swap! g assoc-in [:facts "goal"] {:id "goal" :status :goal :confidence 0.0
                                          :claim "Goal"})
      (graph/add-edge! g {:from "f1" :to "goal" :rel :supports})
      (is (nil? (graph/shortest-path g "goal"))))))

;; ---------------------------------------------------------------------------
;; confidence-gap
;; ---------------------------------------------------------------------------

(deftest confidence-gap-test
  (testing "finds weakest node on critical path and includes :path-cost"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "start"] {:id "start" :status :approved :confidence 1.0
                                           :claim "Known"})
      (swap! g assoc-in [:facts "strong"] {:id "strong" :status :unapproved :confidence 0.9
                                            :claim "Strong fact"})
      (swap! g assoc-in [:facts "weak"] {:id "weak" :status :unapproved :confidence 0.3
                                          :claim "Weak fact"})
      (swap! g assoc-in [:facts "goal"] {:id "goal" :status :goal :confidence 0.0
                                          :claim "Goal"})
      (graph/add-edge! g {:from "start" :to "strong" :rel :supports :weight 0.8})
      (graph/add-edge! g {:from "strong" :to "weak" :rel :supports :weight 0.6})
      (graph/add-edge! g {:from "weak" :to "goal" :rel :supports :weight 0.4})
      (let [gap (graph/confidence-gap g "goal")]
        (is (some? gap))
        (is (= "weak" (:fact-id gap)))
        (is (= 0.3 (:confidence gap)))
        (is (number? (:path-cost gap)))))))

;; ---------------------------------------------------------------------------
;; suggest-next
;; ---------------------------------------------------------------------------

(deftest suggest-next-contradiction-priority-test
  (testing "contradictions are prioritized over weak links"
    (let [g (graph/make-graph)]
      ;; Set facts with :contradicted status (as contradict! would)
      (swap! g assoc-in [:facts "f1"] {:id "f1" :status :contradicted :confidence 0.35
                                        :claim "A"})
      (swap! g assoc-in [:facts "f2"] {:id "f2" :status :contradicted :confidence 0.35
                                        :claim "B"})
      (swap! g update :edges conj {:from "f1" :to "f2" :rel :contradicts :weight 0.0})
      (let [suggestion (graph/suggest-next g "goal")]
        (is (= :resolve-contradiction (:action suggestion)))
        (is (vector? (:facts suggestion)))
        (is (= 2 (count (:facts suggestion))))
        ;; Each fact in the vector should have :id :claim :confidence
        (is (every? #(contains? % :id) (:facts suggestion)))))))

(deftest suggest-next-weakest-link-test
  (testing "suggests strengthening weakest link when no contradictions"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "start"] {:id "start" :status :approved :confidence 1.0
                                           :claim "Known"})
      (swap! g assoc-in [:facts "weak"] {:id "weak" :status :unapproved :confidence 0.3
                                          :claim "Weak claim"})
      (swap! g assoc-in [:facts "goal"] {:id "goal" :status :goal :confidence 0.0
                                          :claim "Goal"})
      (graph/add-edge! g {:from "start" :to "weak" :rel :supports :weight 0.8})
      (graph/add-edge! g {:from "weak" :to "goal" :rel :supports :weight 0.5})
      (let [suggestion (graph/suggest-next g "goal")]
        (is (= :strengthen (:action suggestion)))
        (is (map? (:fact suggestion)))
        (is (= "weak" (get-in suggestion [:fact :fact-id])))
        (is (string? (:suggestion suggestion)))
        (is (str/includes? (:suggestion suggestion) "Corroborate"))))))

(deftest suggest-next-explore-test
  (testing "suggests exploration when no path exists"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:facts "orphan"] {:id "orphan" :status :unapproved :confidence 0.4
                                            :claim "Isolated claim"})
      (let [suggestion (graph/suggest-next g "goal")]
        (is (= :explore (:action suggestion)))
        (is (= "No path to goal exists. Discover connecting facts." (:suggestion suggestion)))))))

;; ---------------------------------------------------------------------------
;; efficiency-metrics
;; ---------------------------------------------------------------------------

(deftest efficiency-metrics-test
  (testing "computes correct ratios with spec key names"
    (let [g (graph/make-graph)]
      (swap! g assoc-in [:stats :tokens-spent] 1000)
      (swap! g assoc-in [:stats :time-spent-ms] 500)
      (swap! g assoc-in [:facts "f1"] {:id "f1" :confidence 0.8 :claim "A"})
      (swap! g assoc-in [:facts "f2"] {:id "f2" :confidence 0.6 :claim "B"})
      (let [m (graph/efficiency-metrics g)]
        (is (= 2 (:total-facts m)))
        (is (= 0.7 (:avg-confidence m)))
        (is (= 1000 (:tokens-spent m)))
        (is (= 500 (:time-spent-ms m)))
        (is (> (:conf-per-token m) 0))
        (is (> (:conf-per-time-ms m) 0))
        (is (= 2.0 (:facts-per-1k-tok m)))))))

(deftest efficiency-metrics-zero-tokens-test
  (testing "handles zero tokens gracefully"
    (let [g (graph/make-graph)]
      (let [m (graph/efficiency-metrics g)]
        (is (= 0.0 (:conf-per-token m)))
        (is (= 0.0 (:conf-per-time-ms m)))
        (is (= 0 (:total-facts m)))
        (is (= 0.0 (:avg-confidence m)))))))

;; ---------------------------------------------------------------------------
;; graph-stats
;; ---------------------------------------------------------------------------

(deftest graph-stats-test
  (testing "returns correct counts"
    (let [g (graph/make-graph)]
      (graph/set-goal! g "Find root cause")
      (graph/add-fact! g {:id "f1" :claim "Log shows errors" :confidence 0.6})
      (graph/add-fact! g {:id "f2" :claim "CPU is high" :confidence 0.5})
      (graph/approve! g "f1")
      (let [stats (graph/graph-stats g)]
        ;; goal + f1 + f2 = 3
        (is (= 3 (:node-count stats)))
        (is (= 1 (:approved-count stats)))
        ;; f2 is unapproved (f1 was unapproved but got approved)
        (is (= 1 (:unapproved-count stats)))
        (is (= "Find root cause" (:goal stats)))))))

;; ---------------------------------------------------------------------------
;; SCI helpers (require *graph* binding)
;; ---------------------------------------------------------------------------

(deftest sci-fact-test
  (testing "sci-fact adds to bound graph"
    (let [g (graph/make-graph)]
      (reset! core/*graph* g)
      (let [fid (graph/sci-fact "Test claim" :confidence 0.6)]
        (is (string? fid))
        (is (contains? (:facts @g) fid))
        (is (<= (get-in @g [:facts fid :confidence]) 0.70)))
      (reset! core/*graph* nil))))

(deftest sci-fact-from-test
  (testing "sci-fact-from creates fact from step output"
    (let [g (graph/make-graph)]
      (reset! core/*graph* g)
      (core/step "my-step" "The service is healthy")
      (let [fid (graph/sci-fact-from "my-step" :confidence 0.65)]
        (is (string? fid))
        (is (= "The service is healthy" (get-in @g [:facts fid :claim]))))
      (reset! core/*graph* nil))))

;; ---------------------------------------------------------------------------
;; Integration — investigate macro via SCI runner
;; ---------------------------------------------------------------------------

(def ^:private test-dir "/tmp/glitch-graph-test")

(defn- write-temp-workflow [filename content]
  (let [path (str test-dir "/" filename)]
    (io/make-parents path)
    (spit path content)
    path))

(defn- cleanup []
  (let [d (io/file test-dir)]
    (when (.exists d)
      (doseq [f (reverse (file-seq d))]
        (.delete f)))))

(deftest investigate-macro-integration-test
  (testing "investigate macro creates graph, adds facts, and returns results"
    (cleanup)
    (.mkdirs (io/file test-dir))
    (prov/reset!)
    (prov/register "mock"
      (fn [opts]
        {:response "{\"finding\": \"OOM detected\", \"confidence\": 0.8}"
         :tokens-in 50 :tokens-out 100}))
    (try
      (let [path (write-temp-workflow "investigate.glitch"
                   "(let [result (investigate \"Find root cause of outage\"
                                  (step \"logs\" \"Error: OOM killed process\")
                                  (fact \"OOM killed the process\" :confidence 0.65 :source-step \"logs\")
                                  (fact \"Memory usage was at 98%\" :confidence 0.60))]
                     (step \"result\" (str (:reachable result))))")
            result (runner/run path :tiers [{:providers ["mock"]}])]
        ;; The workflow should complete without error
        (is (some? (:output result)))
        ;; Graph var should be cleared after investigate
        (is (nil? @core/*graph*)))
      (finally
        (cleanup)))))

(deftest investigate-cleans-up-on-error-test
  (testing "investigate macro clears *graph* even on error"
    (cleanup)
    (.mkdirs (io/file test-dir))
    (prov/reset!)
    (try
      (let [path (write-temp-workflow "investigate-err.glitch"
                   "(investigate \"Goal\"
                      (throw (ex-info \"boom\" {})))")]
        (try
          (runner/run path)
          (catch Exception _))
        ;; Graph should be cleared regardless
        (is (nil? @core/*graph*)))
      (finally
        (cleanup)))))

(deftest llm-tracks-graph-stats-test
  (testing "llm function tracks tokens and time in graph stats"
    (let [g (graph/make-graph)]
      (reset! core/*graph* g)
      (core/set-provider-fn!
        (fn [opts]
          {:response "answer" :tokens-in 50 :tokens-out 100}))
      (core/llm :prompt "test question")
      (is (= 150 (get-in @g [:stats :tokens-spent])))
      ;; time-spent-ms may be 0 for very fast mock providers
      (is (>= (get-in @g [:stats :time-spent-ms]) 0))
      (reset! core/*graph* nil))))
