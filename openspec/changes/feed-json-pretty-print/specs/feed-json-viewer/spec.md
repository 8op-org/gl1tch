## ADDED Requirements

### Requirement: JSON output lines render collapsed by default
The Activity Feed SHALL detect output lines whose trimmed content is valid JSON (object or array) and render them as a single collapsed summary row instead of word-wrapping the raw text.

- Object summary format: `{ … } (N keys)`
- Array summary format: `[ N items ]`
- An expand indicator (e.g. `▸`) SHALL precede the summary when collapsed.
- Detection uses `json.Valid` after trimming; lines that fail validation render as plain text.

#### Scenario: Object JSON line renders collapsed
- **WHEN** a feed entry output line is `{"name":"foo","count":3}`
- **THEN** the Activity Feed shows a single row: `▸ { … } (2 keys)`

#### Scenario: Array JSON line renders collapsed
- **WHEN** a feed entry output line is `[1,2,3,4,5]`
- **THEN** the Activity Feed shows a single row: `▸ [ 5 items ]`

#### Scenario: Non-JSON line renders as plain text
- **WHEN** a feed entry output line is `hello world`
- **THEN** the line renders as plain dimmed text (existing behavior)

### Requirement: Enter key toggles JSON expand/collapse
When the feed cursor is on a collapsed JSON summary row, pressing `enter` SHALL expand it to a pretty-printed view. Pressing `enter` again SHALL collapse it back to the summary.

#### Scenario: Expand collapsed JSON
- **WHEN** the cursor is on a collapsed JSON row and the user presses `enter`
- **THEN** the row expands to a pretty-printed, syntax-highlighted multi-line view (2-space indent)

#### Scenario: Collapse expanded JSON
- **WHEN** the cursor is on an already-expanded JSON row and the user presses `enter`
- **THEN** the view collapses back to the single summary row

#### Scenario: Enter on non-JSON line is a no-op in feed
- **WHEN** the cursor is on a plain text output line and the user presses `enter`
- **THEN** nothing happens (no change to feed state)

### Requirement: Expanded JSON is capped at 20 visible lines
When a JSON value expands to more than 20 pretty-printed lines, the Activity Feed SHALL show the first 20 lines followed by a `… N more lines` overflow indicator.

#### Scenario: Large JSON truncated at 20 lines
- **WHEN** a JSON object pretty-prints to 35 lines and is expanded
- **THEN** 20 lines are shown followed by `… 15 more lines`

### Requirement: Hint bar shows enter hint on JSON cursor line
When the feed is focused and the cursor rests on a JSON summary or expanded-JSON line, the hint bar SHALL include `enter` with description `expand`/`collapse` as appropriate.

#### Scenario: Hint shown on collapsed JSON line
- **WHEN** the feed cursor is on a collapsed JSON row
- **THEN** the hint bar shows `enter  expand`

#### Scenario: Hint shown on expanded JSON line
- **WHEN** the feed cursor is on an expanded JSON first-line row
- **THEN** the hint bar shows `enter  collapse`

### Requirement: JSON expand state resets on new feed entries
When new feed entries are prepended to the Activity Feed, the JSON expand state (`feedJSONExpanded`) SHALL be cleared to avoid stale index references.

#### Scenario: Expand state cleared on new entry
- **WHEN** a new feed entry is prepended while a JSON row is expanded
- **THEN** the previously expanded row returns to collapsed state
