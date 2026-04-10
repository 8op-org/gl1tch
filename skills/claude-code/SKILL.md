---
name: gl1tch
description: Use the gl1tch CLI to automate GitHub tasks, run workflows, and compose shell+LLM pipelines. Invoke when the user asks to review PRs, triage issues, check CI, run any glitch workflow, or set up glitch in their project.
---

# gl1tch — GitHub Automation CLI

## When to use

- User asks to review a PR, check issues, or see CI status
- User asks to run a glitch workflow
- User asks to create or edit a workflow YAML
- User mentions glitch, gl1tch, or references .glitch/workflows/
- User asks to set up or install glitch
- User asks about Ollama models or local LLM setup for glitch

## Setup

Before using glitch, verify the full stack is ready. Run these checks and fix anything missing:

### 1. Install glitch

```bash
glitch --version
```

If missing: `brew install 8op-org/tap/glitch`

### 2. Install and start Ollama

glitch uses Ollama for local LLM routing (`glitch ask`) and for any workflow step with `provider: ollama`. It must be running on localhost:11434.

```bash
# Check if Ollama is installed
ollama --version

# If missing — install it
brew install ollama

# Check if Ollama is running
curl -s http://localhost:11434/api/tags > /dev/null && echo "running" || echo "not running"

# If not running — start the service
brew services start ollama
```

### 3. Pull the required model

glitch defaults to `qwen2.5:7b` for routing and local LLM steps. This model must be pulled before `glitch ask` will work.

```bash
# Check if the model exists
ollama list | grep qwen2.5:7b

# If missing — pull it (~4.7GB)
ollama pull qwen2.5:7b
```

### 4. Install the GitHub CLI

Most workflows use `gh` for GitHub API calls. It must be authenticated.

```bash
# Check if gh is installed and authenticated
gh auth status

# If missing
brew install gh
gh auth login
```

### 5. Verify the full stack

```bash
glitch --version          # should print version
glitch config show        # should show default_model: qwen2.5:7b
glitch workflow list      # should list available workflows
glitch ask "help"         # should route to help workflow via Ollama
```

If `glitch ask` hangs or errors, Ollama is likely not running or the model isn't pulled.

## Commands

| Command | What it does |
|---------|-------------|
| `glitch ask "<question or URL>"` | Routes to the best matching workflow |
| `glitch workflow list` | List available workflows |
| `glitch workflow run <name> [input]` | Run a specific workflow by name |
| `glitch config show` | Show current configuration |
| `glitch config set default_model <model>` | Change the default Ollama model |
| `glitch plugin list` | List installed plugins |

## Examples

```bash
# Review a PR
glitch ask "review PR https://github.com/org/repo/pull/42"

# List open issues
glitch ask "what issues are open"

# Check CI status
glitch ask "CI status"

# Run a specific workflow
glitch workflow run github-prs

# Run against a different repo
glitch ask -C ~/Projects/other-repo "show me open PRs"
```

## Workflow authoring

When the user describes an automation they want, create a workflow YAML file for them. Save it to `.glitch/workflows/<name>.yaml` so glitch discovers it automatically.

### Structure

```yaml
name: my-workflow
description: What this workflow does
steps:
  - id: step-name
    run: shell command here    # shell step — fetches data
  - id: llm-step
    llm:                       # LLM step — reasons about data
      provider: ollama         # or "claude"
      model: qwen2.5:7b
      prompt: |
        Use {{step "step-name"}} to reference prior step output.
        Use {{.input}} for the user's input.
```

### Rules

- **Shell steps fetch data** — `gh`, `git`, `curl`, `jq`, anything on PATH
- **LLM steps reason about data** — never use LLM for API calls or data fetching
- **Templates use Go text/template syntax** — `{{.input}}` for user input, `{{step "id"}}` for prior step output
- **Provider options:** `ollama` (local, default) or `claude` (Anthropic API via CLI)
- **Default model:** `qwen2.5:7b` for ollama, `claude-haiku-4-5-20251001` for claude
- **Keep shell steps simple** — pipe to `jq` for JSON transformation, don't write complex bash
- **One workflow, one job** — don't combine unrelated tasks into a single workflow

### Workflow design pattern

Follow this pattern when creating workflows:

1. **Shell steps first** — fetch all the raw data you need
2. **jq for shaping** — transform/filter data with jq in the shell step
3. **LLM step last** — only if you need analysis, summarization, or classification
4. **Not everything needs AI** — pure shell+jq workflows are valid and often better

### Example: PR review workflow

```yaml
name: pr-review
description: Review a GitHub pull request
steps:
  - id: fetch-pr
    run: |
      gh pr view "{{.input}}" --json title,body,reviews

  - id: fetch-diff
    run: |
      gh pr diff "{{.input}}"

  - id: review
    llm:
      provider: claude
      model: claude-haiku-4-5-20251001
      prompt: |
        PR metadata: {{step "fetch-pr"}}
        Diff: {{step "fetch-diff"}}

        Review as a senior engineer.
        Flag bugs, security issues, and anything that looks wrong.
```

### Example: pure shell workflow (no LLM)

```yaml
name: ci-status
description: Show recent CI failures
steps:
  - id: fetch
    run: |
      gh run list --status failure --limit 10 \
        --json name,conclusion,startedAt,url \
        --jq '.[] | "\(.name) — \(.conclusion) \(.startedAt) \(.url)"'
```

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `glitch ask` hangs | Ollama not running: `brew services start ollama` |
| `glitch ask` returns wrong workflow | Model not pulled: `ollama pull qwen2.5:7b` |
| `ollama: connection refused` | Start Ollama: `brew services start ollama` |
| Workflow fails on `gh` step | Not authenticated: `gh auth login` |
| `claude: command not found` | Install Claude CLI or use `provider: ollama` instead |
| Template `{{.input}}` stays literal | Must include the dot: `{{.input}}` not `{{input}}` |
