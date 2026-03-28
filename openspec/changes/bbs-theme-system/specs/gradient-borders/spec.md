## ADDED Requirements

### Requirement: Gradient border color stops in theme YAML
The `theme.yaml` schema SHALL accept an optional `gradient_border` field containing 2–4 hex color stops. When present, the panel border renderer SHALL interpolate foreground colors across each border side.

#### Scenario: Two-stop gradient on top border
- **WHEN** `gradient_border: ["#bd93f9", "#ff79c6"]` is declared in `theme.yaml`
- **THEN** the top border SHALL transition from purple on the left to pink on the right via per-character color interpolation

#### Scenario: Four-stop gradient applied to all sides
- **WHEN** four color stops are declared
- **THEN** each border side (top, right, bottom, left) SHALL receive its own linear interpolation across the four stops

#### Scenario: Single-color gradient_border is treated as solid
- **WHEN** only one hex color is provided in `gradient_border`
- **THEN** the border SHALL render as a solid color using that hex value (equivalent to current `BorderForeground` behavior)

#### Scenario: Omitting gradient_border uses current solid border
- **WHEN** `gradient_border` is absent from `theme.yaml`
- **THEN** border rendering SHALL use the existing `palette.accent` solid border color unchanged

### Requirement: Gradient border renderer produces terminal-safe output
The gradient renderer SHALL produce output compatible with tmux (screen-256color), iTerm2, and Kitty by using 24-bit true-color escape sequences (`\x1b[38;2;R;G;Bm`) with a reset after each border line.

#### Scenario: Each border character has an explicit color reset
- **WHEN** a gradient border is rendered
- **THEN** every color transition SHALL be preceded by an explicit `\x1b[0m` reset to prevent color bleed into adjacent content

#### Scenario: Gradient renders without color bleed under tmux
- **WHEN** the terminal reports `TERM=screen-256color` (tmux)
- **THEN** the gradient border SHALL use nearest 256-color approximation (`\x1b[38;5;Nm`) instead of 24-bit sequences

### Requirement: Gradient interpolation uses linear RGB blending
Color stops SHALL be interpolated using linear RGB channel interpolation (not HSL or perceptual). The number of steps equals the border side length in characters.

#### Scenario: Mid-point color matches linear blend
- **WHEN** gradient stops are `#000000` (black) and `#ffffff` (white) on a 10-char border
- **THEN** character 5 SHALL be colored approximately `#7f7f7f` (mid-grey)

### Requirement: Panel-level gradient override
Individual panels SHALL be able to override the theme's `gradient_border` via `header_style.panels.<panel>.gradient_border` in `theme.yaml`.

#### Scenario: Panel-specific gradient overrides theme gradient
- **WHEN** `header_style.panels.pipelines.gradient_border: ["#ff5555", "#50fa7b"]` is set
- **THEN** the pipelines panel SHALL display a red-to-green gradient border regardless of the theme-level gradient
