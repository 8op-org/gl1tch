# `(par ...)` — Parallel Step Execution

**Date:** 2026-04-15
**Status:** Approved
**Form count after:** ~13 (under 15-form threshold)

## Summary

Add `(par ...)` as a new compound form to the gl1tch workflow engine. All child steps inside a `par` block run concurrently. Steps after the block can reference any parallel step's output. Fail-fast semantics: first error cancels all siblings.

## Motivation

The workflow engine runs steps sequentially. Independent work — fan-out to multiple providers, parallel data gathering, concurrent gate checks — waits unnecessarily. The `site-update` workflow has two independent chains (doc enrichment and changelog generation) that could run simultaneously. Phase gates are also independent and bottleneck on the slowest one.

## Syntax

```scheme
;; Top-level parallel steps
(par
  (step "a" (run "curl ..."))
  (step "b" (llm :prompt "..."))
  (retry 2 (step "c" (run "..."))))

;; Parallel gates inside a phase
(phase "verify" :retries 1
  (step "work" (run "..."))
  (par
    (gate "g1" (run "..."))
    (gate "g2" (run "..."))
    (gate "g3" (run "..."))))

;; Composes with timeout
(timeout "30s"
  (par
    (step "a" (run "curl ..."))
    (step "b" (run "curl ..."))))

;; Composes with retry on individual children
(par
  (retry 2 (step "a" (llm ...)))
  (step "b" (run "...")))
```

Children of `par` are any form that `convertForm` handles: `step`, `retry`, `timeout`, `catch`, `cond`, `map`, and nested `par`.

## Semantics

- All children start concurrently via goroutines.
- Outputs stored under each child's own step ID — flat, no namespace. `{{step "a"}}` works identically whether `a` ran inside `par` or sequentially.
- **Fail-fast:** First error cancels all sibling goroutines via context cancellation. The workflow halts with that error.
- Steps *after* the `par` block can reference any parallel step output via `{{step "id"}}`.
- Steps *inside* a `par` block must not reference sibling steps (they haven't completed yet). Template rendering produces an empty string for unresolved refs. Parse-time validation deferred for v1.

## Data Structures

### types.go

Add one field to `Step`:

```go
ParSteps []Step `yaml:"-"` // par: concurrent child steps
```

`Form` field gets value `"par"` (alongside `"cond"`, `"map"`, `"catch"`).

No new types. `par` is a `Step` with `Form: "par"` and children in `ParSteps`. Slots into `WorkflowItem`, phase step/gate lists, and map bodies without structural changes.

### runCtx

Add mutex for concurrent step output writes:

```go
type runCtx struct {
    // ... existing fields ...
    mu sync.Mutex // guards steps map during par execution
}
```

All writes to `rctx.steps` during `executePar` go through lock/unlock. Sequential execution is unaffected — mutex is uncontended.

## Parser Changes (sexpr.go)

### convertForm

New case:

```go
case "par":
    s, err := convertPar(n, defs)
    if err != nil {
        return nil, err
    }
    return []Step{s}, nil
```

### convertPar

1. Iterate children after the `par` symbol.
2. Each child dispatched through `convertForm` (recursive — `retry`, `timeout`, `catch` all work inside `par`).
3. Collect resulting steps into `ParSteps`.
4. Reject blocks with fewer than 2 children (parse error).
5. Return `Step{Form: "par", ID: "par-N", ParSteps: collected}` with auto-generated ID.

### Phase support

`convertPhase` already iterates child forms. A `(par ...)` inside a phase produces a single step with `Form: "par"`. Gates inside `par` get `IsGate: true` from the `(gate ...)` converter — no special handling needed.

## Runner Changes (runner.go)

### executePar

```go
func executePar(ctx context.Context, rctx *runCtx, steps []Step) ([]stepOutcome, error) {
    g, gctx := errgroup.WithContext(ctx)
    outcomes := make([]stepOutcome, len(steps))
    for i, s := range steps {
        i, s := i, s
        g.Go(func() error {
            outcome, err := executeStep(rctx.withContext(gctx), s)
            if err != nil {
                return err
            }
            rctx.mu.Lock()
            rctx.steps[s.ID] = outcome.output
            rctx.mu.Unlock()
            outcomes[i] = outcome
            return nil
        })
    }
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return outcomes, nil
}
```

### Integration points

1. **`executeStep`** — add `"par"` case alongside `"cond"` and `"map"`, dispatch to `executePar`.
2. **`runBareStep`** — already calls `executeStep`, so `par` steps flow through naturally.
3. **Context threading** — `runCtx` does not currently carry a `context.Context`. Thread one through: `Run()` creates a background context, `executeStep` passes it down, `executePar` derives a cancellable child via `errgroup.WithContext`.

### Telemetry

Each goroutine populates its slot in a `[]stepOutcome` array (index-based, no mutex needed for the array). After `g.Wait()` returns, accumulate totals and index LLM calls sequentially. This avoids concurrent ES writes.

## Testing

### Parser tests (sexpr_test.go)

| Test | Verifies |
|------|----------|
| `TestSexprWorkflow_Par` | Basic `(par (step ...) (step ...))` parses to `Form:"par"` with 2 `ParSteps` |
| `TestSexprWorkflow_ParInPhase` | `par` wrapping gates inside a phase |
| `TestSexprWorkflow_ParWithRetry` | `(par (retry 2 (step ...)) (step ...))` composes |
| `TestSexprWorkflow_ParSingleChild` | Rejected at parse time (needs >= 2 children) |

### Runner tests (runner_test.go)

| Test | Verifies |
|------|----------|
| `TestRun_Par_Basic` | Two shell steps run, both outputs available to subsequent step via `{{step "id"}}` |
| `TestRun_Par_FailFast` | One step fails, other is cancelled, workflow returns error |
| `TestRun_Par_InPhase` | Parallel gates inside a phase, all pass |
| `TestRun_Par_WithTimeout` | `(timeout "1s" (par ...))` cancels slow children |
| `TestRun_Par_Telemetry` | LLM steps inside `par` accumulate tokens/cost correctly |

Tests use shell steps (`echo`, `sleep`, `exit 1`) to verify concurrency and cancellation without real LLM providers.

## Files Changed

| File | Change |
|------|--------|
| `internal/pipeline/types.go` | Add `ParSteps []Step` to `Step` |
| `internal/pipeline/sexpr.go` | Add `convertPar`, wire into `convertForm` |
| `internal/pipeline/runner.go` | Add `executePar`, mutex on `runCtx`, context threading, telemetry collection |
| `internal/pipeline/sexpr_test.go` | 4 parser tests |
| `internal/pipeline/runner_test.go` | 5 runner tests |
| `go.mod` | Add `golang.org/x/sync` for `errgroup` (if not already present) |
