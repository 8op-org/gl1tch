## ADDED Requirements

### Requirement: Pipeline step badges render one per line in the activity feed
When a feed entry contains pipeline steps, the step badges in the activity feed body SHALL be rendered vertically — one step badge per line — instead of the previous horizontal wrapping layout. Each step badge line SHALL be indented with two spaces and show the step glyph, step ID, and status color.

#### Scenario: Steps render vertically
- **WHEN** a feed entry has three steps and is visible in the activity feed
- **THEN** each step appears on its own line, not packed horizontally with `·` separators

#### Scenario: Step badge format preserved
- **WHEN** a step is in `done` status
- **THEN** its badge line shows the success glyph and step ID in the success color, indented by two spaces

#### Scenario: Running step shown in accent color
- **WHEN** a step is in `running` status
- **THEN** its badge line renders in the yellow/accent color

#### Scenario: Failed step shown in error color
- **WHEN** a step is in `failed` status
- **THEN** its badge line renders in the error color

### Requirement: Each step's output lines appear beneath its badge
When a step has associated output lines (stored in `StepInfo.lines`), those output lines SHALL be rendered immediately below that step's badge line, each indented by four spaces, up to a maximum of 5 lines per step. If a step has more than 5 output lines, only the last 5 SHALL be shown.

#### Scenario: Step output shown beneath badge
- **WHEN** a step has 3 output lines
- **THEN** all 3 lines appear directly below the step badge, indented by 4 spaces

#### Scenario: Step output capped at 5 lines
- **WHEN** a step has 8 output lines
- **THEN** only the last 5 lines appear beneath the badge

#### Scenario: Step with no output shows badge only
- **WHEN** a step has zero output lines
- **THEN** only the badge line is rendered for that step with no blank padding beneath it
