## ADDED Requirements

### Requirement: Widget manifests define external plugin binaries
A widget manifest (`widget.yaml`) in `~/.config/orcai/widgets/<name>/` SHALL define `name`, `binary` (path or name), `description`, and an optional `subscribe` list of event types. Orcai SHALL discover all valid manifests at startup.

#### Scenario: Widget discovered from user config directory
- **WHEN** `~/.config/orcai/widgets/weather/widget.yaml` exists with a valid manifest
- **THEN** orcai registers the weather widget and prepares to launch it

#### Scenario: Malformed manifest is skipped with error log
- **WHEN** a `widget.yaml` is missing the required `binary` field
- **THEN** orcai logs a warning and continues loading other widgets without crashing

### Requirement: Orcai launches widget binaries in tmux panes
Each discovered widget binary SHALL be launched by orcai in a tmux pane at startup. Orcai SHALL NOT dictate the tmux layout; window/pane positioning is the user's or widget's responsibility.

#### Scenario: Widget binary starts in its own pane
- **WHEN** orcai initializes and a widget manifest is valid
- **THEN** the widget binary process is started via tmux and runs independently

### Requirement: Widgets communicate via framed JSON protocol over Unix socket
Orcai SHALL serve a Unix socket at the bus daemon address. Widget binaries MAY connect to this socket to send commands to orcai and receive subscribed events as newline-delimited JSON frames.

#### Scenario: Widget receives a subscribed event
- **WHEN** a session starts and a widget has `session.started` in its `subscribe` list
- **THEN** the widget receives `{"event": "session.started", "payload": {...}}` on its socket connection

#### Scenario: Widget sends a command to orcai
- **WHEN** a widget writes `{"cmd": "notify", "payload": {"text": "Updated"}}` to the socket
- **THEN** orcai processes the command and displays the notification

#### Scenario: Widget that ignores the socket still functions
- **WHEN** a widget binary never connects to the bus socket
- **THEN** orcai does not block or error; the widget renders freely in its pane

### Requirement: Disconnected widget clients are pruned silently
When a widget binary exits or loses its socket connection, orcai SHALL remove it from the subscriber registry without propagating errors to other widgets or crashing.

#### Scenario: Widget crash does not affect orcai or other widgets
- **WHEN** a widget binary exits unexpectedly
- **THEN** orcai removes it from subscribers and continues operating normally

### Requirement: First-party widgets use the same contract as contributor widgets
Built-in widgets (welcome splash, sysop panel) SHALL be implemented as widget binaries with their own `widget.yaml` manifests. They SHALL receive no privileged access beyond what the protocol provides.

#### Scenario: Welcome widget uses bus protocol for theme
- **WHEN** the active theme changes
- **THEN** the welcome widget receives `theme.changed` via the socket and re-renders using the new palette
