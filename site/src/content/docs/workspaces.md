---
title: "Workspaces"
order: 8
description: "A workspace is a self-contained project environment: declared resources, scoped workflows, nested runs, and shared defau"
---

A workspace is a self-contained project environment: declared resources, scoped workflows, nested runs, and shared defaults. You run `glitch workspace init` once, then every subsequent command — `glitch run`, `glitch observe`, `glitch workspace gui` — knows where it is and what it's working on.

## Quick start

```bash
# Scaffold a new workspace in the current directory
glitch workspace init ~/Projects/my-project

# Add a git repo as a declared resource (clone + pin to a SHA)
glitch workspace add https://github.com/acme/backend --pin main

# Add a local folder as a symlinked resource
glitch workspace add ~/notes --as notes

# Add a tracker alias (no filesystem materialization)
glitch workspace add acme/ops --as ops-tracker

# Make this workspace active so every command auto-resolves it
glitch workspace use my-project

# Check what's active, pinned, and recently run
glitch workspace status
```

## What a workspace gives you

- **Declared resources** — pinned git clones, symlinked folders, and tracker aliases listed inline in `workspace.glitch`. Workflows reference them by name instead of absolute paths.
- **Scoped workflows** — workflows in `<workspace>/workflows/` override the global set. No more `--path` on every invocation.
- **Shared defaults** — one place to set the default model, provider, Elasticsearch URL, and params for every workflow in the workspace.
- **Nested runs** — batch fan-out and `call-workflow` produce parent/child run trees, rendered as a tree in the GUI.
- **Auto-discovery** — walk-up from the current directory, `GLITCH_WORKSPACE` env var, or a sticky active workspace. No flag needed in day-to-day use.
- **A reshaped CLI** — top-level `glitch run`, everything else under `glitch workspace`.

## The workspace.glitch file

The format is s-expressions, same as workflows. A full example with all resource types and defaults:

```glitch
(workspace "my-project"
  :description "Platform command center"
  :owner "your-name"

  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    (params
      :repo "acme/backend"
      :env "staging"))

  (resource "backend"
    :type "git"
    :url "https://github.com/acme/backend"
    :ref "main"
    :pin "abc123def456...")

  (resource "frontend"
    :type "git"
    :url "https://github.com/acme/frontend"
    :ref "v9.0.0"
    :pin "f00ba4...")

  (resource "notes"
    :type "local"
    :path "~/notes")

  (resource "ops-tracker"
    :type "tracker"
    :repo "acme/ops"))
```

A minimal workspace is just a name:

```glitch
(workspace "scratch")
```

### Top-level fields

| Field | Description |
|-------|-------------|
| `"<name>"` | Workspace name, used in the registry and run logs |
| `:description` | What this workspace is for |
| `:owner` | Your name or team |
| `(defaults ...)` | Workspace-wide defaults for model, provider, Elasticsearch, and params |
| `(resource "<name>" ...)` | One per declared resource; repeat for each |

### Defaults

| Field | Description |
|-------|-------------|
| `:model` | Default model for `(llm ...)` steps |
| `:provider` | Default provider (`ollama`, `copilot`, `claude`, etc.) |
| `:elasticsearch` | Default Elasticsearch URL for search/index/observe |
| `:websearch` | SearXNG instance URL (e.g. `"http://localhost:8080"`). Required by the `(websearch ...)` workflow form. |
| `(params :<key> "<value>")` | Key-value pairs injected as `~param.<key>` in every workflow |

### Resource types

| Type | Required fields | Tool-managed | Materialization |
|------|-----------------|--------------|-----------------|
| `git` | `:url`, `:ref` | `:pin` | Clone under `<workspace>/resources/<name>/`, pinned to a SHA |
| `local` | `:path` | — | Symlink at `<workspace>/resources/<name>` pointing to the expanded path |
| `tracker` | `:repo` (`org/name`) | — | Named alias only; no filesystem entry |

Local resources ship symlink-only. If you need a snapshot of a non-git folder, promote it to a git repo first. Trackers are useful when you want to reference a GitHub repo by name in a workflow but don't need its code checked out.

## Referencing resources from workflows

Resources are exposed to workflows via `~resource.<name>.<field>` template bindings:

```glitch
(step "backend-status"
  (run "cd ~resource.backend.path && git status --short"))

(step "frontend-head"
  (run "git -C ~resource.frontend.path rev-parse HEAD"))

(step "tracker-info"
  (run "gh repo view ~resource.ops-tracker.repo"))
```

Available fields per type:

| Expression | git | local | tracker |
|------------|-----|-------|---------|
| `~resource.<name>.path` | clone path | symlink path | empty |
| `~resource.<name>.url` | git URL | empty | empty |
| `~resource.<name>.ref` | declared ref | empty | empty |
| `~resource.<name>.pin` | resolved SHA | empty | empty |
| `~resource.<name>.repo` | inferred from URL if GitHub | empty | declared `:repo` |

Missing resources evaluate to empty string, matching `~param.missing` behavior. In global mode (no active workspace), `~resource.*` is always empty.

### Cross-resource iteration with map-resources

Loop over resources of a given type — parallel to `map`, but pulled from workspace state rather than a prior step:

```glitch
(map-resources :type "git"
  (step "git-status"
    (run "cd ~resource.item.path && git status --short")))
```

Each iteration binds the current resource to `.resource.item`, so `~resource.item.path`, `~resource.item.ref`, and the rest are all available in the body.

## The CLI surface

Three tiers: hot paths at top level, management under `glitch workspace`, infrastructure at top level.

### Hot paths

```bash
glitch run <workflow> [input]   # run a workflow (workspace workflows first, global fallback)
glitch observe <query>          # observer queries on indexed activity
```

`glitch run` accepts `--set key=value`, `--path` (override workflow resolution), and `--results-dir`. In workspace mode it resolves from `<workspace>/workflows/`; in global mode from `~/.config/glitch/workflows/`.

### Workspace management

| Command | Purpose |
|---------|---------|
| `glitch workspace init [path]` | Scaffold a workspace; adds it to the registry |
| `glitch workspace use <name>` | Set the active workspace (sticky across shell sessions) |
| `glitch workspace list` | List registered workspaces |
| `glitch workspace status` | Active workspace, resources with staleness, recent runs |
| `glitch workspace gui` | Launch the workspace GUI |
| `glitch workspace register <path>` | Add an existing workspace dir to the registry |
| `glitch workspace unregister <name>` | Remove from the registry (files untouched) |
| `glitch workspace add <url\|path> [--as name] [--pin ref]` | Add a resource; type inferred from input |
| `glitch workspace rm <name>` | Remove a resource; deletes the clone or symlink |
| `glitch workspace sync [name...]` | Update resources to their declared refs |
| `glitch workspace pin <name> <ref>` | Set a new ref, sync, and write the pin |
| `glitch workspace workflow list` | List workflows in the active workspace |
| `glitch workspace workflow new <name>` | Scaffold a new workflow file |
| `glitch workspace workflow edit <name>` | Open a workflow in `$EDITOR` |

### Infrastructure

```bash
glitch up          # start Elasticsearch + Kibana via Docker Compose
glitch down        # stop them
glitch index       # index a repo for code search
glitch config      # manage global config
glitch plugin      # manage plugins
```

## Workspace discovery

When you run any command, `glitch` picks a workspace using this precedence (first match wins):

1. `--workspace <path>` — explicit flag override
2. `GLITCH_WORKSPACE` env var — useful for CI and scripts
3. CWD walk-up — walk up from `$PWD` looking for a directory containing `workspace.glitch`
4. Active workspace — whatever you last `glitch workspace use`'d, stored in `~/.config/glitch/state.glitch`
5. **None** — global mode: global workflows, no resources, CWD-relative results

In practice: run `glitch workspace use my-project` once, and every command from any directory uses that workspace until you switch. `cd` into a workspace directory and the walk-up wins automatically, even if a different workspace is active.

## Nested runs

Runs form a tree. Two things create children:

### 1. Batch fan-out

`glitch batch <config>` creates a parent batch run, and each `{issue × variant × iteration}` becomes a child. The on-disk layout mirrors the tree:

```
results/acme/backend/morning-briefing-r8k2/
├── README.md
├── evidence/
├── run.json
└── children/
    ├── pr-review-a1b2/
    │   ├── README.md
    │   └── evidence/
    └── pr-review-c3d4/
        └── ...
```

### 2. call-workflow

Sub-workflow invocation via the `call-workflow` s-expr form. The runner starts a new run as a child of the current run, runs the named workflow to completion, and returns the child's final output as this step's output:

```glitch
(step "list-prs"
  (run "gh pr list --repo acme/backend --json number | jq -r '.[].number'"))

(map "list-prs"
  (step "review-each-pr"
    (call-workflow "pr-review"
      :set (repo "acme/backend")
      :set (pr "~param.item"))))

(step "summary"
  (llm :prompt ```
    Summarize these PR reviews:
    ~(step review-each-pr)
    ```))
```

`call-workflow` inherits the active workspace and resources. Cycle protection refuses to start a child whose name is already on the workflow call stack.

Retries, catch, and map do **not** create child runs — those are timeline events or inline step expansion on the same run. Nesting reflects workflow composition only, so the GUI tree stays readable.

## Workspace directory layout

```
~/Projects/my-project/
├── workspace.glitch           # declarative config: name, defaults, resources, pins
├── workflows/                 # workflows scoped to this workspace
├── resources/                 # materialized resources (clones, symlinks)
│   ├── backend/               # git clone, pinned SHA in workspace.glitch
│   ├── frontend/              # git clone
│   └── notes/                 # symlink to ~/notes
├── results/                   # run outputs, organized by GitHub namespace
│   └── acme/backend/
│       └── issue-42/
│           ├── README.md
│           ├── evidence/
│           ├── run.json
│           └── children/      # nested sub-workflow runs
│               └── pr-review-a1b2/
│                   └── ...
└── .glitch/
    ├── glitch.db              # local run database
    ├── resources.glitch       # per-resource fetched timestamps (not committed)
    └── .gitignore             # auto-generated
```

Workflows resolve from `<workspace>/workflows/` only — the global workflow set is skipped in workspace mode. If you want a global workflow available locally, copy or symlink it in. Explicit beats magical.

## Default params

Params declared in `(defaults (params ...))` become template variables in every workflow:

```glitch
(workspace "acme"
  (defaults
    (params
      :repo "acme/backend"
      :team "platform")))
```

Any workflow can now use `~param.repo` and `~param.team` without `--set` flags. Explicit `--set` on the command line overrides workspace defaults when both are present.

## The workspace registry

Known workspaces live in `~/.config/glitch/workspaces.glitch`:

```glitch
(workspaces
  (workspace "my-project" :path "~/Projects/my-project")
  (workspace "sandbox"    :path "~/Projects/sandbox")
  (workspace "labs"       :path "~/Projects/labs"))
```

Populated automatically by `workspace init` and `workspace register`; hand-editable; name collisions are rejected. The active workspace is tracked separately in `~/.config/glitch/state.glitch`:

```glitch
(state :active-workspace "my-project")
```

## Example: running from any directory

```bash
# Once, from anywhere:
glitch workspace use my-project

# Now every command picks up that workspace's workflows, resources, and defaults:
glitch run morning-briefing
glitch run pr-review --set pr=42

# Or cd into a different workspace — walk-up wins over the active one:
cd ~/Projects/sandbox
glitch run build-release
```

## Example: daily briefing using resources

```glitch
(workflow "morning-briefing"
  :description "Aggregate open PRs across all pinned git resources"

  (map-resources :type "git"
    (step "prs"
      (run "gh pr list --repo ~resource.item.repo --json number,title,state | jq '.'")))

  (step "briefing"
    (llm :prompt ```
      Summarize open PRs across my repos as a terse bullet list:
      ~(step prs)
      ```)))
```

No hardcoded repo list. Add or remove resources in `workspace.glitch` and the briefing picks up the change the next time it runs.

## Global config is s-expr too

`~/.config/glitch/config.glitch` holds global provider and tier configuration:

```glitch
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

One format across global config, workspace config, and workflows.

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the step forms and templates your workspace workflows use
- [Plugins](/docs/plugins) — reusable data-gathering subcommands that compose with workflows
- [Local Models](/docs/local-models) — setting up Ollama for your workspace's default model
- [Compare](/docs/compare) — parent/child runs across multiple providers