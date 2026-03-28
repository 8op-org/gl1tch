## Context

The orcai binary has grown a long tail of subcommands inherited from earlier development phases — git UI, weather widget, ollama manager, session picker, kill, new, code. These are either superseded by the switchboard TUI or were never fully adopted. Meanwhile, two new friction points exist:

1. The `^spc j` jump window launches via `tmux display-popup … orcai _jump`, which creates a visual mismatch (its own palette load, no shared modal frame) and requires the external dispatch path in `main.go`.
2. The cron TUI runs in a dedicated `orcai-cron` tmux session (a separate OS process). When the user changes the active theme inside the switchboard, the cron TUI's in-memory registry is never notified — it only picks up the new theme on next startup because the active theme is persisted to disk but not broadcast cross-process.

## Goals / Non-Goals

**Goals:**
- Reduce the public CLI surface to commands that are actively used: `orcai` (no args), `cron`, `pipeline`, `agent`, `bridge`, `config`, `completion`, and internal `_opsx`.
- Remove all associated source files, tests, and helper binaries for pruned commands.
- Embed the jump window as a modal overlay inside the switchboard TUI so it shares the active theme, border, and overlay rendering infrastructure.
- Make the cron TUI detect active-theme changes on disk automatically, without restart.

**Non-Goals:**
- Rewriting the jump window's filtering or window-enumeration logic.
- Adding bus/socket cross-process pubsub for theme changes (overkill when disk polling at 5 s is sufficient and already available).
- Changing the `config` subcommand behaviour (left as-is pending a separate decision).

## Decisions

### 1. Disk polling for cron theme sync (vs. busd subscription)

**Decision**: Add a 5-second tick in the cron TUI that re-reads the active theme name from disk. If it differs from the last known name, reload the bundle and trigger a re-render.

**Alternatives considered**:
- *busd subscription*: The cron TUI would connect to the orcai bus socket and subscribe to `theme.changed`. More correct architecturally, but the busd socket path is resolved at bootstrap time and the cron process may start independently; the plumbing adds significant surface area for a problem that polling solves adequately.
- *inotify / FSEvents file watch*: OS-level file watch on `~/.config/orcai/themes/.active`. More responsive, but platform-specific and adds a dependency.

**Rationale**: Disk polling is already the cron TUI's source of truth for config (it re-reads `cron.yaml` on changes). A 5-second lag on theme switch is imperceptible in practice.

### 2. Jump window as switchboard modal (vs. keeping popup)

**Decision**: Embed `jumpwindow.Model` inside the switchboard TUI. When `^spc j` fires, the bootstrap keybinding sends a key event directly to the switchboard window (e.g. `tmux send-keys -t orcai:0 J`) that toggles the jump modal overlay, instead of spawning a popup.

**Alternatives considered**:
- *Keep `display-popup` but load theme from disk*: Simpler, but the popup still has its own alt-screen and doesn't feel native. Modal overlays with `lipgloss` are already used for help, theme picker, and job windows.
- *New tmux pane split inside switchboard window*: Avoids the key-routing complexity but disrupts layout and is harder to dismiss cleanly.

**Rationale**: The switchboard already has a modal overlay pattern (help modal, theme picker, agent runner). Reusing it for jump gives a coherent UX and removes the standalone dispatch path.

### 3. `sysop` and `welcome` removal

Both commands run `switchboard.Run()` with no meaningful distinction. The `orcai` binary with no args calls `bootstrap.Run()`, which already launches the switchboard session. Keeping named aliases adds noise to `--help` without benefit.

**Decision**: Remove `sysop.go`, `welcome.go`, and the `orcai-sysop` / `orcai-welcome` helper binaries. The bootstrap path is the canonical entrypoint.

## Risks / Trade-offs

- **External scripts calling removed subcommands** → Documented as BREAKING in proposal; users relying on `orcai git` or `orcai weather` must update. Risk is low given these commands were niche.
- **Jump modal key routing** → Sending `tmux send-keys` to the switchboard window for an internal keypress is fragile if the user has remapped keys. The keybinding name (`jump-window`) should be user-overridable via `keybindings.yaml`. Use an unlikely default (e.g. `ctrl+j` in the switchboard's own keymap) that is distinct from terminal passthrough.
- **`_welcome` removal** → `cmd/orcai-welcome/main.go` wraps the bootstrap exec path for use as a tmux default command. Removing it requires ensuring `bootstrap.Run()` is invoked directly or via `orcai` with no args in all contexts where `orcai-welcome` was used.

## Migration Plan

1. Delete pruned `cmd/*.go` files and `cmd/orcai-*/` directories.
2. Update `main.go` dispatch switch to remove deleted commands and `_jump` / `_welcome` cases.
3. Update `internal/bootstrap/bootstrap.go` keybinding strings: replace `display-popup … orcai _jump` with `send-keys -t orcai:0 <jump-key>`; remove `ollama` popup binding.
4. Add jump modal to `internal/switchboard/`: new `jump_modal.go`, integrate into `switchboard.go` Update/View.
5. Add disk-poll tick to `internal/crontui/init.go` and wire into `Update()`.
6. Update `internal/keybindings/`: remove `launch-session-picker` default.
7. Run full test suite; update any tests referencing removed commands.
