## ADDED Requirements

### Requirement: Sidebar component is independently reusable
The `internal/buildershared` package SHALL export a `Sidebar` BubbleTea sub-model that renders a searchable, scrollable list of named items and reports selection changes via messages, with no dependency on prompt or pipeline domain types.

#### Scenario: Sidebar renders item list
- **WHEN** the Sidebar is initialized with a list of names
- **THEN** it renders each name as a selectable row with the active item highlighted

#### Scenario: Sidebar filters on search input
- **WHEN** the user types characters into the sidebar search field
- **THEN** only items whose names contain the search string (case-insensitive) are shown

#### Scenario: Sidebar emits selection message
- **WHEN** the user presses Enter on a highlighted item
- **THEN** the Sidebar emits a `SidebarSelectMsg` carrying the selected item name

### Requirement: EditorPanel component is independently reusable
The `internal/buildershared` package SHALL export an `EditorPanel` BubbleTea sub-model that provides a name input field and a multi-line prompt/content textarea, usable by both the prompt builder and pipeline builder.

#### Scenario: EditorPanel exposes typed content
- **WHEN** the user edits the name input or the content textarea
- **THEN** `EditorPanel.Name()` and `EditorPanel.Content()` return the current values

#### Scenario: EditorPanel tab-cycles focus between name and content
- **WHEN** the user presses Tab while the EditorPanel is focused
- **THEN** focus alternates between the name input and the content textarea

### Requirement: RunnerPanel component is independently reusable
The `internal/buildershared` package SHALL export a `RunnerPanel` BubbleTea sub-model that streams and displays test-runner output lines and exposes a method to inject a starting prompt.

#### Scenario: RunnerPanel appends streaming output
- **WHEN** the parent sends a `RunLineMsg` to the RunnerPanel
- **THEN** the new line is appended to the scrollable output area

#### Scenario: RunnerPanel clears on new run
- **WHEN** `RunnerPanel.Clear()` is called before a new run starts
- **THEN** all previous output lines are removed
