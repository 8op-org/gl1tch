## ADDED Requirements

### Requirement: Orcai starts a local Unix socket event bus at startup
Orcai SHALL start a Unix socket server before launching any widget binaries. The socket path SHALL be `$XDG_RUNTIME_DIR/orcai/bus.sock` when `$XDG_RUNTIME_DIR` is set, falling back to `~/.cache/orcai/bus.sock`.

#### Scenario: Socket is available before widgets launch
- **WHEN** orcai initializes
- **THEN** the bus socket is listening before any widget binary process is started

#### Scenario: Fallback socket path is used when XDG_RUNTIME_DIR is unset
- **WHEN** `$XDG_RUNTIME_DIR` is not set in the environment
- **THEN** orcai creates and listens on `~/.cache/orcai/bus.sock`

### Requirement: Widgets register subscriptions on connect
Upon connecting to the bus socket, a widget SHALL send a registration frame declaring its name and the event types it wants to receive. Orcai SHALL only deliver events matching the declared subscriptions.

#### Scenario: Widget receives only subscribed events
- **WHEN** a widget registers with `subscribe: ["session.started"]`
- **THEN** it receives `session.started` events but not `theme.changed` events

#### Scenario: Widget with empty subscribe list receives no events
- **WHEN** a widget connects and declares no subscriptions
- **THEN** no events are delivered to it; it may still send commands

### Requirement: Core orcai components publish events to the bus
The session manager, theme switcher, and pipeline runner SHALL publish structured events to the internal bus. The bus daemon SHALL fan these out to all subscribed widget connections.

#### Scenario: Session start event published and delivered
- **WHEN** orcai launches a new AI provider session
- **THEN** a `session.started` event with provider, model, and window index is delivered to all subscribed widgets

#### Scenario: Theme change event published and delivered
- **WHEN** the active theme is switched
- **THEN** a `theme.changed` event with the full resolved palette is delivered to all subscribed widgets

### Requirement: Bus daemon shuts down cleanly with orcai
When orcai exits, the bus daemon SHALL close all widget connections, remove the socket file, and terminate cleanly.

#### Scenario: Socket file cleaned up on exit
- **WHEN** orcai shuts down normally
- **THEN** the bus socket file is removed from the filesystem
