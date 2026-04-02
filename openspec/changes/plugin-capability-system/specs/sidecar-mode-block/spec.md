## ADDED Requirements

### Requirement: Mode block schema
A sidecar YAML MAY include a top-level `mode:` block. When present it SHALL declare:
- `trigger` (string, required): slash command that activates this widget (e.g. `/mud`)
- `logo` (string, required): text rendered via TDF font when the mode is active
- `speaker` (string, required): label shown in the chat panel for plugin output (max 6 chars)
- `exit_command` (string, required): user input that deactivates the mode (e.g. `quit`)
- `on_activate` (string, optional): command piped to the plugin binary on first launch

A sidecar without a `mode:` block SHALL behave identically to the current behaviour.

#### Scenario: Valid mode block parsed
- **WHEN** a sidecar declares a complete `mode:` block
- **THEN** `SidecarSchema.Mode` is populated with all declared fields

#### Scenario: Missing mode block is zero-value safe
- **WHEN** a sidecar omits the `mode:` block
- **THEN** `SidecarSchema.Mode` is a zero-value struct and no widget behaviour is registered

#### Scenario: Mode block missing required field is rejected
- **WHEN** a sidecar `mode:` block omits `trigger`
- **THEN** `LoadWidgetRegistry` logs a WARN and skips that sidecar's widget registration

### Requirement: Slash command activation
When a `mode:` block is loaded, gl1tch SHALL register the declared `trigger` as a dynamic slash command. Typing the trigger in the gl1tch chat panel SHALL:
1. Set the panel into widget mode for this plugin
2. Swap the TDF header to render `logo`
3. If `on_activate` is set, pipe it to the plugin binary and display the output as a `speaker`-labelled message
4. Route all subsequent non-slash user input to the plugin binary (one-shot per command)

#### Scenario: Trigger activates widget mode
- **WHEN** the user types `/mud` and a sidecar declares `trigger: /mud`
- **THEN** gl1tch enters widget mode, swaps the logo, and runs `on_activate` if set

#### Scenario: Trigger already in mode shows hint
- **WHEN** the user types the trigger while already in widget mode
- **THEN** gl1tch shows a bot message indicating the mode is already active

### Requirement: Exit command deactivates mode
When in widget mode, if the user submits the value matching `exit_command`, gl1tch SHALL:
1. Exit widget mode
2. Restore the original TDF header (`GL1TCH`)
3. Show a bot message confirming deactivation

#### Scenario: Exit command restores normal mode
- **WHEN** user types `quit` while in widget mode with `exit_command: quit`
- **THEN** gl1tch exits widget mode and restores the GL1TCH logo

### Requirement: Widget input routing
While in widget mode, non-slash user input SHALL be piped to the plugin binary on stdin. The binary's stdout SHALL be captured and displayed as a chat message labelled with `speaker`. Stderr is merged with stdout. The binary is invoked once per command (one-shot, not a persistent process).

#### Scenario: Command routed to plugin binary
- **WHEN** user types `go north` while in widget mode
- **THEN** gl1tch runs the plugin binary with `go north` on stdin and displays the output

#### Scenario: Binary not found shows install hint
- **WHEN** the plugin binary is not on PATH
- **THEN** gl1tch displays a bot message with the install command from the sidecar description
