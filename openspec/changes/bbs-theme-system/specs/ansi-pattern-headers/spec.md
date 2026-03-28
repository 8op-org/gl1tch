## ADDED Requirements

### Requirement: CP437 pattern library with named patterns
The system SHALL provide a pattern library of at least 8 named patterns using Unicode equivalents of CP437 characters (128–255 range): `waves`, `texture`, `checkerboard`, `double-line`, `shadow`, `noise`, `brick`, `scanline`.

#### Scenario: Pattern tiles to exact panel width
- **WHEN** a panel header is rendered with pattern `waves` at width W
- **THEN** the pattern row SHALL be exactly W visible characters wide, tiling or truncating the pattern sequence to fit

#### Scenario: Pattern mirrors on bottom row
- **WHEN** a panel header renders a top and bottom pattern row
- **THEN** the bottom row SHALL use the mirror/inverse variant of the same pattern (e.g., upper half-block ▀ on top → lower half-block ▄ on bottom)

#### Scenario: Unknown pattern name falls back to block chars
- **WHEN** a theme declares `header_pattern: "unknown-pattern"`
- **THEN** the header engine SHALL fall back to the existing `▄`/`▀`/`█` block character rendering

### Requirement: TDF font renderer for panel title text
The system SHALL implement a TDF (TheDraw Font) parser in `internal/tdf` that reads `.tdf` binary files and renders a string as ANSI block-letter art.

#### Scenario: Render panel title with TDF font
- **WHEN** a theme declares `header_font: "standard"` and a panel header is rendered
- **THEN** the title text SHALL be rendered as ANSI block letters using the named TDF font, centered within the header bar

#### Scenario: TDF render width clamps to panel width
- **WHEN** the TDF-rendered title is wider than the panel
- **THEN** the renderer SHALL fall back to plain centered text for that panel rather than truncating mid-character

#### Scenario: Missing TDF font falls back to plain text
- **WHEN** the named font file is not found in `internal/assets/fonts/tdf/`
- **THEN** the header SHALL render the plain text title centered in the accent bar (no error, no panic)

#### Scenario: Empty header_font disables TDF rendering
- **WHEN** `header_font` is omitted or empty in `theme.yaml`
- **THEN** the header SHALL render using CP437 pattern decoration with plain centered text

### Requirement: Header title centered full-width
The panel header bar SHALL span exactly 100% of the terminal/panel width and the title text SHALL be horizontally centered within that bar.

#### Scenario: Title centered at varying terminal widths
- **WHEN** the terminal is resized to any width
- **THEN** the next render of the panel header SHALL recompute title centering at the new width with no off-by-one misalignment

#### Scenario: Header bars align horizontally across panels
- **WHEN** multiple panels are rendered side by side
- **THEN** the top edge of each panel header SHALL form a continuous horizontal line at the same terminal row

### Requirement: Theme YAML declares header pattern and font
The `theme.yaml` schema SHALL accept optional `header_pattern` (string, pattern name) and `header_font` (string, TDF font name) fields at the theme root level.

#### Scenario: Theme with header_pattern and header_font
- **WHEN** a `theme.yaml` declares both `header_pattern: "waves"` and `header_font: "fire"`
- **THEN** all panels SHALL render the waves pattern border with the fire TDF font for titles

#### Scenario: Theme without new fields uses existing behavior
- **WHEN** a `theme.yaml` omits `header_pattern` and `header_font`
- **THEN** panel headers SHALL render using the existing `▄`/`▀`/`█` block character logic unchanged
