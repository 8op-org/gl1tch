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
title: "Pipelines"
description: "Pipeline execution model, DAG step ordering, context passing, and the step lifecycle."
order: 2
---
```

A pipeline in gl1tch is a YAML-defined workflow where steps are executed in order determined by their dependencies. Steps run sequentially or in parallel depending on the `needs` declarations, forming a directed acyclic graph (DAG). Data flows between steps through a shared execution context that holds variables, outputs, and structured state.


## Pipeline Definition

A pipeline is a YAML file at `~/.config/glitch/pipelines/<name>.pipeline.yaml` with a simple structure:

```yaml
name: my-pipeline
version: "1"
vars:
  timeout: "30s"
  model: "llama3.2"

steps:
  - id: fetch-data
    executor: builtin.http_get
    url: "https://api.example.com/data"
    
  - id: process
    executor: ollama
    model: "{{ param.model }}"
    prompt: "Summarize: {{ step.fetch-data.data.body }}"
    needs: [fetch-data]
    
  - id: review
    executor: claude
    model: "claude-haiku"
    prompt: "Review this summary: {{ step.process.data.output }}"
    needs: [process]
    retry:
      max_attempts: 3
      interval: "2s"
      on: "on_failure"
```

Each step declares an `executor` (the plugin or builtin that runs it), optional configuration fields (like `model`, `prompt`, `url`), and dependencies via `needs`.


## Execution Model

The pipeline runner builds a DAG from all steps and their `needs` declarations, then executes steps in topologically sorted order. Steps with no dependencies run immediately; dependent steps wait for their prerequisites to complete.

Steps execute within a shared **execution context** — a key-value store that holds pipeline variables, step outputs, and temporary state. Variables are scoped by prefix: `param.*` for pipeline-level vars, `step.<id>.data.*` for step outputs, and `cwd` for the process working directory.

Before any step runs, the runner performs setup:
- Creates a fresh `ExecutionContext`
- Loads pipeline-level variables (`vars` block) with `param.` prefix
- Sets `param.input` to any user-supplied input
- Injects brain context if `use_brain` is enabled
- Initializes step state trackers


## DAG Step Ordering

Steps execute when all dependencies listed in `needs` have completed. The `needs` field is an optional list of step IDs:

```yaml
steps:
  - id: step-a
  - id: step-b
    needs: [step-a]          # waits for step-a
  - id: step-c
    needs: [step-a, step-b]  # waits for both step-a and step-b
  - id: step-d
    # no needs — runs immediately alongside step-a
```

Steps without dependencies form parallel streams. In the example above, step-a and step-d run concurrently. step-b waits only for step-a. step-c waits for both. The runner uses goroutines to execute independent steps in parallel, avoiding unnecessary blocking.

> [!TIP]
> Design pipelines to maximize parallelism. If step-c only depends on output from step-b, don't list step-a in its `needs` — this allows step-a and step-d to run alongside other steps.


## Context and Variable Passing

Variables flow between steps through template expansion. When a step's configuration is prepared, any string field containing `{{ ... }}` is evaluated against the current execution context.

Template syntax uses dot notation to access context values:

- `{{ param.model }}` — a pipeline variable
- `{{ step.process.data.output }}` — structured data from a prior step
- `{{ cwd }}` — the current working directory
- `{{ step.fetch.data.status_code }}` — nested fields in step output

When a step completes, its executor may produce structured output that becomes available as `step.<id>.data.<key>` for all downstream steps. For example, an HTTP step might output:

```yaml
steps:
  - id: call-api
    executor: builtin.http_get
    url: "https://api.example.com/users"
    # produces: step.call-api.data.status_code, step.call-api.data.body
    
  - id: process-response
    prompt: "Parse this JSON: {{ step.call-api.data.body }}"
    needs: [call-api]
```

The pipeline also sets `step.<id>.out` (legacy) for backward compatibility, but structured `step.<id>.data.*` is preferred for new work.

> [!WARNING]
> Variables are expanded at step-launch time, not at definition time. If you use `{{ step.x.data.y }}` in a step that runs before step-x completes, the template expands to empty. Always list dependencies in `needs` to ensure ordering.


## Step Lifecycle

Each step progresses through a fixed lifecycle:

1. **Waiting** — the step is defined but its dependencies haven't completed
2. **Ready** — all dependencies are done; the step is queued for execution
3. **Running** — the executor is active (plugin binary executing, model generating, etc.)
4. **Done** — the step completed without error
5. **Failed** — the step returned an error (after exhausting retries)

The runner prints a structured log line at each transition:

```
[step:fetch-data] status:running
[step:fetch-data] status:done
[step:process] status:running
[step:process] status:failed
```

Input and output steps (if present) do not emit status lines — they are handled synchronously before the main DAG executes.

### Initialization and Cleanup

Although not yet exposed in the YAML, steps support `init` and `cleanup` phases for resource allocation and release:

- **init**: Called before the executor runs; establishes connections, allocates memory, or prepares state
- **execute**: The main step logic; returns output and optional error
- **cleanup**: Called after execute (whether succeeded or failed); closes connections, deallocates resources

Future versions may expose these hooks in pipeline YAML for plugin authors to define custom lifecycle behavior.


## Control Flow: Retry, Error Branching, and Loops

### Retry Policy

Steps can declare a `retry` block to automatically re-run on failure:

```yaml
steps:
  - id: flaky-call
    executor: builtin.http_get
    url: "https://unreliable-api.example.com/data"
    retry:
      max_attempts: 3
      interval: "2s"
      on: "on_failure"
```

The runner will re-execute the step up to 3 times, waiting 2 seconds between attempts, only if the previous attempt failed. If all attempts fail, the step enters the Failed state.

### Error Branching

By default, a failed step terminates the pipeline. Use `on_failure` to route to a recovery step:

```yaml
steps:
  - id: fetch
    executor: builtin.http_get
    url: "https://api.example.com"
    on_failure: use-cache
    
  - id: use-cache
    executor: builtin.log
    message: "API failed; using cached data instead"
    # this step runs only if 'fetch' fails
```

The `use-cache` step is held back during normal execution and only enqueued if `fetch` fails. If `use-cache` completes (regardless of its own exit status), the pipeline continues to any steps that depend on it.

### Loops

Use `for_each` to iterate a step over a list of inputs, collecting outputs into an array:

```yaml
steps:
  - id: items
    executor: builtin.http_get
    url: "https://api.example.com/items"
    
  - id: analyze-each
    executor: ollama
    model: "llama3.2"
    for_each: "{{ step.items.data.results }}"
    prompt: "Analyze: {{ item }}"
    # produces: step.analyze-each.data.results (array of outputs)
    needs: [items]
```

The `for_each` field references a list in the context. The step is executed once per item, and `{{ item }}` expands to the current element. All outputs are collected into `step.<id>.data.results`.


## Event Publishing

Steps can publish their output to the event bus (busd) for downstream consumers:

```yaml
steps:
  - id: notification
    executor: builtin.log
    message: "Pipeline checkpoint reached"
    publish_to: "app.notifications"
```

After the step completes, the runner emits a bus event on topic `app.notifications` with the step's output data. This allows the console, cron scheduler, or other services to react to pipeline progress in real time.


## Builtin Steps

Several steps are compiled directly into the binary and do not require an external plugin:

| Step | Executor | Purpose |
|------|----------|---------|
| Assert | `builtin.assert` | Validate a condition; fail if false |
| Set Data | `builtin.set_data` | Add a key-value pair to context |
| Log | `builtin.log` | Print a message to stdout |
| Sleep | `builtin.sleep` | Pause execution for a duration |
| HTTP GET | `builtin.http_get` | Fetch a URL and store response |

These steps have no external dependencies and are always available.


## Working Example

Here is a complete, runnable pipeline that chains local and cloud models:

```yaml
name: code-review-pipeline
version: "1"
vars:
  local_model: "mistral"
  claude_model: "claude-haiku"

steps:
  - id: input
    executor: builtin.set_data
    data:
      code: |
        function add(a, b) {
        return a + b
        }
    
  - id: local-review
    executor: ollama
    model: "{{ param.local_model }}"
    prompt: |
      Review this code and list issues:
      {{ step.input.data.code }}
    needs: [input]
    
  - id: cloud-refine
    executor: claude
    model: "{{ param.claude_model }}"
    prompt: |
      The local model found these issues:
      {{ step.local-review.data.output }}
      
      Rewrite the code to fix them.
    needs: [local-review]
    
  - id: report
    executor: builtin.log
    message: |
      Final code:
      {{ step.cloud-refine.data.output }}
    needs: [cloud-refine]
```

Running this pipeline:

```bash
glitch pipeline run code-review-pipeline
```

The runner will execute `input`, then both `local-review` (waiting for `input`), then `cloud-refine` (waiting for `local-review`), then `report` (waiting for `cloud-refine`). At each step transition, a status line is printed.


## See Also

- [Brain Context](/docs/brain) — how pipelines read and write brain notes
- [Providers](/docs/providers) — available executors (ollama, claude, plugins)
- [Cron Scheduling](/docs/cron) — run pipelines on a schedule

