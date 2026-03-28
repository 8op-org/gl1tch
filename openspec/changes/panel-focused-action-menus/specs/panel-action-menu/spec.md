## ADDED Requirements

### Requirement: Panel renders its own action-hint footer when focused
Each panel and pane in the switchboard and cron TUI SHALL render a single-row action-hint footer inside its own border box. The footer SHALL be visible only when the panel is focused. When the panel is not focused the footer row SHALL be absent and the panel SHALL NOT reserve vertical space for it.

#### Scenario: Focused panel shows hint footer
- **WHEN** a switchboard panel or cron TUI pane gains focus
- **THEN** the panel renders a single-row hint footer at the bottom of its content area, inside its border

#### Scenario: Unfocused panel hides hint footer
- **WHEN** a switchboard panel or cron TUI pane does not have focus
- **THEN** the panel renders no hint footer row and does not reserve vertical space for one

#### Scenario: Hint footer fits panel width
- **WHEN** the hint footer is rendered
- **THEN** the footer row is exactly `panelWidth` characters wide, truncated if hints would overflow

### Requirement: Shared HintBar helper in panelrender
The `panelrender` package SHALL expose a `HintBar(hints []Hint, width int, pal Palette) string` function that formats a row of key/description hint pairs. All panels and panes SHALL use this function to render their hint footers.

#### Scenario: HintBar formats hints consistently
- **WHEN** `HintBar` is called with a non-empty hints slice
- **THEN** each hint is rendered with an accented key and dimmed description, separated by ` · `, padded to `width`

#### Scenario: HintBar returns empty string for empty hints
- **WHEN** `HintBar` is called with a nil or empty hints slice
- **THEN** it returns an empty string

### Requirement: Global bottom bars removed
The `viewBottomBar` function in the switchboard and the `viewHintBar` function in the cron TUI SHALL be removed once all panels carry their own footers. No full-terminal-width hint bar SHALL appear below the panel grid.

#### Scenario: No global bottom bar rendered
- **WHEN** any panel has focus
- **THEN** the switchboard renders no full-terminal-width hint bar below the panel grid

#### Scenario: Cron TUI has no standalone hint bar
- **WHEN** the full cron TUI is open
- **THEN** hints appear inside the active pane border, not in a separate bar below both panes

### Requirement: Panel height arithmetic accounts for footer row
Each panel view function SHALL subtract one row from its content height when a footer is present (focused state) so that the panel content does not overflow or clip.

#### Scenario: Content height reduced when footer shown
- **WHEN** a panel is focused and renders a footer row
- **THEN** the scrollable or list content area is one row shorter than when the panel is unfocused

#### Scenario: Content height unchanged when footer absent
- **WHEN** a panel is not focused
- **THEN** the content area height is unchanged relative to the header-only layout
