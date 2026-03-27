## Context

The cron/scheduler infrastructure (`internal/cron/`) and result store were introduced in the `pipeline-cron-inbox-db` change. The scheduler hot-reloads `~/.config/orcai/cron.yaml` via fsnotify within ~1 s of a write. The TUI currently has no way to write to that file — users must edit it manually and the scheduler must already be running as a daemon.

Two UI surfaces need scheduling:
1. **Agent runner modal** (`internal/switchboard/switchboard.go`) — focus slot 3 currently opens a dir picker for WORKING DIRECTORY. We replace this with a SCHEDULE text input (blank = run once in cwd, non-blank = cron expression → write to `cron.yaml`).
2. **Pipeline launcher** — after selecting a pipeline, the TUI immediately opens a dir picker. We insert a mode-select step before the dir picker: **Run now** (existing path) or **Schedule recurring** (cron expression input → write to `cron.yaml`).

## Goals / Non-Goals

**Goals:**
- Users can schedule an agent run as a recurring cron job from the agent modal without leaving the TUI.
- Users can schedule a pipeline as a recurring cron job from the pipeline launcher without leaving the TUI.
- A `cron.WriteEntry` / `cron.RemoveEntry` API safely mutates `cron.yaml` from the TUI process; the daemon picks up changes automatically.
- An integration test verifies a pipeline completes end-to-end when invoked via the `orcai` CLI.

**Non-Goals:**
- UI to list, edit, or delete existing cron entries (future work).
- Cron expression validation beyond basic syntax (let robfig/cron surface parse errors).
- Starting the cron daemon from the TUI (daemon lifecycle is separate).

## Decisions

### 1. Reuse focus slot 3 in the agent modal as SCHEDULE, not a new slot

The modal currently cycles through 4 focus areas (provider → model → prompt → cwd). Slot 3 (cwd) opens a dir picker overlay; there is no free text input. We replace that slot with a `textarea` pre-filled with blank (meaning "run now in default cwd"). The label changes from `WORKING DIRECTORY` to `SCHEDULE (cron expression, blank = run now)`.

**Alternatives considered:**
- Add a 5th slot after cwd for schedule. Rejected: makes the modal taller and the cwd field is rarely useful for agents (they inherit the switchboard's launchCWD).
- Keep cwd slot, add schedule as an optional overlay after submit. Rejected: more state, harder to discover.

### 2. Pipeline mode-select overlay before dir picker

Insert a lightweight two-item menu (`Run now` / `Schedule recurring`) that appears when a pipeline is selected. Choosing `Run now` proceeds to the existing dir picker. Choosing `Schedule recurring` swaps to a single text input for the cron expression; confirming writes to `cron.yaml` and shows a feed item (no dir picker needed — the scheduler runs pipelines from the pipeline's own directory or the configured `cwd` in `cron.yaml`).

**Alternatives considered:**
- Add a toggle inside the existing dir picker. Rejected: dir picker is already a separate model; adding a mode toggle there couples unrelated concerns.
- Always show dir picker then ask about schedule. Rejected: confusing if no tmux window ever opens.

### 3. `cron.WriteEntry` / `cron.RemoveEntry` with file-level locking

New file `internal/cron/writer.go`. Uses `os.OpenFile` with `O_RDWR|O_CREATE` + `syscall.Flock` (or a pure-Go flock shim for portability) to prevent concurrent writes. Reads the YAML, appends/removes the entry, and atomically renames a temp file into place. The running daemon's fsnotify watcher fires within ~1 s.

**Alternatives considered:**
- Write from the daemon via an IPC socket. Rejected: adds daemon complexity; TUI writing directly is simpler since the daemon hot-reloads anyway.
- No locking. Rejected: TUI could be opened in multiple windows simultaneously.

### 4. Integration test via `orcai` CLI

A `_test.go` file in `internal/pipeline/` (or a new `test/integration/` package) builds the binary with `go build`, runs `orcai pipeline run <fixture.yaml>`, and asserts exit 0 within 5 s. Uses a minimal fixture pipeline with a single `shell` step (`echo ok`).

**Why not a unit test?** The proposal explicitly asks for end-to-end confidence that the CLI path works, which requires the full binary.

## Risks / Trade-offs

- **fsnotify race on slow filesystems**: The daemon may not pick up `cron.yaml` changes before the user expects the next scheduled run. Mitigation: show the confirmed schedule time in the feed entry so the user can verify.
- **Cron expression UX**: Free text input for cron expressions is error-prone. Mitigation: parse the expression with `robfig/cron` immediately on submit and show an inline error if invalid; do not write to `cron.yaml` on parse failure.
- **Agent modal slot repurpose**: Removing the dir picker from slot 3 may surprise users who relied on it. Mitigation: agents already default to the switchboard's `launchCWD`; the dir picker was rarely used. Document the change in the feed hint bar.

## Open Questions

- Should the pipeline mode-select overlay also allow changing the working directory when scheduling (i.e., store `cwd` in the cron entry)? Current proposal: use blank cwd in the cron entry (scheduler defaults to pipeline file's directory). Revisit if users request per-schedule cwd overrides.
