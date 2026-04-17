# Code Graph Design

**Date:** 2026-04-17
**Status:** Draft
**Inspired by:** [colbymchenry/codegraph](https://github.com/colbymchenry/codegraph)

## Goal

Add semantic code intelligence to `glitch index` and `glitch observe`. Replace the
regex symbol extractor with tree-sitter AST parsing. Build a per-repo symbol graph
with six edge kinds in Elasticsearch. Enable structured code navigation, impact
analysis, and graph-aware answers from the observer.

## Non-Goals

- SQLite storage for the graph (ES is the single store)
- LSP integration
- Embedding/vector search on symbols (existing knowledge index handles that)
- New CLI commands (all queries go through `glitch observe`)

---

## Elasticsearch Indices

### `glitch-symbols-{repo}`

One document per symbol.

| Field        | Type    | Notes                                           |
|--------------|---------|-------------------------------------------------|
| `id`         | keyword | Deterministic: SHA256 of `file:kind:name:start` |
| `file`       | keyword | Relative path from repo root                    |
| `kind`       | keyword | function, method, type, interface, class, field, const, var, import, export |
| `name`       | text + keyword | Symbol name, text for FTS, keyword for exact    |
| `signature`  | text    | Full declaration line(s)                         |
| `language`   | keyword | Detected from file extension                    |
| `start_line` | integer | First line of symbol                             |
| `end_line`   | integer | Last line of symbol                              |
| `parent_id`  | keyword | Containing symbol ID (method -> type, etc.)      |
| `docstring`  | text    | Leading comment block if present                 |
| `file_hash`  | keyword | SHA256 of source file (for incremental indexing)  |
| `repo`       | keyword |                                                  |
| `indexed_at` | date    |                                                  |

### `glitch-edges-{repo}`

One document per relationship.

| Field       | Type    | Notes                                    |
|-------------|---------|------------------------------------------|
| `source_id` | keyword | Symbol ID                                |
| `target_id` | keyword | Symbol ID                                |
| `kind`      | keyword | contains, imports, exports, extends, implements, calls |
| `file`      | keyword | File where the edge was observed         |
| `repo`      | keyword |                                          |

### Existing `glitch-code-{repo}`

Unchanged. Content chunks remain for BM25 text search.

---

## Tree-Sitter Extraction

### Dependency

`github.com/smacker/go-tree-sitter` — bundles grammars for Go, Python, JS, TS,
Ruby, Rust, Java, C, C++, C#, PHP, Scala, Swift, Kotlin, Haskell, and more.

### LanguageExtractor

Generic config struct. One instance per language, no language-specific code paths.

```go
type SymbolQuery struct {
    Query string     // tree-sitter S-expression query
    Kind  string     // symbol kind to assign (function, method, type, etc.)
    // Capture names: @name (required), @signature (optional), @docstring (optional)
}

type LanguageExtractor struct {
    Language      string
    Grammar       *sitter.Language
    Extensions    []string
    SymbolQueries []SymbolQuery
    ImportQuery   string          // tree-sitter query for import extraction
    ExportQuery   string          // tree-sitter query for export extraction
    CallQuery     string          // tree-sitter query for call site extraction
    PathResolver  func(importPath string, fromFile string, repoRoot string) string
}
```

Tree-sitter queries are declarative. Adding a new language = adding query strings +
a path resolver function.

### Three-Phase Pipeline

**Phase 1 — Extract (per-file, no cross-file knowledge):**
- Parse source with tree-sitter -> AST
- Run symbol queries -> emit symbol docs with IDs
- Run import/export queries -> emit unresolved import stubs
- Run call queries -> emit unresolved call stubs (callee name + file + line)
- Compute file SHA256

**Phase 2 — Resolve Imports (cross-file, after all symbols indexed):**
- Walk unresolved import stubs
- Resolve import path to target file (language-specific PathResolver)
- Match imported names to symbol IDs in target file
- Emit `imports` edges
- Emit `contains` edges (parent -> child from Phase 1 parent_id)
- Emit `extends` / `implements` edges from type declarations

**Phase 3 — Resolve Calls (after imports resolved):**
- For each unresolved call stub, resolve callee name:
  1. Local scope: same-file symbols
  2. Imported symbols: follow resolved import edges
  3. Unresolved: drop (no edge emitted)
- Emit `calls` edges

Unresolved references are dropped, not guessed. Accurate partial graph > noisy
complete graph.

---

## Incremental Indexing

**Change detection:** file SHA256 hash comparison.

1. On walk, compute SHA256 per file
2. ES terms aggregation on `glitch-symbols-{repo}`: `{file -> file_hash}` for all files
3. Classify files as: unchanged (skip), changed (delete + re-extract), new (extract), deleted (delete)
4. Bulk delete symbols + edges for changed/removed files
5. Run three-phase pipeline on changed + new files only
6. First run is always full

`--full` flag forces full re-index (skips change detection).

---

## Observer Integration

### Expanded Index List

`internal/observer/query.go` adds `glitch-symbols-{repo}` and `glitch-edges-{repo}`
to the search scope when `--repo` is specified.

### Graph Traversal

New `--depth` flag on `glitch observe` (default: 1).

When observer detects a graph-shaped question (callers, callees, impact, dependencies),
it:

1. Queries `glitch-symbols-{repo}` for the target symbol(s)
2. BFS traversal on `glitch-edges-{repo}` following relevant edge kinds to `--depth`
3. Collects the subgraph (symbols + edges)
4. Feeds the subgraph + source snippets to LLM for synthesis

For non-graph questions, observer behavior is unchanged.

---

## Research Tool Additions

Three new tools added to `ToolSet`:

### `search_symbols`

Query `glitch-symbols-{repo}` by name, kind, file, language. Returns structured
symbol data (name, kind, file, line range, signature).

### `search_edges`

Query `glitch-edges-{repo}` by source/target symbol, edge kind. "What calls X",
"what does X call", "what imports X", "what implements Y".

### `symbol_context`

Given a symbol ID, returns: the symbol doc + its edges + source code snippet
(fetched from chunk index by file path + line range). One call, full picture.

---

## CLI Changes

### `glitch index` (enhanced)

```
glitch index [path] [flags]

Flags:
  --repo string       Override repo name (default: directory name)
  --es-url string     Elasticsearch URL (default: from workspace or localhost:9200)
  --languages string  Comma-separated language filter (default: all detected)
  --full              Force full re-index, skip change detection
  --symbols-only      Only index symbols + edges, skip content chunks
  --stats             Print index stats (symbol count, edge count, files, languages)
```

### `glitch observe` (new flag)

```
  --depth int         BFS traversal depth for graph queries (default: 1)
```

---

## Modified Packages

| Package                    | Change                                                        |
|----------------------------|---------------------------------------------------------------|
| `internal/indexer/`        | Replace regex extractor with tree-sitter pipeline. Add LanguageExtractor configs. Three-phase extract/resolve. Incremental via file hash. |
| `internal/observer/`       | Expand index list. Add graph BFS traversal. Handle `--depth`. |
| `internal/esearch/mappings.go` | Add symbol + edge index mappings.                         |
| `cmd/index.go`             | Add flags: --repo, --es-url, --languages, --full, --symbols-only, --stats. |
| `cmd/observe.go`           | Add --depth flag.                                             |
| Research `ToolSet`         | Add search_symbols, search_edges, symbol_context tools.       |

---

## Testing

### Unit Tests

- LanguageExtractor: parse known Go/Python/TS source strings, assert correct symbols + kinds + line ranges
- Three-phase pipeline: mock ES, verify symbol docs + edge docs produced correctly
- Incremental indexing: verify skip/delete/re-extract logic based on hash comparison
- Path resolvers: per-language import path -> file resolution
- Graph BFS: verify traversal depth limiting, edge kind filtering

### Integration Tests

- Index a small multi-file Go package, query symbols + edges from ES, verify graph integrity
- Index, modify one file, re-index, verify only changed file was re-processed
- Observer with graph query: "what calls X" returns correct subgraph

### Smoke Tests

- `glitch index .` on gl1tch repo itself — completes without error, stats show > 0 symbols
- `glitch index . --full` — full re-index completes
- `glitch observe "what functions are in internal/indexer" --repo gl1tch` — returns symbol list
- `glitch observe "what calls IndexRepo" --repo gl1tch --depth 2` — returns caller chain
- Verify against smoke pack baseline (ensemble/kibana/oblt-cli/observability-robots)

---

## Implementation Order

1. **ES mappings** — symbol + edge index definitions in `internal/esearch/mappings.go`
2. **LanguageExtractor framework** — generic struct + tree-sitter parsing loop
3. **Language configs** — Go first, then all bundled grammars
4. **Phase 1: Extract** — symbol + import + call site extraction per file
5. **Phase 2: Resolve imports** — cross-file import resolution + contains/extends/implements edges
6. **Phase 3: Resolve calls** — call graph edges
7. **Incremental indexing** — file hash change detection
8. **CLI flags** — `cmd/index.go` enhancements
9. **Research tools** — search_symbols, search_edges, symbol_context
10. **Observer integration** — expanded index list + graph BFS + --depth flag
11. **Tests + smoke tests**
