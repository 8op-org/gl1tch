---
title: "Your First Pipeline"
description: "Install gl1tch and run your first AI-powered automation in under five minutes."
order: 1
---

gl1tch runs AI-powered automations called pipelines. This page gets you from zero to a working pipeline — everything on this page has been tested on a fresh machine.

## Requirements

- **tmux** — gl1tch runs inside a tmux session
- **An AI provider** — Ollama (local, free) or the Claude CLI (cloud)
- **20GB+ disk** if using Ollama (the default coding model is ~5GB; `llama3.2` is ~2GB)
- **4GB+ RAM** to run a local model

Install tmux if you don't have it:

```bash
brew install tmux          # macOS
sudo apt install tmux      # Ubuntu / Debian
```


## Install

```bash
brew install 8op-org/tap/glitch
```

**Linux** — download the release tarball and run the install script:

```bash
curl -L https://github.com/8op-org/gl1tch/releases/latest/download/glitch_linux_amd64.tar.gz \
  | tar xz -C /tmp/glitch-install
cd /tmp/glitch-install && ./install.sh
```

Use `glitch_linux_arm64.tar.gz` on ARM64 machines. The script installs `glitch` and all provider binaries to `~/.local/bin` — make sure it's on your `$PATH`.


## Set up an AI provider

You need at least one before you can run pipelines.

### Ollama (local, no API cost)

```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3.2          # ~2GB, good all-around
```

On Linux, Ollama installs as a systemd service and starts automatically. On macOS, start it with `ollama serve`.

### Claude CLI (cloud)

```bash
brew install claude           # macOS
# or download from https://claude.ai/download
claude                        # authenticate on first run
```


## Initialize your workspace

```bash
glitch config init
```

This writes your default config, executor wrappers, and the `wf-git-pulse` example pipeline to `~/.config/glitch`. You only need to run this once.

Verify your provider is reachable:

```bash
glitch model          # prints the best available provider/model
```


## Run your first pipeline

```bash
glitch pipeline run wf-git-pulse
```

This pipeline shows what's happening in the current git repo — recent commits, diff stat, and working tree status. It runs entirely with shell steps, no AI model required.

```
[step:log] status:done
[step:diffstat] status:done
[step:status] status:done

=== recent commits ===
3a9f1c2 feat: add pipeline resume support
8d2e047 fix: brain context injection on retry
...
```


## Open the console

```bash
glitch
```

This opens your workspace in a tmux session. Everything available from the command line is here — plus conversation history, brain context, and the inline docs viewer (`/docs`).


## Ask gl1tch something

From the console or the terminal:

```bash
glitch ask "summarize my last 5 commits"
```

gl1tch routes the request to a pipeline if one matches, or handles it directly.


## Write your own pipeline

Save this as `~/.config/glitch/pipelines/git-summary.pipeline.yaml`:

```yaml
name: git-summary
version: "1"

steps:
  - id: log
    executor: shell
    vars:
      cmd: "git log --oneline -10"

  - id: summarize
    executor: ollama
    model: llama3.2
    needs: [log]
    prompt: |
      Here are the last 10 git commits:

      {{steps.log.output}}

      Summarize what's been worked on in 3-4 sentences. Be specific about what changed.
```

Run it:

```bash
glitch pipeline run git-summary
```

Swap `executor: ollama` + `model: llama3.2` for `executor: claude` + `model: claude-haiku-4-5-20251001` if you're using Claude instead.


## Next steps

- [Pipelines](/docs/pipelines/pipelines) — what's inside a pipeline and how steps connect
- [Executors](/docs/pipelines/executors) — all available executors
- [Console](/docs/pipelines/console) — your gl1tch workspace in detail
- [Examples](/docs/pipelines/examples) — ready-to-run pipelines for real developer workflows
