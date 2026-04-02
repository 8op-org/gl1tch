## ADDED Requirements

### Requirement: ICE class emitted as BUSD signal from runner
When Ollama evaluation returns a non-empty `ICEClass`, the pipeline runner SHALL emit a `game.ice.encountered` BUSD signal with fields: `ice_class` (string), `run_id` (string), `pipeline_name` (string).

#### Scenario: ICE signal emitted on non-nil class
- **WHEN** `EvaluateResult.ICEClass` is `"black-ice"` after a pipeline run
- **THEN** `game.ice.encountered` SHALL be published with `ice_class: "black-ice"` before the run scoring completes

#### Scenario: No signal when ICE class is empty
- **WHEN** `EvaluateResult.ICEClass` is `""` or the evaluate call returns no class
- **THEN** no `game.ice.encountered` signal SHALL be emitted

### Requirement: ICE badge displayed in score card output
The score signal handler SHALL render an ICE badge line when the scored payload includes a non-empty `ice_class` field.

#### Scenario: ICE badge in score output
- **WHEN** `game.run.scored` includes `ice_class: "trace-ice"`
- **THEN** the rendered score card SHALL contain a visually distinct ICE badge line (e.g., `[ICE DETECTED] trace-ice`)

#### Scenario: No badge when ice_class absent
- **WHEN** `game.run.scored` has no `ice_class` field or empty string
- **THEN** no ICE badge line SHALL appear in the score card
