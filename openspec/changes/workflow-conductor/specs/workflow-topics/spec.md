## ADDED Requirements

### Requirement: Workflow BUSD topic constants

The file `internal/busd/topics/topics.go` SHALL export the following string constants in a new `Workflow*` group. No existing constants SHALL be modified or removed.

| Constant              | Value                        |
|-----------------------|------------------------------|
| `WorkflowRunStarted`  | `"workflow.run.started"`     |
| `WorkflowRunCompleted`| `"workflow.run.completed"`   |
| `WorkflowRunFailed`   | `"workflow.run.failed"`      |
| `WorkflowStepStarted` | `"workflow.step.started"`    |
| `WorkflowStepDone`    | `"workflow.step.done"`       |
| `WorkflowStepFailed`  | `"workflow.step.failed"`     |

#### Scenario: Constants compile without conflict
- **WHEN** `internal/busd/topics/topics.go` is compiled after adding the new constants
- **THEN** the package compiles without error and no existing constant values are changed

#### Scenario: Switchboard wildcard subscription captures workflow events
- **WHEN** a BUSD subscriber registers for topic prefix `workflow.*`
- **THEN** it receives all six workflow event types published during a conductor run
