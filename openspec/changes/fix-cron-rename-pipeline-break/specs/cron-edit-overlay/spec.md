## ADDED Requirements

### Requirement: Edit overlay preserves uneditable fields
When a cron entry is saved through the edit overlay, ALL fields of the original `cron.Entry` SHALL be preserved in the written result. Fields not exposed as text inputs (`Args`, `WorkingDir`) MUST be copied from the original entry and MUST NOT revert to zero values.

#### Scenario: Rename preserves WorkingDir
- **WHEN** a user opens the edit overlay for an entry that has `working_dir` set
- **WHEN** the user changes only the `Name` field and confirms
- **THEN** the saved entry SHALL have the same `working_dir` as the original

#### Scenario: Edit preserves Args
- **WHEN** a user opens the edit overlay for an entry that has `args` set
- **WHEN** the user changes any editable field and confirms
- **THEN** the saved entry SHALL have the same `args` map as the original

#### Scenario: New entry has zero values for uneditable fields
- **WHEN** a user creates a brand-new cron entry (no original)
- **THEN** `Args` SHALL be nil and `WorkingDir` SHALL be empty string
