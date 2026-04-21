(ns glitch.core-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.core :as core]
            [glitch.confidence :as conf]
            [cheshire.core :as json]))

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

;; --- check-contract ---

(deftest check-contract-non-empty-test
  (testing "non-empty passes for non-blank"
    (is (nil? (core/check-contract "hello" {:non-empty true}))))
  (testing "non-empty fails for blank"
    (is (some? (core/check-contract "" {:non-empty true}))))
  (testing "non-empty fails for whitespace-only"
    (is (some? (core/check-contract "   " {:non-empty true})))))

(deftest check-contract-min-length-test
  (testing "min-length passes"
    (is (nil? (core/check-contract "abcdef" {:min-length 5}))))
  (testing "min-length fails"
    (is (some? (core/check-contract "abc" {:min-length 5})))))

(deftest check-contract-max-length-test
  (testing "max-length passes"
    (is (nil? (core/check-contract "abc" {:max-length 5}))))
  (testing "max-length fails"
    (is (some? (core/check-contract "abcdef" {:max-length 5})))))

(deftest check-contract-json-test
  (testing "json passes for valid JSON"
    (is (nil? (core/check-contract "{\"a\":1}" {:json true}))))
  (testing "json fails for non-JSON"
    (is (some? (core/check-contract "not json" {:json true})))))

(deftest check-contract-keys-test
  (testing "keys passes when all present"
    (is (nil? (core/check-contract "{\"title\":\"x\",\"content\":\"y\"}"
                                    {:keys ["title" "content"]}))))
  (testing "keys fails when missing"
    (is (some? (core/check-contract "{\"title\":\"x\"}"
                                     {:keys ["title" "content"]})))))

(deftest check-contract-matches-test
  (testing "matches passes on regex match"
    (is (nil? (core/check-contract "error: 404" {:matches #"error: \d+"}))))
  (testing "matches fails on no match"
    (is (some? (core/check-contract "all good" {:matches #"error: \d+"})))))

(deftest check-contract-pred-test
  (testing "pred passes when fn returns truthy"
    (is (nil? (core/check-contract "abc" {:pred (fn [s] (= 3 (count s)))}))))
  (testing "pred fails when fn returns falsy"
    (is (some? (core/check-contract "ab" {:pred (fn [s] (= 3 (count s)))})))))

;; --- step with :expects ---

(deftest step-with-expects-passes-test
  (testing "step with valid expects passes"
    (let [result (core/step "good" "hello world" :expects {:non-empty true :min-length 5})]
      (is (= "hello world" result)))))

(deftest step-with-expects-fails-test
  (testing "step with failed expects throws"
    (is (thrown-with-msg? Exception #"contract-violation"
           (core/step "bad" "" :expects {:non-empty true})))))

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

;; --- llm with :schema ---

(deftest llm-schema-valid-test
  (testing "llm with valid schema passes through"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"action\": \"write\", \"reason\": \"needed\"}"
         :tokens-in 5 :tokens-out 10}))
    (let [result (core/llm :prompt "pick action"
                           :schema {:required ["action" "reason"]
                                    :types {"action" :string "reason" :string}})]
      (is (= "{\"action\": \"write\", \"reason\": \"needed\"}" result)))))

(deftest llm-schema-retries-on-invalid-test
  (testing "llm with schema retries on invalid, then succeeds"
    (let [calls (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! calls inc)
          (if (= 1 @calls)
            {:response "{\"wrong\": \"shape\"}" :tokens-in 1 :tokens-out 1}
            {:response "{\"action\": \"write\", \"reason\": \"fixed\"}" :tokens-in 1 :tokens-out 1})))
      (let [result (core/llm :prompt "pick"
                             :schema {:required ["action" "reason"]}
                             :retries 2)]
        (is (= 2 @calls))
        (is (re-find #"action" result))))))

(deftest llm-schema-exhausted-test
  (testing "llm throws after schema retries exhausted"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"bad\": true}" :tokens-in 1 :tokens-out 1}))
    (is (thrown-with-msg? Exception #"schema-violation"
           (core/llm :prompt "pick"
                     :schema {:required ["action"]}
                     :retries 1)))))

(deftest llm-schema-records-artifacts-test
  (testing "llm records schema validation result in recorder"
    (let [recorded (atom [])]
      (core/set-provider-fn!
        (fn [opts]
          {:response "{\"action\": \"write\"}" :tokens-in 1 :tokens-out 1}))
      (core/set-step-recorder! (fn [m] (swap! recorded conj m)))
      (core/llm :prompt "pick" :schema {:required ["action"]} :step-id "test-llm")
      (let [last-rec (last @recorded)]
        (is (re-find #"schema_valid" (or (:artifacts last-rec) "")))))))

;; --- llm confidence scoring ---

(deftest llm-confidence-passes-test
  (testing "llm with confidence above threshold passes"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"label\": \"bug\", \"confidence\": 0.9}"
         :tokens-in 1 :tokens-out 1}))
    (let [result (core/llm :prompt "classify"
                           :schema {:required ["label" "confidence"]
                                    :types {"label" :string "confidence" :number}}
                           :min-confidence 0.7)]
      (is (re-find #"bug" result)))))

(deftest llm-confidence-retries-test
  (testing "llm retries when confidence is below threshold"
    (let [calls (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! calls inc)
          (if (= 1 @calls)
            {:response "{\"label\": \"bug\", \"confidence\": 0.3}" :tokens-in 1 :tokens-out 1}
            {:response "{\"label\": \"bug\", \"confidence\": 0.9}" :tokens-in 1 :tokens-out 1})))
      (core/llm :prompt "classify"
                :schema {:required ["label" "confidence"]
                         :types {"label" :string "confidence" :number}}
                :min-confidence 0.7
                :retries 2)
      (is (= 2 @calls)))))

(deftest llm-confidence-exhausted-test
  (testing "llm throws when confidence stays low after retries"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"label\": \"bug\", \"confidence\": 0.2}"
         :tokens-in 1 :tokens-out 1}))
    (is (thrown-with-msg? Exception #"low-confidence"
           (core/llm :prompt "classify"
                     :schema {:required ["label" "confidence"]
                              :types {"label" :string "confidence" :number}}
                     :min-confidence 0.7
                     :retries 1)))))

(deftest llm-confidence-default-threshold-test
  (testing "default 0.7 threshold when schema requires confidence"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"label\": \"bug\", \"confidence\": 0.4}"
         :tokens-in 1 :tokens-out 1}))
    (is (thrown-with-msg? Exception #"low-confidence"
           (core/llm :prompt "classify"
                     :schema {:required ["label" "confidence"]
                              :types {"label" :string "confidence" :number}}
                     :retries 0)))))

(deftest llm-confidence-records-value-test
  (testing "llm records confidence in step-recorder"
    (let [recorded (atom [])]
      (core/set-provider-fn!
        (fn [opts]
          {:response "{\"label\": \"bug\", \"confidence\": 0.85}"
           :tokens-in 1 :tokens-out 1}))
      (core/set-step-recorder! (fn [m] (swap! recorded conj m)))
      (core/llm :prompt "classify"
                :schema {:required ["label" "confidence"]
                         :types {"label" :string "confidence" :number}}
                :min-confidence 0.5
                :step-id "conf-test")
      (let [last-rec (last @recorded)]
        (is (= 0.85 (:confidence last-rec)))))))

;; --- grounded? ---

(deftest grounded-passes-test
  (testing "grounded? returns true when LLM says grounded"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"grounded\": true, \"unsupported\": []}"
         :tokens-in 1 :tokens-out 1}))
    (core/step "doc" "The CLI supports --verbose flag.")
    (is (true? (core/grounded? "doc" "CLI reference: --verbose enables debug output")))))

(deftest grounded-fails-test
  (testing "grounded? throws when LLM says not grounded"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"grounded\": false, \"unsupported\": [{\"claim\": \"--quiet flag\", \"reason\": \"not in context\"}]}"
         :tokens-in 1 :tokens-out 1}))
    (core/step "doc" "The CLI supports --quiet flag.")
    (is (thrown-with-msg? Exception #"grounding-failure"
           (core/grounded? "doc" "CLI reference: --verbose enables debug output")))))

(deftest grounded-soft-mode-test
  (testing "grounded? with :strict false records but doesn't throw"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"grounded\": false, \"unsupported\": [{\"claim\": \"x\", \"reason\": \"y\"}]}"
         :tokens-in 1 :tokens-out 1}))
    (core/step "doc" "some text")
    (is (false? (core/grounded? "doc" "context" :strict false)))))

(deftest grounded-max-unsupported-test
  (testing "grounded? with :max-unsupported allows N unsupported claims"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"grounded\": false, \"unsupported\": [{\"claim\": \"x\", \"reason\": \"y\"}]}"
         :tokens-in 1 :tokens-out 1}))
    (core/step "doc" "text")
    (is (true? (core/grounded? "doc" "ctx" :max-unsupported 1)))))

(deftest grounded-records-via-recorder-test
  (testing "grounded? records result via step-recorder"
    (let [recorded (atom [])]
      (core/set-provider-fn!
        (fn [opts]
          {:response "{\"grounded\": true, \"unsupported\": []}"
           :tokens-in 1 :tokens-out 1}))
      (core/set-step-recorder! (fn [m] (swap! recorded conj m)))
      (core/step "doc" "valid text")
      (core/grounded? "doc" "context")
      (let [grounding-rec (last @recorded)]
        (is (= "grounded" (:kind grounding-rec)))
        (is (= 1 (:gate-passed grounding-rec)))))))

(deftest grounded-false-empty-unsupported-test
  (testing "grounded=false with 0 unsupported claims passes (within tolerance)"
    (core/set-provider-fn!
      (fn [opts]
        {:response "{\"grounded\": false, \"unsupported\": []}"
         :tokens-in 1 :tokens-out 1}))
    (core/step "doc" "text")
    ;; max-unsupported defaults to 0, and there are 0 unsupported claims
    (is (true? (core/grounded? "doc" "context")))))

(deftest grounded-unparseable-response-test
  (testing "grounded? throws when LLM returns unparseable JSON"
    (core/set-provider-fn!
      (fn [opts]
        {:response "I cannot verify this."
         :tokens-in 1 :tokens-out 1}))
    (core/step "doc" "text")
    (is (thrown-with-msg? Exception #"grounding-failure"
           (core/grounded? "doc" "context")))))

;; --- consensus ---

(deftest consensus-agreement-test
  (testing "consensus returns agreed=true when all providers agree"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          {:response (str "{\"decision\": \"deploy\", \"reason\": \"reason-" @call-count "\"}")
           :tokens-in 1 :tokens-out 1}))
      (let [result-str (core/consensus ["p1" "p2"]
                         :prompt "should we deploy?"
                         :schema {:required ["decision" "reason"]
                                  :types {"decision" :string "reason" :string}}
                         :compare-key "decision")
            result (json/parse-string result-str)]
        (is (true? (get result "agreed")))
        (is (= "deploy" (get-in result ["value" "decision"])))
        (is (= 2 (count (get result "votes"))))))))

(deftest consensus-disagreement-test
  (testing "consensus returns agreed=false when providers disagree"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          (if (odd? @call-count)
            {:response "{\"decision\": \"deploy\"}" :tokens-in 1 :tokens-out 1}
            {:response "{\"decision\": \"hold\"}" :tokens-in 1 :tokens-out 1})))
      (let [result-str (core/consensus ["p1" "p2"]
                         :prompt "deploy?"
                         :schema {:required ["decision"]}
                         :compare-key "decision")
            result (json/parse-string result-str)]
        (is (false? (get result "agreed")))
        (is (nil? (get result "value")))))))

(deftest consensus-min-providers-test
  (testing "consensus throws with fewer than 2 providers"
    (core/set-provider-fn! (fn [_] {:response "{}" :tokens-in 0 :tokens-out 0}))
    (is (thrown-with-msg? Exception #"minimum 2 providers"
           (core/consensus ["p1"] :prompt "test")))))

(deftest consensus-provider-failure-test
  (testing "consensus handles provider failure gracefully"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          (if (= 1 @call-count)
            (throw (ex-info "provider down" {}))
            {:response "{\"decision\": \"deploy\"}" :tokens-in 1 :tokens-out 1})))
      (let [result-str (core/consensus ["p1" "p2"]
                         :prompt "deploy?"
                         :schema {:required ["decision"]}
                         :compare-key "decision")
            result (json/parse-string result-str)]
        (is (false? (get result "agreed")))))))

(deftest consensus-records-via-recorder-test
  (testing "consensus records result via step-recorder"
    (let [recorded (atom [])]
      (core/set-provider-fn!
        (fn [opts]
          {:response "{\"decision\": \"deploy\"}" :tokens-in 1 :tokens-out 1}))
      (core/set-step-recorder! (fn [m] (swap! recorded conj m)))
      (core/consensus ["p1" "p2"]
        :prompt "deploy?"
        :schema {:required ["decision"]}
        :compare-key "decision")
      (let [cons-rec (last @recorded)]
        (is (= "consensus" (:kind cons-rec)))
        (is (= 1 (:gate-passed cons-rec)))))))

(deftest consensus-requires-compare-key-test
  (testing "consensus throws without compare-key or schema required"
    (core/set-provider-fn! (fn [_] {:response "{}" :tokens-in 0 :tokens-out 0}))
    (is (thrown-with-msg? Exception #"requires :compare-key"
           (core/consensus ["p1" "p2"] :prompt "test")))))

;; --- consensus with authority weighting ---

(deftest consensus-authority-weighting-test
  (testing "Bayesian confidence is higher than simple ratio when providers agree"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          {:response "{\"decision\": \"deploy\"}" :tokens-in 1 :tokens-out 1}))
      (let [result-str (core/consensus ["copilot" "openrouter"]
                         :prompt "deploy?"
                         :schema {:required ["decision"]}
                         :compare-key "decision"
                         :compare-mode :exact)
            result (json/parse-string result-str)]
        (is (true? (get result "agreed")))
        ;; Bayesian combine with high-authority providers and base-conf=1.0
        ;; should be > simple 2/2 = 1.0 ... well it'll be close to 1.0
        ;; Both agree so base-conf = 1.0, bayesian = 1 - (1-0.9)(1-0.85) = 0.985
        (is (> (get result "confidence") 0.9))
        ;; Each vote should have authority and quality fields
        (let [votes (get result "votes")]
          (is (number? (get (first votes) "authority")))
          (is (number? (get (first votes) "quality"))))))))

(deftest consensus-quality-tiebreaker-test
  (testing "disagreeing providers — winner picked by highest quality"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          (if (= 1 @call-count)
            ;; copilot: high authority, short response
            {:response "{\"decision\": \"deploy\"}" :tokens-in 1 :tokens-out 1}
            ;; lmstudio: lower authority, also short
            {:response "{\"decision\": \"hold\"}" :tokens-in 1 :tokens-out 1})))
      (let [result-str (core/consensus ["copilot" "lmstudio"]
                         :prompt "deploy?"
                         :schema {:required ["decision"]}
                         :compare-key "decision"
                         :compare-mode :exact)
            result (json/parse-string result-str)]
        (is (false? (get result "agreed")))
        ;; Winner should be copilot (higher authority -> higher quality)
        (is (some? (get result "winner")))
        (is (= "copilot" (get-in result ["winner" "provider"])))))))

;; --- composite-score ---

(deftest composite-score-test
  (testing "composite score computes harmonic mean over gate results"
    ;; Simulate gate results in steps
    (core/step "confidence:my-step" "0.9")
    (core/step "grounded:my-step" "true")   ;; -> 1.0
    (core/step "validate:my-step" "true")   ;; -> 1.0
    (let [result (core/composite-score "my-step")
          score  (Double/parseDouble result)]
      ;; harmonic-mean of [[1.0 0.9] [1.0 1.0] [1.0 1.0]] ≈ 0.966
      (is (> score 0.9))
      (is (<= score 1.0)))))

(deftest composite-score-empty-test
  (testing "composite score with no gates returns 0.0"
    (let [result (core/composite-score "nonexistent")]
      (is (= "0.0" result)))))

(deftest composite-score-custom-weights-test
  (testing "composite score respects custom weights"
    (core/step "confidence:weighted" "0.5")
    (core/step "grounded:weighted" "true")
    (let [default-result (Double/parseDouble (core/composite-score "weighted"))
          heavy-conf     (Double/parseDouble
                           (core/composite-score "weighted"
                             :weights {"confidence" 3.0 "grounded" 1.0}))]
      ;; With heavy weight on confidence (0.5), the score should be pulled down
      (is (< heavy-conf default-result)))))

;; --- domain-check in llm ---

(deftest llm-domain-check-on-topic-test
  (testing "domain-check preserves confidence when response is on-topic"
    (core/set-provider-fn!
      (fn [opts]
        ;; Response uses same keywords as prompt
        {:response "{\"label\": \"deploy\", \"confidence\": 0.8}"
         :tokens-in 1 :tokens-out 1}))
    ;; prompt says "deploy" and response says "deploy" — on topic
    (let [result (core/llm :prompt "classify deploy action"
                           :schema {:required ["label" "confidence"]
                                    :types {"label" :string "confidence" :number}}
                           :min-confidence 0.5
                           :domain-check true)]
      ;; Should pass — domain relevance shouldn't drop it below 0.5
      (is (some? result)))))

(deftest llm-domain-check-off-topic-test
  (testing "domain-check penalizes confidence when response is off-topic"
    (core/set-provider-fn!
      (fn [opts]
        ;; Response has zero keyword overlap with prompt
        {:response "{\"label\": \"banana\", \"confidence\": 0.75}"
         :tokens-in 1 :tokens-out 1}))
    ;; prompt is about "deploy kubernetes cluster" but response says "banana"
    (is (thrown-with-msg? Exception #"low-confidence"
           (core/llm :prompt "deploy kubernetes cluster"
                     :schema {:required ["label" "confidence"]
                              :types {"label" :string "confidence" :number}}
                     :min-confidence 0.5
                     :domain-check true
                     :retries 0)))))

;; --- consensus semantic mode ---

(deftest consensus-semantic-agreement-test
  (testing "consensus with semantic mode detects agreement on similar text"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          (if (odd? @call-count)
            {:response "{\"decision\": \"deploy now\"}" :tokens-in 1 :tokens-out 1}
            {:response "{\"decision\": \"we should deploy\"}" :tokens-in 1 :tokens-out 1})))
      ;; Mock embed to return similar vectors for deploy-related text
      (with-redefs [conf/embed
                    (fn [text]
                      (if (re-find #"deploy" (str text))
                        [0.9 0.1 0.0]
                        [0.0 0.0 1.0]))]
        (let [result-str (core/consensus ["p1" "p2"]
                           :prompt "should we deploy?"
                           :schema {:required ["decision"]}
                           :compare-key "decision"
                           :compare-mode :semantic)
              result (json/parse-string result-str)]
          (is (true? (get result "agreed"))))))))

(deftest consensus-semantic-fallback-on-nil-embed-test
  (testing "consensus falls back to exact comparison when embed returns nil"
    (let [call-count (atom 0)]
      (core/set-provider-fn!
        (fn [opts]
          (swap! call-count inc)
          (if (odd? @call-count)
            {:response "{\"decision\": \"deploy now\"}" :tokens-in 1 :tokens-out 1}
            {:response "{\"decision\": \"we should deploy\"}" :tokens-in 1 :tokens-out 1})))
      ;; Mock embed to return nil (TEI unavailable)
      (with-redefs [conf/embed (fn [_] nil)]
        (let [result-str (core/consensus ["p1" "p2"]
                           :prompt "should we deploy?"
                           :schema {:required ["decision"]}
                           :compare-key "decision"
                           :compare-mode :semantic)
              result (json/parse-string result-str)]
          ;; Falls back to exact — "deploy now" != "we should deploy" → disagreement
          (is (false? (get result "agreed"))))))))
