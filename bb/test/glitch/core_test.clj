(ns glitch.core-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.core :as core]))

;; Reset state before each test
(use-fixtures :each
  (fn [f]
    (core/reset!)
    (core/set-input! "")
    (core/set-params! {})
    (core/set-step-recorder! nil)
    (core/set-provider-fn! nil)
    (reset! core/*call-stack* [])
    (core/set-workflows-dir! ".")
    (f)))

;; --- step / ref ---

(deftest step-and-ref-test
  (testing "step records value and ref retrieves it"
    (core/step :greeting "hello world")
    (is (= "hello world" (core/ref :greeting))))

  (testing "step converts non-string to string"
    (core/step :num 42)
    (is (= "42" (core/ref :num))))

  (testing "ref returns nil for unknown step"
    (is (nil? (core/ref :nonexistent)))))

;; --- input / params ---

(deftest input-test
  (testing "set-input! and input round-trip"
    (core/set-input! "some input text")
    (is (= "some input text" (core/input)))))

(deftest params-test
  (testing "set-params! and params round-trip"
    (core/set-params! {:repo "gl1tch" :branch "main"})
    (is (= {:repo "gl1tch" :branch "main"} (core/params))))

  (testing "param retrieves by keyword"
    (core/set-params! {:repo "gl1tch"})
    (is (= "gl1tch" (core/param :repo))))

  (testing "param returns default when missing"
    (core/set-params! {})
    (is (= "fallback" (core/param :missing "fallback")))))

;; --- sh ---

(deftest sh-test
  (testing "sh runs a command and returns trimmed stdout"
    (is (= "hello" (core/sh "echo" "hello"))))

  (testing "sh throws on non-zero exit"
    (is (thrown? Exception (core/sh "false")))))

;; --- workflow macro ---

(deftest workflow-test
  (testing "workflow macro returns name, output, and steps"
    (let [result (core/workflow "test-wf"
                   :description "a test workflow"
                   (core/step :a "alpha")
                   (core/step :b "beta"))]
      (is (= "test-wf" (:name result)))
      (is (= "beta" (:output result)))
      (is (= "alpha" (get (:steps result) :a)))
      (is (= "beta" (get (:steps result) :b))))))

;; --- gate ---

(deftest gate-test
  (testing "gate returns truthy predicate"
    (is (true? (core/gate :check true))))

  (testing "gate returns falsy predicate"
    (is (false? (core/gate :check2 false))))

  (testing "gate records in steps"
    (core/gate :g1 true)
    (is (= "true" (core/ref :g1))))

  (testing "gate calls step-recorder when set"
    (let [recorded (atom nil)]
      (core/set-step-recorder! (fn [m] (reset! recorded m)))
      (core/gate :g2 false)
      (is (= "gate" (:kind @recorded)))
      (is (= 0 (:gate-passed @recorded))))))

;; --- retry ---

(deftest retry-test
  (testing "retry succeeds on first attempt"
    (is (= "ok" (core/retry 3 "ok"))))

  (testing "retry retries on failure then succeeds"
    (let [counter (atom 0)]
      (is (= "done"
             (core/retry 3
               (swap! counter inc)
               (when (< @counter 3)
                 (throw (ex-info "not yet" {})))
               "done")))))

  (testing "retry throws after exhaustion"
    (is (thrown? Exception
                (core/retry 2
                  (throw (ex-info "always fail" {})))))))

;; --- last-output ---

(deftest last-output-test
  (testing "last-output returns empty string when no steps"
    (is (= "" (core/last-output))))

  (testing "last-output returns the most recent step"
    (core/step :first "one")
    (core/step :second "two")
    (is (= "two" (core/last-output)))))

;; --- call-workflow cycle detection ---

(deftest call-workflow-cycle-test
  (testing "cycle detection throws"
    ;; Simulate a workflow already on the stack
    (swap! core/*call-stack* conj "wf-a")
    (is (thrown-with-msg? Exception #"call-workflow cycle"
                         (core/call-workflow "wf-a")))))

;; --- call-workflow file not found ---

(deftest call-workflow-not-found-test
  (testing "throws when workflow file doesn't exist"
    (is (thrown-with-msg? Exception #"not found"
                         (core/call-workflow "nonexistent-workflow")))))

;; --- par ---

(deftest par-test
  (testing "par executes forms concurrently and collects results"
    (let [results (core/par
                    (do (Thread/sleep 10) "a")
                    (do (Thread/sleep 10) "b")
                    "c")]
      (is (= ["a" "b" "c"] results)))))

;; --- with-timeout ---

(deftest with-timeout-test
  (testing "with-timeout returns value when fast enough"
    (is (= "fast" (core/with-timeout 2 "fast"))))

  (testing "with-timeout throws on slow execution"
    (is (thrown-with-msg? Exception #"timeout"
                         (core/with-timeout 1
                           (Thread/sleep 5000)
                           "slow")))))

;; --- phase ---

(deftest phase-test
  (testing "phase executes body and returns last value"
    (is (= "done" (core/phase "build"
                    (core/step :compile "compiled")
                    "done"))))

  (testing "phase records start and end via step-recorder"
    (let [records (atom [])]
      (core/set-step-recorder! (fn [m] (swap! records conj m)))
      (core/phase "deploy"
        (core/step :ship "shipped"))
      ;; Should have: phase start, step record, phase end = 3 records
      (is (= 3 (count @records)))
      (is (= "started" (:output (first @records))))
      (is (= "phase" (:kind (first @records))))
      (is (= "phase" (:kind (last @records)))))))

;; --- step-recorder integration ---

(deftest step-recorder-test
  (testing "step-recorder is called for each step"
    (let [records (atom [])]
      (core/set-step-recorder! (fn [m] (swap! records conj m)))
      (core/step :x "val-x")
      (core/step :y "val-y")
      (is (= 2 (count @records)))
      (is (= "step" (:kind (first @records))))
      (is (= "val-x" (:output (first @records)))))))

;; --- save / read-file ---

(deftest save-and-read-file-test
  (testing "save creates file with parent dirs, read-file reads it back"
    (let [path "/tmp/glitch-test/nested/dir/test.txt"]
      (core/save path "hello file")
      (is (= "hello file" (core/read-file path)))
      ;; cleanup
      (clojure.java.io/delete-file path true))))

;; --- get-steps ---

(deftest get-steps-test
  (testing "get-steps returns all recorded steps"
    (core/step :a "1")
    (core/step :b "2")
    (let [steps (core/get-steps)]
      (is (= "1" (get steps :a)))
      (is (= "2" (get steps :b))))))

;; --- llm ---

(deftest llm-throws-without-provider-test
  (testing "llm throws when no provider is set"
    (is (thrown-with-msg? Exception #"no provider function set"
                         (core/llm :prompt "hello")))))

(deftest llm-calls-provider-test
  (testing "llm calls provider-fn and returns response"
    (core/set-provider-fn!
      (fn [opts]
        {:response (str "echo: " (:prompt opts))
         :tokens-in 5
         :tokens-out 10}))
    (is (= "echo: hello" (core/llm :prompt "hello")))))

(deftest llm-records-via-step-recorder-test
  (testing "llm records call via step-recorder"
    (let [recorded (atom nil)]
      (core/set-provider-fn!
        (fn [opts]
          {:response "answer" :tokens-in 1 :tokens-out 2}))
      (core/set-step-recorder! (fn [m] (reset! recorded m)))
      (core/llm :prompt "question" :model "test-model")
      (is (= "llm" (:kind @recorded)))
      (is (= "question" (:prompt @recorded)))
      (is (= "answer" (:output @recorded)))
      (is (= "test-model" (:model @recorded))))))

;; --- json-extract ---

(deftest json-extract-plain-test
  (is (= "{\"a\":1}" (core/json-extract "{\"a\":1}"))))

(deftest json-extract-with-prefix-test
  (is (= "{\"action\":\"dev\"}"
         (core/json-extract "Here is the result:\n{\"action\":\"dev\"}"))))

(deftest json-extract-with-markdown-fence-test
  (is (= "{\"x\":1}"
         (core/json-extract "```json\n{\"x\":1}\n```"))))

(deftest json-extract-nested-test
  (is (= "{\"a\":{\"b\":2}}"
         (core/json-extract "thinking...\n{\"a\":{\"b\":2}}\ndone"))))

(deftest json-extract-with-braces-in-strings-test
  (is (= "{\"msg\":\"{hello}\"}"
         (core/json-extract "prefix {\"msg\":\"{hello}\"} suffix"))))

(deftest json-extract-array-test
  (is (= "[1,2,3]"
         (core/json-extract "result: [1,2,3]"))))

;; --- validate-schema ---

(deftest validate-schema-valid-test
  (testing "valid JSON passes schema"
    (is (nil? (core/validate-schema
                {"action" "write" "reason" "because"}
                {:required ["action" "reason"]
                 :types {"action" :string "reason" :string}})))))

(deftest validate-schema-missing-key-test
  (testing "missing required key returns violation"
    (let [v (core/validate-schema {"action" "write"} {:required ["action" "reason"]})]
      (is (some? v))
      (is (re-find #"reason" (first v))))))

(deftest validate-schema-wrong-type-test
  (testing "wrong type returns violation"
    (let [v (core/validate-schema {"count" "not-a-number"} {:types {"count" :number}})]
      (is (some? v)))))

(deftest validate-schema-enum-test
  (testing "value not in enum returns violation"
    (let [v (core/validate-schema
              {"action" "delete"}
              {:enum {"action" ["write" "edit" "skip"]}})]
      (is (some? v))))

  (testing "value in enum passes"
    (is (nil? (core/validate-schema
                {"action" "write"}
                {:enum {"action" ["write" "edit" "skip"]}})))))

(deftest validate-schema-bool-type-test
  (testing "bool type check"
    (is (nil? (core/validate-schema {"pass" true} {:types {"pass" :bool}})))
    (is (some? (core/validate-schema {"pass" "yes"} {:types {"pass" :bool}})))))

(deftest validate-schema-array-type-test
  (testing "array type check"
    (is (nil? (core/validate-schema {"items" [1 2 3]} {:types {"items" :array}})))
    (is (some? (core/validate-schema {"items" "nope"} {:types {"items" :array}})))))

(deftest validate-schema-object-type-test
  (testing "object type check"
    (is (nil? (core/validate-schema {"meta" {"a" 1}} {:types {"meta" :object}})))
    (is (some? (core/validate-schema {"meta" 42} {:types {"meta" :object}})))))

;; --- validate (standalone) ---

(deftest validate-fn-passes-test
  (testing "validate returns true when step output matches schema"
    (core/step "data" "{\"pass\": true, \"issues\": []}")
    (is (true? (core/validate "data" {:required ["pass" "issues"]
                                       :types {"pass" :bool "issues" :array}})))))

(deftest validate-fn-throws-test
  (testing "validate throws when step output fails schema"
    (core/step "data" "{\"pass\": true}")
    (is (thrown-with-msg? Exception #"schema-violation"
           (core/validate "data" {:required ["pass" "issues"]})))))

(deftest validate-fn-non-json-test
  (testing "validate throws when step output is not JSON"
    (core/step "plain" "just text")
    (is (thrown-with-msg? Exception #"schema-violation"
           (core/validate "plain" {:required ["key"]})))))
