## Why

The signal board supports line-by-line navigation but offers no way to collect and act on multiple entries at once. Users need to select specific feed entries as context for a new agent run, and the current agent runner modal is too narrow to comfortably compose prompts.

## What Changes

- Add multi-line marking to the signal board: pressing `space` (or a dedicated mark key) while a line is selected toggles a highlight on that entry
- 'r' key on a focused signal board opens the agent runner modal with the titles/content of all marked entries pre-injected into the prompt textarea
- The agent runner modal now opens at 90% of the terminal width instead of the current hard-capped 90-character maximum
- Executed pipeline steps in the activity feed are rendered vertically (one step per line) with each step's output shown beneath it, instead of the current horizontal wrapping layout

## Capabilities

### New Capabilities

- `feed-line-marking`: Multi-line highlight/mark in the signal board and 'r'-to-agent-runner context injection

### Modified Capabilities

- `agent-runner-modal-layout`: Agent runner modal width changes from a hard cap of 90 columns to 90% of terminal width; prompt textarea pre-population from marked feed entries
- `feed-step-display`: Pipeline step badges in the activity feed change from horizontal-wrapped rows to a vertical per-step layout with output lines shown beneath each step

## Impact

- `internal/switchboard/signal_board.go`: new `marked map[string]bool` field on `SignalBoard`, toggle logic, 'r' key handler
- `internal/switchboard/switchboard.go`: `viewAgentModalBox` width formula, agent modal open path to accept injected prompt text, step rendering loop in the feed body builder
- `internal/switchboard/switchboard_test.go`: tests for mark toggle, 'r'-key injection, modal width, vertical step layout
