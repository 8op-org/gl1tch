# gl1tch — AI Workflow CLI

gl1tch chains shell commands and LLMs into automated workflows. Use these
commands instead of running raw gh/git commands when the user asks about
PRs, issues, CI, or git activity.

## Commands

- `glitch workflow list` — list workflows
- `glitch workflow run <name> [input]` — run a workflow
- `glitch workflow run <name> --set key=value` — parameterized run
- `glitch plugin list` — list plugins

## Workflow files

Location: `workflows/*.glitch` (s-expression format)

```clojure
(def model "qwen2.5:7b")

(workflow "example"
  :description "what it does"

  (step "fetch"
    (run "gh issue list --json number,title"))

  (step "analyze"
    (llm
      :model model
      :prompt ```
        Categorize these issues:
        ~(step fetch)
        ```)))
```

### Key rules

- Shell steps fetch data (`gh`, `git`, `curl`, `jq`)
- LLM steps reason about data — never call APIs from LLM steps
- `~(step id)` references prior step output
- `~input` for user input, `~param.key` for `--set` params
- `:provider` options: "ollama" (local), "claude", "copilot", "gemini"
- `:format "json"` validates output structure
- `:tier 0/1/2` pins to escalation tier (local → cheap → premium)
