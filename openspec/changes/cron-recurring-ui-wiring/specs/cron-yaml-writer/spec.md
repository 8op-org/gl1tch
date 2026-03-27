## ADDED Requirements

### Requirement: WriteEntry atomically adds an entry to cron.yaml
`cron.WriteEntry(entry Entry)` SHALL read `~/.config/orcai/cron.yaml`, append the entry (or replace an existing entry with the same `name`), and write the result atomically via a temp-file rename. It SHALL acquire a file-level lock before reading to prevent concurrent writes.

#### Scenario: First entry creates the file
- **WHEN** `cron.yaml` does not exist and `WriteEntry` is called
- **THEN** `cron.yaml` is created with the single entry under `entries:`

#### Scenario: Subsequent entries are appended
- **WHEN** `cron.yaml` already has one or more entries and `WriteEntry` is called with a new name
- **THEN** the new entry is appended and existing entries are preserved

#### Scenario: Entry with duplicate name is replaced
- **WHEN** `WriteEntry` is called with a name that matches an existing entry
- **THEN** the existing entry is replaced in-place (same position) with the new data

#### Scenario: Concurrent writes do not corrupt the file
- **WHEN** two goroutines call `WriteEntry` simultaneously
- **THEN** both entries are eventually present and the YAML is valid (lock serializes access)

### Requirement: RemoveEntry atomically removes a named entry from cron.yaml
`cron.RemoveEntry(name string)` SHALL read `cron.yaml`, remove the entry whose `name` matches, and write the result atomically. It SHALL be a no-op (no error) if the name is not found.

#### Scenario: Named entry is removed
- **WHEN** `RemoveEntry` is called with a name that exists in `cron.yaml`
- **THEN** the entry is absent from the file after the call and all other entries are preserved

#### Scenario: Non-existent name is a no-op
- **WHEN** `RemoveEntry` is called with a name not present in `cron.yaml`
- **THEN** the file is unchanged and no error is returned

### Requirement: Integration test confirms pipeline executes end-to-end via CLI
A Go integration test SHALL build the `orcai` binary, run `orcai pipeline run <fixture>` against a minimal fixture pipeline, and assert the process exits 0 within 5 seconds.

#### Scenario: Simple shell-step pipeline succeeds
- **WHEN** the integration test runs `orcai pipeline run testdata/simple.pipeline.yaml`
- **THEN** the process exits with status 0 within 5 seconds
- **AND** stdout contains the expected output from the fixture's shell step
