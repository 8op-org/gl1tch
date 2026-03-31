## ADDED Requirements

### Requirement: orcai pipeline-builder launches the pipeline builder TUI
The CLI SHALL expose `orcai pipeline-builder` as a subcommand that opens the full-screen two-column pipeline builder TUI using an alt-screen BubbleTea program.

#### Scenario: Subcommand opens TUI
- **WHEN** the user runs `orcai pipeline-builder`
- **THEN** the terminal enters alt-screen and the two-column pipeline builder TUI is displayed

### Requirement: Pipeline builder uses the same two-column layout as prompt builder
The pipeline builder TUI SHALL use the same `buildershared` sidebar, EditorPanel, and RunnerPanel components as the prompt builder, rendering left sidebar and right (feedback loop + chat input).

#### Scenario: Pipeline builder renders shared layout
- **WHEN** the pipeline builder TUI initializes
- **THEN** the left column shows the saved-pipelines sidebar and the right column shows the pipeline content editor above and the agent runner chat input below

### Requirement: Pipeline builder editor shows YAML pipeline content
The EditorPanel inside the pipeline builder SHALL display the pipeline YAML representation in its content textarea, allowing the user to edit it directly.

#### Scenario: Selecting a pipeline loads its YAML
- **WHEN** the user selects a pipeline from the sidebar
- **THEN** the EditorPanel content textarea is populated with the pipeline's YAML content

### Requirement: ctrl+s saves current pipeline
The pipeline builder TUI SHALL save the current pipeline to disk as a `.pipeline.yaml` file when the user presses `ctrl+s`.

#### Scenario: Save writes pipeline file
- **WHEN** the user presses ctrl+s and the pipeline name field is non-empty
- **THEN** the pipeline YAML is written to the pipelines directory and the sidebar refreshes

### Requirement: ctrl+r reinjects pipeline prompt into test runner
The pipeline builder TUI SHALL inject the first step's prompt into the RunnerPanel and start a test run when the user presses `ctrl+r`.

#### Scenario: ctrl+r starts pipeline test run
- **WHEN** the user presses ctrl+r
- **THEN** the RunnerPanel clears and begins streaming the pipeline execution output
