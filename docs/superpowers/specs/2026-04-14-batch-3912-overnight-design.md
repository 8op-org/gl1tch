# Overnight Batch: Issue #3912 Skill Creator Guide

**Date**: 2026-04-14
**Issue**: elastic/observability-robots#3912
**Goal**: Generate all 7 sub-issue deliverables using glitch workflows with local LLMs, run Claude comparison for each, capture telemetry in Elasticsearch, and visualize results in Kibana.

## Sub-Issues

| Issue | Deliverable | Pattern | Exemplar Skill | Depends On |
|-------|------------|---------|----------------|------------|
| #3916 | `index.md` — guide structure + intro | — | — | — |
| #3918 | `01-wrapper-curl.md` | Tool Wrapper | `github-cli` | #3916 |
| #3919 | `02-generator-mermaid.md` | Generator | `github-issues` | #3916 |
| #3920 | `03-reviewer-kiss.md` | Reviewer | `docs-humanizer` | #3916 |
| #3921 | `04-inversion-invitations.md` | Inversion | `lab-documentation` | #3916 |
| #3922 | `05-pipeline-pr-checks.md` | Pipeline | `epic-story-splitter` | #3916 |
| #3923 | `index-final.md` — humanization + finalize | — | — | #3918-#3922 |

## Architecture

### Workflow Shapes

**Shape 1: Intro** (#3916) — research repo, analyze all 5 patterns, write intro, self-review.

**Shape 2: Use-case** (#3918-#3922) — research repo, read exemplar skill for this pattern, analyze specific pattern, write narrative use-case doc, self-review against acceptance criteria.

**Shape 3: Humanization** (#3923) — read all 5 prior outputs + index, humanize all docs, rewrite final index with links, review for AI writing patterns.

### Sexpr Workflow Format

All workflows written as `.glitch` files using s-expression syntax. Stored in `~/Projects/stokagent/workflows/batch-3912/`.

Each issue gets two files:
- `{issue}-local.glitch` — all LLM steps use Ollama with local models
- `{issue}-claude.glitch` — all LLM steps use `provider: "claude"` with model specified

Example structure (use-case shape):

```lisp
(def issue "3918")
(def pattern "wrapper")
(def use-case "01-wrapper-curl")
(def exemplar "github-cli")
(def provider "ollama")
(def coder "qwen3-coder:30b")
(def writer "qwen3.5:35b-a3b")
(def reviewer "qwen3:8b")

(workflow "3918-wrapper-curl-local"
  :description "Issue #3918: Tool Wrapper use-case guide (local models)"

  (step "skill-inventory" (run "..."))
  (step "read-exemplar" (run "..."))
  (step "read-index" (run "..."))
  (step "analyze"
    (llm :provider provider :model coder
      :prompt ```...analyze the wrapper pattern...```))
  (step "write"
    (llm :provider provider :model writer
      :prompt ```...write narrative doc...```))
  (step "review"
    (llm :provider provider :model reviewer
      :prompt ```...review against criteria...```))
  (step "save-doc"
    (save "results/3918/01-wrapper-curl.md" :from "write"))
  (step "save-review"
    (save "results/3918/review.md" :from "review")))
```

Claude variant swaps `def` bindings:

```lisp
(def provider "claude")
(def coder "haiku")
(def writer "sonnet")
(def reviewer "haiku")
```

### Model Selection

| Role | Local Model | Claude Model | Rationale |
|------|-------------|-------------|-----------|
| Code analysis | qwen3-coder:30b | haiku | Pattern analysis from skill code |
| Writing | qwen3.5:35b-a3b | sonnet | Narrative documentation |
| Review | qwen3:8b | haiku | Structured PASS/FAIL evaluation |

## Code Changes

### 1. Provider model passthrough

**File**: `internal/provider/provider.go` — `RunProvider()`

Currently the provider registry renders `{{.prompt}}` into the command template. Add `{{.model}}` support:

```go
// Template data includes both prompt and model
data := map[string]string{"prompt": prompt, "model": model}
```

**File**: `internal/provider/provider.go` — `ProviderRegistry.RunProvider()` signature

Add `model` parameter:

```go
func (r *ProviderRegistry) RunProvider(name, model, prompt string) (string, error)
```

**File**: `~/.config/glitch/providers/claude.yaml`

```yaml
name: claude
command: claude -p --model {{.model}} --output-format text
```

When `model` is empty, the runner skips `--model` entirely by using the stdin path (no `{{.prompt}}` or `{{.model}}` in template). Concrete logic: if the provider template contains `{{.model}}` and model is non-empty, render normally. If model is empty and template contains `{{.model}}`, use the provider's base command (strip `--model {{.model}}`) and pipe via stdin. Simplest implementation: two provider files aren't needed — just check `model != ""` before rendering.

### 2. Telemetry enhancements

**File**: `internal/esearch/telemetry.go` — `LLMCallDoc`

Add fields:

```go
WorkflowName    string `json:"workflow_name"`
Issue           string `json:"issue"`
ComparisonGroup string `json:"comparison_group"`
TokensTotal     int64  `json:"tokens_total"`
```

**File**: `internal/esearch/mappings.go` — `LLMCallsMapping`

Add to mapping:

```json
"workflow_name":      { "type": "keyword" },
"issue":              { "type": "keyword" },
"comparison_group":   { "type": "keyword" },
"tokens_total":       { "type": "long" }
```

**New index**: `glitch-workflow-runs` — one doc per complete workflow execution:

```go
type WorkflowRunDoc struct {
    RunID           string  `json:"run_id"`
    WorkflowName    string  `json:"workflow_name"`
    Issue           string  `json:"issue"`
    ComparisonGroup string  `json:"comparison_group"`
    TotalSteps      int     `json:"total_steps"`
    LLMSteps        int     `json:"llm_steps"`
    TotalTokensIn   int64   `json:"total_tokens_in"`
    TotalTokensOut  int64   `json:"total_tokens_out"`
    TotalCostUSD    float64 `json:"total_cost_usd"`
    TotalLatencyMS  int64   `json:"total_latency_ms"`
    ReviewPass      bool    `json:"review_pass"`
    Timestamp       string  `json:"timestamp"`
}
```

### 3. Pipeline runner telemetry context

**File**: `internal/pipeline/runner.go` — `RunOpts`

Extend with context fields:

```go
type RunOpts struct {
    Telemetry       *esearch.Telemetry
    Issue           string
    ComparisonGroup string // "local" or "claude"
}
```

The runner populates `WorkflowName` from `w.Name`. `Issue` and `ComparisonGroup` are derived from the workflow name by convention:
- Names ending in `-local` → `comparison_group: "local"`, strip suffix for base name
- Names ending in `-claude` → `comparison_group: "claude"`, strip suffix for base name
- Leading digits in the base name → `issue` (e.g., `"3918-wrapper-curl-local"` → issue `"3918"`, group `"local"`)

This avoids extra CLI flags. The convention is enforced by the batch workflow naming.

After all steps complete, the runner indexes a `WorkflowRunDoc` summarizing the run.

The `ReviewPass` field is derived by scanning the last LLM step's output for "OVERALL: PASS" or "OVERALL: FAIL" (case-insensitive).

### 4. Delete and recreate ES index

The `glitch-llm-calls` mapping needs new fields. Since we're pre-1.0 — wipe and recreate:

```bash
curl -X DELETE http://localhost:9200/glitch-llm-calls
curl -X DELETE http://localhost:9200/glitch-workflow-runs
```

The runner's `EnsureIndices()` recreates them with the new mapping on next run.

## File Layout

```
~/Projects/stokagent/workflows/batch-3912/
  3916-local.glitch
  3916-claude.glitch
  3918-local.glitch
  3918-claude.glitch
  3919-local.glitch
  3919-claude.glitch
  3920-local.glitch
  3920-claude.glitch
  3921-local.glitch
  3921-claude.glitch
  3922-local.glitch
  3922-claude.glitch
  3923-local.glitch
  3923-claude.glitch

~/Projects/stokagent/scripts/
  batch-3912.sh          # overnight runner

~/Projects/gl1tch/results/
  3916/index.md
  3916/review.md
  3916-claude/index.md
  3916-claude/review.md
  3918/01-wrapper-curl.md
  3918/review.md
  3918-claude/01-wrapper-curl.md
  3918-claude/review.md
  ... (same pattern for 3919-3923)
```

## Batch Runner

**File**: `~/Projects/stokagent/scripts/batch-3912.sh`

```bash
#!/bin/bash
set -euo pipefail

cd ~/Projects/gl1tch
go build -o /tmp/glitch-batch .
GLITCH="/tmp/glitch-batch"

# Wipe old telemetry (pre-1.0, no migrations)
curl -sf -X DELETE http://localhost:9200/glitch-llm-calls 2>/dev/null || true
curl -sf -X DELETE http://localhost:9200/glitch-workflow-runs 2>/dev/null || true
curl -sf -X DELETE http://localhost:9200/glitch-tool-calls 2>/dev/null || true
curl -sf -X DELETE http://localhost:9200/glitch-research-runs 2>/dev/null || true

# Phase 1: Intro
for variant in local claude; do
  echo ">>> #3916 ($variant) — $(date)"
  $GLITCH workflow run "3916-$variant" 2>&1 | tee "results/3916-$variant.log"
done

# Phase 2: Use cases (depend on #3916)
for issue in 3918 3919 3920 3921 3922; do
  for variant in local claude; do
    echo ">>> #$issue ($variant) — $(date)"
    $GLITCH workflow run "${issue}-$variant" 2>&1 | tee "results/${issue}-$variant.log"
  done
done

# Phase 3: Humanization (depends on all use cases)
for variant in local claude; do
  echo ">>> #3923 ($variant) — $(date)"
  $GLITCH workflow run "3923-$variant" 2>&1 | tee "results/3923-$variant.log"
done

echo ">>> Batch complete — $(date)"
echo ">>> Dashboard: http://localhost:5601/app/dashboards#/view/glitch-llm-dashboard"
```

Estimated runtime: ~45 minutes total.

## Kibana Dashboard

Rebuild using Vega visualizations (API-safe, won't have the Lens rendering issue):

1. **Token Usage by Model** — stacked bar chart, `model` as color, `tokens_total` as metric
2. **Latency: Local vs Claude** — grouped horizontal bars, `comparison_group` as color, `workflow_name` as category
3. **Cost: Local vs Claude** — donut chart, `comparison_group` slices, `cost_usd` metric
4. **Confidence: Review Pass Rate** — bar chart from `glitch-workflow-runs`, `review_pass` count by `comparison_group` per `issue`
5. **Per-Issue Comparison Table** — data table showing issue, local tokens/latency/pass, claude tokens/latency/pass side by side

Panel 4 (confidence comparison) is the key demo metric — it proves local models produce correct results.

## Implementation Order

1. Code changes: provider model passthrough, telemetry schema, pipeline runner context
2. Provider YAML: update `claude.yaml` with `{{.model}}`
3. ES: wipe and recreate indices
4. Workflows: write all 14 `.glitch` files to stokagent
5. Batch runner script
6. Test: run #3916 pair, verify telemetry lands correctly
7. Kibana: rebuild dashboard with Vega panels
8. Run overnight batch
