## ADDED Requirements

### Requirement: modal package exposes RenderHelp
The `internal/modal` package SHALL provide a `RenderHelp(cfg Config, offset, w, h int) string` function that loads README content, renders it as markdown, and returns a scrollable popup box sized at 80% of the terminal width (minimum 40 columns).

#### Scenario: 80% width popup
- **WHEN** `RenderHelp` is called with terminal width `w`
- **THEN** the rendered popup box inner width SHALL be `w * 4 / 5`, clamped to a minimum of 40

#### Scenario: Fallback content when README missing
- **WHEN** no `README.md` is found on disk
- **THEN** `RenderHelp` SHALL render the built-in fallback text

#### Scenario: Scroll offset respected
- **WHEN** `offset` is greater than zero
- **THEN** the visible content window SHALL start at the given line offset

### Requirement: crontui has a help overlay
The crontui TUI component SHALL display a help overlay when the user presses `?`, rendered via `modal.RenderHelp`, sized at 80% terminal width, centered over the background.

#### Scenario: Open help overlay
- **WHEN** user presses `?` in crontui and no other overlay is open
- **THEN** the help overlay SHALL appear centered over the crontui view

#### Scenario: Close help overlay with esc
- **WHEN** the help overlay is open and user presses `esc`
- **THEN** the overlay SHALL close and crontui SHALL return to normal state

#### Scenario: Scroll help overlay
- **WHEN** the help overlay is open and user presses `j` or `]`
- **THEN** the scroll offset SHALL increment by 1 or 10 respectively

#### Scenario: Scroll up in help overlay
- **WHEN** the help overlay is open and user presses `k` or `[`
- **THEN** the scroll offset SHALL decrement by 1 or 10 (clamped to 0)

### Requirement: switchboard help modal uses shared RenderHelp
The switchboard help modal SHALL delegate to `modal.RenderHelp` and SHALL NOT duplicate README loading or markdown rendering logic.

#### Scenario: Switchboard help unchanged
- **WHEN** user opens the help modal in switchboard via `^spc h`
- **THEN** the help content and behavior SHALL be identical to before this change
