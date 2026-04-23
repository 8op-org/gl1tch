(ns glitch.promote-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.promote :as promote]))

(deftest slugify-test
  (testing "basic slugification"
    (is (= "check-pr-accuracy" (promote/slugify "Check PR Accuracy")))
    (is (= "hello-world" (promote/slugify "  Hello   World  ")))
    (is (= "test123" (promote/slugify "test!@#123")))))

(deftest task-shape-extraction-test
  (testing "extracts task from advise entries in session"
    (let [session [{:type :sh :id "fetch" :args ["curl ..."] :output "data"}
                   {:type :advise :task "check PR accuracy"
                    :recommendation {:approach "primitive"}}
                   {:type :llm :id "analyze" :output "result"}]]
      (let [advise-entries (filter #(= :advise (:type %)) session)]
        (is (= "check PR accuracy" (:task (last advise-entries)))))))

  (testing "returns nil when no advise entries"
    (let [session [{:type :sh :id "fetch" :args ["curl"] :output "data"}]]
      (let [advise-entries (filter #(= :advise (:type %)) session)]
        (is (nil? (when (seq advise-entries) (:task (last advise-entries)))))))))
