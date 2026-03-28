## ADDED Requirements

### Requirement: Dir picker overlay performs fuzzy search across directories under ~/
The dir picker overlay SHALL be a self-contained BubbleTea component that walks the user's home directory asynchronously (depth ≤ 3) and presents matching directory names filtered by the user's typed query. It SHALL display at most 50 results at a time.

#### Scenario: Walk starts when overlay opens
- **WHEN** the dir picker overlay opens
- **THEN** an async walk of `~/` begins immediately and streams results to the component

#### Scenario: Results appear while typing
- **WHEN** the user types a query
- **THEN** the displayed list is filtered to directories whose names fuzzy-match the query, updating in real time

#### Scenario: Walk is depth-limited
- **WHEN** the home directory walk runs
- **THEN** directories deeper than 3 levels below `~/` are not included

#### Scenario: At most 50 results shown
- **WHEN** the filtered result set exceeds 50 items
- **THEN** only the top 50 ranked matches are displayed

### Requirement: Dir picker overlay is navigable and dismissable
The overlay SHALL support keyboard navigation (up/down arrows), selection (Enter), and dismissal without selection (Esc). On selection it SHALL emit a message containing the chosen absolute path. On dismissal it SHALL emit a cancel message.

#### Scenario: Arrow keys move selection
- **WHEN** the user presses the down arrow key
- **THEN** the highlighted item moves to the next entry in the list

#### Scenario: Enter confirms selection
- **WHEN** the user presses Enter on a highlighted directory
- **THEN** the overlay closes and emits a `DirSelectedMsg` with the absolute path

#### Scenario: Esc cancels without selection
- **WHEN** the user presses Esc
- **THEN** the overlay closes and emits a `DirCancelledMsg`; the caller's CWD field is unchanged

### Requirement: Dir picker overlay is reusable across contexts
The dir picker component SHALL be instantiable from any switchboard overlay (agent runner, pipeline launcher) by creating a new model instance. It SHALL NOT hold any context-specific state beyond the query and walk results.

#### Scenario: Instantiated by agent runner
- **WHEN** the agent runner overlay opens the dir picker
- **THEN** the dir picker model is initialized fresh with no pre-existing query

#### Scenario: Instantiated by pipeline launcher
- **WHEN** the pipeline launcher opens the dir picker
- **THEN** the same component code is used with a new model instance
