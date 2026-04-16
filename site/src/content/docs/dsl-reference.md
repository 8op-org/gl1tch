---
title: "DSL Reference"
order: 7
description: "Reference for advanced gl1tch workflow forms — threading, collection operations, conditionals, Elasticsearch integration, embeddings, and data transforms."
---

## Overview

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it, start there.

The forms below extend your workflows with data pipelines, conditional logic, native Elasticsearch access, and vector embeddings. Each form composes with the others — you can thread a search into a filter into a reduce, or gate a whole pipeline behind a `when` predicate.

## Threading macro

`(->)` pipes data between forms. Each form in the thread implicitly receives the output of the previous one — no need to name intermediate steps.

````glitch
(-> (step "commits"
      (run "git log --oneline -20"))
    (step "changelog"
      (llm :model "qwen2.5:7b"
        :prompt ```
          Turn these commits into a changelog:
          {{step "commits"}}
          ```)))
````

Inside a thread, SDK forms like `search`, `flatten`, and `filter` are auto-wrapped in steps. Collection forms (`map`, `filter`, `reduce`) automatically use the previous step as their source:

````glitch
(-> (search :index "docs" :query "{\"match_all\":{}}" :ndjson)
    (filter
      (step "has-content"
        (run "echo '{{.param.item}}' | jq -e '.content | length > 0'")))
    (step "summarize"
      (llm :model "qwen2.5:7b"
        :prompt ```
          Summarize these documents:
          {{step "has-content"}}
          ```)))
````

You can also thread `flatten`, `each`/`map`, and `reduce` — they all pick up the previous step automatically.

## Collection forms

### filter

Iterates over a step's output (one item per line), runs a body step for each, and keeps items where the body output is truthy (non-empty, not `"false"`, not `"0"`).

```glitch
(step "files"
  (run "find . -name '*.go' -maxdepth 2"))

(filter "files"
  (step "has-tests"
    (run "test -f '{{.param.item | replace \".go\" \"_test.go\"}}' && echo yes")))
```

Inside the body step, `{{.param.item}}` is the current line and `{{.param.item_index}}` is the zero-based index — same as `map`.

### reduce

Folds over a step's output with an accumulator. The body runs once per item, and each run's output becomes the next accumulator value. The final accumulator is the step's output.

````glitch
(step "issues"
  (run "gh issue list --limit 10 --json title,body"))

(reduce "issues"
  (step "digest"
    (llm :model "qwen2.5:7b"
      :prompt ```
        Running summary so far:
        {{.param.accumulator}}

        New item:
        {{.param.item}}

        Update the summary to include the new item. Be concise.
        ```)))
````

The body step receives `{{.param.item}}`, `{{.param.item_index}}`, and `{{.param.accumulator}}`. The accumulator starts as an empty string.

## Conditionals

### when

Runs a shell predicate — if it exits 0, the body executes. If it exits non-zero, the whole form is skipped.

```glitch
(when "test -f .env"
  (step "load-env"
    (run "cat .env")))
```

The body can be any form — a step, a `map`, a `par` block, anything.

### when-not

The inverse of `when`. Runs the body only if the predicate fails (exits non-zero).

```glitch
(when-not "git diff --cached --quiet"
  (step "review"
    (llm :model "qwen2.5:7b"
      :prompt "Review these staged changes: {{step \"diff\"}}")))
```

Use `when-not` to gate on "something has changed" patterns — `git diff --quiet` exits 0 when there are no changes.

## Elasticsearch forms

Native Elasticsearch integration. These forms talk directly to your ES cluster — no shell commands needed.

### search

Query an index and get results back as JSON:

```glitch
(step "recent-errors"
  (search :index "logs-*"
    :query "{\"bool\":{\"must\":[{\"match\":{\"level\":\"error\"}}]}}"
    :size 50
    :sort "{\"@timestamp\":\"desc\"}"
    :fields ("message" "level" "@timestamp")))
```

| Keyword | Required | What it does |
|---------|----------|-------------|
| `:index` | yes | Index name or pattern |
| `:query` | no | Raw JSON query body (default: match_all) |
| `:size` | no | Max hits, default 10 |
| `:fields` | no | Source filter — list of field names |
| `:sort` | no | Raw JSON sort clause |
| `:ndjson` | no | Output as NDJSON (one hit per line) instead of a JSON array |
| `:es` | no | Override the ES URL for this step |

Add `:ndjson` when you want to pipe results into `map`, `filter`, or `reduce` — they operate line by line.

### index

Index a document into Elasticsearch:

```glitch
(step "store"
  (index :index "summaries"
    :doc "{{step \"summary\"}}"
    :id "{{.param.repo}}-latest"))
```

| Keyword | Required | What it does |
|---------|----------|-------------|
| `:index` | yes | Target index name |
| `:doc` | yes | JSON document (template-rendered) |
| `:id` | no | Explicit document `_id` |
| `:upsert` | no | Set to `false` to skip if document exists (op_type=create) |
| `:es` | no | Override the ES URL |

You can also auto-embed a field at index time:

```glitch
(step "store-with-vectors"
  (index :index "docs"
    :doc "{{step \"doc-json\"}}"
    :embed :field "content" :model "nomic-embed-text"))
```

This generates a vector embedding for the `content` field using Ollama and stores it alongside the document.

### delete

Delete documents matching a query:

```glitch
(step "cleanup"
  (delete :index "temp-results"
    :query "{\"match\":{\"session\":\"old-session\"}}"))
```

| Keyword | Required | What it does |
|---------|----------|-------------|
| `:index` | yes | Target index name |
| `:query` | yes | Raw JSON query for delete-by-query |
| `:es` | no | Override the ES URL |

## Embedding

Generate vector embeddings from text via Ollama.

```glitch
(step "vectors"
  (embed :input "{{step \"summary\"}}" :model "nomic-embed-text"))
```

| Keyword | Required | What it does |
|---------|----------|-------------|
| `:input` | yes | Text to embed (template-rendered) |
| `:model` | yes | Ollama embedding model name |

The output is a JSON array of floats — your embedding vector. Use it with `index :embed` for end-to-end vector search pipelines.

## Data transforms

### flatten

Converts a JSON array into NDJSON (one JSON object per line). This bridges JSON array output from `search` or `http-get` into line-oriented forms like `map`, `filter`, and `reduce`.

```glitch
(step "hits"
  (search :index "logs-*" :query "{\"match_all\":{}}" :size 100))

(step "as-lines"
  (flatten "hits"))

(map "as-lines"
  (step "process"
    (run "echo '{{.param.item}}' | jq '.message'")))
```

Or skip the manual `flatten` — use `:ndjson` on `search` to get line-oriented output directly.

Inside a `(->)` thread, `(flatten)` with no arguments automatically flattens the previous step:

```glitch
(-> (search :index "docs" :query "{\"match_all\":{}}")
    (flatten)
    (each
      (step "process"
        (run "echo '{{.param.item}}' | jq '.title'"))))
```

### assoc (template function)

Sets a key on a JSON object string. Use it in templates to enrich data flowing through your pipeline:

```glitch
(step "enrich"
  (run "echo '{{step \"data\" | assoc \"status\" \"reviewed\"}}'"))
```

`assoc` takes the key, the value, and the JSON string (piped). It returns the updated JSON.

### pick (template function)

Extracts a single field from a JSON object string. Supports dot notation for nested access:

```glitch
(step "get-title"
  (run "echo '{{pick \"title\" (step \"issue-json\")}}'"))

;; Nested access with dot notation
(step "get-email"
  (run "echo '{{pick \"author.email\" (step \"commit-json\")}}'"))
```

`pick` returns the field value as a string. If the field holds a nested object or array, it returns the JSON representation.

## Complete form reference

### New forms (this page)

| Form | Description |
|------|-------------|
| `(-> form1 form2 ...)` | Threading macro — pipe data between forms |
| `(filter "source" (step ...))` | Keep items where body output is truthy |
| `(reduce "source" (step ...))` | Fold items with an accumulator |
| `(when "pred" (step ...))` | Execute body if predicate exits 0 |
| `(when-not "pred" (step ...))` | Execute body if predicate exits non-zero |
| `(search :index "name" ...)` | Query Elasticsearch |
| `(index :index "name" :doc "..." ...)` | Index a document to Elasticsearch |
| `(delete :index "name" :query "...")` | Delete by query from Elasticsearch |
| `(embed :input "..." :model "...")` | Generate vector embeddings via Ollama |
| `(flatten "step-id")` | JSON array to NDJSON (one object per line) |

### New template functions

| Function | Description |
|----------|-------------|
| `assoc "key" "val" jsonStr` | Set a field on a JSON object |
| `pick "key" jsonStr` | Extract a field (supports dot notation) |

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — core forms, step references, LLM options, and control flow
- [Plugins](/docs/plugins) — package reusable subcommands and compose them into workflows
- [Batch Comparison Runs](/docs/batch-comparison-runs) — run multiple model variants side by side
