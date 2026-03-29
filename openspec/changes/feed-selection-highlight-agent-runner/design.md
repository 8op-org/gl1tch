## Context

The signal board (`internal/switchboard/signal_board.go`, `switchboard.go`) renders a scrollable list of `feedEntry` items with cursor navigation via `j/k`. Each entry has an `id`, `title`, `status`, `lines` (output), and `steps []StepInfo`.

The agent runner modal (`viewAgentModalBox`) opens as a floating overlay centered over the switchboard. Its width is currently `min(max(w-4, 60), 90)` — hard-capped at 90 columns regardless of terminal width. The prompt is a `textarea.Model` that the user fills manually.

Pipeline steps are rendered horizontally inside the feed body, wrapping with `·` separators when they exceed `width-4`. Output lines from each step are shown below but are not associated per-step.

## Goals / Non-Goals

**Goals:**
- Let the user mark (highlight) one or more signal board rows by pressing `m`
- Pressing `r` on a focused signal board collects marked entries' titles into the agent modal prompt
- Agent runner modal opens at 90% of terminal width
- Pipeline step entries in the feed are rendered one-per-line (vertical), each followed by any output lines that belong to that step

**Non-Goals:**
- Persisting the marked set across sessions
- Marking entries outside the signal board (e.g., from the pipeline launcher)
- Per-step output streaming (steps do not yet carry per-step output; this change only alters the step badge layout)

## Decisions

### Decision: Mark state stored as `map[string]bool` on `SignalBoard`

**Alternatives considered:**
- Slice of indices: indices shift when the filtered list changes; IDs are stable.
- Bit-flag on `feedEntry`: mutates the shared feed slice for a UI-only concern; cleaner to keep it in the `SignalBoard` struct.

**Rationale:** `feedEntry.id` is already a stable, unique string. A `map[string]bool` keyed on entry ID handles filter/scroll changes without index math.

### Decision: Mark key is `m`, confirm-and-open key is `r`

**Alternatives considered:**
- `space` for mark: conflicts with future scrolling gestures and is not visually distinct in the hint bar.
- `enter` for run: `enter` already navigates to the tmux window for that entry.

**Rationale:** `m` = "mark" is mnemonic and unused in the current hint set; `r` = "run" is consistent with how the launcher uses `enter` for immediate launch but `r` for the runner modal.

### Decision: Injected prompt is the list of marked entry titles, newline-separated

**Alternatives considered:**
- Injecting full `lines` output: output can be very long and overwhelms the textarea.
- Injecting JSON: unnecessary structure for a text prompt.

**Rationale:** Titles give the agent enough context. The user can edit the textarea before submitting.

### Decision: Modal width becomes `w * 9 / 10` (integer arithmetic, floored)

**Alternatives considered:**
- `w - 4`: was the previous formula; too narrow on wide terminals.
- `lipgloss`-based percentage: no dependency needed for a single multiply-and-divide.

**Rationale:** 90% of terminal width scales naturally. A minimum of 60 columns is preserved for narrow terminals.

### Decision: Step display is vertical — one badge line per step, then that step's output lines

`StepInfo` does not currently carry per-step output lines. The vertical layout change renders each step badge on its own line. Associating output with steps is deferred until `StepInfo` carries a `lines []string` field (a separate change).

## Risks / Trade-offs

- [Risk] Marked set goes stale when entries are archived or filtered out → Mitigation: when building the injected prompt, silently skip IDs not present in the current feed.
- [Risk] Large marked sets produce a very long injected prompt → Mitigation: cap injection at the first 20 marked entries, appending a note if truncated.
- [Risk] Vertical step layout increases feed height significantly for pipelines with many steps → Mitigation: the existing `feedLinesPerEntry = 10` cap continues to apply to output lines; step badges are not capped but are individually short.

## Open Questions

- Should clearing marks happen automatically after `r` submits, or should marks persist so the user can run the same context again? (Default: clear on submit.)
- Should the modal open with focus on the prompt field (slot 2) when launched via `r`? (Default: yes, skip provider/model selection to the prompt.)
