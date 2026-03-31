## MODIFIED Requirements

### Requirement: Jump window sysop entries launch new builder subcommands
The jump window sysop column entries for prompts and pipelines SHALL invoke `orcai prompt-builder` and `orcai pipeline-builder` respectively (in new tmux windows), replacing the legacy `orcai prompts tui` and `orcai pipeline build` commands.

#### Scenario: Prompts entry opens prompt-builder
- **WHEN** the user selects the prompts sysop entry in the jump window
- **THEN** a new tmux window is created running `orcai prompt-builder`

#### Scenario: Pipelines entry opens pipeline-builder
- **WHEN** the user selects the pipelines sysop entry in the jump window
- **THEN** a new tmux window is created running `orcai pipeline-builder`
