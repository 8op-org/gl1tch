## ADDED Requirements

### Requirement: Agent modal exposes a SCHEDULE field
The agent runner modal SHALL replace focus slot 3 (previously WORKING DIRECTORY / dir picker) with a free-text SCHEDULE input. The label SHALL read `SCHEDULE (cron expression, blank = run now)`. The field SHALL default to blank.

#### Scenario: Blank schedule runs agent immediately
- **WHEN** the user submits the agent modal with the SCHEDULE field blank
- **THEN** the agent launches immediately in a new tmux window (existing behavior)

#### Scenario: Valid cron expression schedules the agent
- **WHEN** the user enters a valid 5-field cron expression in the SCHEDULE field and submits
- **THEN** the system writes a new entry to `~/.config/orcai/cron.yaml` with `kind: agent`, the selected provider/model, and the entered prompt
- **AND** a feed item appears confirming the schedule (e.g., `scheduled: <name> @ <expression>`) instead of opening a tmux window

#### Scenario: Invalid cron expression is rejected
- **WHEN** the user enters an unparseable cron expression and submits
- **THEN** the SCHEDULE field displays an inline error message
- **AND** no entry is written to `cron.yaml`
- **AND** the modal remains open

#### Scenario: Tab cycles through all agent modal fields including SCHEDULE
- **WHEN** the modal is open and the user presses Tab repeatedly
- **THEN** focus cycles: provider → model → prompt → SCHEDULE → provider
