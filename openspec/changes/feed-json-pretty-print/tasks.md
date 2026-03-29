## 1. Feedline Parser Infrastructure

- [x] 1.1 Create `internal/switchboard/feed_parsers.go` with `FeedLineParser` function type and a package-level `feedParsers []FeedLineParser` registry
- [x] 1.2 Implement `runFeedLineParsers(raw string, width int, expanded bool) ([]string, bool)` that iterates the registry and returns on first match

## 2. JSON Parser

- [x] 2.1 Implement `jsonFeedLineParser` — detect valid JSON (trim → check `{`/`[` prefix → `json.Valid`), return collapsed summary row when `expanded: false`
- [x] 2.2 Implement collapsed summary: object shows `▸ { … } (N keys)`, array shows `▸ [ N items ]` with dim/accent styling from `ANSIPalette`
- [x] 2.3 Implement expanded pretty-print path: `json.MarshalIndent` → syntax-highlight keys (accent), string values (dim), numbers/booleans (success color), cap at 20 lines + `… N more` overflow
- [x] 2.4 Register `jsonFeedLineParser` as first entry in `feedParsers`

## 3. Renderer Integration

- [x] 3.1 Add `feedCachedLines []string` and `feedJSONExpanded map[string]bool` fields to `Model`
- [x] 3.2 Refactor `viewActivityFeed`: use `logicalIdx` counter in `appendRow`; add `appendExtra` for non-navigable expansion lines
- [x] 3.3 Run `runFeedLineParsers` for each `entry.lines` output line inside `viewActivityFeed`, using `logicalIdx` for expanded-state key
- [x] 3.4 Run `runFeedLineParsers` for each `step.lines` output line inside `viewActivityFeed` (step output also collapses JSON)
- [x] 3.5 Mirror the parser calls in `feedRawLines` so raw-line indices stay aligned with rendered indices

## 4. Marking Fix

- [x] 4.1 Replace all `totalFeedLines(m.feed, m.feedPanelWidth())` call sites with `m.feedLineCount()` — a helper that returns `len(feedRawLines())` with fallback to 1
- [x] 4.2 `feedToggleMark` reads from `feedRawLines` which is now consistent with renderer (single source of truth)
- [x] 4.3 Clear `feedJSONExpanded` and reset `feedCursor`/`feedScrollOffset` when new entries are prepended (in the `FeedEntryMsg` handler)

## 5. Enter Key Handler

- [x] 5.1 Add `enter` handler in `handleEnter` — checks if cursor rawLine starts with `jsonIndicator`, toggles `feedJSONExpanded[feedCursor]`
- [x] 5.2 Update hint bar: when cursor is on a JSON line show `enter  expand` or `enter  collapse` based on expand state

## 6. Tests

- [x] 6.1 Unit test `jsonFeedLineParser` — collapsed summary for object and array, expanded output capped at 20 lines, non-JSON passthrough
- [x] 6.2 Unit test `runFeedLineParsers` — first-match semantics, no-match fallthrough
- [x] 6.3 Integration test: `FeedLineCount()` stable across navigation; enter toggles expand/collapse in view
- [x] 6.4 Verify existing mark-mode tests still pass (`switchboard_test.go`)
