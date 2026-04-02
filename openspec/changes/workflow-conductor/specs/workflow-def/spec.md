## ADDED Requirements

### Requirement: Workflow YAML schema

The system SHALL load and validate `.workflow.yaml` files from `~/.config/glitch/workflows/`. A workflow file MUST declare a `name`, a `version`, and a top-level `steps` list. Each step MUST have an `id` (unique within the workflow) and a `type` field. Valid `type` values are: `pipeline-ref`, `agent-ref`, `decision`, and `parallel`.

#### Scenario: Valid workflow loaded
- **WHEN** a `.workflow.yaml` file is present with a valid `name`, `version`, and `steps` list
- **THEN** `WorkflowDef` is populated without error and all steps are accessible by ID

#### Scenario: Duplicate step ID rejected
- **WHEN** a workflow YAML contains two steps with the same `id`
- **THEN** loading returns an error naming the duplicate ID

#### Scenario: Unknown step type rejected
- **WHEN** a step has a `type` value not in the valid set
- **THEN** loading returns an error naming the step ID and unknown type

### Requirement: pipeline-ref step type

A `pipeline-ref` step MUST declare a `pipeline` field naming a pipeline in `~/.config/glitch/pipelines/`. It MAY include a `vars` map of string overrides and an `input` field (supports `{{ctx.<step_id>.output}}` template references). The step output SHALL be written into `WorkflowContext` under the step's `id`.

#### Scenario: Pipeline resolved by name
- **WHEN** a `pipeline-ref` step declares `pipeline: my-pipeline`
- **THEN** the runner resolves `~/.config/glitch/pipelines/my-pipeline.pipeline.yaml` before execution

#### Scenario: Missing pipeline fails fast
- **WHEN** the referenced `.pipeline.yaml` file does not exist
- **THEN** the workflow fails at that step with a clear "pipeline not found" error before execution begins

#### Scenario: Input template expanded from context
- **WHEN** a `pipeline-ref` step declares `input: "{{ctx.prior_step.output}}"`
- **THEN** the runner substitutes the value from `WorkflowContext["prior_step.output"]` before passing to `pipeline.Run()`

### Requirement: agent-ref step type

An `agent-ref` step MUST declare an `agent` field naming an installed APM agent. The runner SHALL resolve the agent's materialized pipeline at `~/.config/glitch/pipelines/apm.<agent>.pipeline.yaml`. It MAY include a `vars` map and an `input` field.

#### Scenario: Agent pipeline resolved
- **WHEN** an `agent-ref` step declares `agent: glab-issue`
- **THEN** the runner resolves `apm.glab-issue.pipeline.yaml` and dispatches via `pipeline.Run()`

#### Scenario: Uninstalled agent fails fast
- **WHEN** the agent's materialized pipeline file does not exist
- **THEN** the workflow fails at that step with "agent pipeline not found: apm.<name>.pipeline.yaml"

### Requirement: decision step type

A `decision` step MUST declare a `model` (Ollama model name), a `prompt` (template string), and an `on` map of branch-name → next-step-id. It MAY declare a `default_branch` used when Ollama returns an error. The prompt MAY reference `{{ctx.*}}` values.

#### Scenario: Branch selected from JSON output
- **WHEN** Ollama returns `{"branch": "bugfix"}` and `on.bugfix` is defined
- **THEN** the conductor sets the next step to the ID mapped by `on.bugfix`

#### Scenario: Unknown branch fails workflow
- **WHEN** Ollama returns a branch name not present in the `on` map and no `default_branch` is set
- **THEN** the workflow fails with "decision node returned unknown branch: <name>"

#### Scenario: Ollama error uses default branch
- **WHEN** Ollama returns an HTTP error or malformed JSON and `default_branch` is set
- **THEN** the conductor uses `default_branch` to continue without failing

### Requirement: parallel step type

A `parallel` step MUST declare a `branches` list, each containing a nested `steps` list. All branches SHALL execute concurrently. The workflow MUST NOT proceed to the next step until all branches complete. If any branch fails, remaining branches are cancelled and the workflow fails.

#### Scenario: All branches succeed
- **WHEN** a `parallel` step has two branches that both complete successfully
- **THEN** both branches' context outputs are merged into `WorkflowContext` and the next sequential step runs

#### Scenario: One branch fails
- **WHEN** one branch in a `parallel` step returns an error
- **THEN** the other branch is cancelled and the workflow fails with the branch error
