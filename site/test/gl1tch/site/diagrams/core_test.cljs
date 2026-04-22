(ns gl1tch.site.diagrams.core-test
  (:require [cljs.test :refer-macros [deftest is testing]]
            [gl1tch.site.diagrams.core :as layout]))

(deftest topological-layers
  (testing "linear chain produces sequential layers"
    (let [nodes [{:id :a} {:id :b} {:id :c}]
          edges [[:a :b] [:b :c]]
          layers (layout/assign-layers nodes edges)]
      (is (= 3 (count layers)))
      (is (= [:a] (first layers)))
      (is (= [:b] (second layers)))
      (is (= [:c] (nth layers 2)))))
  (testing "diamond produces 3 layers"
    (let [nodes [{:id :a} {:id :b} {:id :c} {:id :d}]
          edges [[:a :b] [:a :c] [:b :d] [:c :d]]
          layers (layout/assign-layers nodes edges)]
      (is (= 3 (count layers)))
      (is (= [:a] (first layers)))
      (is (= #{:b :c} (set (second layers))))
      (is (= [:d] (nth layers 2))))))

(deftest position-nodes
  (testing "positions are computed for lr direction"
    (let [spec {:direction :lr
                :nodes [{:id :a :label "A"} {:id :b :label "B"}]
                :edges [[:a :b]]}
          result (layout/compute spec)]
      (is (every? #(contains? % :x) (:nodes result)))
      (is (every? #(contains? % :y) (:nodes result)))
      (is (< (:x (first (:nodes result)))
             (:x (second (:nodes result))))))))
