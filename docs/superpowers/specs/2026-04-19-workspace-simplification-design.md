# Workspace Simplification: Project Marker Convention

**Goal:** Replace the workspace DSL (macro, resources, defaults, config file) with a `.glitch/` directory convention. Workflows are Janet programs — they import what they need. The framework just needs to know where the project root is.

**Scope:** janet/ directory only. No Go changes.

---

## 1. Project Discovery

The `.glitch/` directory is the project marker. No files inside it are required.

**`workspace.janet` becomes `project.janet`** with two exports:

```janet
(defn find-root [dir]
  "Walk up from dir looking for a .glitch/ directory. Returns the
   parent directory containing .glitch/, or nil."
  ...)

(defn resolve [&named flag]
  "Resolve the project root.
   Priority: explicit --project flag > GLITCH_PROJECT env var > walk cwd."
  ...)
```

Both return a directory path (the parent of `.glitch/`), not a table. `resolve` returns `nil` if no project is found.

The module-level var `*current-ws*`, the `workspace` macro, `defaults`, `resource`, `load`, `get-resource`, `resources-by-type`, and `find-workspace-file` are all deleted.

---

## 2. CLI Changes

### `glitch init` (replaces `glitch workspace init/status/resources`)

```
glitch init [name]
```

- Creates `.glitch/` in the current directory
- Prints the path
- `name` argument is ignored (kept for compatibility but unused)
- No files are generated inside `.glitch/`

The `glitch workspace` command and all subcommands (`init`, `status`, `resources`) are removed.

### `glitch run` flag rename

- `--workspace` / `-w` flag renamed to `--project` / `-P`
- Passed to `project/resolve` as `:flag`
- Model comes from `--model` flag only, hardcoded default is `"gemma4"`
- No config file fallback for model

### `glitch gui` flag rename

- `--workspace` / `-w` renamed to `--project` / `-P`
- GUI receives project root path (string), not a table

---

## 3. Workflow Path Resolution

`cmd-run` resolves workflow files by searching these directories in order:

1. `<project-root>/workflows/` (if project found)
2. `.glitch/workflows/` (relative to cwd)
3. `workflows/` (relative to cwd)

This is the same logic as today minus the workspace table indirection — it uses the project root path directly.

---

## 4. Store Scoping

`store/open-for-workspace` renamed to `store/open-for-project`. Same behavior: puts the database at `<project-root>/.glitch/glitch.db`.

`cmd-run` calls `store/open-for-project` when a project root is found, otherwise `store/open` (global db at `~/.local/share/glitch/glitch.db`).

The `workspace` column in the `runs` table stores the project root path instead of a workspace name. No schema change needed — it's already `TEXT NOT NULL DEFAULT ''`.

---

## 5. Runner Changes

The `workspace` named parameter on `runner/run` is renamed to `project`. It's still a string passed through to `store/record-run` for the `workspace` column. The runner doesn't use it for anything else.

---

## 6. GUI Changes

- `build-routes` takes `project-root` (string or nil) instead of `workspace` (table)
- `/api/workspace` endpoint renamed to `/api/project`, returns `{:root path}` or `{}`
- `/api/workspace/resources` endpoint removed

---

## 7. Test Changes

`test-workspace.janet` is renamed to `test-project.janet` and rewritten:

- Test `find-root`: create a tmp dir with `.glitch/` inside, verify discovery
- Test `find-root` walking up: create nested dirs, verify it finds the right root
- Test `find-root` returns nil when no `.glitch/` exists
- Test `resolve` with explicit flag
- Test `resolve` with env var
- No resource/defaults/macro tests (those concepts are gone)

---

## 8. Files Changed

| File | Action |
|------|--------|
| `janet/src/glitch/workspace.janet` | Delete |
| `janet/src/glitch/project.janet` | Create (two functions) |
| `janet/src/glitch/main.janet` | Modify (replace ws/ with project/, replace commands) |
| `janet/src/glitch/gui.janet` | Modify (use project root string) |
| `janet/src/glitch/store.janet` | Modify (rename function) |
| `janet/src/glitch/runner.janet` | Modify (rename param) |
| `janet/test/test-workspace.janet` | Delete |
| `janet/test/test-project.janet` | Create |

---

## 9. What's Not Changing

- Store schema (same columns, same db format)
- Plugin system (`glitch plugin`)
- Provider loading
- Workflow execution (runner, core, steps)
- The `.glitch/` directory structure beyond the marker convention

---

## 10. Migration

No migration needed. This is pre-1.0 — existing `.glitch/workspace.janet` files just become unused. Users run `glitch init` in any project directory to start fresh.
