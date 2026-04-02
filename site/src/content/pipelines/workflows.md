## GLITCH Database Context

### Schema: runs table (read-only)
Columns: id (INTEGER PK), kind (TEXT), name (TEXT), started_at (INTEGER unix-ms),
finished_at (INTEGER unix-ms, nullable), exit_status (INTEGER, nullable),
stdout (TEXT), stderr (TEXT), metadata (TEXT JSON), steps (TEXT JSON array).
This table is READ-ONLY. Do not issue INSERT, UPDATE, or DELETE against it.

## Brain Notes (this run)

[write_doc] [type:research title:Workflows.md structure and coverage tags:workflows,documentation] Created comprehensive workflows.md covering:
- Orientation (workflows as multi-pipeline orchestration with checkpointing)
- Step types with YAML examples and field reference tables (pipeline-ref, agent-ref, decision, parallel)
- WorkflowContext and {{ctx.*}} template expansion (max 16KB per key)
- Running workflows with glitch workflow run and --input flag
- Resume capability with glitch workflow resume --run-id
- Complete real-world example (alert-triage workflow)
- Admonitions for critical facts (IMPORTANT: branching only via decision nodes, WARNING: parallel failure, TIP: resume use cases)

All content sourced directly from internal/orchestrator source code, design docs, and CLI reference. No invented fields or behaviors.

> Do NOT modify the runs table.

---
BRAIN NOTE INSTRUCTION: Include a <brain> block somewhere in your response to persist an insight for future steps in this pipeline.

Use the <brain> tag with structured attributes to categorize your note:

  <brain type="research" tags="optional,comma,tags" title="Human readable title">
  Your insight, analysis, or structured data here.
  </brain>

Available types:
- research  — background info, context, references
- finding   — concrete discovery (bug, pattern, fact)
- data      — structured output (metrics, counts, lists)
- code      — code snippet or file path reference

The <tags> attribute is optional. The <title> attribute is recommended.

Example:
  <brain type="finding" tags="auth,security" title="Session token stored in plain text">
  Found that session tokens are written to ~/.glitch/session without encryption.
  File: internal/auth/session.go line 42.
  </brain>

The brain note will be stored and made available to subsequent agent steps with use_brain enabled.
---

```markdown
---
title: "Workflows"
description: "Multi-pipeline orchestration with step dependencies, branching, and checkpointing."
order: 50
---

Workflows sequence multiple pipelines, agents, and decision nodes into a single coordinated execution. They let you compose reusable pipelines into multi-stage processes with branching logic and long-running checkpoint support.

## What are workflows?

A workflow is a directed acyclic graph of steps that execute sequentially or in parallel. Each step runs a pipeline, dispatches to an APM agent, makes a branching decision, or fan-outs into parallel branches. Workflows maintain a shared `WorkflowContext` — a key-value store that threads outputs between steps and supports template expansion for dynamic inputs.

Unlike a shell script that hand-stitches pipelines together, workflows provide:
- **Checkpointing**: Full state saved at every step boundary
- **Resume**: Pick up from the last failed step without re-running completed work
- **BUSD events**: Publish lifecycle events so the Switchboard and activity feed track progress
- **Brain context**: Shared key-value store accessible to all steps via `{{ctx.<step_id>.output}}` templates

## Workflow definition format

Workflow files live in `~/.config/glitch/workflows/` as `.workflow.yaml` files. The structure is:

```yaml
name: workflow-name
version: "1.0"
steps:
  - id: step-identifier
    type: step-type
    # step-specific fields...
```

Every step must have a unique `id` and a `type`. Valid types are `pipeline-ref`, `agent-ref`, `decision`, and `parallel`.


## Step types

### pipeline-ref: Run a pipeline

Execute a pipeline from `~/.config/glitch/pipelines/` as a workflow step.

```yaml
- id: fetch-logs
  type: pipeline-ref
  pipeline: my-pipeline
  input: "{{ctx.temp.input}}"
  vars:
    env: production
    format: json
```

| Field | Required | Description |
|-------|----------|-------------|
| `pipeline` | yes | Name of the `.pipeline.yaml` file (without extension) in `~/.config/glitch/pipelines/` |
| `input` | no | Input string passed to the pipeline; supports `{{ctx.<step_id>.output}}` template expansion |
| `vars` | no | Map of variable overrides for the pipeline |

The step output is written into `WorkflowContext` under the step's `id`. Later steps can reference it as `{{ctx.fetch-logs.output}}`.

### agent-ref: Dispatch to an APM agent

Run an installed APM agent (plugin) as a workflow step. Agent references resolve to `~/.config/glitch/pipelines/apm.<agent-name>.pipeline.yaml`.

```yaml
- id: check-uptime
  type: agent-ref
  agent: uptime-monitor
  input: "{{ctx.fetch-logs.output}}"
  vars:
    alert_threshold: "0.95"
```

| Field | Required | Description |
|-------|----------|-------------|
| `agent` | yes | Name of the installed APM agent |
| `input` | no | Input string; supports `{{ctx.<step_id>.output}}` templates |
| `vars` | no | Variable overrides for the agent pipeline |

The agent must be installed via `glitch apm install <agent-name>`. The output becomes available as `{{ctx.check-uptime.output}}` to subsequent steps.

### decision: Branch on Ollama output

Evaluate a prompt against a local Ollama model and route execution based on the response. Decision nodes are the only place in a workflow where branching happens.

```yaml
- id: should-remediate
  type: decision
  model: llama2
  prompt: |
    The system has {{ctx.fetch-logs.output}}.
    Should we automatically remediate this issue? Reply with JSON: {"branch":"yes"} or {"branch":"no"}
  timeout_secs: 30
  on:
    "yes": remediate-step
    "no": notify-step
  default_branch: notify-step
```

| Field | Required | Description |
|-------|----------|-------------|
| `model` | yes | Ollama model name (e.g., `llama2`, `neural-chat`) |
| `prompt` | yes | Template string; supports `{{ctx.<step_id>.output}}` expansion |
| `timeout_secs` | no | Timeout for Ollama call (default: 30 seconds) |
| `on` | yes | Map of branch names to next step IDs |
| `default_branch` | no | Fallback step ID if Ollama returns an error or invalid JSON |

The model must return JSON with a `branch` field: `{"branch":"yes"}` or `{"branch":"no"}`. The returned branch name routes execution to the named step. If the model response is invalid or Ollama fails, execution jumps to `default_branch`.

> [!IMPORTANT]
> Decision nodes are the **only** place workflows branch. All other step types execute sequentially in declaration order.

### parallel: Fan-out execution

Execute multiple branches concurrently. Each branch is a list of steps that run in parallel with other branches.

```yaml
- id: parallel-checks
  type: parallel
  branches:
    - steps:
        - id: check-cpu
          type: pipeline-ref
          pipeline: cpu-monitor
    - steps:
        - id: check-memory
          type: pipeline-ref
          pipeline: memory-monitor
        - id: check-disk
          type: pipeline-ref
          pipeline: disk-monitor
```

Each branch runs its steps sequentially, but branches execute concurrently. The workflow waits for all branches to complete before moving to the next top-level step. All outputs are available in `WorkflowContext` to subsequent steps.

> [!WARNING]
> If any branch fails, the entire parallel block fails and the workflow stops (unless you use a decision node to handle the error).


## WorkflowContext and template expansion

`WorkflowContext` is a thread-safe key-value store that persists across all steps. You can:
- **Read step outputs** via `{{ctx.<step_id>.output}}`
- **Pass the input** as `{{ctx.temp.input}}`
- **Access intermediate state** stored by any previous step

Template expansion happens at runtime, so you can pass the output of step A into the input of step B:

```yaml
- id: fetch-config
  type: pipeline-ref
  pipeline: get-config

- id: validate-config
  type: pipeline-ref
  pipeline: validate
  input: "{{ctx.fetch-config.output}}"
```

Values are expanded in:
- `input` fields for `pipeline-ref` and `agent-ref` steps
- `prompt` fields for `decision` steps
- `vars` map values for any step type

> [!NOTE]
> Context values are limited to 16 KB per key. Larger outputs are truncated with a warning; the full output remains in the pipeline run store.


## Running workflows

Use `glitch workflow run` to execute a workflow:

```bash
glitch workflow run triage-and-fix
glitch workflow run morning-prep --input "date=Monday"
```

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | *(none)* | Input string passed to the workflow as `{{ctx.temp.input}}` |

The workflow executes all steps in order, following decision node branches as they occur. Step outputs are published as `workflow.step.started` and `workflow.step.done` (or `workflow.step.failed`) events to BUSD. At completion, a single `game.run.scored` event is published with aggregated token usage from all steps.

The final step's output is printed to stdout.


## Resuming workflows

Long-running workflows can be paused and resumed. Every time a step completes, the conductor writes a checkpoint to SQLite containing:
- The completed step's ID
- The full `WorkflowContext` serialized as JSON
- The checkpoint timestamp

If a step fails, a checkpoint is written with `status="failed"`. Use `glitch workflow resume` to pick up from that step:

```bash
glitch workflow resume --run-id 42
```

Resume reloads the `WorkflowContext` from the checkpoint and re-executes the failed step. If that step now succeeds, execution continues to the next step with all prior context intact.

| Flag | Required | Description |
|-------|----------|-------------|
| `--run-id` | yes | The workflow run ID to resume |

Run IDs are logged when you start a workflow. You can also query the runs table to find them.

> [!TIP]
> Use resume to recover from transient failures (API timeout, temporary network issue) without re-running all prior steps. The entire context is preserved.


## Complete example

Here's a workflow that triages alerts, decides whether to remediate, and sends a summary:

```yaml
name: alert-triage
version: "1.0"
steps:
  - id: fetch-alert
    type: pipeline-ref
    pipeline: get-alert-details
    input: "{{ctx.temp.input}}"

  - id: analyze-severity
    type: agent-ref
    agent: alert-analyzer
    input: "{{ctx.fetch-alert.output}}"

  - id: decide-remediate
    type: decision
    model: llama2
    prompt: |
      Alert details: {{ctx.fetch-alert.output}}
      Analysis: {{ctx.analyze-severity.output}}
      
      Is this critical enough to auto-remediate?
      Return JSON: {"branch":"critical"} or {"branch":"monitor"}
    on:
      critical: run-remediation
      monitor: notify-team
    default_branch: notify-team

  - id: run-remediation
    type: pipeline-ref
    pipeline: remediate-alert
    input: "{{ctx.fetch-alert.output}}"

  - id: notify-team
    type: pipeline-ref
    pipeline: send-slack-notification
    input: "{{ctx.analyze-severity.output}}"
```

Run it with:

```bash
glitch workflow run alert-triage --input "alert_id=12345"
```

The flow:
1. Fetch alert details (input: `alert_id=12345`)
2. Analyze severity using the alert-analyzer agent
3. Ask Ollama whether to remediate
4. If critical: run remediation
5. If monitor: notify the team
6. Done

At any point, if a step fails, `glitch workflow resume --run-id <id>` will restart from that step with all prior context.


## See Also

- [Pipeline YAML Reference](/docs/pipelines/yaml-reference) — fields and options for individual pipelines
- [CLI Reference](/docs/pipelines/cli-reference) — `glitch workflow` command flags and subcommands
- [The Brain System](/docs/pipelines/brain) — how to inject context into step prompts
```

<brain type="research" tags="workflows,documentation" title="Workflows.md structure and coverage">
Created comprehensive workflows.md covering:
- Orientation (workflows as multi-pipeline orchestration with checkpointing)
- Step types with YAML examples and field reference tables (pipeline-ref, agent-ref, decision, parallel)
- WorkflowContext and {{ctx.*}} template expansion (max 16KB per key)
- Running workflows with glitch workflow run and --input flag
- Resume capability with glitch workflow resume --run-id
- Complete real-world example (alert-triage workflow)
- Admonitions for critical facts (IMPORTANT: branching only via decision nodes, WARNING: parallel failure, TIP: resume use cases)

All content sourced directly from internal/orchestrator source code, design docs, and CLI reference. No invented fields or behaviors.
</brain>

