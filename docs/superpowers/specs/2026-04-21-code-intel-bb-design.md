# Code Intelligence — Babashka Port Design Spec

**Date:** 2026-04-21
**Status:** Active
**Supersedes:** 2026-04-17-code-graph-design.md (Go version)

## Goal

Port the code intelligence system to the Babashka glitch CLI. Build a semantic
code indexer that extracts symbols and relationships via ast-grep, stores them
in Elasticsearch, and exposes queries through CLI, MCP tools, and REPL.

## Architecture

### New Modules

| Module | File | Purpose |
|--------|------|---------|
| `glitch.es` | `bb/src/glitch/es.clj` | Thin ES REST client |
| `glitch.index` | `bb/src/glitch/index.clj` | Three-phase extraction pipeline + CLI |
| `glitch.index.languages` | `bb/src/glitch/index/languages.clj` | Per-language ast-grep rule configs |

### Modified Modules

| Module | Change |
|--------|--------|
| `glitch.main` | Register `glitch index` command |
| `glitch.mcp.handlers` | Add 3 new tools, replace regex `handle-symbols` |
| `glitch.mcp.tools` | Tool schemas for new MCP tools |
| `glitch.core` | Add `search-symbols`, `search-edges`, `symbol-context` DSL functions |
| `glitch.main` (cmd-up) | Add `sg` to prerequisite checks |

### Data Flow

```
repo files
  → filter by language + incremental hash check
  → ast-grep per file (shell out to `sg`)
  → phase 1: extract symbols + contains edges
  → phase 2: resolve imports/exports (cross-file)
  → phase 3: resolve calls + references (cross-file)
  → bulk index to Elasticsearch
```

## Elasticsearch Client (`glitch.es`)

Thin HTTP wrapper. ~60 lines. Uses `babashka.http-client` + `cheshire`.

### Functions

```clojure
(create-index url index mappings)     ;; PUT /{index}
(bulk-index url index docs)           ;; POST /_bulk
(search url index query)              ;; POST /{index}/_search
(delete-index url index)              ;; DELETE /{index}
(doc-count url index)                 ;; GET /{index}/_count
(terms-agg url index field values)    ;; POST /{index}/_search (terms agg)
(index-exists? url index)             ;; HEAD /{index}
```

### Configuration

- Default: `http://localhost:9200`
- Override: `--es-url` flag or `GLITCH_ES_URL` env var

## Elasticsearch Indices

### `glitch-symbols-{repo}`

One document per symbol.

```json
{
  "id":         "repo:file:name:line",
  "file":       "src/main.go",
  "kind":       "function",
  "name":       "IndexRepo",
  "signature":  "func IndexRepo(path string) error",
  "language":   "go",
  "start_line": 42,
  "end_line":   89,
  "parent_id":  "repo:file:Package:1",
  "docstring":  "IndexRepo indexes a repository.",
  "file_hash":  "sha256...",
  "repo":       "gl1tch",
  "indexed_at": "2026-04-21T00:00:00Z"
}
```

### `glitch-edges-{repo}`

One document per relationship.

```json
{
  "source":  "repo:file:Caller:10",
  "target":  "repo:file:Callee:50",
  "kind":    "calls",
  "file":    "src/main.go",
  "repo":    "gl1tch"
}
```

### Edge Kinds

| Kind | Meaning |
|------|---------|
| `contains` | Structural parent-child (class contains method) |
| `imports` | Module/file imports |
| `exports` | Module/file exports |
| `extends` | Inheritance |
| `implements` | Interface/trait implementation |
| `calls` | Function/method invocation |
| `references` | Non-call symbol usage (field access, type use) |

## AST Extraction via ast-grep

### Tool

`sg` (ast-grep CLI), installed via Homebrew (`brew install ast-grep`).

### Rule Files

Located in `bb/resources/ast-grep-rules/{language}/`:

```
bb/resources/ast-grep-rules/
  go/
    symbols.yml      # func, method, type, interface, struct
    imports.yml      # import statements
    calls.yml        # function/method calls
    references.yml   # non-call usage
  python/
    symbols.yml      # def, class, method
    imports.yml      # import, from...import
    calls.yml
    references.yml
  javascript/        # shared with TypeScript
    symbols.yml      # function, class, const/let arrow fns
    imports.yml      # import/require
    calls.yml
    references.yml
  rust/
    symbols.yml      # fn, struct, enum, trait, impl
    imports.yml      # use statements
    calls.yml
    references.yml
  java/
    symbols.yml      # class, interface, method
    imports.yml
    calls.yml
    references.yml
  c/
    symbols.yml      # function, struct, typedef, enum
    imports.yml      # #include
    calls.yml
    references.yml
```

### Invocation

```bash
sg scan --rule <rule.yml> --json <file>
```

Parse JSON output → symbol/edge docs.

### Language Detection

By file extension. Map in `glitch.index.languages`:

```clojure
{".go"    "go"
 ".py"    "python"
 ".js"    "javascript"
 ".ts"    "javascript"
 ".jsx"   "javascript"
 ".tsx"   "javascript"
 ".rs"    "rust"
 ".java"  "java"
 ".c"     "c"
 ".h"     "c"}
```

## Three-Phase Pipeline (`glitch.index`)

### Phase 1: Extract

For each source file:
1. Compute SHA256 hash
2. Check against ES (skip if hash matches — incremental)
3. Run ast-grep symbol rules → symbol docs
4. Build `contains` edges from parent-child nesting
5. Run ast-grep import rules → raw import records
6. Collect into `{:symbols [...] :edges [...] :imports [...]}`

### Phase 2: Resolve Imports

Cross-file pass:
1. Build symbol lookup table `{name → symbol-id}`
2. For each raw import, match to known symbols
3. Create `imports`/`exports` edges
4. Unresolved imports are dropped (not guessed)

### Phase 3: Resolve Calls + References

Cross-file pass:
1. Run ast-grep call rules per file → raw call records
2. Run ast-grep reference rules per file → raw ref records
3. Match call targets to symbol table → `calls` edges
4. Match ref targets to symbol table → `references` edges
5. Unresolved dropped

### Bulk Index

After all three phases:
1. Delete stale symbols (files whose hash changed)
2. Bulk index new/updated symbols
3. Bulk index all edges for changed files

## CLI Command

```
glitch index [flags]
  --repo PATH        repo root (default: cwd)
  --es-url URL       ES endpoint (default: localhost:9200 / GLITCH_ES_URL)
  --languages LANGS  comma-separated filter (default: auto-detect all)
  --full             reindex all files, ignore hashes
  --stats            print index summary, don't index

glitch index query [flags]
  --name PATTERN     symbol name (wildcard OK)
  --kind KIND        filter by symbol kind
  --language LANG    filter by language
  --file PATTERN     filter by file path
  --edges            query edges instead of symbols
  --source NAME      edge source filter
  --target NAME      edge target filter
  --depth N          BFS traversal depth (default: 1)
  --context NAME     shorthand: symbol + all its edges
  --repo REPO        repo name (default: cwd basename)
  --es-url URL       ES endpoint
```

## MCP Tools

### `glitch_search_symbols`

Search the symbol index.

```json
{
  "name": "glitch_search_symbols",
  "description": "Search indexed code symbols by name, kind, language, or file path",
  "inputSchema": {
    "type": "object",
    "properties": {
      "name":     {"type": "string", "description": "Symbol name (wildcard OK)"},
      "kind":     {"type": "string", "description": "Symbol kind: function, method, class, type, interface, trait, struct, enum"},
      "language": {"type": "string", "description": "Language filter"},
      "file":     {"type": "string", "description": "File path pattern"},
      "repo":     {"type": "string", "description": "Repo name (default: cwd)"},
      "limit":    {"type": "integer", "description": "Max results (default: 20)"}
    },
    "required": ["name"]
  }
}
```

### `glitch_search_edges`

Search the edge index with optional BFS traversal.

```json
{
  "name": "glitch_search_edges",
  "description": "Search code relationships (calls, imports, contains, etc.) with optional depth traversal",
  "inputSchema": {
    "type": "object",
    "properties": {
      "source":  {"type": "string", "description": "Source symbol name"},
      "target":  {"type": "string", "description": "Target symbol name"},
      "kind":    {"type": "string", "description": "Edge kind: calls, imports, contains, extends, implements, references"},
      "depth":   {"type": "integer", "description": "BFS traversal depth (default: 1)"},
      "repo":    {"type": "string", "description": "Repo name"},
      "limit":   {"type": "integer", "description": "Max results (default: 50)"}
    }
  }
}
```

### `glitch_symbol_context`

Full context for a single symbol: definition + all edges.

```json
{
  "name": "glitch_symbol_context",
  "description": "Get a symbol's definition and all its relationships (callers, callees, parent, children, implementors)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "name": {"type": "string", "description": "Symbol name"},
      "repo": {"type": "string", "description": "Repo name"}
    },
    "required": ["name"]
  }
}
```

## REPL Functions

Injected into DSL context via `glitch.core`:

```clojure
(search-symbols {:name "Index*" :kind "function" :language "go"})
;; => [{:name "IndexRepo" :kind "function" :file "..." ...} ...]

(search-edges {:source "IndexRepo" :kind "calls" :depth 2})
;; => [{:source "IndexRepo" :target "BulkIndex" :kind "calls"} ...]

(symbol-context "IndexRepo")
;; => {:symbol {...} :callers [...] :callees [...] :parent {...} :children [...]}
```

These functions require an ES URL. Default from `GLITCH_ES_URL` env var or
`http://localhost:9200`. Can be overridden with `:es-url` option.

## Dev Setup

Add `sg` (ast-grep) to `cmd-up` tool checks:

```clojure
(let [tools ["bb" "curl" "gh" "rg" "sg"] ...)
```

## Workflow Update

`update-code-intel-doc.glitch` rewritten to dogfood new tools:

1. Run `glitch index` on gl1tch repo (ensure index is fresh)
2. Use `search-symbols` to pull real symbol data as ground truth
3. Use `search-edges` to show real call graphs
4. Read design spec for feature descriptions
5. Read voice reference for style
6. LLM generates doc grounded in actual indexed data + spec
7. Grounding check
8. Save

## Implementation Order

1. ES client (`glitch.es`)
2. ast-grep rule files (all 6 languages)
3. Extraction pipeline (`glitch.index` + `glitch.index.languages`)
4. CLI command registration (`glitch.main`)
5. `glitch up` update (add `sg`)
6. MCP tools (handlers + schemas)
7. REPL functions (`glitch.core`)
8. Workflow rewrite (`update-code-intel-doc.glitch`)
9. Tests
