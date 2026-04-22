(ns gl1tch.site.router-test
  (:require [cljs.test :refer-macros [deftest is testing]]
            [gl1tch.site.router :as router]))

(deftest route-matching
  (testing "matches root"
    (is (= :home (:name (router/match "/")))))
  (testing "matches doc slug"
    (let [m (router/match "/docs/getting-started")]
      (is (= :docs/page (:name m)))
      (is (= "getting-started" (get-in m [:params :slug])))))
  (testing "matches labs index"
    (is (= :labs/index (:name (router/match "/labs")))))
  (testing "matches lab slug"
    (let [m (router/match "/labs/repl-driven-docs")]
      (is (= :labs/page (:name m)))
      (is (= "repl-driven-docs" (get-in m [:params :slug])))))
  (testing "matches changelog"
    (is (= :changelog (:name (router/match "/changelog"))))))
