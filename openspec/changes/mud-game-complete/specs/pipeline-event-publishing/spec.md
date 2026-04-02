## MODIFIED Requirements

### Requirement: Runner publishes pipeline step lifecycle events
The runner SHALL publish to `orcai.pipeline.step.started` when a step begins and to `orcai.pipeline.step.done` or `orcai.pipeline.step.failed` when it ends. The payload SHALL be JSON: `{"pipeline": "<name>", "step": "<id>", "status": "<status>", "duration_ms": <int>}`.

#### Scenario: Step started event published
- **WHEN** the runner begins executing step `build`
- **THEN** an event on `orcai.pipeline.step.started` with `step: "build"` is published

#### Scenario: Step done event published with duration
- **WHEN** step `build` completes successfully
- **THEN** an event on `orcai.pipeline.step.done` with `status: "done"` and `duration_ms` set is published

#### Scenario: Step failed event published
- **WHEN** step `build` fails after all retry attempts
- **THEN** an event on `orcai.pipeline.step.failed` with `status: "failed"` is published

## ADDED Requirements

### Requirement: `game.run.scored` payload includes achievements and ice_class
The `game.run.scored` BUSD signal payload SHALL include an `achievements` field ([]string, never nil — use empty slice) populated from `EvaluateResult.Achievements`, and an `ice_class` field (string, empty string when none) populated from `EvaluateResult.ICEClass`.

#### Scenario: Achievements propagated in scored payload
- **WHEN** `EvaluateResult.Achievements` contains `["speed-demon"]` after a run
- **THEN** the `game.run.scored` payload SHALL include `achievements: ["speed-demon"]`

#### Scenario: Empty achievements never nil
- **WHEN** `EvaluateResult.Achievements` is nil or empty
- **THEN** the `game.run.scored` payload SHALL include `achievements: []` (not null/omitted)

#### Scenario: ICE class propagated in scored payload
- **WHEN** `EvaluateResult.ICEClass` is `"black-ice"`
- **THEN** the `game.run.scored` payload SHALL include `ice_class: "black-ice"`

#### Scenario: Empty ICE class when none detected
- **WHEN** `EvaluateResult.ICEClass` is empty
- **THEN** the `game.run.scored` payload SHALL include `ice_class: ""`
