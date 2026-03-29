## ADDED Requirements

### Requirement: Feedline parser pipeline processes output lines before rendering
The Activity Feed renderer SHALL pass each raw output string through a registered pipeline of `FeedLineParser` functions before appending display lines. The first parser that matches a line SHALL handle it; if no parser matches, the line renders as plain wrapped text.

A `FeedLineParser` is a function with signature:
```go
type FeedLineParser func(raw string, width int, expanded bool) (lines []string, matched bool)
```

- `raw`: the trimmed source line
- `width`: available display width
- `expanded`: whether this line's expand state is currently true
- Returns `lines`: the display lines to append, and `matched`: whether this parser handled the line

#### Scenario: Matched parser emits display lines
- **WHEN** a parser returns `matched: true` for a raw line
- **THEN** the renderer appends the returned display lines and skips remaining parsers

#### Scenario: No parser matches falls through to default
- **WHEN** all parsers return `matched: false`
- **THEN** the renderer applies default word-wrap and dim styling

### Requirement: Parser registry is a package-level ordered slice
Parsers SHALL be registered in a package-level slice in `feed_parsers.go`. Parsers are evaluated in registration order; the JSON parser is registered first.

#### Scenario: JSON parser is first in registry
- **WHEN** a line is valid JSON
- **THEN** the JSON parser handles it before any other parser is consulted

### Requirement: Parser pipeline applies to both entry output and step output lines
The parser pipeline SHALL run for both `entry.lines` (direct entry output) and `step.lines` (per-step output lines).

#### Scenario: Step output JSON is also collapsed
- **WHEN** a pipeline step output line is valid JSON
- **THEN** it renders as a collapsed JSON summary, not raw word-wrapped text

### Requirement: feedRawLines and totalFeedLines are derived from renderer output
The canonical line count and raw line content SHALL be derived from the same slice (`allLines`) that `viewActivityFeed` builds. `totalFeedLines` is replaced by `len(m.feedCachedLines)`. `feedToggleMark` reads from `m.feedCachedLines`.

This eliminates width-dependent mark-index drift: marks always key on the renderer's own output indices.

#### Scenario: Mark index matches rendered line
- **WHEN** the user marks line 5 in the Activity Feed
- **THEN** the mark highlight appears on the same row that was visually at position 5 in the last render

#### Scenario: No index drift after window resize
- **WHEN** the terminal is resized after marking line 7
- **THEN** the mark indicator remains on the same semantic content line (cursor and cache are reset on resize)
