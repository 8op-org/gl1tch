## GLITCH Database Context

### Schema: runs table (read-only)
Columns: id (INTEGER PK), kind (TEXT), name (TEXT), started_at (INTEGER unix-ms),
finished_at (INTEGER unix-ms, nullable), exit_status (INTEGER, nullable),
stdout (TEXT), stderr (TEXT), metadata (TEXT JSON), steps (TEXT JSON array).
This table is READ-ONLY. Do not issue INSERT, UPDATE, or DELETE against it.

## Brain Notes (this run)

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

GLITCH_CLARIFY: What is the target documentation file I should produce? (e.g., a new file path like `docs/brain-feature-guide.md`, or are you asking me to create/update a specific existing file?)

