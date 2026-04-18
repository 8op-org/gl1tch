# Run Detail Page Design

**Date:** 2026-04-17
**Route:** `/run/:id`

## Overview

A dedicated page for viewing pipeline run details. Graph-centric layout with a slide-over detail panel. Supports both live monitoring (2s polling) and post-mortem debugging. Replaces the cramped inline expansion in WorkflowDetail.

## Route & Navigation

- New route: `/run/:id` added to App.svelte (4th route)
- RunDialog already pushes to `/run/{id}` after kicking off a workflow
- WorkflowDetail run rows link to `/run/:id` instead of expanding inline
- Breadcrumb: `Workflows / <workflow-name> / Run #<id>`
- Nested runs: `Workflows / <name> / Run #<parent-id> / <step-name> / Run #<child-id>`

## Page Layout

### Header Bar (~80px, fixed)

- **Left**: Breadcrumb trail with clickable links back to workflow list and workflow detail
- **Right**: Run status badge (animated pulse for running), "Re-run" button

### Metadata Strip (~48px, single row of pills)

Status | Duration | Model | Tokens (in/out) | Cost | Started timestamp

For running runs: duration ticks client-side every second, tokens/cost update on each poll.

### Graph Area (remaining viewport height)

Full-width pipeline graph. No page-level scroll — graph container manages its own overflow.

## Pipeline Graph

### Auto-Fit

ELK layout runs at natural node sizes (180x68). The SVG uses a `viewBox` to scale the graph to fit the container width. No horizontal scrollbar for typical workflows (up to ~8-10 nodes). For larger workflows, zoom controls appear.

### Zoom & Pan

- Mouse wheel zooms
- Click-drag pans
- Minimap in bottom-left corner (viewport rectangle, like Figma)
- Reset-zoom button snaps to fit-all

### Edge Visibility Fix

- Edges bumped to 2px stroke width
- Subtle glow via `filter: drop-shadow` matching status color
- Status-based colors: cyan (pass), magenta (fail), amber (running)
- Running edges: animated dashed stroke

### Nested Run Drill-Down

When a step has `kind === 'call-workflow'` and a child run exists:
1. Node shows a "drill in" icon
2. Click fetches child run via `/api/runs/{child-id}`
3. Current graph replaced with child run's graph (crossfade transition)
4. Breadcrumb updates to show nesting path
5. "Back to parent" button appears above graph

## Slide-Over Detail Panel

Opens from right edge when a graph node is clicked. Graph container shrinks to make room (300ms ease animation).

### Panel Specs

- **Width**: 420px fixed
- **Close**: X button, Escape key, or click same node again
- **Node swap**: Clicking a different node swaps content in-place (no close/reopen)

### Panel Header

Step name, kind badge (colored), status badge, duration.

### Tabs

1. **Output** (default) — Markdown-rendered step output. Shell steps (`kind === 'run'`) in monospace code block. Empty states: "Step still running..." with spinner for active steps, "No output captured" for finished steps with no data.

2. **Prompt** — Read-only CodeMirror editor with scheme syntax highlighting. For shell steps, shows the command executed.

3. **Metrics** — Key-value display: Model, Duration, Tokens In, Tokens Out, Exit Status, Gate Passed. Labels muted, values bold.

4. **Files** — Lists files from `step.artifacts` (see Result Files section). Each file shows name + size. Click opens CodeMirror read-only viewer in-panel. Back arrow to return to list. Download button per file.

### Live Behavior

Panel content refreshes each poll cycle. Output tab auto-scrolls as new content appends. If user scrolls up, auto-scroll pauses. "Jump to bottom" pill to re-enable.

## Result Files Linking

### Backend: Expose run_id to Templates

Add `run_id` to the `data` map in `runSingleStep()` so workflow authors can use `~run_id` in save paths. Additive change — existing workflows unaffected.

### Backend: Record Artifacts on Step Records

New `artifacts` column on `steps` table (TEXT, JSON array of strings). After a `save` or `write_file` step succeeds, append the resolved file path to the array. No path convention imposed — linking is by recording what was actually written.

Schema change:
```sql
ALTER TABLE steps ADD COLUMN artifacts TEXT;
```

Pre-1.0 so no migration — just add the column to the CREATE TABLE in schema.go. Wipe and restart.

### Backend: Include Artifacts in API Response

Add `Artifacts` field to `stepEntry` struct. Deserialize from JSON in the query. The existing `/api/runs/{id}` endpoint returns it with step data.

### Frontend: Files Tab

Reads `step.artifacts` array. Fetches each file via existing `/api/results/{path}` endpoint. No new API endpoints.

### Frontend: Artifacts Bar

Below the graph, collapsible bar showing total artifact count across all steps. Expands to flat file list grouped by step name. Click opens file in slide-over panel.

## Live Polling

### Lifecycle

1. On mount: fetch `/api/runs/{id}` once
2. If `finished_at` is null: start 2-second `setInterval`
3. Each tick: fetch `/api/runs/{id}`, diff against current state
4. When `finished_at` appears: clear interval

### Diff Strategy

Compare steps by `step_id`:
- **New step**: Add node to graph, ELK re-layout with crossfade
- **Status changed**: Transition node border color
- **Output grew**: Append in slide-over Output tab if open

### ELK Re-Layout on New Steps

Existing nodes animate to new positions via CSS transition on `transform`. New nodes fade in at ELK-computed position. Graph stays stable.

### Error Handling

Poll failure: subtle toast "Connection lost, retrying..." — retry on next interval. After 5 consecutive failures: pause polling, show "Reconnect" button. Never clear existing data on error.

## Components

### New

- `RunDetail.svelte` — page component, owns polling lifecycle and state
- `RunHeader.svelte` — breadcrumb + status + re-run button
- `MetadataStrip.svelte` — row of metric pills
- `ArtifactsBar.svelte` — collapsible file list below graph
- `SlideOverPanel.svelte` — generic slide-over container (could be reused)

### Modified

- `App.svelte` — add `/run/:id` route
- `PipelineGraph.svelte` — add auto-fit viewBox, zoom/pan, minimap, edge glow, drill-down support
- `GraphNode.svelte` — add drill-in icon for call-workflow nodes
- `NodePanel.svelte` — refactor into SlideOverPanel content; add Files tab with artifact support
- `WorkflowDetail.svelte` — run rows link to `/run/:id` instead of inline expand
- `api.js` — no new endpoints needed, existing `getRun()` and `getResult()` suffice

### Backend Modified

- `internal/pipeline/runner.go` — add `run_id` to template data map; record artifacts after save/write_file
- `internal/store/schema.go` — add `artifacts` column to steps table
- `internal/gui/api_runs.go` — include `artifacts` in stepEntry response

## Test Coverage

### Playwright E2E Tests

New test file: `gui/e2e/run-detail.spec.js`

- Navigate to `/run/:id` for a completed run — page loads, header shows correct breadcrumb
- Metadata strip shows status, duration, model, tokens, cost
- Pipeline graph renders with correct number of nodes
- Click a node — slide-over panel opens with step name and tabs
- Output tab shows step output content
- Prompt tab shows step prompt
- Metrics tab shows duration, tokens, exit status
- Files tab shows artifacts (requires a test workflow with a save step)
- Close panel via X button, Escape key, and re-click node
- Click different node — panel swaps content without close animation
- Breadcrumb links navigate back to workflow list and workflow detail
- Graph edges are visible (2px, glow)

### Live Polling Tests (requires running a workflow)

- Start a workflow, navigate to `/run/:id` — page shows running state
- Nodes appear as steps execute
- Status transitions from running to pass/fail
- Polling stops after run completes

### Nested Run Tests (if call-workflow exists in test workflows)

- Drill-in icon appears on call-workflow nodes
- Click drill-in — child graph replaces parent
- Breadcrumb shows nesting path
- Back button returns to parent graph
