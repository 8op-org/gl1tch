## ADDED Requirements

### Requirement: Signal board entries can be individually marked
The signal board SHALL support marking individual feed entries. Pressing `m` while the signal board is focused SHALL toggle the marked state of the currently selected entry. A marked entry SHALL render with a distinct highlight background color on its entire row. Unmarked entries SHALL render as before.

#### Scenario: Mark toggles on
- **WHEN** the user presses `m` with the cursor on an unmarked entry
- **THEN** that entry's row renders with the mark highlight background and the entry is tracked as marked

#### Scenario: Mark toggles off
- **WHEN** the user presses `m` with the cursor on an already-marked entry
- **THEN** that entry's row returns to normal rendering and is no longer tracked as marked

#### Scenario: Multiple entries can be marked simultaneously
- **WHEN** the user navigates to different entries and presses `m` on each
- **THEN** all selected entries are highlighted independently

#### Scenario: Mark persists across scrolling
- **WHEN** an entry is marked and the user scrolls so the entry is off-screen, then scrolls back
- **THEN** the entry still renders with the mark highlight

### Requirement: 'r' key opens agent runner modal with marked entries injected
When at least one entry is marked in the signal board and the signal board is focused, pressing `r` SHALL open the agent runner modal overlay with the titles of all marked entries pre-populated in the prompt textarea, one title per line, up to a maximum of 20 entries. If more than 20 entries are marked only the first 20 SHALL be injected, followed by a note indicating truncation. The modal SHALL open with focus set to the prompt field (slot 2). Marks SHALL be cleared after the modal opens.

#### Scenario: 'r' opens modal with injected context
- **WHEN** one or more entries are marked and the user presses `r`
- **THEN** the agent runner modal opens with the marked entries' titles in the prompt textarea

#### Scenario: Focus starts on prompt field
- **WHEN** the modal is opened via 'r'
- **THEN** the modal's active focus slot is 2 (the prompt textarea), not 0 (provider)

#### Scenario: Marks are cleared after modal opens
- **WHEN** the user presses `r` to open the modal
- **THEN** all signal board entries return to their normal (unmarked) rendering

#### Scenario: Injection capped at 20 entries
- **WHEN** more than 20 entries are marked and the user presses `r`
- **THEN** only the first 20 titles appear in the prompt, followed by a line reading `... (N more)`

#### Scenario: 'r' with no marks does nothing
- **WHEN** no entries are marked and the user presses `r`
- **THEN** the agent runner modal does NOT open

### Requirement: Mark hint appears in the signal board hint bar
When the signal board is focused and not in search mode, the hint bar SHALL show `m mark` as a hint alongside the existing navigation hints. When at least one entry is marked, an additional hint `r run` SHALL appear.

#### Scenario: Mark hint shown when focused
- **WHEN** the signal board is focused and not searching
- **THEN** the hint bar includes the key `m` with description `mark`

#### Scenario: Run hint shown when entries are marked
- **WHEN** the signal board is focused, not searching, and at least one entry is marked
- **THEN** the hint bar includes the key `r` with description `run`
