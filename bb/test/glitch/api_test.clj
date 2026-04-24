(ns glitch.api-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.api :as api]
            [glitch.core :as g]))

(use-fixtures :each
  (fn [f]
    (reset! api/tool-registry {})
    (reset! api/current-model nil)
    (g/set-provider-fn! nil)
    (f)
    (reset! api/tool-registry {})
    (reset! api/current-model nil)
    (g/set-provider-fn! nil)))

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

(deftest agent-test
  (testing "agent calls provider and returns text response"
    (api/register-tool! {:name        "echo"
                         :description "Echo the input"
                         :parameters  {:type "object"
                                       :properties {"msg" {:type "string"}}
                                       :required ["msg"]}
                         :fn          (fn [args] (str "echo: " (get args "msg")))})
    ;; Stub provider: returns text immediately (no tool calls)
    (g/set-provider-fn!
      (fn [opts]
        {:response   "done"
         :tokens-in  0
         :tokens-out 0}))
    (let [result (api/agent ["echo"] "test prompt")]
      (is (string? result))
      (is (= "done" result))))

  (testing "agent with no tools calls provider with empty tool-defs"
    (g/set-provider-fn!
      (fn [opts]
        (is (empty? (:tool-defs opts)))
        {:response "ok" :tokens-in 0 :tokens-out 0}))
    (is (= "ok" (api/agent [] "prompt"))))

  (testing "agent with opts map passes through options"
    (g/set-provider-fn!
      (fn [opts]
        (is (= "be helpful" (:system opts)))
        (is (= true (:agentic opts)))
        {:response "ok" :tokens-in 0 :tokens-out 0}))
    (is (= "ok" (api/agent {:system "be helpful"} [] "prompt"))))

  (testing "agent throws when tool not in registry"
    (g/set-provider-fn! (fn [_] {:response "" :tokens-in 0 :tokens-out 0}))
    (is (thrown-with-msg? clojure.lang.ExceptionInfo
                          #"tool not registered"
                          (api/agent ["missing-tool"] "prompt")))))

(deftest use-provider-test
  (testing "use-provider! sets the provider fn"
    (api/use-provider! "lmstudio")
    ;; provider fn should now be set (not nil)
    (is (some? @g/*provider-fn*)))
  (testing "use-provider! returns nil"
    (is (nil? (api/use-provider! "lmstudio")))))

(deftest use-model-test
  (testing "use-model! stores the model name"
    (api/use-model! "gemma4")
    (is (= "gemma4" @api/current-model)))
  (testing "use-model! returns nil"
    (is (nil? (api/use-model! "gemma4")))))
