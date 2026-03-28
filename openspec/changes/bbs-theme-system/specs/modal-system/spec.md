## ADDED Requirements

### Requirement: Shared modal rendering package
The system SHALL provide an `internal/modal` package that exports pure rendering functions usable from any BubbleTea program without importing switchboard.

#### Scenario: Render confirm modal
- **WHEN** a caller invokes `modal.RenderConfirm(cfg modal.Config, w, h int)` with a title, message, and action labels
- **THEN** the function SHALL return a bordered overlay string centered in the given dimensions with the title bar, message body, and action hint rendered using the bundle's modal colors

#### Scenario: Render alert modal
- **WHEN** a caller invokes `modal.RenderAlert(cfg modal.Config, message string, w, h int)`
- **THEN** the function SHALL return a bordered overlay string with the message and a dismiss hint, using the bundle's modal colors

#### Scenario: Render scrollable content modal
- **WHEN** a caller invokes `modal.RenderScroll(cfg modal.Config, lines []string, offset, w, h int)`
- **THEN** the function SHALL return a bordered overlay with the visible window of lines starting at offset, scroll indicators at top/bottom when content overflows, and the bundle's modal colors applied throughout

#### Scenario: Fallback when bundle is nil
- **WHEN** `modal.Config.Bundle` is nil
- **THEN** all render functions SHALL fall back to Dracula hardcoded colors identical to the current switchboard behavior

### Requirement: Modal config accepts theme bundle and ANSI header
The `modal.Config` struct SHALL accept a `*themes.Bundle`, optional TDF font name, optional ANSI header pattern, and label strings for confirm/dismiss actions.

#### Scenario: Modal uses bundle accent as border color
- **WHEN** a non-nil bundle is provided and `modal.RenderConfirm` is called
- **THEN** the modal border color SHALL match `bundle.Palette.Accent`

#### Scenario: Modal title bar uses bundle title colors
- **WHEN** a non-nil bundle is provided
- **THEN** the title bar background SHALL match `bundle.Modal.TitleBG` (resolved via `Bundle.ResolveRef`) and foreground SHALL match `bundle.Modal.TitleFG`

### Requirement: Switchboard delegates modal rendering to internal/modal
The switchboard package SHALL NOT contain inline modal rendering logic; all modal views SHALL delegate to `internal/modal`.

#### Scenario: Quit confirm modal renders identically before and after refactor
- **WHEN** the user triggers the quit confirm chord in switchboard
- **THEN** the rendered output SHALL be visually identical to the pre-refactor output (border, title bar, message, action hints)

### Requirement: Crontui quit confirm uses internal/modal
The crontui package SHALL use `modal.RenderConfirm` for its quit confirmation overlay instead of inline rendering.

#### Scenario: Crontui quit modal respects active bundle
- **WHEN** crontui has a non-nil bundle and the user triggers quit
- **THEN** the modal SHALL use that bundle's modal colors rather than hardcoded Dracula colors

### Requirement: Go plugin host exposes modal.API
The plugin host SHALL pass a `modal.API` interface to loaded plugins, allowing plugin-owned BubbleTea views to render themed modals.

#### Scenario: Plugin renders confirm modal
- **WHEN** a plugin calls `modalAPI.RenderConfirm(title, message string)` from within its `View()` method
- **THEN** the plugin receives a rendered string using the currently active global theme bundle
