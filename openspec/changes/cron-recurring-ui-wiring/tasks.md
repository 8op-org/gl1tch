## 1. cron.WriteEntry / cron.RemoveEntry

- [x] 1.1 Create `internal/cron/writer.go` with `WriteEntry(entry Entry) error` — reads `cron.yaml`, upserts entry by name, writes atomically via temp-file rename with file-level flock
- [x] 1.2 Implement `RemoveEntry(name string) error` in the same file — reads, filters, writes atomically; no-op if name not found
- [x] 1.3 Write unit tests for `WriteEntry` (create, append, upsert) and `RemoveEntry` (remove, no-op) in `internal/cron/writer_test.go`

## 2. Agent Runner Modal — SCHEDULE field

- [x] 2.1 Replace the dir picker trigger in focus slot 3 of the agent modal with a `textarea` SCHEDULE input; update the label from `WORKING DIRECTORY` to `SCHEDULE (cron expr, blank = run now)`
- [x] 2.2 Update `handleAgentModal` tab/shift-tab cycle to include the SCHEDULE textarea (slot 3) instead of opening the dir picker
- [x] 2.3 On submit: if SCHEDULE is blank, preserve existing agent launch behavior; if non-blank, parse with `robfig/cron` — on error show inline error in the textarea, on success call `cron.WriteEntry` and add a confirmation feed item
- [x] 2.4 Update `renderAgentModal` to render the SCHEDULE textarea with appropriate styling (dimmed when not focused, error state in red)
- [x] 2.5 Update switchboard tests to cover the new SCHEDULE focus slot and both submit paths (run-now and schedule)

## 3. Pipeline Launcher — Mode-Select Overlay

- [x] 3.1 Add `pipelineLaunchMode` state to the switchboard model (`none` / `modeSelect` / `scheduleInput`) and a `pipelineScheduleInput` textarea
- [x] 3.2 When a pipeline is selected (Enter), show the mode-select overlay instead of immediately opening the dir picker; render two items: `Run now` and `Schedule recurring`
- [x] 3.3 `Run now` selection proceeds to existing dir picker flow (unchanged)
- [x] 3.4 `Schedule recurring` selection transitions to `scheduleInput` state, rendering a cron expression textarea
- [x] 3.5 On confirming a schedule: parse with `robfig/cron` — on error show inline error; on success call `cron.WriteEntry` with `kind: pipeline`, add confirmation feed item, reset state
- [x] 3.6 Esc from either mode-select or schedule input resets `pipelineLaunchMode` to `none`
- [x] 3.7 Update `renderLauncher` / view rendering to include the mode-select and schedule-input overlays

## 4. Integration Test — Pipeline via CLI

- [x] 4.1 Create `testdata/simple.pipeline.yaml` fixture in `internal/pipeline/` (single `shell` step: `echo integration-ok`)
- [x] 4.2 Write `internal/pipeline/integration_cli_test.go` with build tag `//go:build integration`; test builds the `orcai` binary with `go build`, runs `orcai pipeline run testdata/simple.pipeline.yaml`, and asserts exit 0 and `integration-ok` in stdout within 5 s
- [x] 4.3 Confirm the integration test passes locally with `go test -tags integration ./internal/pipeline/...`

## 5. Jump Window — Sysop Section (cron daemon)

- [x] 5.1 Add `listSysopWindows() []window` to `internal/jumpwindow/jumpwindow.go` that queries the `orcai-cron` tmux session (`tmux list-windows -t orcai-cron -F "#{window_index}:#{window_id}:#{window_name}:#{@orcai-label}"`) and returns its windows; returns nil (no error) if the session does not exist
- [x] 5.2 Add a `sysop []window` field to the jump window `model`; populate it alongside `windows` in `newModel()` via `listSysopWindows()`
- [x] 5.3 Render the sysop windows at the bottom of the jump window list under a `— sysop —` section header (below the existing `— active jobs —` section); only show the section if `len(sysop) > 0`
- [x] 5.4 Update navigation (up/down/j/k) and Enter to cover sysop entries: selection index spans both `filtered` and `sysop` slices; selecting a sysop entry runs `tmux switch-client -t orcai-cron` followed by `tmux select-window -t <id>` before quitting
- [x] 5.5 Auto-start the cron daemon: in `cron.WriteEntry` (or the UI layer calling it), if the `orcai-cron` tmux session does not exist after writing, run `orcai cron start` as a background subprocess so the daemon is live without requiring a manual step
- [x] 5.6 Add a unit test in `internal/jumpwindow/jumpwindow_test.go` (or new file) verifying that `listSysopWindows` returns nil when `orcai-cron` session is absent (mock the exec call or use an unreachable session name)
