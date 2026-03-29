## 1. Signal Board Mark State

- [x] 1.1 Add `marked map[string]bool` field to `SignalBoard` struct in `signal_board.go`
- [x] 1.2 Add `toggleMark(id string)` method on `SignalBoard` that adds or removes the entry ID from `marked`
- [x] 1.3 Add `hasMarks() bool` helper method on `SignalBoard`
- [x] 1.4 Add `markedTitles(feed []feedEntry) []string` helper method that returns titles of marked entries in feed order, capped at 20

## 2. Signal Board Key Handling

- [x] 2.1 Add `m` key case in the signal board key handler in `switchboard.go`: call `toggleMark` on the currently selected entry's ID
- [x] 2.2 Add `r` key case in the signal board key handler: if `hasMarks()` is true, build the injected prompt string from `markedTitles`, open the agent modal with that prompt, focus slot 2, and clear all marks
- [x] 2.3 Ensure `r` with no marks is a no-op (no modal open)

## 3. Signal Board Row Rendering

- [x] 3.1 In `buildSignalBoard` row rendering loop, detect if `absIdx` entry ID is in `marked` and apply a highlight background (`pal.SelBG` or equivalent) to the full row content when marked
- [x] 3.2 Add `m mark` hint to the `sbHints` slice in the non-search focused state
- [x] 3.3 Add `r run` hint conditionally when `m.signalBoard.hasMarks()` is true

## 4. Agent Runner Modal: Width and Pre-populated Prompt

- [x] 4.1 Change modal width formula in `viewAgentModalBox` from `min(max(w-4, 60), 90)` to `max(w*9/10, 60)` clamped to `w-2`
- [x] 4.2 Add an `initialPrompt string` field to the agent modal open path (e.g., a field on the model or passed through the open action)
- [x] 4.3 When opening the modal via `r`, set `initialPrompt` to the joined marked titles and set `agentModalFocus = 2`
- [x] 4.4 In `viewAgentModalBox` / modal init, if `initialPrompt` is non-empty, pre-populate the `textarea.Model` with that string on open
- [x] 4.5 Clear `initialPrompt` after it has been applied to the textarea (on first render after open)

## 5. Feed Step Vertical Layout

- [x] 5.1 Add `lines []string` field to `StepInfo` struct in `switchboard.go` to hold per-step output
- [x] 5.2 Replace the horizontal step wrapping loop in the feed body builder with a vertical loop: one `boxRow`-style line per step badge, indented 2 spaces
- [x] 5.3 After each step badge line, render `step.lines` (last 5 at most), each indented 4 spaces

## 6. Tests

- [x] 6.1 Add test: toggling mark on an entry adds it to `marked`, toggling again removes it
- [x] 6.2 Add test: `r` key with marked entries opens the modal with `AgentModalOpen() == true` and the prompt textarea contains the expected titles
- [x] 6.3 Add test: `r` key with no marks does not open the modal
- [x] 6.4 Add test: `viewAgentModalBox` width is `w*9/10` for a given terminal width
- [x] 6.5 Add test: `buildSignalBoard` output includes `m mark` hint when focused and not searching
- [x] 6.6 Add test: `buildSignalBoard` output includes `r run` hint when focused and entries are marked
- [x] 6.7 Add test: vertical step layout — feed body contains one line per step rather than a single wrapped line
