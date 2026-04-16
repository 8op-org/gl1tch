---
title: "Workspaces"
order: 8
description: "Project-scoped configuration that sets defaults for model, provider, Elasticsearch URL, and params. One file, zero boilerplate."
---

```glitch
(workspace "my-project"
  :description "observability pipeline workspace"
  :owner "adam"

  (repos "acme/backend" "acme/frontend")

  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    (params :repo "acme/backend")))
```

That's a `workspace.glitch` file. Drop it at your project root and every gl1tch command picks up those defaults automatically — no flags, no env vars, no config files to keep in sync.

## What a workspace does

A workspace sets project-level defaults so you don't repeat yourself across commands and workflows:

- **Model and provider** — every `(llm ...)` step uses these unless overridden
- **Elasticsearch URL** — flows into search, index, and observe commands automatically
- **Default params** — key-value pairs injected into every workflow run
- **Repos** — the list of repositories this workspace covers
- **Name and owner** — metadata for run logs and telemetry

Without a workspace, you pass `--model`, `--provider`, and `--set` flags on every command. With one, you set them once.

## The workspace.glitch file

The format is s-expressions, same as workflows. Here's the full schema:

```glitch
(workspace "<name>"
  :description "<what this workspace is for>"
  :owner "<your name>"

  (repos "<org/repo-a>" "<org/repo-b>")

  (defaults
    :model "<model name>"
    :provider "<provider name>"
    :elasticsearch "<ES URL>"
    (params
      :<key> "<value>"
      :<key> "<value>")))
```

Every field is optional except the name. A minimal workspace:

```glitch
(workspace "scratch")
```

### Fields

| Field | Where | Description |
|-------|-------|-------------|
| `"<name>"` | top-level | Workspace name, shown in logs and results |
| `:description` | top-level | What this workspace is for |
| `:owner` | top-level | Your name or team |
| `(repos ...)` | top-level | Repositories this workspace covers |
| `:model` | `(defaults ...)` | Default model for LLM steps |
| `:provider` | `(defaults ...)` | Default provider (`ollama`, `copilot`, etc.) |
| `:elasticsearch` | `(defaults ...)` | Elasticsearch URL for search/index/observe |
| `(params ...)` | `(defaults ...)` | Key-value pairs injected as `{{.param.<key>}}` |

## The --workspace flag

Pass `--workspace` on any gl1tch command to point at a workspace directory:

```bash
glitch --workspace ~/Projects/my-project workflow run triage
glitch --workspace ~/Projects/my-project ask "what broke in the last hour?"
glitch --workspace ~/Projects/my-project plugin github fetch-issue --issue 42
```

The flag is persistent — it applies to every subcommand. When set, gl1tch:

1. Reads `workspace.glitch` from that directory
2. Applies model and provider defaults to the active config
3. Resolves workflows from `<workspace>/workflows/`
4. Resolves results to `<workspace>/results/`

If you don't pass `--workspace`, gl1tch resolves the workspace name automatically by walking up from your current directory looking for a `workspace.glitch` file. If it doesn't find one, it looks for a `.glitch/` directory. If neither exists, it falls back to the current directory name.

## Workspace Elasticsearch URL

The `:elasticsearch` default flows into workflow runs automatically. Any workflow that uses Elasticsearch-backed steps — search, index, observe — picks up the URL without you specifying it.

```glitch
;; workspace.glitch
(workspace "observability"
  (defaults
    :elasticsearch "http://localhost:9200"))
```

```bash
# this workflow's ES queries hit localhost:9200 automatically
glitch --workspace ~/Projects/observability workflow run ingest-logs
```

No `--es-url` flag, no hardcoded URL in your workflow files. Change the ES URL in one place and every workflow picks it up.

## Default params

Params in your workspace become template variables in every workflow:

```glitch
(workspace "acme"
  (defaults
    (params
      :repo "acme/backend"
      :team "platform")))
```

Now any workflow can use `{{.param.repo}}` and `{{.param.team}}` without `--set` flags:

```bash
# these are equivalent:
glitch --workspace ~/Projects/acme workflow run triage
glitch workflow run triage --set repo=acme/backend --set team=platform
```

Explicit `--set` flags override workspace params when both are present.

## Example: multi-repo project

A realistic workspace for an observability platform with multiple repos, local Ollama, and a shared Elasticsearch instance:

```glitch
(workspace "observability-platform"
  :description "APM + logging + metrics pipeline"
  :owner "platform-team"

  (repos
    "acme/apm-server"
    "acme/log-collector"
    "acme/metrics-gateway"
    "acme/shared-libs")

  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    (params
      :repo "acme/apm-server"
      :env "staging")))
```

Your project layout:

```
~/Projects/observability-platform/
  workspace.glitch              ;; the file above
  workflows/
    triage.glitch               ;; issue triage pipeline
    ingest-logs.glitch          ;; log ingestion workflow
    cross-review.glitch         ;; variant comparison
  results/
    acme/apm-server/            ;; per-repo results
    acme/log-collector/
  .glitch/
    plugins/                    ;; project-local plugins
```

Run everything from the workspace:

```bash
glitch --workspace ~/Projects/observability-platform workflow run triage --set issue=123
glitch --workspace ~/Projects/observability-platform workflow list
glitch --workspace ~/Projects/observability-platform ask "summarize recent failures"
```

Or just `cd` into the directory and skip the flag — gl1tch finds the `workspace.glitch` automatically:

```bash
cd ~/Projects/observability-platform
glitch workflow run triage --set issue=123
```

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the step forms and templates your workspace workflows use
- [Plugins](/docs/plugins) — reusable data-gathering subcommands that compose with workflows
- [Local Models](/docs/local-models) — setting up Ollama for your workspace's default model
- [Batch Comparison Runs](/docs/batch-comparison-runs) — running workflows across multiple providers
