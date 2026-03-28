## Why

The help modal exists only in switchboard and uses full terminal width; crontui has no help overlay at all, and the markdown rendering + README-loading logic is not shared. Adding help access to crontui requires duplicating that logic, and the full-width layout is visually noisy compared to an 80%-width popup.

## What Changes

- The `modal` package gains an 80%-width sizing constraint in `RenderScroll` (and a shared `RenderHelp` entry-point that handles content loading + markdown rendering).
- `internal/switchboard` delegates to the shared helper instead of owning `readmeContent`, `renderMarkdown`, and `fallbackReadme`.
- `internal/crontui` gains a help overlay (model state, keybinding `?`, view method) backed by the same shared helper.

## Capabilities

### New Capabilities

- `shared-help-overlay`: A reusable help-overlay renderer in `internal/modal` that loads README content, renders it as markdown, and returns a scrollable 80%-width popup box for any TUI component.

### Modified Capabilities

- none

## Impact

- `internal/modal/modal.go` — new `RenderHelp` / sizing update in `RenderScroll`
- `internal/switchboard/help_modal.go` — remove `readmeContent`, `renderMarkdown`, `fallbackReadme`; delegate to `modal.RenderHelp`
- `internal/crontui/model.go` — add `helpOpen bool`, `helpScrollOffset int`
- `internal/crontui/update.go` — add `?` keybinding, scroll handling
- `internal/crontui/view.go` — add `viewHelpModal` call in overlay dispatch
