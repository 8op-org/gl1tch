## ADDED Requirements

### Requirement: Agent runner modal width is 90% of terminal width
The agent runner modal overlay SHALL be sized to 90% of the current terminal width (`w * 9 / 10`), with a minimum of 60 columns and a maximum of the full terminal width minus 2 columns. This replaces the previous hard cap of 90 columns.

#### Scenario: Modal uses 90% width on a wide terminal
- **WHEN** the terminal is 160 columns wide and the agent modal opens
- **THEN** the modal is 144 columns wide (`160 * 9 / 10`)

#### Scenario: Minimum width enforced on narrow terminal
- **WHEN** the terminal is 50 columns wide and the agent modal opens
- **THEN** the modal is at most 48 columns wide (clamped to `w - 2`) and at least 60 columns if the terminal allows

#### Scenario: Standard terminal width
- **WHEN** the terminal is 120 columns wide and the agent modal opens
- **THEN** the modal is 108 columns wide (`120 * 9 / 10`)

### Requirement: Agent runner modal accepts a pre-populated prompt
The agent runner modal SHALL accept an optional initial prompt string when opened. When a non-empty initial prompt is provided, the prompt textarea SHALL be pre-populated with that string and the modal focus SHALL start at slot 2 (the prompt field).

#### Scenario: Modal pre-populates prompt from caller
- **WHEN** the modal is opened with a non-empty initial prompt string
- **THEN** the prompt textarea contains that string when the modal first renders

#### Scenario: Modal focus starts on prompt when pre-populated
- **WHEN** the modal is opened with a non-empty initial prompt string
- **THEN** the active focus slot is 2, not 0

#### Scenario: Normal modal open without pre-population
- **WHEN** the modal is opened without an initial prompt (e.g., from the agent runner panel)
- **THEN** the prompt textarea is empty and focus starts at slot 0 (provider), unchanged from current behavior
