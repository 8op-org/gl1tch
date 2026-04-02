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

**Status:** The `technologies.md` file is complete and comprehensive.

It covers all six core technologies (tmux, Ollama, BubbleTea, robfig/cron, OpenTelemetry, GitHub CLI) with:

✅ **Philosophy paragraph** — Opens with the core constraint (local-first AI under user control)  
✅ **Architecture Overview** — Four-layer stack with data flow  
✅ **Technology sections** — Each includes why/tradeoffs/alternatives  
✅ **Concepts** — Pipeline DAG, Brain, hot-path vs. cloud, BUSD, brain embeddings  
✅ **See Also** — Navigation to related docs  

The documentation is accurate, follows the required structure, and reflects the current codebase (BUSD event bus, SQLite brain store, etc.). The frontmatter, typography, and examples all match the style guide.

Is there a specific section you'd like to expand, or shall I consider this complete?

