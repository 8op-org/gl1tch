# GUI Redesign v2 — Spec

**Date**: 2026-04-17
**Status**: Draft

## Overview

Complete redesign of the gl1tch workspace GUI. Kill the hover-slideout sidebar, consolidate runs and results under workflows, add a pipeline flow graph for run visualization, fix broken settings, replace CodeMirror with Monaco, and add comprehensive Playwright tests.

## 1. Navigation — Icon Activity Bar

Replace the collapsible hover-slideout sidebar with a fixed 48px icon rail on the left.

**Items:**
- **Top**: Workflows (grid icon)
- **Bottom**: Settings (gear icon)
- **Very top**: Workspace switcher (avatar/icon)

Runs and Results are no longer top-level routes. They are accessed through the Workflow Detail page.

**Behavior:**
- Always visible, never collapses, never slides
- Active icon: cyan left border + glow
- No text labels — icons only, tooltip on hover

**Removed routes:**
- `/runs` — absorbed into Workflow Detail
- `/results` — absorbed into Run Detail (node panel)

**Remaining routes:**
- `/` — Workflow List
- `/workflow/:name` — Workflow Detail (tabs: Runs, Source, Metadata)
- `/settings` — Settings

## 2. Workflow List Page

Grid of workflow cards as the landing page.

**Card contents:**
- Name, description, tags
- Last run status badge (pass/fail/running)
- Last run time (relative, e.g. "3h ago")
- Run count (total executions)

**Page features:**
- Search bar + tag filter pills at top
- View toggle: grid / list
- Click card → navigates to Workflow Detail

## 3. Workflow Detail Page

Core page of the application. Breadcrumb at top: `Workflows / <workflow-name>`.

**Three tabs:**

### Runs Tab (default)
Vertical list of runs for this workflow, newest first. Each row shows:
- Status badge, run ID, started time, duration
- Model, tokens in/out, cost
- Click to expand inline → renders the Pipeline Flow Graph (section 4)

### Source Tab
- Monaco editor with glitch syntax highlighting
- Save button (Cmd/Ctrl+S shortcut)
- Dirty detection indicator

### Metadata Tab
- Name, description, tags, author, version
- Workflow parameters definition
- Read from workflow source file

**Run button** in top-right header — opens run dialog with param form.

## 4. Pipeline Flow Graph (Run Detail)

When a run row is expanded in the Runs tab, the pipeline flow graph renders below it. Uses **elkjs** for layout.

### Nodes
- **Shape**: Rounded rectangle, dark surface (`#1a2230`), 1px border
- **Border color by status**:
  - Cyan (`#00e5ff`) — pass
  - Magenta (`#ff2d6f`) — fail
  - Amber (`#ffb800`) — running (pulse animation)
  - Muted gray (`#5a6a7a`) — pending/skipped
- **Node content**: step ID, kind badge (llm/run/cond/map), duration, token count

### Edges
- Animated edges between nodes, flowing left-to-right
- Edge color matches source node status
- Smooth bezier curves via elkjs routing

### Branching
- `par` steps fan out vertically into parallel lanes, then converge
- `call-workflow` nodes are expandable — click to unfold child workflow's graph inline (nested, indented)
- `map` shows collapsed stack icon with item count, expandable to per-item runs

### Node Detail Panel
Click a node to open a split pane on the right:
- **Output**: rendered markdown
- **Prompt**: collapsible section
- **Metrics**: model, duration, tokens in/out, cost, gate result
- **Result files**: list of files produced by this step, viewable inline with Monaco editor

## 5. Settings Page

Keep existing structure, fix broken sections.

### Workflow Defaults (working)
- Model, provider, params

### Workspace (working)
- Name, Kibana URL

### Resources (broken — fix)
- Table: name, type, ref, pin, last-synced timestamp, status indicator (synced/stale/error)
- Actions: sync, pin, remove — all functional
- Add resource form

### Repos (broken — fix)
- List cloned repos: path, remote URL, current ref, sync status
- Sync action per repo

## 6. Tech Changes

| Change | From | To |
|--------|------|----|
| Editor | CodeMirror 6 | Monaco Editor |
| Graph layout | N/A | elkjs |
| Navigation | Hover-slideout sidebar | Fixed icon activity bar |
| Runs page | Standalone route | Embedded in Workflow Detail |
| Results page | Standalone route | Embedded in node detail panel |

**Keep:** Svelte 5, Vite 6, cyberpunk design system (colors, fonts, glows), existing API layer.

**Remove:**
- `Sidebar.svelte` (hover-slideout)
- `routes/RunList.svelte` (standalone)
- `routes/ResultsBrowser.svelte` (standalone)
- `FileTree.svelte` (replaced by Monaco's built-in tree or simplified file list in node panel)

**Add:**
- `ActivityBar.svelte` — fixed icon rail
- `PipelineGraph.svelte` — elkjs-powered flow graph
- `GraphNode.svelte` — individual step node
- `NodePanel.svelte` — right-side detail panel with Monaco viewer
- Monaco editor integration (replace all CodeMirror instances)

## 7. API Changes

No new endpoints needed. Existing API supports everything:

- `GET /api/workflows` — list with last_run_status (add run_count if missing)
- `GET /api/runs?parent_id=N` — filter runs by workflow
- `GET /api/runs/{id}` — run with steps array (feeds the graph)
- `GET /api/runs/{id}/tree` — nested run tree for call-workflow expansion
- `GET /api/results/{path}` — file content for node panel
- Resource/repo endpoints — already exist, just need frontend fixes

**New endpoint:** `GET /api/workflows/{name}/runs` — list runs for a specific workflow, filtered server-side by workflow_name. Required for the Runs tab to load efficiently without client-side filtering.

## 8. Playwright Test Plan

Each spec file maps to a page or feature:

### `e2e/activity-bar.spec.js`
- Icon rail renders with correct icons
- Active state (cyan border) on current route
- Workspace switcher opens and switches

### `e2e/workflow-list.spec.js`
- Cards render with name, description, tags, status badge, run count
- Search filters cards by name/description
- Tag filter pills toggle and filter
- Grid/list view toggle works
- Click card navigates to workflow detail

### `e2e/workflow-detail.spec.js`
- Breadcrumb renders and navigates back
- Three tabs render and switch correctly
- Runs tab: timeline loads newest first, run rows show correct metadata
- Source tab: Monaco editor loads workflow source, save works, dirty detection
- Metadata tab: displays name, description, tags, author, version
- Run button opens dialog, submits, new run appears in timeline

### `e2e/pipeline-graph.spec.js`
- Expanding a run renders the flow graph
- Nodes render with correct status colors (pass=cyan, fail=magenta, running=amber)
- Edges connect nodes left-to-right
- Parallel steps fan out into lanes and converge
- Click node opens detail panel on right
- Detail panel shows output, prompt, metrics
- Result files listed and openable in Monaco

### `e2e/nested-runs.spec.js`
- call-workflow nodes show expand indicator
- Expanding renders child workflow graph inline
- Multiple nesting levels render correctly
- Child run metrics display correctly

### `e2e/settings.spec.js`
- Workflow defaults: model/provider/params save and persist
- Resources: add resource, sync, pin, remove — all functional
- Repos: list repos, sync, show status
- Form dirty detection works

### Test infrastructure
- Runs against real workspace (same as today)
- Chromium only
- Screenshot + trace on failure
- 30s timeout per test
