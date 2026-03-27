## Why

Pipelines and agents run once and vanish — there is no way to schedule recurring work, no place to review results, and no shared memory between runs. These three gaps together prevent ORCAI from acting as a persistent, self-improving workspace.

## What Changes

- Add `orcai cron` subcommand that launches a long-lived scheduler daemon in a dedicated tmux session, executing pipelines and agent runs on cron expressions using Charm's `log` library for structured output.
- Add a background SQLite store (`orcai.db`) that captures every pipeline and agent run result; pipelines can query and write to this store mid-run, enabling iterative, data-driven workflows.
- Add an **Inbox** panel to the switchboard left sidebar (between the existing nav items) showing a live feed of run results; selecting an entry opens a full-screen mutt-style modal with the full output, metadata, and re-run controls.

## Capabilities

### New Capabilities

- `pipeline-cron-scheduler`: Recurring pipeline and agent execution via cron expressions. The `orcai cron` subcommand starts a Charm-powered daemon in a named tmux session; each schedule entry maps a cron expression to a pipeline or agent with optional argument overrides.
- `pipeline-result-store`: SQLite-backed result store (`orcai.db`) that captures run metadata (pipeline name, timestamps, exit status, stdout/stderr). Pipelines access it via an injected `db` step type and template variable `{{db.*}}`; agents receive it as a tool in their context.
- `switchboard-inbox`: New left-sidebar panel labelled **Inbox** in the switchboard TUI. Displays a scrollable list of recent run results from the result store, fully theme-aware. Selecting a result opens a full-screen modal with mutt-style navigation (next/prev, reply-as-rerun, delete).

### Modified Capabilities

- `pipeline-execution-context`: Pipeline `ExecutionContext` gains a `db` accessor so steps can read from and write to the result store via dot-path syntax (e.g. `{{db.last.out}}`).

## Impact

- New `internal/cron/` package (scheduler, Charm log integration).
- New `internal/store/` package (SQLite schema, query helpers via `modernc.org/sqlite` — no CGo).
- New `internal/inbox/` package (BubbleTea model for inbox panel and modal).
- `cmd/orcai/main.go` gains `cron` subcommand via Cobra.
- `internal/switchboard/` updated to include inbox panel in layout and sidebar nav.
- `pipeline/` runner updated to write results to store after each run.
- No breaking changes to existing pipeline YAML syntax; `db` step type is additive.
- Single binary preserved — SQLite driver is pure Go.
