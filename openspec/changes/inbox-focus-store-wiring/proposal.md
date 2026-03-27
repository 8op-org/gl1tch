## Why

Two regressions exist in the inbox/store implementation shipped in `pipeline-cron-inbox-db`:

1. **Wrong bottom-bar hints when inbox is focused.** `viewBottomBar` has no case for `inboxFocused`, so it falls through to the default launcher hints ("enter launch · ctrl+s submit …"). The user sees actions that don't apply to the inbox panel.

2. **Pipeline results never reach the database or the inbox.** The store is passed into the switchboard's `inboxModel` but is never kept on the `Model` struct, so it can't be forwarded when a pipeline is launched. More critically, pipelines launched from the switchboard run as an external `orcai pipeline run …` CLI invocation. The CLI handler (`cmd/pipeline.go`) calls `pipeline.Run()` with no `WithRunStore` option, so `RecordRunStart` / `RecordRunComplete` are never executed and the `runs` table stays empty.

## What Changes

- Add an `inboxFocused` case to `viewBottomBar` that shows inbox-appropriate hints (`enter` open · `d` delete · `r` re-run · `tab` focus · `i` inbox).
- Persist the `*store.Store` on the switchboard `Model` struct so it is accessible after construction.
- Open the store in `cmd/pipeline.go` (`pipelineRunCmd`) and pass it to `pipeline.Run()` via `WithRunStore`, so every pipeline execution — whether triggered from the TUI or the CLI — records its run in the database.

## Capabilities

### New Capabilities
- `inbox-focus-hints`: Correct bottom-bar key hints when the Inbox panel has focus.

### Modified Capabilities
- `pipeline-result-store`: Pipeline CLI runner (`cmd/pipeline.go`) must open the store and pass it to `pipeline.Run()` so results are recorded.
- `switchboard-inbox`: Switchboard `Model` must retain the store reference for future use (e.g., re-run dispatch from the inbox).

## Impact

- `internal/switchboard/switchboard.go`: add `store *store.Store` field to `Model`; populate in `NewWithStore`; add `inboxFocused` case in `viewBottomBar`.
- `cmd/pipeline.go`: open store, defer close, pass `WithRunStore(s)` to `pipeline.Run()`.
