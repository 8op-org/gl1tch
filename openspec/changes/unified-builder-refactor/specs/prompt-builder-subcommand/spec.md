## ADDED Requirements

### Requirement: orcai prompt-builder launches the prompt builder TUI
The CLI SHALL expose `orcai prompt-builder` as a subcommand that opens the full-screen two-column prompt builder TUI using an alt-screen BubbleTea program.

#### Scenario: Subcommand opens TUI
- **WHEN** the user runs `orcai prompt-builder`
- **THEN** the terminal enters alt-screen and the two-column prompt builder TUI is displayed

### Requirement: Prompt builder uses two-column layout
The prompt builder TUI SHALL render a left sidebar column and a right column containing the feedback loop panel above and the agent runner chat input below.

#### Scenario: Layout renders both columns
- **WHEN** the prompt builder TUI initializes
- **THEN** the left column shows the saved-prompts sidebar and the right column shows the feedback loop area and chat input

#### Scenario: Tab moves focus between columns
- **WHEN** the user presses Tab
- **THEN** focus cycles: sidebar → feedback loop → chat input → sidebar

### Requirement: ctrl+s saves current prompt
The prompt builder TUI SHALL save the current name and content to disk when the user presses `ctrl+s`.

#### Scenario: Save writes prompt file
- **WHEN** the user presses ctrl+s and the name field is non-empty
- **THEN** the prompt is saved and the sidebar list refreshes to include it

### Requirement: ctrl+r reinjects prompt into test runner
The prompt builder TUI SHALL inject the current prompt content into the RunnerPanel and start a test run when the user presses `ctrl+r`.

#### Scenario: ctrl+r starts test run
- **WHEN** the user presses ctrl+r
- **THEN** the RunnerPanel clears and begins streaming the test run output for the current prompt
