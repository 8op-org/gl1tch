## Why

Each TUI panel's context-sensitive action hints are rendered in a single shared bottom bar (`viewBottomBar` / `viewHintBar`) that lives outside every panel — panels are passive boxes that receive no footer of their own. This makes it harder to see which panel owns which actions at a glance, and duplicates focus-dispatching logic across panels while making it impossible to show per-panel menus inside the panel boundaries.

## What Changes

- Each switchboard panel (Launcher, Agent, Signal Board, Activity Feed, Inbox, Cron) gains an embedded footer row rendered inside its own box, showing its action hints only when that panel is focused.
- The cron TUI's two panes (Jobs, Logs) each gain an embedded footer row rendered inside their own box, showing hints only for the active pane.
- A shared `panelrender.HintBar(hints []Hint, width int, pal palette) string` helper is introduced so all panels render hints the same way without duplicating hint-formatting code.
- The global `viewBottomBar` in `switchboard.go` and `viewHintBar` in `crontui/view.go` are removed (or reduced to a no-op stub) once all panels carry their own footers.

## Capabilities

### New Capabilities
- `panel-action-menu`: Each panel/pane renders its own single-row action-hint footer inside its border, visible only when the panel is focused; a shared `HintBar` helper in `panelrender` provides the formatted row.

### Modified Capabilities

## Impact

- `internal/switchboard/switchboard.go` — `viewBottomBar`, all panel view functions (`buildLauncherSection`, `buildAgentSection`, `buildInboxSection`, `buildCronSection`, `viewActivityFeed`)
- `internal/switchboard/signal_board.go` — `buildSignalBoard`
- `internal/crontui/view.go` — `View`, `viewJobList`, `viewLogPane`, `viewHintBar`
- `internal/panelrender/panelrender.go` — new `HintBar` export
- Panel height calculations in switchboard and cron TUI will need adjustment to account for the footer row
