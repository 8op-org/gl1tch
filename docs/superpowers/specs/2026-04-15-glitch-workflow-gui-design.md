# glitch workflow gui

**Date:** 2026-04-15
**Status:** Approved

## Problem

gl1tch workflows produce structured results (markdown plans, PR bodies, reviews, JSON classifications) but there is no way to browse, edit, run, or review them without the terminal. The Kibana dashboard covers telemetry. The Astro dashboard covers personal productivity. Neither covers workflow management.

## Decision

Add `glitch workflow gui` as a built-in command that starts a local web server for managing workflows in any gl1tch workspace.

## Scope (v1)

1. Browse workflows in the workspace (list, view source)
2. Edit workflows with CodeMirror (sexpr syntax highlighting)
3. Run a workflow (set params, kick it off)
4. View results with markdown rendering and syntax highlighting
5. Re-run a workflow (same or modified params)
6. Embedded Kibana telemetry — per-run and per-workflow aggregate views

### Explicitly deferred

- Live tail of running workflows (WebSocket streaming)
- Batch management UI
- Plugin browser
- Run diffing
- Config editor

## Architecture

```
glitch binary
  cmd/gui.go              Cobra command, starts HTTP server
  internal/gui/
    server.go             Router, static files, go:embed for prod
    api.go                REST endpoints
    ws.go                 WebSocket stub (v1: run status polling, later: streaming)
  gui/                    Svelte + Vite frontend (top-level, alongside site/)
    src/
      routes/             Pages: workflows, editor, results, run
      lib/                CodeMirror setup, markdown renderer, API client
      App.svelte
    vite.config.js
    package.json
```

### Dev vs prod

- **Dev:** `glitch workflow gui --dev` — Go server on `:8374`, Vite dev server on `:5173` with proxy to `/api`. Hot reload for frontend.
- **Prod:** `glitch workflow gui` — Go serves embedded `gui/dist/` via `go:embed`. Single binary, single port.

### Go internals used

The GUI command imports gl1tch packages directly (no subprocess shelling):

- `internal/pipeline` — run workflows, read step definitions
- `internal/store` — query past runs and results from SQLite
- `internal/dashboard` — extend existing Kibana seeding with per-workflow saved searches
- `internal/esearch` — query ES for telemetry data
- Filesystem — read/write `.glitch` files, read result artifacts

## REST API

```
GET  /api/workflows              List .glitch files in workspace
GET  /api/workflows/:name        Read workflow source
PUT  /api/workflows/:name        Save edited workflow
POST /api/workflows/:name/run    Start a run (params in JSON body)
GET  /api/runs                   List past runs from SQLite store
GET  /api/runs/:id               Run detail + result file listing
GET  /api/results/*path          Read a result file (md, json, txt)
GET  /api/kibana/workflow/:name  Kibana iframe URL for workflow aggregate view
GET  /api/kibana/run/:id         Kibana iframe URL for specific run
WS   /api/runs/:id/stream        Live step output (stub in v1)
```

All endpoints scoped to the resolved workspace path.

## Pages

### Workflow list

Left sidebar listing all `.glitch` files in `workflows/`. Click to open in editor. Shows file name and first-line comment as description.

### Editor

CodeMirror 6 with custom sexpr language mode (Lezer grammar, ~50 lines for s-expression syntax). Save button writes back via `PUT /api/workflows/:name`. "Run" button opens the run dialog.

### Run dialog

Modal form. Extracts `{{.param.*}}` references from the workflow source to auto-generate input fields. User fills in values (e.g. repo, issue), clicks Start. Posts to `POST /api/workflows/:name/run`.

### Run view

Shows run metadata (workflow, params, start time, status). v1: polls `GET /api/runs/:id` for status updates. Displays step list with pass/fail indicators. Links to result files when complete. Embedded Kibana panel showing LLM calls, latency, and cost for this specific run (filtered by run_id).

### Workflow overview (telemetry)

Each workflow page includes an embedded Kibana panel showing aggregate telemetry across all runs: average latency per step, total cost over time, success/failure rate. Useful for spotting trends across batch runs.

### Results browser

Tree view of `results/` directory. Click a file to view:
- `.md` files rendered as HTML (marked + highlight.js)
- `.json` files with syntax highlighting
- `.txt` files as plain text

## Frontend stack

- **Framework:** Svelte 5 + Vite
- **Editor:** CodeMirror 6 with custom Lezer grammar for sexpr
- **Markdown:** marked + highlight.js
- **Routing:** svelte-spa-router (hash-based, no server routing needed)
- **Styling:** Plain CSS, no framework. Dark theme default (matches terminal workflow).

## Key decisions

| Decision | Rationale |
|----------|-----------|
| Lives in gl1tch, not separate repo | Workspace is a gl1tch concept; `workflow gui` is a natural CLI extension |
| Svelte, not Astro | Astro is SSG-first; real-time editor and run management need reactivity |
| Import internals, not shell out | GUI is inside the binary; direct access to pipeline/store is cleaner |
| Port 8374, bind 127.0.0.1 | Local-only tool, no auth needed |
| go:embed for prod | Single binary distribution, no runtime deps |
| Custom Lezer grammar | Sexpr is simple (~50 lines); no existing grammar fits the gl1tch dialect |
| Deferred: live streaming | WebSocket plumbing exists but v1 uses polling for simplicity |
| Kibana embed, not custom charts | Kibana already renders telemetry well; embedding avoids rebuilding charting |

## File changes

### New files (gl1tch repo)

- `cmd/gui.go` — Cobra command registration
- `internal/gui/server.go` — HTTP server, router, static file serving
- `internal/gui/api.go` — REST handlers
- `internal/gui/ws.go` — WebSocket stub
- `internal/gui/kibana.go` — Kibana saved search creation + iframe URL generation
- `gui/` — Entire Svelte frontend directory
- `gui/src/lib/sexpr.grammar` — Lezer grammar for sexpr highlighting

### Modified files

- `cmd/workflow.go` — Add `gui` subcommand
- `go.mod` / `go.sum` — No new Go deps needed (net/http is stdlib)
- `.gitignore` — Add `gui/node_modules/`, `gui/dist/`
