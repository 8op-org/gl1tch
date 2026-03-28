## MODIFIED Requirements

### Requirement: Header title spans full terminal width and is centered
The switchboard header bar (currently displaying "ABBS Switchboard" left-aligned with a narrow accent bar) SHALL be rerendered to span exactly 100% of the terminal width, with the title text horizontally centered in the bar.

#### Scenario: Header fills full terminal width
- **WHEN** the terminal is any width W
- **THEN** the header bar SHALL be exactly W characters wide with no trailing whitespace gap between the title region and the right edge

#### Scenario: Title is horizontally centered
- **WHEN** the header bar is rendered
- **THEN** the title text SHALL have equal padding on its left and right sides (within ±1 character for odd-width arithmetic)

#### Scenario: Header recenters on terminal resize
- **WHEN** the terminal is resized (tea.WindowSizeMsg received)
- **THEN** the next View() call SHALL recalculate the centered title at the new width

#### Scenario: Secondary panels below header left-align with header left edge
- **WHEN** the switchboard is rendered with the full-width header
- **THEN** the left edge of each panel column SHALL align with the left edge of the header bar (no indent offset from the header)
