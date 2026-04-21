# DSL Reference

Extended forms beyond the core workflow syntax. These cover data transforms, confidence checking, and advanced control flow.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it, start there.

## Confidence framework

Built-in functions for validating LLM output quality, schema conformance, and factual grounding.

### Schema validation with `validate`

Check that a step's JSON output matches a schema:

```glitch
(step "classify"
  (llm :format "json"
    :prompt "Classify this issue. Return JSON with type, severity, summary."))

(validate "classify"
  {:required ["type" "severity" "summary"]
   :types    {"type"     :string
              "severity" :string
              "summary"  :string}
   :enum     {"type"     ["bug" "feature" "refactor" "documentation"]
              "severity" ["low" "medium" "high" "critical"]}})
```

Returns true on success, throws `:schema-violation` on failure. Schema options:

| Key | Description |
|-----|-------------|
| `:required` | Keys that must be present |
| `:types` | Expected types per key (`:string`, `:number`, `:bool`, `:array`, `:object`) |
| `:enum` | Allowed values per key |

### Inline schema on `llm`

Pass `:schema` directly to `(llm ...)` for automatic validation with retries:

```glitch
(step "classify"
  (llm :format "json"
    :schema {:required ["type" "confidence"]
             :types {"type" :string "confidence" :number}}
    :retries 2
    :min-confidence 0.7
    :prompt "Classify this issue..."))
```

If the response doesn't match the schema, the LLM is re-prompted with the violation details. If `:min-confidence` is set and the response includes a `"confidence"` field below the threshold, the LLM is pushed to try harder.

### Grounding with `grounded?`

Verify that LLM output is factually supported by context:

```glitch
(step "docs"
  (run "cat README.md"))

(step "summary"
  (llm :prompt "Summarize this project:\n~(step docs)"))

(grounded? "summary" (ref "docs")
  :strict true
  :max-unsupported 0)
```

A separate LLM call compares the output against the context and flags unsupported claims. Options:

| Key | Default | Description |
|-----|---------|-------------|
| `:provider` | default | Which provider runs the grounding check |
| `:strict` | `true` | Throw on grounding failure vs. return false |
| `:max-unsupported` | `0` | Tolerated number of unsupported claims |

### Consensus with `consensus`

Run the same prompt through multiple providers and compare:

```glitch
(step "result"
  (consensus ["copilot" "claude" "lmstudio"]
    :prompt "What language is this project written in? Return JSON: {\"language\": \"...\"}"
    :schema {:required ["language"] :types {"language" :string}}
    :compare-key "language"))
```

Returns JSON: `{"agreed": true/false, "value": {...}, "votes": [...]}`. Requires at least 2 providers. Agreement is checked on the `:compare-key` field (case-insensitive).

## Contract checking with `check-contract`

Validate step output against structural expectations without parsing JSON:

```glitch
(step "report"
  (llm :prompt "Write a detailed analysis...")
  :expects {:non-empty true
            :min-length 200
            :matches #"## Summary"})
```

The `:expects` keyword on `(step ...)` triggers contract checking. Contract options:

| Key | Description |
|-----|-------------|
| `:non-empty` | Output must not be blank |
| `:min-length` | Minimum character count |
| `:max-length` | Maximum character count |
| `:json` | Output must parse as valid JSON |
| `:keys` | JSON output must contain these keys |
| `:matches` | Output must match this regex pattern |
| `:pred` | Custom predicate function |

Contract violations throw `:contract-violation` with the step ID, expected contract, and actual value.

## JSON extraction

`json-extract` pulls the first JSON object or array from noisy LLM output:

```glitch
(step "parsed"
  (json-extract (ref "raw-llm-output")))
```

Handles markdown fences, thinking indicators, and other LLM noise. Tracks nesting depth and respects string escaping.

## Threading with `->`

Pipe data through transforms in `(def ...)` context:

```glitch
(def examples
  (-> (glob "examples/*.glitch")
      (map read-file)
      (join "\n\n")))

(def commands
  (-> (read-file "valid-commands.txt")
      (lines)
      (filter (contains "glitch"))
      (join "\n")))
```

Available in `(def)` context only (parse time). Functions: `read-file`, `glob`, `map`, `filter`, `lines`, `join`, `split`, `trim`, `upper`, `lower`, `replace`, `contains`, `flatten`.

## Parallel execution with `par`

Run steps concurrently:

```glitch
(par
  (step "prs"    (run "gh pr list --json number,title"))
  (step "issues" (run "gh issue list --json number,title"))
  (step "commits" (run "git log --oneline -10")))
```

All three steps run in parallel. `(par ...)` returns a vector of results. Use this when steps are independent and you want to cut wall-clock time.

## call-workflow

Execute a child workflow inline:

```glitch
(step "review"
  (call-workflow "pr-review"
    :set (repo "acme/backend")
    :set (pr "42")))
```

The child inherits the active workspace and resources. Cycle detection prevents recursive loops. The child's final output becomes this step's output.

## Tool control on LLM steps

By default, all available MCP tools are injected into every `(llm ...)` call. Control which tools are available:

```glitch
;; All tools (default)
(step "with-tools"
  (llm :prompt "Search the codebase for auth patterns..."))

;; Specific tools only
(step "search-only"
  (llm :tools ["glitch_search"]
    :prompt "Find all error handlers..."))

;; No tools — plain LLM call
(step "plain"
  (llm :tools []
    :prompt "Summarize this text..."))
```

For agentic multi-round tool use:

```glitch
(step "agent"
  (llm :agentic true :max-rounds 5
    :prompt "Find and fix the bug in auth.clj..."))
```

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — core forms, templates, and control flow
- [Phases & Gates](/docs/phases-and-gates) — verification checkpoints
- [Compare Runs](/docs/compare) — evaluating across providers
