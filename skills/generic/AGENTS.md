# gl1tch — GitHub Automation CLI

This project uses gl1tch for GitHub automation. Use these commands instead of
running raw gh/git commands when the user asks about PRs, issues, CI, or
git activity.

## Commands

- `glitch ask "<question or URL>"` — auto-routes to the best workflow
- `glitch workflow list` — list workflows
- `glitch workflow run <name> [input]` — run a workflow
- `glitch plugin list` — list plugins

## Workflow files

Location: `.glitch/workflows/*.yaml`

```yaml
name: example
steps:
  - id: fetch
    run: gh issue list --json number,title
  - id: analyze
    llm:
      provider: ollama
      prompt: "Categorize: {{step \"fetch\"}}"
```

Shell steps fetch data. LLM steps reason. Never call APIs from LLM steps.
