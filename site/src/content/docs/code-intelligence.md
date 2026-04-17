---
title: "Code Intelligence"
order: 10
description: "Index repos, query with natural language, glitch up/down"
---

Code intelligence lets you index any repository into Elasticsearch and then query it with natural language. Ask things like "what functions call the auth handler?" or "show me all symbol definitions in the billing package" and get grounded, data-backed answers.

## Starting and stopping the stack

`glitch up` and `glitch down` manage the Elasticsearch and Kibana stack via Docker Compose. You need Docker running.

```bash
# Start Elasticsearch + Kibana (detached)
glitch up

# Stop them
glitch down
```

`glitch up` also seeds Kibana with dashboards automatically. If Kibana is still starting up when the seed runs, the command prints a notice — just run `glitch up` again once Kibana is ready.

Elasticsearch listens on `http://localhost:9200`. Kibana is available at `http://localhost:5601`.

## Indexing a repository

`glitch index` extracts symbols, call edges, and code chunks from a repository and pushes them into Elasticsearch:

```bash
# Index the current directory
glitch index

# Index a specific path
glitch index ~/Projects/my-app

# Force a full re-index (clears existing data for the repo)
glitch index ~/Projects/my-app --full

# Index symbols and call graph only, skip content chunks
glitch index ~/Projects/my-app --symbols-only

# Print index stats when done
glitch index ~/Projects/my-app --stats
```

### Flags

| Flag | Default | What it does |
|------|---------|--------------|
| `[path]` | `.` | Repository root to index |
| `--repo` | directory name | Override the repository identifier stored in ES |
| `--es-url` | `http://localhost:9200` | Elasticsearch URL |
| `--languages` | all supported | Comma-separated filter, e.g. `go,python` |
| `--full` | false | Force full re-index — clears existing data first |
| `--symbols-only` | false | Index symbols and call edges, skip code chunk content |
| `--stats` | false | Print a summary of indexed symbols and edges after completion |

### What gets indexed

`glitch index` does language-aware extraction. For each supported language it extracts:

- **Symbols** — function and method definitions, types, classes, constants
- **Call edges** — which symbols call which other symbols
- **Code chunks** — file sections with enough context for semantic search

This produces a queryable call graph alongside full-text and vector search over the codebase content.

### Language support

Pass `--languages` to restrict extraction to specific languages:

```bash
# Only index Go and TypeScript files
glitch index --languages go,typescript
```

Without `--languages`, `glitch` indexes all supported languages it finds in the repo.

## Querying with natural language

`glitch observe` runs a natural language question against indexed data and synthesizes an answer:

```bash
# Ask a question against all indexed repos
glitch observe "what functions handle authentication?"

# Scope to a specific repo
glitch observe "where are database migrations applied?" --repo elastic/kibana

# Use a different provider for answer synthesis
glitch observe "summarize the billing package" \
  --provider ollama \
  --model qwen3:8b

# Increase traversal depth for deeper call graph exploration
glitch observe "trace the call chain from the HTTP handler to the database" \
  --depth 3
```

### Flags

| Flag | Default | What it does |
|------|---------|--------------|
| `[question]` | required | Natural language question |
| `--repo` | all repos | Scope the query to a specific repository |
| `--provider` | `copilot` | LLM provider for query generation and answer synthesis |
| `--model` | — | Model name (provider-specific) |
| `--depth` | `1` | BFS traversal depth — how many hops to follow in the call graph |

### How depth works

`--depth` controls how far `glitch observe` traverses the call graph when building context for your question. At depth 1, it retrieves the directly matching symbols. At depth 2, it also pulls in what those symbols call and what calls them. Higher depth gives richer context but takes longer and uses more tokens.

For most questions, depth 1 is sufficient. Use depth 2 or 3 when you're tracing execution flows or trying to understand how a function fits into the broader call chain.

## Using indexed data in workflows

Once a repo is indexed, workflows can query it directly using the `(search ...)` form from the [DSL Reference](/docs/dsl-reference):

```glitch
(workflow "auth-audit"
  :description "Find all functions that touch the auth layer"

  (step "auth-symbols"
    (search :index "glitch-symbols"
      :query "{\"match\":{\"package\":\"auth\"}}"
      :fields ("name" "file" "line")
      :size 50
      :ndjson))

  (step "analysis"
    (llm :provider "ollama" :model "qwen3:8b"
      :prompt ```
        Review these authentication-related symbols and identify
        any that look risky or inconsistent:
        ~(step auth-symbols)
        ```)))
```

The `(index ...)` and `(delete ...)` forms let you write and clean up documents from workflows too. See [DSL Reference](/docs/dsl-reference) for the full ES form reference.

## Typical workflow

```bash
# 1. Start the stack
glitch up

# 2. Index your repo
glitch index ~/Projects/my-app --stats

# 3. Ask questions
glitch observe "which packages have the most outbound dependencies?" \
  --repo my-app

# 4. Run an analysis workflow that uses the indexed data
glitch run code-health-check --set repo=my-app
```

## Next steps

- [DSL Reference](/docs/dsl-reference) — `(search ...)`, `(index ...)`, and `(delete ...)` forms for ES access in workflows
- [Workspaces](/docs/workspaces) — set a workspace-level Elasticsearch URL so you don't need `--es-url` on every command
- [Getting Started](/docs/getting-started) — initial setup including Docker and Ollama
