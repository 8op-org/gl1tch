## ADDED Requirements

### Requirement: ConductorRunner executes a WorkflowDef

`ConductorRunner.Run(ctx, def WorkflowDef, input string, opts ...ConductorOption)` SHALL execute the workflow's steps in declaration order, skipping steps whose `id` does not appear in the execution path determined by decision nodes. It MUST return the final step's output string and any error.

#### Scenario: Sequential steps run in order
- **WHEN** a workflow declares steps A, B, C with no decision nodes
- **THEN** A runs first, then B, then C; each step's output is in `WorkflowContext` before the next starts

#### Scenario: Context cancellation stops execution
- **WHEN** the caller cancels the context mid-workflow
- **THEN** the current step's pipeline run is cancelled and `Run()` returns `context.Canceled`

### Requirement: BUSD events published for workflow lifecycle

`ConductorRunner` SHALL publish `workflow.run.started` before the first step and `workflow.run.completed` or `workflow.run.failed` after the final step. It SHALL publish `workflow.step.started` and `workflow.step.done` or `workflow.step.failed` around each step execution. All payloads SHALL be JSON.

#### Scenario: Run started event payload
- **WHEN** a workflow run begins
- **THEN** `workflow.run.started` is published with `{"workflow_name": "<name>", "run_id": <id>}`

#### Scenario: Step done event payload
- **WHEN** a step completes successfully
- **THEN** `workflow.step.done` is published with `{"step_id": "<id>", "run_id": <id>, "duration_ms": <n>}`

#### Scenario: BUSD unavailable does not block execution
- **WHEN** no BUSD publisher is configured
- **THEN** the workflow runs to completion without error; events are silently dropped

### Requirement: Checkpoint at every step boundary

After each step completes (success or failure), `ConductorRunner` SHALL write the current `WorkflowContext` and the completed step's ID to `workflow_checkpoints` in SQLite. On failure, the checkpoint records the failed step ID so resume can restart from it.

#### Scenario: Checkpoint written after step success
- **WHEN** a step finishes successfully
- **THEN** a row is inserted into `workflow_checkpoints` with `status="done"`, the step ID, and the serialized context JSON

#### Scenario: Checkpoint written after step failure
- **WHEN** a step returns an error
- **THEN** a row is inserted into `workflow_checkpoints` with `status="failed"` and the failed step ID

### Requirement: Resume from last checkpoint

`glitch workflow resume --run-id <id>` SHALL reload the `workflow_checkpoints` for that run, restore `WorkflowContext`, and re-execute from the last failed or incomplete step.

#### Scenario: Resume restores context and reruns failed step
- **WHEN** a run failed at step B and the user runs `glitch workflow resume --run-id <id>`
- **THEN** the runner restores context from the last checkpoint, skips step A, and re-executes step B

#### Scenario: Resume of completed run is rejected
- **WHEN** the user attempts to resume a run with `status="completed"` in `workflow_runs`
- **THEN** the command returns an error: "run <id> is already complete"

### Requirement: Game scoring aggregated at workflow scope

When game mode is enabled, `ConductorRunner` SHALL suppress per-pipeline `game.run.scored` events and instead publish a single aggregate `game.run.scored` event after the workflow completes, with `TokenUsage` summed across all `pipeline-ref` and `agent-ref` steps.

#### Scenario: Aggregate game event published after workflow
- **WHEN** two pipeline-ref steps each consume tokens and the workflow completes
- **THEN** exactly one `game.run.scored` event is published with the sum of both steps' token usage

#### Scenario: Game mode disabled suppresses all game events
- **WHEN** `game.enabled: false` is set in config
- **THEN** no `game.run.scored` event is published at any scope
