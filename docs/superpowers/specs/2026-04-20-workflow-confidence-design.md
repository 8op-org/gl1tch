# Workflow Confidence & Accuracy Framework

**Date:** 2026-04-20  
**Status:** Design approved  
**Scope:** 5 new core primitives for gl1tch workflow engine

## Overview

Five independent, composable primitives that improve accuracy and confidence in workflow outputs. Each is a function in `core.clj`, exposed as an SCI binding in `runner.clj`, and recorded in the SQLite store.

Design principles:
- **Layered primitives** — each mechanism is independently usable, they compose but don't depend on each other
- **Fail fast** — violations throw immediately at the boundary where they're detected
- **Engine records, workflow decides** — the engine validates and persists metrics; workflows control orchestration and recovery
- **No invisible prompt mutation** — the engine never injects text into prompts without the workflow author's knowledge

## 1. Schema Enforcement

### 1.1 `llm` with `:schema`

Validates LLM JSON output against a structural schema. Retries on mismatch.

```clj
(step "router"
  (llm :prompt "Pick an action..."
       :schema {:required ["action" "reason"]
                :types {"action" :string "reason" :string}
                :enum {"action" ["write" "edit" "skip"]}}
       :retries 2))
```

**Behavior:**

1. Call provider as normal
2. Run `json-extract` on response
3. Validate extracted JSON against `:schema`
4. If invalid and retries remain: re-call with appended error message ("your response didn't match: missing key 'action'")
5. If invalid and retries exhausted: throw `ex-info` with `{:kind :schema-violation :expected schema :got parsed}`
6. If valid: return the raw string response (step output stays a string)
7. Record validation result in `artifacts` column: `{"schema_valid": true}` or `{"schema_valid": false, "violations": [...]}`

### 1.2 Standalone `(validate step-id schema)`

Validates any step's existing output against a schema. No retry.

```clj
(step "merged" (json/encode ...))
(validate "merged" {:required ["pass" "issues"] :types {"pass" :bool "issues" :array}})
```

**Behavior:**

- Reads `(ref step-id)`, runs `json-extract`, validates against schema
- Returns `true` on success, throws on failure (same ex-info shape)
- Records in step-recorder as `{:kind "validate" :step-id step-id :gate-passed 1/0}`
- No retry — checking existing data, not generating new data

### 1.3 Schema shape

Minimal, purpose-built. Not JSON Schema.

```clj
{:required ["key1" "key2"]                              ;; keys that must exist
 :types {"key" :string|:bool|:number|:array|:object}    ;; type checks
 :enum {"key" ["allowed" "values"]}}                    ;; allowed values
```

No nesting, no conditionals, no `$ref`. Covers 95% of LLM output validation needs without the complexity.

## 2. Step Contracts

### 2.1 `step` with `:expects`

Validates step output immediately after evaluation. Fails fast at the step boundary.

```clj
(step "write-page"
  (llm :prompt "Generate the page content...")
  :expects {:non-empty true
            :min-length 100
            :json true
            :keys ["title" "content"]})
```

**Implementation note:** The `step` function signature changes from `(defn step [id body])` to `(defn step [id body & {:keys [expects]}])`. The SCI macro for `step` in `runner.clj` must also be updated to pass keyword args through. Since `step` is called as `(step "id" expr :expects {...})`, the expr evaluates first (positional), then keyword args are extracted.

**Behavior:**

1. Step body evaluates, produces string output as today
2. If `:expects` is present, engine checks the output immediately
3. On violation: throw `ex-info` with `{:kind :contract-violation :step-id id :expected expects :got output}`
4. On pass: record normally, add `{"contract": "pass"}` to `artifacts`
5. No retry — contracts catch bugs. Use `(retry 2 (step ...))` for retry.

### 2.2 Contract predicates

| Key | Check |
|---|---|
| `:non-empty` | `(not (str/blank? output))` |
| `:min-length` | `(>= (count output) n)` |
| `:max-length` | `(<= (count output) n)` |
| `:json` | `json-extract` succeeds and parses |
| `:keys` | parsed JSON contains all listed keys |
| `:matches` | regex pattern matches output |
| `:pred` | custom `(fn [output] bool)` — escape hatch |

### 2.3 Interaction with `:schema`

`:schema` on `llm` validates inside the LLM call (with retry). `:expects` on `step` validates after the step body (no retry). They stack:

```clj
(step "router"
  (llm :prompt "..." :schema {:required ["action"]})  ;; LLM retries on bad schema
  :expects {:min-length 10})                          ;; step fails if still too short
```

Different failure modes, different concerns. Schema catches structural problems at generation time. Contracts catch output-level problems at the step boundary.

## 3. Confidence Scoring

### 3.1 `llm` with `:min-confidence`

Extracts and enforces a confidence score from LLM responses.

```clj
(step "classify"
  (llm :prompt "Classify this issue... Include a confidence score 0.0-1.0."
       :schema {:required ["label" "confidence"]
                :types {"label" :string "confidence" :number}}
       :min-confidence 0.7
       :retries 2))
```

**Behavior:**

1. LLM returns JSON with a `"confidence"` field (workflow author prompts for it)
2. After schema validation passes, engine extracts `confidence` from parsed response
3. If `confidence < :min-confidence`: retry with "your confidence was 0.4, think harder and be more specific"
4. If still below threshold after retries: throw `{:kind :low-confidence :confidence 0.4 :min 0.7}`
5. Record `confidence` value in the `confidence` REAL column in `steps`

### 3.2 Default behavior

- `:min-confidence` defaults to `0.7` only when `:schema` includes `"confidence"` in `:required`
- If you don't ask for confidence in the schema, the engine doesn't enforce it
- Override with `:min-confidence 0.0` to record but not enforce

### 3.3 Key constraint

The engine never injects "include a confidence score" into prompts. The workflow author decides when confidence matters and prompts for it explicitly. The engine only validates and records.

### 3.4 Store changes

```sql
ALTER TABLE steps ADD COLUMN confidence REAL;
```

Recorded on every step that returns a confidence value. `NULL` for steps without it.

### 3.5 Querying trends

```sql
SELECT avg(confidence), min(confidence) 
FROM steps WHERE step_id = 'router' AND confidence IS NOT NULL;
```

No new API — just SQLite queries over accumulated data.

## 4. Grounding Assertions

### 4.1 `(grounded? step-id context)`

Verifies that a step's output is factually supported by provided ground truth context.

```clj
(step "write-docs"
  (llm :prompt "Write documentation for the CLI..."))

(grounded? "write-docs" (param "context"))
```

**Behavior:**

1. Reads `(ref step-id)` — the output to verify
2. Calls the LLM with a fixed verification prompt (see 4.3)
3. Expected response: `{"grounded": true, "unsupported": []}` or `{"grounded": false, "unsupported": [{"claim": "...", "reason": "..."}]}`
4. If `grounded` is false: throw `{:kind :grounding-failure :step-id id :unsupported [...]}`
5. Records as `{:kind "grounded" :step-id step-id :gate-passed 1/0}` in steps table
6. Stores unsupported claims in `artifacts` column

### 4.2 Options

```clj
;; Use a different provider for verification (recommended — don't self-verify)
(grounded? "write-docs" (param "context") :provider "lmstudio")

;; Soft mode — record but don't throw
(grounded? "write-docs" (param "context") :strict false)

;; Allow up to N unsupported claims before failing
(grounded? "write-docs" (param "context") :max-unsupported 1)
```

### 4.3 Verification prompt

Hardcoded in the engine. Consistent and testable:

```
You are a factual verification system. Compare the OUTPUT against the CONTEXT (ground truth).

Identify any claims, commands, features, or examples in the OUTPUT that are NOT directly supported by the CONTEXT.

Do not flag style issues, opinions, or reasonable inferences. Only flag factual claims that contradict or have no basis in the context.

CONTEXT:
{context}

OUTPUT:
{output}

Return JSON: {"grounded": true/false, "unsupported": [{"claim": "exact text", "reason": "why unsupported"}]}
```

### 4.4 Default provider

Uses current provider wiring if no `:provider` specified. Documentation recommends using a different model than the one that generated the content for stronger verification.

## 5. Consensus

### 5.1 `(consensus providers opts)`

Runs the same prompt through multiple providers and compares responses.

```clj
(step "critical-decision"
  (consensus ["copilot" "lmstudio"]
    :prompt "Should we deploy this change?"
    :schema {:required ["decision" "reason"]
             :types {"decision" :string "reason" :string}
             :enum {"decision" ["deploy" "hold" "rollback"]}}))
```

**Behavior:**

1. Calls each provider in parallel (reuses `par` infrastructure) with the same prompt
2. Applies `:schema` validation to each response independently (retries per-provider)
3. Compares responses on `:compare-key` (default: first key in `:required`)
4. Returns a result map (serialized as JSON string for step output)
5. Records: `kind = "consensus"`, `gate_passed = 1/0`, `artifacts` = full votes, `confidence` = agreement ratio

### 5.2 Return shape

```clj
;; Agreement
{"agreed" true
 "value" {"decision" "deploy" "reason" "..."}
 "votes" [{"provider" "copilot" "response" {...}}
           {"provider" "lmstudio" "response" {...}}]}

;; Disagreement
{"agreed" false
 "value" nil
 "votes" [{"provider" "copilot" "response" {"decision" "deploy" ...}}
           {"provider" "lmstudio" "response" {"decision" "hold" ...}}]}
```

### 5.3 Agreement logic

- Compares `:compare-key` values across voters
- Exact string match after normalization (lowercase, trim)
- All voters match: `agreed = true`, value = first voter's full response
- Not all match: `agreed = false`, value = nil

```clj
;; Only compare "decision", ignore differing "reason" text
(consensus ["copilot" "lmstudio"]
  :prompt "..."
  :schema {:required ["decision" "reason"]}
  :compare-key "decision")
```

### 5.4 Workflow decides recovery

```clj
(step "verdict"
  (let [result (json/decode (ref "critical-decision"))]
    (if (get result "agreed")
      (get-in result ["value" "decision"])
      ;; Disagreement — escalate to stronger model
      (llm :prompt (str "Two models disagreed: " (json/encode (get result "votes"))
                        "\nMake the final call.")
           :provider "claude"
           :schema {:required ["decision"]}))))
```

### 5.5 Options

| Key | Default | Purpose |
|---|---|---|
| `:prompt` | required | Prompt sent to all voters |
| `:schema` | nil | Schema validation per voter |
| `:compare-key` | first `:required` key | Field that determines agreement |
| `:model` | nil | Override model for all voters |
| `:min-confidence` | nil | Per-voter confidence threshold |

### 5.6 Constraints

- Minimum 2 providers required (throws if < 2)
- Providers must be distinct (no self-voting)
- Provider failure is recorded but doesn't fail consensus — if 2/3 succeed, compare those 2
- If only 1 succeeds: `agreed = false` (can't have consensus with one voice)

## Store Schema Changes

```sql
-- New column on steps table
ALTER TABLE steps ADD COLUMN confidence REAL;
```

The existing `artifacts` TEXT column (currently unused) stores per-step metadata:
- Schema validation results
- Contract pass/fail
- Grounding unsupported claims
- Consensus vote details

All stored as JSON strings.

## SCI Bindings

New bindings added to `runner.clj` `make-sci-ctx`:

```clj
'validate    validate
'grounded?   grounded?
'consensus   consensus
```

The `step` and `llm` functions already exist — they gain new keyword argument handling internally.

## Composition Example

Full pipeline using all 5 mechanisms:

```clj
(workflow "reviewed-deploy"
  :description "Deploy decision with full verification"

  ;; 1. Gather context
  (step "diff" (sh "git diff HEAD~1"))

  ;; 2. Consensus on the decision (schema + confidence per voter)
  (step "decision"
    (consensus ["copilot" "lmstudio"]
      :prompt (str "Analyze this diff and decide: deploy, hold, or rollback.\n"
                   (ref "diff"))
      :schema {:required ["decision" "confidence" "reason"]
               :types {"decision" :string "confidence" :number "reason" :string}
               :enum {"decision" ["deploy" "hold" "rollback"]}}
      :compare-key "decision"
      :min-confidence 0.7))

  ;; 3. If agreed, generate deploy notes
  (step "notes"
    (let [result (json/decode (ref "decision"))]
      (when-not (get result "agreed")
        (throw (ex-info "consensus failed — manual review required" {:votes (get result "votes")})))
      (llm :prompt (str "Write deploy notes for: " (get-in result ["value" "reason"]))
           :schema {:required ["summary" "risks"]}))
    :expects {:non-empty true :min-length 50})

  ;; 4. Ground the notes against the actual diff
  (grounded? "notes" (ref "diff") :provider "lmstudio"))
```

## Testing Strategy

Each primitive is independently testable:

- **Schema enforcement:** Unit tests with valid/invalid JSON, edge cases (empty, nested, missing keys)
- **Step contracts:** Unit tests per predicate, integration test for `:expects` on `step`
- **Confidence:** Unit test threshold logic, integration test retry-on-low-confidence
- **Grounding:** Mock provider returning grounded/ungrounded responses, verify throw/pass behavior
- **Consensus:** Mock 2-3 providers with agreeing/disagreeing responses, verify return shape

Integration tests: run actual `.glitch` workflows that exercise all 5 mechanisms end-to-end using the `lmstudio` provider.
