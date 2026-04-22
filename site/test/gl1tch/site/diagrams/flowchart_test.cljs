(ns gl1tch.site.diagrams.flowchart-test
  (:require [cljs.test :refer-macros [deftest is testing]]
            [gl1tch.site.diagrams.flowchart :as fc]))

(deftest renders-svg
  (testing "flowchart returns an svg hiccup vector"
    (let [spec {:direction :lr
                :nodes [{:id :a :label "Start"} {:id :b :label "End"}]
                :edges [[:a :b]]}
          result (fc/render spec)]
      (is (= :svg (first result)))
      (is (map? (second result))))))

(deftest parses-inline-form
  (testing "parses inline diagram EDN to spec"
    (let [inline [[:node :a "Start"]
                  [:node :b "End"]
                  [:edge :a :b]]
          spec (fc/parse-inline {:type :flowchart :direction :lr} inline)]
      (is (= 2 (count (:nodes spec))))
      (is (= 1 (count (:edges spec))))
      (is (= :lr (:direction spec))))))
