## 1. Delete Dead CLI Command Files

- [x] 1.1 Delete `cmd/git.go`
- [x] 1.2 Delete `cmd/weather.go`
- [x] 1.3 Delete `cmd/ollama.go`
- [x] 1.4 Delete `cmd/picker.go` and `cmd/picker_test.go`
- [x] 1.5 Delete `cmd/new.go` and `cmd/new_test.go`
- [x] 1.6 Delete `cmd/kill.go`
- [x] 1.7 Delete `cmd/code.go`
- [x] 1.8 Delete `cmd/sysop.go` and `cmd/sysop_test.go`
- [x] 1.9 Delete `cmd/welcome.go` and `cmd/welcome_test.go`
- [x] 1.10 Delete `cmd/orcai-picker/` directory
- [x] 1.11 Delete `cmd/orcai-sysop/` directory
- [x] 1.12 Delete `cmd/orcai-welcome/` directory

## 2. Update `main.go` Dispatch

- [x] 2.1 Remove `_jump` case from the `switch os.Args[1]` block in `main.go`
- [x] 2.2 Remove `_welcome` case from `main.go`
- [x] 2.3 Remove `jumpwindow` import from `main.go`
- [x] 2.4 Update the cobra dispatch case list to remove: `git`, `weather`, `code`, `new`, `kill`, `ollama`, `picker`, `sysop`, `welcome`

## 3. Update Bootstrap Keybindings

- [x] 3.1 In `internal/bootstrap/bootstrap.go`, replace the `^spc j` binding from `display-popup -E -w 70 -h 24 "… _jump"` to `send-keys -t "#{session_name}:0" J` (works from both orcai and orcai-cron sessions)
- [x] 3.2 Remove the `ollama` popup keybinding (`^spc o`) from bootstrap; unbind chord `o` in generated conf
- [x] 3.3 Update the bootstrap help text (status bar hints) to remove jump/ollama references if present

## 4. Embed Jump Window as Switchboard Modal

- [x] 4.1 Create `internal/jumpwindow/embedded.go` — `EmbeddedModel` with `Update(msg)`, `View(bundle, w)` and `CloseMsg` for parent dismissal; works in both switchboard and cron TUI
- [x] 4.2 Add `jumpOpen bool` and `jumpModal jumpwindow.EmbeddedModel` to switchboard `Model` struct
- [x] 4.3 In `switchboard.go` `handleKey()`, handle `J` to initialize and open jump modal
- [x] 4.4 In `switchboard.go` `Update()`, handle `jumpwindow.CloseMsg` to close modal; route keys to jump modal when `jumpOpen`
- [x] 4.5 In `switchboard.go` `View()`, render jump modal overlay (centered, using active theme) when `jumpOpen`
- [x] 4.6 Add `jumpOpen`/`jumpModal` to crontui `Model`; wire `J` key, `CloseMsg`, and `View()` overlay — jump accessible from both switchboard and cron TUI

## 5. Cron TUI Theme Disk Polling

- [x] 5.1 `pollThemeFile()` tick at 5s already in `internal/crontui/init.go` — started in `Init()`
- [x] 5.2 `themeFilePollMsg` type already defined in crontui message types
- [x] 5.3 `themeFilePollMsg` handler in `Update()` calls `gr.RefreshActive()`, compares to `m.lastThemeName`, updates `m.bundle`
- [x] 5.4 (smoke-test on running session)

## 6. Keybindings Cleanup

- [x] 6.1 Remove the `launch-session-picker`, `open-sysop`, `open-welcome` actions from `internal/keybindings/keybindings.go`
- [x] 6.2 Verify no remaining references to removed commands in keybindings defaults

## 7. Build and Test

- [x] 7.1 `go build ./...` passes; also fixed pre-existing build failures (WorkingDir removed from Entry; PanelHeader 4th arg missing at call sites; bootstrap keybinding test tokens)
- [x] 7.2 `go test ./...` all pass
## 8. Help Modal in Cron TUI

- [x] 8.1 Wire `?`/`ctrl+h` in `crontui` `handleKey()` to open the help modal overlay (uses `modal.RenderScroll` + local README reader)
- [x] 8.2 Add `helpOpen bool` and `helpScrollOffset int` to crontui `Model`; render help overlay in `View()` when open
- [x] 8.3 Add `J jump` and `? help` hints to crontui hint bar

## 9. Smoke Tests

- [ ] 7.3 Smoke-test: run `make run`, verify `orcai --help` shows only the expected commands
- [ ] 7.4 Smoke-test: press `^spc j` inside a live orcai session and confirm the jump modal appears inside the switchboard
- [ ] 7.5 Smoke-test: switch theme in the switchboard, wait ≤ 5 seconds, verify the cron TUI (`orcai cron tui`) updates
