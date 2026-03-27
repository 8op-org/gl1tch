## Context

The inbox panel was added in `pipeline-cron-inbox-db`. The store is opened at startup in `bootstrap`/`switchboard.Run()` and passed to `NewWithStore`, which forwards it to `inbox.New`. However:

- The store reference is dropped after construction — it is not stored on `Model`.
- Pipelines launched from the TUI exec `orcai pipeline run <yaml>` in a new tmux window. That subprocess calls `cmd/pipeline.go` → `pipeline.Run()` with no store, so nothing is recorded.
- `viewBottomBar` has no `inboxFocused` branch; the user sees launcher hints that don't apply.

## Goals / Non-Goals

**Goals:**
- Inbox shows correct key hints (open · delete · re-run · tab · i shortcut) when focused.
- Every `orcai pipeline run` invocation (TUI-launched or CLI-direct) records its run in `~/.local/share/orcai/orcai.db`.
- Switchboard `Model` retains the store for future operations (e.g., re-run dispatch).

**Non-Goals:**
- Agent runner store integration (separate task).
- Changing how the cron scheduler records runs (already correct).
- Modifying the inbox BubbleTea model internals.

## Decisions

**Store retained on Model** — Add `store *store.Store` to `Model` and populate it in `NewWithStore`. Cost: one pointer. Benefit: any future code in the switchboard (re-run, delete from inbox, etc.) can access it without threading it through call args.

**Store opened in CLI pipeline runner** — `pipelineRunCmd` in `cmd/pipeline.go` calls `store.Open()`, defers `Close()`, and passes `WithRunStore(s)`. This is the correct seam: all pipeline runs, regardless of whether triggered from TUI or CLI, go through this function. The store open is a fast no-op on subsequent calls (WAL mode). Nil-safe guard already exists in runner.go so a nil store is handled gracefully.

**Bottom-bar hint set for inbox** — Added as a new `case m.inboxFocused:` before the default case in `viewBottomBar`. Hints: `enter` open · `d` delete · `r` re-run · `q`/`Esc` close modal · `tab` focus cycle · `i` inbox.

## Risks / Trade-offs

- Opening the store in `cmd/pipeline.go` adds a small startup cost (~1ms) for every CLI pipeline invocation. Acceptable given the value of recorded results.
- If `~/.local/share/orcai/` doesn't exist yet, `store.Open()` must create it — this is already handled by `store.Open` from the previous change.
