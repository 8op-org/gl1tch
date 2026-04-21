(ns glitch.store-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.store :as store]))

(def ^:dynamic *db* nil)

(defn temp-db-fixture [f]
  (let [path (str "/tmp/glitch-test-" (System/currentTimeMillis) ".edn")
        db   (store/open path)]
    (binding [*db* db]
      (try
        (f)
        (finally
          (store/close db)
          (let [file (java.io.File. path)]
            (when (.exists file) (.delete file))))))))

(use-fixtures :each temp-db-fixture)

(deftest test-open-creates-db
  (testing "open returns a store map with :conn and :path"
    (is (some? (:conn *db*)))
    (is (some? (:path *db*)))))

(deftest test-record-run-and-get-run
  (testing "insert and retrieve a run"
    (let [id (store/record-run *db*
               {:name          "test-workflow"
                :input         "hello world"
                :workflow-file "test.glitch"
                :model         "gemma4"})
          run (store/get-run *db* id)]
      (is (integer? id))
      (is (= "test-workflow" (:run/name run)))
      (is (= "hello world" (:run/input run)))
      (is (= "test.glitch" (:run/workflow run)))
      (is (= "gemma4" (:run/model run)))
      (is (integer? (:run/started-at run)))
      (is (nil? (:run/output run))))))

(deftest test-record-step-and-get-steps
  (testing "insert steps and retrieve by run-id"
    (let [run-id (store/record-run *db* {:name "step-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id    run-id
                    :step-id   "s1"
                    :prompt    "do the thing"
                    :output    "done"
                    :model     "gemma4"
                    :duration  150
                    :kind      "llm"})
          _      (store/record-step *db*
                   {:run-id    run-id
                    :step-id   "s2"
                    :prompt    "next thing"
                    :output    "ok"
                    :model     "gemma4"
                    :duration  80
                    :kind      "shell"})
          steps  (store/get-steps *db* run-id)]
      (is (= 2 (count steps)))
      (is (= #{"s1" "s2"} (set (map :step/id steps))))
      (is (= "do the thing" (:step/prompt (first (filter #(= "s1" (:step/id %)) steps)))))
      (is (= 150 (:step/duration (first (filter #(= "s1" (:step/id %)) steps))))))))

(deftest test-finish-run
  (testing "finish-run updates output and exit status"
    (let [id  (store/record-run *db* {:name "finish-test" :input "x"})
          _   (store/finish-run *db* id "all done" 0
                {:tokens-in 500 :tokens-out 200 :cost 0.02})
          run (store/get-run *db* id)]
      (is (= "all done" (:run/output run)))
      (is (= 0 (:run/exit run)))
      (is (integer? (:run/finished-at run))))))

(deftest test-list-runs
  (testing "list-runs returns runs in descending order with limit"
    (dotimes [i 5]
      (store/record-run *db* {:name (str "run-" i) :input (str i)}))
    (let [runs (store/list-runs *db* :limit 3)]
      (is (= 3 (count runs))))))

(deftest test-list-runs-workflow-filter
  (testing "list-runs with workflow filter"
    (store/record-run *db* {:name "a" :input "x" :workflow-file "deploy.glitch"})
    (store/record-run *db* {:name "b" :input "y" :workflow-file "deploy.glitch"})
    (store/record-run *db* {:name "c" :input "z" :workflow-file "other.glitch"})
    (let [runs (store/list-runs *db* :workflow "deploy.glitch")]
      (is (= 2 (count runs)))
      (is (every? #(= "deploy.glitch" (:run/workflow %)) runs)))))

(deftest test-step-confidence-column
  (testing "record-step persists confidence value"
    (let [run-id (store/record-run *db* {:name "conf-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "done"
                    :kind "llm" :confidence 0.85})
          steps  (store/get-steps *db* run-id)]
      (is (= 0.85 (:step/confidence (first steps)))))))

(deftest test-step-upsert
  (testing "recording a step with same run-id+step-id replaces the old row"
    (let [run-id (store/record-run *db* {:name "upsert-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "first"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "second"})
          steps  (store/get-steps *db* run-id)]
      (is (= 1 (count steps)))
      (is (= "second" (:step/output (first steps)))))))

(deftest test-record-fact-and-edges
  (testing "facts and edges are recorded and retrieved"
    (let [run-id (store/record-run *db* {:name "fact-test" :input "x"})
          _      (store/record-fact *db*
                   {:id "f1" :run-id run-id :claim "the sky is blue"
                    :confidence 0.9 :status :approved})
          _      (store/record-fact *db*
                   {:id "f2" :run-id run-id :claim "grass is green"
                    :confidence 0.8 :status :unapproved})
          _      (store/record-fact-edge *db*
                   {:run-id run-id :from-id "f1" :to-id "f2"
                    :rel :supports :weight 0.7})
          facts  (store/get-facts *db* run-id)
          edges  (store/get-fact-edges *db* run-id)]
      (is (= 2 (count facts)))
      (is (= 1 (count edges))))))

(deftest test-persistence
  (testing "data survives close and reopen"
    (let [path (str "/tmp/glitch-persist-test-" (System/currentTimeMillis) ".edn")
          db1  (store/open path)
          id   (store/record-run db1 {:name "persist" :input "x"})]
      (store/finish-run db1 id "done" 0)
      (store/close db1)
      (let [db2 (store/open path)
            run (store/get-run db2 id)]
        (try
          (is (= "done" (:run/output run)))
          (finally
            (store/close db2)
            (.delete (java.io.File. path))))))))
