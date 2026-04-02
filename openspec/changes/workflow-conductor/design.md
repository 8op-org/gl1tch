## Context

gl1tch's pipeline runner (`internal/pipeline/runner.go`) executes a single named DAG. It is well-tested and used everywhere. Adding multi-pipeline coordination directly to the runner would bloat it and risk regressions. Instead, the orchestrator sits **above** the runner as a thin coordination layer that composes existing primitives without modifying them.

The token-score-gamification spec established the pattern of using Ollama with `format:json` for structured LLM decisions (achievement evaluation, narration). The orchestrator reuses this exact pattern for `decision` step routing — Ollama is the only LLM in the hot path, and its output is always validated JSON.

## Goals / Non-Goals

**Goals:**
- Coordinate multi-pipeline and multi-agent workflows via a declarative YAML schema
- Use Ollama `format:json` for runtime branching decisions (no remote LLM in the decision path)
- Maintain a shared `WorkflowContext` (key-value store) so step outputs are available to later steps via `{{ctx.<step_id>.output}}`
- Publish BUSD events for workflow and step lifecycle (`workflow.run.*`, `workflow.step.*`)
- Checkpoint full workflow state at every step boundary; support `glitch workflow resume --run-id`
- Additive only: zero changes to `pipeline.Run()`, `executor.Manager`, or existing BUSD topics

**Non-Goals:**
- Distributed / multi-host execution
- Replacing the pipeline runner for single-pipeline use cases
- A visual workflow editor or DSL beyond YAML
- Real-time streaming aggregation across parallel branches (outputs are joined after all branches complete)

## Decisions

### D1: Orchestrator as a thin wrapper over `pipeline.Run()`, not a rewrite

**Decision**: `StepDispatcher` calls `pipeline.Run()` for `pipeline-ref` steps and an `AgentRunner` for `agent-ref` steps. The orchestrator never re-implements step execution.

**Alternatives considered**:
- Extend `pipeline.Runner` to support workflow concepts directly — rejected because it entangles two concerns and would require modifying tested code.
- Embed pipeline YAML inline in workflow YAML — rejected because it duplicates the pipeline schema and breaks reuse of existing `.pipeline.yaml` files.

**Rationale**: Keeps pipeline.Run() the single source of truth for step execution. The orchestrator only handles sequencing, context threading, and BUSD events at the workflow level.

---

### D2: Ollama `format:json` for decision nodes, never freeform

**Decision**: `DecisionNode.Evaluate()` calls Ollama's `/api/generate` with `format: "json"` and a required `branch` field in the response schema. If parsing fails, the workflow fails fast (no silent fallback to a default branch).

**Alternatives considered**:
- Parse freeform LLM output for branch name — rejected because freeform output is unreliable and changes with model version.
- Use a Go decision function instead of LLM — rejected because it would require recompilation to change routing logic; Ollama prompts in YAML are operator-configurable.

**Rationale**: Consistent with the token-score-gamification narration pattern. Failing fast on bad JSON surfaces model/prompt issues immediately.

---

### D3: `WorkflowContext` is an in-memory map serialized to SQLite at each checkpoint

**Decision**: `WorkflowContext` is a `map[string]string` held in memory during a run. At each step boundary, the full context is JSON-serialized into the `workflow_checkpoints` table alongside the step ID and status.

**Alternatives considered**:
- Store each context key as a separate row — more queryable but requires schema changes every time a new key type is added.
- Keep context only in memory — no resume support; rejected because single-pipeline resume already exists in gl1tch.

**Rationale**: Simple, uniform, and maps cleanly onto the existing `step_checkpoints` pattern. Resume is implemented by deserializing the last checkpoint row.

---

### D4: `parallel` blocks fan-out via goroutines, join before proceeding

**Decision**: `parallel` step blocks launch each branch as a goroutine calling `StepDispatcher.Dispatch()`. The conductor waits for all branches with `errgroup`. Branch outputs are merged into `WorkflowContext` keyed by branch step ID. First error cancels remaining branches.

**Alternatives considered**:
- Sequential execution of parallel blocks (simpler) — rejected; the whole point of `parallel` is concurrency.
- Allow partial failure continuation — deferred to a future `allow_failure` flag per branch.

**Rationale**: `errgroup` with context cancellation is idiomatic Go and matches how the pipeline runner handles parallel DAG steps.

---

### D5: New BUSD topic namespace `workflow.*` — no collision with `pipeline.*`

**Decision**: Add `workflow.run.started`, `workflow.run.completed`, `workflow.run.failed`, `workflow.step.started`, `workflow.step.done`, `workflow.step.failed` as constants to `internal/busd/topics/topics.go`. The Switchboard TUI subscribes to `workflow.*` separately.

**Rationale**: Pipeline events remain pipeline events. Workflow events are higher-level and consumers (TUI, cron, scoring) can subscribe selectively.

---

### D6: Game scoring aggregated at workflow scope

**Decision**: `ConductorRunner` accumulates `TokenUsage` from each `pipeline-ref` step's `ScoreTeeWriter` (via a channel). After the final step completes, it publishes a single `game.run.scored` event with the aggregate token usage. Per-pipeline `game.run.scored` events are suppressed when run inside a workflow (pipeline-level game mode disabled via `game: false` injected into run options).

**Rationale**: Consistent with the token-score spec's requirement of "exactly one event per run." A workflow is a single logical run.

## Risks / Trade-offs

- **Ollama unavailability on `decision` steps** → Workflow fails at the decision node. Mitigation: decision nodes should have a `default_branch` field used when Ollama returns an error.
- **Large `WorkflowContext` at checkpoint** → JSON blob can grow large if steps produce verbose outputs. Mitigation: context values are truncated to 16 KB per key at write time; full output remains in pipeline run store.
- **Agent-ref steps have no native executor** → `agent-ref` requires the APM pipeline for that agent to be materialized. If the agent is not installed, the step fails immediately. Mitigation: `StepDispatcher` checks for the materialized pipeline file and returns a clear error before dispatch.
- **No visual progress for parallel branches in TUI** → The Switchboard only shows workflow-level events today; parallel branch progress is invisible. Mitigation: each branch publishes `workflow.step.*` events so the activity feed shows them.

## Migration Plan

1. Add `workflow_runs` and `workflow_checkpoints` tables via additive migration in `store.Open()` — no data loss, backward compatible.
2. Add topic constants to `topics.go` — compile-time safe, existing subscribers unaffected.
3. Register `workflowCmd` in `cmd/root.go`.
4. No config migration required; workflow files live in `~/.config/glitch/workflows/` which is created on first use.
5. Rollback: remove `cmd/workflow.go` and `internal/orchestrator/` — no other code references them.

## Open Questions

- Should `decision` nodes support a `timeout` field to cap Ollama latency? (Likely yes — default 30s.)
- Should `workflow resume` re-run the failed step or skip to the next? (Current plan: re-run from failed step.)
- Should `agent-ref` steps support inline `vars` override the same way pipeline steps do?
