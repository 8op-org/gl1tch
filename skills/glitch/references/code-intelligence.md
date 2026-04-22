# Code Intelligence

`glitch index` uses ast-grep to extract symbols and relationships, storing them in Elasticsearch.

## Setup

Requires Elasticsearch running locally (or set `GLITCH_ES_URL`).

**Environment variables:**
- `GLITCH_ES_URL` — Elasticsearch endpoint (default: `http://localhost:9200`)
- `GLITCH_TEI_URL` — Text Embeddings Inference endpoint (default: `http://localhost:8090`)

## Supported Languages

Go, Python, JavaScript, Rust, Java, C, Clojure

## ES Indices

- `glitch-symbols-<repo>` — functions, vars, types, macros, protocols
- `glitch-edges-<repo>` — calls, imports, contains, extends, implements, references

## CLI Usage

```bash
# Index a repo
glitch index                                    # index current repo
glitch index --repo ~/Projects/foo              # index specific repo
glitch index --languages go,clojure             # limit languages
glitch index --full                             # skip hash check, full reindex
glitch index --stats                            # show index stats only

# Query symbols
glitch index query --name "IndexRepo"
glitch index query --kind function --language go

# Query relationships
glitch index query --edges --source "IndexRepo"
glitch index query --context "IndexRepo"        # definition + all edges
```

## Workflow Usage

```clojure
;; Search indexed symbols
(search-symbols {:name "Index*" :kind "function" :language "go"})

;; Search code relationships (calls, imports, extends, implements, references, contains)
(search-edges {:source "IndexRepo" :kind "calls" :depth 2})

;; Full context: definition + all relationships
(symbol-context "IndexRepo")
```
