# GL1TCH — your AI, your terminal, your rules

```
  _____ _     __ _______ _____ _    _
 / ____| |   /_ |__   __/ ____| |  | |
| |  __| |    | |  | | | |    | |__| |
| | |_ | |    | |  | | | |    |  __  |
| |__| | |____| |  | | | |____| |  | |
 \_____|______|_|  |_|  \_____|_|  |_|
```

GL1TCH is a tmux-native AI workspace. Run pipelines, coordinate agents, and keep everything visible in one terminal session — your GL1TCH, your way.

## Install

**macOS**
```bash
brew install 8op-org/tap/glitch
```

**Linux** — download the release tarball and run the install script:
```bash
curl -L https://github.com/8op-org/gl1tch/releases/latest/download/glitch_linux_amd64.tar.gz \
  | tar xz -C /tmp/glitch-install
cd /tmp/glitch-install && ./install.sh
```

Use `glitch_linux_arm64.tar.gz` on ARM64. The script installs `glitch` and all provider binaries to `~/.local/bin` — make sure it's on your `$PATH`.

## Requirements

- **tmux** — `brew install tmux` / `apt install tmux`
- **An AI provider** — Ollama (local, free) or Claude CLI (cloud)

## Getting Started

### Launch

```bash
glitch
```

This creates (or reattaches to) your GL1TCH session inside tmux. Detach and reconnect anytime — your jobs keep running.

On first launch, run `/init` to walk through provider setup.

### Your First Pipeline

```
/pipeline create a pipeline that summarizes my git log every morning
```

GL1TCH generates the pipeline, shows you a preview, and asks before saving. No YAML editing required.

## Commands

Type `/help` in the console to see all commands. Common ones:

| Command | What it does |
|---|---|
| `/init` | First-run wizard — configure providers and preferences |
| `/models` | Pick a provider and model |
| `/model [name]` | Switch provider or model inline |
| `/pipeline [name]` | Run a pipeline, or ask GL1TCH to build one |
| `/rerun [name]` | Rerun a pipeline by name |
| `/cron` | Get help scheduling recurring jobs |
| `/brain [query]` | Search your notes, or start an interactive brain session |
| `/prompt [name]` | Load or build a system prompt with AI |
| `/session [new\|name\|#]` | Manage chat sessions |
| `/s` | Shorthand for `/session` |
| `/cwd [path]` | Set working directory |
| `/terminal [cmd]` | Open a split pane, optionally running a command |
| `/themes` | Open the theme picker |
| `/clear` | Clear chat history |
| `/help` | Full command list |
| `/quit` | Exit GL1TCH |

## Navigation

| Key | Action |
|---|---|
| `tab` | Cycle focus between panels |
| `j` / `k` | Move selection up/down |
| `enter` | Launch / open selected item |
| `esc` | Back / close overlay |
| `T` | Open theme picker |

### Chord Shortcuts

Press `^spc` (ctrl+space) then a key:

| Chord | Action |
|---|---|
| `^spc h` | Help screen |
| `^spc t` | Switch to console |
| `^spc m` | Theme picker |
| `^spc c` | New window |
| `^spc d` | Detach session |
| `^spc r` | Reload GL1TCH (picks up new binary) |
| `^spc q` | Quit |
| `^spc [` / `]` | Previous / next window |
| `^spc x` / `X` | Kill pane / window |
| `^spc a` | Jump to GL1TCH assistant |

## Themes

Press `T` or `/themes` to open the theme picker.

Built-in themes: **Dracula**, **Nord**, **Catppuccin Mocha**, **Tokyo Night**, **Rose Piné**, **Solarized Dark**, **Kanagawa**.

Custom themes live in `~/.config/glitch/themes/`.

## Releasing

Use the `/release` skill in Claude Code to cut a new release:

```
/release
```

The skill guides you through: branch → PR → merge to protected main → changelog curation → semver tag → GitHub Actions release build.

Requires: `gh` CLI authenticated, `goreleaser` installed, `task` installed.
