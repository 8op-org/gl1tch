---
name: gl1tch
description: Use the gl1tch CLI to automate GitHub tasks, run workflows, and compose shell+LLM pipelines. Invoke when the user asks to review PRs, triage issues, check CI, or run any glitch workflow.
---

# gl1tch — GitHub Automation CLI

## When to use

- User asks to review a PR, check issues, or see CI status
- User asks to run a glitch workflow
- User asks to create or edit a workflow YAML
- User mentions glitch, gl1tch, or references .glitch/workflows/

## Commands

| Command | What it does |
|---------|-------------|
| `glitch ask "<question or URL>"` | Routes to the best matching workflow |
| `glitch workflow list` | List available workflows |
| `glitch workflow run <name> [input]` | Run a specific workflow by name |
| `glitch config show` | Show current configuration |
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

Workflows live in `.glitch/workflows/*.yaml`. Structure:

```yaml
name: my-workflow
description: What this workflow does
steps:
  - id: step-name
    run: shell command here    # shell step
  - id: llm-step
    llm:                       # LLM step
      provider: ollama         # or "claude"
      model: qwen2.5:7b
      prompt: |
        Use {{step "step-name"}} to reference prior step output.
        Use {{.input}} for the user's input.
```

Key rules:
- Shell steps fetch data (`gh`, `git`, `curl`, etc.)
- LLM steps reason about data — never use LLM for API calls
- Templates use Go text/template syntax: `{{.param.input}}` or `{{step "id"}}`
- Provider options: "ollama" (local, default) or "claude" (Anthropic API)

## Installation check

Before using glitch, verify it's installed:
```bash
glitch --version
```

If not installed: `brew install 8op-org/tap/glitch`
