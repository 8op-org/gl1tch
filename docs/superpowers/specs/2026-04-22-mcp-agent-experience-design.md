# MCP Agent Experience Design

**Date:** 2026-04-22
**Status:** Approved

## Problem

The glitch MCP server has two issues:

1. **Connection failure** — missing `protocolVersion` in the initialize response causes Claude Code to reject the handshake.
2. **Poor tool surface** — 9 tools with overlapping concerns, redundant capabilities (file read, grep), and an empty eval sandbox. Agents make poor tool choices when the options are noisy.

## Design Principles

- Only expose tools the agent can't do natively (no file read, no grep)
- Fewer tools with clear descriptions = better agent decisions
- ES code intelligence is first-class (always available)
- Descriptions tell agents *when* to use a tool and *what they get back*

## Protocol Fix (done)

Add `protocolVersion: "2024-11-05"` to the initialize response. Sync server version to `0.3.0` to match `glitch version`.

## Tool Surface: 7 Tools

### Remove (3 tools)

| Tool | Reason |
|------|--------|
| `glitch_read_file` | Agent has native Read tool |
| `glitch_search` | Agent has native Grep/rg |
| `glitch_symbols` | Superseded by ES-backed `glitch_search_symbols` |

### Keep (5 tools, improved descriptions)

#### `glitch_run`

Execute a glitch workflow file and return its output. Use this to run automation pipelines defined in `.glitch/workflows/`. Pass input text and/or key-value parameters. Returns the workflow's stdout on success, or the error message on failure.

**Schema:** `file` (required), `input` (optional string), `set` (optional object of key-value params)

#### `glitch_check`

Validate a glitch workflow file for syntax errors without executing it. Returns "ok" if valid, or a description of the syntax error found.

**Schema:** `file` (required)

**Handler change:** Shell out to `glitch check <file>` instead of bare `read-string`, so it validates against the real parser.

#### `glitch_search_symbols`

Search the code intelligence index for symbol definitions (functions, methods, classes, types, structs, interfaces, traits, enums). Supports wildcard matching with `*`. Use this instead of grep when you need structured symbol metadata across a repository.

**Schema:** `name` (required), `kind`, `language`, `file`, `repo`, `limit` (all optional)

#### `glitch_search_edges`

Query code relationships in the intelligence index: calls, imports, contains, extends, implements, references. Supports depth traversal for multi-hop queries (e.g. "what calls the functions that call X"). Use this to understand how code connects.

**Schema:** `source`, `target`, `kind`, `depth`, `repo`, `limit` (all optional)

#### `glitch_symbol_context`

Get a complete picture of a symbol: its definition plus all relationships (callers, callees, parent, children, implementors). Use this when you need to understand a symbol's role in the codebase in one call rather than multiple `search_edges` queries.

**Schema:** `name` (required), `repo` (optional)

### Add (2 tools)

#### `glitch_eval`

Evaluate a Clojure expression in a context with the full glitch DSL loaded (glitch.core). Use this to programmatically compose and execute workflow steps, query state, or build pipelines dynamically. The glitch DSL provides: `llm`, `sh`, `ref`, `call-workflow`, `search-symbols`, `search-edges`, `symbol-context`, `gate`, `consensus`, `validate`, `json-extract`, and more.

**Schema:** `expression` (required string)

**Handler change:** SCI can't load babashka namespaces directly — build the SCI context by explicitly mapping glitch.core public vars into an SCI namespace. Wire up the provider function so `llm` calls work. Include `clojure.string` and `cheshire.core` (SCI built-ins). This gives the agent a live glitch runtime, not an empty sandbox.

**Available DSL functions:** `llm`, `sh`, `ref`, `input`, `params`, `param`, `search`, `save`, `read-file`, `call-workflow`, `json-extract`, `validate`, `validate-schema`, `gate`, `consensus`, `composite-score`, `search-symbols`, `search-edges`, `symbol-context`, `trace`, `grounded?`

#### `glitch_list_workflows`

List available glitch workflows in `.glitch/workflows/` with their filenames and descriptions. Use this to discover what automation is available before running a workflow.

**Schema:** `path` (optional, defaults to `.glitch/workflows/`)

**Returns:** JSON array of `{name, file, description}` objects. Description is extracted from the first comment or docstring in the workflow file.

## Handler Changes Summary

| Handler | Change |
|---------|--------|
| `handle-search` | Delete |
| `handle-symbols` | Delete |
| `handle-eval` | Rebuild — SCI context with glitch.core referred, provider wired up |
| `handle-read-file` | Delete |
| `handle-check` | Shell out to `glitch check` instead of `read-string` |
| `handle-list-workflows` | New — scan directory, extract descriptions |
| ES handlers | No changes (descriptions updated in tool definitions only) |

## Test Changes

- Update `mcp_test.clj` to expect 7 tools instead of 8
- Add test for `glitch_list_workflows`
- Update protocol test to assert `protocolVersion` in initialize response (done)

## Implementation Notes

- **Subagent per task** — each implementation task gets its own parallel subagent
- **Update documentation** — update `site/content/docs/mcp-server.edn` and `skills/claude-code/SKILL.md` to reflect the new tool surface (removed tools, new tools, updated descriptions)
- **Code review at end** — run code-reviewer agent against the full changeset before merging
- **Push to main** — after review passes, push directly to main
