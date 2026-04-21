# Workflow Confidence & Accuracy Framework — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 5 composable primitives (schema enforcement, step contracts, confidence scoring, grounding assertions, consensus) to the gl1tch Babashka workflow engine.

**Architecture:** Each primitive is a function in `bb/src/glitch/core.clj`, exposed as an SCI binding in `bb/src/glitch/runner.clj`, with results recorded via the existing step-recorder into the SQLite store. Primitives are independent — they compose but don't depend on each other.

**Tech Stack:** Babashka, SCI (Small Clojure Interpreter), Cheshire (JSON), go-sqlite3 pod

**Spec:** `docs/superpowers/specs/2026-04-20-workflow-confidence-design.md`

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `bb/src/glitch/store.clj` | Modify | Add `confidence REAL` column to steps DDL |
| `bb/src/glitch/core.clj` | Modify | Schema validation, step contracts, confidence, `validate`, `grounded?`, `consensus` |
| `bb/src/glitch/runner.clj` | Modify | Wire new SCI bindings + update `step` macro for `:expects` |
| `bb/test/glitch/core_test.clj` | Modify | Unit tests for all 5 primitives |
| `bb/test/glitch/store_test.clj` | Modify | Test confidence column persists |
| `bb/test/glitch/runner_test.clj` | Modify | Integration tests — SCI workflows using all primitives |

---

### Task 1: Store — Add confidence column

**Files:**
- Modify: `bb/src/glitch/store.clj:52-64` (steps DDL)
- Modify: `bb/src/glitch/store.clj:175-197` (record-step)
- Test: `bb/test/glitch/store_test.clj`

- [ ] **Step 1: Write the failing test**

Add to `bb/test/glitch/store_test.clj`:

```clj
(deftest test-step-confidence-column
  (testing "record-step persists confidence value"
    (let [run-id (store/record-run *db* {:name "conf-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "done"
                    :kind "llm" :confidence 0.85})
          steps  (store/get-steps *db* run-id)]
      (is (= 0.85 (:confidence (first steps))))))

  (testing "confidence is nil when not set"
    (let [run-id (store/record-run *db* {:name "no-conf" :input "y"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "done" :kind "step"})
          steps  (store/get-steps *db* run-id)]
      (is (nil? (:confidence (first steps)))))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — confidence column doesn't exist or isn't in INSERT

- [ ] **Step 3: Add confidence column to schema DDL**

In `bb/src/glitch/store.clj`, update the steps CREATE TABLE (line 52-65):

```clj
"CREATE TABLE IF NOT EXISTS steps (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id      INTEGER NOT NULL,
    step_id     TEXT NOT NULL,
    prompt      TEXT,
    output      TEXT,
    model       TEXT,
    duration_ms INTEGER,
    kind        TEXT,
    exit_status INTEGER,
    tokens_in   INTEGER,
    tokens_out  INTEGER,
    gate_passed INTEGER,
    artifacts   TEXT,
    confidence  REAL,
    UNIQUE(run_id, step_id)
  )"
```

- [ ] **Step 4: Update record-step to persist confidence**

In `bb/src/glitch/store.clj`, update `record-step` (line 176-197):

```clj
(defn record-step
  "Insert or replace a step record. `rec` is a map with keys:
     :run-id :step-id :prompt :output :model :duration :kind
     :exit :tokens-in :tokens-out :gate :artifacts :confidence"
  [db rec]
  (sql-execute! db
    ["INSERT OR REPLACE INTO steps
        (run_id, step_id, prompt, output, model,
         duration_ms, kind, exit_status,
         tokens_in, tokens_out, gate_passed, artifacts, confidence)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
     (:run-id rec)
     (:step-id rec)
     (:prompt rec)
     (:output rec)
     (:model rec)
     (:duration rec)
     (:kind rec)
     (:exit rec)
     (:tokens-in rec)
     (:tokens-out rec)
     (:gate rec)
     (:artifacts rec)
     (:confidence rec)]))
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS (including new confidence tests + all existing store tests)

- [ ] **Step 6: Commit**

```bash
git add bb/src/glitch/store.clj bb/test/glitch/store_test.clj
git commit -m "feat(store): add confidence REAL column to steps table"
```

---

### Task 2: Schema validation helpers in core.clj

**Files:**
- Modify: `bb/src/glitch/core.clj` (add `validate-schema` and `validate` functions after `json-extract`)
- Test: `bb/test/glitch/core_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/core_test.clj`:

```clj
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — `validate-schema` and `validate` don't exist

- [ ] **Step 3: Implement validate-schema**

Add to `bb/src/glitch/core.clj` after the `json-extract` function (after line 242):

```clj
(defn validate-schema
  "Validate a parsed JSON map against a schema.
   Returns nil on success, or a vector of violation strings on failure."
  [parsed schema]
  (let [violations (atom [])]
    ;; :required — check keys exist
    (doseq [k (:required schema)]
      (when-not (contains? parsed k)
        (swap! violations conj (str "missing required key: " k))))
    ;; :types — check value types
    (doseq [[k expected-type] (:types schema)]
      (when (contains? parsed k)
        (let [v (get parsed k)
              ok (case expected-type
                   :string  (string? v)
                   :number  (number? v)
                   :bool    (boolean? v)
                   :array   (vector? v)
                   :object  (map? v)
                   true)]
          (when-not ok
            (swap! violations conj (str "key '" k "' expected " (name expected-type)
                                        ", got " (type v)))))))
    ;; :enum — check allowed values
    (doseq [[k allowed] (:enum schema)]
      (when (contains? parsed k)
        (let [v (get parsed k)]
          (when-not (some #{v} allowed)
            (swap! violations conj (str "key '" k "' value '" v
                                        "' not in " (pr-str allowed)))))))
    (let [v @violations]
      (when (seq v) v))))
```

- [ ] **Step 4: Implement validate (standalone)**

Add to `bb/src/glitch/core.clj` right after `validate-schema`:

```clj
(defn validate
  "Validate a step's output against a schema. Returns true on success, throws on failure.
   Records result via step-recorder."
  [step-id schema]
  (let [output (ref step-id)
        extracted (json-extract (str output))
        parsed (try
                 (cheshire.core/parse-string extracted)
                 (catch Exception _
                   nil))
        violations (if (nil? parsed)
                     [(str "failed to parse JSON from step '" step-id "'")]
                     (validate-schema parsed schema))
        passed (nil? violations)]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id (str "validate:" step-id)
                 :kind "validate"
                 :output (str passed)
                 :gate-passed (if passed 1 0)
                 :artifacts (when violations
                              (cheshire.core/generate-string {:violations violations}))}))
    (if passed
      true
      (throw (ex-info (str "schema-violation: " (str/join "; " violations))
                      {:kind :schema-violation
                       :step-id step-id
                       :expected schema
                       :got extracted
                       :violations violations})))))
```

Also add `cheshire.core` to the ns require:

```clj
(ns glitch.core
  (:require [babashka.process :as bp]
            [clojure.string :as str]
            [clojure.java.io :as io]
            [cheshire.core :as json]))
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add bb/src/glitch/core.clj bb/test/glitch/core_test.clj
git commit -m "feat(core): add schema validation and standalone validate function"
```

---

### Task 3: Schema enforcement on llm (`:schema` + `:retries`)

**Files:**
- Modify: `bb/src/glitch/core.clj:191-209` (llm function)
- Test: `bb/test/glitch/core_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/core_test.clj`:

```clj
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
      ;; Last recorded entry should have artifacts
      (let [last-rec (last @recorded)]
        (is (re-find #"schema_valid" (or (:artifacts last-rec) "")))))))
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — llm doesn't handle `:schema` yet

- [ ] **Step 3: Implement schema enforcement in llm**

Replace the `llm` function in `bb/src/glitch/core.clj` (lines 191-209):

```clj
(defn llm [& {:keys [prompt model provider skill step-id schema retries
                      min-confidence] :as opts}]
  (when-not @*provider-fn*
    (throw (ex-info "llm: no provider function set — call set-provider-fn! first" {})))
  (let [full-prompt (if skill
                      (str (slurp skill) "\n\n" prompt)
                      prompt)
        max-retries (or retries 0)
        sid         (or step-id "llm")]
    (loop [attempt 0
           current-prompt full-prompt]
      (let [start    (System/nanoTime)
            result   (@*provider-fn* (assoc opts :prompt current-prompt))
            elapsed  (/ (- (System/nanoTime) start) 1e6)
            response (:response result)]
        ;; Schema validation (when :schema is provided)
        (if schema
          (let [extracted (json-extract (str response))
                parsed   (try (json/parse-string extracted) (catch Exception _ nil))
                violations (if (nil? parsed)
                             ["response is not valid JSON"]
                             (validate-schema parsed schema))
                valid?   (nil? violations)]
            (if valid?
              ;; Valid — record and return
              (do
                (when-let [recorder @*step-recorder*]
                  (recorder {:step-id sid :prompt current-prompt
                             :output response :model (or model "")
                             :duration (Math/round elapsed) :kind "llm"
                             :tokens-in (or (:tokens-in result) 0)
                             :tokens-out (or (:tokens-out result) 0)
                             :artifacts (json/generate-string {:schema_valid true})
                             :confidence (when (and parsed (contains? parsed "confidence"))
                                           (get parsed "confidence"))}))
                response)
              ;; Invalid — retry or throw
              (if (< attempt max-retries)
                (recur (inc attempt)
                       (str current-prompt "\n\nYour response didn't match the required schema: "
                            (str/join "; " violations) ". Please try again."))
                (do
                  (when-let [recorder @*step-recorder*]
                    (recorder {:step-id sid :prompt current-prompt
                               :output response :model (or model "")
                               :duration (Math/round elapsed) :kind "llm"
                               :tokens-in (or (:tokens-in result) 0)
                               :tokens-out (or (:tokens-out result) 0)
                               :artifacts (json/generate-string
                                            {:schema_valid false :violations violations})}))
                  (throw (ex-info (str "schema-violation: " (str/join "; " violations))
                                  {:kind :schema-violation
                                   :expected schema
                                   :got extracted
                                   :violations violations}))))))
          ;; No schema — pass through (original behavior)
          (do
            (when-let [recorder @*step-recorder*]
              (recorder {:step-id sid :prompt full-prompt
                         :output response :model (or model "")
                         :duration (Math/round elapsed) :kind "llm"
                         :tokens-in (or (:tokens-in result) 0)
                         :tokens-out (or (:tokens-out result) 0)}))
            response))))))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS (new schema tests + existing llm tests)

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/core.clj bb/test/glitch/core_test.clj
git commit -m "feat(core): schema enforcement on llm with retry"
```

---

### Task 4: Step contracts (`:expects` on step)

**Files:**
- Modify: `bb/src/glitch/core.clj:59-65` (step function)
- Modify: `bb/src/glitch/runner.clj:20-63` (SCI step macro)
- Test: `bb/test/glitch/core_test.clj`
- Test: `bb/test/glitch/runner_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/core_test.clj`:

```clj
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — `check-contract` doesn't exist, `step` doesn't accept `:expects`

- [ ] **Step 3: Implement check-contract**

Add to `bb/src/glitch/core.clj` after `validate`:

```clj
(defn check-contract
  "Check a step output against a contract map.
   Returns nil on success, or a vector of violation strings on failure."
  [output expects]
  (let [violations (atom [])]
    (when (and (:non-empty expects) (str/blank? output))
      (swap! violations conj "output is empty"))
    (when-let [min-len (:min-length expects)]
      (when (< (count output) min-len)
        (swap! violations conj (str "output length " (count output) " < min " min-len))))
    (when-let [max-len (:max-length expects)]
      (when (> (count output) max-len)
        (swap! violations conj (str "output length " (count output) " > max " max-len))))
    (when (:json expects)
      (let [extracted (json-extract output)]
        (when-not (try (json/parse-string extracted) true (catch Exception _ false))
          (swap! violations conj "output is not valid JSON"))))
    (when-let [required-keys (:keys expects)]
      (let [extracted (json-extract output)
            parsed (try (json/parse-string extracted) (catch Exception _ nil))]
        (if (nil? parsed)
          (swap! violations conj "output is not valid JSON (keys check)")
          (doseq [k required-keys]
            (when-not (contains? parsed k)
              (swap! violations conj (str "missing key: " k)))))))
    (when-let [pattern (:matches expects)]
      (when-not (re-find pattern output)
        (swap! violations conj (str "output does not match pattern: " pattern))))
    (when-let [pred-fn (:pred expects)]
      (when-not (pred-fn output)
        (swap! violations conj "custom predicate failed")))
    (let [v @violations]
      (when (seq v) v))))
```

- [ ] **Step 4: Update step to accept :expects**

Replace the `step` function in `bb/src/glitch/core.clj` (lines 59-65):

```clj
(defn step [id body & {:keys [expects]}]
  (let [val (str body)]
    ;; Check contract if :expects provided
    (when expects
      (let [violations (check-contract val expects)]
        (when violations
          (throw (ex-info (str "contract-violation: " (str/join "; " violations))
                          {:kind :contract-violation
                           :step-id id
                           :expected expects
                           :got val
                           :violations violations})))))
    (swap! *steps* assoc id val)
    (swap! *step-order* conj id)
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id id :output val :kind "step"
                 :artifacts (when expects
                              (json/generate-string {:contract "pass"}))}))
    val))
```

- [ ] **Step 5: Run tests to verify core tests pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 6: Update the SCI step macro in runner.clj**

The SCI `step` macro in `bb/src/glitch/runner.clj` sci-macros string needs to handle the keyword args after the body. Since `step` is now a regular function with `& {:keys [expects]}`, the existing SCI binding `'step g/step` already works — the SCI eval calls the host function directly. No macro change needed for `step` since it's bound as a function, not a macro.

Verify by adding a runner integration test to `bb/test/glitch/runner_test.clj`:

```clj
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
```

- [ ] **Step 7: Run all tests**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 8: Commit**

```bash
git add bb/src/glitch/core.clj bb/test/glitch/core_test.clj bb/test/glitch/runner_test.clj
git commit -m "feat(core): step contracts with :expects keyword"
```

---

### Task 5: Confidence scoring on llm

**Files:**
- Modify: `bb/src/glitch/core.clj` (llm function — already has schema loop from Task 3)
- Test: `bb/test/glitch/core_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/core_test.clj`:

```clj
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
    ;; No explicit :min-confidence, but schema requires "confidence" — default 0.7
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — confidence logic not yet in llm

- [ ] **Step 3: Add confidence enforcement to llm**

Update the valid-schema branch of the `llm` function's loop in `bb/src/glitch/core.clj`. After schema passes, before returning, add confidence check. Replace the valid? branch:

```clj
            (if valid?
              ;; Schema passed — check confidence
              (let [conf-val    (when (contains? parsed "confidence")
                                  (get parsed "confidence"))
                    threshold   (cond
                                  min-confidence min-confidence
                                  (and (some #{"confidence"} (:required schema))
                                       (nil? min-confidence))
                                  0.7
                                  :else nil)
                    conf-ok?    (or (nil? threshold)
                                   (nil? conf-val)
                                   (>= conf-val threshold))]
                (if conf-ok?
                  ;; All good — record and return
                  (do
                    (when-let [recorder @*step-recorder*]
                      (recorder {:step-id sid :prompt current-prompt
                                 :output response :model (or model "")
                                 :duration (Math/round elapsed) :kind "llm"
                                 :tokens-in (or (:tokens-in result) 0)
                                 :tokens-out (or (:tokens-out result) 0)
                                 :artifacts (json/generate-string {:schema_valid true})
                                 :confidence conf-val}))
                    response)
                  ;; Confidence too low — retry or throw
                  (if (< attempt max-retries)
                    (recur (inc attempt)
                           (str current-prompt "\n\nYour confidence was " conf-val
                                ", which is below the required " threshold
                                ". Think harder and be more specific."))
                    (do
                      (when-let [recorder @*step-recorder*]
                        (recorder {:step-id sid :prompt current-prompt
                                   :output response :model (or model "")
                                   :duration (Math/round elapsed) :kind "llm"
                                   :tokens-in (or (:tokens-in result) 0)
                                   :tokens-out (or (:tokens-out result) 0)
                                   :artifacts (json/generate-string
                                                {:schema_valid true :low_confidence true})
                                   :confidence conf-val}))
                      (throw (ex-info (str "low-confidence: " conf-val " < " threshold)
                                      {:kind :low-confidence
                                       :confidence conf-val
                                       :min threshold}))))))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/core.clj bb/test/glitch/core_test.clj
git commit -m "feat(core): confidence scoring on llm with threshold enforcement"
```

---

### Task 6: Grounding assertions — `grounded?`

**Files:**
- Modify: `bb/src/glitch/core.clj` (add `grounded?` function)
- Test: `bb/test/glitch/core_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/core_test.clj`:

```clj
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
    ;; 1 unsupported claim, max-unsupported 1 — should pass
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — `grounded?` doesn't exist

- [ ] **Step 3: Implement grounded?**

Add to `bb/src/glitch/core.clj` after `check-contract`:

```clj
(def ^:private grounding-prompt
  "You are a factual verification system. Compare the OUTPUT against the CONTEXT (ground truth).

Identify any claims, commands, features, or examples in the OUTPUT that are NOT directly supported by the CONTEXT.

Do not flag style issues, opinions, or reasonable inferences. Only flag factual claims that contradict or have no basis in the context.

CONTEXT:
%s

OUTPUT:
%s

Return JSON: {\"grounded\": true/false, \"unsupported\": [{\"claim\": \"exact text\", \"reason\": \"why unsupported\"}]}")

(defn grounded?
  "Verify that a step's output is factually supported by provided context.
   Options: :provider, :strict (default true), :max-unsupported (default 0)"
  [step-id context & {:keys [provider strict max-unsupported]
                       :or {strict true max-unsupported 0}}]
  (let [output   (ref step-id)
        prompt   (format grounding-prompt context output)
        response (@*provider-fn* {:prompt prompt :provider provider})
        extracted (json-extract (:response response))
        parsed   (try (json/parse-string extracted) (catch Exception _ nil))
        grounded (get parsed "grounded")
        unsupported (or (get parsed "unsupported") [])
        num-unsupported (count unsupported)
        passed   (or grounded (<= num-unsupported max-unsupported))]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id (str "grounded:" step-id)
                 :kind "grounded"
                 :output (str passed)
                 :gate-passed (if passed 1 0)
                 :artifacts (json/generate-string
                              {:grounded grounded :unsupported unsupported})}))
    (if (or passed (not strict))
      passed
      (throw (ex-info (str "grounding-failure: " num-unsupported " unsupported claims")
                      {:kind :grounding-failure
                       :step-id step-id
                       :unsupported unsupported})))))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/core.clj bb/test/glitch/core_test.clj
git commit -m "feat(core): grounding assertions with grounded? function"
```

---

### Task 7: Consensus — `consensus`

**Files:**
- Modify: `bb/src/glitch/core.clj` (add `consensus` function)
- Test: `bb/test/glitch/core_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/core_test.clj`:

```clj
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
      ;; Only 1/2 succeeds — not enough for consensus
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — `consensus` doesn't exist

- [ ] **Step 3: Implement consensus**

Add to `bb/src/glitch/core.clj` after `grounded?`:

```clj
(defn consensus
  "Run the same prompt through multiple providers and compare responses.
   Returns a JSON string with {agreed, value, votes}.
   Options: :prompt (required), :schema, :compare-key, :model, :min-confidence"
  [providers & {:keys [prompt schema compare-key model min-confidence]}]
  (when (< (count providers) 2)
    (throw (ex-info "consensus requires minimum 2 providers" {:providers providers})))
  (let [cmp-key  (or compare-key (first (:required schema)))
        ;; Call each provider (sequentially for simplicity — par is a macro, not composable here)
        votes    (reduce
                   (fn [acc pname]
                     (try
                       (let [response (@*provider-fn*
                                        (cond-> {:prompt prompt :provider pname}
                                          model (assoc :model model)))
                             extracted (json-extract (:response response))
                             parsed    (try (json/parse-string extracted) (catch Exception _ nil))
                             violations (when (and schema parsed)
                                          (validate-schema parsed schema))]
                         (if (and parsed (nil? violations))
                           (conj acc {:provider pname :response parsed})
                           acc))
                       (catch Exception _
                         acc)))
                   []
                   providers)
        ;; Compare on cmp-key
        values   (when cmp-key
                   (map #(-> % :response (get cmp-key)
                             str str/lower-case str/trim)
                        votes))
        agreed   (and (>= (count votes) 2)
                      (apply = values))
        result   {"agreed" agreed
                  "value"  (when agreed (:response (first votes)))
                  "votes"  (mapv (fn [v] {"provider" (:provider v)
                                          "response" (:response v)})
                                 votes)}
        confidence (if (empty? votes) 0.0
                     (/ (count (filter #(= (first values)
                                           (-> % :response (get cmp-key)
                                               str str/lower-case str/trim))
                                       votes))
                        (double (count votes))))]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id "consensus"
                 :kind "consensus"
                 :output (json/generate-string result)
                 :gate-passed (if agreed 1 0)
                 :confidence confidence
                 :artifacts (json/generate-string {:votes (get result "votes")})}))
    (json/generate-string result)))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/core.clj bb/test/glitch/core_test.clj
git commit -m "feat(core): consensus multi-provider voting"
```

---

### Task 8: Wire SCI bindings in runner.clj

**Files:**
- Modify: `bb/src/glitch/runner.clj:110-154` (make-sci-ctx user-ns)
- Test: `bb/test/glitch/runner_test.clj`

- [ ] **Step 1: Write the failing tests**

Add to `bb/test/glitch/runner_test.clj`:

```clj
;; ---------------------------------------------------------------------------
;; Confidence framework SCI bindings
;; ---------------------------------------------------------------------------

(deftest validate-in-sci-test
  (testing "validate works in SCI context"
    (prov/register "mock"
      (fn [opts] {:response "{\"action\": \"write\"}" :tokens-in 0 :tokens-out 0}))
    (let [path (write-temp-workflow "validate-sci.glitch"
                 "(step \"data\" (json/encode {\"pass\" true \"issues\" []}))
                  (validate \"data\" {:required [\"pass\" \"issues\"]
                                      :types {\"pass\" :bool \"issues\" :array}})
                  (step \"ok\" \"validated\")")
          result (runner/run path)]
      (is (= "validated" (:output result))))))

(deftest grounded-in-sci-test
  (testing "grounded? works in SCI context"
    (prov/register "mock"
      (fn [opts]
        {:response "{\"grounded\": true, \"unsupported\": []}"
         :tokens-in 0 :tokens-out 0}))
    (let [path (write-temp-workflow "grounded-sci.glitch"
                 "(step \"doc\" \"The CLI has --verbose.\")
                  (grounded? \"doc\" \"CLI ref: --verbose flag exists\")
                  (step \"ok\" \"grounded\")")
          result (runner/run path :tiers [{:providers ["mock"]}])]
      (is (= "grounded" (:output result))))))

(deftest consensus-in-sci-test
  (testing "consensus works in SCI context"
    (prov/register "mock"
      (fn [opts]
        {:response "{\"decision\": \"deploy\", \"reason\": \"looks good\"}"
         :tokens-in 0 :tokens-out 0}))
    (let [path (write-temp-workflow "consensus-sci.glitch"
                 "(step \"verdict\"
                    (consensus [\"mock\" \"mock\"]
                      :prompt \"deploy?\"
                      :schema {:required [\"decision\" \"reason\"]}
                      :compare-key \"decision\"))")
          result (runner/run path :tiers [{:providers ["mock"]}])]
      (is (re-find #"agreed" (:output result))))))

(deftest llm-schema-in-sci-test
  (testing "llm with :schema works in SCI context"
    (prov/register "mock"
      (fn [opts]
        {:response "{\"action\": \"write\", \"reason\": \"needed\"}"
         :tokens-in 0 :tokens-out 0}))
    (let [path (write-temp-workflow "schema-sci.glitch"
                 "(step \"router\"
                    (llm :prompt \"pick action\"
                         :provider \"mock\"
                         :schema {:required [\"action\" \"reason\"]
                                  :types {\"action\" :string \"reason\" :string}}))")
          result (runner/run path)]
      (is (re-find #"action" (:output result))))))
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: FAIL — `validate`, `grounded?`, `consensus` not in SCI context

- [ ] **Step 3: Add SCI bindings**

In `bb/src/glitch/runner.clj`, update `make-sci-ctx` user-ns (around line 110-154). Add after the `'json-extract` binding:

```clj
                 ;; confidence framework
                 'validate       g/validate
                 'validate-schema g/validate-schema
                 'check-contract g/check-contract
                 'grounded?      g/grounded?
                 'consensus      g/consensus
```

- [ ] **Step 4: Run all tests**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/runner.clj bb/test/glitch/runner_test.clj
git commit -m "feat(runner): wire confidence framework SCI bindings"
```

---

### Task 9: Integration test — full workflow exercising all 5 primitives

**Files:**
- Test: `bb/test/glitch/runner_test.clj`

- [ ] **Step 1: Write the integration test**

Add to `bb/test/glitch/runner_test.clj`:

```clj
;; ---------------------------------------------------------------------------
;; Full integration — all 5 confidence primitives in one workflow
;; ---------------------------------------------------------------------------

(deftest full-confidence-integration-test
  (testing "workflow uses schema, contract, confidence, grounding, and consensus"
    (let [call-count (atom 0)]
      (prov/register "mock"
        (fn [opts]
          (swap! call-count inc)
          (let [p (:prompt opts)]
            (cond
              ;; Grounding verification prompt
              (re-find #"factual verification" (str p))
              {:response "{\"grounded\": true, \"unsupported\": []}"
               :tokens-in 5 :tokens-out 10}

              ;; Default — schema-valid response with confidence
              :else
              {:response "{\"decision\": \"deploy\", \"confidence\": 0.9, \"reason\": \"all clear\"}"
               :tokens-in 5 :tokens-out 10}))))
      (let [path (write-temp-workflow "full-confidence.glitch"
                   ";; 1. Schema + confidence on LLM
                    (step \"analysis\"
                      (llm :prompt \"Analyze this change\"
                           :provider \"mock\"
                           :schema {:required [\"decision\" \"confidence\" \"reason\"]
                                    :types {\"decision\" :string \"confidence\" :number \"reason\" :string}
                                    :enum {\"decision\" [\"deploy\" \"hold\" \"rollback\"]}}
                           :min-confidence 0.7)
                      :expects {:non-empty true :min-length 10})

                    ;; 2. Validate the output schema standalone
                    (validate \"analysis\" {:required [\"decision\"]})

                    ;; 3. Ground the analysis against context
                    (grounded? \"analysis\" \"Change log: all tests pass, no breaking changes\"
                               :strict false)

                    ;; 4. Consensus from two providers
                    (step \"verdict\"
                      (consensus [\"mock\" \"mock\"]
                        :prompt \"Final call?\"
                        :schema {:required [\"decision\"]}
                        :compare-key \"decision\"))")
            result (runner/run path :tiers [{:providers ["mock"]}])]
        ;; Verify the workflow completed
        (is (some? (:output result)))
        ;; Verify steps were recorded
        (is (some? (get (:steps result) "analysis")))
        (is (some? (get (:steps result) "verdict")))
        ;; Consensus result should show agreement
        (is (re-find #"agreed" (get (:steps result) "verdict")))))))
```

- [ ] **Step 2: Run all tests**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add bb/test/glitch/runner_test.clj
git commit -m "test: full integration test for confidence framework"
```

---

### Task 10: Wire test namespace in test runner

**Files:**
- Modify: `bb/src/glitch/test_runner.clj` (no new ns needed — all tests are in existing files)

This task is a no-op since all tests were added to existing test files (`core_test.clj`, `store_test.clj`, `runner_test.clj`). The test runner already discovers them. Verify:

- [ ] **Step 1: Run full test suite**

Run: `cd bb && bb -cp src:test:providers -m glitch.test-runner`
Expected: ALL PASS, test count increased from baseline

- [ ] **Step 2: Delete any test .db files left over**

Run: `rm -f /tmp/glitch-test-*.db /tmp/glitch-test-*.db-wal /tmp/glitch-test-*.db-shm`
