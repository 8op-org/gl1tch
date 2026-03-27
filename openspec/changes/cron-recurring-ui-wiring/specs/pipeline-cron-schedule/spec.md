## ADDED Requirements

### Requirement: Pipeline launcher shows mode-select before running
After the user selects a pipeline to launch, the system SHALL display a two-item mode-select overlay before proceeding. The options SHALL be `Run now` and `Schedule recurring`.

#### Scenario: User chooses Run now
- **WHEN** the user selects `Run now` from the mode-select overlay
- **THEN** the existing dir picker overlay opens (unchanged behavior)

#### Scenario: User chooses Schedule recurring
- **WHEN** the user selects `Schedule recurring` from the mode-select overlay
- **THEN** a cron expression text input replaces the dir picker
- **AND** the user can enter a 5-field cron expression and confirm

#### Scenario: Escape closes the mode-select overlay
- **WHEN** the mode-select overlay is open and the user presses Esc
- **THEN** the overlay closes without launching or scheduling the pipeline

### Requirement: Confirming a pipeline schedule writes to cron.yaml
When the user confirms a valid cron expression in the pipeline schedule input, the system SHALL write a new entry to `~/.config/orcai/cron.yaml` with `kind: pipeline` and the pipeline's name and YAML path.

#### Scenario: Valid schedule is confirmed
- **WHEN** the user enters a valid cron expression and confirms
- **THEN** a new entry is appended to `cron.yaml`
- **AND** a feed item appears: `scheduled: <pipeline-name> @ <expression>`
- **AND** no tmux window is opened

#### Scenario: Invalid cron expression in pipeline schedule
- **WHEN** the user enters an invalid cron expression and confirms
- **THEN** an inline error is displayed in the schedule input
- **AND** nothing is written to `cron.yaml`
