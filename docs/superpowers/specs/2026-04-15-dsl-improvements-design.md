# DSL Improvements: Renames, ES Forms, Embed, Template Functions

**Date:** 2026-04-15
**Status:** Draft

## Summary

Three changes to the gl1tch workflow DSL:

1. **Rename forms** for readability — shorter, human-friendly names with permanent aliases for old names
2. **New ES + embed forms** — first-class `(search)`, `(index)`, `(delete)`, `(embed)` to replace curl boilerplate
3. **Template functions** — string manipulation inside `{{ }}` templates so workflows stop shelling out for trivial ops

## 1. Form Renames

Old names become permanent aliases — both old and new resolve to the same converter. No deprecation, no migration pressure.

| Current | New | Rationale |
|---------|-----|-----------|
| `json-pick` | `pick` | Shorter, obvious in context |
| `http-get` | `fetch` | Matches common usage (JS fetch, curl) |
| `http-post` | `send` | Natural pair with fetch |
| `read-file` | `read` | No ambiguity inside a step |
| `write-file` | `write` | No ambiguity inside a step |
| `map` | `each` | Reads naturally for iteration |

Forms that stay as-is: `run`, `llm`, `save`, `lines`, `merge`, `glob`, `plugin`, `retry`, `timeout`, `catch`, `cond`, `par`, `let`, `phase`, `gate`, `def`, `workflow`, `step`.

### Implementation

In `convertStep()` (sexpr.go), add alias cases:

```go
case "pick", "json-pick":
    // ...
case "fetch", "http-get":
    // ...
case "send", "http-post":
    // ...
case "read", "read-file":
    // ...
case "write", "write-file":
    // ...
```

In `convertForm()`, add:

```go
case "each", "map":
    // ...
```

Docs and new workflows use the new names. Old workflows keep working forever.

## 2. New Forms

### 2a. `(search)`

ES `_search` with parsed response. Returns hits as a JSON array of `_source` objects.

```lisp
(search :index "glitch-{{.workspace}}-knowledge-myrepo"
        :query {"term" {"type" "doc"}}
        :size 50
        :fields ("title" "content" "path")
        :es "http://localhost:9200")
```

| Keyword | Required | Default | Description |
|---------|----------|---------|-------------|
| `:index` | yes | — | ES index name (template-rendered) |
| `:query` | yes | — | ES query DSL as inline map `{...}` |
| `:size` | no | 10 | Max hits |
| `:fields` | no | all | `_source` field filter |
| `:es` | no | workspace config, fallback `http://localhost:9200` | ES base URL |

**Output:** JSON array of `_source` objects from hits. No envelope — just the docs.

**Errors:** Non-2xx response or connection failure fails the step (retryable via `(retry)`).

### 2b. `(index)`

Index a single document into ES.

```lisp
(index :index "glitch-{{.workspace}}-knowledge-myrepo"
       :doc "{{step \"synthesize\"}}"
       :id "summary-architecture"
       :embed :field "content" :provider "ollama" :model "nomic-embed-text"
       :es "http://localhost:9200")
```

| Keyword | Required | Default | Description |
|---------|----------|---------|-------------|
| `:index` | yes | — | ES index name (template-rendered) |
| `:doc` | yes | — | JSON document body (template-rendered) |
| `:id` | no | auto | Document `_id` |
| `:embed` | no | — | Auto-embed before indexing (see below) |
| `:es` | no | workspace config | ES base URL |

**`:embed` sub-keywords:**

| Keyword | Required | Default | Description |
|---------|----------|---------|-------------|
| `:field` | yes | — | Which field in `:doc` to embed |
| `:provider` | yes | — | Embedding provider (same as `(llm)` providers) |
| `:model` | yes | — | Embedding model name |

When `:embed` is present, the runtime:
1. Parses `:doc` as JSON
2. Extracts the named field
3. Calls the embedding provider
4. Adds an `embedding` field to the doc
5. Indexes the enriched doc

**Output:** ES index response JSON (`_id`, `_version`, `result`).

### 2c. `(delete)`

ES `_delete_by_query`.

```lisp
(delete :index "glitch-{{.workspace}}-knowledge-myrepo"
        :query {"term" {"type" "summary"}}
        :es "http://localhost:9200")
```

| Keyword | Required | Default | Description |
|---------|----------|---------|-------------|
| `:index` | yes | — | ES index name (template-rendered) |
| `:query` | yes | — | ES query DSL as inline map `{...}` |
| `:es` | no | workspace config | ES base URL |

**Output:** Delete response JSON (`deleted` count).

### 2d. `(embed)`

Standalone embedding — returns the vector without indexing.

```lisp
(embed :input "{{step \"content\"}}"
       :provider "ollama"
       :model "nomic-embed-text")
```

| Keyword | Required | Default | Description |
|---------|----------|---------|-------------|
| `:input` | yes | — | Text to embed (template-rendered) |
| `:provider` | yes | — | Provider name (same routing as `(llm)`) |
| `:model` | yes | — | Embedding model name |

**Output:** JSON array of floats (the embedding vector).

### ES Connection Resolution

Order of precedence for the ES URL:
1. Per-step `:es` keyword (explicit override)
2. Workspace config (`workspace.elasticsearch.url` or equivalent)
3. Default: `http://localhost:9200`

This lets users point at remote clusters without changing workflows — just set the workspace config.

## 3. Template Functions

Add Go template functions for common string operations. These work everywhere templates are rendered (prompts, URLs, paths, doc bodies).

### Functions

| Function | Signature | Example | Result |
|----------|-----------|---------|--------|
| `split` | `split sep str` | `{{split "/" .param.repo}}` | `["elastic", "ensemble"]` |
| `join` | `join sep list` | `{{step "tags" \| split "\n" \| join ", "}}` | `"foo, bar"` |
| `last` | `last list` | `{{split "/" .param.repo \| last}}` | `"ensemble"` |
| `first` | `first list` | `{{split "/" .param.repo \| first}}` | `"elastic"` |
| `upper` | `upper str` | `{{upper .param.name}}` | `"FOO"` |
| `lower` | `lower str` | `{{lower .param.name}}` | `"foo"` |
| `trim` | `trim str` | `{{trim .param.input}}` | strips whitespace |
| `trimPrefix` | `trimPrefix prefix str` | `{{trimPrefix "refs/" .param.ref}}` | `"heads/main"` |
| `trimSuffix` | `trimSuffix suffix str` | `{{trimSuffix ".git" .param.url}}` | `"github.com/foo/bar"` |
| `replace` | `replace old new str` | `{{replace "/" "-" .param.repo}}` | `"elastic-ensemble"` |
| `truncate` | `truncate n str` | `{{truncate 500 .param.content}}` | first 500 chars |
| `contains` | `contains substr str` | `{{if contains "fix" .param.title}}...{{end}}` | bool |
| `hasPrefix` | `hasPrefix prefix str` | `{{if hasPrefix "feat" .param.title}}...{{end}}` | bool |
| `hasSuffix` | `hasSuffix suffix str` | `{{if hasSuffix ".go" .param.path}}...{{end}}` | bool |

### Implementation

Register functions via `template.FuncMap` in the render function (runner.go). All functions are standard Go `strings` package operations — no external dependencies.

```go
funcMap := template.FuncMap{
    "split":      strings.Split,
    "join":       strings.Join,
    "last":       func(s []string) string { return s[len(s)-1] },
    "first":      func(s []string) string { return s[0] },
    "upper":      strings.ToUpper,
    "lower":      strings.ToLower,
    "trim":       strings.TrimSpace,
    "trimPrefix": strings.TrimPrefix,
    "trimSuffix": strings.TrimSuffix,
    "replace":    func(old, new, s string) string { return strings.ReplaceAll(s, old, new) },
    "truncate":   func(n int, s string) string { /* rune-safe truncation */ },
    "contains":   strings.Contains,
    "hasPrefix":  strings.HasPrefix,
    "hasSuffix":  strings.HasSuffix,
}
```

## Before/After: knowledge-synthesis.glitch

### Before (current)

```lisp
(step "query-docs"
  (run ```
REPO="{{.param.repo}}"
REPO_NAME=$(echo "$REPO" | cut -d/ -f2)
INDEX="glitch-{{.workspace}}-knowledge-${REPO_NAME}"
ES="http://localhost:9200"

curl -sf "${ES}/${INDEX}/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"term":{"type":"doc"}},"size":50,"_source":["title","content","path"]}' \
  2>/dev/null | python3 -c "
import json, sys
data = json.load(sys.stdin)
for hit in data['hits']['hits']:
    s = hit['_source']
    path = s.get('path', s.get('title', 'unknown'))
    content = s.get('content', '')[:500]
    print(f'--- {path} ---')
    print(content)
    print()
" 2>/dev/null || echo "(no docs indexed yet)"
```))

(step "delete-old-summaries"
  (run ```
REPO="{{.param.repo}}"
REPO_NAME=$(echo "$REPO" | cut -d/ -f2)
INDEX="glitch-{{.workspace}}-knowledge-${REPO_NAME}"
ES="http://localhost:9200"

curl -sf -X POST "${ES}/${INDEX}/_delete_by_query" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"term":{"type":"summary"}}}' > /dev/null 2>&1
echo "Cleared old summaries"
```))

(step "index-summaries"
  (run ```
REPO="{{.param.repo}}"
REPO_NAME=$(echo "$REPO" | cut -d/ -f2)
EMBED_SCRIPT="$HOME/Projects/stokagent/scripts/embed.sh"
INDEXER="$HOME/Projects/stokagent/scripts/index-json-to-es.py"

cat '{{stepfile "synthesize"}}' | python3 "$INDEXER" "$REPO_NAME" "summary" "$EMBED_SCRIPT"
```))
```

### After

```lisp
(def repo-name "{{split \"/\" .param.repo | last}}")
(def knowledge-index "glitch-{{.workspace}}-knowledge-{{repo-name}}")

(step "query-docs"
  (search :index knowledge-index
          :query {"term" {"type" "doc"}}
          :size 50
          :fields ("title" "content" "path")))

(step "delete-old-summaries"
  (delete :index knowledge-index
          :query {"term" {"type" "summary"}}))

(step "index-summaries"
  (each "synthesize"
    (index :index knowledge-index
           :doc "{{.param.item}}"
           :embed :field "content" :provider "ollama" :model "nomic-embed-text")))
```

~60 lines of shell/curl/python → ~15 lines of DSL.

## Type Changes (types.go)

New struct fields on `Step`:

```go
// ES forms
Search    *SearchStep  `yaml:"-"`
Index     *IndexStep   `yaml:"-"`
Delete    *DeleteStep  `yaml:"-"`
Embed     *EmbedStep   `yaml:"-"`
```

New types:

```go
type SearchStep struct {
    IndexName string
    Query     string            // raw JSON query body
    Size      int
    Fields    []string
    ESURL     string            // override ES URL
}

type IndexStep struct {
    IndexName string
    Doc       string            // template-rendered JSON
    DocID     string            // optional explicit _id
    ESURL     string
    EmbedField   string         // field to embed (empty = no embedding)
    EmbedProvider string
    EmbedModel   string
}

type DeleteStep struct {
    IndexName string
    Query     string            // raw JSON query body
    ESURL     string
}

type EmbedStep struct {
    Input    string             // template-rendered text
    Provider string
    Model    string
}
```

## Files Changed

| File | Change |
|------|--------|
| `internal/pipeline/sexpr.go` | Add alias cases in `convertForm()`/`convertStep()`, add `convertSearch()`, `convertIndex()`, `convertDelete()`, `convertEmbed()` |
| `internal/pipeline/runner.go` | Add `executeSearch()`, `executeIndex()`, `executeDelete()`, `executeEmbed()` in `runSingleStep()`. Add `template.FuncMap` to render function. |
| `internal/pipeline/types.go` | Add `SearchStep`, `IndexStep`, `DeleteStep`, `EmbedStep` types and fields on `Step` |
| `internal/pipeline/es.go` | New file — ES HTTP client helper (shared by search/index/delete). Handles URL resolution, Content-Type headers, response parsing. |
| `internal/pipeline/embed.go` | New file — embedding provider dispatch (reuses LLM provider routing for embedding API calls) |
| `docs/site/workflow-syntax.md` | Update with new form names, document new forms and template functions |

## Out of Scope

- DSL-native expressions in `(def)` — currently defs are strings with Go template interpolation, which clashes visually with the sexpr style. A future pass could make defs evaluate DSL expressions like `(split "/" :param.repo :last)` instead of `"{{split \"/\" .param.repo | last}}"`. Noted as a known wart.
- NDJSON streaming — still capture-then-iterate via `(each)`
- `(each)` as side-effect-only (no output collection) — revisit if needed
- Bulk ES operations (`_bulk` API) — single-doc `(index)` is sufficient for now
- Workspace ES config schema — use existing config mechanism, details TBD during implementation
