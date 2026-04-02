## ADDED Requirements

### Requirement: Signals block schema
A sidecar YAML MAY include a top-level `signals:` block containing a list of signal declarations. Each entry SHALL declare:
- `topic` (string, required): BUSD topic or wildcard pattern (e.g. `mud.*`, `weather.updated`)
- `handler` (string, required): named handler to invoke when the topic fires (e.g. `companion`, `score`, `log`)

A sidecar without a `signals:` block is unaffected.

#### Scenario: Valid signals block parsed
- **WHEN** a sidecar declares a `signals:` block with one or more entries
- **THEN** `SidecarSchema.Signals` is a populated slice

#### Scenario: Missing signals block is zero-value safe
- **WHEN** a sidecar omits `signals:`
- **THEN** `SidecarSchema.Signals` is nil and no signal subscriptions are registered

#### Scenario: Unknown handler name logs warning
- **WHEN** a signal declaration references a handler name not in the registry
- **THEN** gl1tch logs a WARN at startup and skips that signal subscription

### Requirement: BUSD subscriptions derived from signals
At startup, gl1tch SHALL collect all `topic` values from loaded signal declarations and add them to the BUSD subscription list alongside existing hardcoded topics. The bus connection SHALL subscribe to all declared topics.

#### Scenario: Plugin topics subscribed at startup
- **WHEN** a sidecar declares `signals: [{topic: "mud.*", handler: companion}]`
- **THEN** the BUSD client subscribes to `mud.*` at connection time

#### Scenario: Multiple plugins with overlapping topics
- **WHEN** two sidecars declare the same topic with different handlers
- **THEN** both handlers are invoked when the topic fires

### Requirement: Signal dispatch to named handler
When a subscribed topic fires, gl1tch SHALL look up the handler name in the `SignalHandlerRegistry` and invoke it with the event topic and payload. If the handler is not found, the event is silently dropped and a DEBUG log is emitted.

#### Scenario: Companion handler invoked on matching event
- **WHEN** `mud.room.entered` fires and a signal declaration maps `mud.*` to `companion`
- **THEN** the companion Ollama narration goroutine is started with the event payload

#### Scenario: Score handler invoked on matching event
- **WHEN** `claude.token.usage` fires and a signal maps it to `score`
- **THEN** the XP engine receives the token usage data

#### Scenario: Log handler writes to debug log
- **WHEN** any topic fires and a signal maps it to `log`
- **THEN** the event is written to `~/.local/share/glitch/plugin-signals.log`
