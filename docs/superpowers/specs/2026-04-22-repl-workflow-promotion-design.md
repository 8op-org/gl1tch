# REPL-to-Workflow Promotion

**Date:** 2026-04-22
**Status:** Approved

## Problem

glitch has two modes that don't naturally flow into each other: the REPL (where daily work happens inside Emacs) and workflows (where reusable value accumulates as `.glitch` files). Nobody starts by saying "I'm going to write a workflow." You start by doing the thing, and only later realize it was worth keeping. The path from REPL exploration to saved workflow is manual and high-friction, so most useful sessions evaporate.

## Design

Three layers, each building on the previous:

### Layer 1: Session Recording

The REPL maintains a persistent session that accumulates across evaluations within a single REPL lifetime.

**What gets recorded:**
- Every `step` call (id, body output, duration)
- Every `sh` / `search` / `llm` call (args, output, tokens)
- Every `ref` resolution
- Timestamp for each entry

**Storage:** An atom in the REPL process, flushed to `~/.local/share/glitch/sessions/` as EDN on each entry. One file per session, named by start timestamp. Sessions rotate when the REPL restarts.

**New REPL functions:**
- `(session)` — show current session summary (step count, duration, what you've done)
- `(session :clear)` — start fresh without restarting the REPL
- `(session :slice :from "step-id")` — mark a range for later promotion

Recording is ambient. No change to how you work.

### Layer 2: Promotion

One function call to turn a REPL session into a workflow:

```clojure
(promote "fetch ES logs and summarize errors")
```

**Flow:**
1. Grab current session (or last `:slice` if set)
2. Send recorded steps to LLM: "Turn this REPL session into a clean `.glitch` workflow. Parameterize the obvious inputs. Keep the step IDs."
3. Show generated workflow in REPL output
4. Ask for confirmation (y/n/edit)
5. On yes: save to `.glitch/workflows/<slugified-description>.glitch`
6. Index with description as metadata

**The LLM step cleans up** false starts, typos, and exploratory dead ends while preserving essential logic. Uses the configured default provider.

**Scope limits:**
- No interactive editor popup — open the file in Emacs after if needed
- No attempt to infer parameters beyond obvious ones (input text, file paths)
- No automatic git commit

**Range promotion:**
```clojure
(promote "desc" :from "step-id" :to "step-id")
```

**Edge cases:**
- Empty session: refuse with message
- Session with only `sh` calls and no `step` wrappers: wrap in steps during promotion

### Layer 3: Recall

Find things by what they do, not by filename. Available in both the REPL and via MCP.

**Index:** `~/.local/share/glitch/index.edn` — a flat list of entries:
```clojure
{:path "path/to/workflow.glitch"
 :description "fetch ES logs and summarize errors"
 :promoted-from "sessions/2026-04-22T14-30-00.edn"  ; nil if hand-authored
 :created "2026-04-22T14:32:00Z"
 :tags ["es" "logs" "errors" "summarize"]}
```

Tags are LLM-extracted from the description.

**REPL function:**
```clojure
(recall "that ES log thing")
```

- Search index by fuzzy match on description and tags
- Workflow found: return path, optionally run with `(recall "..." :run true :input "...")`
- Only session found (never promoted): offer to promote
- Multiple matches: show numbered list, user picks

**MCP tool — `glitch_recall`:**
- New tool alongside existing 7
- Claude searches by intent: `glitch_recall("summarize ES errors")`
- Returns workflow path + description
- Claude runs it via existing `glitch_run`

**No vector embeddings.** Keyword/fuzzy match over descriptions and tags in a flat EDN file.

## Data Flow

```
REPL session atoms
  → flush to ~/.local/share/glitch/sessions/*.edn
    → (promote) → LLM cleanup → .glitch/workflows/*.glitch
      → index entry in ~/.local/share/glitch/index.edn
        → (recall) / glitch_recall MCP tool
```

## Files Changed

| File | Change |
|------|--------|
| `bb/src/glitch/repl.clj` | Session recording, `promote`, `recall`, `session` functions injected into REPL namespace |
| `bb/src/glitch/session.clj` | New — session persistence, index read/write, fuzzy search |
| `bb/src/glitch/promote.clj` | New — LLM-assisted session-to-workflow conversion |
| `bb/src/glitch/mcp/tools.clj` | Add `glitch_recall` tool definition |
| `bb/src/glitch/mcp/handlers.clj` | Handler for recall |
| `bb/src/glitch/core.clj` | Instrument `step`, `sh`, `llm`, `search` with recording hooks |

## Not Changed

- Runner/workflow execution (`.glitch` files work exactly as before)
- Provider system
- Plugin system
- Existing MCP tools
- Workflow DSL syntax
