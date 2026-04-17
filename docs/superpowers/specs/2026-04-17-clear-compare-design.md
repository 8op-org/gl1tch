# CLEAR Compare Design

**Date:** 2026-04-17
**Status:** Draft
**Scope:** Runner, sexpr parser, ES telemetry, batch integration

## Problem

Compare/batch workflows have two gaps:

1. **No contract** — workflows jump into branches without declaring what they're measuring or why. The judge has no stated objective to evaluate against.
2. **No reflection** — a winner is picked but nothing is learned. There's no structured output capturing what the comparison proved, what models are good/bad at, or what future runs should do differently.

## Solution: CLEAR as Runner Phases

Map the CLEAR framework (Contract, Listen, Explore, Action, Review) onto the compare/batch execution model. Most phases already exist — this design adds the missing bookends (Contract and Review) and tags existing phases for telemetry.

## Design

### 1. Contract Enforcement

`:objective` becomes a required field on `(compare ...)` blocks.

```scheme
(compare
  :objective "determine which model produces valid, actionable migration plans"
  :criteria ["json_validity" "plan_completeness" "actionability"]
  (branch "local" (step "plan" (llm :model "qwen2.5:7b" :prompt "...")))
  (branch "cloud" (step "plan" (llm :provider "claude" :prompt "...")))
  (review ...))
```

**Runner behavior:**

- Parser extracts `:objective` during sexpr parsing
- Runner validates `:objective` is non-empty before executing the compare block — hard fail if missing
- `:objective` is injected into the judge/review prompt automatically, so the judge evaluates branches against the stated objective
- `:objective` is recorded in the run's ES document (`glitch-runs` and `glitch-cross-reviews`)

**CLI implicit compare:** When `--variant` flags inject implicit compare blocks, the user must also pass `--objective "..."` or the runner refuses.

### 2. Phase Telemetry

The 5 CLEAR phases get recorded as metadata on existing telemetry. No new indices — richer annotations on what's already there.

**Phase tagging on step records:**

| Step Type | Phase Tag |
|---|---|
| Shared/seed `(run ...)` steps before branches | `listen` |
| Branch steps during parallel execution | `explore` |
| Judge/review step | `action` |
| Reflection step (new) | `review` |

The `contract` phase isn't a step — it's metadata on the compare block itself (`:objective` + `:criteria` stored on the parent run document).

**ES mapping additions on `glitch-llm-calls` and `glitch-cross-reviews`:**

```json
"phase": { "type": "keyword" },
"objective": { "type": "text" }
```

### 3. Reflection Step

After the judge picks a winner, the runner automatically executes a reflection step. This is the "Review" in CLEAR.

**What it does:** The runner calls the LLM with a structured prompt containing the original objective, criteria and scores from the judge, each branch's final output, and the winner with margin of victory.

**Prompt output structure:**

1. **FINDING** — one sentence on what the comparison proved
2. **MODEL_INSIGHT** — per-model strengths/weaknesses on these criteria
3. **CONFIDENCE** — high/medium/low based on decisiveness of the result
4. **RECOMMENDATION** — one sentence on what future runs should do differently

**Model selection:** Uses the same model as the review/judge step.

**ES index:** `glitch-learnings`

```json
{
  "run_id": "...",
  "compare_id": "...",
  "objective": "...",
  "scope": "compare",
  "finding": "...",
  "model_insight": { "local": "...", "cloud": "..." },
  "confidence": "high",
  "recommendation": "...",
  "criteria": ["json_validity", "plan_completeness"],
  "winner": "cloud",
  "margin": 7,
  "models_tested": ["qwen2.5:7b", "claude-sonnet-4-20250514"],
  "timestamp": "..."
}
```

**Cost:** One extra LLM call per compare block. Cheap — the prompt is small and the output is structured.

### 4. Batch Integration

Reflection happens at two levels in batch mode:

**Per-compare reflection:** Each compare block within a variant workflow gets its own reflection step. Scope: `"compare"`.

**Batch-level reflection:** After all items complete and the cross-review runs, the batch runner executes one final reflection aggregating patterns across all per-compare learnings. Scope: `"batch"`, linked via `batch_run_id`.

**Manifest update:** The existing batch manifest gets a "Learnings" section at the bottom, pulled from the batch-level reflection.

### 5. CLI Surface

**No new commands.** Reflection is automatic.

**Changes to existing commands:**

- `glitch workflow run`: `--objective "..."` required when `--variant` flags are used (implicit compare). Reflection output printed to terminal after winner announcement.
- `glitch observe`: Gains ability to query `glitch-learnings` index via existing query surface.

**Terminal output example:**

```
-- Compare: implementation ----------------------
Objective: determine which model produces valid migration plans

  local   (qwen2.5:7b)   18/30
  cloud   (claude)        27/30

Winner: cloud

-- Learning -------------------------------------
Finding: Cloud model produced valid JSON in all cases; local model
failed JSON validation in the plan step.
Confidence: high
Recommendation: Use local model for drafting prose plans, escalate
to cloud for structured output requirements.
```

### 6. Deferred: Read-Side Feedback Loop

Learnings are queryable via `glitch observe` but do not auto-influence future runs. The read side (querying learnings at contract time to inform the judge, or at routing time to influence tier selection) is deferred until the shape of the learnings data is trusted.

## Phase Mapping Summary

| CLEAR Phase | gl1tch Mechanism | Status |
|---|---|---|
| Contract | `:objective` + `:criteria` on compare blocks | **New** |
| Listen | Shared/seed steps before branches | Exists, add phase tag |
| Explore | Parallel branch execution | Exists, add phase tag |
| Action | Judge/review picks winner | Exists, add phase tag |
| Review | Reflection step + ES learning doc | **New** |
