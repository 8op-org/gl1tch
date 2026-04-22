# DSL Reference

Complete reference for the glitch workflow DSL — all primitives, LLM options, control flow, validation, and utilities.

## Core Primitives

| Form | Description |
|------|-------------|
| `(workflow "name" :description "..." body...)` | Workflow definition |
| `(step "id" body)` | Evaluate body, record output under id |
| `(ref "id")` | Retrieve output of a previous step |
| `(input)` | User input passed to the workflow |
| `(params)` | All runtime parameters as a map |
| `(param "key")` | Single parameter from `--set key=value` |
| `(sh "command")` | Shell command (bash -c), returns stdout |
| `(run "command")` | Alias for `sh` |
| `(save "path" content)` | Write content to file |
| `(read-file "path")` | Read file contents |
| `(write-file "path" content)` | Alias for save |
| `(last-output)` | Output of the most recent step |
| `(get-steps)` | Map of all step outputs |
| `(search "query" :limit 10 :path ".")` | Ripgrep wrapper |

## Triple-Backtick Strings

The preprocessor converts ` ``` ` blocks into `(str ...)` forms. Use `~(ref "id")` and `~(param "key")` for interpolation:

```clojure
(step "summarize"
  (llm :prompt ```
    Summarize this data for a developer:
    ~(ref "fetch")

    User asked: ~(input)
    Repo: ~(param "repo")
    ```))
```

## LLM Invocation

```clojure
(llm
  :prompt "..."                    ;; required
  :provider "lmstudio"             ;; optional, uses default if omitted
  :model "qwen3:8b"                ;; optional
  :schema {:required ["key"]       ;; JSON schema validation
           :types {"key" :string}
           :enum {"status" ["ok" "fail"]}}
  :retries 2                       ;; retry on schema/confidence failure
  :min-confidence 0.8              ;; minimum confidence threshold
  :skill "path/to/skill.md"       ;; prepend skill content to prompt
  :tools [...]                     ;; tool definitions for tool-calling providers
  :agentic true                    ;; enable multi-round tool calling
  :max-rounds 5                    ;; cap tool-calling rounds
  :domain-check true               ;; apply domain-relevance adjustment
  :step-id "custom-id"             ;; custom step recording ID
  :format "json")                  ;; output format hint
```

## Control Flow

```clojure
;; Concurrent execution
(par
  (step "a" (sh "curl http://api1"))
  (step "b" (sh "curl http://api2")))

;; Retry on failure
(retry 3
  (step "flaky" (sh "curl -sf https://api.example.com/data")))

;; Timeout (seconds)
(with-timeout 120
  (step "slow" (llm :prompt "Analyze everything...")))

;; Gate — boolean check (non-fatal, records pass/fail)
(gate "has-tests" (not (str/blank? (ref "find-tests"))))

;; Phase — progress grouping
(phase "research"
  (step "fetch" (sh "gh issue view 42 --json body"))
  (step "analyze" (llm :prompt (str "..." (ref "fetch")))))

;; Call sub-workflow (resolves from same directory)
(call-workflow "site-write"
  :input "update the getting-started page"
  :set {"page" "getting-started" "instructions" "add babashka syntax"})

;; Conditional (standard Clojure)
(cond
  (= action "write") (call-workflow "write" :set {"page" slug})
  (= action "dev")   (call-workflow "dev")
  :else              (str "unknown: " action))
```

## Validation & Confidence

```clojure
;; Contract validation on step output
(step "classify" (llm :prompt "...")
  :expects {:non-empty true :json true :keys ["type" "severity"]})

;; Schema validation against a step
(validate "classify"
  {:required ["type" "severity"]
   :types {"type" :string "severity" :string}
   :enum {"severity" ["low" "medium" "high" "critical"]}})

;; Grounding check — verify output against source context
(grounded? "summary" (ref "raw-data")
  :provider "claude" :strict true :max-unsupported 0)

;; Multi-provider consensus
(consensus ["claude" "openrouter" "lmstudio"]
  :prompt "Classify this issue: ..."
  :schema {:required ["severity"]}
  :compare-key "severity")

;; Composite quality score (harmonic mean of gate results)
(composite-score "classify")
```

## Available Utilities

Full Clojure standard library via SCI, plus:

| Symbol | Description |
|--------|-------------|
| `str/*` | All of `clojure.string` (e.g., `str/trim`, `str/split`) |
| `json/decode` | Parse JSON string to map |
| `json/encode` | Map to JSON string |
| `json-extract` | Extract first JSON object/array from LLM noise |
| `mkdir-p` | Create directories |
| `slurp` / `spit` | Read/write files |
| `atom` / `swap!` / `reset!` | Concurrency primitives |
| `future-call` / `deref` | Async execution |
| `def` / `let` / `cond` / `when` / `map` / `filter` / `reduce` | Standard Clojure |

## File Locations

- **Project-local**: `.glitch/workflows/` — loaded for the project
- **Global plugins**: `~/.config/glitch/plugins/` — available everywhere
- **Workflow dir**: `call-workflow` resolves siblings in the parent directory of the running file
