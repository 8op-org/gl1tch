# Compare as Core DSL and CLI Primitive

**Date:** 2026-04-15
**Status:** Draft

## Problem

Comparison is currently bolted onto `glitch ask` via `--compare` and naming conventions (`issue-to-pr-local`, `issue-to-pr-claude`). It's invisible to the DSL, unavailable from `glitch workflow run`, and requires duplicate workflow files for what's often "same workflow, different model." Comparison should be a first-class concept at every layer: DSL form for authoring, CLI flag for ad-hoc use, unified telemetry feeding Kibana dashboards in the GUI.

## Design

### 1. DSL Grammar: `(compare ...)`, `(branch ...)`, `(review ...)`

`(compare ...)` is a flow-control form, peer to `(par ...)`. It contains named `(branch ...)` blocks and an optional `(review ...)` judge.

#### Step-level compare (single alternatives)

```scheme
(step "analyze"
  (compare
    (branch "fast"    (llm :model "qwen2.5:7b" :prompt "..."))
    (branch "quality" (llm :provider "claude" :prompt "..."))
    (review :criteria ["accuracy" "specificity"])))
```

#### Multi-step compare (full sequences)

```scheme
(compare
  :id "implementation"
  (branch "local"
    (step "plan" (llm :model "qwen2.5:7b" :prompt "..."))
    (step "code" (run "apply-patch {{step \"plan\"}}")))
  (branch "cloud"
    (step "plan" (llm :provider "claude" :prompt "..."))
    (step "code" (run "apply-patch {{step \"plan\"}}")))
  (review
    :criteria ["plan_completeness" "code_quality" "accuracy"]
    :model "qwen2.5:7b"))
```

#### Custom review prompt (override default judge)

```scheme
(compare
  :id "tone-check"
  (branch "formal" (llm :prompt "Write formally: ..."))
  (branch "casual" (llm :prompt "Write casually: ..."))
  (review
    :prompt "Which response better matches our brand voice? ..."
    :model "qwen2.5:7b"))
```

#### Grammar rules

- `(compare ...)` can appear inside a `(step ...)` body or at the workflow top level
- When inside a step, the winning branch's output becomes that step's output
- When at the top level, `:id` is required and used for downstream `{{step "id"}}` references (returns winner's final step output)
- `(branch "name" ...)` contains one or more step bodies (llm, run, save) or full `(step ...)` forms
- `(review ...)` is optional; omitted = default judge with generic quality criteria
- At least two branches are required

### 2. Branch Execution and Scoping

#### Parallel execution

- All branches launch simultaneously, reusing the `par` execution machinery
- Each branch gets its own copy of current params and accumulated step results
- Branches can read outer step results (`{{step "setup"}}` from before the compare)
- Branches cannot see sibling branch state

#### Step namespacing

- Steps inside branches are namespaced: `<compare-id>/<branch-name>/<step-id>`
  - Example: `implementation/local/plan`, `implementation/cloud/plan`
- Within a branch, `{{step "plan"}}` resolves locally (branch scope first, then outer workflow scope)
- From outside the compare:
  - `{{step "implementation"}}` — winner's final step output (default)
  - `{{step "implementation" :variant "local"}}` — specific branch's output
  - `{{step "implementation" :scores}}` — structured score data
  - `{{step "implementation" :winner}}` — winning branch name as a string

#### Seed step optimization

- If branches contain identical `(run ...)` steps (same command text after template expansion), the runner detects this and executes once, seeding all branches with the result
- Preserves the "shell steps own data fetching" pattern efficiently

#### Failure handling

- If a branch fails, it's excluded from review
- Sibling branches continue to completion
- If all branches fail, the compare block fails
- If only one branch succeeds, it wins by default (no review needed)

### 3. Review System

#### Default judge (implicit)

When `(review ...)` is omitted, glitch uses a built-in judge prompt. It receives all branch outputs, scores each on generic criteria (coherence, completeness, relevance), and picks a winner. Model defaults to the workflow's `(def model ...)` or the global default (qwen2.5:7b).

#### Criteria mode

```scheme
(review :criteria ["accuracy" "specificity" "actionability"])
```

Generates a structured prompt asking the judge to score each branch 1-10 on each criterion. Uses the numeric scoring format from `parseCrossReviewNumeric` — scores >= 7 count as passed. Output format:

```
VARIANT: local
accuracy: 8/10
specificity: 6/10
actionability: 9/10
total: 23/30

VARIANT: cloud
accuracy: 9/10
specificity: 9/10
actionability: 8/10
total: 26/30

WINNER: cloud
```

#### Custom prompt mode

```scheme
(review
  :prompt "Which response better matches our brand voice? ..."
  :model "qwen2.5:7b")
```

Full control over the judge prompt. Template access to branch outputs via `{{branch "local"}}` and `{{branch "cloud"}}`. Must emit the structured `VARIANT: / total: / WINNER:` format for the scoring pipeline to parse.

#### Workflow-level rollup

After all compare blocks finish, a regular step can inspect the full picture:

```scheme
(step "rollup"
  (llm :prompt ```
    Compare block "implementation" chose {{step "implementation" :winner}}.
    Compare block "tone-check" chose {{step "tone-check" :winner}}.
    Does this combination make sense?
    ```))
```

No special form needed — the template accessors give rollup steps everything.

### 4. CLI Integration

#### `glitch workflow run` gains comparison

```bash
# Ad-hoc: same workflow, different models
glitch workflow run analyze \
  --variant ollama:qwen2.5:7b \
  --variant claude \
  --variant copilot

# With custom review criteria
glitch workflow run analyze \
  --variant ollama:qwen2.5:7b \
  --variant claude \
  --review-criteria "accuracy,completeness"

# Explicit results dir
glitch workflow run analyze \
  --variant ollama:qwen2.5:7b \
  --variant claude \
  --results-dir ./my-results
```

When `--variant` flags are present, the runner wraps every `(llm ...)` step that is NOT already inside a `(compare ...)` block in an implicit compare — one branch per variant. Explicit compare blocks in the workflow are left untouched. The workflow author writes no compare logic; comparison comes from the CLI.

#### `glitch workflow run --compare`

```bash
# Discovers sibling workflows by naming convention
glitch workflow run issue-to-pr --compare

# Combines naming convention + extra variant
glitch workflow run issue-to-pr --compare --variant gemini
```

Naming convention (`-local`, `-claude`) still works for structurally different workflows. These become branches in a top-level compare.

#### `glitch ask` unchanged

`glitch ask` with `--compare` continues to work. Internally, the batch system constructs compare blocks instead of managing variant workflows directly — same external behavior, new internal implementation.

### 5. Telemetry: Run Index and Compare Scores

#### Run documents (new)

Every `pipeline.Run` call emits a run document to ES, regardless of comparison:

| Field | Type | Description |
|-------|------|-------------|
| `run_id` | keyword | Unique run identifier |
| `workflow_name` | keyword | Workflow that was executed |
| `workspace` | keyword | Workspace context |
| `source` | keyword | `"cli"`, `"gui"`, or `"batch"` |
| `status` | keyword | `"running"`, `"completed"`, `"failed"` |
| `params` | object | Params map passed to the run |
| `has_compare` | boolean | Whether run involved any compare blocks |
| `duration_ms` | long | Total runtime |
| `timestamp` | date | When the run started |

Index name: `glitch-runs`

#### Cross-review documents (extended)

The existing `glitch-cross-reviews` index gains new fields:

| Field | Type | Description |
|-------|------|-------------|
| `compare_id` | keyword | Links to a specific `(compare ...)` block |
| `scope` | keyword | `"step"` for DSL-level, `"workflow"` for CLI/batch |
| `criteria_name` | keyword | Individual criterion name |
| `criteria_score` | integer | Raw score (1-10) |
| `workflow_name` | keyword | Source workflow |
| `workspace` | keyword | Workspace context |

Existing fields unchanged: `run_id`, `variant`, `passed`, `total`, `confidence`, `winner`. Current dashboards continue to work.

### 6. GUI Integration

#### Runs section

All runs appear in the GUI's runs section regardless of trigger source. The `source` field shows as a badge (`CLI`, `GUI`, `BATCH`) so users can distinguish origin at a glance.

#### Comparison tab

Any workflow run involving `(compare ...)` blocks shows a comparison tab with:

- Side-by-side branch outputs
- Per-branch scores and criteria breakdown
- Winner highlight
- Manifest summary (for multi-compare runs)

#### Embedded Kibana dashboards

Three dashboard views embeddable in the workflow GUI:

1. **Model leaderboard** — which model/provider wins most often, aggregated across compare blocks. Bar chart, filterable by workspace and time range.
2. **Criteria breakdown** — per-criteria scores across variants for a specific compare block. Radar or grouped bar chart showing variant strengths and weaknesses.
3. **Run history** — timeline of compare results for a workflow. Shows winner drift over time — useful for detecting model degradation or prompt improvements.

Dashboards are filtered to the current workflow/workspace context when embedded.

## Migration

- Existing `--compare` on `glitch ask` continues to work unchanged
- Existing naming convention workflows (`-local`, `-claude`) continue to work
- The batch system (`internal/batch`) is refactored to construct compare blocks internally
- No breaking changes to workflow files or CLI flags

## Non-Goals

- No migration tooling for converting existing variant workflow sets into compare blocks (manual for now)
- No distributed/remote branch execution — all branches run on the local machine
- No automatic model discovery — variants are explicitly named
