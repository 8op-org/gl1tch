(ns glitch.mcp.search-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.mcp.search :as search]))

(deftest normalize-scores-test
  (testing "empty returns empty"
    (is (= [] (search/normalize-scores []))))
  (testing "single item gets score 1"
    (is (= [{:id 1 :score 1.0}]
           (search/normalize-scores [{:id 1 :score 5.0}]))))
  (testing "all same score get 1"
    (is (= [{:id 1 :score 1.0} {:id 2 :score 1.0}]
           (search/normalize-scores [{:id 1 :score 3.0} {:id 2 :score 3.0}]))))
  (testing "different scores scale linearly"
    (let [result (search/normalize-scores [{:id 1 :score 0.0} {:id 2 :score 10.0}])]
      (is (= 0.0 (:score (first result))))
      (is (= 1.0 (:score (second result)))))))

(deftest merge-scores-test
  (let [kw [{:id 1 :score 1.0} {:id 2 :score 0.5}]
        sem [{:id 2 :score 1.0} {:id 3 :score 0.8}]
        merged (search/merge-scores kw sem)]
    (testing "merged results contain all IDs"
      (is (= #{1 2 3} (set (map :id merged)))))
    (testing "results sorted descending by score"
      (is (apply >= (map :score merged))))
    (testing "ID 2 appears in both and has highest combined score"
      (is (= 2 (:id (first merged)))))))
