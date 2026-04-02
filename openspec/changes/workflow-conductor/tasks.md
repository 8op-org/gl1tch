## 1. Foundation — Store & Topics

- [x] 1.1 Add `workflow_runs` and `workflow_checkpoints` tables to `internal/store/schema.go` via additive migration
- [x] 1.2 Add store methods: `CreateWorkflowRun`, `CompleteWorkflowRun`, `FailWorkflowRun`, `GetWorkflowRun`, `SaveWorkflowCheckpoint`, `LoadWorkflowCheckpoints`
- [x] 1.3 Add `WorkflowRunStarted`, `WorkflowRunCompleted`, `WorkflowRunFailed`, `WorkflowStepStarted`, `WorkflowStepDone`, `WorkflowStepFailed` constants to `internal/busd/topics/topics.go`

## 2. WorkflowDef — YAML Schema & Loader

- [x] 2.1 Define `WorkflowDef`, `WorkflowStep`, `ParallelBranch`, `DecisionStep` structs in `internal/orchestrator/def.go`
- [x] 2.2 Implement `LoadWorkflow(r io.Reader) (*WorkflowDef, error)` using `gopkg.in/yaml.v3` in `internal/orchestrator/loader.go`
- [x] 2.3 Implement `(*WorkflowDef).Validate()` — checks unique step IDs, valid step types, decision `on` map non-empty
- [x] 2.4 Implement `FindWorkflow(name string) (string, error)` in `internal/orchestrator/loader.go` — resolves name to `~/.config/glitch/workflows/<name>.workflow.yaml`

## 3. WorkflowContext

- [x] 3.1 Implement `WorkflowContext` struct in `internal/orchestrator/context.go` — `sync.RWMutex`-protected `map[string]string` with `Set`, `Get`, `Marshal`, `Unmarshal`
- [x] 3.2 Enforce 16 KB truncation in `Set()` with stderr warning
- [x] 3.3 Implement `ExpandTemplate(s string, wctx *WorkflowContext) string` — replaces `{{ctx.<key>}}` references; unknown keys expand to `""`

## 4. DecisionNode

- [x] 4.1 Implement `DecisionNode` struct in `internal/orchestrator/decision.go` with `Evaluate(ctx context.Context, wctx *WorkflowContext) (branch string, err error)`
- [x] 4.2 POST to Ollama `/api/generate` with `format: "json"`, `stream: false`; apply configurable timeout (default 30s) via `context.WithTimeout`
- [x] 4.3 Validate response JSON has `branch` string field; return typed errors for missing/wrong-type field
- [x] 4.4 Apply `default_branch` fallback in `ConductorRunner` (not in `DecisionNode` itself)

## 5. StepDispatcher

- [x] 5.1 Implement `StepDispatcher` in `internal/orchestrator/dispatcher.go` with `Dispatch(ctx, step WorkflowStep, wctx *WorkflowContext, mgr *executor.Manager, opts ...pipeline.RunOption) (string, error)`
- [x] 5.2 `pipeline-ref` dispatch: resolve pipeline file, expand `input` template, call `pipeline.Run()` with game mode suppressed
- [x] 5.3 `agent-ref` dispatch: resolve `apm.<name>.pipeline.yaml`, delegate to same pipeline.Run() path
- [x] 5.4 Write step output into `WorkflowContext` under `<step_id>.output` after successful dispatch

## 6. ConductorRunner

- [x] 6.1 Implement `ConductorRunner` struct in `internal/orchestrator/conductor.go` with functional options: `WithStore`, `WithBusPublisher`, `WithExecutorManager`, `WithBrainInjector`
- [x] 6.2 Implement `(*ConductorRunner).Run(ctx, def *WorkflowDef, input string) (string, error)` — sequential execution, decision branching, parallel fan-out via `errgroup`
- [x] 6.3 Publish `workflow.run.started` before first step; `workflow.run.completed` or `workflow.run.failed` after last
- [x] 6.4 Publish `workflow.step.started` / `workflow.step.done` / `workflow.step.failed` around each `StepDispatcher.Dispatch()` call
- [x] 6.5 Write checkpoint after every step (success and failure) using store methods from task 1.2
- [x] 6.6 Implement `(*ConductorRunner).Resume(ctx, runID int64) (string, error)` — load checkpoints, restore `WorkflowContext`, re-execute from last failed step
- [x] 6.7 Implement game aggregation: accumulate `TokenUsage` from each step via channel; publish single `game.run.scored` after workflow completes; suppress per-pipeline game events

## 7. CLI

- [x] 7.1 Create `cmd/workflow.go` with `workflowCmd` (`glitch workflow`) and subcommands `run` and `resume`
- [x] 7.2 `glitch workflow run <name> [--input "..."]`: load workflow, build executor manager (same as `pipeline run`), construct `ConductorRunner`, call `Run()`
- [x] 7.3 `glitch workflow resume --run-id <id>`: open store, call `ConductorRunner.Resume()`
- [x] 7.4 Register `workflowCmd` in `cmd/root.go`

## 8. Tests

- [x] 8.1 Table-driven unit tests for `WorkflowDef.Validate()` covering duplicate IDs, unknown types, empty decision `on` map
- [x] 8.2 Unit tests for `WorkflowContext`: Set/Get, truncation at 16 KB, Marshal/Unmarshal round-trip, template expansion
- [x] 8.3 Unit tests for `DecisionNode.Evaluate()` using an `httptest.Server` stub for Ollama — covers success, timeout, missing branch field, non-string branch, HTTP error
- [x] 8.4 Integration test for `ConductorRunner.Run()` using `StubExecutor` pipelines — covers sequential, decision branching, parallel fan-out, and context propagation
- [x] 8.5 Integration test for `ConductorRunner.Resume()` — simulate failure at step 2 of 3, verify step 1 skipped and step 2 re-executed with restored context
