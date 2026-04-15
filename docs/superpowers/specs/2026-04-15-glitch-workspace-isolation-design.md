# Glitch Workspace Isolation

**Date:** 2026-04-15
**Status:** Draft

## Problem

Glitch workspaces scope workflows and results to a directory, but nothing enforces isolation between concurrent runs or between workspaces. Today:

1. **SQLite is global.** Every run writes to `~/.local/share/glitch/glitch.db` regardless of which workspace invoked it. Two parallel batch runs against different workspaces interleave rows in the same tables with no workspace tag. The GUI can't filter runs by workspace because the data doesn't carry that context.

2. **Results overwrite silently.** Two concurrent runs targeting the same `<org>/<repo>/<issue>-<number>` directory race on `os.WriteFile` — the last writer wins. There's no run-scoped subdirectory and no advisory lock. A batch run and a manual `glitch ask` hitting the same issue will corrupt each other's evidence directory.

3. **Config bleeds across workspaces.** `workspace.glitch` defines `(defaults :model ... :provider ...)`, but the global `~/.config/glitch/config.yaml` is always loaded first. If two workspaces want different default models, the user must remember to set the flag or hope the workspace file was parsed — there's no clear precedence chain or validation that the workspace config actually took effect.

4. **No workspace identity in telemetry.** `run.json` has no `workspace` field. Elasticsearch-indexed runs can't be grouped or filtered by workspace. Cross-workspace dashboards are impossible.

## Solution

Workspace-scoped state at every layer — database, filesystem, config, and telemetry — with run-level isolation for concurrent safety.

## Design

### 1. Per-workspace SQLite database

Move the database into the workspace directory when `--workspace` is set:

```
<workspace>/.glitch/glitch.db
```

Without `--workspace`: behavior unchanged — global `~/.local/share/glitch/glitch.db`.

The `.glitch/` directory inside the workspace holds all glitch-managed state. It should be gitignored by default (glitch init can append it).

```go
// store.OpenForWorkspace returns a workspace-scoped store.
func OpenForWorkspace(workspacePath string) (*Store, error) {
    dbPath := filepath.Join(workspacePath, ".glitch", "glitch.db")
    return OpenAt(dbPath)
}
```

SQLite WAL mode (already enabled) handles concurrent readers. For concurrent writers within the same workspace, SQLite's built-in busy timeout is sufficient — set `_pragma=busy_timeout(5000)` alongside the existing WAL pragma.

**What this gives us:** Each workspace's run history, research hints, and step logs are self-contained. Deleting a workspace directory deletes all its state. The GUI, when launched with `--workspace`, queries only that workspace's database.

### 2. Run-scoped result directories

Every run gets a unique subdirectory under the issue/PR path:

```
results/<org>/<repo>/<issue|pr>-<number>/<run-id>/
├── README.md
├── run.json
├── summary.md
├── evidence/
│   ├── 001-grep_code.txt
│   └── 002-read_file.txt
└── implementation/
    └── plan.md
```

The `<run-id>` is a short timestamp-plus-suffix: `20260415T103000-a1b2`. Format: `YYYYMMDDTHHMMSS-<4 hex>`. Human-sortable, unique enough for concurrent runs, and filesystem-safe.

```go
func runScopedDir(baseDir string, result LoopResult) string {
    base := resultDir(baseDir, result)
    stamp := time.Now().UTC().Format("20060102T150405")
    suffix := fmt.Sprintf("%04x", rand.Intn(0xFFFF))
    return filepath.Join(base, stamp+"-"+suffix)
}
```

A symlink `latest` in the parent directory points to the most recent run:

```
results/elastic/observability-robots/issue-3920/
├── latest -> 20260415T103000-a1b2
├── 20260415T103000-a1b2/
└── 20260415T091500-c3d4/
```

This preserves the "latest result is at a predictable path" property that downstream workflows depend on, while keeping history.

**Variant runs** (the `--claude` / `--copilot` double-dash convention from the workspace model spec) are orthogonal — the variant suffix goes on the parent, not the run directory:

```
results/elastic/observability-robots/issue-3920--claude/latest/
results/elastic/observability-robots/issue-3920--copilot/latest/
```

### 3. Config precedence chain

Explicit, documented, no surprises:

```
1. CLI flags              (highest — --model, --provider)
2. workspace.glitch       (defaults block)
3. ~/.config/glitch/config.yaml
4. Built-in defaults      (lowest — qwen2.5:7b / ollama)
```

Implementation: the existing `config.Load()` returns the merged config. Add a `config.ApplyWorkspace(ws *workspace.Workspace, cfg *Config)` that overlays workspace defaults onto the loaded config before CLI flags are applied.

```go
func ApplyWorkspace(ws *Workspace, cfg *Config) {
    if ws == nil {
        return
    }
    if ws.Defaults.Model != "" {
        cfg.Model = ws.Defaults.Model
    }
    if ws.Defaults.Provider != "" {
        cfg.Provider = ws.Defaults.Provider
    }
}
```

The call site in `cmd/root.go` applies in order: load config → apply workspace → apply flags.

### 4. Workspace identity in run metadata

Add `workspace` to `run.json`:

```json
{
  "workspace": "stokagent",
  "workspace_path": "/Users/stokes/Projects/stokagent",
  "run_id": "20260415T103000-a1b2",
  "repo": "elastic/observability-robots",
  "ref_type": "issue",
  "ref_number": 3920,
  ...
}
```

The `workspace` field is the name from `workspace.glitch`; `workspace_path` is the resolved `--workspace` flag value. Both are empty strings when running without a workspace.

The SQLite `runs` table gets a `workspace TEXT NOT NULL DEFAULT ''` column. The GUI filters on it when `--workspace` is set. Elasticsearch-indexed runs can be grouped by workspace in Kibana.

### 5. `.glitch/` directory convention

Every workspace gets a `.glitch/` directory for managed state:

```
<workspace>/
├── .glitch/
│   ├── glitch.db          # workspace-scoped SQLite
│   └── .gitignore         # contains "*" — never commit state
├── workflows/
├── results/
└── workspace.glitch
```

`glitch init` (or first run with `--workspace`) creates `.glitch/` and writes the inner `.gitignore`. This is the only glitch-managed directory — `workflows/` and `results/` remain user-visible and committable.

## What changes in glitch core

1. **`internal/store`:** new `OpenForWorkspace(path)` constructor; `busy_timeout` pragma added to connection string
2. **`internal/research/results.go`:** `resultDir()` → `runScopedDir()` with timestamp-hex suffix; `latest` symlink management
3. **`cmd/root.go`:** config load order enforced as flags → workspace → global → defaults
4. **`internal/research/results.go`:** `runJSON` struct gains `Workspace` and `WorkspacePath` fields
5. **`internal/store/schema.go`:** `workspace` column on `runs` table
6. **`cmd/root.go`:** `.glitch/` directory auto-creation when `--workspace` is set

## What does NOT change

- Behavior without `--workspace` — global database, CWD-relative results, global config only
- Workflow file format or s-expression syntax
- Pipeline execution, step runners, provider protocol
- The `workspace.glitch` parser — it already handles the fields we need
- Observer, indexer, or capability system
- GUI server code beyond filtering on the workspace column
