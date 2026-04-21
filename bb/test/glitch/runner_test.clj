(ns glitch.runner-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.runner :as runner]
            [glitch.core :as g]
            [glitch.provider :as prov]
            [clojure.java.io :as io]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Helpers
;; ---------------------------------------------------------------------------

(def ^:private test-dir "/tmp/glitch-runner-test")

(defn- write-temp-workflow
  "Write content to a temp workflow file. Returns the absolute path."
  [filename content]
  (let [path (str test-dir "/" filename)]
    (io/make-parents path)
    (spit path content)
    path))

(defn- cleanup []
  (let [d (io/file test-dir)]
    (when (.exists d)
      (doseq [f (reverse (file-seq d))]
        (.delete f)))))

;; ---------------------------------------------------------------------------
;; Fixtures
;; ---------------------------------------------------------------------------

(use-fixtures :each
  (fn [f]
    (cleanup)
    (.mkdirs (io/file test-dir))
    ;; Reset core + provider state
    (g/reset!)
    (g/set-input! "")
    (g/set-params! {})
    (g/set-step-recorder! nil)
    (g/set-provider-fn! nil)
    (reset! g/*call-stack* [])
    (g/set-workflows-dir! ".")
    (prov/reset!)
    (f)
    (cleanup)))

;; ---------------------------------------------------------------------------
;; Basic workflow evaluation
;; ---------------------------------------------------------------------------

(deftest simple-step-workflow-test
  (testing "run evaluates a workflow with step/ref and returns correct output and steps"
    (let [path   (write-temp-workflow "simple.glitch"
                   "(step \"hello\" \"world\")
                    (step \"greeting\" (str \"hi \" (ref \"hello\")))")
          result (runner/run path)]
      (is (= "hi world" (:output result)))
      (is (= "world" (get (:steps result) "hello")))
      (is (= "hi world" (get (:steps result) "greeting"))))))

(deftest workflow-with-input-test
  (testing "run passes input to the workflow"
    (let [path   (write-temp-workflow "with-input.glitch"
                   "(step \"echo\" (str \"got: \" (input)))")
          result (runner/run path :input "test-input")]
      (is (= "got: test-input" (:output result))))))

(deftest workflow-with-params-test
  (testing "run passes params to the workflow"
    (let [path   (write-temp-workflow "with-params.glitch"
                   "(step \"repo\" (param :repo \"fallback\"))
                    (step \"branch\" (param :branch \"main\"))")
          result (runner/run path :params {:repo "gl1tch" :branch "dev"})]
      (is (= "gl1tch" (get (:steps result) "repo")))
      (is (= "dev" (get (:steps result) "branch"))))))

(deftest workflow-with-seed-steps-test
  (testing "run pre-populates seed-steps before eval"
    (let [path   (write-temp-workflow "seeded.glitch"
                   "(step \"combined\" (str (ref \"pre\") \"-post\"))")
          result (runner/run path :seed-steps {"pre" "seeded"})]
      (is (= "seeded-post" (:output result))))))

;; ---------------------------------------------------------------------------
;; Workflow macro
;; ---------------------------------------------------------------------------

(deftest workflow-macro-test
  (testing "workflow macro works in SCI context"
    (let [path   (write-temp-workflow "macro.glitch"
                   "(workflow \"my-wf\"
                      :description \"a test\"
                      (step \"a\" \"alpha\")
                      (step \"b\" \"beta\"))")
          result (runner/run path)]
      ;; The workflow macro returns a map; the last expression in the file
      ;; is that map, but runner/run returns last-output from step tracking
      (is (= "beta" (:output result)))
      (is (= "alpha" (get (:steps result) "a")))
      (is (= "beta" (get (:steps result) "b"))))))

;; ---------------------------------------------------------------------------
;; LLM dispatch with mock provider
;; ---------------------------------------------------------------------------

(deftest llm-dispatch-test
  (testing "llm dispatches to a registered mock provider"
    (prov/register "mock"
      (fn [opts]
        {:response (str "mock: " (:prompt opts))
         :tokens-in 0
         :tokens-out 0}))
    (let [path   (write-temp-workflow "llm-test.glitch"
                   "(step \"answer\" (llm :prompt \"hello\" :provider \"mock\"))")
          result (runner/run path)]
      (is (= "mock: hello" (:output result))))))

(deftest llm-with-model-default-test
  (testing "llm uses the default model from run options"
    (prov/register "mock"
      (fn [opts]
        {:response (str "model=" (:model opts))
         :tokens-in 0
         :tokens-out 0}))
    (let [path   (write-temp-workflow "llm-model.glitch"
                   "(step \"m\" (llm :prompt \"test\" :provider \"mock\"))")
          result (runner/run path :model "gemma4")]
      (is (= "model=gemma4" (:output result))))))

;; ---------------------------------------------------------------------------
;; String and JSON namespace access
;; ---------------------------------------------------------------------------

(deftest str-namespace-test
  (testing "str/ namespace functions work in SCI"
    (let [path   (write-temp-workflow "str-ns.glitch"
                   "(step \"upper\" (str/upper-case \"hello\"))
                    (step \"joined\" (str/join \", \" [\"a\" \"b\" \"c\"]))")
          result (runner/run path)]
      (is (= "HELLO" (get (:steps result) "upper")))
      (is (= "a, b, c" (get (:steps result) "joined"))))))

(deftest json-namespace-test
  (testing "json/ namespace functions work in SCI"
    (let [path   (write-temp-workflow "json-ns.glitch"
                   "(step \"encoded\" (json/encode {:a 1 :b 2}))")
          result (runner/run path)]
      (is (str/includes? (:output result) "\"a\""))
      (is (str/includes? (:output result) "\"b\"")))))

;; ---------------------------------------------------------------------------
;; call-workflow (inter-workflow calls)
;; ---------------------------------------------------------------------------

(deftest call-workflow-test
  (testing "call-workflow evaluates a child workflow via SCI"
    (write-temp-workflow "child.glitch"
      "(step \"child-step\" (str \"child got: \" (input)))")
    (let [path   (write-temp-workflow "parent.glitch"
                   "(let [result (call-workflow \"child\" :input \"from-parent\")]
                      (step \"parent-step\" (str \"child said: \" (:output result))))")
          result (runner/run path :workflows-dir test-dir)]
      (is (= "child said: child got: from-parent" (:output result))))))

(deftest call-workflow-cycle-detection-test
  (testing "call-workflow detects cycles"
    (write-temp-workflow "cycle-a.glitch"
      "(call-workflow \"cycle-b\")")
    (write-temp-workflow "cycle-b.glitch"
      "(call-workflow \"cycle-a\")")
    (is (thrown-with-msg? Exception #"call-workflow cycle"
          (runner/run (str test-dir "/cycle-a.glitch")
                      :workflows-dir test-dir)))))

;; ---------------------------------------------------------------------------
;; par macro
;; ---------------------------------------------------------------------------

(deftest par-macro-test
  (testing "par executes forms concurrently in SCI"
    (let [path   (write-temp-workflow "par-test.glitch"
                   "(let [results (par
                                    (do (Thread/sleep 10) \"a\")
                                    (do (Thread/sleep 10) \"b\")
                                    \"c\")]
                      (step \"par-result\" (str/join \",\" results)))")
          result (runner/run path)]
      (is (= "a,b,c" (:output result))))))

;; ---------------------------------------------------------------------------
;; retry macro
;; ---------------------------------------------------------------------------

(deftest retry-macro-test
  (testing "retry succeeds after failures"
    (let [path   (write-temp-workflow "retry-test.glitch"
                   "(let [counter (atom 0)]
                      (step \"retried\"
                        (retry 3
                          (swap! counter inc)
                          (when (< @counter 2)
                            (throw (ex-info \"not yet\" {})))
                          \"success\")))")
          result (runner/run path)]
      (is (= "success" (:output result)))))

  (testing "retry throws after exhaustion"
    (let [path (write-temp-workflow "retry-fail.glitch"
                 "(retry 2 (throw (ex-info \"always fail\" {})))")]
      (is (thrown-with-msg? Exception #"retry exhausted"
            (runner/run path))))))

;; ---------------------------------------------------------------------------
;; with-timeout macro
;; ---------------------------------------------------------------------------

(deftest with-timeout-macro-test
  (testing "with-timeout returns when fast enough"
    (let [path   (write-temp-workflow "timeout-ok.glitch"
                   "(step \"fast\" (with-timeout 5 \"quick\"))")
          result (runner/run path)]
      (is (= "quick" (:output result)))))

  (testing "with-timeout throws on slow execution"
    (let [path (write-temp-workflow "timeout-fail.glitch"
                 "(with-timeout 1 (Thread/sleep 5000) \"slow\")")]
      (is (thrown-with-msg? Exception #"timeout"
            (runner/run path))))))

;; ---------------------------------------------------------------------------
;; phase macro
;; ---------------------------------------------------------------------------

(deftest phase-macro-test
  (testing "phase executes body and returns last value"
    (let [path   (write-temp-workflow "phase-test.glitch"
                   "(step \"result\" (phase \"build\"
                                       (step \"compile\" \"compiled\")
                                       \"done\"))")
          result (runner/run path)]
      (is (= "done" (:output result)))
      (is (= "compiled" (get (:steps result) "compile"))))))

;; ---------------------------------------------------------------------------
;; gate
;; ---------------------------------------------------------------------------

(deftest gate-test
  (testing "gate works in SCI context"
    (let [path   (write-temp-workflow "gate-test.glitch"
                   "(gate \"check\" true)
                    (step \"after-gate\" \"passed\")")
          result (runner/run path)]
      (is (= "passed" (:output result)))
      (is (= "true" (get (:steps result) "check"))))))

;; ---------------------------------------------------------------------------
;; sh
;; ---------------------------------------------------------------------------

(deftest sh-test
  (testing "sh runs shell commands in SCI"
    (let [path   (write-temp-workflow "sh-test.glitch"
                   "(step \"echo\" (sh \"echo\" \"hello world\"))")
          result (runner/run path)]
      (is (= "hello world" (:output result))))))

;; ---------------------------------------------------------------------------
;; list-workflows
;; ---------------------------------------------------------------------------

(deftest list-workflows-test
  (testing "list-workflows finds .glitch and .clj files"
    (write-temp-workflow "alpha.glitch" "(step \"a\" \"1\")")
    (write-temp-workflow "beta.clj" "(step \"b\" \"2\")")
    (write-temp-workflow "gamma.txt" "not a workflow")
    (let [wfs (runner/list-workflows test-dir)]
      (is (= 2 (count wfs)))
      (is (= "alpha" (:name (first wfs))))
      (is (= "beta" (:name (second wfs))))))

  (testing "list-workflows returns nil for non-existent dir"
    (is (nil? (runner/list-workflows "/tmp/nonexistent-glitch-dir")))))

;; ---------------------------------------------------------------------------
;; step :expects contracts in SCI
;; ---------------------------------------------------------------------------

(deftest step-expects-in-sci-test
  (testing "step with :expects works in SCI context"
    (let [path (write-temp-workflow "expects-pass.glitch"
                 "(step \"good\" \"hello world\" :expects {:non-empty true :min-length 5})")
          result (runner/run path)]
      (is (= "hello world" (:output result)))))

  (testing "step with :expects throws in SCI on violation"
    (let [path (write-temp-workflow "expects-fail.glitch"
                 "(step \"bad\" \"\" :expects {:non-empty true})")]
      (is (thrown-with-msg? Exception #"contract-violation"
            (runner/run path))))))

;; ---------------------------------------------------------------------------
;; Error handling
;; ---------------------------------------------------------------------------

(deftest workflow-error-test
  (testing "run propagates workflow errors"
    (let [path (write-temp-workflow "error.glitch"
                 "(throw (ex-info \"boom\" {:reason \"test\"}))" )]
      (is (thrown-with-msg? Exception #"boom"
            (runner/run path))))))
