(ns gl1tch.site.content-test
  (:require [cljs.test :refer-macros [deftest is testing]]
            [gl1tch.site.content :as content]))

(deftest doc-index-sorting
  (testing "docs sort by :order"
    (let [docs [{:slug "b" :order 2 :title "B"}
                {:slug "a" :order 1 :title "A"}]
          sorted (content/sort-docs docs)]
      (is (= "a" (:slug (first sorted))))
      (is (= "b" (:slug (second sorted)))))))

(deftest extract-toc
  (testing "extracts headings from sections"
    (let [doc {:sections [{:heading "Install" :level 2 :body []}
                          {:heading "Config" :level 2 :body []}
                          {:heading "Advanced" :level 3 :body []}]}
          toc (content/extract-toc doc)]
      (is (= 3 (count toc)))
      (is (= "Install" (:heading (first toc))))
      (is (= 3 (:level (nth toc 2)))))))
