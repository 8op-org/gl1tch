## ADDED Requirements

### Requirement: Post-score hook fires tuner asynchronously
After a run is scored (XP computed, user_score updated, achievements recorded), the scoring code SHALL check tuner trigger conditions and, if met, launch the tuner in a background goroutine. The hook SHALL NOT block the scoring path. The hook SHALL pass a detached context (not the run's context) so tuner cancellation does not affect the run lifecycle.

#### Scenario: Tuner fires without blocking scoring
- **WHEN** scoring completes and trigger conditions are met
- **THEN** the tuner goroutine is launched and the scoring function returns immediately

#### Scenario: Tuner uses detached context
- **WHEN** the run context is cancelled after scoring
- **THEN** the tuner goroutine continues running to completion unaffected
