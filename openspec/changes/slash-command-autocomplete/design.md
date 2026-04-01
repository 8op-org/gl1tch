## Context

The GLITCH chat input (`glitchChatPanel`) uses a single `charmbracelet/bubbles/textinput.Model`. Slash commands are parsed only after the user hits Enter. No inline feedback is given while typing. The codebase already has a reusable fuzzy-scoring function in `internal/modal/fuzzypicker.go` (`fuzzyScore`) and a consistent pattern for rendering overlay boxes (used by the model-picker and theme-picker in `glitch_panel.go`).

## Goals / Non-Goals

**Goals:**
- Show a ranked suggestion list above the chat input whenever the value starts with `/`
- Filter suggestions in real-time as the user continues typing
- Let the user navigate and select with Tab / Up / Down / Enter
- Insert the selected command (with a trailing space) into the input on selection
- Dismiss the list cleanly on Esc or when the text no longer starts with `/`
- Zero new external dependencies

**Non-Goals:**
- Argument-level autocomplete (e.g. completing theme names after `/themes `) — deferred
- A separate modal picker for slash commands — the inline overlay is sufficient
- Persisting or learning from command usage frequency

## Decisions

### D1 — Inline overlay, not a modal

A full-screen or focus-stealing modal would break typing flow. The existing `modelPickerBox` overlay pattern in `glitch_panel.go` renders a lipgloss-styled box directly above the input row; this is the right pattern to reuse.

*Alternatives considered:* Full FuzzyPickerModel modal — rejected because it steals focus and requires two keystrokes (open, then close) to dismiss.

### D2 — Self-contained autocomplete state inside `glitchChatPanel`

Add three fields to the panel struct:
```go
acSuggestions []slashSuggestion  // filtered list
acCursor      int                // selected index
acActive      bool               // overlay visible?
```
`slashSuggestion` is a small local struct `{ cmd, hint string }`.

*Alternatives considered:* Extracting a separate `autocomplete` package — unnecessary abstraction for what is a ~150-line addition to a single file.

### D3 — Reuse `fuzzyScore` from `internal/modal/fuzzypicker.go`

The scoring function is already exported-enough for internal reuse and handles prefix/substring/scattered matching correctly. Filter+rank on every keypress; the list is small (≤15 commands) so no performance concern.

### D4 — Intercept Tab / Up / Down before the textinput gets them

BubbleTea message routing lets the panel consume navigation keys when `acActive == true` and pass everything else through to `p.input.Update(msg)` as normal. No modification to textinput or any shared component.

### D5 — Insert with trailing space, cursor positioned after

On selection: `p.input.SetValue(selected.cmd + " ")` and `p.input.CursorEnd()`. This is the most ergonomic behaviour (user can immediately type arguments).

## Risks / Trade-offs

- **Narrow panel widths** → suggestion box may overflow: mitigate by capping box width to `p.width` and truncating hint text
- **New commands added later** → the static command list in the autocomplete source needs to stay in sync with the switch statement: mitigate by defining suggestions as a package-level `var` adjacent to the switch, with a comment pointing both ways
- **Tab key conflict** — Tab currently does nothing in the input; safe to repurpose for autocomplete navigation

## Migration Plan

Purely additive change. No existing behaviour is modified. No data migration. Rollback = revert the single file.

## Open Questions

- Should `/quit` and `/exit` be excluded from suggestions to avoid accidental exits? (Lean: keep them, users should see they exist)
