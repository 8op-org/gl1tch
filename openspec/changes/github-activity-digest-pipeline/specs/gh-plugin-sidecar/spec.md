## ADDED Requirements

### Requirement: gh sidecar plugin descriptor
The system SHALL provide an ORCAI sidecar YAML at `orcai-plugins/plugins/gh/gh.yaml` that wraps the `gh` CLI binary, enabling pipeline steps to call `gh` subcommands by setting `ORCAI_ARGS`.

#### Scenario: Pipeline step calls gh issue list
- **WHEN** a pipeline step sets `plugin: gh` and `vars.args: "issue list --repo elastic/ensemble --json number,title,state,updatedAt,labels --limit 30"`
- **THEN** the step executes `gh issue list --repo elastic/ensemble ...` and returns the JSON output as `step.<id>.data.value`

#### Scenario: Missing gh binary
- **WHEN** `gh` is not on PATH
- **THEN** the pipeline step fails with a non-zero exit code and an error message surfaced by `builtin.assert`

### Requirement: gh sidecar installed as wrapper
The sidecar descriptor SHALL be installable by copying `gh.yaml` to `~/.config/orcai/wrappers/gh.yaml` with no compiled binary required.

#### Scenario: Sidecar installed from YAML only
- **WHEN** user copies `gh.yaml` to `~/.config/orcai/wrappers/`
- **THEN** `orcai pipeline run` resolves `plugin: gh` steps without error
