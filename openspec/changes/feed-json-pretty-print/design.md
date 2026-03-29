## Context

The Activity Feed renders entries with a title line, optional cwd, step badge lines, and per-entry output lines. Output lines currently word-wrap raw text. When a pipeline step calls a JSON API (GitHub, Ollama, etc.) the entire response lands as a single line — it word-wraps to dozens of rows, is visually unreadable, and inflates the line count used for navigation math.

The recent `totalFeedLines` rewrite added panel-width-dependent word-wrap counting to keep cursor bounds tight after the scrolling fix. That same width is threaded into `feedRawLines` and `feedToggleMark`. The result: if `feedPanelWidth()` returns a different value at mark-time vs render-time (e.g., window resize mid-session), or if the wrap math diverges between `totalFeedLines` and `viewActivityFeed`, marks land on incorrect lines.

The root cause is that marking borrows its index space from a different computation than the renderer. The renderer's `allLines` slice is the only authoritative index.

## Goals / Non-Goals

**Goals:**
- JSON lines in feed output render collapsed by default; `enter` toggles expand/collapse
- Expanded JSON is pretty-printed with syntax highlighting, capped at 20 lines
- Marking indices are always derived from the renderer's own output — no separate width-dependent computation
- A feedline parser interface lets future formatters plug in without touching the renderer loop

**Non-Goals:**
- Inbox Detail JSON viewer (separate change if desired)
- Nested collapse (inner JSON objects within an expanded object)
- Editable JSON
- Parser hot-reload or config-file registration

## Decisions

### 1. Feedline parser as a function type, not an interface

A `type FeedLineParser func(raw string, width int, expanded bool) (lines []string, matched bool)` is simpler than an interface with a `Matches` + `Render` split. Parsers are pure functions; they receive the line and width and return display lines. The registry is a package-level slice iterated in order.

**Alternative considered**: Interface with `Match(string) bool` + `Render(...)` methods. Adds ceremony with no benefit here — the function type is sufficient and easier to test.

### 2. Marking uses `appendRow` index directly

`viewActivityFeed` already tracks the current line index via `idx := len(*lines)` inside `appendRow`. That index is the mark key. `feedToggleMark` stores content by calling `feedRawLines` which mirrors the same loop structure — they must stay in sync. **The fix**: remove width-dependent wrap counting from `totalFeedLines` and instead return `len(allLines)` from the renderer itself. Navigation (`handleDown`, `handleUp`, `G`) call a new `m.feedLineCount()` helper that builds `allLines` in a dry run (or reuses the last render if we cache it).

**Simpler alternative**: Cache `allLines` on the model after each render. Since BubbleTea renders are pure and frequent, the cache is always fresh. Mark indices read directly from the cached slice — no separate raw-line builder needed.

**Decision**: Cache `allLines []string` on the model (set during `viewActivityFeed`). `feedToggleMark` uses `m.feedCachedLines`. `totalFeedLines` is replaced by `len(m.feedCachedLines)`. This eliminates the width-math duplication entirely.

### 3. JSON expand state keyed by stable string key

JSON expand state is `feedJSONExpanded map[string]bool` keyed by `fmt.Sprintf("%d:%d", entryIdx, lineIdx)` — entry position in `m.feed` slice + absolute output-line index within that entry. Entry index is stable within a session; line index is stable for a given entry's output slice. If a new entry is prepended, existing keys shift — but expand state from old entries is rare enough that losing it on new-entry-prepend is acceptable.

**Alternative**: Key by `entryID + ":" + lineIdx`. Requires entries to carry a stable ID. More correct but adds complexity not justified by the use case.

### 4. JSON summary format

- Object: `{ … }` (N keys) — shows key count
- Array: `[ N items ]` — shows element count
- Parsing happens once at render time per line per expand state

## Risks / Trade-offs

- **Renderer caching**: Caching `allLines` means the model grows by the full rendered line slice. For a feed with 200 entries × 10 output lines × typical JSON expansion this stays well under 1 MB. Acceptable.
- **Key drift on prepend**: If `m.feed` is capped at 200 entries and new entries are prepended, `entryIdx` for existing entries shifts by 1. JSON expand state becomes stale. Mitigation: clear `feedJSONExpanded` when new entries are prepended (same place `feedScrollOffset` and `feedCursor` are reset).
- **Parser order sensitivity**: If two parsers both match a line, only the first wins. Document the registry order. JSON parser is registered first; future parsers append.

## Migration Plan

1. Add `feed_parsers.go` with parser type and JSON parser
2. Refactor `viewActivityFeed` to: (a) run parsers per output line, (b) set `m.feedCachedLines` after building `allLines`
3. Replace `totalFeedLines(m.feed, m.feedPanelWidth())` calls with `len(m.feedCachedLines)` (with fallback to 1 when cache is empty)
4. Update `feedToggleMark` to read from `m.feedCachedLines`
5. Add `enter` key handler in feed context to toggle JSON expand state
6. Update hint bar to show `enter` hint when cursor is on expandable line

No database changes. No config changes. Rollback: revert the four files changed.

## Open Questions

- Should step output lines (inside `entry.steps`) also run through the parser pipeline? **Tentative yes** — step output is where JSON API responses most often appear.
- Cap for expanded JSON: 20 lines enough? If a response is 500 keys, 20 lines + `… N more` is fine for the feed context.
