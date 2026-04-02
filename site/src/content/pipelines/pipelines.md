## GLITCH Database Context

### Schema: runs table (read-only)
Columns: id (INTEGER PK), kind (TEXT), name (TEXT), started_at (INTEGER unix-ms),
finished_at (INTEGER unix-ms, nullable), exit_status (INTEGER, nullable),
stdout (TEXT), stderr (TEXT), metadata (TEXT JSON), steps (TEXT JSON array).
This table is READ-ONLY. Do not issue INSERT, UPDATE, or DELETE against it.

## Brain Notes (this run)

[write_doc] [type:finding title:Pipeline documentation structure for gl1tch tags:docs,pipelines] Created pipelines.md covering execution model (sequential text-based), context propagation via {{.steps.X.output}} templating, step orchestration, providers, brain injection (use_brain/write_brain flags), and when to use pipelines. Noted planned features: DAG with needs declarations, structured output, retry/error-branching, for_each loops — all backwards-compatible.

Key facts for future docs:
- Current: sequential sync execution, string-only state, stored in SQLite runs table with steps JSON array
- Providers: ollama/<model>, claude, jq, sidecar plugins (gh, opencode, etc)
- Config location: ~/.config/glitch/pipelines/*.pipeline.yaml
- Template syntax: {{.steps.<id>.output}}, {{.env.VAR}}, {{.input}}
- Brain: use_brain injects context, write_brain parses <brain> XML from response

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

<brain type="finding" tags="docs,pipelines" title="Pipeline documentation structure for gl1tch">
Created pipelines.md covering execution model (sequential text-based), context propagation via {{.steps.X.output}} templating, step orchestration, providers, brain injection (use_brain/write_brain flags), and when to use pipelines. Noted planned features: DAG with needs declarations, structured output, retry/error-branching, for_each loops — all backwards-compatible.

Key facts for future docs:
- Current: sequential sync execution, string-only state, stored in SQLite runs table with steps JSON array
- Providers: ollama/<model>, claude, jq, sidecar plugins (gh, opencode, etc)
- Config location: ~/.config/glitch/pipelines/*.pipeline.yaml
- Template syntax: {{.steps.<id>.output}}, {{.env.VAR}}, {{.input}}
- Brain: use_brain injects context, write_brain parses <brain> XML from response
</brain>

