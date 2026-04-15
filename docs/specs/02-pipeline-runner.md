# Pipeline Runner Specification

## Overview

This document defines the execution semantics of the gl1tch pipeline runner — how workflows are executed, how data flows between steps, and how error recovery works.

## Definitions

- **Run** — a single execution of a workflow, identified by a unique `RunID`.
- **Steps map** — the `map[string]string` that accumulates step outputs, keyed by step ID.
- **Outcome** — the internal result of a single step execution, including output, token counts, cost, and escalation metadata.

## Execution Model

### Step Ordering

Steps MUST execute sequentially in declaration order (top to bottom). There is no implicit parallelism.

When a workflow has `Items` (phases and steps interleaved), the runner walks the `Items` list. When a workflow has only flat `Steps`, it walks the `Steps` list. These are mutually exclusive paths — `Items` takes priority when populated.

### Step Output

Every step produces a string output. The output is stored in `steps[step.ID]` immediately after successful execution.

`Result.Output` is the output of the last step in the workflow. For phased workflows, this is the last step in the last phase (work steps only, not gates).

Steps with no meaningful output (e.g., a `save` that writes to disk) still produce an output (the written content). A step that produces nothing stores an empty string.

### Data Flow

A step MUST be able to reference any prior step's output via `{{step "id"}}`. "Prior" means any step whose ID is already in the steps map at the time of template rendering.

A reference to a step that hasn't executed yet renders as an empty string. This is not an error — it's a silent empty substitution. This behavior exists because Go's `text/template` returns zero values for missing map keys.

`{{stepfile "id"}}` writes the referenced step's output to a temporary file and returns the file path. This MUST be used when step output contains characters that would break shell escaping (quotes, newlines, special characters). The temp file is not cleaned up automatically.

### Seed Steps

`RunOpts.SeedSteps` pre-populates the steps map before execution begins. Steps whose IDs match a seed key are skipped entirely — their seeded output is used by subsequent steps.

This is used by batch mode: data-gathering shell steps run once, then all variant runs are seeded with those results.

## Retry Semantics

`Step.Retry` specifies the maximum number of retries (not total attempts). Total attempts = `1 + Retry`.

On each attempt:
- The step re-executes from scratch (shell command re-runs, LLM re-prompts)
- If the attempt succeeds, the output is stored and execution continues
- If the attempt fails, the next attempt begins

If all attempts fail, the error from the last attempt propagates.

### Interaction with Timeout

If a step has both `Retry` and `Timeout`, they share the same deadline context. The timeout applies to ALL attempts combined, not per-attempt. Once the deadline expires:
- The current attempt fails immediately
- No further retries are attempted
- The timeout error propagates

## Timeout Semantics

`Step.Timeout` is a Go duration string (e.g., `"30s"`, `"2m"`, `"1h"`). Invalid duration strings are a runtime error (not a parse error).

A timeout creates a `context.WithTimeout`. The context is passed to the step execution. When the context expires:
- Shell commands are killed
- LLM calls fail with context error
- The step fails with the context error

## Catch Semantics

When `Step.Form == "catch"`:
1. The primary step executes (including its own retries if any)
2. If the primary succeeds, execution continues normally — fallback is never touched
3. If the primary fails after all retries, the fallback step executes with a fresh context (no timeout from the primary)
4. Both `step.ID` and `fallback.ID` are set to the fallback's output in the steps map
5. If the fallback also fails, that error propagates

The dual-write to both IDs means downstream steps can reference either the primary or fallback ID and get the same result.

## Cond Semantics

When `Step.Form == "cond"`:
1. Predicates are evaluated in declaration order
2. Each predicate is a shell command, template-rendered before execution
3. Exit code 0 = true (predicate matches), non-zero = false
4. The first matching predicate's step executes via the normal `executeStep` path (retries, timeouts apply)
5. `"else"` always matches — it SHOULD be the last branch
6. If no branch matches, the cond step's output is empty string and no error is raised
7. The cond step ID is auto-generated as `cond-{line}` if not explicitly set

## Map Semantics

When `Step.Form == "map"`:
1. The source step's output is retrieved from the steps map
2. If the source step has no output, this is a runtime error
3. Output is split by newlines; empty lines are skipped
4. For each non-empty line:
   - The body step is cloned with ID `{body.ID}-{index}`
   - `item` and `item_index` are injected into the params map
   - The cloned step executes via the normal `executeStep` path
   - Iteration is sequential — items are NOT processed in parallel
5. All iteration outputs are collected; the map step's output is newline-joined
6. The map step ID is auto-generated as `map-{line}` if not explicitly set

Original params are restored after each iteration — map params don't leak.

## Phase Execution

When a workflow item is a Phase:
1. All work steps execute sequentially
2. All gate steps execute sequentially after work steps complete
3. Gate evaluation:
   - **Shell gates:** non-zero exit code = failure
   - **LLM gates:** response text is checked for "PASS" or "FAIL" keywords (case-insensitive, markdown bold markers stripped)
4. If any gate fails:
   - If retries remain, the entire phase re-executes from step 1
   - Steps map is NOT cleared between retries — work steps re-execute and overwrite their outputs
5. If all retries exhausted with gate failures:
   - A `VerificationReport` is generated with per-gate results
   - The report is stored as `steps["_verification_report"]`
   - The workflow errors with the phase failure

### Gate Evaluation Detail

LLM gate responses are evaluated by scanning for keywords:
- Contains "PASS" (case-insensitive, after stripping `*`) → gate passes
- Contains "FAIL" → gate fails
- Contains neither → gate fails (fail-safe default)

## Error Propagation

All errors halt the workflow immediately. There is no "continue on error" mode.

Errors are wrapped with context at each level:
- Step level: `"step {id}: {error}"`
- Phase level: `"phase {id}: {error}"`
- Fallback level: `"step {id} fallback {fallback.id}: {error}"`

There are no custom error types. All errors are `error` interface values wrapped with `fmt.Errorf`. Callers cannot programmatically distinguish error kinds (e.g., timeout vs. provider failure) without string matching.

## Result

A successful workflow run returns:

```go
Result {
    Workflow string            // workflow name
    Output   string            // last step's output
    Steps    map[string]string // all step outputs keyed by ID
}
```

The Steps map includes:
- All explicitly defined step outputs
- Map iteration step outputs (with `{body.ID}-{index}` keys)
- Fallback outputs (keyed by both primary and fallback IDs)
- Seeded step outputs
- `_verification_report` if any phase produced one

## Conformance

See [`spec/02-pipeline-runner.glitch`](../../spec/02-pipeline-runner.glitch) for the conformance workflow that exercises data flow, retry, timeout, catch, cond, map, and phase semantics.
