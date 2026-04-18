# WebSearch Form Design

**Date:** 2026-04-17
**Status:** Approved

## Summary

Add a `(websearch ...)` workflow form backed by SearXNG, with the endpoint configured as a workspace default. Follows the same pattern as Elasticsearch: connection config in `workspace.glitch`, execution as a workflow step.

## Workspace Config

Add `:websearch` to the defaults block:

```
(workspace "myproject"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    :websearch "http://localhost:8080"))
```

If omitted, the `(websearch ...)` form errors with: "no websearch endpoint configured — add `:websearch` to workspace defaults".

## Workflow Form

```
(websearch "query string"
  :engines ("google" "stackoverflow" "wikipedia")
  :results 5
  :lang "en")
```

### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| query (1st arg) | yes | — | Search query string. Supports template refs (`~(step id)`, `~param.key`). |
| `:engines` | no | SearXNG defaults | List of SearXNG engine names to target. |
| `:results` | no | 5 | Max number of results to return. |
| `:lang` | no | `"en"` | Language filter. |

### Output Format

JSON array, one object per result:

```json
[
  {
    "title": "Why pods get evicted",
    "url": "https://example.com/article",
    "content": "Snippet text from the search engine...",
    "engine": "google"
  }
]
```

### Composability

The form returns structured JSON that can be piped into LLM steps or iterated with `(each ...)`:

```
(step "find-sources"
  (websearch "~param.topic best practices" :results 3))

(each "find-sources"
  (step "fetch-page"
    (fetch "~item.url"))
  (step "summarize"
    :model "qwen2.5:7b"
    "Summarize this page:\n~(step fetch-page)"))
```

## Backend

SearXNG — a self-hosted metasearch engine that aggregates results from Google, Bing, DuckDuckGo, Wikipedia, and others.

- No API keys required
- Local-first — runs as a container alongside Ollama
- JSON API: `GET /search?q=...&format=json`
- Per-query engine selection
- No rate limits or costs
- Users bring their own instance

## Implementation Scope

### Go packages touched

1. **`internal/workspace/`** — parse `:websearch` from defaults, add `WebSearch string` field to workspace struct, serialize back
2. **`internal/pipeline/types.go`** — new `WebSearchStep` struct (query, engines, results, lang)
3. **`internal/pipeline/sexpr.go`** — parse `(websearch ...)` form into `WebSearchStep`
4. **`internal/pipeline/runner.go`** — execute `WebSearchStep` by hitting SearXNG JSON API, return results as JSON string
5. **`docs/site/workflow-syntax.md`** — document the new form
6. **`docs/site/workspaces.md`** — document the `:websearch` default

### Not in scope

- No new resource type — websearch stays in defaults
- No content extraction — use `(fetch ...)` for full page content after search
- No result caching — every `(websearch ...)` call hits SearXNG live
- No SearXNG installation/management — users bring their own instance

### Error handling

- No `:websearch` in workspace defaults → clear error at step execution time when `(websearch ...)` is encountered
- SearXNG unreachable → step fails with connection error, retryable via `(retry ...)`
- Zero results → returns empty JSON array `[]`, not an error
