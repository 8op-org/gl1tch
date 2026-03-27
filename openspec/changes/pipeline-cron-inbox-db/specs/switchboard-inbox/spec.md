## ADDED Requirements

### Requirement: Inbox panel in the switchboard left sidebar
The switchboard TUI SHALL include an **Inbox** panel as a new entry in the left sidebar navigation. The Inbox panel SHALL appear between the existing sidebar items. The panel SHALL display a scrollable list of recent pipeline and agent run results sourced from the result store, polled on a 5-second interval via `tea.Tick`. The panel title and all UI elements SHALL use the active Lipgloss theme (Dracula palette, monospace font) and SHALL respond to theme changes at runtime.

#### Scenario: Inbox panel appears in sidebar
- **WHEN** the switchboard is rendered
- **THEN** the left sidebar contains an "Inbox" entry navigable via the existing sidebar keybindings

#### Scenario: Inbox list shows recent runs
- **WHEN** the Inbox panel is focused
- **THEN** a scrollable list of run results is shown, each item displaying `name`, `kind`, elapsed time, and exit status (ok/error indicator)

#### Scenario: Inbox refreshes every 5 seconds
- **WHEN** 5 seconds elapse while the Inbox panel is visible
- **THEN** the list is re-queried from the store and updated in place

#### Scenario: Inbox is theme-aware
- **WHEN** the active theme changes
- **THEN** all Inbox UI elements (list, item delegates, borders, indicators) update to reflect the new theme colors

### Requirement: Full-screen inbox detail modal with mutt-style navigation
Pressing Enter (or the configured confirm key) on an Inbox list item SHALL open a full-screen modal overlay displaying the complete run result. The modal SHALL show: run metadata header (name, kind, started/finished timestamps, duration, exit status), stdout content, and stderr content (if non-empty). The modal SHALL support mutt-style keybindings: `n`/`p` to navigate to the next/previous result without closing, `r` to re-run the pipeline or agent, `d` to delete the result from the store, and `q`/`Esc` to close.

#### Scenario: Enter opens detail modal
- **WHEN** an item is selected in the Inbox list and Enter is pressed
- **THEN** a full-screen modal overlays the switchboard showing the full run output and metadata

#### Scenario: n/p navigate between results in modal
- **WHEN** the detail modal is open and `n` is pressed
- **THEN** the modal content updates to show the next run result without closing and re-opening

#### Scenario: r re-runs the pipeline
- **WHEN** the detail modal is open and `r` is pressed
- **THEN** the pipeline or agent associated with the result is queued for immediate execution and a confirmation toast is shown

#### Scenario: d deletes the result
- **WHEN** the detail modal is open and `d` is pressed
- **THEN** a confirmation prompt appears; on confirmation the result row is deleted from the store and the modal advances to the next item (or closes if no items remain)

#### Scenario: q/Esc closes the modal
- **WHEN** the detail modal is open and `q` or Esc is pressed
- **THEN** the modal is dismissed and focus returns to the Inbox list

### Requirement: Inbox item styling distinguishes success from failure
Inbox list items SHALL use color-coded status indicators: a green dot (or themed success color) for `exit_status = 0`, a red dot (or themed error color) for non-zero `exit_status`, and a yellow spinner for in-flight runs (`exit_status = NULL`). Item text SHALL be truncated to fit the panel width with ellipsis.

#### Scenario: Successful run shows green indicator
- **WHEN** an inbox item has `exit_status = 0`
- **THEN** the item renders a success-colored indicator using the theme's success color

#### Scenario: Failed run shows error indicator
- **WHEN** an inbox item has `exit_status != 0`
- **THEN** the item renders an error-colored indicator using the theme's error color

#### Scenario: In-flight run shows spinner
- **WHEN** an inbox item has `exit_status = NULL`
- **THEN** the item renders an animated spinner in the theme's accent color

### Requirement: Inbox panel emits run-complete event on store update
When the result store records a new completed run while the switchboard is running in the same process, it SHALL publish a `RunCompleted` event on the existing event bus. The Inbox panel SHALL subscribe to this event and refresh immediately (without waiting for the 5-second poll tick).

#### Scenario: In-process run triggers immediate inbox refresh
- **WHEN** a pipeline run completes within the same process as the switchboard
- **THEN** the Inbox list updates within one render cycle rather than waiting for the next poll tick
