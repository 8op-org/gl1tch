---
title: "Plugins"
order: 6
description: "Plugins are directories of `.glitch` files. Each file is a subcommand. No compilation, no release pipeline, no Homebrew "
---

Plugins are directories of `.glitch` files. Each file is a subcommand. No compilation, no release pipeline, no Homebrew tap — edit and run.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it yet, start there — plugins use the same step forms, templates, and control flow.

## Structure

A plugin is a directory. The directory name is the plugin name. Each `.glitch` file (except `plugin.glitch`) becomes a subcommand:

```
.glitch/plugins/github/
  plugin.glitch         ;; optional manifest — shared defs and metadata
  fetch-issue.glitch    ;; → glitch plugin github fetch-issue
  fetch-related.glitch  ;; → glitch plugin github fetch-related
  repo-context.glitch   ;; → glitch plugin github repo-context
  es-context.glitch     ;; → glitch plugin github es-context
```

That gives you:

```bash
glitch plugin github fetch-issue --repo acme/backend --issue 42
glitch plugin github repo-context --repo acme/backend
glitch plugin list
```

## Plugin manifest

Optional. Drop a `plugin.glitch` in your plugin directory to add metadata and shared definitions:

```glitch
;; .glitch/plugins/github/plugin.glitch

(plugin "github"
  :description "GitHub issue and repo context gathering"
  :version "1.0.0")

(def default-repo "acme/backend")
```

- `:description` and `:version` show up in `glitch plugin list` and `--help`
- `(def ...)` bindings are injected into every subcommand's scope before execution
- Subcommands can override shared defs with their own `(def ...)`
- No dependency declarations, no lifecycle hooks — keep it simple

Without a manifest, the directory name is the plugin name and there are no shared defs.

## Subcommand files

Each `.glitch` file declares arguments with `(arg ...)` at the top, then defines a standard workflow. Arguments become CLI flags automatically.

### A simple subcommand

```glitch
;; .glitch/plugins/github/fetch-issue.glitch
;;
;; Usage: glitch plugin github fetch-issue --repo acme/backend --issue 42

(arg "repo" :default "acme/backend" :description "org/repo")
(arg "issue" :description "Issue number")

(workflow "fetch-issue"
  :description "Fetch GitHub issue JSON"

  (step "issue"
    (run "gh issue view {{.param.issue}} --repo {{.param.repo}} --json number,title,body,labels,comments,assignees,milestone")))
```

One arg with a default (`repo`), one required arg (`issue`). That's it — gl1tch generates `--help` from your `(arg ...)` declarations:

```bash
$ glitch plugin github fetch-issue --help
github fetch-issue — Fetch GitHub issue JSON

Flags:
  --repo    org/repo (default: acme/backend)
  --issue   Issue number (required)
```

### Argument types

| Keyword | Description |
|---------|-------------|
| `:default` | Value if not provided. Omit for required args. |
| `:type` | `:string` (default), `:flag` (boolean), `:number` |
| `:description` | Shown in auto-generated `--help` |

```glitch
(arg "since" :default "yesterday" :description "Time range for query")
(arg "authored" :type :flag :description "PRs you authored")
(arg "limit" :type :number :default "50" :description "Max results")
```

Flags on the CLI: `--since week --authored --limit 20`

### Using control flow in plugins

Plugins use the same forms as workflows. Here is a subcommand that uses `catch` for graceful degradation when a GraphQL endpoint is unavailable:

````glitch
;; .glitch/plugins/github/fetch-related.glitch
;;
;; Usage: glitch plugin github fetch-related --repo acme/backend --issue 42

(arg "repo" :default "acme/backend" :description "org/repo")
(arg "issue" :description "Issue number")

(workflow "fetch-related"
  :description "Fetch linked PRs and recent repo activity"

  (catch
    (step "related"
      (run ```
        REPO="{{.param.repo}}"
        ISSUE="{{.param.issue}}"
        echo "=== LINKED PRS ==="
        gh api graphql -f query="
        {
          repository(owner: \"$(echo $REPO | cut -d/ -f1)\",
                     name: \"$(echo $REPO | cut -d/ -f2)\") {
            issue(number: $ISSUE) {
              timelineItems(itemTypes: [CROSS_REFERENCED_EVENT], first: 20) {
                nodes {
                  ... on CrossReferencedEvent {
                    source {
                      ... on PullRequest {
                        number title state url
                      }
                    }
                  }
                }
              }
            }
          }
        }" 2>/dev/null \
          | jq -r '.data.repository.issue.timelineItems.nodes[]?.source
              | select(. != null)
              | "\(.state) #\(.number) \(.title) \(.url)"' 2>/dev/null
        echo ""
        echo "=== RECENT REPO ACTIVITY ==="
        gh api "repos/$REPO/commits?per_page=10" \
          --jq '.[] | "\(.sha[0:7]) \(.commit.message | split("\n")[0])"' 2>/dev/null
        ```))
    (step "fallback"
      (run "echo 'no linked PRs found'"))))
````

### A heavier subcommand

Gather local repo structure, recent commits, and config files — all in one shell step:

````glitch
;; .glitch/plugins/github/repo-context.glitch
;;
;; Usage: glitch plugin github repo-context --repo acme/backend

(arg "repo" :default "acme/backend" :description "org/repo")

(workflow "repo-context"
  :description "Local repo structure and recent commits"

  (step "context"
    (run ```
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"

      if [ ! -d "$REPO_PATH" ]; then
        echo "Repository not cloned locally."
        exit 0
      fi

      echo "=== DIRECTORY STRUCTURE ==="
      find "$REPO_PATH" -maxdepth 3 -type f \
        -not -path '*/.git/*' \
        -not -path '*/node_modules/*' \
        -not -path '*/vendor/*' \
        -not -path '*/dist/*' | head -300

      echo ""
      echo "=== RECENT COMMITS ==="
      git -C "$REPO_PATH" log --oneline -20 2>/dev/null

      echo ""
      echo "=== KEY CONFIG FILES ==="
      for f in Makefile Dockerfile package.json go.mod pyproject.toml; do
        if [ -e "$REPO_PATH/$f" ]; then
          echo "--- $f ---"
          head -30 "$REPO_PATH/$f" 2>/dev/null
          echo ""
        fi
      done
      ```)))
````

## Composing plugins into workflows

This is where it gets interesting. Plugins handle data gathering. Workflows compose them and add LLM reasoning. The `(plugin ...)` form works like any step — output is captured and available downstream via `{{step "id"}}`.

### Issue-to-PR pipeline

A real production workflow that turns a GitHub issue into a PR plan. Four plugin calls gather data, then LLM steps classify, research, draft, and review:

````glitch
;; issue-to-pr.glitch
;;
;; Run with: glitch workflow run issue-to-pr --set repo=acme/backend --set issue=42

(workflow "issue-to-pr"
  :description "Issue-to-PR with tiered escalation"

  ;; --- Data gathering (plugin calls) ---

  (step "fetch-issue"
    (github/fetch-issue
      :repo "{{.param.repo}}"
      :issue "{{.param.issue}}"))

  (step "fetch-related"
    (github/fetch-related
      :repo "{{.param.repo}}"
      :issue "{{.param.issue}}"))

  (step "repo-context"
    (github/repo-context
      :repo "{{.param.repo}}"))

  (step "es-context"
    (github/es-context
      :repo "{{.param.repo}}"
      :issue "{{.param.issue}}"))

  ;; --- LLM reasoning ---

  ;; Classification: fast, low stakes — pin to local tier
  (step "classify"
    (llm :tier 0 :format "json"
      :prompt ```
      Classify this GitHub issue and extract structured information.

      ISSUE:
      {{step "fetch-issue"}}

      RELATED:
      {{step "fetch-related"}}

      Respond with ONLY valid JSON:
      {
        "type": "bug|feature|refactor|documentation",
        "complexity": "small|medium|large",
        "summary": "one line summary",
        "key_requirements": ["req 1", "req 2"],
        "acceptance_criteria": ["criterion 1", "criterion 2"]
      }
      ```))

  ;; Research: needs repo context — still local
  (step "research"
    (llm :tier 0
      :prompt ```
      You are a senior engineer producing an implementation plan.

      ISSUE: {{step "fetch-issue"}}
      CLASSIFICATION: {{step "classify"}}
      REPOSITORY STRUCTURE: {{step "repo-context"}}
      CODE INDEX: {{step "es-context"}}
      LINKED CONTEXT: {{step "fetch-related"}}

      Produce a detailed plan:
      1. Exact file paths to modify or create
      2. What to change and why
      3. Code snippets where helpful
      4. Testing strategy
      5. Risks and mitigations
      ```))

  ;; Build PR artifacts
  (step "build-pr"
    (llm :tier 0
      :prompt ```
      Generate PR artifacts.

      ISSUE: {{step "fetch-issue"}}
      CLASSIFICATION: {{step "classify"}}
      IMPLEMENTATION PLAN: {{step "research"}}

      Output with delimiters:
      ---PR_TITLE---
      Short imperative title (under 70 chars)
      ---END_PR_TITLE---
      ---PR_BODY---
      ## Summary / ## Changes / ## Test Plan
      ---END_PR_BODY---
      ```))

  ;; Self-review against acceptance criteria
  (step "review"
    (llm :tier 0
      :prompt ```
      Review this PR plan against acceptance criteria.
      CLASSIFICATION: {{step "classify"}}
      PR ARTIFACTS: {{step "build-pr"}}
      IMPLEMENTATION PLAN: {{step "research"}}
      For each criterion: PASS or FAIL with one-line reason.
      Then: OVERALL: PASS or OVERALL: FAIL
      ```))

  ;; --- Save results ---

  (step "save-plan"
    (save "results/{{.param.repo}}/issue-{{.param.issue}}/plan.md" :from "research"))

  (step "save-review"
    (save "results/{{.param.repo}}/issue-{{.param.issue}}/review.md" :from "review"))

  (step "save-classify"
    (save "results/{{.param.repo}}/issue-{{.param.issue}}/classification.json" :from "classify"))

  (step "save-pr"
    (save "results/{{.param.repo}}/issue-{{.param.issue}}/pr-body.md" :from "build-pr")))
````

Notice the pattern: plugin calls are deterministic and cacheable — they run `gh`, `git`, `curl`. LLM steps get pre-shaped data and do the reasoning. Swap a plugin subcommand and every workflow that calls it picks up the change.

### Cross-review with glob and timeout

Compare variant outputs across multiple LLM providers. Uses `glob` to find review files and `timeout` to cap the grading step:

````glitch
;; cross-review.glitch — neutral grader for batch comparison runs

(def provider "ollama")
(def model "qwen3:8b")

(workflow "cross-review"
  :description "Compare variant outputs and pick the best one"

  (step "find-reviews"
    (glob "*/review.md"
      :dir "results/{{.param.repo}}/issue-{{.param.issue}}/iteration-1"))

  (step "collect-variants"
    (run ```
      BASE="results/{{.param.repo}}/issue-{{.param.issue}}/iteration-1"
      echo "=== VARIANT RESULTS ==="
      for variant in local claude copilot; do
        DIR="$BASE/$variant"
        if [ -d "$DIR" ]; then
          echo ""
          echo "--- $variant ---"
          echo "REVIEW:"
          cat "$DIR/review.md" 2>/dev/null || echo "(no review)"
          echo ""
          echo "PR TITLE:"
          cat "$DIR/pr-title.txt" 2>/dev/null || echo "(no title)"
        fi
      done
      ```))

  (timeout "90s"
    (step "grade"
      (llm :provider provider :model model
        :prompt ```
        Compare these variant outputs for issue #{{.param.issue}}.

        {{step "collect-variants"}}

        Score each variant 1-10 on:
        1. Plan completeness
        2. Plan specificity
        3. PR quality
        4. Review accuracy

        WINNER: <variant name>
        REASON: <one sentence>
        ```)))

  (step "save-cross-review"
    (save "results/{{.param.repo}}/issue-{{.param.issue}}/cross-review.md" :from "grade")))
````

### Knowledge synthesis with retry and map

Index a repo's documentation into Elasticsearch, then synthesize summaries. Uses `map` for batch file processing and `retry` for the LLM synthesis step:

````glitch
;; repo-ingest.glitch — clone/pull repo, find docs, index each file

(workflow "repo-ingest"
  :description "Ingest repo documentation into ES knowledge index"

  (step "ensure-repo"
    (run ```
      REPO="{{.param.repo}}"
      REPO_NAME=$(echo "$REPO" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      if [ ! -d "$REPO_PATH" ]; then
        gh repo clone "$REPO" 2>&1
      else
        git -C "$REPO_PATH" pull --ff-only 2>&1 || echo "using existing"
      fi
      ```))

  (step "find-docs"
    (run ```
      REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
      REPO_PATH="$HOME/Projects/$REPO_NAME"
      find "$REPO_PATH" -type f \( -name "README*" -o -name "*.md" \
        -o -name "Makefile" -o -name "go.mod" -o -name "pyproject.toml" \) \
        -not -path '*/.git/*' -not -path '*/node_modules/*' \
        -size -100k 2>/dev/null | sort
      ```))

  (map "find-docs"
    (step "process-doc"
      (run ```
        FILE="{{.param.item}}"
        CONTENT=$(cat "$FILE" 2>/dev/null | head -500)
        if [ -z "$CONTENT" ]; then exit 0; fi
        HASH=$(echo "$CONTENT" | shasum -a 256 | cut -d' ' -f1)
        # ... check for changes, index to ES ...
        echo "INDEXED: {{.param.item}}"
        ```))))
````

Then a separate workflow synthesizes the indexed knowledge:

````glitch
;; knowledge-synthesis.glitch — read from ES, produce materialized summaries

(def provider "copilot")

(workflow "knowledge-synthesis"
  :description "Synthesize all knowledge into materialized summaries"

  (step "query-docs"
    (run "curl -sf 'http://localhost:9200/glitch-knowledge-*/_search' ..."))

  (step "query-architecture"
    (run "curl -sf 'http://localhost:9200/glitch-knowledge-*/_search' ..."))

  (retry 2
    (step "synthesize"
      (llm :provider provider
        :prompt ```
        Create a comprehensive knowledge synthesis for {{.param.repo}}.

        DOCUMENTATION: {{step "query-docs"}}
        ARCHITECTURE: {{step "query-architecture"}}

        Produce JSON summaries covering:
        architecture overview, key patterns, common pitfalls,
        decision log, contributor quick-start, testing guide.
        ```))))
````

## CLI reference

```bash
glitch plugin list                              # list all discovered plugins
glitch plugin <name> --help                     # plugin description + subcommands
glitch plugin <name> <subcommand> --help        # args from (arg ...) forms
glitch plugin <name> <subcommand> [--flags]     # run it
```

Output goes straight to stdout — pipe it however you want:

```bash
glitch plugin github fetch-issue --repo acme/backend --issue 42 | jq '.title'
```

## Discovery

Two paths, same as workflows:

1. `.glitch/plugins/<name>/` — project-local, wins on conflict
2. `~/.config/glitch/plugins/<name>/` — user-global

No config registries. No binary auto-discovery. Your plugins are `.glitch` directories.

## Output contract

Your plugin's final step stdout IS the output.

- **From the CLI:** prints to terminal, composes with pipes (`| jq`, `| less`)
- **From a workflow:** becomes `{{step "id"}}` like any other step
- **Stderr** from inner shell steps passes through to terminal

## Design principles

**Shell steps own data fetching.** `gh`, `git`, `curl`, `jq` — anything on your PATH. Free, deterministic, cacheable.

**LLM steps own reasoning.** Summarizing, classifying, planning, reviewing. Expensive, so feed pre-processed data.

**Plugins are the reusable data layer.** Write a plugin subcommand once (`fetch-issue`), compose it into many workflows (`issue-to-pr`, `pr-review`, `cross-review`). Change the plugin, every workflow picks it up.

**Workflows are the orchestration layer.** They compose plugin calls with LLM steps, control flow, and save steps. The workflow decides what to do with the data. The plugin decides how to get it.