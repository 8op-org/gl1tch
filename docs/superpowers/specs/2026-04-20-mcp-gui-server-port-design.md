# MCP Server + GUI HTTP Server Port to Babashka

**Date:** 2026-04-20
**Status:** Draft

## Summary

Port the Janet MCP stdio server and Go GUI HTTP server to Babashka, completing the engine migration. The Svelte frontend stays unchanged — we only replace the Go HTTP backend.

## Scope

### MCP Server (~1,000 lines implementation)

Stdio JSON-RPC 2.0 server exposing 8 tools for code search, indexing, workflow execution, and evaluation.

**Modules:**

| File | Purpose | Lines (est) |
|------|---------|-------------|
| `src/glitch/mcp.clj` | Entry point — stdin line loop, context init | ~50 |
| `src/glitch/mcp/protocol.clj` | JSON-RPC 2.0 parse/format/dispatch | ~60 |
| `src/glitch/mcp/tools.clj` | Tool schema definitions (8 tools) | ~80 |
| `src/glitch/mcp/handlers.clj` | Tool implementations (search, index, run, eval, grep, symbols, read) | ~120 |
| `src/glitch/mcp/indexer.clj` | File walker, chunker, symbol extraction, SQLite FTS5 | ~350 |
| `src/glitch/mcp/search.clj` | Hybrid FTS5 + cosine similarity, score merging | ~140 |
| `src/glitch/mcp/embeddings.clj` | LM Studio `/v1/embeddings` HTTP client | ~40 |
| `src/glitch/mcp/vecmath.clj` | dot-product, magnitude, cosine-similarity | ~20 |

**Tools exposed:**
1. `glitch_search` — hybrid semantic + keyword search
2. `glitch_index` — index/reindex repository
3. `glitch_run` — execute workflow
4. `glitch_eval` — evaluate Clojure/SCI expression
5. `glitch_check` — validate Clojure syntax
6. `glitch_grep` — regex file search
7. `glitch_symbols` — symbol name search via FTS5
8. `glitch_read_file` — read file (200 line cap)

**Key dependencies:**
- `go-sqlite3` pod (FTS5, blob storage)
- `babashka.http-client` (LM Studio embeddings)
- `babashka.process` (subprocess: grep, shasum, glitch)
- `cheshire.core` (JSON)
- SCI (eval/check tools)

**SQLite schema:**
```sql
CREATE TABLE chunks (
  id INTEGER PRIMARY KEY,
  repo TEXT, path TEXT, content TEXT,
  language TEXT, symbols TEXT, hash TEXT,
  embedding BLOB, indexed_at TEXT
);
CREATE VIRTUAL TABLE chunks_fts USING fts5(path, content, symbols, content=chunks, content_rowid=id);
CREATE TABLE index_meta (repo TEXT PRIMARY KEY, model TEXT, dimensions INTEGER, indexed_at TEXT);
```

### GUI HTTP Server (~2,200 lines Go → ~800 lines Clojure)

Babashka HTTP server that serves the pre-built Svelte SPA and implements 23 REST API endpoints.

**Modules:**

| File | Purpose | Lines (est) |
|------|---------|-------------|
| `src/glitch/gui.clj` | Server init, routing, SPA fallback, static serving | ~80 |
| `src/glitch/gui/workflows.clj` | List, get, put, run workflows; config loading | ~150 |
| `src/glitch/gui/runs.clj` | List runs, get run + steps, run tree | ~120 |
| `src/glitch/gui/resources.clj` | CRUD resources, sync, pin (workspace.glitch parsing) | ~200 |
| `src/glitch/gui/results.clj` | Read/write result files | ~50 |
| `src/glitch/gui/workspace.clj` | Get/put workspace, list/switch workspaces | ~80 |
| `src/glitch/gui/telemetry.clj` | ES client, index/search, bulk API | ~120 |

**Approach:** Use `org.httpkit.server` (available in bb) with Ring-style request maps and manual routing (no middleware framework needed for 23 routes).

**Endpoints (matching Go server exactly):**

```
GET  /api/workflows              — list workflows from .glitch/workflows/
GET  /api/workflows/:name        — get workflow source + metadata + params
PUT  /api/workflows/:name        — save workflow source
POST /api/workflows/:name/run    — execute workflow (background), return run_id
GET  /api/workflows/actions/:ctx — workflows matching action context

GET  /api/runs                   — list runs (optional ?workflow= filter, ?parent_id=)
GET  /api/runs/:id               — get run + steps
GET  /api/runs/:id/tree          — get run tree (parent-child hierarchy)

GET  /api/results/*path          — read result file
PUT  /api/results/*path          — save result file

GET  /api/workspace              — current workspace info
PUT  /api/workspace              — update workspace config
GET  /api/workspaces             — list registered workspaces
POST /api/workspaces/use         — switch active workspace

GET  /api/workspace/resources    — list resources from workspace.glitch
POST /api/workspace/resources    — add resource (infer type from URL)
DELETE /api/workspace/resources/:name — remove resource
POST /api/workspace/sync         — sync all resources
POST /api/workspace/sync/:name   — sync specific resource
POST /api/workspace/pin          — pin resource to ref

GET  /api/kibana/workflow/:name  — workflow telemetry from ES
GET  /api/kibana/run/:id         — run telemetry from ES
GET  /api/providers              — list available providers
```

**Static file serving:** Serve `gui/dist/` directory with SPA fallback (non-API GETs that don't match a file → index.html).

**Store:** Reuse `glitch.store` module (SQLite) for runs/steps database.

**Telemetry:** Port ES client from Go — thin HTTP wrapper for bulk indexing, search, and index management. Connects to `localhost:9200`, nil-safe (no-ops if ES unreachable).

**Workflow execution:** `handleRunWorkflow` spawns a future that runs the workflow via the existing `glitch.runner`, records steps to store, and indexes telemetry to ES.

**CLI integration:** Add `glitch gui` command to `main.clj` that starts the server on `localhost:3000`.

## Architecture Decisions

1. **httpkit over ring-jetty** — already bundled in bb, lightweight, async-capable
2. **Reuse existing store** — GUI endpoints read/write same `.glitch/glitch.db`
3. **No websockets initially** — polling for run status (matches Go server behavior)
4. **SPA fallback** — any non-`/api/` GET that doesn't match a static file returns `index.html`
5. **glitch_eval uses SCI** — clean break from Janet; evaluates in same SCI context as runner
6. **glitch_check validates Clojure** — uses `clojure.core/read-string` to detect syntax errors
7. **Embeddings via LM Studio** — same model (`nomic-embed-text`) and endpoint as Janet version
8. **Vector storage as JSON blob** — same `pack-f32`/`unpack-f32` approach (JSON-encoded float arrays in BLOB column)

## Testing Strategy

**MCP tests:**
- `test/glitch/mcp/protocol_test.clj` — unit tests for parse/format/dispatch
- `test/glitch/mcp/vecmath_test.clj` — math correctness
- `test/glitch/mcp/indexer_test.clj` — chunking, symbol extraction, DB operations
- `test/glitch/mcp/search_test.clj` — score normalization, merge logic
- `test/glitch/mcp_test.clj` — integration test (spawn server, send JSON-RPC)

**GUI tests:**
- `test/glitch/gui_test.clj` — start server, hit endpoints, verify JSON responses
- `test/glitch/gui/telemetry_test.clj` — ES client unit tests (mock HTTP)
- Svelte e2e tests (existing Playwright suite) run against the bb server

## Implementation Order

### Phase 1: MCP Server
1. `mcp/vecmath.clj` + test (pure math, no deps)
2. `mcp/embeddings.clj` (HTTP client)
3. `mcp/protocol.clj` + test (JSON-RPC layer)
4. `mcp/tools.clj` (schema definitions)
5. `mcp/indexer.clj` + test (SQLite, file walking, chunking)
6. `mcp/search.clj` + test (hybrid search)
7. `mcp/handlers.clj` (wire everything together)
8. `mcp.clj` + integration test (stdio loop)

### Phase 2: GUI HTTP Server
9. `gui/telemetry.clj` + test (ES client — needed by workflow run)
10. `gui.clj` (server init, routing, static file serving)
11. `gui/workflows.clj` (list/get/put/run workflows)
12. `gui/runs.clj` (list/get runs, run tree)
13. `gui/resources.clj` (CRUD, sync, pin)
14. `gui/results.clj` (read/write result files)
15. `gui/workspace.clj` (workspace management)
16. `gui_test.clj` (endpoint integration tests)

### Phase 3: CLI + Cleanup
17. Wire `glitch mcp` and `glitch gui` commands into `main.clj`
18. Delete old files (janet/, Go source, Taskfile.yml)
19. Update memory files

## Out of Scope

- GUI Svelte frontend changes (stays as-is, served from `gui/dist/`)
- WebSocket/SSE for real-time updates
- Batch/compare module
