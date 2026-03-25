## MODIFIED Requirements

### Requirement: Dashboard is persistent and does not auto-exit
Window 0 SHALL display the live dashboard indefinitely after launch. The dashboard SHALL NOT exit to `$SHELL` on arbitrary keypresses. Only an explicit quit action (`q` or `ctrl+c`) SHALL trigger the shell handoff.

#### Scenario: Dashboard stays open after launch
- **WHEN** orcai starts and window 0 opens
- **THEN** the dashboard remains visible and does not exit until the user presses `q` or `ctrl+c`

#### Scenario: Arbitrary keypresses do not close dashboard
- **WHEN** the user presses any key other than `q`, `ctrl+c`, or `enter`
- **THEN** the dashboard remains open and does not exit to shell

#### Scenario: q quits to shell
- **WHEN** the user presses `q`
- **THEN** the dashboard exits and the process is replaced by `$SHELL`

### Requirement: Dashboard is implemented as a first-party widget binary
The welcome dashboard SHALL be implemented as a widget binary following the widget plugin contract. It SHALL have a `widget.yaml` manifest and connect to the bus daemon to receive `theme.changed` and `session.started`/`session.ended` events. It SHALL NOT be a baked-in BubbleTea component in `internal/welcome/`.

#### Scenario: Welcome widget receives theme event and re-renders
- **WHEN** the active theme changes
- **THEN** the welcome widget receives the `theme.changed` event and re-renders the banner and status cards using the new palette

#### Scenario: Welcome widget manifest is discoverable
- **WHEN** orcai starts
- **THEN** the welcome widget manifest is found and the widget binary is launched in window 0

### Requirement: Dashboard displays banner from active theme
The dashboard banner SHALL use the active theme's ANSI art (`splash.ans`) and palette rather than hardcoded color constants. The banner SHALL scale to terminal width.

#### Scenario: Banner uses theme splash art
- **WHEN** the active theme has a `splash.ans` asset
- **THEN** the dashboard renders that ANSI art as its banner

#### Scenario: Banner scales to terminal width
- **WHEN** the terminal is resized
- **THEN** the banner redraws at the new width without truncation or overflow

### Requirement: Dashboard shows one session card per active tmux window
The dashboard SHALL render one card for each non-home tmux window in the `orcai` session. Each card SHALL display the window name, provider, status indicator (● streaming / ○ idle), input token count, output token count, and estimated cost in USD. Windows with no received telemetry SHALL display "no data" in the metrics area. Cards SHALL use box-drawing borders in the active theme palette.

#### Scenario: Card rendered for each active window
- **WHEN** the orcai session has N non-home windows
- **THEN** the dashboard shows N session cards

#### Scenario: Card shows telemetry data
- **WHEN** a window has received at least one `orcai.telemetry` event
- **THEN** its card shows provider name, status icon, input tokens (rounded to k), output tokens, and cost formatted as `$0.000`

#### Scenario: Card shows no-data placeholder
- **WHEN** a window has not yet received any telemetry
- **THEN** its card shows "no data" in the metrics area

#### Scenario: No sessions shows empty state
- **WHEN** no non-home windows exist
- **THEN** the dashboard shows a "no active sessions" message instead of cards

### Requirement: Status indicator reflects streaming vs idle state
A card's status indicator SHALL update in real time as telemetry events arrive. A `status: "streaming"` event SHALL show `●` in the theme's success color. A `status: "done"` event SHALL show `○` in the theme's dim color.

#### Scenario: Streaming status shown in success color
- **WHEN** a `orcai.telemetry` event with `status: "streaming"` is received
- **THEN** the corresponding card displays `●` in the active theme's `success` color

#### Scenario: Idle status shown dimmed
- **WHEN** a `orcai.telemetry` event with `status: "done"` is received
- **THEN** the corresponding card displays `○` in the active theme's `dim` color

### Requirement: Dashboard shows aggregate totals row
Below all session cards the dashboard SHALL render a single totals row summing input tokens, output tokens, and cost across all sessions that have telemetry data. The row SHALL be clearly labelled "TOTAL".

#### Scenario: Totals row sums all sessions
- **WHEN** multiple sessions have telemetry data
- **THEN** the totals row shows the sum of all input tokens, output tokens, and cost

#### Scenario: Totals row shows zero when no data
- **WHEN** no sessions have telemetry data
- **THEN** the totals row shows 0 for all fields

### Requirement: Dashboard subscribes to orcai bus events
The dashboard widget SHALL connect to the bus daemon socket on startup and subscribe to `orcai.telemetry`, `session.started`, `session.ended`, and `theme.changed` events. Each received event SHALL update the corresponding session card or re-render the theme immediately. The dashboard SHALL retry connection for up to 3 seconds before proceeding without telemetry.

#### Scenario: Bus connection established on startup
- **WHEN** the dashboard widget starts and the bus socket is available
- **THEN** the dashboard connects to the bus and begins receiving events

#### Scenario: Bus unavailable — dashboard still renders
- **WHEN** the bus socket is not available after the 3-second retry window
- **THEN** the dashboard renders with "no data" for all sessions and does not crash

#### Scenario: Telemetry event updates card in real time
- **WHEN** an `orcai.telemetry` event is received for a window
- **THEN** that window's card updates without requiring a manual refresh

### Requirement: Window list refreshes periodically
The dashboard SHALL refresh the list of active tmux windows every 5 seconds via a tick command. New windows SHALL appear as cards; killed windows SHALL be removed.

#### Scenario: New window appears after tick
- **WHEN** a new tmux window is created and 5 seconds elapse
- **THEN** a new card appears in the dashboard for that window

#### Scenario: Killed window removed after tick
- **WHEN** a tmux window is killed and 5 seconds elapse
- **THEN** its card is removed from the dashboard

### Requirement: Enter opens the provider picker popup
Pressing `Enter` on the dashboard SHALL open the provider/model picker in a tmux display-popup, identical to the current welcome screen behaviour.

#### Scenario: Enter opens picker
- **WHEN** the user presses `Enter`
- **THEN** a tmux display-popup opens running `orcai _picker`

### Requirement: Footer shows chord-key hints
The dashboard footer SHALL display `^spc n new · ^spc p build` hints in the active theme's dim color, consistent with the status-bar hints.

#### Scenario: Footer shows navigation hints
- **WHEN** the dashboard is rendered
- **THEN** the footer contains `^spc n new` and `^spc p build`
