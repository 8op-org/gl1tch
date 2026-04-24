(ns glitch.api-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.api :as api]))

(use-fixtures :each
  (fn [f]
    (reset! api/tool-registry {})
    (f)))

(deftest register-tool-test
  (testing "registers a tool in the registry"
    (api/register-tool! {:name "search"
                         :description "Search the index"
                         :parameters {:type "object"
                                      :properties {"query" {:type "string"}}
                                      :required ["query"]}
                         :fn (fn [_] "result")})
    (is (= 1 (count @api/tool-registry)))
    (is (= "search" (:name (get @api/tool-registry "search"))))))

(deftest list-tools-test
  (testing "returns empty when no tools registered"
    (is (= [] (api/list-tools))))
  (testing "returns tool names after registration"
    (api/register-tool! {:name "search" :description "Search" :parameters {} :fn identity})
    (is (= ["search"] (api/list-tools)))))

(deftest remove-tool-test
  (testing "removes a registered tool"
    (api/register-tool! {:name "search" :description "Search" :parameters {} :fn identity})
    (api/remove-tool! "search")
    (is (= [] (api/list-tools))))
  (testing "no-ops on unknown tool"
    (is (nil? (api/remove-tool! "nonexistent")))))
