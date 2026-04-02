## ADDED Requirements

### Requirement: Pending ICE encounter persisted in SQLite
When `game.ice.encountered` is received, the signal handler SHALL insert a row into the `ice_encounters` table with fields: `id`, `ice_class`, `run_id`, `deadline` (current time + configurable timeout, default 24h), `resolved` (bool, default false), `outcome` (TEXT, default NULL).

The timeout SHALL be configurable in the pack YAML under `ice_encounter.timeout_hours` (integer, default 24). A shorter value (e.g., 1) is appropriate when the MUD plugin is active and the player is expected to be at the terminal. A longer value (e.g., 24) is appropriate for unattended or CI pipeline runs where the player may not be present.

#### Scenario: Encounter row created on signal
- **WHEN** `game.ice.encountered` arrives with `ice_class: "black-ice"` and `run_id: "abc"`
- **THEN** a row SHALL be inserted with `resolved: false` and `deadline` set to current time plus the configured `timeout_hours`

#### Scenario: Default timeout is 24 hours
- **WHEN** no `ice_encounter.timeout_hours` is set in the pack
- **THEN** the deadline SHALL be 24 hours from the time of the signal

### Requirement: `glitch game ice` presents encounter and accepts resolution
`glitch game ice` SHALL read the most recent unresolved encounter from `ice_encounters`. It SHALL display the ICE class and present two options: fight or jack-out. The user's choice SHALL be recorded as `outcome` and the row marked `resolved: true`.

#### Scenario: Fight resolves encounter with win
- **WHEN** user runs `glitch game ice` and selects "fight" with an active encounter
- **THEN** `outcome: "win"` is recorded, `resolved: true`, and no streak penalty is applied

#### Scenario: Jack-out resolves encounter with loss
- **WHEN** user runs `glitch game ice` and selects "jack-out" with an active encounter
- **THEN** `outcome: "loss"` is recorded, `resolved: true`, and streak is decremented by 1 (minimum 0)

#### Scenario: No active encounter exits cleanly
- **WHEN** user runs `glitch game ice` and no unresolved encounter exists
- **THEN** a message "No active ICE encounter." is printed and the command exits 0

### Requirement: Timed-out encounters auto-resolve as losses
On each `glitch` startup, the store SHALL check for unresolved encounters whose `deadline` has passed and mark them as `outcome: "loss"`, `resolved: true`, applying the streak penalty.

#### Scenario: Expired encounter auto-resolves on startup
- **WHEN** an encounter exists with `deadline` in the past and `resolved: false`
- **THEN** on next startup it SHALL be marked `resolved: true`, `outcome: "loss"`, and streak decremented

### Requirement: Streak decrement is bounded at zero
Streak decrement from ICE loss SHALL never produce a negative streak value.

#### Scenario: Streak at zero remains zero after loss
- **WHEN** current streak is 0 and an ICE encounter loss is applied
- **THEN** streak remains 0
