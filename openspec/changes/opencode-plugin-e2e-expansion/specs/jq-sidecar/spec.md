## ADDED Requirements

### Requirement: Binary reads JSON from stdin and applies a jq filter
The `orcai-jq` binary SHALL read its entire stdin as a JSON string, then apply a `jq` filter expression to it and write the result to stdout.

#### Scenario: Filter applied to JSON input
- **WHEN** stdin contains `{"name":"orcai","version":"1.0"}` and the filter is `.name`
- **THEN** stdout contains `"orcai"` and the binary exits 0

#### Scenario: Empty stdin produces error
- **WHEN** stdin is closed with no data
- **THEN** the binary exits non-zero with a descriptive error

### Requirement: Filter resolved from --filter flag or ORCAI_JQ_FILTER env var
The binary SHALL accept a `--filter` flag. If not provided, it SHALL fall back to `ORCAI_JQ_FILTER`. If neither is set, it SHALL default to the identity filter `.` (pass-through).

#### Scenario: --filter flag sets filter expression
- **WHEN** `--filter .items[0].id` is passed
- **THEN** the jq expression `.items[0].id` is applied to stdin

#### Scenario: ORCAI_JQ_FILTER used as fallback
- **WHEN** `--filter` is absent and `ORCAI_JQ_FILTER=.result.count` is in the environment
- **THEN** `.result.count` is applied to stdin

#### Scenario: Default filter is identity
- **WHEN** neither `--filter` nor `ORCAI_JQ_FILTER` is set
- **THEN** stdin JSON is passed through unchanged to stdout

### Requirement: jq execution errors surfaced as non-zero exit
If the `jq` binary is not found, or if the filter expression is invalid, or if the input is not valid JSON, the binary SHALL exit non-zero and write the error to stderr.

#### Scenario: Invalid JSON input
- **WHEN** stdin contains `not json`
- **THEN** the binary exits non-zero and stderr contains an error message

#### Scenario: Invalid filter expression
- **WHEN** the filter expression is syntactically invalid jq
- **THEN** the binary exits non-zero with jq's error in stderr

#### Scenario: jq not installed
- **WHEN** `jq` is not found on PATH
- **THEN** the binary exits non-zero with a message indicating `jq` is required

### Requirement: Sidecar YAML declares the jq provider
A sidecar file `jq.yaml` SHALL be provided in `plugins/jq/`. It SHALL declare `name: jq`, a description, `command: orcai-jq`, and document `vars.filter` as the jq filter expression.

#### Scenario: Sidecar loaded by orcai
- **WHEN** `jq.yaml` is placed in `~/.config/orcai/wrappers/` and `orcai-jq` is on PATH
- **THEN** `orcai` registers `jq` as an available provider plugin

### Requirement: Plugin repository layout for jq plugin
The `plugins/jq/` directory SHALL contain `main.go`, `go.mod`, `Makefile`, and `jq.yaml` with the same structure as `plugins/ollama/`.

#### Scenario: make install places binary and sidecar
- **WHEN** `make install` is run from `plugins/jq/`
- **THEN** `orcai-jq` is available via `which orcai-jq` and `jq.yaml` exists in `~/.config/orcai/wrappers/`
