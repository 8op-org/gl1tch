## ADDED Requirements

### Requirement: Pipeline launch prompts for working directory via overlay modal
When the user triggers a pipeline run, the dir picker overlay SHALL appear before execution begins, allowing the user to set the working directory. The selected CWD SHALL be passed to the pipeline runner as a context variable.

#### Scenario: Dir picker appears on pipeline launch trigger
- **WHEN** the user selects a pipeline to run
- **THEN** the dir picker overlay opens before the pipeline executes

#### Scenario: Pipeline executes with selected CWD
- **WHEN** the user selects a directory and confirms
- **THEN** the pipeline runner starts with `cwd` set to the selected absolute path in the execution context

#### Scenario: Pipeline launch cancelled via Esc
- **WHEN** the user presses Esc in the dir picker
- **THEN** the pipeline does not execute and the UI returns to its previous state

#### Scenario: Pipeline dir picker defaults to orcai launch directory
- **WHEN** the dir picker opens for pipeline launch
- **THEN** the orcai launch directory is highlighted as the default selection

### Requirement: Pipeline CWD is passed as execution context variable
The pipeline execution context SHALL receive a `cwd` key containing the absolute path selected in the dir picker. Pipeline steps MAY reference this via `{{cwd}}` template interpolation.

#### Scenario: CWD available in pipeline template interpolation
- **WHEN** a pipeline step contains `{{cwd}}` in its arguments
- **THEN** the value is replaced with the selected working directory path at execution time
