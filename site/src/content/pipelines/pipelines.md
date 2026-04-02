## GLITCH Database Context

### Schema: runs table (read-only)
Columns: id (INTEGER PK), kind (TEXT), name (TEXT), started_at (INTEGER unix-ms),
finished_at (INTEGER unix-ms, nullable), exit_status (INTEGER, nullable),
stdout (TEXT), stderr (TEXT), metadata (TEXT JSON), steps (TEXT JSON array).
This table is READ-ONLY. Do not issue INSERT, UPDATE, or DELETE against it.

## Brain Notes (this run)

[write_doc] tags.
```

When `use_brain: true`, the runner prepends a context block to the prompt describing past reviews, patterns, and learned insights stored in the `brain_notes` table. When `write_brain: true`, the runner parses any `<brain>...

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
title: "Pipelines"
description: "DAG-based multi-step workflows with structured context, retry policies, and error handling."
order: 1
---

Pipelines are YAML-defined workflows that chain together multiple steps, each executed by a provider (local Ollama, Claude, or custom plugins). Steps run in a directed acyclic graph (DAG) determined by their dependencies, with structured context flowing between them. Each step has a clear lifecycle, supports conditional execution and retries, and can publish its output to external systems.


## Pipeline YAML Structure

Every pipeline is a `.pipeline.yaml` file defining a name, version, and ordered list of steps.

```yaml
name: example-pipeline
version: "1"

vars:
  model: llama3.2
  temperature: 0.7

steps:
  - id: analyze
    executor: ollama
    model: "{{ vars.model }}"
    prompt: |
      Analyze this code for security issues:
      {{ param.input }}
    
  - id: summarize
    executor: claude
    model: claude-sonnet-4-6
    needs: [analyze]
    prompt: |
      Write a 1-sentence summary of these findings:
      {{ step.analyze.data.summary }}
```

The `vars` block sets pipeline-wide defaults. The `steps` array defines the execution sequence. Each step has an `id`, `executor` (provider name), and `prompt` (task description). Template variables like `{{ param.input }}` and `{{ step.analyze.data.summary }}` are substituted at runtime.


## DAG Execution Model

Pipelines execute as a directed acyclic graph, not linearly. When a step declares `needs: [stepA, stepB]`, it runs only after both dependencies complete. Steps with no `needs` clause run immediately in parallel.

```yaml
steps:
  - id: fetch_data
    executor: builtin
    action: http_get
    url: https://api.example.com/data
    
  - id: process_data
    executor: ollama
    model: mistral
    needs: [fetch_data]
    prompt: |
      Clean and summarize this data:
      {{ step.fetch_data.data.body }}
      
  - id: validate_result
    executor: builtin
    action: assert
    needs: [process_data]
    condition: "{{ step.process_data.data.tokens }} > 0"
```

This pipeline runs `fetch_data` first. Once it completes, `process_data` runs (blocking on `fetch_data`). In parallel, if another step had no dependencies, it would run concurrently. After `process_data` finishes, `validate_result` runs.

The runner uses goroutines to parallelize independent steps, respecting `needs` constraints. A failure in any step stops all steps that depend on it (unless `on_failure` routing is configured).


## Step Lifecycle

Every step passes through three phases: `init`, `execute`, and `cleanup`. This allows providers to allocate resources, do work, and release them cleanly.

**init** — The provider prepares: connects to Ollama, authenticates with Claude, loads data from disk, etc. If `init` fails, the step is marked failed and dependent steps do not run.

**execute** — The provider runs the actual task. For LLM steps, this is the model inference. For plugins, this is the binary execution. The output is captured and stored in the context.

**cleanup** — The provider releases resources: closes connections, deletes temporary files, etc. Cleanup runs even if `execute` failed, unless the provider is completely unavailable.

Structured output from each step is stored in a typed data map, not a flat string. This allows downstream steps to extract specific fields.

```yaml
steps:
  - id: list_files
    executor: builtin
    action: log
    message: "Files in {{ param.cwd }}"
    # After this step, step.list_files.data will contain the output
    
  - id: next_step
    needs: [list_files]
    prompt: |
      Process these items:
      {{ step.list_files.data.output }}
```


## Context and Variables

During execution, the pipeline builds a context object accessible in all prompts and conditions. Three namespaces are available:

**{{ param.* }}** — User-supplied input. For example, `glitch pipeline run my-pipeline --input "arg=value"` sets `{{ param.input }}` to the string value. Additional `--input` flags are parsed as key=value pairs and stored in `{{ param.key }}`.

**{{ vars.* }}** — Pipeline-level defaults defined in the `vars` block or overridden at runtime.

**{{ step.<id>.data.* }}** — Output from completed steps. Each step stores its output in a typed map; accessing a missing key is an error.

```yaml
vars:
  model: llama3.2

steps:
  - id: gather
    executor: builtin
    action: http_get
    url: "{{ param.api_url }}"
    
  - id: process
    needs: [gather]
    executor: ollama
    model: "{{ vars.model }}"
    prompt: |
      Process this response:
      {{ step.gather.data.body }}
```

Template substitution happens at step dispatch time, after all dependencies are satisfied. Missing variables or undefined step outputs cause the step to fail immediately.


## Retry Policies

Steps can be retried on failure. The `retry` block controls behavior.

```yaml
steps:
  - id: call_api
    executor: builtin
    action: http_get
    url: https://api.example.com/data
    retry:
      max_attempts: 3
      interval: 5s
      on: on_failure  # always | on_failure (default: on_failure)
```

If the step fails, the runner waits `interval` seconds, then re-executes it. After `max_attempts` total attempts (including the first), if still failing, the step is marked failed.

The `on` condition controls when retries trigger. `on_failure` (default) retries only if the step returned an error. `always` retries unconditionally (useful for polls).

```yaml
retry:
  max_attempts: 5
  interval: 2s
  on: always
```

Retries are scoped to a single step; they do not cascade to dependent steps or re-run the entire pipeline.


## Error Handling and Branching

By default, a step failure stops the pipeline. The `on_failure` field routes execution to a recovery step instead.

```yaml
steps:
  - id: primary_model
    executor: ollama
    model: mistral
    prompt: "Analyze this input: {{ param.input }}"
    on_failure: fallback_model
    
  - id: fallback_model
    executor: claude
    prompt: |
      The local analysis failed. Try again:
      {{ param.input }}
```

If `primary_model` fails, instead of aborting, the pipeline jumps to `fallback_model`. The original error is available in context as `{{ step.primary_model.error }}` (if the fallback step is configured to use it).

`on_failure` is mutually exclusive with `needs`; a recovery step cannot have dependencies. If a recovery step also fails, the pipeline is marked failed and stops.


## Conditional Execution

Steps can be skipped based on a condition. Use the `condition` field with a template expression.

```yaml
steps:
  - id: check_env
    executor: builtin
    action: set_data
    key: is_prod
    value: "{{ param.env == 'prod' }}"
    
  - id: deploy_prod
    condition: "{{ step.check_env.data.is_prod }}"
    needs: [check_env]
    executor: builtin
    action: log
    message: "Deploying to production"
```

If the condition evaluates to false, the step is skipped. Dependent steps then fail (they cannot run if the step they need is skipped), unless the skipped step is optional.

Conditions are evaluated after dependencies complete, but before `execute` is called.


## Built-in Steps

The runner includes a small set of compiled-in executors for common tasks. These do not require plugins.

### builtin.log

Log a message to the pipeline output.

```yaml
- id: announce
  executor: builtin
  action: log
  message: "Processing: {{ param.input }}"
```

### builtin.set_data

Set a variable in the context for use by downstream steps.

```yaml
- id: extract_key
  executor: builtin
  action: set_data
  key: extracted_value
  value: "{{ step.prior.data.field }}"
```

Output is stored in `step.<id>.data.key` (the value you set).

### builtin.assert

Fail the step if a condition is false.

```yaml
- id: validate
  executor: builtin
  action: assert
  condition: "{{ step.prior.data.count > 0 }}"
  message: "Validation failed: count is zero"
```

### builtin.http_get

Fetch a URL and store the response body and status code.

```yaml
- id: fetch
  executor: builtin
  action: http_get
  url: "https://api.example.com/v1/data"
  headers:
    Authorization: "Bearer {{ param.api_key }}"
```

Output: `step.<id>.data.body` (response text), `step.<id>.data.status_code` (HTTP status).

### builtin.sleep

Pause for a duration.

```yaml
- id: wait
  executor: builtin
  action: sleep
  duration: 10s
```


## Plugin Executors

Custom executors are installed as plugins with hierarchical names. Plugins are resolved from the plugin registry using the executor ID.

```yaml
steps:
  - id: run_code
    executor: providers.claude.chat
    model: claude-sonnet-4-6
    prompt: "Write a Python function to sort a list"
    
  - id: run_local
    executor: providers.ollama.chat
    model: mistral
    prompt: "Summarize this text: {{ param.text }}"
```

Plugin names follow the pattern `<category>.<action>`. For example, `providers.claude.chat` is a plugin in the `providers` category with action `chat`. Each plugin binary receives the step's prompt and context via environment variables and stdin.

Plugins are responsible for their own lifecycle: they may allocate connections, temp files, or other resources in `init`, use them in `execute`, and clean up in `cleanup`.


## For-Each Loops

Steps can iterate over a list of inputs, executing once per item and collecting outputs into an array.

```yaml
steps:
  - id: process_items
    executor: ollama
    model: llama3.2
    for_each: "{{ param.items }}"
    prompt: |
      Analyze this item:
      {{ each.value }}
```

The `for_each` expression is a template that evaluates to a list. For each item, the step runs with `{{ each.key }}` (index) and `{{ each.value }}` (the item). The output is collected into `step.<id>.data.results` as an array.

```yaml
vars:
  codes:
    - "function add(a, b) { return a + b; }"
    - "function greet(name) { return 'Hi ' + name; }"

steps:
  - id: review_each
    executor: claude
    model: claude-sonnet-4-6
    for_each: "{{ vars.codes }}"
    prompt: |
      Review this code snippet:
      {{ each.value }}
      
  - id: summarize
    needs: [review_each]
    prompt: |
      All reviews:
      {{ step.review_each.data.results }}
```

If `for_each` has 100 items, the step is not executed 100 times in the current design—iteration is intended for moderate-size lists (< 50 items). Very large loops should be refactored into separate pipelines.


## Event Publishing

Step output can be published to external systems. Use the `publish_to` field to send events to the event bus.

```yaml
steps:
  - id: analyze
    executor: ollama
    model: mistral
    prompt: "Analyze: {{ param.input }}"
    publish_to: "analysis.results"
```

When the step completes, its output is published as a BUSD event on the topic `analysis.results`. Other parts of the system can subscribe and react in real-time (e.g., a TUI panel updates, a cron job triggers, etc.).

The event payload includes step metadata (id, duration, status) and the full context at that point.


## Brain Integration

Pipelines can automatically inject historical knowledge and write back learned insights. Use the `use_brain` and `write_brain` flags on steps or at the pipeline level.

```yaml
name: code-review-with-memory
version: "1"
use_brain: true
write_brain: true

steps:
  - id: review
    executor: claude
    model: claude-sonnet-4-6
    prompt: |
      Review this code:
      {{ param.code }}
      
      If you identify patterns or lessons, write them in <brain> tags.
```

When `use_brain: true`, the runner prepends a context block to the prompt describing past reviews, patterns, and learned insights stored in the `brain_notes` table. When `write_brain: true`, the runner parses any `<brain>...</brain>` blocks in the LLM response and stores them for future reference.

Brain notes are scoped to the current working directory, so different projects maintain separate memories.


## Example: Multi-Step Analysis Pipeline

```yaml
name: github-digest
version: "1"

vars:
  repo: my-org/my-repo
  model: mistral

steps:
  - id: fetch_issues
    executor: builtin
    action: http_get
    url: "https://api.github.com/repos/{{ vars.repo }}/issues"
    headers:
      Authorization: "Bearer {{ param.gh_token }}"
      
  - id: normalize
    executor: providers.jq.filter
    needs: [fetch_issues]
    filter: ".[] | {number, title, state, updated_at}"
    
  - id: local_summary
    executor: providers.ollama.chat
    model: "{{ vars.model }}"
    needs: [normalize]
    prompt: |
      Summarize these GitHub issues in 2 sentences:
      {{ step.normalize.data.results }}
      
  - id: enhance
    executor: claude
    model: claude-haiku
    needs: [local_summary]
    prompt: |
      Edit this summary for clarity:
      {{ step.local_summary.data.output }}
    publish_to: "digests.ready"
    write_brain: true
```

This pipeline fetches issues, normalizes JSON with `jq`, summarizes locally with Ollama (fast), enhances with Claude (quality), publishes the result, and stores learnings for future reference.


## See Also

- [Pipeline CLI Reference](/docs/pipelines/cli-reference) — `glitch pipeline run`, `glitch pipeline resume`, flags and examples
- [Brain Notes](/docs/pipelines/brain) — How to use automatic context injection and learned insights
- [Workflows](/docs/pipelines/workflows) — Multi-pipeline orchestration with branching decisions
```

