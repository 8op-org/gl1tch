---
title: "Your First Pipeline"
description: "Install gl1tch and run your first AI-powered automation in under five minutes."
order: 1
---

gl1tch runs AI-powered automations called pipelines. You tell it what you want — it routes to the right pipeline, or builds one on the spot. This page gets you from zero to a working pipeline in under five minutes.

## Install

```bash
go install github.com/8op-org/gl1tch/cmd/glitch@latest
```

You need Go 1.22+ and at least one AI provider: [Ollama](https://ollama.ai) running locally, or the [Claude CLI](https://claude.ai/download) authenticated. No Docker, no cloud account required.

## Run your first pipeline

gl1tch ships with `wf-git-pulse` — a pipeline that shows what's happening in any git repo right now:

```bash
glitch pipeline run wf-git-pulse
```

```
[pipeline] starting: wf-git-pulse
[step:pulse] status:running
[step:pulse] status:done

=== recent commits ===
067ce08 feat(console): mud-chat-reply signal handler
68a8da1 feat(model): add glitch model subcommand for plugin model discovery
7cc9125 Merge pull request #40 from 8op-org/feature/router-improvements
389150d feat(router): five intent routing improvements with full test coverage
aa2faf2 chore: delete dead EditorPanel from buildershared

=== diff stat since last commit ===
 internal/console/signal_handlers.go | 38 +++++++++++++++++++++++++++++++++++++
 1 file changed, 38 insertions(+)

=== untracked / modified ===
 M site/src/content/pipelines/quickstart.md
```

That's a real pipeline run — `git log`, `git diff --stat`, and `git status` chained together in one step.

## Add AI to the pipeline

Ask gl1tch to summarize the same commits. gl1tch defaults to Claude Haiku — fast, cheap, and capable enough for most pipeline steps:

```bash
glitch ask --provider claude "summarize my last 5 commits"
```

gl1tch fetches your commits with `git log`, passes them to Haiku, and streams the result:

```
[step:fetch] status:running
[step:fetch] status:done
[step:summarize] status:running
[step:summarize] status:done

**067ce08 — MUD chat reply handler**
Added in-game chat integration. When players mention "glitch" in MUD chat,
gl1tch generates a reply and publishes it back to the game's web chat UI.

**68a8da1 — glitch model subcommand**
New command that outputs the best available model in "provider/model" format.
Enables plugins to resolve the user's configured model without hardcoding names.

**389150d — Five intent routing improvements**
Major routing overhaul: removed a gate blocking natural-language invocations,
added fast-path extraction for cron expressions, and added near-miss clarification.
```

No model flag needed — gl1tch picks the cheapest model for the provider automatically.

> **Note:** Pipelines that use tools, run shell commands autonomously, or coordinate multi-agent workflows require a model with tool/function-calling support. Haiku handles summarization and generation steps well; for agentic tasks upgrade to `--model claude-sonnet-4-6` or `--model claude-opus-4-6`.

## Use Ollama instead

If you'd rather keep everything local:

```bash
glitch ask --provider ollama "summarize my last 5 commits"
```

gl1tch picks your default Ollama model (`qwen2.5:latest` unless you configure otherwise) and runs the same pipeline locally — no API key, no outbound traffic.

## Review a PR

Pass gl1tch a GitHub PR URL and it routes to `pr-review` automatically:

```bash
glitch ask "https://github.com/8op-org/gl1tch/pull/40"
```

```
[route] → pr-review (95%)
[step:fetch_diff] status:running
[step:fetch_comments] status:running
[step:fix] status:running
[step:fix] status:done
```

`pr-review` fetches the diff and reviewer comments, then produces corrected code. Requires `gh` authenticated.

## Open the console

For ongoing sessions — asking questions, running pipelines, switching between projects:

```bash
glitch
```

Everything available from the command line is available here, plus conversation history, brain context, and the inline docs viewer (`/docs`).

## Next steps

- [Pipelines](/docs/pipelines/pipelines) — What's inside a pipeline and how steps connect
- [Console](/docs/pipelines/console) — Your gl1tch workspace in detail
- [Brain](/docs/pipelines/brain) — How gl1tch remembers context across sessions
- [Examples](/docs/pipelines/examples) — Ready-to-run pipelines for real developer workflows
