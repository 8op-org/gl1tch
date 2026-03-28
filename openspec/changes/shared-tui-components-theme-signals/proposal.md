## Why

Sub-TUIs like `crontui` rely on a fragile 5-second file-poll to detect cross-process theme changes, meaning in-session theme switches in switchboard take up to 5 seconds to appear — and any future sub-command will need to re-implement this same workaround. The busd event bus already exists and is already the declared mechanism for `theme.changed` events; it just isn't wired into TUI components yet.

## What Changes

- **Wire busd `theme.changed` into all TUI components** — replace the file-poll hack in `crontui` (and any future sub-TUI) with a busd subscriber so cross-process theme changes are delivered in real time.
- **Extract shared TUI shell** — the private `boxTop` / `boxBot` / `boxRow` wrappers and in-switchboard theme-plumbing boilerplate are moved (or exposed) so every sub-command starts from the same foundation without copying code.
- **busd client helper for TUI processes** — a small `internal/tuikit` (or `internal/busd` extension) package provides a ready-to-use `tea.Cmd` that subscribes to `theme.changed` and returns a typed msg, eliminating the need for each TUI to hand-roll the same subscription loop.
- **Remove file-poll fallback** once busd subscription is the primary mechanism — `pollThemeFile` in crontui is deleted.
- **switchboard publishes `theme.changed` via busd** when the user picks a new theme, instead of (or in addition to) writing the active-theme file.

## Capabilities

### New Capabilities

- `tui-theme-bus-subscriber`: A reusable `tea.Cmd` + helper that any BubbleTea model can use to subscribe to `theme.changed` events over busd, receiving typed theme-change messages without polling.

### Modified Capabilities

- `status-bar-session-controls`: The switchboard theme-picker now publishes `theme.changed` to busd on activation; this changes observable behavior for any listener on the bus.

## Impact

- `internal/busd` — new exported helper(s) for connecting a TUI process as a subscriber
- `internal/themes` — unchanged types; `TopicThemeChanged` constant already correct
- `internal/crontui` — `pollThemeFile` / `themeFilePollMsg` removed; replaced with busd subscription cmd
- `internal/switchboard` — theme-picker apply path publishes to busd
- `internal/panelrender` — `boxTop` / `boxBot` / `boxRow` wrappers promoted from switchboard private functions to exported panelrender functions so crontui and future sub-TUIs can use them without re-implementing
- Any future sub-command (`gitui`, `jumpwindow`, etc.) automatically gets live theme updates by using the new busd subscriber cmd
