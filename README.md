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
| `^spc d` | Detach session |
| `^spc r` | Reload GL1TCH (picks up new binary) |
| `^spc q` | Quit |

## Themes

Press `T` or `/themes` to open the theme picker. Custom themes live in `~/.config/glitch/themes/`.

| Theme | Variant | Accent |
|---|---|---|
| **Dracula** | dark | `#ff79c6` |
| **GL1TCH** | dark | `#bd93f9` |
| **Tokyo Night** | dark | `#7aa2f7` |
| **Tokyo Night Storm** | dark | `#7aa2f7` |
| **Tokyo Night Moon** | dark | `#82aaff` |
| **Kanagawa** | dark | `#7e9cd8` |
| **Kanagawa Dragon** | dark | `#7fb4ca` |
| **Nord** | dark | `#88c0d0` |
| **Nordic** | dark | `#81a1c1` |
| **Catppuccin Mocha** | dark | `#cba6f7` |
| **Catppuccin Macchiato** | dark | `#8aadf4` |
| **Catppuccin Frappé** | dark | `#8caaee` |
| **Catppuccin Latte** | light | `#1e66f5` |
| **Rose Piné** | dark | `#c4a7e7` |
| **Rose Piné Dawn** | light | `#286983` |
| **Gruvbox** | dark | `#458588` |
| **Gruvbox Light** | light | `#076678` |
| **Everforest Dark** | dark | `#83c092` |
| **Everforest Light** | light | `#8da101` |
| **One Dark** | dark | `#61afef` |
| **One Half Dark** | dark | `#61afef` |
| **One Half Light** | light | `#0184bc` |
| **One Light** | light | `#4078f2` |
| **Solarized Dark** | dark | `#268bd2` |
| **Solarized Light** | light | `#268bd2` |
| **GitHub Dark** | dark | `#58a6ff` |
| **GitHub Light** | light | `#0969da` |
| **Night Owl** | dark | `#82aaff` |
| **Night Owl Light** | light | `#4876d6` |
| **Nightfly** | dark | `#82aaff` |
| **Moonfly** | dark | `#74b2ff` |
| **Ayu Dark** | dark | `#39bae6` |
| **Ayu Light** | light | `#399ee6` |
| **Iceberg Dark** | dark | `#84a0c6` |
| **Iceberg Light** | light | `#2d539e` |
| **Zenbones Dark** | dark | `#67afc1` |
| **Zenbones Light** | light | `#407a9e` |
| **Tomorrow Night** | dark | `#81a2be` |
| **Tomorrow** | light | `#4271ae` |
| **Seoul256 Dark** | dark | `#5f87af` |
| **Seoul256 Light** | light | `#5f87af` |
| **Cobalt2** | dark | `#ffc600` |
| **Shades of Purple** | dark | `#fad000` |
| **Panda** | dark | `#ff75b5` |
| **Sonokai** | dark | `#76cce0` |
| **Deus** | dark | `#7aa2f7` |
| **Lucario** | dark | `#00bff3` |
| **Apprentice** | dark | `#87af87` |
| **Jellybeans** | dark | `#8197bf` |
| **Srcery** | dark | `#68a8e4` |
| **Gotham** | dark | `#195466` |
| **Miasma** | dark | `#78824b` |
| **Borland** | dark | `#ffff55` |
| **Tender** | dark | `#b3deef` |
| **Noctis Lux** | light | `#2d6b97` |

## Releasing

Use the `/release` skill in Claude Code to cut a new release:

```
/release
```

The skill guides you through: branch → PR → merge to protected main → changelog curation → semver tag → GitHub Actions release build.

Requires: `gh` CLI authenticated, `goreleaser` installed, `task` installed.
