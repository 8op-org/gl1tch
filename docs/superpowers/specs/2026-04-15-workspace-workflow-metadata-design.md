# Workspace & Workflow Metadata Enrichment

**Date:** 2026-04-15
**Status:** Draft

## Problem

The glitch GUI has thin APIs — workflow list returns only name+description, runs show basic SQLite rows, and results are raw file browsing. There's no workspace-level identity file. A human looking at the GUI gets minimal context about what they're seeing, and glitch itself can't leverage richer metadata for filtering, grouping, or cost visibility.

## Solution

Additive metadata across three layers — workspace manifest, workflow keywords, and SQLite columns — all optional, all backward-compatible, all surfaced through enriched GUI API endpoints.

## Layer 1: `workspace.glitch` File Format

New file at `<workspace>/workspace.glitch` (e.g., `~/Projects/stokagent/workspace.glitch`). Lives alongside the `workflows/` and `results/` directories. Parsed by the same s-expression engine. Loaded when `--workspace` is set; absence is not an error — glitch operates without it.

```glitch
(workspace "stokagent"
  :description "Cross-repo research and automation for Elastic observability"
  :owner "adam"

  ;; Target repos this workspace covers
  (repos
    "elastic/observability-robots"
    "elastic/ensemble"
    "elastic/oblt-cli")

  ;; Config overrides (take precedence over ~/.config/glitch/config.yaml)
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"))
```

### Parsing

The `(workspace ...)` form is the top-level form, analogous to `(workflow ...)`. The parser:

- Reads `:description`, `:owner` as string keywords
- Reads `(repos ...)` as a list of string atoms
- Reads `(defaults ...)` as keyword-value pairs for config overrides

### What glitch does with it

- `glitch gui` renders a workspace home page with name, description, owner, and repo list.
- `(defaults ...)` overrides global config when running workflows in this workspace.
- `(repos ...)` gives the GUI a scoped list for filtering runs/results instead of showing everything.
- CLI can validate that `glitch ask elastic/kibana#1234` targets a repo this workspace covers (warning, not error).

### Go types

```go
type Workspace struct {
    Name        string
    Description string
    Owner       string
    Repos       []string
    Defaults    WorkspaceDefaults
}

type WorkspaceDefaults struct {
    Model    string
    Provider string
}
```

## Layer 2: Workflow Metadata Enrichment

New optional keywords on the `(workflow ...)` form. Existing workflows keep working — all new fields are optional.

```glitch
(workflow "pr-review"
  :description "Review PRs for code quality and test coverage"
  :tags ("review" "ci" "code-quality")
  :author "adam"
  :version "1.0"
  :created "2026-04-01"

  (step "fetch" ...))
```

### New fields

| Keyword | Type | Purpose |
|---------|------|---------|
| `:tags` | s-expression list of strings | GUI filtering, grouping, search |
| `:author` | string | Track who wrote/owns the workflow |
| `:version` | string | Semver-ish, human-managed |
| `:created` | string (date) | When the workflow was first written |

### What glitch does with it

- `handleListWorkflows` returns all fields — GUI can render tag chips, group by category, show author badges.
- `glitch workflow list` CLI gains `--tag` filter (e.g., `glitch workflow list --tag review`).
- Tags are freeform strings, no controlled vocabulary — the user decides what taxonomy makes sense.
- Version is informational only, no enforcement or auto-increment.

### Go changes

- `Workflow` struct gets `Tags []string`, `Author string`, `Version string`, `Created string`.
- `convertWorkflow()` in `internal/pipeline/sexpr.go` handles the new keywords.
- `workflowEntry` in `internal/gui/api_workflows.go` includes the new fields in JSON responses.

## Layer 3: SQLite Run Enrichment

New columns on the existing `runs` and `steps` tables. Pre-1.0, so wipe-and-restart — no migration code.

### `runs` table additions

| Column | Type | Source |
|--------|------|--------|
| `workflow_file` | TEXT | Which `.glitch` file was executed |
| `repo` | TEXT | Target repo (e.g., `elastic/ensemble`) |
| `model` | TEXT | Primary model used |
| `tokens_in` | INTEGER | Total input tokens |
| `tokens_out` | INTEGER | Total output tokens |
| `cost_usd` | REAL | Estimated cost |
| `variant` | TEXT | Variant label if applicable |

### `steps` table additions

| Column | Type | Source |
|--------|------|--------|
| `kind` | TEXT | `run`, `llm`, `save`, `gate`, etc. |
| `exit_status` | INTEGER | 0/1 for shell steps, null for LLM |
| `tokens_in` | INTEGER | Per-step token count |
| `tokens_out` | INTEGER | Per-step token count |
| `gate_passed` | INTEGER | Boolean — did this gate pass? NULL if not a gate |

### What glitch does with it

- `handleListRuns` returns model, repo, cost, token totals — GUI run list shows at a glance what each run cost and targeted.
- `handleGetRun` returns per-step kind, gate status, token breakdown — GUI run detail can render a step-by-step timeline with pass/fail indicators on gates.
- `run.json` continues to be written as the portable/archival format, generated from the same data.

### Go changes

- Schema in `internal/store/schema.go` updated with new columns.
- Runner writes the new fields as steps execute.
- GUI API structs (`runEntry`, `stepEntry`) include new fields.

## Layer 4: GUI API Changes

### New endpoint

- `GET /api/workspace` — returns the parsed `workspace.glitch` (name, description, owner, repos, defaults). The GUI home page consumes this.

### Enriched endpoints

`GET /api/workflows` response changes from:

```json
[{"name": "pr-review", "file": "pr-review.glitch", "description": "..."}]
```

to:

```json
[{
  "name": "pr-review",
  "file": "pr-review.glitch",
  "description": "...",
  "tags": ["review", "ci"],
  "author": "adam",
  "version": "1.0",
  "created": "2026-04-01"
}]
```

`GET /api/runs` response adds: `workflow_file`, `repo`, `model`, `tokens_in`, `tokens_out`, `cost_usd`.

`GET /api/runs/{id}` step entries add: `kind`, `exit_status`, `tokens_in`, `tokens_out`, `gate_passed`.

No new auth, no pagination changes, no breaking changes — all new fields are additive. GUI renders what's present, ignores what's missing.

## What changes in glitch core

1. **New parser:** `parseWorkspaceFile()` in `internal/pipeline/` (or new `internal/workspace/` package) for `workspace.glitch`.
2. **Workflow struct enrichment:** four new fields, keyword handling in `convertWorkflow()`.
3. **SQLite schema:** new columns on `runs` and `steps` tables, wipe-and-restart.
4. **Runner instrumentation:** write enriched metadata as steps execute.
5. **GUI API:** new `/api/workspace` endpoint, enriched response structs on existing endpoints.
6. **CLI:** `glitch workflow list --tag` filter.

## What does NOT change

- Workflow execution semantics — no behavioral changes.
- S-expression grammar — only new keywords on existing forms.
- Result directory structure — `run.json` stays as-is, just populated from richer data.
- Global config at `~/.config/glitch/config.yaml`.
- Any behavior when `workspace.glitch` is absent.
