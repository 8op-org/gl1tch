---
name: glitch
description: Complete reference for the gl1tch CLI (8op-org/gl1tch) — workflow authoring, CLI commands, providers, batch runs, workspace model, observer, and installation. Use when the user mentions glitch, wants to create/edit/run workflows, automate tasks with glitch, review PRs via glitch, run batch comparisons, query indexed data, install/update glitch, or describes a task that could be a glitch workflow.
---

# glitch

gl1tch is a GitHub co-pilot CLI that orchestrates shell commands and LLM calls into automated workflows. Core design principle: **shell does the grunt work, LLM does the thinking.** Shell steps fetch and shape data (free, deterministic). LLM steps reason about it (expensive, so feed pre-processed data).

## Installation

Install from the Homebrew tap. Remove any locally built shadows first:

```bash
# 1. Remove locally built shadows
rm -f ~/.local/bin/glitch
rm -f "$(go env GOPATH)/bin/glitch"

# 2. Tap
brew tap 8op-org/tap 2>/dev/null || true

# 3. Reinstall
brew reinstall glitch

# 4. Verify
brew list --versions glitch
glitch --version
```

**Prerequisites:** Ollama must be running locally (`ollama serve`) with `qwen2.5:7b` pulled (`ollama pull qwen2.5:7b`). GitHub CLI (`gh`) must be authenticated.

| Formula | Source repo |
|---------|------------|
| `glitch` | `8op-org/gl1tch` |

The tap auto-updates via goreleaser's `brews:` section. `HOMEBREW_TAP_GITHUB_TOKEN` must be set as a repo secret.

## CLI Command Reference

```bash
# Workflow management
glitch workflow list                           # list available workflows
glitch wf list                                 # alias
glitch workflow run <name> [input]             # run a named workflow
glitch workflow run <name> --set key=value     # parameterized run
glitch workflow run <name> --path /some/repo   # run against different directory

# Observer — query indexed activity via Elasticsearch
glitch observe "How many issues were resolved this week?"
glitch observe "Show all PRs that failed CI" --repo elastic/kibana

# Infrastructure
glitch up                                      # start ES + Kibana via docker-compose
glitch down                                    # stop ES + Kibana
glitch index [path]                            # index repo into ES for code search

# Configuration
glitch config show
glitch config set default_model qwen3:8b
```

## Workspace Model

The `--workspace` flag scopes workflows and results to a directory. Designed for cross-repo work where one command center (e.g., stokagent) manages workflows and results for multiple target repos.

```bash
glitch --workspace ~/Projects/stokagent run issue-to-pr --set repo=elastic/observability-robots --set issue=3920
```

Typical usage with a shell alias:

```bash
alias gl='glitch --workspace ~/Projects/stokagent'
```

When `--workspace` is set:

- **Workflows:** resolved from `<workspace>/workflows/` only. Global `~/.config/glitch/workflows/` is skipped.
- **Results:** written to `<workspace>/results/<org>/<repo>/<issue|pr>-<number>/`.
- **Without `--workspace`:** current behavior — global workflows, CWD-relative results.

### Result directory structure

```
<workspace>/results/<org>/<repo>/
  issue-3920/
    README.md          # rollup — frontmatter metadata + action-ready content
    evidence/          # raw tool call outputs, numbered
      001-github-issue.md
      002-grep-results.md
      003-file-read.md
    plan.md            # implementation plan (if goal=implement)
    review.md          # post-impl review (if applicable)
    run.json           # machine-readable run metadata
```

#### README.md rollup format

```markdown
---
repo: elastic/observability-robots
ref: issue-3920
title: "Fix flaky CI in integration tests"
status: researched | planned | implemented
created: 2026-04-14T10:30:00Z
model: qwen2.5:7b
---

## Summary
<2-3 sentence findings>

## Recommendation
<what to do, concrete>

## Response Draft
<copy-paste ready PR comment, issue reply, or PR body>

## Evidence Index
- [001-github-issue.md](evidence/001-github-issue.md) — original issue body and comments
- [002-grep-results.md](evidence/002-grep-results.md) — relevant code matches
```

#### Variant runs

For comparing outputs across models/tools:

```
results/elastic/observability-robots/
  issue-3920/           # default run
  issue-3920--claude/   # variant
  issue-3920--copilot/  # variant
```

#### run.json schema

```json
{
  "repo": "elastic/observability-robots",
  "ref_type": "issue",
  "ref_number": 3920,
  "workflow": "pr-review",
  "status": "researched",
  "created": "2026-04-14T10:30:00Z",
  "duration_seconds": 45,
  "model": "qwen2.5:7b",
  "variant": null
}
```

## Workflow Authoring

gl1tch supports two workflow formats:

- **`.glitch` (s-expression)** — preferred format, Lisp-like syntax
- **`.yaml`** — legacy YAML format, still supported

### File Locations

- **Global workflows**: `~/.config/glitch/workflows/` — loaded by all `glitch` runs
- **Project-local workflows**: `workflows/` — override globals for the project
- **Workspace workflows**: `<workspace>/workflows/` — when `--workspace` is set (replaces global)
- **Config**: `~/.config/glitch/config.yaml`

Loading order: global dir first, then `workflows/` (local overrides global). With `--workspace`, only `<workspace>/workflows/` is loaded.

### S-Expression Format (.glitch) — Preferred

```clojure
;; comments start with ;;

;; Top-level bindings
(def model "qwen2.5:7b")
(def provider "ollama")

(workflow "workflow-name"
  :description "what it does"

  (step "step-id"
    (run "shell command here"))

  (step "another-step"
    (llm
      :provider provider        ;; resolves def binding
      :model model
      :prompt ```
        Multiline prompt with triple-backtick delimiters.
        Use ~(step step-id) for prior step output.
        Use ~input for user input.
        Use ~param.key for --set key=value params.
        ```))

  (step "write-file"
    (save "results/~param.repo/output.md" :from "another-step"))

  ;; Disable a step without deleting it:
  #_(step "skipped"
    (run "echo this is disabled")))
```

### S-Expression Syntax Reference

| Form | Description |
|------|-------------|
| `(def name "value")` | Top-level binding, substituted in keyword values and run commands |
| `(workflow "name" ...)` | Workflow definition, one per file |
| `:description "text"` | Workflow metadata keyword |
| `(step "id" ...)` | Step definition, id must be unique |
| `(run "command")` | Shell step (sh -c) |
| `(run varname)` | Shell step using a def binding |
| `(llm :prompt "..." [:provider "x"] [:model "y"])` | LLM call |
| `(llm ... :skill "name")` | LLM call with skill content prepended to prompt |
| `(llm ... :tier 1)` | LLM call pinned to a specific escalation tier |
| `(llm ... :format "json")` | LLM call with output format hint |
| `(save "path" :from "step-id")` | Write step output to file |
| `(retry N (step ...))` | Retry step up to N times on failure |
| `(timeout "30s" (step ...))` | Kill step after duration (Go duration string) |
| `(let ((name val) ...) body...)` | Scoped bindings — like def but lexically scoped |
| `(catch (step ...) (step ...))` | Run primary step; on failure, run fallback instead |
| `(cond (pred (step ...)) ...)` | Multi-branch conditional — predicates are shell commands (exit 0 = true) |
| `(map "step-id" (step ...))` | Iterate over prior step output (newline-split), run body per item |
| `` ``` `` | Triple-backtick multiline string (auto-dedented) |
| `#_(...)` | Reader discard — comments out entire s-expression |
| `;; text` | Line comment |

### Control Flow Forms

Forms compose — `retry` can wrap `timeout`, `let` can contain `retry`, etc.

```clojure
;; Retry a flaky API call up to 3 times
(retry 3
  (step "fetch"
    (run "curl -sf https://api.example.com/data")))

;; Kill an LLM step if it hangs beyond 2 minutes
(timeout "2m"
  (step "analyze"
    (llm :prompt "Analyze: ~(step fetch)")))

;; Compose: retry + timeout
(retry 2
  (timeout "30s"
    (step "flaky-slow"
      (run "curl -sf https://slow-api.example.com"))))

;; Scoped bindings (shadows outer defs within body)
(let ((endpoint "https://api.example.com")
      (token "abc123"))
  (step "call"
    (run "curl -H 'Auth: ~param.token' endpoint"))
  (step "parse"
    (run "echo '~(step call)' | jq '.data'")))

;; Error recovery — if primary fails, run fallback
(catch
  (step "try"
    (run "gh api graphql -f query='...'"))
  (step "fallback"
    (run "gh issue view ~param.issue --json body")))

;; Multi-branch conditional — predicates are shell commands
(cond
  ("test -f critical.log"
    (step "alert"
      (run "notify-send 'Critical issue found'")))
  ("test -f warning.log"
    (step "warn"
      (run "echo 'Warnings detected'")))
  (else
    (step "ok"
      (run "echo 'All clear'"))))

;; Iterate over prior step output (one item per line)
;; ~param.item and ~param.item_index available in body
(step "list-files"
  (run "find . -name '*.go' -maxdepth 2"))

(map "list-files"
  (step "check"
    (run "wc -l ~param.item")))
```

### Template Expressions (Tilde Interpolation)

| Expression | Description |
|-----------|-------------|
| `~input` | User input passed to the workflow |
| `~param.key` | Runtime parameter from `--set key=value` |
| `~param.item` | Current item in a `(map ...)` iteration |
| `~param.item_index` | Current index (0-based) in a `(map ...)` iteration |
| `~(step id)` | Output of a previous step |
| `~(stepfile id)` | Write step output to temp file, return path |

### YAML Format (Legacy)

```yaml
name: workflow-name
description: what it does

steps:
  - id: step-id
    run: |
      shell command here

  - id: another-step
    llm:
      provider: claude
      model: claude-haiku-4-5-20251001
      prompt: |
        Prompt with ~(step step-id) and ~input

  - id: write-file
    save: "results/output.md"
    save_step: another-step
```

### The Cardinal Rule: Shell First, LLM Last

**Shell steps collect and prepare data** — `gh`, `git`, `curl`, `jq`, date math, text processing. Fast, deterministic, free.

**LLM steps synthesize the result** — summarizing, prioritizing, formatting, judgment calls. Expensive, so feed pre-processed data.

#### What belongs in shell steps

- API calls: `gh api graphql`, `gh pr view`, `curl`
- Data filtering: `jq` selectors, `grep`, `awk`
- Date computation: `date -v`, arithmetic
- Git operations: `git log`, `git diff`, `git status`
- Text extraction: `sed`, `cut`, field selection

#### What belongs in LLM steps

- Summarizing or explaining data for a human reader
- Prioritizing items based on fuzzy criteria
- Writing natural language reports from structured data
- Making judgment calls (e.g., "is this PR risky?")

#### Anti-patterns

- Asking the LLM to parse JSON (use `jq`)
- Asking the LLM to calculate dates (use `date`)
- Putting all logic in one massive LLM step
- Using multiple LLM steps when one suffices
- Relying on LLM/MCP for live API calls (shell steps own data fetching)

## Workflow Patterns

### Pattern 1: Simple fetch + format

```clojure
(def model "qwen3:8b")

(workflow "git-status"
  :description "Summarize current git state"

  (step "status"
    (run "git status --short"))

  (step "summary"
    (llm :model model
      :prompt ```
        Summarize this git status for a developer:
        ~(step status)
        ```)))
```

### Pattern 2: Multi-source aggregation

```clojure
(def model "qwen3:8b")

(workflow "morning-briefing"
  :description "Aggregate multiple sources into daily briefing"

  (step "prs"
    (run "gh pr list --author @me --json number,title,state | jq '.'"))

  (step "reviews"
    (run "gh pr list --search 'review-requested:@me' --json number,title,url | jq '.'"))

  (step "issues"
    (run "gh issue list --assignee @me --json number,title,labels | jq '.'"))

  (step "briefing"
    (llm :model model
      :prompt ```
        Create a morning briefing from these sources:

        My PRs:
        ~(step prs)

        Pending reviews:
        ~(step reviews)

        My issues:
        ~(step issues)

        Format: bullet list, no emoji, terse.
        ```)))
```

### Pattern 3: Parameterized with --set

```clojure
(workflow "parameterized"
  :description "Pass runtime params with --set"

  (step "fetch"
    (run "gh issue view ~param.issue --repo ~param.repo --json number,title,body"))

  (step "analyze"
    (llm :prompt ```
      Analyze this issue:
      ~(step fetch)
      ```))

  (step "save-it"
    (save "results/~param.repo/~param.issue.md" :from "analyze")))
```

Run with: `glitch workflow run parameterized --set issue=3442 --set repo=elastic/ensemble`

### Pattern 4: Issue-to-PR pipeline (full)

The most complex pattern. Structure:

1. `fetch-issue` — `gh issue view` with full JSON
2. `fetch-related` — linked PRs via GraphQL timeline
3. `repo-context` — local repo structure, recent commits, config files
4. `prior-results` — previous iteration feedback (for iterative improvement)
5. `classify` — LLM extracts type, complexity, requirements, acceptance criteria
6. `research` — LLM produces detailed implementation plan
7. `build-pr` — LLM generates PR title, body, and next steps
8. `review` — LLM grades against acceptance criteria (PASS/FAIL per criterion)
9. `save-results` — shell step writes all artifacts to results dir

### Pattern 5: Using stepfile for complex shell escaping

When step output contains characters that break shell escaping:

```clojure
(step "use-prior-output"
  (run "cat '~(stepfile big-json-step)' | jq '.items[]'"))
```

`~(stepfile id)` writes the step output to a temp file and returns the path.

## Provider & Model Configuration

Config at `~/.config/glitch/config.yaml`:

```yaml
default_model: qwen3:8b
default_provider: ollama

providers:
  openrouter:
    type: openai-compatible
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_API_KEY
    default_model: google/gemma-3-12b-it:free

tiers:
  - providers: [ollama]
    model: qwen3:8b
  - providers: [openrouter]
    model: google/gemma-3-12b-it:free
  - providers: [copilot, claude]
```

### Provider Reference

| Provider | How it runs | Notes |
|----------|------------|-------|
| `ollama` | Local Ollama server at `localhost:11434` | Free, default. Requires `ollama serve` |
| `claude` | `claude -p --output-format text` | Strong reasoning, tool-use agent |
| `copilot` | `gh copilot explain` / Copilot CLI | Premium requests |
| `gemini` | `gemini -p` | Google agent |
| `openrouter` | OpenAI-compatible API | Free tiers available |
| (omitted) | Uses `default_provider` from config | |

### Available Models (Current)

| Tier | Models |
|------|--------|
| **Local (Ollama)** | qwen3:8b (default), qwen2.5:7b, qwen3-coder:30b, qwen3.5:35b-a3b |
| **OpenRouter free** | google/gemma-3-12b-it:free, meta-llama/llama-4-scout:free |
| **Claude** | claude-haiku-4-5-20251001, claude-sonnet-4-6-20250514 |
| **Copilot** | uses default model |

### Tiered Escalation

When no provider is pinned in a workflow step and tiers are configured, glitch uses smart routing:

1. Try tier 1 (e.g., local Ollama)
2. Self-eval the response with local LLM (score 1-5)
3. If score < threshold (default: 4), escalate to next tier
4. Escalation reasons: malformed output, empty response, hallucination, provider error, structural failure, low eval score

## Batch Comparison Runs

Batch runs execute the same workflow across multiple LLM variants to compare output quality.

**Concept:** batch = issues x variants x iterations + cross-review + manifest.

### Naming Convention

```
issue-to-pr-local.glitch      # ollama/qwen2.5:7b
issue-to-pr-claude.glitch     # claude CLI
issue-to-pr-copilot.glitch    # copilot CLI
issue-to-pr-gemma.glitch      # openrouter/gemma
cross-review.glitch            # neutral grader
```

Keep pipeline structure identical across variants — only change `(def provider ...)` and `(def model ...)`.

### Batch Script Pattern

```bash
#!/bin/bash
set -euo pipefail

cd ~/Projects/gl1tch
go build -o /tmp/glitch-batch .
GLITCH="/tmp/glitch-batch"

VARIANTS="local claude copilot"

for variant in $VARIANTS; do
  echo ">>> ($variant) — $(date)"
  $GLITCH workflow run "task-name-$variant" 2>&1 || echo "WARN: $variant failed"
done
```

### Results Structure

```
results/<issue>/
  iteration-1/
    local/
      classification.json
      plan.md
      review.md
      pr-title.txt
      pr-body.md
    claude/
      ...
    cross-review.md
  manifest.md
```

## Observer & Elasticsearch

glitch indexes workflow events to Elasticsearch for observation and telemetry.

```bash
# Start infrastructure
glitch up          # docker-compose: ES + Kibana

# Index a repo for code search
glitch index ~/Projects/ensemble

# Query indexed data
glitch observe "How many issues were resolved this week?"
glitch observe "Show all PRs that failed CI" --repo elastic/kibana

# Stop
glitch down
```

### ES Indices

| Index | Contents |
|-------|----------|
| `glitch-events` | Raw workflow events |
| `glitch-research-runs` | Research loop executions |
| `glitch-tool-calls` | Tool invocations |
| `glitch-llm-calls` | LLM call telemetry (tokens, cost, latency) |
| `glitch-code-<repo>` | Chunked source + symbols (from `glitch index`) |

Kibana dashboard at `http://localhost:5601/app/dashboards#/view/glitch-llm-dashboard`.

## Plugin System

### Naming Convention

- **Repos:** `gl1tch-<plugin>` (e.g., `gl1tch-github`)
- **Binaries:** `glitch-<plugin>` (e.g., `glitch-github`)

### Custom Providers

YAML-defined in `~/.config/glitch/providers/`:

```yaml
name: "my-provider"
command: "command template with ~prompt and ~model"
```

### Release Pipeline

Plugins use GoReleaser + GitHub Actions. Tag a release → GoReleaser builds binaries → auto-updates the Homebrew tap formula at `8op-org/homebrew-tap`.

## Project Reference

- **Repo**: `8op-org/gl1tch` at `~/Projects/gl1tch`
- **Language**: Go (Cobra CLI)
- **Module**: `github.com/8op-org/gl1tch`
- **Config**: `~/.config/glitch/config.yaml`
- **Environment**: `~/.config/glitch/.env` and `./.env`
- **Global workflows**: `~/.config/glitch/workflows/`
- **Custom providers**: `~/.config/glitch/providers/`

### Key Packages

| Package | Path | Purpose |
|---------|------|---------|
| `cmd` | `cmd/` | CLI commands (Cobra) |
| `pipeline` | `internal/pipeline/` | Workflow execution engine |
| `sexpr` | `internal/sexpr/` | S-expression lexer/parser |
| `provider` | `internal/provider/` | Multi-LLM routing and execution |
| `research` | `internal/research/` | Evidence gathering tool-use loop |
| `router` | `internal/router/` | Input → workflow matching |
| `batch` | `internal/batch/` | Multi-issue batch orchestration |
| `observer` | `internal/observer/` | Natural language ES query engine |
| `esearch` | `internal/esearch/` | Elasticsearch HTTP client |
| `store` | `internal/store/` | SQLite persistence |
