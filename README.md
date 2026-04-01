# GL1TCH — your AI, your terminal, your rules

```
  _____ _     __ _______ _____ _    _
 / ____| |   /_ |__   __/ ____| |  | |
| |  __| |    | |  | | | |    | |__| |
| | |_ | |    | |  | | | |    |  __  |
| |__| | |____| |  | | | |____| |  | |
 \_____|______|_|  |_|  \_____|_|  |_|
```

GL1TCH is a tmux-native AI workspace. It runs pipelines, coordinates AI agents, and keeps everything visible in one terminal session. Your GL1TCH is your own.

## Getting Started

### Launch

```
glitch
```

This creates (or reattaches to) your GL1TCH session. Everything runs inside tmux — you can detach and reconnect anytime.

### The Switchboard

The **Switchboard** (window 0) is your control panel. GL1TCH takes the full screen — talk to it to run pipelines, launch agents, check job status, or just get help.

### Navigation

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
| `^spc h` | This help screen |
| `^spc t` | Switch to Switchboard |
| `^spc m` | Theme picker |
| `^spc j` | Jump to any window |
| `^spc c` | New window |
| `^spc d` | Detach session |
| `^spc r` | Reload GL1TCH (picks up new binary) |
| `^spc q` | Quit |
| `^spc [` / `]` | Previous / next window |
| `^spc x` / `X` | Kill pane / window |
| `^spc a` | Jump to GL1TCH assistant |

### Pipelines

Pipelines live in `~/.config/glitch/pipelines/`. Each is a `.pipeline.yaml` file. Ask GL1TCH to run one, or use the pipeline launcher overlay.

### Themes

Press `T` in the Switchboard or `^spc m` to open the theme picker. Themes live in `~/.config/glitch/themes/`.

Built-in themes: **Dracula**, **Nord**, **Catppuccin Mocha**, **Tokyo Night**, **Rose Piné**, **Solarized Dark**, **Kanagawa**.

### Reconnecting

```
glitch
```

If a session is already running, this reattaches. Your jobs keep running while detached.

## Releasing

Use the `/release` skill in Claude Code to cut a new release:

```
/release
```

The skill guides you through the full flow: branch → PR → merge to protected main → changelog curation → semver tag → GitHub Actions release build.

Requires: `gh` CLI authenticated, `goreleaser` installed, `task` installed.
