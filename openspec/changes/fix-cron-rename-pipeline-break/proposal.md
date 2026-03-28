## Why

When a user edits a cron entry (including renaming it), the `Args` and `WorkingDir` fields are silently dropped because the edit overlay only exposes 5 of the 7 fields in `cron.Entry`. Any cron job relying on `working_dir` to resolve a relative pipeline path will fail immediately after the next edit.

## What Changes

- **Fix field loss in `confirmEdit()`**: When building the updated entry, copy `Args` and `WorkingDir` from the original entry so they are never silently discarded.
- **Fix form initialization in `newEditOverlay()`**: Preserve `Args` and `WorkingDir` on the overlay struct so they survive round-trips through the edit form.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `cron-edit-overlay`: The cron edit form must preserve all `cron.Entry` fields (including `Args` and `WorkingDir`) that are not directly editable in the 5-field form.

## Impact

- `/Users/stokes/Projects/orcai/internal/crontui/update.go` — `confirmEdit()`, `newEditOverlay()`
- `/Users/stokes/Projects/orcai/internal/crontui/model.go` — `EditOverlay` struct
- No API, schema, or storage changes required
