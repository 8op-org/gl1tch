## Why

The CLI has accumulated dead commands (git, weather, ollama, picker, new, kill, code) that no longer reflect how orcai is used — the switchboard is the single entrypoint, and internal tooling like the jump window deserves first-class UI treatment rather than a standalone popup that doesn't share theme or modal styling. Additionally, the cron TUI runs in a separate process and misses live theme changes until restart, breaking visual consistency.

## What Changes

- **Remove dead CLI subcommands**: `git`, `weather`, `ollama`, `picker`, `new`, `kill`, `code`, `sysop`, `welcome` are removed from the cobra tree and their source files deleted. The `orcai` binary with no arguments remains the single switchboard entry point.
- **Consolidate `_welcome` / `orcai-welcome`**: Remove the `_welcome` dispatch case and the `orcai-welcome` helper binary; the bootstrap path already handles session launch.
- **Integrate jump window as switchboard modal**: Replace the `tmux display-popup … orcai _jump` keybinding with an in-process modal overlay in the switchboard TUI. The `_jump` dispatch in `main.go` and the `jumpwindow.Run()` standalone entry point are removed; the jump model renders inside the existing modal layer and inherits the active theme.
- **Cron TUI live theme sync**: The cron TUI polls the active theme from disk on a short tick. When the on-disk active theme differs from the current bundle the TUI re-loads and re-renders with the new palette — no restart required.
- **BREAKING**: `orcai sysop`, `orcai welcome`, `orcai git`, `orcai weather`, `orcai ollama`, `orcai picker`, `orcai new`, `orcai kill`, `orcai code` are no longer valid invocations.

## Capabilities

### New Capabilities

- `jump-window-modal`: Jump window rendered as a first-class modal overlay inside the switchboard TUI, sharing theme, border, and overlay infrastructure with other modals.
- `cron-theme-live-sync`: Cron TUI detects active-theme changes on disk and re-renders without restart.

### Modified Capabilities

- `core-subcommand-dispatch`: Picker, sysop, and welcome are no longer registered cobra subcommands; the only entry points are `orcai` (no args) and the remaining subcommands (cron, pipeline, agent, bridge, config, completion).

## Impact

- `cmd/`: delete `git.go`, `weather.go`, `ollama.go`, `picker.go`, `new.go`, `new_test.go`, `kill.go`, `code.go`, `sysop.go`, `sysop_test.go`, `welcome.go`, `welcome_test.go`; remove `_welcome` case from `main.go`; remove `_jump` case from `main.go`; update the command list in `main.go`'s dispatch switch.
- `cmd/orcai-picker/`, `cmd/orcai-sysop/`, `cmd/orcai-welcome/`: delete all three helper binary directories.
- `internal/bootstrap/`: update the `^spc j` keybinding to send a key event to the switchboard window instead of `display-popup … orcai _jump`; remove `ollama` popup binding.
- `internal/switchboard/`: add jump modal model and integrate into overlay/key handler pipeline.
- `internal/crontui/`: add a disk-poll tick alongside the existing theme-channel subscription so theme changes propagate cross-process.
- `internal/jumpwindow/`: package remains but `Run()` entry point is deprecated in favour of the switchboard-embedded modal.
