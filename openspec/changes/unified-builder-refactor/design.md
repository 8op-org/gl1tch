## Context

Both `internal/promptbuilder` and `internal/pipelineeditor` implement three-panel TUI layouts with a sidebar list, an editor area, and a test runner. They were built independently and have diverged: promptbuilder uses a terminal-escape-based renderer while pipelineeditor uses Lipgloss/BubbleTea widgets. The jump window lives in `internal/jumpwindow` and is embedded only in the switchboard; its sysop entries hard-code the legacy commands `orcai prompts tui` and the old `orcai pipeline build`.

The desired end state is:
1. A shared `internal/buildershared` package that owns the sidebar, test runner panel, and editor widget.
2. Both builders use a unified two-column layout: sidebar left, (feedback loop + agent runner chat) right.
3. Jump window promoted to `orcai widget jump-window`, launching the new subcommands.

## Goals / Non-Goals

**Goals:**
- Single sidebar implementation used by both builders.
- Single editor/test-runner implementation used by both builders.
- `orcai prompt-builder` and `orcai pipeline-builder` as top-level-adjacent subcommands launchable from the CLI and from tmux via jump window.
- Consistent two-column layout: left = sidebar, right-top = feedback loop panel, right-bottom = agent runner chat input.
- Feedback loop: on first send, panel clears and awaits response; `ctrl+s` saves; `ctrl+r` reinjects first prompt into test runner.
- `orcai widget jump-window` as an independently callable subcommand.

**Non-Goals:**
- Changing pipeline execution, store, or plugin layers.
- Merging prompts and pipelines into a single data model.
- Removing backward-compatible `orcai prompts tui` / `orcai pipeline build` (they can remain as aliases or be deprecated later).

## Decisions

### 1. New `internal/buildershared` package instead of making one builder depend on the other

**Decision**: Extract shared UI components into `internal/buildershared`.

**Rationale**: Importing `promptbuilder` from `pipelineeditor` or vice versa creates a dependency cycle risk and tight coupling between domain concerns. A neutral shared package avoids this.

**Alternatives considered**:
- Extend promptbuilder and have pipelineeditor import it — rejected because it couples pipeline domain logic to the prompt builder's data model.
- Duplicate and accept divergence — rejected; we've already seen the cost of this.

### 2. Right-column layout: feedback loop above, agent runner chat below

**Decision**: The right column is split vertically. The upper portion shows the current prompt/pipeline content and becomes the streaming response display once the user submits. The lower portion is a persistent chat input (repurposed agent runner input widget).

**Rationale**: Mirrors familiar chat-app patterns. The feedback loop area doubles as both the draft editor and the response display, keeping the screen uncluttered.

**Alternatives considered**:
- Separate tabs for draft vs response — rejected as more keystrokes to a common workflow.
- Modal overlay for response — rejected; already present in current design and users have to dismiss it.

### 3. `orcai widget jump-window` as new subcommand surface

**Decision**: Add `cmd/widget.go` with a `widget` cobra command group; `jump-window` is a subcommand that runs the same embedded model as today but as a standalone alt-screen program.

**Rationale**: Jump window is useful from other subcommands (e.g., `orcai prompt-builder` could bind a key to spawn it). Making it a proper subcommand allows `os/exec` invocation without reimplementing the TUI.

**Alternatives considered**:
- Keep it internal-only and import the package — works but doesn't enable external invocation from scripts or other subcommands that don't embed it.

### 4. Shared components use interface-based composition, not embedding

**Decision**: `buildershared` exports `Sidebar`, `EditorPanel`, and `RunnerPanel` as independent BubbleTea sub-models with `Update(tea.Msg) (Self, tea.Cmd)` and `View() string` methods. Parent models compose them as fields.

**Rationale**: Embedding BubbleTea models via struct embedding causes message-routing confusion when parent and child both handle the same message types. Explicit field composition with delegation is idiomatic in this codebase.

### 5. Feedback loop clears on first send, not on every send

**Decision**: The feedback loop area shows the draft prompt/pipeline. On first `ctrl+enter` (or send), it clears to a loading indicator. Subsequent sends append to the history in the same area.

**Rationale**: Matches the UX description: "once you send your first prompt, it clears out waiting for the response." Preserving history after the first interaction is useful context.

## Risks / Trade-offs

- **Regression in promptbuilder rendering** → Mitigation: keep existing `BubbleModel` as a thin wrapper until the shared components are verified; flip behind a build tag if needed.
- **pipelineeditor refactor scope creep** → Mitigation: the task list gates each component extraction separately; pipeline YAML editor (FocusYAML pane) stays in pipelineeditor and is not shared.
- **Jump window embedded vs standalone modes** → Both modes must still work. The `embedded` flag on the model continues to control whether it sends `CloseMsg` vs `tea.Quit`.

## Open Questions

- Should `ctrl+s` in the feedback loop save both the prompt text and the pipeline YAML, or only the prompt/pipeline name + content that was last edited in the sidebar?
- Should `orcai pipeline build` be deprecated immediately or kept as an alias pointing to `orcai pipeline-builder`?
