## 1. Extend modal package with shared RenderHelp

- [x] 1.1 Move `readmeContent`, `renderMarkdown`, and `fallbackReadme` from `internal/switchboard/help_modal.go` into `internal/modal/modal.go`
- [x] 1.2 Add `RenderHelp(cfg Config, offset, w, h int) string` to `internal/modal/modal.go` that computes `innerW = w * 4 / 5` (min 40), calls `renderMarkdown` + `RenderScroll`
- [x] 1.3 Add `glamour` import to `internal/modal/modal.go` (already in go.mod)

## 2. Update switchboard to delegate to modal.RenderHelp

- [x] 2.1 In `internal/switchboard/help_modal.go`, replace the body of `viewHelpModal` to call `modal.RenderHelp(cfg, m.helpScrollOffset, w, h)`
- [x] 2.2 Delete the now-unused `readmeContent`, `renderMarkdown`, and `fallbackReadme` symbols from `internal/switchboard/help_modal.go`
- [x] 2.3 Verify `go build ./internal/switchboard/...` passes and help modal still works visually

## 3. Add help overlay to crontui

- [x] 3.1 Add `helpOpen bool` and `helpScrollOffset int` fields to `Model` in `internal/crontui/model.go`
- [x] 3.2 In `internal/crontui/update.go`, add `?` keybinding in the main key handler (when no overlay is open) to set `helpOpen = true`, `helpScrollOffset = 0`
- [x] 3.3 In `internal/crontui/update.go`, add a help-overlay key block (checked before main keys when `helpOpen`): `esc` closes, `j`/`]` scroll down by 1/10, `k`/`[` scroll up by 1/10
- [x] 3.4 Add `viewHelpModal() string` method to `internal/crontui/view.go` calling `modal.RenderHelp(cfg, m.helpScrollOffset, m.width, m.height)`
- [x] 3.5 In `internal/crontui/view.go` overlay dispatch block, add a check for `m.helpOpen` and return `renderOverlay(content, m.viewHelpModal(), m.width, m.height, bgColor)`

## 4. Verification

- [x] 4.1 Run `go build ./...` — no errors
- [x] 4.2 Run `go test ./internal/modal/... ./internal/switchboard/... ./internal/crontui/...` — all pass
