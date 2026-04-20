(ns glitch.store-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.store :as store]))

;; The pod is loaded by glitch.store; we can require its ns after that.
(require '[pod.babashka.go-sqlite3 :as sql])

(def ^:dynamic *db* nil)

(defn temp-db-fixture [f]
  (let [path (str "/tmp/glitch-test-" (System/currentTimeMillis) ".db")
        db   (store/open path)]
    (binding [*db* db]
      (try
        (f)
        (finally
          (store/close db)
          (let [file (java.io.File. path)]
            (when (.exists file) (.delete file))
            ;; Clean up WAL/SHM files too
            (let [wal (java.io.File. (str path "-wal"))
                  shm (java.io.File. (str path "-shm"))]
              (when (.exists wal) (.delete wal))
              (when (.exists shm) (.delete shm)))))))))

(use-fixtures :each temp-db-fixture)

;; ---------------------------------------------------------------------------
;; open
;; ---------------------------------------------------------------------------

(deftest test-open-creates-db-and-tables
  (testing "open creates the db file"
    (is (.exists (java.io.File. *db*))))

  (testing "schema tables exist"
    (let [tables (->> (sql/query *db*
                        "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
                      (map :name)
                      set)]
      (is (contains? tables "runs"))
      (is (contains? tables "steps"))
      (is (contains? tables "research_events")))))

;; ---------------------------------------------------------------------------
;; record-run + get-run
;; ---------------------------------------------------------------------------

(deftest test-record-run-and-get-run
  (testing "insert and retrieve a run"
    (let [id (store/record-run *db*
               {:name          "test-workflow"
                :input         "hello world"
                :workflow-file "test.glitch"
                :model         "gemma4"
                :variant       nil
                :workspace     "ws1"
                :parent-run-id nil
                :workflow-name "test-wf"})
          run (store/get-run *db* id)]
      (is (integer? id))
      (is (= id (:id run)))
      (is (= "test-workflow" (:name run)))
      (is (= "hello world" (:input run)))
      (is (= "test.glitch" (:workflow_file run)))
      (is (= "gemma4" (:model run)))
      (is (= "ws1" (:workspace run)))
      (is (= "test-wf" (:workflow_name run)))
      (is (= "workflow" (:kind run)))
      (is (integer? (:started_at run)))
      (is (nil? (:output run)))
      (is (nil? (:exit_status run))))))

;; ---------------------------------------------------------------------------
;; record-step + get-steps
;; ---------------------------------------------------------------------------

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
                    :kind      "llm"
                    :exit      0
                    :tokens-in 100
                    :tokens-out 50
                    :gate      nil
                    :artifacts nil})
          _      (store/record-step *db*
                   {:run-id    run-id
                    :step-id   "s2"
                    :prompt    "next thing"
                    :output    "ok"
                    :model     "gemma4"
                    :duration  80
                    :kind      "shell"
                    :exit      0
                    :tokens-in nil
                    :tokens-out nil
                    :gate      nil
                    :artifacts nil})
          steps  (store/get-steps *db* run-id)]
      (is (= 2 (count steps)))
      (is (= "s1" (:step_id (first steps))))
      (is (= "s2" (:step_id (second steps))))
      (is (= "do the thing" (:prompt (first steps))))
      (is (= 150 (:duration_ms (first steps)))))))

;; ---------------------------------------------------------------------------
;; finish-run
;; ---------------------------------------------------------------------------

(deftest test-finish-run
  (testing "finish-run updates output, exit_status, and totals"
    (let [id  (store/record-run *db* {:name "finish-test" :input "x"})
          _   (store/finish-run *db* id "all done" 0
                {:tokens-in 500 :tokens-out 200 :cost 0.02})
          run (store/get-run *db* id)]
      (is (= "all done" (:output run)))
      (is (= 0 (:exit_status run)))
      (is (= 500 (:tokens_in run)))
      (is (= 200 (:tokens_out run)))
      (is (= 0.02 (:cost_usd run)))
      (is (integer? (:finished_at run)))))

  (testing "finish-run without totals defaults to zero"
    (let [id  (store/record-run *db* {:name "finish-no-totals" :input "y"})
          _   (store/finish-run *db* id "result" 1)
          run (store/get-run *db* id)]
      (is (= "result" (:output run)))
      (is (= 1 (:exit_status run)))
      (is (= 0 (:tokens_in run))))))

;; ---------------------------------------------------------------------------
;; list-runs
;; ---------------------------------------------------------------------------

(deftest test-list-runs
  (testing "list-runs returns runs in descending order with limit"
    (dotimes [i 5]
      (store/record-run *db* {:name (str "run-" i) :input (str i)}))
    (let [runs (store/list-runs *db* :limit 3)]
      (is (= 3 (count runs)))
      ;; Descending order: newest first
      (is (> (:id (first runs)) (:id (second runs)))))))

(deftest test-list-runs-parent-filter
  (testing "list-runs with parent-id filter"
    (let [parent-id (store/record-run *db* {:name "parent" :input "p"})
          child1    (store/record-run *db* {:name "child1" :input "c1"
                                            :parent-run-id parent-id})
          child2    (store/record-run *db* {:name "child2" :input "c2"
                                            :parent-run-id parent-id})
          _orphan   (store/record-run *db* {:name "orphan" :input "o"})
          children  (store/list-runs *db* :parent-id parent-id)]
      (is (= 2 (count children)))
      (is (every? #(= parent-id (:parent_run_id %)) children)))))

(deftest test-list-runs-workflow-filter
  (testing "list-runs with workflow filter"
    (store/record-run *db* {:name "a" :input "x" :workflow-file "deploy.glitch"})
    (store/record-run *db* {:name "b" :input "y" :workflow-file "deploy.glitch"})
    (store/record-run *db* {:name "c" :input "z" :workflow-file "other.glitch"})
    (let [runs (store/list-runs *db* :workflow "deploy.glitch")]
      (is (= 2 (count runs)))
      (is (every? #(= "deploy.glitch" (:workflow_file %)) runs)))))

;; ---------------------------------------------------------------------------
;; step upsert (INSERT OR REPLACE)
;; ---------------------------------------------------------------------------

(deftest test-step-upsert
  (testing "recording a step with same run-id+step-id replaces the old row"
    (let [run-id (store/record-run *db* {:name "upsert-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "first"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "second"})
          steps  (store/get-steps *db* run-id)]
      (is (= 1 (count steps)))
      (is (= "second" (:output (first steps)))))))
