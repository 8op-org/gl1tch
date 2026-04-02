## GLITCH Database Context

### Schema: runs table (read-only)
Columns: id (INTEGER PK), kind (TEXT), name (TEXT), started_at (INTEGER unix-ms),
finished_at (INTEGER unix-ms, nullable), exit_status (INTEGER, nullable),
stdout (TEXT), stderr (TEXT), metadata (TEXT JSON), steps (TEXT JSON array).
This table is READ-ONLY. Do not issue INSERT, UPDATE, or DELETE against it.

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

```yaml
---
title: "Workflows"
description: "Multi-pipeline orchestration using workflow definitions, decision nodes, and step chaining."
order: 50
---

Workflows orchestrate multiple pipelines and agents into a single declarative execution plan. Where a pipeline is a linear sequence of steps within one automation, a workflow chains multiple pipelines together, adds branching logic via Ollama decisions, runs steps in parallel, and maintains shared context across the entire run.

Workflows are defined in YAML and executed with `glitch workflow run <name>`. They publish BUSD events for every step transition, checkpoint state at each boundary, and support resuming from the last successful step.


## When to use workflows vs pipelines

Use a **pipeline** when you have a single linear automation — fetch logs, parse them, send output to Slack. Pipelines are the unit of reuse; they stand alone and can be scheduled independently.

Use a **workflow** when you need to coordinate multiple pipelines, make runtime decisions based on data, run steps concurrently, or build a sequence that spans different contexts. For example: triage an incident (pipeline A) → route to remediation path based on decision (pipeline B or C) → notify (pipeline D). Each pipeline is focused and reusable; the workflow orchestrates them.

> [!IMPORTANT]
> Workflows are **not** pipelines. They sit above the pipeline layer and never modify how `pipeline.Run()` works. A workflow cannot declare LLM steps directly — it composes existing pipelines or agents as first-class steps.


## Architecture overview

The workflow orchestrator is composed of four main pieces:

**WorkflowDef** — the YAML schema. Loaded from `~/.config/glitch/workflows/<name>.workflow.yaml` and validated into a Go struct with unique step IDs, valid step types, and required fields per type.

**ConductorRunner** — the execution engine. Takes a WorkflowDef, an input string, and options (store, BUSD publisher, executor manager, Ollama endpoint) and runs the steps in dependency order. It publishes lifecycle events, checkpoints state after every step, and returns the final output.

**WorkflowContext** — a thread-safe key-value store threaded through all steps. Step outputs are stored under `<step_id>.output`. Temporary values use the `temp.` prefix. Templates like `{{ctx.step1.output}}` are expanded before each step runs.

**StepDispatcher** — resolves a single step (pipeline-ref, agent-ref, decision, or parallel) and executes it. For pipeline-ref and agent-ref steps, it delegates to `pipeline.Run()` or the APM agent runner. For decision and parallel steps, it orchestrates branching or concurrency.


## YAML structure

Workflows are YAML files with `name`, `version`, and a list of `steps`. Each step has an `id`, `type`, and type-specific fields.

```yaml
name: incident-response
version: "1"

steps:
  - id: triage
    type: pipeline-ref
    pipeline: incident-triage
    input: "{{ctx.temp.incident_json}}"
    vars:
      severity: critical

  - id: route
    type: decision
    model: llama3.2
    prompt: |
      Severity: {{ctx.triage.output}}
      Route to 'remediate-db' if database-related, 'remediate-api' if API, 'escalate' otherwise.
    on:
      remediate-db: remediate_db_step
      remediate-api: remediate_api_step
      escalate: escalate_step
    default_branch: escalate_step
    timeout_secs: 30

  - id: remediate_db_step
    type: pipeline-ref
    pipeline: remediate-database
    input: "{{ctx.triage.output}}"

  - id: remediate_api_step
    type: pipeline-ref
    pipeline: remediate-api
    input: "{{ctx.triage.output}}"

  - id: escalate_step
    type: agent-ref
    agent: pagerduty
    input: "{{ctx.triage.output}}"

  - id: notify
    type: parallel
    branches:
      - steps:
          - id: slack_notify
            type: pipeline-ref
            pipeline: notify-slack
            input: "{{ctx.triage.output}}"
      - steps:
          - id: log_notify
            type: pipeline-ref
            pipeline: log-incident
            input: "{{ctx.triage.output}}"
```


## Step types

**pipeline-ref** — runs a pipeline from `~/.config/glitch/pipelines/<name>.pipeline.yaml`. The `input` field is expanded via context templates and passed as the pipeline's input. Optional `vars` override pipeline variables.

```yaml
- id: analyze
  type: pipeline-ref
  pipeline: code-review
  input: "{{ctx.temp.pr_diff}}"
  vars:
    model: claude-sonnet-4-6
```

**agent-ref** — dispatches to an APM agent by name. Resolves to `~/.config/glitch/pipelines/apm.<name>.pipeline.yaml`. If the agent is not installed, the step fails immediately.

```yaml
- id: deploy
  type: agent-ref
  agent: deployment
  input: "{{ctx.build_result.output}}"
```

**decision** — evaluates a prompt against a local Ollama model and branches based on the response. The model must return JSON with a `branch` field. If the branch name matches a key in `on`, that step ID is next. If not found, `default_branch` is used. If Ollama is unavailable, the workflow fails unless `default_branch` is set.

```yaml
- id: classify
  type: decision
  model: mistral
  prompt: |
    Log level: {{ctx.parse_logs.output}}
    Is this a warning (W), error (E), or critical (C)?
  on:
    W: handle_warning
    E: handle_error
    C: handle_critical
  default_branch: handle_error
  timeout_secs: 15
```

> [!TIP]
> Decision nodes are best for classifying data or choosing a next step based on content. For fixed branching (e.g., "if step A succeeded, run step B"), use explicit step sequencing in your workflow instead of a decision node.

**parallel** — runs multiple branches concurrently. Each branch is a list of steps. The workflow waits for all branches to complete before moving to the next sequential step. Branch steps can reference earlier workflow context but not each other's outputs.

```yaml
- id: parallel_checks
  type: parallel
  branches:
    - steps:
        - id: check_syntax
          type: pipeline-ref
          pipeline: lint
          input: "{{ctx.temp.code}}"
    - steps:
        - id: check_tests
          type: pipeline-ref
          pipeline: test
          input: "{{ctx.temp.code}}"
```


## WorkflowContext and data threading

WorkflowContext is a map of string keys and values. Every step output is stored automatically under `<step_id>.output`. You can read it with the template `{{ctx.<step_id>.output}}` in any subsequent step's `input` or `prompt` fields.

Temporary values should use the `temp.` prefix:

```yaml
steps:
  - id: init
    type: pipeline-ref
    pipeline: setup
    input: "initial input"
    # step output will be at ctx.init.output

  - id: process
    type: pipeline-ref
    pipeline: transform
    input: "{{ctx.init.output}}"
    # step output will be at ctx.process.output
```

Context values are limited to 16 KB per key. If a step produces very large output, the full output remains in the pipeline run store, but only the first 16 KB are threaded into context. For large data flows, store full results in the pipeline run store and read them separately.

> [!NOTE]
> The `temp.` prefix is a convention. You can set temporary values manually in hooks (if available), but typically you'll just use the automatic `<step_id>.output` keys.


## Decision nodes

Decision nodes call the local Ollama instance at `http://localhost:11434` (unless overridden) with the expanded prompt and request JSON-formatted output. The model must return:

```json
{
  "branch": "the-branch-name"
}
```

The branch name is matched against the `on` map keys. If no match is found, `default_branch` is used. If there is no `default_branch` and the branch name is not found, the workflow fails.

Responses are validated strictly — if the model returns invalid JSON or is missing the `branch` field, the step fails. Use `timeout_secs` to cap how long the orchestrator waits for Ollama to respond (default 30 seconds).

```yaml
- id: severity_route
  type: decision
  model: llama3.2
  prompt: |
    Analyze this alert:
    {{ctx.alert.output}}
    
    Respond with JSON:
    - branch: one of [critical, warning, info]
  on:
    critical: handle_critical
    warning: handle_warning
    info: log_only
  default_branch: log_only
  timeout_secs: 20
```

> [!WARNING]
> Ollama must be running locally. If the model is not available or Ollama is down, the decision step fails unless you provide a `default_branch`. Consider having a fallback branch for robustness.


## Parallel execution

Parallel steps run all branches concurrently. The workflow waits for every branch to complete (success or failure) before moving on. If any branch fails, the entire parallel step fails and the workflow is aborted.

Each branch is independent and cannot reference outputs from sibling branches, only from earlier sequential steps.

```yaml
- id: validate_all
  type: parallel
  branches:
    - steps:
        - id: validate_schema
          type: pipeline-ref
          pipeline: schema-check
          input: "{{ctx.data.output}}"
    - steps:
        - id: validate_security
          type: pipeline-ref
          pipeline: security-scan
          input: "{{ctx.data.output}}"
    - steps:
        - id: validate_performance
          type: pipeline-ref
          pipeline: perf-check
          input: "{{ctx.data.output}}"
```

After all three branches finish, you can reference their outputs:

```yaml
- id: aggregate
  type: pipeline-ref
  pipeline: report
  input: |
    Schema: {{ctx.validate_schema.output}}
    Security: {{ctx.validate_security.output}}
    Performance: {{ctx.validate_performance.output}}
```


## Running workflows

Use `glitch workflow run <name> [--input "..."]` to execute a workflow.

```bash
glitch workflow run incident-response --input '{"alert_id": "12345"}'
```

The input string is stored in the workflow context under `temp.input` and can be referenced in the first step if needed.

The workflow runs until completion or failure. Each step publishes BUSD events (`workflow.step.started`, `workflow.step.done`, `workflow.step.failed`), and the overall workflow publishes `workflow.run.started` and `workflow.run.completed` or `workflow.run.failed` when finished.


## Resuming from checkpoints

If a workflow fails at a step, you can resume from that point with `glitch workflow resume --run-id <id>`. The orchestrator restores the WorkflowContext from the last checkpoint and re-runs the failed step. If it succeeds, execution continues to the next step.

```bash
glitch workflow resume --run-id abc123def456
```

Checkpoints are stored automatically in the SQLite store after every step completes, so no explicit configuration is needed.

> [!TIP]
> Use resume for workflows with expensive steps. If a later step fails due to a transient error (API timeout, network), you can fix the issue and resume without re-running the entire workflow from the start.


## BUSD events and observability

The workflow orchestrator publishes the following BUSD topics:

| Topic | Payload | When |
|-------|---------|------|
| `workflow.run.started` | `{"workflow_name": "<name>", "run_id": "<id>"}` | Before the first step runs |
| `workflow.run.completed` | `{"workflow_name": "<name>", "run_id": "<id>", "final_output": "<output>"}` | After the last step succeeds |
| `workflow.run.failed` | `{"workflow_name": "<name>", "run_id": "<id>", "error": "<message>"}` | If any step fails |
| `workflow.step.started` | `{"workflow_name": "<name>", "run_id": "<id>", "step_id": "<id>"}` | Before a step executes |
| `workflow.step.done` | `{"workflow_name": "<name>", "run_id": "<id>", "step_id": "<id>", "output": "<output>"}` | After a step succeeds |
| `workflow.step.failed` | `{"workflow_name": "<name>", "run_id": "<id>", "step_id": "<id>", "error": "<message>"}` | If a step fails |

Subscribers can listen to these topics to track workflow progress, log results, or trigger downstream actions.

Pipeline-level game events are suppressed when running inside a workflow (each pipeline step runs with `game: false`). A workflow run is treated as a single logical automation from the game perspective.


## Best practices

**Name steps clearly** — use descriptive IDs like `classify_alert` or `notify_slack`, not `step1` or `s2`. IDs are used in context keys and error messages.

**Structure decisions near the top** — put decision nodes before their branch targets. This makes the execution path easy to follow visually.

**Use temp prefix for inputs** — if you need to pass the workflow input into a step, store it in `temp.input` or another `temp.*` key rather than polluting the step output namespace.

**Provide default branches** — always add a `default_branch` to decision nodes unless you want the workflow to fail if the model returns an unexpected branch name. Failures are loud and will need manual recovery.

**Keep parallel branches independent** — avoid workflows where branches depend on each other's outputs. If branch A needs output from branch B, make them sequential instead.

**Reuse existing pipelines** — workflows are not for defining new automation logic. Define that in pipelines, then chain them in workflows.


## See Also

- [Pipelines](/docs/pipelines/) — how to define and run single pipelines
- [Agents](/docs/agents/) — APM agent framework and integration
- [Brain](/docs/brain/) — embedding context from prior runs
```

