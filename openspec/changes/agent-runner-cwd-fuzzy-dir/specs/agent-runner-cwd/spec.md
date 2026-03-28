## ADDED Requirements

### Requirement: Agent runner overlay displays a WORKING DIRECTORY field
The agent runner overlay SHALL include a WORKING DIRECTORY field below the PROMPT section. It SHALL default to the directory orcai was launched from (`os.Getwd()` captured at startup). The field SHALL display the current value and allow the user to open the dir picker overlay to change it.

#### Scenario: Default CWD shown on open
- **WHEN** the agent runner overlay opens
- **THEN** the WORKING DIRECTORY field displays the orcai launch directory

#### Scenario: User opens dir picker from CWD field
- **WHEN** the WORKING DIRECTORY field is focused and the user presses Enter or a trigger key
- **THEN** the dir picker overlay opens with the current CWD value pre-filled as the filter

#### Scenario: CWD updates after dir picker selection
- **WHEN** the user selects a directory in the dir picker overlay
- **THEN** the WORKING DIRECTORY field in the agent runner overlay updates to the selected path

#### Scenario: Agent job launched with selected CWD
- **WHEN** the user submits the agent runner overlay (ctrl+s)
- **THEN** the agent session is started with the working directory set to the value in the WORKING DIRECTORY field

### Requirement: Tab focus cycles through all agent runner overlay fields
The agent runner overlay SHALL include WORKING DIRECTORY in its tab focus cycle: PROVIDER → MODEL → PROMPT → WORKING DIRECTORY → (back to PROVIDER).

#### Scenario: Tab reaches WORKING DIRECTORY
- **WHEN** the user presses Tab while PROMPT is focused
- **THEN** focus moves to the WORKING DIRECTORY field

#### Scenario: Tab wraps from WORKING DIRECTORY back to PROVIDER
- **WHEN** the user presses Tab while WORKING DIRECTORY is focused
- **THEN** focus moves back to the PROVIDER list
