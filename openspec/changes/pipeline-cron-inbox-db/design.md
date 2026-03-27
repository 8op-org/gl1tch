## Context

ORCAI pipelines and agents are currently fire-and-forget: invoked manually, results printed to stdout, and discarded. There is no scheduling mechanism, no persistent result history, and no way for one run to inform the next. The switchboard TUI reflects this — it has a pipeline runner and feed, but no inbox or historical record.

The goal is to add three tightly coupled capabilities that together form a self-improving loop:
1. A **cron daemon** that fires pipelines and agents on a schedule.
2. A **result store** (SQLite) that captures every run outcome and makes it queryable from within pipelines.
3. An **Inbox panel** in the switchboard that surfaces stored results in a mutt-style reader.

The single-binary constraint and no-CGo requirement (for easy distribution) are hard constraints.

## Goals / Non-Goals

**Goals:**
- `orcai cron` subcommand launches a persistent scheduler daemon in a named tmux session.
- A pure-Go SQLite store captures run results; pipelines and agents can query it mid-run.
- The switchboard gains an Inbox panel with a full-screen modal result reader.
- The result store is the connective tissue: cron writes to it, the inbox reads from it, pipelines consume it.

**Non-Goals:**
- Distributed/multi-host scheduling (single machine only).
- A web UI or REST API for the inbox.
- Real-time push updates to a remote inbox (polling within the TUI session is sufficient).
- Email sending or IMAP integration (the mutt metaphor is UX only).

## Decisions

### Decision: Pure-Go SQLite via `modernc.org/sqlite`
**Rationale**: Keeps the single-binary constraint. `modernc.org/sqlite` is a CGo-free port of SQLite that produces a standard Go binary. The alternative (`mattn/go-sqlite3`) requires CGo and a C compiler, which complicates cross-compilation and distribution.
**Alternatives considered**: BoltDB/bbolt (no SQL, harder to query from pipelines), DuckDB (CGo, overkill), flat JSON files (no concurrent write safety, no query capability).

### Decision: `orcai cron` as a Cobra subcommand that self-daemonizes into tmux
**Rationale**: Keeps a single binary. `orcai cron start` launches a new tmux session named `orcai-cron` and execs `orcai cron run` inside it. `orcai cron list`, `orcai cron stop`, and `orcai cron logs` round out the interface. The Charm `log` library provides structured, leveled output inside the tmux session.
**Alternatives considered**: A separate `orcaid` binary (breaks single-binary requirement), systemd service (not portable to macOS), launchd plist (macOS-only, requires sudo for user sessions).

### Decision: Cron schedule definitions live in `~/.config/orcai/cron.yaml`
**Rationale**: Consistent with the existing `~/.config/orcai/pipelines/` and `~/.config/orcai/wrappers/` convention. Each entry maps a name, cron expression, and pipeline or agent invocation. The scheduler watches this file for changes via `fsnotify` and hot-reloads without restart.
**Alternatives considered**: Inline cron fields in pipeline YAML (mixes concerns, makes pipelines less reusable), a database table for schedules (adds complexity, harder to edit by hand).

### Decision: `db` step type in pipelines; `db` tool in agent context
**Rationale**: The existing step type extension point (plugin dispatch) can be extended with a built-in `db` step type handled directly by the runner before plugin dispatch. This requires no YAML schema changes for other step types. Agents receive a `query_db` tool injected by the agent runner, consistent with how other context tools are injected.
**Alternatives considered**: Environment variable with DB path (too low-level, no query abstraction), a sidecar HTTP server for DB access (adds a port and process dependency).

### Decision: Inbox panel as a new BubbleTea model in `internal/inbox/`
**Rationale**: Follows the existing pattern of `internal/feed/`, `internal/signalboard/`, and `internal/helpmodal/`. The inbox list view is a `list.Model` variant with themed item delegates. The detail modal reuses the full-screen overlay pattern from `internal/helpmodal/`. Items are polled from the store on a configurable interval (default 5s) using a `tea.Tick` command.
**Alternatives considered**: Embedding inbox items into the existing feed (conflates live streaming with historical records), a separate TUI process (breaks the single-window UX).

### Decision: Result store schema — minimal, append-only
```sql
CREATE TABLE runs (
  id          INTEGER PRIMARY KEY,
  kind        TEXT NOT NULL,   -- 'pipeline' | 'agent'
  name        TEXT NOT NULL,
  started_at  INTEGER NOT NULL, -- unix millis
  finished_at INTEGER,
  exit_status INTEGER,          -- 0=ok, non-zero=error, NULL=running
  stdout      TEXT,
  stderr      TEXT,
  metadata    TEXT              -- JSON blob for arbitrary kv
);
```
Runs are never updated in-place after completion (append-only). The `metadata` column holds arbitrary JSON so pipelines can annotate results without schema migrations.

## Risks / Trade-offs

- **SQLite write contention**: The cron daemon and TUI may write concurrently. Mitigation: enable WAL mode (`PRAGMA journal_mode=WAL`) and use a single writer goroutine in the store package with a channel-based queue.
- **`modernc.org/sqlite` binary size**: Adds ~5MB to the binary. Acceptable; noted for release notes.
- **Inbox poll latency**: The 5s default means results appear up to 5s after a run completes. Mitigation: runs can post a `RunCompleted` event to the existing event bus so the inbox can react immediately within the same process.
- **Cron expression parsing**: Using `robfig/cron/v3` (the de-facto standard). Risk of expression syntax mismatch with user expectations (e.g., 5-field vs 6-field). Mitigation: document supported syntax; default to 5-field (no seconds).
- **tmux session collision**: `orcai cron start` will fail if `orcai-cron` session already exists. Mitigation: detect and offer `--force` to kill-and-restart.

## Migration Plan

1. Add `modernc.org/sqlite` and `robfig/cron/v3` to `go.mod`.
2. Create `internal/store/` with schema migration on first open (no external migration tool needed at this scale).
3. Add `internal/cron/` scheduler; wire into `cmd/orcai/` via Cobra.
4. Extend pipeline runner to write results to store; add `db` step type.
5. Add `internal/inbox/` BubbleTea model; integrate into switchboard layout.
6. Add `orcai-cron` session management helpers to `internal/tmux/` (reuse existing tmux package).
7. Existing pipelines run unchanged — no YAML changes required.

## Open Questions

- Should `orcai cron logs` tail the tmux pane output or write to a dedicated log file? (Leaning toward a rotating log file at `~/.local/share/orcai/cron.log` for persistence across tmux session restarts.)
- Inbox item retention policy: keep all, or auto-prune after N days/rows? (Suggest configurable, default 30 days or 1000 rows, whichever comes first.)
- Agent `query_db` tool: read-only or read-write? (Leaning read-write with explicit write step in pipeline context; agents get read-only by default to prevent runaway writes.)
