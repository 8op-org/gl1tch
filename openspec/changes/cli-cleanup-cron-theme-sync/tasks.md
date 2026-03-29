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

- [x] 3.1 In `internal/bootstrap/bootstrap.go`, replace the `^spc j` binding from `display-popup -E -w 70 -h 24 "… _jump"` to `send-keys -t "#{session_name}:0" J`
- [x] 3.2 Remove the `ollama` popup keybinding (`^spc o`) from bootstrap; unbind chord `o` in generated conf
- [x] 3.3 Bootstrap now launches `orcai` (no args) for the switchboard session instead of `orcai sysop`; removed `resolveCompanion` helper

## 4. Embed Jump Window as Switchboard Modal

- [x] 4.1 Create `internal/jumpwindow/embedded.go` — `EmbeddedModel` with `Update(msg)`, `View()`, `SetSize(w,h)` and `CloseMsg` for parent dismissal; added `embedded bool` to inner model to send CloseMsg instead of tea.Quit
- [x] 4.2 Add `jumpOpen bool` and `jumpModal jumpwindow.EmbeddedModel` to switchboard `Model` struct
- [x] 4.3 In `switchboard.go` `handleKey()`, handle `J` to initialize and open jump modal
- [x] 4.4 In `switchboard.go` `Update()`, handle `jumpwindow.CloseMsg` to close modal; route keys to jump modal when `jumpOpen`
- [x] 4.5 In `switchboard.go` `View()`, render jump modal overlay (centered, using active theme) when `jumpOpen`
- [x] 4.6 Add `jumpOpen`/`jumpModal` to crontui `Model`; wire `J` key, `CloseMsg`, and `View()` overlay

## 5. Cron TUI Theme Disk Polling

- [ ] ~~5.1–5.3 Skipped — themeState already handles cross-process theme sync~~

## 6. Keybindings Cleanup

- [x] 6.1 Remove the `launch-session-picker`, `open-sysop`, `open-welcome` actions from `internal/keybindings/keybindings.go`
- [x] 6.2 No remaining references to removed commands in keybindings defaults

## 7. Build and Test

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./...` all pass

## 8. Help Modal in Cron TUI

- [x] 8.1–8.2 Already implemented (`helpOpen`, `handleHelpKey`, `viewHelpModal` all present)
- [x] 8.3 Added `J jump` and `? help` hints to crontui hint bar

## 9. Smoke Tests

- [ ] 7.3 Smoke-test: run `make run`, verify `orcai --help` shows only the expected commands
- [ ] 7.4 Smoke-test: press `^spc j` inside a live orcai session and confirm the jump modal appears inside the switchboard
- [ ] 7.5 Smoke-test: switch theme in the switchboard, wait ≤ 5 seconds, verify the cron TUI (`orcai cron tui`) updates
