# gl1tch GUI Redesign — Cyberpunk Dev Tool

**Date:** 2026-04-15
**Status:** Approved
**Depends on:** 2026-04-15-glitch-workflow-gui-design.md (v1 GUI, functional)

## Problem

The v1 GUI is functional (5 routes, 9 API endpoints, 20 passing Playwright tests) but visually bare. Flat lists, no hierarchy, no icons, minimal styling, no loading states. It needs to look and feel like a real product.

## Scope

Redesign every visual surface of the existing GUI. Add a full file browser with editing and a workflow action system to the results view. No new backend endpoints except where noted. No changes to the Go embed/server architecture.

### In scope

- Custom CSS design system (no component library)
- Sidebar navigation replacing top nav
- Workflow card grid with filtering
- Editor with metadata side panel
- Restyled runs table and run detail
- Full file browser with preview/edit modes
- Workflow action system: `(action ...)` keyword in workflow metadata, GUI discovers and surfaces actions dynamically
- highlight.js CSS import (currently missing)
- Loading states, transitions, visual feedback
- New sexpr keyword: `(action "context:name")` — requires parser update in `internal/pipeline/sexpr.go`

### Out of scope

- Sexpr CodeMirror grammar (separate task)
- WebSocket live streaming
- Batch management UI
- Plugin browser
- Auth/multi-user

## Visual Identity

### Palette

| Token | Value | Use |
|---|---|---|
| `--bg-deep` | `#0a0e14` | Page background |
| `--bg-surface` | `#111820` | Panels, cards |
| `--bg-elevated` | `#1a2230` | Hover, active items |
| `--neon-cyan` | `#00e5ff` | Primary actions, links, active nav |
| `--neon-magenta` | `#ff2d6f` | Danger, errors, FAIL status |
| `--neon-amber` | `#ffb800` | Warnings, RUNNING status |
| `--neon-green` | `#00ff9f` | Success, PASS status |
| `--text-primary` | `#e0e6ed` | Body text |
| `--text-muted` | `#5a6a7a` | Secondary text, labels |
| `--border` | `#1e2a3a` | Panel borders |

### Glow effects

Used sparingly for interactive feedback only:
- Active nav items: `box-shadow: 0 0 8px var(--neon-cyan)`
- Status badges: faint glow matching their status color
- Focus rings: neon cyan glow replacing browser default

### Typography

- UI text: Inter or system sans-serif, 13px base
- Code/data: JetBrains Mono or Fira Code, 12px
- Headings: medium weight, slightly larger. No bold-everything.

### Surfaces

1px solid borders with `--border` color. No heavy shadows. Panels feel like glass on a dark void. The cyberpunk comes from color and glow, not gradients or heavy effects.

## Layout Shell

CSS Grid: sidebar + main content area.

```
+--------+----------------------------------+
|        |                                  |
| SIDEBAR|         MAIN CONTENT             |
|  56px  |          flex: 1                 |
| (200px |                                  |
|  hover)|                                  |
|        |                                  |
+--------+----------------------------------+
```

### Sidebar (56px collapsed, 200px expanded)

- Top: gl1tch logo (monospace, cyan glow on hover)
- Collapsed by default, expands on hover or toggle
- Nav items: Workflows, Runs, Results — icon + label
- Active item: cyan left border + icon glow
- Bottom: workspace name (truncated), settings gear

### Icons

Inline SVG only. No icon library dependency. ~8 icons needed: workflow, play, folder, terminal, settings, search, file, chevron. Source from Lucide (tree-shakeable, ~300 bytes per icon) or hand-draw.

### Main content area

- Top bar per route: breadcrumb + page title + action buttons (right-aligned)
- Consistent 24px padding
- Independent scroll from sidebar

## Route: Workflow List (Card Grid)

### Filter bar

- Search input with magnifying glass icon, cyan border on focus
- Tag filter pills: click to toggle active, active tags glow cyan
- Sort dropdown: name, last run, recently modified

### Workflow card

```
+-----------------------------------+
| cross-review.glitch               |
|                                   |
| Cross-repo PR review with         |
| convention checks                 |
|                                   |
| [review] [elastic] [v1.2]        |
|                                   |
| * PASS  -  2h ago  -  @adam       |
+-----------------------------------+
```

- Background: `--bg-surface`, border: `--border`
- Hover: border shifts to `--neon-cyan` at low opacity
- Name: monospace, cyan, clickable (opens editor)
- Description: first line from workflow source, `--text-muted`
- Tags: small pills with `--bg-elevated` background
- Footer: last run status (colored dot), relative time, author
- No runs: "Never run" in muted text

Responsive grid: 3 columns wide, 2 medium, 1 narrow.

## Route: Workflow Editor

### Layout

Two-panel: code editor (flex: 1) + metadata side panel (250px, collapsible).

### Top bar

- Breadcrumb: "Workflows / cross-review.glitch" — Workflows links back
- Right: Save button (muted until dirty, then cyan), Run button (neon cyan, always visible)
- Save indicator: dot that flashes green on save success

### Editor pane

- CodeMirror with One Dark theme, background adjusted to `--bg-deep`
- Auto-save on blur or Cmd+S
- Line numbers in `--text-muted`

### Metadata panel (right, 250px)

- Tags, author, version, created date
- Recent runs: last 5, each with status dot + relative time, clickable
- Toggle to collapse — editor takes full width

### Run dialog (modal)

- Overlay with backdrop blur
- Neon cyan border on dialog
- Input fields: dark background, cyan focus ring
- Dynamic form from extracted workflow params
- "No parameters required" message when none
- "Start Run" button: neon cyan with glow on hover
- On success: navigate to run detail

## Route: Runs List

Data table:

| Column | Content |
|---|---|
| ID | `#047` (monospace, cyan link) |
| Workflow | name (monospace) |
| Status | colored dot + text (PASS/FAIL/RUNNING) |
| Duration | `4m 12s` or `--` if running |
| Started | relative time (`2h ago`) |

- Status dots: neon green/magenta/amber with faint glow
- Running rows: animated pulse on status dot
- Hover: row highlights `--bg-elevated`
- Click: navigates to run detail
- Filter dropdown: by workflow name, by status

## Route: Run Detail

### Top bar

- Breadcrumb: "Runs / #047 cross-review"
- Status badge (large, glowing)
- Re-run button (right-aligned, neon cyan)

### Metadata row

Single row: started time, duration, model, token count. Monospace values.

### Steps table

Compact table: step number, step name, model, duration, status dot. Each step on one row.

### Output section

Full rendered markdown. highlight.js CSS must be imported for syntax-colored code blocks.

### Telemetry section

Collapsible. Kibana iframe embed (400px height). Section hidden if no telemetry URL available.

## Route: Results Browser

### Layout

Resizable split pane: file tree (left, 250px default) + preview/edit pane (right, flex: 1).

### File tree (left pane)

- Collapsible directory tree with indent guides
- File type icons with color coding: `.md` (cyan), `.diff` (green), `.json` (amber), `.yaml` (magenta), `.glitch` (cyan)
- Directories: cyan icon. Files: muted icon, accent on select.
- Selected file: `--bg-elevated` background + cyan left border
- Right-click context menu: copy path, open in editor, delete
- Drag handle on divider to resize (min 200px, max 50% viewport)

### Preview/Edit pane (right)

**Header:** breadcrumb (full path, each segment clickable) + mode toggle ("Preview" | "Edit")

**Preview mode (default):**
- Markdown: rendered HTML with syntax-highlighted code blocks
- JSON: pretty-printed with syntax colors
- YAML: syntax highlighted
- Diff: red/green diff rendering
- Images: inline display
- Unknown: raw monospace text

**Edit mode:**
- CodeMirror with appropriate language mode
- Save button (cyan when dirty)
- Saves via `PUT /api/results/{path}` (new endpoint)

**Empty state:** "Select a file to preview" with muted folder icon, centered.

### Workflow Action System

Instead of hardcoded buttons, the GUI dynamically discovers actions from workflows via a new `(action ...)` metadata keyword. Any workflow can declare itself as a GUI action, and the GUI surfaces it in the appropriate context.

**Action keyword syntax:**

```scheme
(workflow
  (name "issue-to-pr-claude")
  (tags "pr" "claude")
  (action "results:create-pr")    ;; surfaces in results browser
  (par
    (arg repo :default "elastic/observability-robots")
    (arg issue :required true)))
```

**Action contexts:**
- `results:*` — actions in the results browser (e.g., `results:create-pr`, `results:analyze`, `results:summarize`)
- `workflow:post-run` — actions triggered after a run completes
- `workflow:*` — actions on the workflow list or editor (e.g., `workflow:lint`, `workflow:validate`)

**How the GUI discovers actions:**
1. `GET /api/workflows` already returns workflow metadata (tags, author, version)
2. Add `actions` field to the response — list of action strings from `(action ...)` keywords
3. GUI groups workflows by action context and surfaces them where they belong

**Action bar in results browser:**

When workflows declare `results:*` actions, an action bar appears at the top of the preview pane:

```
+------------------------------------------------------+
| Actions:  [Create PR (Claude)] [Create PR (Copilot)] |
+------------------------------------------------------+
```

Each button corresponds to a workflow with a matching action. Clicking one opens the run dialog for that workflow, pre-filled with context from the current result directory:
- `repo` param: extracted from result path
- `issue` param: extracted from directory name
- Other params: shown in the dialog for user override

The workflow does the actual work (calling `gh`, creating branches, etc.) — the GUI just launches it with context.

**Action bar in run detail:**

When workflows declare `workflow:post-run` actions, they appear as buttons after a run completes. Example: a "summarize" workflow that takes a run ID and generates a summary.

### New backend endpoints

- `PUT /api/results/{path}` — save edited file content
- `GET /api/workflows/actions/{context}` — returns workflows matching an action context (e.g., `results:create-pr`)

## Components Inventory

Reusable components to extract:

1. **StatusBadge** — dot + text, color by status (pass/fail/running)
2. **Card** — surface container with hover border effect
3. **Modal** — overlay + backdrop blur + cyan border
4. **Breadcrumb** — path segments as links
5. **FilterBar** — search + tag pills + sort dropdown
6. **SplitPane** — resizable two-panel layout with drag handle
7. **FileTree** — recursive directory tree with icons
8. **ActionBar** — dynamic action strip populated from workflow action metadata

## Testing

All 20 existing Playwright tests must continue to pass after the redesign. New tests needed:

- File tree navigation (expand/collapse directories)
- Preview/edit mode toggle
- File editing and save
- Action bar renders workflows with matching action context
- Action button opens run dialog with pre-filled params
- Sidebar nav collapses/expands
- Card grid filtering by tag
- Resizable split pane drag

## Migration

This is a visual overhaul of existing components, not a rewrite. The approach:

1. Replace `app.css` with the new design system (CSS custom properties, base styles)
2. Update each Svelte component one at a time — layout, classes, markup
3. Extract shared components (StatusBadge, Modal, etc.) into `src/lib/components/`
4. Add new features (file editing, action system) after visual overhaul
5. Run Playwright tests after each component to catch regressions
