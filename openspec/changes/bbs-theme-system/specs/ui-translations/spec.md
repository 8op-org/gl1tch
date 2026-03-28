## ADDED Requirements

### Requirement: Translations YAML loader
The system SHALL load `~/.config/orcai/translations.yaml` at startup and expose a `translations.Provider` interface with a single method `T(key, fallback string) string`.

#### Scenario: Known key returns translated value
- **WHEN** `translations.yaml` contains `pipelines_panel_title: "▓▒░ PIPELINE OPS ░▒▓"`
- **THEN** `provider.T("pipelines_panel_title", "Pipelines")` SHALL return `"▓▒░ PIPELINE OPS ░▒▓"`

#### Scenario: Unknown key returns fallback
- **WHEN** `translations.yaml` does not contain a key
- **THEN** `provider.T(key, fallback)` SHALL return the fallback string unchanged

#### Scenario: Missing translations file is a no-op
- **WHEN** `~/.config/orcai/translations.yaml` does not exist
- **THEN** the provider SHALL return the fallback for every key with no error logged to the user

### Requirement: ANSI escape sequence expansion in translation values
Translation values SHALL support `\e[...m` and `\033[...m` escape notation, expanded to raw ANSI bytes at load time.

#### Scenario: Color escape in translation value
- **WHEN** `translations.yaml` contains `switchboard_title: "\e[38;2;189;147;249mORCAI\e[0m"`
- **THEN** `provider.T("switchboard_title", "ORCAI")` SHALL return the value with raw ESC bytes (0x1B) in place of `\e`

#### Scenario: Malformed escape sequence rendered safely
- **WHEN** a translation value contains an incomplete escape (e.g., `"\e["` with no closing `m`)
- **THEN** the system SHALL apply `\x1b[0m` (reset) after the value when rendering to prevent terminal state corruption

### Requirement: Translatable keys cover all major UI labels
The following keys SHALL be translatable: panel names (`pipelines_panel_title`, `agent_runner_panel_title`, `signal_board_panel_title`, `activity_feed_panel_title`, `inbox_panel_title`, `cron_panel_title`), status bar tokens, modal titles (`quit_modal_title`, `help_modal_title`, `theme_picker_title`), and the main header title (`switchboard_header_title`).

#### Scenario: Custom panel name appears in panel header
- **WHEN** `cron_panel_title` is set to `"⚡ CRON DAEMON ⚡"`
- **THEN** the cron panel header SHALL display `"⚡ CRON DAEMON ⚡"` as its title text

#### Scenario: Custom modal title appears in quit confirm
- **WHEN** `quit_modal_title` is set to `">>> BAIL OUT <<<"`
- **THEN** the quit confirm modal title bar SHALL display `">>> BAIL OUT <<<"`

### Requirement: Translations persist across theme switches
The translations provider SHALL continue returning the same values when the active theme changes; translations are independent of theme.

#### Scenario: Theme change does not reset translations
- **WHEN** the user switches from Dracula to Nord
- **THEN** all translated labels SHALL remain in effect after the theme reload

### Requirement: Plugin context receives translations provider
The plugin host SHALL inject the `translations.Provider` into the plugin context so plugins can look up user-configured labels.

#### Scenario: Plugin reads custom label
- **WHEN** a plugin calls `ctx.Translations().T("my_plugin_label", "Default Label")`
- **THEN** it SHALL receive the user-configured value if present, or "Default Label" otherwise
