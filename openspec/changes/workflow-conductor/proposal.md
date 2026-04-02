## Why

gl1tch's pipeline runner executes a single DAG in isolation — there is no layer above it that can sequence multiple pipelines, dispatch to APM agents as first-class steps, or use Ollama to make branching decisions at runtime. As workflows grow more complex (e.g., triage → remediate → notify), users are forced to hand-stitch pipelines together in shell scripts, losing checkpointing, brain context, and game scoring in the process.

## What Changes

- **New** `internal/orchestrator/` package: `WorkflowDef`, `ConductorRunner`, `WorkflowContext`, `DecisionNode`, `StepDispatcher`
- **New** `~/.config/glitch/workflows/*.workflow.yaml` schema with step types: `pipeline-ref`, `agent-ref`, `decision`, `parallel`
- **New** `cmd/workflow.go`: `glitch workflow run <name> [--input "..."]` CLI entry point
- **New** BUSD topic namespace: `workflow.run.*` and `workflow.step.*` constants in `internal/busd/topics/topics.go`
- **New** SQLite tables: `workflow_runs`, `workflow_checkpoints` added via additive migration in `internal/store/schema.go`
- **No changes** to `pipeline.Run()`, `executor.Manager`, or existing BUSD topics

## Capabilities

### New Capabilities

- `workflow-def`: YAML schema for multi-step workflows referencing pipelines, agents, decisions, and parallel blocks; loaded and validated into a `WorkflowDef` Go struct
- `conductor-runner`: Executes a `WorkflowDef` graph — drives step dispatch, publishes BUSD events, checkpoints state at every step boundary, supports resume
- `workflow-context`: Shared key-value store threaded through all steps; pipeline outputs written in as `{{ctx.<step_id>.output}}`; readable by subsequent step prompts via template expansion
- `decision-node`: Step type that calls Ollama with `format:json` and a user-defined prompt+schema; structured JSON output drives branch selection declared in the YAML
- `workflow-topics`: New BUSD topic constants `workflow.run.started`, `workflow.run.completed`, `workflow.run.failed`, `workflow.step.started`, `workflow.step.done`, `workflow.step.failed`

### Modified Capabilities

_(none — existing pipeline, executor, and store interfaces unchanged)_

## Impact

- `internal/store/schema.go`: additive migration adds `workflow_runs` and `workflow_checkpoints` tables
- `internal/busd/topics/topics.go`: new constants added (no removals)
- `cmd/root.go`: `workflowCmd` registered as a subcommand
- Depends on existing `pipeline.Run()`, `executor.Manager`, `store.Store`, `busd.ConnectedClient`, and `brainrag.BrainInjector`
- No new external dependencies; Ollama called via existing HTTP pattern used by `internal/router` and `internal/brainrag`
