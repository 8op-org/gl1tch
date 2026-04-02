## ADDED Requirements

### Requirement: Quest events emitted as individual BUSD signals
For each entry in `EvaluateResult.QuestEvents`, the pipeline runner SHALL emit one `game.quest.event` BUSD signal with fields: `event_type` (string), `payload` (map[string]any), `run_id` (string).

#### Scenario: Quest event published per entry
- **WHEN** `EvaluateResult.QuestEvents` contains two entries after a pipeline run
- **THEN** two separate `game.quest.event` signals SHALL be published, one per entry

#### Scenario: No signals when quest events empty
- **WHEN** `EvaluateResult.QuestEvents` is nil or empty
- **THEN** no `game.quest.event` signals SHALL be emitted

### Requirement: Quest event logged by default handler
The existing log signal handler SHALL subscribe to `game.quest.event` and append each event to the plugin signals log file.

#### Scenario: Quest event appears in log
- **WHEN** `game.quest.event` is received
- **THEN** the event SHALL be appended to `~/.local/share/glitch/plugin-signals.log` in the same format as other logged events
