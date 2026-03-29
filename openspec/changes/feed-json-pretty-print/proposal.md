## Why

Pipeline step output and agent responses frequently contain raw JSON — GitHub API results, tool call outputs, search results — that renders as dense single-line blobs in the Activity Feed. These blobs are unreadable at a glance and wrap across dozens of screen columns. Collapsible pretty-printed JSON lets users quickly inspect structured output without leaving the feed.

The recent scrolling/cursor-bounds fix introduced panel-width-dependent line counting into the marking system, which broke mark indices — marks now land on the wrong lines. The fix is to decouple marking from wrap-count arithmetic and instead derive mark indices directly from the renderer's own output (the `allLines` slice), which is already the ground truth.

A feedline parser layer makes both features composable: as the renderer walks each raw output string, parsers run first and may emit multiple styled display lines. JSON is the first parser; future parsers (diffs, log levels, ANSI passthrough) plug in without touching core renderer logic.

## What Changes

- **Feedline parser system**: A small `FeedLineParser` interface (or function signature) that takes a raw line string and width and returns zero or more display lines. Parsers are tried in order; the first match wins.
- **JSON parser**: Detects lines where the trimmed value is valid JSON (`{…}` or `[…]`). Collapsed default renders a single summary row (`{ … }` or `[ N items ]`). `enter` on the cursor line toggles expand/collapse. Expanded renders pretty-printed with 2-space indent, syntax-highlighted (keys in accent, strings dimmed, numbers/booleans in success color), capped at 20 visible lines with `… N more` overflow.
- **Marking fix**: `feedToggleMark` and `totalFeedLines` stop calling `feedPanelWidth()` at mark time. Instead, `viewActivityFeed` passes its `allLines` slice length as the authoritative count; marks are keyed on the `appendRow` index, which is identical between the renderer and the raw-line builder because both use the same parser pipeline.
- **Hint bar update**: `enter` hint added when feed cursor is on an expandable JSON line.

## Capabilities

### New Capabilities
- `feed-json-viewer`: Collapsible, syntax-highlighted JSON viewer embedded in Activity Feed line renderer
- `feed-line-parsers`: Extensible feedline parser pipeline — parsers registered in order, first match wins

### Modified Capabilities
- `feed-step-output`: Step output lines run through the parser pipeline; JSON blobs collapse instead of word-wrapping raw text

## Impact

- `internal/switchboard/switchboard.go` — `viewActivityFeed`, `feedRawLines`, `totalFeedLines`, key handler (`enter`), marking logic
- New file: `internal/switchboard/feed_parsers.go` — `FeedLineParser` type, `jsonFeedLineParser`, parser registry
- New model state: `feedJSONExpanded map[string]bool` keyed by `"entryIndex:lineIndex"`
- No new dependencies — `encoding/json` already imported
