# Work-on-Issue Workflow + Provider System

**Date**: 2026-04-13
**Status**: Draft

## Overview

Add an end-to-end `work-on-issue` workflow that takes a GitHub issue reference, builds rich context, hands implementation off to any AI coding tool, and produces a results folder with everything needed to open a PR. Replace the current hardcoded provider functions with a YAML-based provider registry so new AI tools can be added without recompiling.

## Provider System

### What changes

The current `provider.go` has hardcoded `RunClaude` and `RunOllama` functions. The pipeline runner has a switch statement that dispatches to one or the other. This doesn't scale.

Replace with a provider registry: YAML files in `~/.config/glitch/providers/` that define how to invoke any AI tool.

### Provider YAML format

```yaml
name: claude
command: claude -p --output-format text "{{.prompt}}"
```

```yaml
name: codex
command: codex -p --full-auto "{{.prompt}}"
```

```yaml
name: copilot
command: gh copilot suggest "{{.prompt}}"
```

Each provider file defines:
- `name`: referenced in workflow steps via `provider: <name>`
- `command`: a Go `text/template` string with `{{.prompt}}` as the placeholder

### Provider loading

1. On startup, load all `.yaml` files from `~/.config/glitch/providers/`
2. Built-in `ollama` provider remains hardcoded (uses HTTP API, not CLI — needed for fast router classification)
3. All other providers use the command template pattern: render prompt into template, exec via shell, capture stdout
4. If a workflow references a provider that doesn't exist, fail with a clear error listing available providers

### Pipeline runner changes

The switch statement in `runner.go` (lines 53-63) becomes:

1. If provider is `ollama` → use `RunOllama` (HTTP API, unchanged)
2. Otherwise → look up provider by name in registry, render command template, exec via `RunShell`

`RunClaude` is deleted. Claude becomes a provider YAML file like everything else.

### Plugin system removal

The `cmd/plugin.go` plugin discovery system is removed. The `glitch-github` binary plugin concept is abandoned. GitHub data fetching stays in shell steps within workflows (using `gh` CLI directly). The provider system is the only extension point.

## Issue Resolution

### Parsing issue references

The router gets a new fast-path regex for `work on issue` patterns. When matched, the input is parsed to extract repo and issue number using progressive resolution:

| Input | Parsed repo | Parsed issue |
|-------|-------------|--------------|
| `work on issue 3442` | current git remote origin | 3442 |
| `work on issue observability-robots#3442` | elastic/observability-robots | 3442 |
| `work on issue elastic/observability-robots#3442` | elastic/observability-robots | 3442 |

Resolution logic:
1. Extract issue number (required) and optional repo qualifier from input
2. If bare number: run `git remote get-url origin`, parse `owner/repo` from the URL
3. If `repo#number`: prepend default org `elastic`
4. If `owner/repo#number`: use as-is
5. Pass `repo` and `issue` as template parameters to the workflow: `{{.param.repo}}`, `{{.param.issue}}`

### Router changes

New fast-path in `router.go` before the LLM fallback:
```go
var reWorkOnIssue = regexp.MustCompile(`work on issue\s+(.+)`)
```

When matched, parse the capture group, resolve repo + issue, route to the `work-on-issue` workflow with structured params.

### Template parameter support

Currently the runner only passes `{{.input}}` to templates. Add support for `{{.param.<key>}}` so the router can pass structured data (repo, issue number) alongside the raw input.

The `data` map in `Run()` becomes:
```go
data := map[string]any{
    "input": input,
    "param": params, // map[string]string from router
}
```

## The `work-on-issue` Workflow

### Step 1: `fetch-issue` (shell)

```yaml
- id: fetch-issue
  run: |
    gh issue view {{.param.issue}} --repo {{.param.repo}} \
      --json number,title,body,labels,comments,assignees
```

Fetches the full issue JSON including body, comments, and labels.

### Step 2: `fetch-repo-context` (shell)

```yaml
- id: fetch-repo-context
  run: |
    # Determine local clone path from repo name
    REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
    REPO_PATH="$HOME/Projects/$REPO_NAME"
    
    # Get directory structure (depth 3)
    find "$REPO_PATH" -maxdepth 3 -type f \
      -not -path '*/node_modules/*' \
      -not -path '*/.git/*' \
      -not -path '*/vendor/*' | head -200
    
    echo "---RECENT_COMMITS---"
    git -C "$REPO_PATH" log --oneline -20
```

Grabs the repo structure and recent history for context. The exact path resolution logic may need refinement (check common clone locations).

### Step 3: `classify` (llm — ollama)

```yaml
- id: classify
  llm:
    provider: ollama
    model: qwen2.5:7b
    prompt: |
      Analyze this GitHub issue and classify it.
      
      Issue JSON:
      {{step "fetch-issue"}}
      
      Respond with ONLY valid JSON:
      {
        "type": "code|documentation|test|refactor|infrastructure",
        "affected_files": ["list", "of", "files"],
        "acceptance_criteria": ["list", "of", "criteria"],
        "related_issues": ["list", "of", "issue", "refs"],
        "summary": "one line summary of what needs to happen"
      }
```

Fast local classification. Extracts structure from the issue body so downstream steps don't need to re-parse.

### Step 4: `gather-context` (shell)

```yaml
- id: gather-context
  run: |
    REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
    
    # Query ES for related code if available
    curl -sf "http://localhost:9200/glitch-code-$REPO_NAME/_search" \
      -H 'Content-Type: application/json' \
      -d '{
        "query": {
          "multi_match": {
            "query": "{{.param.issue}}",
            "fields": ["content", "path", "symbols"]
          }
        },
        "size": 10
      }' 2>/dev/null | jq -r '.hits.hits[]._source | "\(.path):\n\(.content)\n---"' \
      || echo "ES not available, skipping code index lookup"
```

Queries Elasticsearch for code related to the issue. Gracefully skips if ES isn't running or the index doesn't exist. The query terms will be refined using the classification output once we add template chaining for shell steps.

### Step 5: `create-branch` (shell)

```yaml
- id: create-branch
  run: |
    REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
    REPO_PATH="$HOME/Projects/$REPO_NAME"
    ISSUE={{.param.issue}}
    
    cd "$REPO_PATH"
    git fetch origin
    git checkout -b fix/${ISSUE} origin/main
    echo "fix/${ISSUE}"
```

Creates a feature branch in the target repo. Branch name follows `fix/<issue-number>` convention.

### Step 6: `build-prompt` (llm — ollama)

```yaml
- id: build-prompt
  llm:
    provider: ollama
    model: qwen2.5:7b
    prompt: |
      Build a detailed implementation prompt for an AI coding assistant.
      
      ISSUE:
      {{step "fetch-issue"}}
      
      CLASSIFICATION:
      {{step "classify"}}
      
      REPO STRUCTURE:
      {{step "fetch-repo-context"}}
      
      RELATED CODE FROM INDEX:
      {{step "gather-context"}}
      
      Write a complete, self-contained prompt that tells the AI assistant:
      1. What repository and branch to work in
      2. Exactly what files to modify or create
      3. What changes to make, with specific guidance
      4. Acceptance criteria to verify against
      5. What NOT to do (don't commit, don't push, don't open a PR)
      
      The prompt should be detailed enough that the assistant can work
      autonomously without asking clarifying questions. Output ONLY the
      prompt text, no wrapper or explanation.
```

Assembles all context into a single implementation prompt. This is what gets handed to the executor.

### Step 7: `implement` (llm — user's chosen provider)

```yaml
- id: implement
  llm:
    provider: claude
    model: claude-sonnet-4-20250514
    prompt: |
      {{step "build-prompt"}}
```

The provider here is whatever the user wants. To switch to codex, change `provider: codex`. This step does the actual implementation work — the provider receives the full prompt and makes the changes.

### Step 8: `build-results` (shell)

```yaml
- id: build-results
  run: |
    REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
    REPO_PATH="$HOME/Projects/$REPO_NAME"
    ISSUE={{.param.issue}}
    RESULTS_DIR="$REPO_PATH/.glitch/results/$ISSUE"
    
    mkdir -p "$RESULTS_DIR"
    
    # Record branch name
    echo "fix/${ISSUE}" > "$RESULTS_DIR/branch.txt"
    
    # Capture the diff as changes
    git -C "$REPO_PATH" diff origin/main --stat > "$RESULTS_DIR/changes.md"
    echo "" >> "$RESULTS_DIR/changes.md"
    git -C "$REPO_PATH" diff origin/main >> "$RESULTS_DIR/changes.md"
    
    # Ensure .glitch/results is gitignored
    if ! grep -q '.glitch/results' "$REPO_PATH/.gitignore" 2>/dev/null; then
      echo '.glitch/results/' >> "$REPO_PATH/.gitignore"
    fi
    
    echo "Results written to $RESULTS_DIR"
```

Then a follow-up LLM step generates summary.md and next-steps.md:

```yaml
- id: build-summary
  llm:
    provider: ollama
    model: qwen2.5:7b
    prompt: |
      Generate two documents from this issue and diff.
      
      ISSUE:
      {{step "fetch-issue"}}
      
      CHANGES:
      {{step "build-results"}}
      
      IMPLEMENTATION NOTES:
      {{step "implement"}}
      
      Document 1 - PR SUMMARY (output between ---SUMMARY_START--- and ---SUMMARY_END---):
      Write a PR title and body. Include: what changed, why, test plan.
      
      Document 2 - NEXT STEPS (output between ---NEXT_START--- and ---NEXT_END---):
      List what the user needs to do to get this up for review:
      - Review the changes locally
      - Run specific tests
      - Push the branch
      - Open the PR (include the gh command with the summary)
```

A final shell step parses the delimiters and writes the files:

```yaml
- id: write-artifacts
  run: |
    REPO_NAME=$(echo "{{.param.repo}}" | cut -d/ -f2)
    REPO_PATH="$HOME/Projects/$REPO_NAME"
    ISSUE={{.param.issue}}
    RESULTS_DIR="$REPO_PATH/.glitch/results/$ISSUE"
    
    CONTENT='{{step "build-summary"}}'
    
    echo "$CONTENT" | sed -n '/---SUMMARY_START---/,/---SUMMARY_END---/p' \
      | sed '1d;$d' > "$RESULTS_DIR/summary.md"
    
    echo "$CONTENT" | sed -n '/---NEXT_START---/,/---NEXT_END---/p' \
      | sed '1d;$d' > "$RESULTS_DIR/next-steps.md"
    
    echo ""
    echo "=== Issue #$ISSUE — Ready for Review ==="
    echo ""
    echo "Branch: fix/$ISSUE"
    echo "Results: $RESULTS_DIR/"
    echo ""
    cat "$RESULTS_DIR/next-steps.md"
```

## Results Folder

The workflow produces `.glitch/results/<issue-number>/` in the target repo:

```
.glitch/results/3442/
  summary.md      # PR title + body, ready to paste or use with gh pr create
  changes.md      # Full diff against main
  next-steps.md   # What to do next (review, test, push, open PR)
  branch.txt      # Branch name
```

`.glitch/results/` is added to `.gitignore` automatically if not already present.

## Changes Required

### New files
- `~/.config/glitch/providers/claude.yaml` — Claude provider template
- `~/.config/glitch/providers/codex.yaml` — Codex provider template  
- `~/.config/glitch/providers/copilot.yaml` — Copilot provider template
- `~/Projects/stokagent/workflows/work-on-issue.yaml` — The workflow (symlinked into `~/.config/glitch/workflows/`)

### Modified files
- `internal/provider/provider.go` — Add `LoadProviders()`, `RunProvider(name, prompt)`. Remove `RunClaude`.
- `internal/pipeline/runner.go` — Replace provider switch with registry lookup. Add `param` support to template data.
- `internal/pipeline/types.go` — Add `Params map[string]string` to `Workflow` or pass through `Run()`.
- `internal/router/router.go` — Add `reWorkOnIssue` fast-path, issue reference parsing, `Match()` returns params.
- `cmd/ask.go` — Pass params from router to `pipeline.Run()`.

### Removed files
- `cmd/plugin.go` — Plugin discovery system removed.

## Out of Scope

- Project board integration (requires `read:project` token scope, separate effort)
- Write-back to GitHub (commenting on issues, updating labels)
- Automatic PR creation (user reviews locally first, future `open-pr` workflow)
- Multi-issue batching
- Workflow resumability / checkpointing
