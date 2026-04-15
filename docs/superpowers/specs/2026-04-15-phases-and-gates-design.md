# Phases & Gates Design

**Date:** 2026-04-15
**Status:** Draft
**Goal:** Add `(phase ...)` and `(gate ...)` forms to the workflow sexpr syntax so workflows can declare verified, retriable units of work — generalizing the triplegate pattern from the site-update workflow into glitch core.

## Problem

Today the pipeline runner executes steps sequentially with no concept of verification checkpoints. The `--compare` flag bolts on cross-review from the outside, but individual workflows can't declare "verify this work before continuing." The site-update workflow design introduced a triplegate pattern (parse check, build check, diff-review) that proved effective, but it's hardcoded to that workflow's needs. We need it as a generic primitive.

## Design

### New Forms

Two new sexpr forms: `phase` and `gate`.

**`(phase "name" [:retries N] ...children)`**

Groups steps and gates into a retriable unit. Children are either `(step ...)` or `(gate ...)`. The runner executes all steps in declaration order, then all gates in declaration order. If any gate fails, the entire phase re-runs (up to `:retries` times, default 0).

**`(gate "name" (run ...) | (llm ...))`**

A verification step that lives inside a phase. Structurally identical to a step — it has an ID, runs a shell command or LLM call, and writes output to the step store. The difference is semantic: its output is evaluated as a pass/fail verdict, and failure triggers phase retry or halt.

### Bare Steps

Steps outside any phase work exactly as today. No gates, no retry, no behavior change. A workflow with zero phases is identical to the current runner. Zero breaking changes.

### Syntax

```scheme
(workflow "site-update"
  :description "regenerate site from stubs"

  (phase "gather"
    (step "stubs" (run "cat docs/site/*.md"))
    (step "examples" (run "cat examples/*.glitch")))

  (phase "generate" :retries 2
    (step "enrich-docs"
      (llm :prompt "Expand these stubs into full docs.
            Stubs: {{step \"stubs\"}}
            Examples: {{step \"examples\"}}
            Previous gate feedback: {{step \"syntax-check\"}}"))

    (gate "syntax-check"
      (run "validate-sexpr.sh"))
    (gate "build-check"
      (run "cd site && npx astro build"))
    (gate "diff-review"
      (llm :tier 2
        :prompt "Compare stubs vs generated docs.
                 Check: no hallucinated features, all stub headings covered,
                 valid sexpr syntax in examples, user-first tone.
                 Stubs: {{step \"stubs\"}}
                 Generated: {{step \"enrich-docs\"}}
                 Respond PASS or FAIL with specific findings.")))

  (phase "output"
    (step "save-docs" (run "cp generated.md site/src/content/docs/"))
    (step "save-report" (run "cp build-report.md site/"))))
```

### Execution Model

```
Phase "generate" (:retries 2)

  Attempt 1:
    run step "enrich-docs"    -> output stored as steps["enrich-docs"]
    run gate "syntax-check"   -> FAIL, output stored as steps["syntax-check"]
    (remaining gates skipped on first failure)

  Attempt 2:
    run step "enrich-docs"    -> prompt references {{step "syntax-check"}} from attempt 1
    run gate "syntax-check"   -> PASS
    run gate "build-check"    -> PASS
    run gate "diff-review"    -> PASS
    -> phase passes, continue to next phase/step

  If attempt 3 also fails:
    -> workflow halts, verification report emitted
```

Key behaviors:

1. **Steps run first, then gates.** Within a phase, all steps execute in order, then all gates execute in order. Gates validate the work steps produced.
2. **First gate failure skips remaining gates.** No point running the diff-review if the syntax check failed.
3. **Gate output persists across retries.** On retry, the step store still contains the previous gate output. LLM steps can reference `{{step "gate-id"}}` to see what went wrong and self-correct.
4. **Phase-scoped retry.** Only the failing phase re-runs. Prior phases and their outputs are stable.
5. **Step store is global.** Steps and gates from all phases share one store, same as today. A gate in phase "generate" can reference a step from phase "gather."

### Gate Pass/Fail Semantics

- **Shell gates:** Exit 0 = PASS. Non-zero = FAIL. Stderr is captured as failure detail and stored in the step store.
- **LLM gates:** Output is parsed for PASS/FAIL using the same logic as `batch/manifest.go`'s `ParseReview`. This keeps scoring compatible with `--compare` manifests.

Both gate types write their full output to the step store under their ID regardless of pass/fail.

### Verification Report

When a phase exhausts all retries, the runner emits a structured report and halts the workflow:

```
Phase "generate" FAILED (3/3 attempts exhausted)

Gate results (final attempt):
  syntax-check: PASS
  build-check:  FAIL - exit 1: Missing frontmatter in generated.md
  diff-review:  (skipped - prior gate failed)

Workflow halted. See gate outputs in step store for details.
```

This becomes the workflow's result output, so `--compare` manifests can reference it.

### Interaction with --compare

No changes to batch orchestration. Each variant workflow self-verifies through its own gates before `--compare` runs the cross-review. This means:

- Cross-review compares gate-verified outputs (higher baseline quality)
- Cross-review focuses on relative quality ("which variant is best") not absolute correctness ("is this even valid")
- A variant whose gates fail gets a natural zero score in the manifest — it produced a verification report, not a result

### Interaction with Tiered Runner

Phases/gates and tiered routing operate at different levels and compose naturally:

- **Tiered runner** (per-step): picks the cheapest provider that produces structurally valid output for a single LLM step. Escalates local -> paid on structural failure or low self-eval.
- **Gates** (per-phase): verify the combined output of a phase actually accomplished the goal. Catch semantic failures that individual step quality can't guarantee.

`RunSmart`'s structural validation is complementary, not redundant — it prevents garbage from reaching the step store. Gates verify the end-to-end result.

### Interaction with Existing Forms

All existing step forms (`cond`, `map`, `catch`, `retry`, `timeout`) work inside phases unchanged. They apply to individual steps within a phase. Phase-level retry is orthogonal — it re-runs the whole phase, which may contain steps with their own retry/catch logic.

### Parser Changes

The sexpr parser needs to recognize two new forms:

- `(phase "name" [:retries N] ...children)` — children are step or gate nodes
- `(gate "name" (run ...) | (llm ...))` — identical structure to step, tagged as gate

The `Step` struct gains a `IsGate bool` field. The `Workflow` struct gains a `Phases []Phase` field alongside the existing `Steps []Step` for bare steps.

```go
type Phase struct {
    ID      string
    Retries int
    Steps   []Step // work steps
    Gates   []Step // verification steps (IsGate = true)
}
```

### Runner Changes

New `executePhase()` function:

1. Execute all phase steps in order (using existing `executeStep`)
2. Execute all phase gates in order
3. If a gate fails: check retry budget
   - Budget remaining: re-run from step 1
   - Budget exhausted: emit verification report, return error
4. All gates pass: continue to next phase/step

The main `Run()` function walks the workflow's top-level items (bare steps and phases) in order, delegating to `executeStep` or `executePhase` as appropriate.

### What Changes

| Component | Change |
|-----------|--------|
| `internal/pipeline/sexpr.go` | Parse `phase` and `gate` forms |
| `internal/pipeline/runner.go` | Add `executePhase()`, update `Run()` to handle phases |
| `internal/pipeline/types.go` | Add `Phase` struct, `IsGate` field on `Step` |
| `internal/batch/manifest.go` | No changes — `ParseReview` already works for gate output |

### What Doesn't Change

- Bare steps outside phases
- Batch/compare orchestration (`internal/batch/`)
- Tiered routing and self-eval (`internal/provider/tiers.go`)
- Template rendering
- Telemetry indexing
- All existing workflow files (no phases = same behavior)

## Out of Scope

- Gate-to-gate dependencies (gates within a phase always run in declaration order)
- Parallel gate execution (gates are sequential; keeps failure-skipping simple)
- Cross-phase retry (only the failing phase retries, not the whole workflow)
- Nested phases (phases cannot contain phases)
- `--compare` changes (it already benefits from self-verifying workflows)

## Examples

### Minimal: Single gate

```scheme
(workflow "deploy-check"
  (phase "deploy" :retries 1
    (step "apply" (run "kubectl apply -f manifest.yaml"))
    (gate "health" (run "kubectl rollout status deploy/app --timeout=60s"))))
```

### Research with quality gate

```scheme
(workflow "research"
  (phase "gather"
    (step "search" (run "glitch search --query '{{.param.query}}'")))

  (phase "analyze" :retries 1
    (step "summary"
      (llm :prompt "Summarize these findings: {{step \"search\"}}
                    Previous feedback: {{step \"fact-check\"}}"))
    (gate "fact-check"
      (llm :tier 2
        :prompt "Verify claims in this summary against the source.
                 Summary: {{step \"summary\"}}
                 Source: {{step \"search\"}}
                 Respond PASS or FAIL with specific findings."))))
```

### Mixed bare steps and phases

```scheme
(workflow "mixed"
  ;; bare step, no gates
  (step "setup" (run "echo starting"))

  ;; gated phase
  (phase "core" :retries 2
    (step "generate" (llm :prompt "..."))
    (gate "validate" (run "check-output.sh")))

  ;; bare step after gated phase
  (step "cleanup" (run "echo done")))
```
