## ADDED Requirements

### Requirement: TUI processes subscribe to theme changes via busd
Any BubbleTea TUI running as a standalone OS process SHALL be able to receive `theme.changed` events from the busd daemon using a shared helper, without implementing its own socket dial or JSON framing logic.

#### Scenario: Theme change delivered to standalone sub-TUI
- **WHEN** the user switches themes in switchboard
- **THEN** a standalone crontui process receives the new theme name within 1 second via busd

#### Scenario: Sub-TUI renders updated theme immediately
- **WHEN** a `theme.changed` busd event is received
- **THEN** the sub-TUI re-renders using the new theme bundle without requiring user interaction

### Requirement: busd subscriber cmd handles daemon unavailability gracefully
If the busd daemon is not running when a TUI process attempts to subscribe, the subscription SHALL fail silently — the TUI SHALL continue to function with its last-known theme rather than crashing or blocking.

#### Scenario: busd not running at TUI startup
- **WHEN** a sub-TUI initializes and the busd socket does not exist
- **THEN** the TUI starts successfully and displays its default or last-known theme

#### Scenario: busd connection lost mid-session
- **WHEN** the busd daemon stops unexpectedly while a sub-TUI is running
- **THEN** the sub-TUI continues to function with the theme active at the time of disconnect

### Requirement: switchboard publishes theme.changed to busd on theme selection
When the user selects a new theme in the switchboard theme picker, switchboard SHALL publish a `theme.changed` event to the busd daemon containing the new theme name.

#### Scenario: Theme picker selection triggers busd publish
- **WHEN** the user confirms a new theme in the switchboard theme picker
- **THEN** a `theme.changed` event with the selected theme name is published to busd

#### Scenario: All connected sub-TUIs receive the event
- **WHEN** a `theme.changed` event is published to busd
- **THEN** all currently connected sub-TUI processes that subscribed to `theme.changed` receive the event

### Requirement: Shared box drawing helpers available to all TUIs
The `panelrender` package SHALL export `BoxTop`, `BoxBot`, and `BoxRow` functions so any sub-TUI can draw themed panel borders without re-implementing or copying private switchboard helpers.

#### Scenario: Sub-TUI renders panel borders using shared helpers
- **WHEN** a sub-TUI calls `panelrender.BoxTop` / `BoxBot` / `BoxRow`
- **THEN** the panel border is rendered consistently with the same visual style as switchboard panels

### Requirement: File-poll theme detection removed from crontui
The `crontui` package SHALL NOT use a periodic file-poll to detect cross-process theme changes. The busd subscription is the sole cross-process theme update mechanism.

#### Scenario: Theme change in switchboard reflected in crontui without polling delay
- **WHEN** the user switches themes in switchboard while crontui is open
- **THEN** crontui reflects the new theme within 1 second (not the previous up-to-5-second poll interval)
