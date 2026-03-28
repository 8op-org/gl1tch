## Context

`cron.Entry` has 7 fields: `Name`, `Schedule`, `Kind`, `Target`, `Args`, `Timeout`, `WorkingDir`. The cron edit overlay (`EditOverlay`) only exposes 5 of these as text inputs. When the user saves an edit, `confirmEdit()` constructs a brand-new `cron.Entry` from those 5 inputs — `Args` and `WorkingDir` are never copied from the original, so they revert to zero values (`nil` and `""`).

The silent data loss affects any cron entry that sets `working_dir` (required when the pipeline path is relative to a project directory) or passes `args`. After a single edit the entry is written back to `cron.yaml` without those fields.

## Goals / Non-Goals

**Goals:**
- Preserve `Args` and `WorkingDir` from the original entry whenever a user edits and saves a cron job.
- No regressions on the 5 editable fields.

**Non-Goals:**
- Exposing `Args` or `WorkingDir` as editable fields in the TUI (separate UX concern).
- Changing how `cron.yaml` is stored or how the scheduler reads it.

## Decisions

**Copy uneditable fields from `ov.original` in `confirmEdit()`**

`EditOverlay` already stores the original entry (`ov.original`) for the rename-removal logic (line 286). The fix is to copy the two uneditable fields from `original` when building `updated`:

```go
updated := cron.Entry{
    Name:       ov.fields[0].Value(),
    Schedule:   ov.fields[1].Value(),
    Kind:       ov.fields[2].Value(),
    Target:     ov.fields[3].Value(),
    Timeout:    ov.fields[4].Value(),
    Args:       ov.original.Args,       // preserved
    WorkingDir: ov.original.WorkingDir, // preserved
}
```

This is the smallest possible change: no struct changes, no new fields in the form, no migration needed.

**Alternative considered — add `Args`/`WorkingDir` fields to the form**

Would fix the bug and allow editing those fields, but adds form complexity and a wider diff. Deferred as a separate improvement.

## Risks / Trade-offs

- If a user intentionally wants to clear `WorkingDir` they cannot do so through the TUI (they must edit `cron.yaml` directly). Acceptable for a bug-fix scope.

## Migration Plan

No migration needed. `cron.yaml` files are not altered by this change. The fix only affects in-memory construction during saves from the TUI.
