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
- `glitch-github-<repo>` — GitHub issues and PRs (via `--sources github`)
- `glitch-gitlab-<repo>` — GitLab issues and MRs (via `--sources gitlab`)

## CLI Usage

```bash
# Index a repo (code symbols)
glitch index                                    # index current repo
glitch index --repo ~/Projects/foo              # index specific repo
glitch index --languages go,clojure             # limit languages
glitch index --full                             # skip hash check, full reindex
glitch index --stats                            # show index stats only

# Index sources (GitHub/GitLab issues, PRs, custom data)
glitch index sources                            # list available sources + detection
glitch index --sources github                   # index GitHub issues/PRs (auto-detects remote)
glitch index --sources gitlab                   # index GitLab issues/MRs
glitch index --sources github,gitlab            # multiple sources
glitch index --sources github --no-code         # skip code indexing
glitch index --sources github --since 2026-01-01  # date filter
glitch index --sources github --limit 200       # max items
glitch index --sources github --state open      # state filter (open/closed/all)

# Query symbols
glitch index query --name "IndexRepo"
glitch index query --kind function --language go

# Query relationships
glitch index query --edges --source "IndexRepo"
glitch index query --context "IndexRepo"        # definition + all edges

# Query source indices
glitch index query --from github --name "rate limiting"
```

## Pluggable Index Sources

The index pipeline supports arbitrary data sources via the `defindexer` protocol. Built-in sources (github, gitlab) auto-detect from git remote. Custom sources go in `.glitch/plugins/`:

```clojure
(ns my-support-tickets
  (:require [glitch.index.sources :refer [defindexer]]))

(defindexer support-tickets
  {:doc        "Index support portal tickets"
   :detect     (fn [_] true)
   :index-name (fn [repo] (str "glitch-support-tickets-" repo))
   :mapping    {:properties
                {:id {:type "keyword"} :title {:type "text"}
                 :body {:type "text"} :severity {:type "keyword"}}}
   :cli-opts   [{:flag "--portal-url" :doc "API base URL"}]
   :fetch      (fn [{:keys [portal-url since limit]}]
                 ;; return seq of ES-ready documents
                 )})
```

Protocol keys:
- `:detect` — `(fn [repo-path])` → truthy if source applies here
- `:index-name` — `(fn [repo-name])` → ES index name
- `:mapping` — ES mapping `{:properties {...}}`
- `:cli-opts` — source-specific CLI flags
- `:fetch` — `(fn [opts])` → seq of documents to index

## Workflow Usage

```clojure
;; Search indexed symbols
(search-symbols {:name "Index*" :kind "function" :language "go"})

;; Search code relationships (calls, imports, extends, implements, references, contains)
(search-edges {:source "IndexRepo" :kind "calls" :depth 2})

;; Full context: definition + all relationships
(symbol-context "IndexRepo")
```
