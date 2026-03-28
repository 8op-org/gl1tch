## 1. Shared HintBar Helper

- [x] 1.1 Define `Hint` struct (`Key`, `Desc string`) in `internal/panelrender/panelrender.go`
- [x] 1.2 Implement `HintBar(hints []Hint, width int, pal Palette) string` in `panelrender` — accent key, dim desc, ` · ` separator, padded to width, returns `""` for empty hints
- [x] 1.3 Define `Palette` interface (or accept concrete type) that exposes `Accent`, `Dim`, `BG` colors needed by `HintBar`

## 2. Switchboard Panels — Embed Footer

- [x] 2.1 Update `buildLauncherSection` to call `HintBar` with launcher hints and append footer inside the panel; subtract 1 from content height when focused
- [x] 2.2 Update `buildAgentSection` to call `HintBar` with agent hints and append footer inside the panel; subtract 1 from content height when focused
- [x] 2.3 Update `buildSignalBoard` to call `HintBar` with signal-board hints (search mode vs normal mode) and append footer inside the panel; subtract 1 from content height when focused
- [x] 2.4 Update `viewActivityFeed` to call `HintBar` with feed hints and append footer inside the panel; subtract 1 from content height when focused
- [x] 2.5 Update `buildInboxSection` to call `HintBar` with inbox hints and append footer inside the panel; subtract 1 from content height when focused
- [x] 2.6 Update `buildCronSection` (embedded cron panel) to call `HintBar` with cron-panel hints and append footer inside the panel; subtract 1 from content height when focused

## 3. Cron TUI — Embed Footer

- [x] 3.1 Update `viewJobList` to call `HintBar` with jobs-pane hints (filter mode vs normal mode) and append footer inside the pane; subtract 1 from content height when `activePane == 0`
- [x] 3.2 Update `viewLogPane` to call `HintBar` with logs-pane hints and append footer inside the pane; subtract 1 from content height when `activePane == 1`

## 4. Remove Global Bottom Bars

- [x] 4.1 Delete `viewBottomBar` from `internal/switchboard/switchboard.go` and remove its call site in `View`
- [x] 4.2 Delete `viewHintBar` from `internal/crontui/view.go` and remove its call site in `View`
- [x] 4.3 Update switchboard `View` height arithmetic: remove the `- 1` previously reserved for the global bottom bar; verify no layout regressions

## 5. Verification

- [x] 5.1 Build (`go build ./...`) with no errors
- [x] 5.2 Manually verify each switchboard panel shows its hint footer only when focused and hides it when unfocused
- [x] 5.3 Manually verify cron TUI pane footers appear only for the active pane
- [x] 5.4 Verify no duplicate hints (global bar gone, per-panel footer present)
- [x] 5.5 Verify panel content does not clip or overflow after height arithmetic changes
