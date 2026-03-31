## Why

The prompt builder and pipeline builder duplicate significant UI logic (sidebar, test runner, editor panel) and diverge in design. Unifying them into shared components and a consistent two-column layout removes duplication, ensures both tools evolve together, and makes the jump window reusable across subcommands.

## What Changes

- Extract shared components from `promptbuilder` into a reusable `internal/buildershared` package: sidebar (list + search), test runner panel, and editor area.
- Redesign both `orcai prompt-builder` and `orcai pipeline-builder` as subcommands under their respective top-level commands, each using the shared components with a two-column layout:
  - **Left column**: sidebar (existing list/search behavior)
  - **Right column top**: prompt/pipeline feedback loop — shows the current prompt/pipeline draft; clears when first prompt is sent, waits for response; `ctrl+s` saves; `ctrl+r` injects first prompt into test runner
  - **Right column bottom**: chat/agent runner prompt input (the existing agent runner modal repurposed as an inline widget)
- **BREAKING**: `orcai pipeline build` is renamed to `orcai pipeline-builder` (new subcommand entry-point); `orcai prompts tui` is supplemented/replaced by `orcai prompt-builder`.
- Pipeline builder inherits the same editor component used in prompt builder.
- Jump window extracted to `orcai widget jump-window` subcommand so it can be called from any parent command, not just from the switchboard overlay.
- Jump window updated to open `orcai prompt-builder` and `orcai pipeline-builder` in new tmux windows (replacing the current `orcai prompts tui` / `orcai pipeline build` calls).

## Capabilities

### New Capabilities

- `builder-shared-components`: Reusable sidebar, test runner, and editor BubbleTea components extracted from promptbuilder, consumable by both prompt-builder and pipeline-builder.
- `prompt-builder-subcommand`: `orcai prompt-builder` — standalone two-column prompt builder TUI with feedback loop and inline agent runner.
- `pipeline-builder-subcommand`: `orcai pipeline-builder` — standalone two-column pipeline builder TUI sharing builder-shared-components and same editor as prompt-builder.
- `builder-feedback-loop`: Shared feedback loop panel: send prompt → clear → await response; `ctrl+s` to save, `ctrl+r` to re-inject into test runner.
- `jump-window-widget`: `orcai widget jump-window` — extracted jump window usable from any subcommand context, opens builder subcommands in tmux windows.

### Modified Capabilities

- `jumpwindow`: Jump window sysop entries updated to call `orcai prompt-builder` and `orcai pipeline-builder` instead of legacy commands, and is now also surfaced as `orcai widget jump-window`.

## Impact

- `internal/promptbuilder` — sidebar, editor, test-runner extracted to `internal/buildershared`; BubbleModel updated to use shared components
- `internal/pipelineeditor` — refactored to use shared editor/sidebar/runner from `internal/buildershared`
- `internal/jumpwindow` — `PipelinesMsg` / prompts entries updated to launch new subcommands; embedded model promoted to `orcai widget jump-window`
- `cmd/pipeline.go` — adds `pipeline-builder` subcommand
- `cmd/cmd_prompts.go` — adds `prompt-builder` subcommand
- `cmd/widget.go` (new) — `orcai widget` group with `jump-window` subcommand
- No changes to pipeline execution, store, or plugin layers
