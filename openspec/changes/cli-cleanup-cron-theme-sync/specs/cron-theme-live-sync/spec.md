## ADDED Requirements

### Requirement: Cron TUI detects active theme changes on disk without restart
The cron TUI SHALL poll the on-disk active theme at a regular interval (≤ 5 seconds). When the active theme name on disk differs from the currently rendered theme, the cron TUI SHALL reload the theme bundle and re-render with the new palette.

#### Scenario: Theme change propagates to cron TUI within 5 seconds
- **WHEN** the user switches the active theme in the switchboard TUI
- **THEN** the cron TUI re-renders with the new theme within 5 seconds, without requiring a restart

#### Scenario: Cron TUI retains theme on unchanged disk
- **WHEN** the active theme on disk has not changed since the last poll
- **THEN** the cron TUI does not re-render or reload the theme bundle

#### Scenario: Cron TUI falls back to Dracula on disk-read failure
- **WHEN** the active theme file cannot be read during a poll cycle
- **THEN** the cron TUI continues rendering with the last successfully loaded theme and logs a warning
