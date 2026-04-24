# glitch REPL API — Design Spec

Date: 2026-04-24

## Problem

The glitch workflow DSL runs via SCI (batch, all-or-nothing). Primitives like `llm`, `sh`,
and `step` are SCI macros defined in `runner.clj` — they don't exist as real Clojure
functions. The nREPL server (`glitch.repl`) starts and injects `glitch.core`, but the
useful primitives aren't reachable. You can't sit in Emacs and type `(llm "...")`.

## Goal

REPL-first agent exploration: open Emacs, connect to `glitch repl`, call `llm`, `sh`,
`deftool`, and `agent` as plain functions, watch agents run, then promote the session
to a replayable workflow.

## Design

### New namespace: `glitch.api`

Four primitives, all real Clojure functions (no SCI):

#### `llm`

```clojure
(llm "summarize the last 10 commits")
;; => string

(llm {:model "gemma4" :system "you are a git expert"} "what broke?")
;; => string

(llm {:schema {:type "object" :properties {:severity {:type "string"}}}}
     "classify this error")
;; => {:severity "high"}
```

Reads `@g/*provider-fn*` (set by `repl/start`). Structured output variant parses
JSON and validates against the schema, retrying up to 3 times on failure.

#### `sh`

```clojure
(sh "git log --oneline -10")
;; => stdout string

(sh {:dir "/some/repo"} "git status")
;; => stdout string
```

Thin wrapper over `babashka.process/shell`. Returns stdout as a string. Throws on
non-zero exit code.

#### `deftool`

```clojure
(deftool search [query]
  "Search the codebase index for relevant files"
  (index-query query))
```

Produces an OpenAI-format tool definition map and registers the handler fn in a
`glitch.api/tool-registry` atom. The docstring becomes the LLM-visible description.

Supporting functions:
- `(list-tools)` — print registered tools
- `(remove-tool :search)` — drop a tool

#### `agent`

```clojure
(agent [search run-tests] "why are the integration tests failing?")

(agent {:system "you are a senior engineer on this repo"}
       [search run-tests]
       "why are the integration tests failing?")
```

Calls `tool_loop/run-loop` with the registered tools. Streams each tool call name
to stderr so you can watch it think. Returns the final text string. Defaults:
`:agentic true`, `:max-rounds 10`.

#### Provider switching

```clojure
(use-provider! "openrouter")
(use-model! "anthropic/claude-sonnet-4-6")
```

Mutates `g/*provider-fn*` in place. Survives the rest of the REPL session.

---

### Changes to `glitch.core`

Extract `llm` and `sh` as real Clojure functions in `glitch.core`. The SCI macros
in `runner.clj` already call through to the provider and process — they delegate
to these functions instead of duplicating logic.

---

### Changes to `glitch.repl`

`repl/start` injects `glitch.api` into the `user` namespace alongside the existing
`session`/`promote`/`recall`:

```
llm, sh, deftool, agent, list-tools, remove-tool, use-provider!, use-model!
```

Session recording stays on — every `llm` and `agent` call records automatically,
so `(promote "investigate ci failures")` still works at the end of a session.

---

## What doesn't change

- The `.glitch` workflow DSL and SCI runner — still the save/replay format
- `session`, `promote`, `recall` — unchanged
- `tool_loop/run-loop` — called by `agent`, not rewritten
- `provider` registry — unchanged; `use-provider!` just updates the fn atom

## Files affected

| File | Change |
|------|--------|
| `bb/src/glitch/api.clj` | **new** — `llm`, `sh`, `deftool`, `agent`, `use-provider!`, `use-model!` |
| `bb/src/glitch/core.clj` | extract `llm`/`sh` as real Clojure fns |
| `bb/src/glitch/repl.clj` | inject `glitch.api` into `user` namespace |
