---
title: "Your First Pipeline"
description: "Install gl1tch and run your first AI-powered automation in under five minutes."
order: 1
---

gl1tch runs AI-powered automations called pipelines. You tell it what you want — it routes to the right pipeline, or builds one on the spot. This page gets you from zero to a working pipeline in under five minutes.

## Install

```bash
brew install 8op-org/tap/glitch
```

Or build from source with Go 1.22+:

```bash
go install github.com/8op-org/gl1tch/cmd/glitch@latest
```

You also need at least one AI provider: [Ollama](https://ollama.ai) running locally, or the [Claude CLI](https://claude.ai/download) authenticated. No Docker, no cloud account required.

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

Ask gl1tch to summarize the same commits. You can type this directly in the console, or run it from your terminal — both work exactly the same way:

```
summarize my last 5 commits
```

From the terminal:

```bash
glitch ask --provider ollama "summarize my last 5 commits"
```

gl1tch fetches your commits with `git log`, passes them to your local model, and streams the result:

```
[step:fetch] status:running
[step:fetch] status:done
[step:summarize] status:running
[step:summarize] status:done

Recent commits added a signal handler for in-game chat interaction and a
subcommand for discovering plugin models. There were also improvements to
the router with enhanced test coverage, and dead code was cleaned up by
removing an unused EditorPanel class.
```

No model flag needed — gl1tch picks your default Ollama model automatically (`qwen2.5:latest` unless you configure otherwise).

> **Note:** Pipelines that use tools, run shell commands autonomously, or coordinate multi-agent workflows require a larger local model with tool/function-calling support. Smaller models handle summarization and generation steps well; for agentic tasks pull a capable model like `ollama pull qwen3:8b` and pass `--model qwen3:8b`.

## Use Claude if you prefer a cloud provider

```bash
glitch ask --provider claude "summarize my last 5 commits"
```

gl1tch picks Claude Haiku by default (the cheapest option). Pass `--model claude-sonnet-4-6` to upgrade.

## Review a PR

Pass gl1tch a GitHub PR URL — in the console or from the terminal:

```
https://github.com/8op-org/gl1tch/pull/40
```

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
