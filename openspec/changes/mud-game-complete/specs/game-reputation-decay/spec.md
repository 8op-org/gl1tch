## ADDED Requirements

### Requirement: Reputation decay applied on startup based on inactive days
On each `glitch` startup, the store SHALL compute the number of full days since the last recorded pipeline run. For each day (up to a configurable cap, default 7), it SHALL apply a `decay_per_day` reduction (default 2) to each MUD faction's reputation score, respecting a floor (default 10).

#### Scenario: One day of inactivity decays reputation
- **WHEN** the last pipeline run was exactly 1 full day ago and `decay_per_day` is 2
- **THEN** each faction's reputation SHALL be reduced by 2 (or to the floor, whichever is higher)

#### Scenario: Decay capped at 7 days maximum
- **WHEN** the last pipeline run was 30 days ago
- **THEN** decay is applied as if only 7 days had passed (cap prevents total lockout from long absences)

#### Scenario: Floor prevents reputation going below minimum
- **WHEN** a faction's reputation is 11 and decay would reduce it by 4
- **THEN** reputation is set to 10 (the floor), not 7

#### Scenario: No decay when run occurred today
- **WHEN** a pipeline run was recorded today (same calendar day)
- **THEN** no decay is applied

### Requirement: Pack defines decay parameters
The pack YAML SHALL support a `reputation_decay` block with fields: `decay_per_day` (integer, default 2), `floor` (integer, default 10), `max_decay_days` (integer, default 7).

#### Scenario: Pack decay parameters override defaults
- **WHEN** pack YAML sets `reputation_decay.decay_per_day: 5`
- **THEN** the startup decay check uses 5 points per day instead of 2

### Requirement: Reputation restored on next active run
No explicit restore mechanic is required — natural reputation gain through MUD activity (existing NPC interaction system) is sufficient. Decay simply creates pressure; it does not permanently reduce reputation.

#### Scenario: Decay does not prevent reputation gain
- **WHEN** reputation is at the floor and the player talks to an NPC that grants +10 reputation
- **THEN** reputation increases normally to floor + 10
