## ADDED Requirements

### Requirement: SignalHandlerRegistry type
A `SignalHandlerRegistry` SHALL be a map from handler name (string) to a handler function of signature `func(topic, payload string)`. The registry is populated at startup with the built-in named handlers. Plugins reference handlers by name in their `signals:` block; they do not supply their own handler code.

Built-in handlers for v1:
- `companion` — starts the Ollama narration goroutine with the event topic and payload as context
- `score` — forwards the payload to the XP/token score engine
- `log` — appends a log line to `~/.local/share/glitch/plugin-signals.log`

#### Scenario: Known handler dispatched on topic fire
- **WHEN** a subscribed BUSD topic fires and the signal declaration names `companion`
- **THEN** `SignalHandlerRegistry["companion"]` is invoked with the topic and payload

#### Scenario: Unknown handler logs warn, drops event
- **WHEN** a signal declaration names a handler not in the registry (e.g. `florp`)
- **THEN** a WARN is logged at startup when loading the sidecar; the event is dropped at dispatch time with a DEBUG log

#### Scenario: Multiple handlers for same topic
- **WHEN** two signal declarations reference the same topic but different handlers (`companion` and `log`)
- **THEN** both handlers are invoked when the topic fires

### Requirement: companion handler
The `companion` handler SHALL read the firing topic and JSON payload and start an Ollama narration goroutine using `game.GameEngine.Respond()` with a plugin-appropriate system prompt. The narration output SHALL be sent to the narration channel for display in the chat panel.

#### Scenario: companion handler narrates mud event
- **WHEN** `mud.room.entered` fires with a JSON payload containing room name
- **THEN** the companion generates a 2-4 line cynical reaction and it appears in the chat panel as a narration message

### Requirement: score handler
The `score` handler SHALL parse the payload as a token usage event and forward it to the XP engine (existing `scoring` package). If the payload cannot be parsed, it is dropped with a DEBUG log.

#### Scenario: score handler updates XP on token usage
- **WHEN** `claude.token.usage` fires with `{"input": 120, "output": 80}`
- **THEN** the XP engine receives the token counts and updates the player score

### Requirement: log handler
The `log` handler SHALL append a single line to `~/.local/share/glitch/plugin-signals.log` in the format: `<RFC3339 timestamp> <topic> <payload>`. The file is created if it does not exist. Write errors are logged at WARN but do not crash gl1tch.

#### Scenario: log handler writes line
- **WHEN** any topic fires and a signal maps it to `log`
- **THEN** a new line appears in `~/.local/share/glitch/plugin-signals.log`
