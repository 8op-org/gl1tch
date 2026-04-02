## GLITCH Database Context

### Schema: runs table (read-only)
Columns: id (INTEGER PK), kind (TEXT), name (TEXT), started_at (INTEGER unix-ms),
finished_at (INTEGER unix-ms, nullable), exit_status (INTEGER, nullable),
stdout (TEXT), stderr (TEXT), metadata (TEXT JSON), steps (TEXT JSON array).
This table is READ-ONLY. Do not issue INSERT, UPDATE, or DELETE against it.

## Brain Notes (this run)

[polish] ` blocks. If one is found and a store is available, it's persisted for the current run. Every subsequent step receives accumulated brain notes in its prompt preamble — automatically, before your prompt text.

Brain injection and storage are controlled by two flags: `use_brain` (read context) and `write_brain` (write context). Both can be set at the pipeline level (applying to all steps) or at the step level (overriding the pipeline setting).

## Writing to the brain

Your prompt instructs the model to emit a brain block. gl1tch finds it, extracts it, stores it.

```yaml
steps:
  - id: audit
    executor: claude
    model: claude-sonnet-4-6
    write_brain: true
    prompt: |
      Audit this codebase for security issues. Be specific.
      Record your key findings in a <brain> block at the end.
```

The model outputs its analysis, then appends:

```
<brain type="finding" tags="security,sql-injection" title="SQL injection vulnerabilities">
SQL injection in user_search (line 42), admin_filter (line 89),
report_query (line 156). All use string concatenation. No parameterized queries.
[read_target] ` block.

## Writing to the brain

Your prompt instructs the model to emit a brain block. gl1tch finds it, extracts it, stores it.

```yaml
steps:
  - id: audit
    executor: claude
    model: claude-sonnet-4-6
    prompt: |
      Audit this codebase for security issues. Be specific.
      Record your key findings in a <brain> block at the end.
```

The model outputs its analysis, then appends:

```
<brain tags="security,sql-injection">
SQL injection in user_search (line 42), admin_filter (line 89),
report_query (line 156). All use string concatenation. No parameterized queries.

> Do NOT modify the runs table.

---
BRAIN NOTE INSTRUCTION: Include a <brain> block somewhere in your response to persist an insight for future steps in this pipeline.

Use the <brain> tag with structured attributes to categorize your note:

  <brain type="research" tags="optional,comma,tags" title="Human readable title">
  Your insight, analysis, or structured data here.
[deep_search] ...
[scan_docs] ` blocks. If one is found and a store is available, it's persisted for the current run. Every subsequent step receives accumulated brain notes in its prompt preamble — automatically, before your prompt text.

There are no YAML flags that control this. Brain scanning and injection are always on when a store is configured (which is the default when running via `glitch pipeline run`). The model decides what's worth remembering by whether it emits a `<brain>` block.

## Writing to the brain

Your prompt instructs the model to emit a brain block. gl1tch finds it, extracts it, stores it.

```yaml
steps:
  - id: audit
    executor: claude
    model: claude-sonnet-4-6
    prompt: |
      Audit this codebase for security issues. Be specific.
      Record your key findings in a <brain> block at the end.
```

The model outputs its analysis, then appends:

```
<brain tags="security,sql-injection">
SQL injection in user_search (line 42), admin_filter (line 89),
report_query (line 156). All use string concatenation. No parameterized queries.

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

---
title: "The Brain System"
description: "Steps write what they learn. Later steps read it. The model decides what's worth remembering."
order: 4
---

## How it works

Every step's output is scanned for `<brain>` blocks. If one is found and a store is available, it's persisted for the current run. Every subsequent step receives accumulated brain notes in its prompt preamble — automatically, before your prompt text.

Brain injection and storage are controlled by two flags: `use_brain` (read context) and `write_brain` (write context). Both can be set at the pipeline level (applying to all steps) or at the step level (overriding the pipeline setting).

## Writing to the brain

Your prompt instructs the model to emit a brain block. gl1tch finds it, extracts it, stores it.

```yaml
steps:
  - id: audit
    executor: claude
    model: claude-sonnet-4-6
    write_brain: true
    prompt: |
      Audit this codebase for security issues. Be specific.
      Record your key findings in a <brain> block at the end.
```

The model outputs its analysis, then appends:

```
<brain type="finding" tags="security,sql-injection" title="SQL injection vulnerabilities">
SQL injection in user_search (line 42), admin_filter (line 89),
report_query (line 156). All use string concatenation. No parameterized queries.
</brain>
```

gl1tch extracts the block, stores it, moves on. The full output — including the brain block — still goes to stdout.

The `<brain>` block supports structured attributes:
- `type`: Optional category (`finding`, `research`, `data`, `code`)
- `tags`: Optional comma-separated list for querying and organization
- `title`: Optional human-readable summary

## Reading from the brain

If brain notes exist for the current run and `use_brain` is enabled, they appear in the next step's prompt preamble automatically.

```yaml
use_brain: true
steps:
  - id: fetch
    executor: shell
    write_brain: true
    prompt: |
      Gather information about the codebase.
      Store findings in <brain> tags for later analysis.

  - id: analyze
    executor: claude
    model: claude-sonnet-4-6
    use_brain: true
    prompt: |
      Using the brain context provided above, 
      synthesize a security improvement plan.
```

In this pipeline, the `analyze` step automatically receives the findings stored by `fetch` in its prompt preamble.

## Controlling brain behavior with flags

### The `use_brain` flag

`use_brain: true` enables read injection — accumulated brain notes from earlier steps are prepended to the step's prompt.

**Pipeline level:**
```yaml
use_brain: true
steps:
  - id: step1
    executor: claude
    # inherits use_brain: true
  - id: step2
    executor: claude
    # inherits use_brain: true
```

**Step level override:**
```yaml
use_brain: true
steps:
  - id: step1
    executor: claude
    # uses pipeline-level use_brain: true
  - id: step2
    executor: shell
    use_brain: false
    # suppresses brain injection for this step only
```

### The `write_brain` flag

`write_brain: true` enables write injection — the step receives instructions to emit `<brain>` blocks, and its output is scanned for them and stored.

**Pipeline level:**
```yaml
write_brain: true
steps:
  - id: audit
    executor: claude
    # scan output for <brain> blocks and store them
  - id: report
    executor: claude
    # scan output for <brain> blocks and store them
```

**Step level override:**
```yaml
write_brain: true
steps:
  - id: research
    executor: claude
    # scan output for <brain> blocks and store them
  - id: filter
    executor: shell
    write_brain: false
    # do not scan or store; shell output is raw data
```

### Using both flags together

When a step has both `use_brain: true` and `write_brain: true`, it reads prior brain context and contributes new findings:

```yaml
steps:
  - id: audit
    executor: claude
    write_brain: true
    prompt: |
      Audit the codebase security. Store findings in <brain> tags.

  - id: prioritize
    executor: claude
    use_brain: true
    write_brain: true
    prompt: |
      Based on the audit findings in the brain context above,
      rank the issues by severity. Store your ranking in <brain> tags.

  - id: plan
    executor: claude
    use_brain: true
    prompt: |
      Using the prioritized issues from the brain context above,
      create a remediation plan.
```

## Brain context in the ask command

The `glitch ask` command supports brain injection with flags:

```bash
# Enable brain injection (default for local providers like Ollama)
glitch ask "explain my pipeline setup"

# Disable brain injection
glitch ask --brain=false "hello"

# Write response back to the brain store
glitch ask --write-brain "document this finding"
```

For remote providers (like Claude API), brain injection is disabled by default:

```bash
# Must opt in explicitly for remote providers
glitch ask -p claude --brain=true "explain my setup"
```

## Tips for effective brain usage

1. **Let the model decide.** Set `write_brain: true` and let the step emit `<brain>` blocks when it finds something worth remembering. Don't force every step to write.

2. **Use structured attributes.** Include `type`, `tags`, and `title` in your `<brain>` blocks to make stored notes queryable and organized.

3. **Data-fetch steps should suppress injection.** If a step pulls live data from an API (like a shell script calling curl), set `use_brain: false` to avoid polluting the model's context with stale cache.

4. **Chain analysis steps.** Build pipelines where early steps research and write findings, and later steps read those findings to synthesize decisions. This creates an audit trail of reasoning.

5. **Set pipeline defaults, override for exceptions.** Define `use_brain: true` and `write_brain: true` at the pipeline level, then set `use_brain: false` only for data-fetch steps that should work with fresh input.

