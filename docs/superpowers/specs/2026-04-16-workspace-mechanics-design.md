# Workspace Mechanics: Resources, Nesting, and CLI

**Date:** 2026-04-16
**Status:** Draft
**Supersedes (partially):** `2026-04-14-workspace-model-design.md` — that spec established `--workspace` as a path-only flag. This one extends the workspace into a full environment with declared resources, sub-workflow composition, parent/child run nesting, auto-discovery, and a reshaped CLI.

## Problem

The workspace is a thin scoping wrapper today: `--workspace <path>` changes where workflows and results resolve, and `workspace.glitch` holds a minimal name/defaults/repos-as-strings declaration. That isn't enough to reach the daily-driver goal:

1. **No resource model.** Workflows reach external repos by passing `--path` / `--repo` every invocation. There's no way to declare "this workspace operates on ensemble, kibana, and observability-robots" with reproducible pins.
2. **Flat runs only.** Runs are single-row in SQLite. Batch fan-out is filesystem-nested but not database-nested; the GUI can't render a run tree because there isn't one.
3. **No workflow composition.** Workflows can't call other workflows, so "daily briefing that runs pr-review for each pending PR" has to be inlined or shelled out.
4. **Path-ceremony CLI.** Every command needs `--workspace` or a shell alias. No auto-detection from CWD. No registry of known workspaces. GUI lives under `workflow gui` instead of being workspace-scoped.
5. **CLI surface leaks.** `workflow` is a top-level namespace even though workflows are workspace-scoped, making the hierarchy inconsistent.

## Solution

Promote the workspace from a path flag into a first-class, self-contained environment with four pieces:

- A **declared resource model** (git repos, local folders, tracker aliases) that materializes into `resources/` and is pinned inline in `workspace.glitch`.
- A **nested run model** (parent/child in SQLite, `children/` on disk) driven by batch fan-out and a new `call-workflow` s-expr form.
- **Workspace discovery** via CWD walk-up + a s-expr registry + active-workspace state, with explicit precedence ordering.
- A **reshaped CLI** that folds workflow management under `workspace` while promoting workflow invocation to a top-level `glitch run`.

All configuration stays s-expr. The existing YAML global config (`~/.config/glitch/config.yaml`) flips to `config.glitch` as a clean break — no migration code, per the pre-1.0 rule.

## Design

### 1. Filesystem layout

A workspace is a directory containing a `workspace.glitch` file at its root:

```
~/Projects/stokagent/
├── workspace.glitch           # declarative config: name, defaults, resources, pins
├── workflows/                 # workflows scoped to this workspace
├── resources/                 # materialized resources (clones, symlinks)
│   ├── ensemble/              # git clone, pinned SHA in workspace.glitch
│   ├── kibana/                # git clone
│   └── notes/                 # symlink to ~/my-notes (local folder)
├── results/                   # run outputs, organized by GitHub namespace
│   └── elastic/ensemble/
│       └── issue-3920/
│           ├── README.md      # rollup
│           ├── evidence/
│           ├── run.json
│           └── children/      # nested sub-workflow runs
│               └── pr-review-a1b2/
│                   ├── README.md
│                   └── evidence/
└── .glitch/
    ├── glitch.db              # SQLite: runs, steps, parent/child links
    ├── resources.glitch       # per-resource fetched timestamps (workspace-local state)
    ├── .gitignore             # auto-generated, prevents committing state
    └── cache/                 # git object cache, transient
```

`resources/` and `results/children/` are new. `workflows/`, `results/<org>/<repo>/<ref>/`, and `.glitch/glitch.db` already exist.

### 2. `workspace.glitch` schema

Extends today's form with a `(resource ...)` block per declared resource. User-edited fields (`:ref`, `:url`, `:path`) and tool-managed fields (`:pin`) live side by side; per-sync timestamps move to `.glitch/state.glitch` to keep `workspace.glitch` committable without churn.

```clojure
(workspace "stokagent"
  :description "Agent command center"
  :owner "adam-stokes"

  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200")

  (resource "ensemble"
    :type "git"
    :url "https://github.com/elastic/ensemble"
    :ref "main"
    :pin "abc123def456...")

  (resource "kibana"
    :type "git"
    :url "https://github.com/elastic/kibana"
    :ref "v9.0.0"
    :pin "f00ba4...")

  (resource "notes"
    :type "local"
    :path "~/my-notes")

  (resource "obs-robots"
    :type "tracker"
    :repo "elastic/observability-robots"))
```

**Resource types (MVP):**

| Type | Fields | Materializes as |
|------|--------|-----------------|
| `git` | `:url` (required), `:ref` (required), `:pin` (tool-managed) | Clone under `resources/<name>/` |
| `local` | `:path` (required) | Symlink at `resources/<name>` pointing to expanded `:path` |
| `tracker` | `:repo` (required, `org/name` form) | No filesystem entry; named alias only |

Local-folder resources ship symlink-only in MVP. If snapshotting a non-git folder is needed later, promote it to a git repo. No `:mode` knob.

Per-resource state file (internal, never committed; separate from the global `~/.config/glitch/state.glitch` covered in §5):

```clojure
;; <workspace>/.glitch/resources.glitch
(resources
  (resource-state "ensemble" :fetched "2026-04-15T10:30:00Z")
  (resource-state "kibana"   :fetched "2026-04-15T10:30:00Z"))
```

### 3. Resource semantics

**Materialization** runs on `glitch workspace add`:

- `git`: clone `:url`, checkout `:ref`, resolve to SHA, write `:pin` back into `workspace.glitch`, record `:fetched` in `.glitch/resources.glitch`.
- `local`: create symlink at `resources/<name>` pointing to expanded `:path`.
- `tracker`: no filesystem change; entry added to `workspace.glitch` only.

**Sync** (`glitch workspace sync [name...]`) updates resources to their current `:ref`:

- `git`: `git fetch`, checkout `:ref`, resolve SHA, update `:pin`, bump `:fetched` in `resources.glitch`. If the clone directory is missing, re-clone from scratch.
- `local`: verify symlink target exists; if the symlink itself is missing, recreate it from `:path`; if the target is gone, warn.
- `tracker`: no-op (optionally probe `gh repo view` for 404 detection).

`--force` re-clones from scratch. Named args sync a subset; no args syncs all.

**Pin** (`glitch workspace pin <name> <ref>`) rewrites `:ref`, immediately syncs, writes `:pin`. Shorthand for the common case.

**Template references** — resources are exposed to workflows via a new `.resource` root in Go text/template:

| Expression | git | local | tracker |
|------------|-----|-------|---------|
| `{{.resource.<name>.path}}` | clone path | symlink path | empty |
| `{{.resource.<name>.url}}` | git URL | empty | empty |
| `{{.resource.<name>.ref}}` | declared ref | empty | empty |
| `{{.resource.<name>.pin}}` | resolved SHA | empty | empty |
| `{{.resource.<name>.repo}}` | inferred from URL if GitHub | empty | declared `:repo` |

Missing resources evaluate to empty string (matches existing `.param.missing` behavior). In global mode (no active workspace), `.resource.*` is always empty.

**Cross-resource iteration** — new s-expr form `map-resources`, parallel to `map`, filters by type and binds each resource to `.resource.item` for the body:

```clojure
(map-resources :type "git"
  (step "status"
    (run "cd {{.resource.item.path}} && git status --short")))
```

### 4. Nested runs

**SQLite schema change.** Two columns added to `runs`:

```sql
ALTER TABLE runs ADD COLUMN parent_run_id TEXT REFERENCES runs(id);
ALTER TABLE runs ADD COLUMN workflow_name TEXT;
```

Existing `.glitch/glitch.db` is wiped on upgrade (clean break, no migration helper).

**Sources of nesting (MVP only populates from these two):**

1. **Batch fan-out.** `glitch batch <config>` creates a parent batch run; each `{issue × variant × iteration}` becomes a child with `parent_run_id` set. The current batch filesystem layout (`iteration-N/variant/`) is replaced by `children/<variant>-<iteration>-<run-id>/` so on-disk structure matches the DB tree. Old batch results under the previous layout are not migrated (pre-1.0 clean break).

2. **Sub-workflow invocation** via new `call-workflow` form:

```clojure
(step "review-each-pr"
  (call-workflow "pr-review" :set (repo "elastic/ensemble") (pr "{{step \"list-prs\"}}")))
```

The runner starts a new run with `parent_run_id = <current-run-id>`, runs the named workflow to completion inheriting the active workspace + resources, and returns the child's final step output as this step's output. `{{step "review-each-pr"}}` reads the child's result like any other step.

**Filesystem reflects the DB:**

```
results/elastic/ensemble/morning-briefing-r8k2/
├── README.md
├── evidence/
├── run.json
└── children/
    ├── pr-review-a1b2/
    │   ├── README.md
    │   ├── evidence/
    │   └── run.json
    └── pr-review-c3d4/
        └── ...
```

**Cycle protection.** The runner maintains a workflow-name call stack; `call-workflow` refuses to start a child whose name is already on the stack. Defensive, one check.

**Not sources of nesting** (stay single-run, by design):

- `retry N` — retries are timeline events on one run
- `catch` — fallback is a step on the same run
- `map` — body expands inline as steps on the parent run

Nesting reflects workflow composition only, not every control-flow construct. Otherwise the GUI tree becomes noise.

### 5. Workspace discovery

Precedence order (first match wins):

1. `--workspace <path>` — explicit override
2. `GLITCH_WORKSPACE` env var — for CI/scripts
3. CWD walk-up — walk up from `$PWD` until a directory containing `workspace.glitch` is found
4. Active workspace from `glitch workspace use` — sticky via `~/.config/glitch/state.glitch`
5. **None** — global mode: global workflows, no resources, CWD-relative results (current fallback)

**Registry** at `~/.config/glitch/workspaces.glitch` tracks known workspaces:

```clojure
(workspaces
  (workspace "stokagent"  :path "~/Projects/stokagent")
  (workspace "gl1tch-dev" :path "~/Projects/gl1tch")
  (workspace "farm"       :path "~/Projects/farm"))
```

Auto-populated by `workspace init` and `workspace register`; hand-editable; name collisions rejected.

**Active workspace state** at `~/.config/glitch/state.glitch`:

```clojure
(state :active-workspace "stokagent")
```

Set by `glitch workspace use <name>`; machine-local.

**Workflow resolution in workspace mode:** only `<workspace>/workflows/`; global `~/.config/glitch/workflows/` skipped. No merging — if a user wants a global workflow available locally, they copy or symlink it into the workspace. Explicit beats magical.

### 6. CLI surface

Three-tier layout: hot paths at top level, workspace-scoped management under `workspace`, infrastructure at top level.

**Hot paths (workspace-aware, top-level):**

```
glitch ask <query>              # smart router
glitch run <workflow> [input]   # run a workflow (replaces `workflow run`)
glitch observe <query>          # observer queries
```

`glitch run` accepts the same flags as today's `workflow run`: `--set k=v`, `--path`, `--results-dir`. In workspace mode it resolves from `<workspace>/workflows/`; in global mode from `~/.config/glitch/workflows/`.

**Workspace management (grouped under `workspace`):**

| Command | Purpose |
|---------|---------|
| `glitch workspace init [path]` | Scaffold new workspace at path (default CWD); adds to registry |
| `glitch workspace use <name>` | Set active workspace (sticky) |
| `glitch workspace list` | List registered workspaces |
| `glitch workspace status` | Active workspace + resources (with staleness) + recent runs |
| `glitch workspace gui` | Launch GUI (replaces `glitch workflow gui`) |
| `glitch workspace register <path> [--as name]` | Add existing workspace dir to registry |
| `glitch workspace unregister <name>` | Remove from registry (files untouched) |
| `glitch workspace add <url\|path> [--as name] [--pin ref]` | Add a resource; type inferred from input |
| `glitch workspace rm <name>` | Remove a resource; deletes clone/symlink |
| `glitch workspace sync [name...]` | Update resources to their declared refs |
| `glitch workspace pin <name> <ref>` | Set new ref, sync, write pin |
| `glitch workspace workflow list` | List workflows in active workspace |
| `glitch workspace workflow new <name>` | Scaffold a new workflow file |
| `glitch workspace workflow edit <name>` | Open workflow in `$EDITOR` |

**Infrastructure (top-level, not workspace-scoped):**

```
glitch up / down
glitch index [path]
glitch config show / set
glitch plugin list
glitch version
```

**Removed commands:**

- `glitch workflow list | run | new` — folded into `workspace workflow` subtree and `run`
- `glitch workflow gui` — replaced by `glitch workspace gui`

### 7. Global config flip

`~/.config/glitch/config.yaml` → `~/.config/glitch/config.glitch`:

```clojure
(config
  :default-model "qwen2.5:7b"
  :default-provider "ollama"
  :eval-threshold 4

  (providers
    (provider "openrouter"
      :type "openai-compatible"
      :base-url "https://openrouter.ai/api/v1"
      :api-key-env "OPENROUTER_API_KEY"
      :default-model "google/gemma-3-12b-it:free")

    (provider "claude" :type "cli"))

  (tiers
    (tier :providers ("ollama") :model "qwen2.5:7b")
    (tier :providers ("openrouter") :model "google/gemma-3-12b-it:free")
    (tier :providers ("copilot" "claude"))))
```

Clean break. If the old `config.yaml` exists and `config.glitch` does not, glitch logs a one-time stderr warning pointing to the new format and continues with defaults. User re-runs `glitch config set ...` or edits `config.glitch` directly.

### 8. GUI changes

Grows the existing SvelteKit GUI outward; no rewrite.

**New endpoints:**

| Endpoint | Purpose |
|----------|---------|
| `GET /api/workspaces` | List registered workspaces |
| `POST /api/workspaces/use` | Switch active workspace |
| `GET /api/workspace/resources` | List resources with staleness |
| `POST /api/workspace/resources` | Add a resource |
| `DELETE /api/workspace/resources/:name` | Remove a resource |
| `POST /api/workspace/sync[/:name]` | Sync all or one |
| `POST /api/workspace/pin` | Pin a resource to a ref |
| `GET /api/runs/:id/tree` | Recursive run tree |
| `GET /api/runs?parent_id=:id` | Direct children of a run |

**New UI surfaces:**

- **Workspace switcher** — persistent dropdown in top bar with active name and registered list; "Switch" invokes `workspace use` server-side
- **Dashboard** — active-workspace landing view with resources (staleness badges), recent runs, pending sync actions
- **Resources panel** — table with type/ref/pin/fetched; row actions for sync/pin/remove; "Add resource" modal
- **Runs tree view** — new default for the runs page; collapsible rows with children (batch variants + `call-workflow` descendants); existing flat list kept as secondary tab
- **Results browser** — file-tree shape mirroring `results/<org>/<repo>/<ref>/children/...`, rollup preview in right pane
- **Settings** (existing `Settings.svelte`) — extended with resources editing; existing workspace config editor retained

**GUI scope-outs (MVP):** no workflow graphical editor, no SSE/websocket log streaming, no cross-workspace run comparison.

### 9. Package landing zones

| Package | Responsibilities added |
|---------|------------------------|
| `internal/workspace/` | `Resource` struct + type enum, walk-up + registry resolution, active-workspace state, `workspace.glitch` round-trip writer |
| `internal/resource/` (new) | Git clone/fetch, symlink creation, pin resolution, staleness checks |
| `internal/pipeline/` | `call-workflow` step runner with parent-run threading, cycle guard, `.resource.*` template binding, `map-resources` form |
| `internal/sexpr/` | Parse `(resource ...)`, `(call-workflow ...)`, `(map-resources ...)` forms; formatting-preserving writer for `workspace.glitch` pin updates |
| `internal/store/` | Schema columns `parent_run_id`, `workflow_name`; `GetRunTree()`, children queries |
| `internal/gui/` + `gui/src/routes/` | New endpoints + tree view + resources panel + workspace switcher |
| `cmd/workspace*.go` | New command files per verb; removal of `cmd/workflow_gui.go` |
| `cmd/run.go` (new) | Top-level `run` command extracted from `workflow run` |

### 10. Rollout phases

Dependency spine. The implementation plan will expand each into ordered steps; breaking ordering is safe within a phase.

1. **Schema + discovery foundation** — `workspace.glitch` resource form in sexpr parser, walk-up/registry/env-var/`use` precedence, `config.yaml` → `config.glitch` flip, SQLite columns added. No user-visible commands yet.
2. **Resource materialization** — `workspace add/rm/sync/pin` commands, git clone cache, symlink creation, tracker aliases, plus `.resource.<name>.<field>` template binding in the pipeline runner. Shippable alone: resources become usable in existing workflows.
3. **Workflow composition** — `call-workflow` form, cycle protection, batch populates parent/child. Nested runs exist in DB and filesystem.
4. **CLI reshape** — `workspace workflow` subtree, top-level `glitch run`, removal of old commands. Pure rename; no new behavior.
5. **GUI** — workspace switcher, dashboard, resources panel, runs tree view, results browser with children expansion.
6. **`map-resources`** — ships last. Shell loops over `resources/*` are the workaround until it lands.

### 11. Breaking changes (pre-1.0, clean breaks only)

| Change | User-visible impact |
|--------|---------------------|
| `config.yaml` → `config.glitch` | One-time stderr warning on first run if old file exists; user re-inits with `glitch config set ...` or edits the new file. No migration code. |
| `workflow list \| run \| new` removed | Replaced by `glitch workspace workflow ...` + `glitch run`. Users update their scripts. |
| `workflow gui` removed | Replaced by `workspace gui`. |
| SQLite gets `parent_run_id`, `workflow_name` | `.glitch/glitch.db` wiped on version upgrade; recreated clean on first run. One-line wipe in `store.Open()`. |
| `workspace.glitch` schema extends | Additive — old workspace.glitch files still parse; resources are new forms. |

### 12. Out of scope (deferred)

Explicitly not part of this spec. Separate mini-specs as follow-ups:

- HTTP-type resources (URLs fetched on sync)
- Elasticsearch-index resources
- Resource tags/groups for filter-based iteration
- Resource post-sync hooks
- Log streaming / SSE in GUI
- Monaco or similar workflow editor embed in GUI
- Workspace export/import bundles
- Per-resource auth configuration (rely on existing `gh auth`, SSH config, env vars)
- Snapshot-copy mode for local folders (promote to git instead)
- `map-resources` can be dropped from MVP if shell loops suffice — reassess after Phase 2 lands

## Success criteria

- User runs `glitch workspace init ~/Projects/stokagent`, then `glitch workspace add https://github.com/elastic/ensemble --pin main`, then `cd ~/Projects/stokagent && glitch ask "review obs-robots#3920"` — no `--workspace` flag needed, resource is available at `{{.resource.ensemble.path}}`.
- `glitch workspace status` shows resources with pins and staleness.
- `glitch workspace gui` opens the GUI, switcher reflects active workspace, runs view renders batch and `call-workflow` trees.
- A workflow using `(call-workflow "pr-review" ...)` produces a child run visible in the tree, with its own `children/<run-id>/` directory populated.
- `glitch run my-flow` works from any CWD inside a workspace; falls back to global workflows when outside.
