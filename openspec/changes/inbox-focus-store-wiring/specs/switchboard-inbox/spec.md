## ADDED Requirements

### Requirement: switchboard retains store reference
`switchboard.Model` MUST hold a `*store.Store` field populated by `NewWithStore`.

#### Scenario: store available
- **WHEN** `NewWithStore(s)` called with a non-nil store
- **THEN** `m.store == s` after construction

#### Scenario: store nil (no-op)
- **WHEN** `NewWithStore(nil)` called
- **THEN** switchboard operates normally; inbox shows empty state
