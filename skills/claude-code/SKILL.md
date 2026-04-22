---
name: glitch
description: Complete reference for the gl1tch CLI (8op-org/gl1tch) — Babashka workflow engine with LLM integration, MCP server, code intelligence, investigation graphs, and plugin system. Use when the user mentions glitch, wants to create/edit/run workflows, automate tasks with glitch, use the MCP server, index code, write plugins, run investigations, or describes a task that could be a glitch workflow. Also use when reviewing or authoring .glitch files.
---

# glitch

gl1tch is a workflow engine built on Babashka (Clojure) that orchestrates shell commands and LLM calls. Core design principle: **shell does the grunt work, LLM does the thinking.** Shell steps fetch and shape data (free, deterministic). LLM steps reason about it (expensive, so feed pre-processed data).

## Installation

```bash
cd ~/Projects/gl1tch/bb
bb install
```

This installs a wrapper script to `~/.local/bin/glitch` that runs:
```bash
exec bb -cp "$HOME/.local/share/glitch/src" -m glitch.main "$@"
```

Source is rsync'd to `~/.local/share/glitch/src`, providers to `~/.config/glitch/providers/`.

**Prerequisites:** Babashka (`bb`), `curl`, `gh` (GitHub CLI, authenticated), `rg` (ripgrep), `sg` (ast-grep, for code intelligence). Run `glitch up` to verify.

| Repo | Source |
|------|--------|
| `8op-org/gl1tch` | `~/Projects/gl1tch` |

## CLI Command Reference

```bash
# Run a workflow
glitch run <file> [input...]                    # execute a .glitch or .clj file
glitch run <file> -p claude                     # use specific provider
glitch run <file> -s key=value                  # set parameter (repeatable)
glitch run <file> -m qwen3:8b                   # override model

# Validate syntax
glitch check <file>                             # parse check (Clojure reader)

# Evaluate Clojure directly
glitch eval <file>                              # load-file a .clj script

# Environment check
glitch up                                       # verify required tools: bb, curl, gh, rg, sg

# Code intelligence
glitch index                                    # index current repo to ES
glitch index --repo ~/Projects/foo              # index specific repo
glitch index --languages go,clojure             # limit languages
glitch index --full                             # skip hash check, full reindex
glitch index --stats                            # show index stats only
glitch index query --name "IndexRepo"           # query symbols
glitch index query --kind function --language go
glitch index query --edges --source "IndexRepo" # query relationships
glitch index query --context "IndexRepo"        # definition + all edges

# REPL
glitch repl                                     # start nREPL on port 1667
glitch repl -p 7777                             # custom port

# Plugins
glitch plugin list                              # list registered plugins
glitch plugin <name> <command> [args...]        # run plugin command
glitch <plugin-name> <command> [args...]        # shorthand

# MCP server (for IDE integration)
glitch mcp                                      # start JSON-RPC stdio server

# Version
glitch version                                  # prints "glitch 0.3.0"
```

## Workflow Authoring

Workflows are `.glitch` files — Clojure s-expressions evaluated via SCI (Small Clojure Interpreter) in a sandbox with all glitch primitives pre-bound.

### File Locations

- **Project-local**: `.glitch/workflows/` — loaded for the project
- **Global plugins**: `~/.config/glitch/plugins/` — available everywhere
- **Workflow dir**: `call-workflow` resolves siblings in the parent directory of the running file

### Syntax

```clojure
;; Comments are standard Clojure

(workflow "name"
  :description "what it does"

  ;; Shell step — runs bash -c
  (step "fetch"
    (sh "gh issue view 42 --json title,body"))

  ;; LLM step — calls configured provider
  (step "analyze"
    (llm
      :provider "claude"
      :model "claude-haiku-4-5-20251001"
      :prompt (str "Analyze this issue:\n" (ref "fetch"))))

  ;; Save output to file
  (save "results/output.md" (ref "analyze")))
```

### Triple-Backtick Strings

The preprocessor converts ` ``` ` blocks into `(str ...)` forms. Use `~(ref "id")` and `~(param "key")` for interpolation inside them:

```clojure
(step "summarize"
  (llm :prompt ```
    Summarize this data for a developer:
    ~(ref "fetch")

    User asked: ~(input)
    Repo: ~(param "repo")
    ```))
```

### DSL Reference

#### Core Primitives

| Form | Description |
|------|-------------|
| `(workflow "name" :description "..." body...)` | Workflow definition |
| `(step "id" body)` | Evaluate body, record output under id |
| `(ref "id")` | Retrieve output of a previous step |
| `(input)` | User input passed to the workflow |
| `(params)` | All runtime parameters as a map |
| `(param "key")` | Single parameter from `--set key=value` |
| `(sh "command")` | Shell command (bash -c), returns stdout |
| `(run "command")` | Alias for `sh` |
| `(save "path" content)` | Write content to file |
| `(read-file "path")` | Read file contents |
| `(write-file "path" content)` | Alias for save |
| `(last-output)` | Output of the most recent step |
| `(get-steps)` | Map of all step outputs |
| `(search "query" :limit 10 :path ".")` | Ripgrep wrapper |

#### LLM Invocation

```clojure
(llm
  :prompt "..."                    ;; required
  :provider "lmstudio"             ;; optional, uses default if omitted
  :model "qwen3:8b"                ;; optional
  :schema {:required ["key"]       ;; JSON schema validation
           :types {"key" :string}
           :enum {"status" ["ok" "fail"]}}
  :retries 2                       ;; retry on schema/confidence failure
  :min-confidence 0.8              ;; minimum confidence threshold
  :skill "path/to/skill.md"       ;; prepend skill content to prompt
  :tools [...]                     ;; tool definitions for tool-calling providers
  :agentic true                    ;; enable multi-round tool calling
  :max-rounds 5                    ;; cap tool-calling rounds
  :domain-check true               ;; apply domain-relevance adjustment
  :step-id "custom-id"             ;; custom step recording ID
  :format "json")                  ;; output format hint
```

#### Control Flow

```clojure
;; Concurrent execution
(par
  (step "a" (sh "curl http://api1"))
  (step "b" (sh "curl http://api2")))

;; Retry on failure
(retry 3
  (step "flaky" (sh "curl -sf https://api.example.com/data")))

;; Timeout (seconds)
(with-timeout 120
  (step "slow" (llm :prompt "Analyze everything...")))

;; Gate — boolean check (non-fatal, records pass/fail)
(gate "has-tests" (not (str/blank? (ref "find-tests"))))

;; Phase — progress grouping
(phase "research"
  (step "fetch" (sh "gh issue view 42 --json body"))
  (step "analyze" (llm :prompt (str "..." (ref "fetch")))))

;; Call sub-workflow (resolves from same directory)
(call-workflow "site-write"
  :input "update the getting-started page"
  :set {"page" "getting-started" "instructions" "add babashka syntax"})

;; Conditional (standard Clojure)
(cond
  (= action "write") (call-workflow "write" :set {"page" slug})
  (= action "dev")   (call-workflow "dev")
  :else              (str "unknown: " action))
```

#### Validation & Confidence

```clojure
;; Contract validation on step output
(step "classify" (llm :prompt "...")
  :expects {:non-empty true :json true :keys ["type" "severity"]})

;; Schema validation against a step
(validate "classify"
  {:required ["type" "severity"]
   :types {"type" :string "severity" :string}
   :enum {"severity" ["low" "medium" "high" "critical"]}})

;; Grounding check — verify output against source context
(grounded? "summary" (ref "raw-data")
  :provider "claude" :strict true :max-unsupported 0)

;; Multi-provider consensus
(consensus ["claude" "openrouter" "lmstudio"]
  :prompt "Classify this issue: ..."
  :schema {:required ["severity"]}
  :compare-key "severity")

;; Composite quality score (harmonic mean of gate results)
(composite-score "classify")
```

#### Investigation Graphs (Bayesian Reasoning)

```clojure
;; Structured investigation with fact tracking
(investigate "The deploy failure was caused by a config change"

  ;; Create facts with confidence scores
  (step "check-logs"
    (sh "kubectl logs deploy/app --tail 100"))
  (fact "logs-show-config-error"
    :claim "Logs contain ConfigMap parse failure"
    :confidence 0.6
    :source-step "check-logs")

  ;; Corroborate from another source
  (step "check-git"
    (sh "git log --oneline -5 -- k8s/"))
  (corroborate-from "config-changed-recently"
    :claim "ConfigMap was modified in last commit"
    :step "check-git"
    :corroborates "logs-show-config-error")

  ;; Query graph state
  (approve! "logs-show-config-error")   ;; mark as trusted (confidence 1.0)
  (reachable? "goal")                    ;; can we reach the goal from approved facts?
  (confidence-gap "goal")                ;; find weakest link
  (suggest-next)                         ;; what to investigate next
  (graph-stats))                         ;; introspection
```

Confidence rules:
- Single-source cap: 0.70 max from one source
- Bayesian combination breaks the cap when multiple sources corroborate
- Contradiction detection: keyword overlap + negation polarity flip
- Geometric decay (0.7^n) when contradictions are detected

#### Code Intelligence Queries

Requires `glitch index` to have been run against the repo with Elasticsearch running.

```clojure
;; Search indexed symbols
(search-symbols {:name "Index*" :kind "function" :language "go"})

;; Search code relationships (calls, imports, extends, implements, references, contains)
(search-edges {:source "IndexRepo" :kind "calls" :depth 2})

;; Full context: definition + all relationships
(symbol-context "IndexRepo")
```

#### Available Utilities

Full Clojure standard library via SCI, plus:

| Symbol | Description |
|--------|-------------|
| `str/*` | All of `clojure.string` (e.g., `str/trim`, `str/split`) |
| `json/decode` | Parse JSON string to map |
| `json/encode` | Map to JSON string |
| `json-extract` | Extract first JSON object/array from LLM noise |
| `mkdir-p` | Create directories |
| `slurp` / `spit` | Read/write files |
| `atom` / `swap!` / `reset!` | Concurrency primitives |
| `future-call` / `deref` | Async execution |
| `def` / `let` / `cond` / `when` / `map` / `filter` / `reduce` | Standard Clojure |

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

- Asking the LLM to parse JSON (use `jq` or `json/decode`)
- Asking the LLM to calculate dates (use `date`)
- Putting all logic in one massive LLM step
- Using multiple LLM steps when one suffices
- Relying on LLM for live API calls (shell steps own data fetching)

## Workflow Patterns

### Pattern 1: Simple fetch + format

```clojure
(workflow "git-status"
  :description "Summarize current git state"

  (step "status"
    (sh "git status --short"))

  (step "summary"
    (llm :prompt (str "Summarize this git status for a developer:\n" (ref "status")))))
```

### Pattern 2: Multi-source aggregation

```clojure
(workflow "morning-briefing"
  :description "Aggregate multiple sources into daily briefing"

  (par
    (step "prs" (sh "gh pr list --author @me --json number,title,state | jq '.'"))
    (step "reviews" (sh "gh pr list --search 'review-requested:@me' --json number,title,url | jq '.'"))
    (step "issues" (sh "gh issue list --assignee @me --json number,title,labels | jq '.'")))

  (step "briefing"
    (llm :prompt ```
      Create a morning briefing from these sources:

      My PRs:
      ~(ref "prs")

      Pending reviews:
      ~(ref "reviews")

      My issues:
      ~(ref "issues")

      Format: bullet list, no emoji, terse.
      ```)))
```

### Pattern 3: Parameterized with --set

```clojure
(workflow "analyze-issue"
  :description "Analyze a GitHub issue"

  (step "fetch"
    (sh (str "gh issue view " (param "issue") " --repo " (param "repo") " --json number,title,body")))

  (step "analyze"
    (llm :prompt (str "Analyze this issue:\n" (ref "fetch"))))

  (save (str "results/" (param "repo") "/" (param "issue") ".md") (ref "analyze")))
```

Run with: `glitch run analyze-issue.glitch -s issue=3442 -s repo=elastic/ensemble`

### Pattern 4: Router with sub-workflows

```clojure
(workflow "site"
  :description "Route freeform input to the right site operation"

  (step "route"
    (llm
      :provider "openrouter"
      :tools []
      :prompt (str "Classify intent. Return JSON: {\"action\": \"...\", \"slug\": \"...\"}.\n"
                   "User said: " (input))))

  (step "dispatch"
    (let [parsed (json/decode (json-extract (ref "route")))
          action (get parsed "action")]
      (cond
        (= action "write-page")
        (:output (call-workflow "site-write"
                   :set {"page" (get parsed "slug")}))
        (= action "dev")
        (:output (call-workflow "site-dev"))
        :else
        (str "unknown action: " action)))))
```

### Pattern 5: Validated LLM with schema

```clojure
(workflow "classify"
  :description "Classify an issue with schema validation"

  (step "fetch" (sh "gh issue view 42 --json title,body,labels"))

  (step "classify"
    (llm
      :prompt (str "Classify:\n" (ref "fetch"))
      :schema {:required ["type" "severity" "confidence"]
               :types {"type" :string "severity" :string "confidence" :number}
               :enum {"severity" ["low" "medium" "high" "critical"]}}
      :retries 2
      :min-confidence 0.7))

  (validate "classify"
    {:required ["type" "severity"]
     :types {"type" :string "severity" :string}}))
```

## Provider & Model Configuration

### Provider Registry

Providers are `.clj` files in `~/.config/glitch/providers/` that call `(glitch.provider/register name fn)` at load time.

Built-in providers (installed by `bb install`):

| Provider | How it runs | Notes |
|----------|------------|-------|
| `lmstudio` | HTTP to localhost:1234 (OpenAI-compatible) | Default local provider. Supports tool calling |
| `claude` | `claude -p` CLI | Strong reasoning, MCP tool support |
| `copilot` | `copilot` CLI | Premium requests |
| `openrouter` | HTTP to openrouter.ai/api/v1 | Free tiers available. Requires `OPENROUTER_API_KEY` |

### Default Provider Fallback

When no provider is specified in a workflow step, glitch tries `lmstudio` first, then falls through default tiers:

1. copilot
2. claude
3. openrouter
4. lmstudio

When a provider IS specified (`:provider "claude"`), it calls that provider directly with no fallback.

### Provider Interface

Every provider implements:
```clojure
(fn [{:keys [prompt model tool-defs]}]
  {:response "..." :tokens-in N :tokens-out M})
```

### Custom Providers

Create a `.clj` file in `~/.config/glitch/providers/`:

```clojure
(ns my-provider
  (:require [glitch.provider :as prov]))

(prov/register "my-provider"
  (fn [{:keys [prompt model]}]
    ;; call your API here
    {:response "..." :tokens-in 0 :tokens-out 0}))
```

## MCP Server

`glitch mcp` starts a JSON-RPC stdio server for IDE integration (Claude Code, Cursor, VS Code Copilot).

### Available Tools

| Tool | Description |
|------|-------------|
| `glitch_search` | Ripgrep wrapper (regex, glob, multiline, PCRE2, context) |
| `glitch_symbols` | Language-aware symbol search (go, python, js, ts, rust, clojure) |
| `glitch_run` | Execute a workflow |
| `glitch_eval` | Evaluate Clojure expressions via SCI |
| `glitch_check` | Syntax validation |
| `glitch_read_file` | Read file (200 lines max) |
| `glitch_search_symbols` | Elasticsearch indexed symbol lookup |
| `glitch_search_edges` | Code relationship queries (calls, imports, extends, implements, references, contains) |
| `glitch_symbol_context` | Full symbol definition + all relationships |

### IDE Configuration

**Claude Code** (`.claude/settings.json`):
```json
{
  "mcpServers": {
    "glitch": {
      "command": "glitch",
      "args": ["mcp"]
    }
  }
}
```

## Plugin System

### Plugin Discovery

Plugins are loaded from:
1. `.glitch/plugins/` (project-local)
2. `~/.config/glitch/plugins/` (global)

### Two Plugin Types

**.glitch file plugins** — each file becomes a command:
```
~/.config/glitch/plugins/github/
  fetch-issue.glitch    -> glitch github fetch-issue
  list-prs.glitch       -> glitch github list-prs
```

**.clj namespace plugins** — self-register via `defcommand`:
```clojure
(ns my-plugin
  (:require [glitch.plugin :refer [defcommand]]))

(defcommand my-func
  "Description of what this does"
  {:args [{:name "repo" :required true}]}
  [opts]
  (str "Result for " (:repo opts)))
```

### Plugin Naming Convention

- **Repos:** `gl1tch-<plugin>` (e.g., `gl1tch-github`)
- **Binaries:** `glitch-<plugin>` (e.g., `glitch-github`)

## Code Intelligence

`glitch index` uses ast-grep to extract symbols and relationships, storing them in Elasticsearch.

**Supported languages:** Go, Python, JavaScript, Rust, Java, C, Clojure

**ES indices:**
- `glitch-symbols-<repo>` — functions, vars, types, macros, protocols
- `glitch-edges-<repo>` — calls, imports, contains, extends, implements, references

```bash
# Index a repo
glitch index --repo ~/Projects/ensemble --languages go,python

# Query from CLI
glitch index query --name "Run*" --kind function
glitch index query --edges --source "IndexRepo" --depth 2
glitch index query --context "IndexRepo"
```

From workflows:
```clojure
(search-symbols {:name "Run*" :kind "function" :language "go"})
(search-edges {:source "IndexRepo" :kind "calls" :depth 2})
(symbol-context "IndexRepo")
```

**Environment variables:**
- `GLITCH_ES_URL` — Elasticsearch endpoint (default: `http://localhost:9200`)
- `GLITCH_TEI_URL` — Text Embeddings Inference endpoint (default: `http://localhost:8090`)

## Data Persistence

EDN-based store at `~/.local/share/glitch/glitch.edn`. Tracks runs, steps, facts, and edges. No external database required.

```clojure
;; State shape
{:counter N
 :runs    {id {:name :input :workflow-file :model :status :output ...}}
 :steps   {id {:run-id :step-id :output :kind :duration ...}}
 :facts   {id {:claim :confidence :source-step ...}}
 :edges   {id {:source :target :kind ...}}}
```

## Project Reference

- **Repo**: `8op-org/gl1tch` at `~/Projects/gl1tch`
- **Runtime**: Babashka (Clojure)
- **Version**: 0.3.0
- **Entry point**: `bb/src/glitch/main.clj`
- **Config**: `~/.config/glitch/providers/`
- **Store**: `~/.local/share/glitch/glitch.edn`
- **Global plugins**: `~/.config/glitch/plugins/`
- **Environment**: `~/.config/glitch/.env` and `./.env`

### Key Modules

| Module | Path | Purpose |
|--------|------|---------|
| `main` | `bb/src/glitch/main.clj` | CLI entry point, arg parsing, command dispatch |
| `core` | `bb/src/glitch/core.clj` | DSL primitives, LLM invocation, validation |
| `runner` | `bb/src/glitch/runner.clj` | SCI sandbox, preprocessor, workflow execution |
| `provider` | `bb/src/glitch/provider.clj` | Provider registry, tiered fallback |
| `graph` | `bb/src/glitch/graph.clj` | Investigation fact graph, Bayesian reasoning |
| `confidence` | `bb/src/glitch/confidence.clj` | Scoring, authority weights, embeddings |
| `index` | `bb/src/glitch/index.clj` | Code intelligence (ast-grep + ES) |
| `store` | `bb/src/glitch/store.clj` | EDN-backed persistence |
| `mcp` | `bb/src/glitch/mcp.clj` | MCP JSON-RPC stdio server |
| `plugin` | `bb/src/glitch/plugin.clj` | Plugin registry, defcommand macro |
| `plugin_loader` | `bb/src/glitch/plugin_loader.clj` | Plugin discovery from filesystem |
| `tool_loop` | `bb/src/glitch/tool_loop.clj` | OpenAI-compatible tool calling loop |
| `repl` | `bb/src/glitch/repl.clj` | nREPL server with DSL pre-loaded |

### Development

```bash
cd ~/Projects/gl1tch/bb
bb install        # install to ~/.local/bin
bb test           # run test suite
bb clean          # clean build artifacts
```

Site development:
```bash
bb site:dev       # shadow-cljs dev server (port 3000, nREPL 7888)
bb site:build     # production build
```
