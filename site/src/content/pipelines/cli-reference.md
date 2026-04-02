---
frontmatter: (keeping exactly as provided, assuming it exists in the original file)
---

# Brain Scanning and Injection Control

The brain store allows agents to persist insights and context that automatically surface in subsequent steps. You control brain injection at two levels: the CLI (`glitch ask`) and the pipeline YAML.

## CLI Flags

The `glitch ask` command provides two brain-related flags:

### `--brain` (default: true)
Enable or disable brain context injection for the prompt.

When enabled, relevant brain notes from the store are automatically prepended to the agent's prompt. When disabled, no brain context is injected even if a store is configured.

```bash
# Use brain context (default)
glitch ask "analyze my codebase"

# Disable brain context
glitch ask --brain=false "analyze my codebase"
```

### `--write-brain` (default: false)
Enable or disable brain write injection. When enabled, the runner appends a write instruction to the prompt, asking the agent to include a `<brain>` block in its response.

```bash
# Enable brain writing
glitch ask --write-brain "audit this code"
```

Both flags can be combined:

```bash
glitch ask --brain=true --write-brain "analyze and remember insights"
```

## Pipeline YAML Flags

For finer control within pipelines, use `use_brain` and `write_brain` fields at the pipeline level (applies to all agent steps) or step level (overrides pipeline setting).

### Pipeline-Level Control

Set `use_brain` and `write_brain` at the top of your pipeline YAML to control brain injection for all agent steps:

```yaml
name: security-audit
use_brain: true       # inject brain context for all steps
write_brain: false    # disable brain writes by default

steps:
  - id: scan
    executor: ollama
    model: llama3.2
    prompt: |
      Scan this codebase for security vulnerabilities.
```

### Step-Level Control

Override pipeline-level flags on individual steps using `use_brain` and `write_brain`:

```yaml
name: feedback-loop
use_brain: false      # disabled by default

steps:
  - id: audit
    executor: claude
    use_brain: true   # override: enable for this step
    write_brain: true
    prompt: |
      Audit this codebase for security issues.
      Record key findings in a <brain> block.

  - id: report
    executor: claude
    use_brain: true   # inject audit findings from step 1
    prompt: |
      Summarize the security findings for a manager.
```

In this example:
- Step `audit` receives no injected brain context (pipeline-level `use_brain: false`), but can write new notes via `write_brain: true`
- Step `report` receives the notes written by `audit` because its `use_brain: true` overrides the pipeline setting
- Brain notes written in `audit` are automatically surfaced in `report` within the same pipeline run

### Tri-State Behavior

Step-level `use_brain` and `write_brain` follow a tri-state model:

| Step Setting | Pipeline Setting | Result |
|---|---|---|
| (not set) | true | Inherit pipeline setting (use_brain: true) |
| (not set) | false | Inherit pipeline setting (use_brain: false) |
| true | false | Use brain injection for this step only |
| false | true | Suppress brain injection for this step |

## Brain Block Format

When `write_brain` is enabled, the agent is instructed to include a `<brain>` block in its response. Use structured attributes to make notes queryable:

```xml
<brain type="finding" tags="auth,security" title="Session Token Vulnerability">
Session tokens are stored in ~/.glitch/session without encryption.
Risk: tokens accessible via filesystem access.
Recommendation: use OS keyring or encrypted storage.
File: internal/auth/session.go, line 42.
</brain>
```

Supported types:
- `research` — background info, context, or references
- `finding` — concrete discovery (bug, pattern, fact)
- `data` — structured output (metrics, counts, lists)
- `code` — code snippet or file path reference

The `tags` and `title` attributes are optional but recommended for discoverability.

## Examples

### Example 1: Single-Step Brain Write

```yaml
name: collect-insights
steps:
  - id: research
    executor: ollama
    model: mistral
    write_brain: true
    prompt: |
      Research best practices for Go error handling.
      Summarize in a <brain> block at the end.
```

The response will be scanned for a `<brain>` block and persisted to the store automatically.

### Example 2: Multi-Step Feedback Loop

```yaml
name: code-review-loop
steps:
  - id: review
    executor: claude
    use_brain: false
    write_brain: true
    prompt: |
      Review this Go function for improvements.
      Record issues found in a <brain> block.

  - id: refactor
    executor: claude
    use_brain: true
    prompt: |
      Based on the review findings, refactor this function.
```

Step `refactor` automatically receives the issues found by step `review` in its context.

### Example 3: Selective Injection

```yaml
name: data-fetch-then-analyze
use_brain: true  # default for all steps

steps:
  - id: fetch
    executor: shell
    use_brain: false  # suppress for shell steps
    prompt: "curl https://api.example.com/data"

  - id: analyze
    executor: claude
    # inherits use_brain: true from pipeline
    prompt: |
      Analyze the fetched data with any relevant
      background knowledge from the brain store.
```


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