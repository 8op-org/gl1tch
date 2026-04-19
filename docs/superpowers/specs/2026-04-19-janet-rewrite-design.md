# glitch-on-Janet: Full Runtime Rewrite

**Date:** 2026-04-19
**Status:** Approved design

## Problem

glitch is a workflow engine with a custom Lisp evaluator (sexpr DSL) hosted inside a Go binary. This creates friction at every layer:

- **Authoring**: Adding a new builtin, provider, or capability requires writing Go code and recompiling.
- **Runtime**: Infrastructure decisions (provider selection, tier escalation, retry policy) are hardcoded in Go, invisible to workflows.
- **Extension**: Users and plugins cannot extend glitch without the Go toolchain.

The custom evaluator (~1950 lines) reimplements what a real Lisp runtime already provides: lexical scoping, closures, threading macros, string interpolation, error handling, and concurrency. The remaining ~4350 lines of Go exist primarily to bridge the DSL to external systems (HTTP, SQLite, shell, filesystem).

## Decision

Rewrite glitch as a **Janet application**. Janet is chosen because:

- Single static binary, ~300KB-1MB, cross-compiles via C
- Instant startup
- Native C FFI (SQLite, HTTP — no shims)
- Cooperative concurrency via fibers + event loop (`ev/gather` = our `par`)
- Built-in JSON, PEG parser, module system
- The glitch DSL's semantics (threading, keyword args, fibers) map 1:1 to Janet
- Users extend glitch by writing Janet — same language as workflows, providers, plugins, workspace config

Go disappears entirely. The current Go codebase serves as a porting reference.

## Architecture

### Module Layout

```
glitch/
├── project.janet              # jpm build file
├── src/
│   ├── glitch/
│   │   ├── main.janet         # CLI entry + arg dispatch
│   │   ├── core.janet         # macros: workflow, step, par, gate, phase
│   │   ├── runner.janet       # workflow loading + execution lifecycle
│   │   ├── provider.janet     # provider registry, tier escalation
│   │   ├── store.janet        # SQLite run/step recording
│   │   ├── workspace.janet    # workspace parsing, resource binding, discovery
│   │   ├── batch.janet        # multi-variant comparison + cross-review
│   │   ├── gui.janet          # HTTP server (spork/httpf), REST API
│   │   ├── http.janet         # HTTP client helpers
│   │   └── util.janet         # JSON helpers, string ops
│   └── stdlib/
│       ├── collections.janet
│       ├── io.janet
│       └── strings.janet
├── providers/                  # provider definitions (Janet, not YAML)
│   ├── ollama.janet
│   ├── claude.janet
│   ├── copilot.janet
│   ├── gemini.janet
│   └── lmstudio.janet
├── gui/                        # Svelte frontend (unchanged)
│   └── ...
└── test/
    └── ...
```

### Core Module (~80 lines)

`workflow`, `step`, `par`, `gate`, `phase` are Janet macros/functions. There is no custom evaluator — Janet IS the evaluator.

```janet
(import glitch)

(glitch/workflow "analyze"
  :description "Analyze a repository"

  (glitch/step "files"
    (glitch/sh "find" (glitch/input) "-name" "*.go"))

  (glitch/step "analysis"
    (glitch/llm :prompt (string "Analyze:\n" (glitch/ref "files"))
                :model "qwen2.5:7b"))

  (glitch/step "save"
    (glitch/save "analysis.md" (glitch/ref "files"))))
```

- `step` records output to a fiber-local step store, calls the step-recorder callback
- `ref` reads from the step store
- `par` expands to `ev/gather` — concurrent fiber execution
- `llm` dispatches through the provider registry
- `sh` wraps `os/spawn` with stdout/stderr capture
- Users have full access to Janet: `ev/spawn`, `net/`, `peg`, `fiber`, `os/`, etc.

### Provider System (~120 lines)

A provider is a Janet function. No YAML. No template rendering.

```janet
# providers/ollama.janet
(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (def res (glitch/http-post "http://localhost:11434/api/generate"
             :body (json/encode {:model model :prompt prompt :stream false})))
  (def parsed (json/decode res))
  {:response (parsed "response")
   :tokens-in (get-in parsed ["prompt_eval_count"] 0)
   :tokens-out (get-in parsed ["eval_count"] 0)
   :latency (parsed "total_duration")
   :cost 0})
```

- Registry is a table of functions, populated by scanning provider directories
- User providers: drop a `.janet` file in `~/.config/glitch/providers/`
- Tier escalation: try providers in order, catch errors, escalate

Default tiers:
1. `ollama` (qwen2.5:7b)
2. `codex`, `gemini`
3. `copilot`, `claude`

### Store (~60 lines)

SQLite via `janet-sqlite3` C module. Same schema as Go version — existing `.db` files are compatible.

**Tables:** `runs`, `steps`, `research_events` (identical columns to Go)

**Key functions:**
- `(open &opt path)` — open/create DB, apply schema
- `(open-for-workspace ws-path)` — workspace-scoped DB
- `(record-run db rec)` — insert run, return ID
- `(finish-run db id output exit-status totals)` — update with final state
- `(record-step db rec)` — insert step record
- `(list-runs db &named parent-id workflow)` — query runs
- `(get-run db id)` — single run with steps

### Workspace (~80 lines)

Workspace files are Janet programs. Parsing is evaluation.

```janet
# .glitch/workspace.janet
(workspace "ensemble"
  :description "Elastic Ensemble project"
  :owner "adam-stokes"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200")
  (resource "ensemble"
    :type "git"
    :url "https://github.com/elastic/ensemble"
    :ref "main"))
```

- `workspace`, `defaults`, `resource` are macros that build a table
- Workspace files can use conditionals, imports, computed paths — they're programs
- Discovery: explicit flag > env var > walk up cwd > active registry
- Registry: `~/.config/glitch/workspaces.janet`

### CLI Dispatch (~200 lines)

No Cobra. Arg parsing is a ~30-line function. Command table maps strings to functions.

**Commands:** `run`, `check`, `eval`, `workspace` (init/use/list/status/add/rm/sync/pin/resources/gui), `config`, `plugin`, `up`, `version`

### GUI Server (~100 lines)

`spork/httpf` for routing. Same REST API the Svelte frontend expects.

**Endpoints:**
- `GET/POST /api/workflows[/:name[/run]]` — list, get, execute
- `GET /api/runs[/:id[/tree]]` — list, detail, tree
- `GET/POST/DELETE /api/workspace/resources[/:name]` — CRUD
- `POST /api/workspace/sync[/:name]` — sync resources
- `POST /api/workspace/pin` — pin git resource
- `GET /api/providers` — list providers
- `GET /api/results/*path` — serve result files
- `GET /*` — Svelte SPA static files + fallback

Workflow execution via `ev/spawn` — non-blocking. Frontend polls for status.

**Svelte frontend:** zero changes required. Same SPA, same JSON API contract, same endpoint paths and response shapes. The frontend does not know or care that the backend switched from Go to Janet.

### Batch/Compare (~100 lines)

Multi-variant runner orchestration. Same concept as Go:
- Seed shared steps once
- Run variants in parallel via `ev/gather`
- Cross-review with LLM judges
- Generate comparison manifest

### Stdlib (~70 lines)

Three Janet files, auto-loaded:
- `collections.janet`: compact, first, pluck, take, unique, without
- `io.janet`: fetch-json, read-lines, write-lines
- `strings.janet`: kebab-case, snake-case, blank?, present?, words, unwords

### Plugin System

Plugins are directories containing Janet files. `glitch plugin <name>` loads `~/.config/glitch/plugins/<name>/main.janet` via `dofile` and calls its `main` function. Same language as everything else.

## Concurrency Model

Janet uses cooperative concurrency via fibers + event loop. This is sufficient for glitch because:

- glitch is I/O-bound: HTTP requests, subprocess spawning, file I/O, SQLite queries
- All I/O operations yield to the event loop automatically
- `ev/gather` provides parallel execution (replaces Go's `par` with `sync.WaitGroup`)
- `ev/thread` available for CPU-bound work that must not block the event loop
- The GUI server and background workflow execution share the same event loop

## Migration Path

The Go codebase is a porting reference, not maintained in parallel. Port order:

1. **Core + CLI** — `core.janet`, `main.janet`, arg parsing (can run trivial workflows)
2. **Providers** — `provider.janet` + individual provider files (LLM calls work)
3. **Store** — `store.janet` (run recording works, existing DBs compatible)
4. **Workspace** — `workspace.janet` (resource binding works)
5. **Runner** — `runner.janet` (full execution lifecycle)
6. **Stdlib** — port the three small libraries
7. **Batch** — `batch.janet` (comparison runs)
8. **GUI** — `gui.janet` (web interface)
9. **Plugins** — plugin loading from Janet

Each module is independently testable. The binary is usable after step 1.

## What Gets Deleted

| Go component | Lines | Replacement |
|---|---|---|
| Evaluator (`eval.go`) | 1050 | 0 — Janet IS the evaluator |
| Builtins (`eval_builtins.go`) | 900 | 0 — thin wrappers in core.janet |
| Sexpr parser (`internal/sexpr/`) | 400 | 0 — Janet's parser |
| Provider system (`internal/provider/`) | 500 | ~120 lines Janet |
| Store (`internal/store/`) | 300 | ~60 lines Janet |
| Workspace (`internal/workspace/`) | 400 | ~80 lines Janet |
| Runner (`internal/pipeline/runner.go`) | 200 | ~60 lines Janet |
| CLI commands (`cmd/`) | 1500 | ~200 lines Janet |
| GUI server (`internal/gui/`) | 600 | ~100 lines Janet |
| Batch (`internal/batch/`) | 400 | ~100 lines Janet |
| **Total** | **~6300** | **~870** |

7:1 reduction. The ratio comes from eliminating the evaluator, the parser, the Go-DSL bridge, and Cobra.

## Risks

- **Janet ecosystem size**: Small community, fewer libraries. Mitigated by: glitch already shells out for most external integrations; HTTP and SQLite are covered by Janet's C FFI.
- **Learning curve**: Team must learn Janet deeply. Mitigated by: Janet is small (the whole language spec fits in one sitting) and similar to the DSL we already built.
- **spork/httpf maturity**: Less battle-tested than Go's net/http. Mitigated by: the GUI server is low-traffic (single user, local). If needed, swap to raw `net/` primitives.
- **Existing workflow migration**: All `.glitch` files need porting to Janet syntax. Mitigated by: the mapping is nearly mechanical; a conversion script handles most of it.

## Non-Goals

- Maintaining the Go codebase in parallel
- Backwards compatibility with `.glitch` sexpr syntax (clean break)
- Embedding Janet inside Go (CGo path rejected)
