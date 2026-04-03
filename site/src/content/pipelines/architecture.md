---
title: "Architecture"
description: "Trace what happens when you run a pipeline — from command to output."
order: 10
---

gl1tch is a single binary that runs in your terminal. Here's what happens from the moment you type `glitch pipeline run` to the moment you see output.

## What Happens When You Run a Pipeline

```text
You type:
  glitch pipeline run code-review

         │
         ▼
  gl1tch reads your pipeline YAML
  ~/.config/glitch/pipelines/code-review.pipeline.yaml

         │
         ▼
  The runner builds a dependency graph from your steps
  (which steps depend on which other steps)

         │
         ▼
  Steps run in order — parallel where possible
  Each step:
    1. Resolves template expressions ({{steps.x.output}})
    2. Calls the executor (ollama, claude, shell, etc.)
    3. Captures the output
    4. Stores it locally

         │
         ▼
  Results appear in your terminal
  Full run history written to ~/.local/share/glitch/glitch.db
```

That's it. No hidden agents. No cloud orchestration. The work happens on your machine.

## Your Workspace

When you run `glitch` without a subcommand, it starts a tmux session named `glitch` and opens the console — a full-screen chat interface where you talk to your assistant, launch pipelines, and review results.

```text
┌────────────────────────────────────────────────────────────┐
│  GL1TCH                                            [header] │
│                                                             │
│  gl1tch: What do you need?                                  │
│                                                             │
│  you: run code-review on the current branch                 │
│                                                             │
│  gl1tch: Running code-review…  ✓ done                       │
│  [result output]                                            │
│                                                             │
│ ╭─────────────────────────────────────────────────────────╮ │
│ │ > type a message or /command                            │ │
│ │                                             agent │ p │ │ │
│ ╰─────────────────────────────────────────────────────────╯ │
└────────────────────────────────────────────────────────────┘
```

The send panel at the bottom lets you select a saved agent, attach a saved prompt, or attach a named pipeline to your message.

### Views inside the workspace

The workspace has three panels. Navigate between them with `tab`. Use keyboard shortcuts from the focused panel:

| Panel | How to focus | What it shows |
|-------|-------------|---------------|
| Chat | default / `tab` | Conversation with your assistant |
| Inbox | `tab` to cycle | All pipeline runs — select one and press `enter` for full output |
| Signal board | `s` from any panel | Live status of running and recent jobs |

From the signal board, `enter` opens the inbox detail for the selected run. From inbox detail, `j`/`k` scrolls the output, `e` opens it in your `$EDITOR`, and `r` sends selected lines back to the agent as follow-up context.

## Where Things Live

| What | Where |
|------|-------|
| Your pipelines | `~/.config/glitch/pipelines/` |
| Your workflows | `~/.config/glitch/workflows/` |
| Your tool wrappers | `~/.config/glitch/wrappers/` |
| Your themes | `~/.config/glitch/themes/` |
| Run history + brain + prompts | `~/.local/share/glitch/glitch.db` |
| Trace logs | `~/.local/share/glitch/traces.jsonl` |
| Cron log | `~/.local/share/glitch/cron.log` |

Everything is on your disk. Nothing requires a network connection for core operation.

## How Scheduling Works

gl1tch runs a second background tmux session named `glitch-cron` alongside the main console. It watches your pipelines and fires the ones with a `cron` field on schedule:

```yaml
name: morning-summary
cron: "0 8 * * 1-5"   # 8am, Monday–Friday
steps:
  ...
```

The cron session starts automatically when you open gl1tch. You can also interact with it directly:

```bash
glitch cron list    # see scheduled pipelines and next run times
glitch cron logs    # tail the cron log
```

## How the Brain Works

The brain is a local store of context built up from your pipeline runs. When a step has `write_brain: true`, its output is indexed and stored. On future runs with `--brain`, gl1tch retrieves relevant context and injects it into your prompts.

Your workspace learns over time without you having to re-explain things.

## Viewing Pipeline Results

After a run finishes, results land in the inbox panel. You can also pull them up directly:

```bash
glitch pipeline results              # most recent run
glitch pipeline results git-digest   # most recent run for a named pipeline
glitch pipeline results git-digest --limit 3  # last 3 runs
```

From inside the workspace, ask gl1tch directly:

```text
show me what git-digest found
```

gl1tch will call `glitch pipeline results git-digest` and display the output in chat. To see it in a separate pane:

```text
/terminal glitch pipeline results git-digest
```

## How glitch ask Works

When you run `glitch ask "summarize my open PRs"`, gl1tch:

1. Routes your prompt to a matching pipeline using your local model.
2. If a match is found, runs that pipeline with your prompt as input.
3. If no match is found, either generates a pipeline on the fly (with your confirmation) or responds directly.

The routing decision stays local. Cloud models are only called if the matched pipeline explicitly uses one.

## See Also

- [Philosophy](/docs/pipelines/philosophy) — why gl1tch works the way it does
- [Pipeline YAML Reference](/docs/pipelines/yaml-reference) — writing your first pipeline
- [Executors](/docs/pipelines/executors) — what runs inside each step
- [CLI Reference](/docs/pipelines/cli-reference) — every command and flag
