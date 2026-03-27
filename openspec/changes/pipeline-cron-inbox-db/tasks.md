## 1. Dependencies and Module Setup

- [x] 1.1 Add `modernc.org/sqlite` to `go.mod` and `go.sum` (pure-Go SQLite, no CGo)
- [x] 1.2 Add `robfig/cron/v3` to `go.mod` for cron expression parsing and scheduling
- [x] 1.3 Add `github.com/charmbracelet/log` to `go.mod` if not already present
- [x] 1.4 Add `github.com/fsnotify/fsnotify` to `go.mod` for `cron.yaml` hot-reload

## 2. Result Store (`internal/store`)

- [x] 2.1 Create `internal/store/store.go` with `Store` struct, `Open()`, and `Close()` — opens `~/.local/share/orcai/orcai.db`, enables WAL mode, applies schema migration
- [x] 2.2 Define `runs` table schema migration in `internal/store/schema.go`
- [x] 2.3 Implement `store.RecordRunStart(kind, name string) (id int64, err error)` — inserts an in-flight row
- [x] 2.4 Implement `store.RecordRunComplete(id int64, exitStatus int, stdout, stderr string) error` — updates the row on completion
- [x] 2.5 Implement single-writer goroutine with channel-based write queue in `internal/store/writer.go`
- [x] 2.6 Implement `store.QueryRuns(limit int) ([]Run, error)` for inbox polling
- [x] 2.7 Implement `store.DeleteRun(id int64) error` for inbox item deletion
- [x] 2.8 Implement `store.AutoPrune(maxAgeDays, maxRows int) error` called after each `RecordRunComplete`
- [x] 2.9 Load pruning config from `~/.config/orcai/config.yaml` under `store.retention` with defaults (30 days, 1000 rows)
- [x] 2.10 Write unit tests for store CRUD, WAL mode, write queue, and pruning logic

## 3. ExecutionContext Store Integration (`pipeline/`)

- [x] 3.1 Add `*store.Store` field to `ExecutionContext` in `pipeline/context.go`
- [x] 3.2 Add `WithStore(s *store.Store) ExecutionContextOption` constructor option
- [x] 3.3 Add `ec.DB() *store.Store` accessor method
- [x] 3.4 Update `pipeline.Run()` to accept and pass through a `*store.Store` (nil-safe)
- [x] 3.5 Update `pipeline.Run()` to call `store.RecordRunStart` before execution and `store.RecordRunComplete` after
- [x] 3.6 Write unit tests for `WithStore`, `ec.DB()`, and nil-store safety

## 4. `db` Built-in Step Type (`pipeline/`)

- [x] 4.1 Add `db` case to step dispatch in `pipeline/runner.go` before plugin manager dispatch
- [x] 4.2 Implement `executeDBStep(ctx, step, ec, store)` in `pipeline/step_db.go` supporting `op: query` and `op: exec`
- [x] 4.3 Implement SQL interpolation using existing `template.Interpolate` with `ec.Snapshot()` vars
- [x] 4.4 For `op: query` — execute SELECT, serialize results as JSON, store at `<step-id>.out` in ExecutionContext
- [x] 4.5 For `op: exec` — execute statement, store `rows_affected` at `<step-id>.rows_affected` in ExecutionContext
- [x] 4.6 Return error and halt pipeline if DB step fails (respecting existing retry settings)
- [x] 4.7 Write integration tests for `db` step query, exec, interpolation, and error paths

## 5. Cron Scheduler (`internal/cron`)

- [x] 5.1 Create `internal/cron/scheduler.go` with `Scheduler` struct wrapping `robfig/cron/v3`
- [x] 5.2 Implement `cron.yaml` loading — parse entries with `name`, `schedule`, `kind`, `target`, `args`, optional `timeout`
- [x] 5.3 Implement `fsnotify` watcher for `cron.yaml` hot-reload (debounce 500ms, re-register all entries)
- [x] 5.4 Implement per-entry goroutine dispatch with optional timeout context
- [x] 5.5 Integrate Charm `log` output to both stderr and `~/.local/share/orcai/cron.log` (rotating, max 10MB)
- [x] 5.6 Wire `store.RecordRunStart` / `store.RecordRunComplete` into each scheduled run
- [x] 5.7 Write unit tests for schedule loading, hot-reload, concurrent dispatch, and timeout cancellation

## 6. `orcai cron` Subcommand (`cmd/orcai`)

- [x] 6.1 Create `cmd/orcai/cmd_cron.go` with Cobra `cron` parent command and `start`, `stop`, `list`, `logs` subcommands
- [x] 6.2 Implement `cron start` — check for existing `orcai-cron` tmux session, create session running `orcai cron run`, respect `--force` flag
- [x] 6.3 Implement `cron stop` — kill `orcai-cron` tmux session, print confirmation
- [x] 6.4 Implement `cron list` — load `cron.yaml`, print each entry with next-fire time
- [x] 6.5 Implement `cron logs` — tail `~/.local/share/orcai/cron.log`
- [x] 6.6 Implement `cron run` (internal) — start the `Scheduler`, block until interrupted
- [x] 6.7 Register `cron` command in `cmd/orcai/main.go` root command
- [x] 6.8 Reuse existing `internal/tmux` helpers for session create/kill/detect

## 7. Inbox BubbleTea Model (`internal/inbox`)

- [x] 7.1 Create `internal/inbox/model.go` with `Model` struct implementing `tea.Model` — list view backed by `[]store.Run`
- [x] 7.2 Implement `tea.Tick`-based 5-second poll command using `store.QueryRuns()`
- [x] 7.3 Implement themed `list.DefaultDelegate` for inbox items with color-coded status indicators (success/error/spinner)
- [x] 7.4 Implement full-screen detail modal in `internal/inbox/modal.go` — metadata header, stdout/stderr sections
- [x] 7.5 Implement mutt-style modal keybindings: `n`/`p` (navigate), `r` (re-run), `d` (delete with confirm), `q`/Esc (close)
- [x] 7.6 Implement re-run dispatch — queue the pipeline/agent for immediate execution and show toast notification
- [x] 7.7 Subscribe to `RunCompleted` event bus event for immediate inbox refresh on in-process runs
- [x] 7.8 Ensure all Lipgloss styles reference the active theme (no hard-coded colors)
- [x] 7.9 Write BubbleTea model unit tests for list rendering, modal open/close, and navigation

## 8. Switchboard Integration

- [x] 8.1 Add `InboxModel` to the switchboard model in `internal/switchboard/model.go`
- [x] 8.2 Add `Inbox` entry to the left sidebar navigation alongside existing panels
- [x] 8.3 Wire sidebar navigation key (existing `[` / `]` or tab) to focus the Inbox panel
- [x] 8.4 Pass `*store.Store` reference to the Inbox model at switchboard init
- [x] 8.5 Propagate `tea.Msg` events (poll tick, `RunCompleted`) through the switchboard Update loop to the Inbox model
- [x] 8.6 Verify theme change messages propagate correctly to Inbox panel and modal

## 9. End-to-End Wiring and Config

- [x] 9.1 Open `store.Store` at application startup in `cmd/orcai/main.go`; pass to switchboard and pipeline runner
- [x] 9.2 Create `~/.local/share/orcai/` directory on first run if it does not exist
- [x] 9.3 Document `cron.yaml` format and `store.retention` config in a `cron.yaml.example` file at `~/.config/orcai/`
- [ ] 9.4 Add `orcai cron` usage to the help modal and/or `orcai --help` output

## 10. Verification

- [x] 10.1 Run `go build ./...` — confirm single binary with no CGo, no linker errors
- [x] 10.2 Run all existing tests — confirm no regressions
- [x] 10.3 Run new unit tests for store, cron scheduler, and inbox model
- [ ] 10.4 Manual smoke test: start `orcai cron`, fire a pipeline, verify result appears in Inbox, open modal, navigate, re-run
- [ ] 10.5 Manual smoke test: `db` step in a pipeline — query last run, use result in next step
- [ ] 10.6 Verify binary size delta is acceptable (expected +5-8MB from SQLite driver)
