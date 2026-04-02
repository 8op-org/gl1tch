## ADDED Requirements

### Requirement: WidgetRegistry type
A `WidgetRegistry` SHALL be a struct that holds a slice of loaded `WidgetConfig` entries. Each `WidgetConfig` wraps the `SidecarSchema.Mode` block together with the sidecar's binary path and speaker label so dispatch code has everything it needs without re-reading disk.

#### Scenario: Registry built from wrappers dir
- **WHEN** `LoadWidgetRegistry(wrappersDir string)` is called at startup
- **THEN** every sidecar YAML with a non-zero `mode:` block is loaded into the registry

#### Scenario: Sidecar without mode block is ignored
- **WHEN** a sidecar YAML has no `mode:` block
- **THEN** it is not added to the `WidgetRegistry`

#### Scenario: Sidecar with invalid mode block logs warning
- **WHEN** a sidecar `mode:` block is missing a required field (e.g. no `trigger`)
- **THEN** `LoadWidgetRegistry` logs a WARN and skips that sidecar

#### Scenario: Duplicate trigger logs warning, first loaded wins
- **WHEN** two sidecars declare the same `trigger` value
- **THEN** the first loaded entry is kept, the second is skipped with a WARN log

### Requirement: Slash command lookup
`WidgetRegistry.FindByTrigger(trigger string)` SHALL return the `WidgetConfig` whose `Mode.Trigger` matches the given string, or `nil` if no match.

#### Scenario: Exact trigger match returns config
- **WHEN** `FindByTrigger("/mud")` is called and a sidecar declares `trigger: /mud`
- **THEN** the matching `WidgetConfig` is returned

#### Scenario: Unknown trigger returns nil
- **WHEN** `FindByTrigger("/unknown")` is called and no sidecar declares that trigger
- **THEN** `nil` is returned

### Requirement: BUSD topic collection
`WidgetRegistry.AllSignalTopics()` SHALL return a deduplicated slice of all `topic` values declared across all loaded sidecars' `signals:` blocks.

#### Scenario: Topics collected from multiple sidecars
- **WHEN** two sidecars declare `signals:` blocks with topics `mud.*` and `weather.updated`
- **THEN** `AllSignalTopics()` returns both

#### Scenario: Duplicate topics deduplicated
- **WHEN** two sidecars declare the same topic
- **THEN** `AllSignalTopics()` returns it once

### Requirement: Registry stored on Model
`switchboard.New()` SHALL call `LoadWidgetRegistry(wrappersDir)` and store the result on the `Model`. The wrappers directory defaults to `~/.config/glitch/wrappers/`. The registry is read-only after startup (no hot-reload).

#### Scenario: Model carries registry after init
- **WHEN** `switchboard.New()` completes
- **THEN** `m.widgetRegistry` is non-nil (may be empty if no widget sidecars are present)
