## ADDED Requirements

### Requirement: Agent runner providers follow a canonical priority order
`buildProviders()` in `internal/picker/picker.go` SHALL sort discovered sidecar providers according to a static `providerPriority` slice before appending them to the output list. The canonical order SHALL be: `claude`, `copilot`, `codex`, `gemini`, `opencode`, `ollama`, `shell`. Providers not in the priority list SHALL appear after all priority providers, in their original discovery order.

#### Scenario: Claude appears before Copilot
- **WHEN** both `claude` and `copilot` sidecars are installed
- **THEN** `buildProviders()` returns Claude before Copilot in the slice

#### Scenario: Codex appears after Copilot and before Gemini
- **WHEN** `claude`, `copilot`, `codex`, and `gemini` sidecars are all installed
- **THEN** the order is Claude → Copilot → Codex → Gemini

#### Scenario: Unknown providers appended after priority providers
- **WHEN** a sidecar named `my-custom-agent` is installed alongside `claude` and `codex`
- **THEN** `buildProviders()` returns Claude → Codex → my-custom-agent

#### Scenario: Shell always appears last among known providers
- **WHEN** multiple providers are installed including shell
- **THEN** `shell` appears after all AI provider cards in the grid

### Requirement: Priority ordering is deterministic across OS restarts
The agent runner grid SHALL display providers in the same order on every launch, regardless of filesystem iteration order on the host OS.

#### Scenario: Order stable across multiple calls
- **WHEN** `buildProviders()` is called multiple times with the same installed sidecars
- **THEN** the returned slice is identical on each call
