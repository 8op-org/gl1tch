## 1. Export Box Helpers from panelrender

- [x] 1.1 Add exported `BoxTop`, `BoxBot`, `BoxRow` functions to `internal/panelrender/panelrender.go` (matching the existing private switchboard wrappers)
- [x] 1.2 Update the private `boxTop`, `boxBot`, `boxRow` wrappers in `internal/switchboard/switchboard.go` to delegate to the new exported panelrender functions
- [x] 1.3 Update `internal/crontui` view code to use `panelrender.BoxTop` / `BoxBot` / `BoxRow` directly (if not already)
- [x] 1.4 Run `go build ./...` to confirm no compilation errors

## 2. busd Theme Subscriber Helper

- [x] 2.1 Create `internal/tuikit/` package with a `ThemeSubscribeCmd` function that dials busd, registers with `subscribe: ["theme.changed"]`, and returns a `tea.Cmd` that blocks on the next event
- [x] 2.2 Define a `ThemeChangedMsg` struct in `internal/tuikit` carrying the new theme name (to be returned by the cmd)
- [x] 2.3 Handle dial failure gracefully: if the busd socket is unavailable, return a no-op `tea.Cmd` (nil) and log a warning
- [x] 2.4 Write unit tests for `ThemeSubscribeCmd` using a test busd daemon (follow existing `busd_test.go` patterns)

## 3. Wire switchboard to Publish theme.changed via busd

- [x] 3.1 In `internal/switchboard/switchboard.go`, locate the theme-picker apply path (where `registry.SetActive` is called)
- [x] 3.2 After `SetActive`, call `busd.Publish(themes.TopicThemeChanged, themes.ThemeChangedPayload{Name: name})` on the daemon instance
- [x] 3.3 Confirm the busd daemon reference is accessible in the theme-picker apply path (it is started at boot); plumb it through if needed
- [x] 3.4 Run the test suite to confirm no regressions: `go test ./internal/switchboard/...`

## 4. Replace File Poll in crontui with busd Subscription

- [x] 4.1 In `internal/crontui/init.go`, replace `pollThemeFile()` cmd in `Init()` with `tuikit.ThemeSubscribeCmd(ctx)`
- [x] 4.2 In `internal/crontui/update.go`, replace the `themeFilePollMsg` handler with a `tuikit.ThemeChangedMsg` handler; re-issue `tuikit.ThemeSubscribeCmd` after each message
- [x] 4.3 Remove `pollThemeFile` function and `themeFilePollMsg` type from `internal/crontui/`
- [x] 4.4 Run `go test ./internal/crontui/...` to confirm no regressions

## 5. Apply Same Pattern to Other Sub-TUIs

- [x] 5.1 Audit `internal/jumpwindow` and any other standalone sub-TUIs for file-poll theme detection; replace with `tuikit.ThemeSubscribeCmd` if found
- [x] 5.2 Confirm `internal/gitui` (if it exists as standalone TUI) uses busd subscription or in-process channel — not file polling

## 6. Validation

- [x] 6.1 Run full test suite: `go test ./...`
- [ ] 6.2 Manual smoke test: launch `orcai`, open `orcai cron` in a second pane, switch themes in switchboard — confirm crontui updates within 1 second
- [ ] 6.3 Manual smoke test: launch `orcai cron` without switchboard running — confirm it starts and displays default theme without error

