## Context

Two TUI components — switchboard and crontui — each need a scrollable help overlay. Switchboard already has one; its content-loading (`readmeContent`, `renderMarkdown`, `fallbackReadme`) and its scroll view (`viewHelpModal`) are private to the switchboard package. The `modal` package already provides `RenderScroll`, but the 80%-width constraint and README rendering are not part of it. Crontui currently has no help overlay.

## Goals / Non-Goals

**Goals:**
- Move README content loading and markdown rendering into `internal/modal` as a shared `RenderHelp` function.
- Make the popup width 80% of the terminal width (clamped to a minimum) instead of `w-4`.
- Add a `?` keybinding in crontui that opens the same help overlay using `modal.RenderHelp`.
- Remove duplicate private code from switchboard's `help_modal.go`.

**Non-Goals:**
- Changing the content shown in the help overlay.
- Adding per-component help content (each component shows the same README).
- Modifying the `RenderConfirm` or `RenderAlert` sizing.

## Decisions

### 1. Add `RenderHelp` to `internal/modal` rather than a new package

`modal` already has `RenderScroll`, `Config`, and `ResolveColors`. Adding `RenderHelp` there keeps the call-site simple (`modal.RenderHelp(cfg, offset, w, h)`) and avoids a new package import for both consumers.

Alternatives considered:
- New `internal/help` package — unnecessary indirection for two files.
- Keep content loading in switchboard, only share sizing — would still require crontui to duplicate the README logic.

### 2. 80% width calculated inside `RenderHelp`, not by callers

Callers pass the raw terminal `w`; `RenderHelp` computes `innerW = w * 4/5` (≥ 40). This makes the constraint consistent and removes sizing math from both consumers.

`RenderScroll` is unchanged so existing callers (e.g. the inbox detail panel) are unaffected.

### 3. `readmeContent`, `renderMarkdown`, `fallbackReadme` move to `internal/modal`

These are self-contained (os.Executable lookup, glamour render, constant). Moving them makes `help_modal.go` in switchboard a thin wrapper:

```go
func (m Model) viewHelpModal(w, h int) string {
    cfg := modal.Config{Bundle: m.activeBundle(), Title: ...}
    return modal.RenderHelp(cfg, m.helpScrollOffset, w, h)
}
```

### 4. Crontui keybinding: `?`

Switchboard uses a chord (`^spc h`). Crontui uses direct single-key bindings (e.g. `e` for edit, `d` for delete, `q` for quit). `?` is the universal TUI help key and is unbound in crontui. Scroll keys match switchboard: `j`/`k`, `[`/`]`, esc to close.

## Risks / Trade-offs

- [glamour import moves to modal] `internal/modal` gains a dependency on `github.com/charmbracelet/glamour`. It already imports `lipgloss`; glamour is already in go.mod. Low risk.
- [os.Executable in modal] Minor — `readmeContent` uses `os.Executable`. Acceptable; modal is already a non-pure rendering package.

## Migration Plan

1. Add `RenderHelp`, `readmeContent`, `renderMarkdown`, `fallbackReadme` to `internal/modal/modal.go`.
2. Update `internal/switchboard/help_modal.go` to call `modal.RenderHelp`; delete the moved symbols.
3. Add `helpOpen bool` + `helpScrollOffset int` to `internal/crontui/model.go`.
4. Add `?` keybinding + scroll + esc handling in `internal/crontui/update.go`.
5. Add overlay dispatch + `viewHelpModal` to `internal/crontui/view.go`.
6. Run `go build ./...` and `go test ./...`.

Rollback: revert the five file changes — no schema/config/binary-format changes.
