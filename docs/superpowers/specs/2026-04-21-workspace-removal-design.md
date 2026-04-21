# Workspace Removal & Shebang/REPL Model

**Date:** 2026-04-21
**Status:** Draft

## Summary

Remove the workspace/project concept from glitch. Workflows become standalone files — run them directly via shebang or build them interactively in a REPL. No `.glitch/` directory, no project root discovery, no cwd-based workflow lookup. Replace SQLite with Datascript. Replace FTS5 search with ripgrep.

## Motivation

Glitch should feel like Emacs: lightweight, scriptable, no project ceremony. A workflow is a file on disk. You point at it, it runs. You open a REPL, you build one interactively. The workspace concept adds indirection without earning its keep.

## Design

### 1. Deletions

**Files removed entirely:**
- `bb/src/glitch/project.clj` — workspace/project root resolution
- `bb/src/glitch/gui/workspace.clj` — workspace REST API stubs
- `bb/src/glitch/mcp/indexer.clj` — SQLite FTS5 indexing
- `bb/src/glitch/mcp/search.clj` — hybrid/keyword search
- `bb/src/glitch/mcp/embeddings.clj` — embedding calls

**Commands removed:**
- `glitch init` — nothing to scaffold
- `glitch gui` — being rewritten in shadow-cljs/reagent

**Code removed from surviving files:**
- `main.clj`: `-P/--project` flag, `resolve-workflow` function, all search-dir logic in `cmd-run`, the `project/resolve` require
- `runner.clj`: `workspace-path` local, `indexer/open-search-db` and `indexer/index-repo` calls, `*search-fn*` binding, MCP tool-context wiring (workspace-path, search-db, tool-handler, available-defs), `list-workflows` function
- `store.clj`: `open-for-project` function, `workspace` column from `runs` table, entire SQLite pod backend (replaced by Datascript)
- `mcp/handlers.clj`: `confined-path` helper, `handle-search`, `handle-index`, `handle-symbols` handlers

### 2. Store Rewrite: SQLite to Datascript

Pure Clojure, no pod dependency. Single global DB persisted as EDN.

**Schema:**
```clojure
{:run/id          {:db/unique :db.unique/identity}
 :run/workflow    {}   ;; absolute path to .glitch file
 :run/input       {}
 :run/output      {}
 :run/exit        {}
 :run/model       {}
 :run/started-at  {}
 :run/finished-at {}
 :step/run-id     {}
 :step/id         {}
 :step/prompt     {}
 :step/output     {}
 :step/model      {}
 :step/duration   {}
 :step/kind       {}
 :step/confidence {}
 :fact/id         {:db/unique :db.unique/identity}
 :fact/run-id     {}
 :fact/claim      {}
 :fact/confidence {}
 :fact/status     {}
 :edge/from       {:db/valueType :db.type/ref}
 :edge/to         {:db/valueType :db.type/ref}
 :edge/rel        {}
 :edge/weight     {}}
```

**Persistence:** Single EDN file at `~/.local/share/glitch/glitch.edn`. Load on `open`, flush on `close` and after each `finish-run` for crash safety.

**Public API unchanged:** `open`, `close`, `record-run`, `finish-run`, `record-step`, `get-run`, `list-runs`, `get-steps`, `record-fact`, `record-fact-edge`, `get-facts`, `get-fact-edges`. Callers don't change.

**Removed:** `open-for-project`, `workspace` parameter from `record-run`, SQLite pod dependency from `bb.edn`.

**Run IDs:** Datascript entity IDs (integers) — same shape as SQLite rowids, callers won't notice.

### 3. CLI Changes

**`glitch run` simplified:**
```
glitch run [options] <file> [input...]

Options:
  -p, --provider <name>   Default provider for LLM calls
  -m, --model <model>     Default model name
  -s, --set <key=value>   Set parameter (repeatable)
```

First positional arg is always a file path. If it doesn't exist, error. No name-based lookup, no search directories.

`-p` overrides the default provider (currently hardcoded to `"lmstudio"` in the runner). Workflows that specify `:provider` in their `(llm ...)` calls override the CLI flag.

**Shebang support:**
```
#!/usr/bin/env glitch run
```
`chmod +x foo.glitch && ./foo.glitch` works. The shebang passes the file path as the first arg to `glitch run`, which is the existing behavior once search-dir logic is removed.

### 4. Runner Changes

- Drop `workspace-path` local — no more `System/getProperty "user.dir"` for indexing
- Drop `indexer/open-search-db` and `indexer/index-repo` calls
- Drop `*search-fn*` binding and the `search` primitive from the SCI context
- Simplify MCP tool-context wiring — no more `workspace-path` or `search-db` in context, but the runner still builds a tool-handler from surviving MCP tools (`glitch_run`, `glitch_eval`, `glitch_check`, `glitch_grep`, `glitch_read_file`, `glitch_search`, `glitch_symbols`) and injects it into `(llm ...)` calls for tool-use
- Drop the `:project` kwarg from the `run` signature
- Add `:provider` kwarg — passed from CLI `-p` flag, used as default when workflow doesn't specify `:provider` in `(llm ...)`
- `workflows-dir` stays — `call-workflow` resolves children relative to the parent workflow's directory, derived from `(.getParent (io/file workflow-path))`

**`call-workflow` unchanged** — resolves children relative to the parent workflow's directory. File-relative, not workspace-relative.

### 5. `glitch repl`

New command. Starts an nREPL server with glitch DSL preloaded.

**What it does:**
1. Starts `babashka.nrepl/start-server!` on a port (default 1667, override with `-p`)
2. Pre-loads all glitch namespaces and binds DSL primitives (`step`, `workflow`, `par`, `llm`, `sh`, etc.) into the `user` namespace
3. Wires up a default provider so `(llm :prompt "...")` works immediately
4. Writes `.nrepl-port` file in cwd (CIDER convention for auto-detection)
5. Prints the port to stderr

**Usage:**
```
glitch repl            # starts nrepl on 1667
glitch repl -p 7888    # custom port
```

**From Emacs:** `M-x cider-connect RET localhost RET 1667` — build workflows interactively, promote to `.glitch` files when ready.

### 6. MCP — Ripgrep-Powered

`glitch mcp` stays as a stdio JSON-RPC server. No daemon.

**Tools removed:**
- `glitch_index` — nothing to index

**Tools replaced:**

`glitch_search` — powered by `rg --json`:
- `query` (required) — search pattern
- `path` (required) — directory to search
- `glob` — file filter (`*.clj`, `*.{ts,tsx}`)
- `fixed` — boolean, literal string mode (`-F`)
- `multiline` — boolean, cross-line matching (`-U`)
- `pcre2` — boolean, lookaround/backreferences (`-P`)
- `context` — lines of context (`-C`)
- `limit` — max results (`-m`)
- `smart_case` — boolean, default true (`-S`)

`glitch_symbols` — curated `rg` patterns per language:
- Clojure: `^\(def[n\-]?\s+QUERY`
- Go: `^func\s+.*QUERY|^type\s+QUERY`
- Python: `^(def|class)\s+QUERY`
- JS/TS: `(function|const|class|export)\s+QUERY`

Takes `query`, `path` (required), optional `language` hint.

**Tools kept (updated):**
- `glitch_run` — `workflow` param is now a file path
- `glitch_eval` — unchanged
- `glitch_check` — drop `confined-path`, path is explicit
- `glitch_grep` — internally replaced by ripgrep, drop `confined-path`
- `glitch_read_file` — drop `confined-path`, path is explicit

### 7. Unchanged

These namespaces/commands have no workspace dependency and are untouched:
- `glitch.core` — all primitives (step, llm, sh, par, gate, validate, grounded?, consensus, composite-score, call-workflow)
- `glitch.confidence` — pure math
- `glitch.graph` — investigation graph
- `glitch.provider` — provider dispatch, tier escalation
- `glitch check`, `glitch eval`, `glitch up`, `glitch version`, `glitch plugin`
- SCI context and macro definitions in runner (minus the `search` primitive)

### 8. Dependency Changes

**bb.edn:**
- Remove: `org.babashka/go-sqlite3` pod
- Add: `datascript` (pure Clojure, no pod needed)

**Runtime requirements:**
- `rg` (ripgrep) — required for MCP search/symbols tools. Added to `glitch up` checks.
