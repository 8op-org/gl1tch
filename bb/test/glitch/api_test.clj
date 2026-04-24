(ns glitch.api-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.api :as api]))

(use-fixtures :each
  (fn [f]
    (reset! api/tool-registry {})
    (f)
    (reset! api/tool-registry {})))

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

(deftest register-tool-returns-nil-test
  (testing "register-tool! returns nil"
    (is (nil? (api/register-tool! {:name "x" :description "x" :parameters {} :fn identity})))))

(deftest list-tools-test
  (testing "returns empty when no tools registered"
    (is (empty? (api/list-tools))))
  (testing "returns tool names after registration"
    (api/register-tool! {:name "search" :description "Search" :parameters {} :fn identity})
    (is (= ["search"] (vec (api/list-tools))))))

(deftest remove-tool-test
  (testing "removes a registered tool"
    (api/register-tool! {:name "search" :description "Search" :parameters {} :fn identity})
    (api/remove-tool! "search")
    (is (empty? (api/list-tools))))
  (testing "no-ops on unknown tool"
    (is (nil? (api/remove-tool! "nonexistent")))))

(deftest deftool-test
  (testing "deftool registers a tool with correct shape"
    (api/deftool my-search [query]
      "Search the codebase for query"
      (str "results for " query))
    (let [t (get @api/tool-registry "my-search")]
      (is (= "my-search" (:name t)))
      (is (= "Search the codebase for query" (:description t)))
      (is (= {:type "object"
              :properties {"query" {:type "string"}}
              :required ["query"]}
             (:parameters t)))
      (is (= "results for hello" ((:fn t) {"query" "hello"})))))

  (testing "deftool with multiple args"
    (api/deftool multi-arg [query limit]
      "Search with limit"
      (str query " " limit))
    (let [t (get @api/tool-registry "multi-arg")]
      (is (= #{"query" "limit"} (set (keys (get-in t [:parameters :properties]))))))))

(deftest deftool-returns-name-test
  (testing "deftool returns the tool name as a string"
    (is (= "ret-tool" (api/deftool ret-tool [x] "desc" x)))))
