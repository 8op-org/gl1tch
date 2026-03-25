## ADDED Requirements

### Requirement: Themes are directory bundles with a manifest and optional assets
A theme bundle SHALL be a directory at `~/.config/orcai/themes/<name>/` containing a `theme.yaml` manifest and optional asset files (`.ans` ANSI art). Bundled themes SHALL ship via `//go:embed` and be available without filesystem reads.

#### Scenario: User installs a community theme
- **WHEN** a directory is placed at `~/.config/orcai/themes/gruvbox/` with a valid `theme.yaml`
- **THEN** orcai discovers it as an available theme at next startup

#### Scenario: Bundled ABS theme is available on fresh install
- **WHEN** orcai starts with no user themes installed
- **THEN** the ABS theme is available and active by default

### Requirement: Theme manifest covers palette, borders, status bar, and ANSI art
A valid `theme.yaml` SHALL support: `name`, `display_name`, `palette` (named semantic colors as hex strings: `bg`, `fg`, `accent`, `dim`, `border`, `error`, `success`), `borders.style` (`light`, `heavy`, or `ascii`), `statusbar.format` and color references, and an optional `splash` path pointing to an `.ans` file within the bundle.

#### Scenario: Palette provides all required semantic color names
- **WHEN** a theme is loaded
- **THEN** all of `bg`, `fg`, `accent`, `dim`, `border`, `error`, `success` are available as resolved hex values

#### Scenario: Missing splash asset falls back gracefully
- **WHEN** a theme's `splash` path points to a non-existent file
- **THEN** orcai falls back to the bundled ABS splash and logs a warning

### Requirement: Theme switching broadcasts palette to all connected widgets
When the active theme changes, orcai SHALL publish a `theme.changed` event on the bus with the full resolved palette as a flat map of color names to hex strings.

#### Scenario: Widget re-renders on theme change
- **WHEN** the user switches from ABS to Gruvbox
- **THEN** all connected widgets subscribed to `theme.changed` receive the full Gruvbox palette and can re-render

### Requirement: Active theme is persisted in orcai config
Orcai SHALL store the name of the active theme in its config file. On restart, the previously active theme SHALL be restored.

#### Scenario: Theme persists across restart
- **WHEN** a user sets the active theme to Gruvbox and restarts orcai
- **THEN** Gruvbox is loaded as the active theme on next startup
