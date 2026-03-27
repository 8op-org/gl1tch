## Why

The cron scheduler and result store exist but are config-file-only — there is no way to schedule a recurring agent run or pipeline directly from the TUI. Users must hand-edit `cron.yaml`, restart the daemon, and guess at the cron syntax; the UI treats every launch as a one-shot run with no memory.

## What Changes

- Add a **SCHEDULE** field to the agent runner modal (replacing the current WORKING DIRECTORY focus slot, which opened a dir picker) so users can optionally enter a cron expression; submitting with a schedule registers a new entry in `cron.yaml` and restarts the daemon, instead of launching immediately.
- Replace the pipeline **dir picker** pre-launch flow with a two-step modal: first choose **Run now** (keeps existing dir picker) or **Schedule recurring** (enters a cron expression inline); the schedule path writes to `cron.yaml` and restarts the daemon.
- Add a `cron.WriteEntry` / `cron.RemoveEntry` helper that atomically updates `~/.config/orcai/cron.yaml` from within the TUI process (the running daemon hot-reloads via fsnotify automatically).
- Add an integration test that starts a pipeline via `orcai` and asserts it completes within a few seconds.

## Capabilities

### New Capabilities

- `agent-runner-cron-schedule`: The agent runner modal gains a SCHEDULE field (cron expression, blank = run once). Submitting with a non-blank schedule writes the entry to `cron.yaml` and shows a confirmation feed item instead of spawning a tmux window.
- `pipeline-cron-schedule`: The pipeline launcher replaces the immediate dir-picker flow with a mode-select step (Run now / Schedule recurring). The recurring path collects a cron expression, writes to `cron.yaml`, and confirms in the feed.
- `cron-yaml-writer`: A `WriteEntry(entry Entry)` and `RemoveEntry(name string)` API in `internal/cron/` that reads, mutates, and atomically writes `cron.yaml` — used by both UI surfaces.

### Modified Capabilities

- `pipeline-cron-scheduler`: Scheduler already hot-reloads on file write via fsnotify; no code change needed, but the spec should document that `WriteEntry` triggers a reload within ~1 s.

## Impact

- `internal/switchboard/switchboard.go`: agent modal focus flow changes (focus slot 3 becomes SCHEDULE text input, not dir picker trigger); pipeline launcher adds a pre-picker mode-select overlay.
- `internal/switchboard/`: new `cronpicker.go` (or inline structs) for the mode-select overlay and schedule input.
- `internal/cron/writer.go`: new file implementing `WriteEntry` / `RemoveEntry` with file-level locking.
- `internal/cron/scheduler.go`: no changes required (hot-reload already works).
- `internal/pipeline/`: integration test covering end-to-end pipeline execution via `orcai` CLI.
- No breaking changes to existing `cron.yaml` schema or pipeline YAML syntax.
