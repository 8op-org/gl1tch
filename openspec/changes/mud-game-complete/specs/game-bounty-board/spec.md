## ADDED Requirements

### Requirement: Tuner generates bounty contracts in evolved pack
The tuner's Ollama analysis prompt SHALL request 3–5 bounty contracts. The generate prompt SHALL include them in the evolved pack YAML under a `bounty_contracts` list. Each contract SHALL have fields: `id` (string), `description` (string), `objective_type` (string), `objective_value` (number), `xp_reward` (integer), `room_id` (string), `valid_until` (ISO8601 date string).

#### Scenario: Evolved pack contains bounty contracts
- **WHEN** the tuner runs a full analyze+generate cycle
- **THEN** the saved evolved pack YAML SHALL contain a `bounty_contracts` list with at least one entry

#### Scenario: Contracts have all required fields
- **WHEN** the evolved pack is parsed
- **THEN** each entry in `bounty_contracts` SHALL have non-empty `id`, `description`, `objective_type`, `xp_reward`, and `valid_until` fields

### Requirement: Bounty completion triggers XP burst
When `game.bounty.completed` is received with a valid contract `id`, the signal handler SHALL award the contract's `xp_reward` to the player and mark the contract as completed in SQLite.

#### Scenario: XP burst on bounty completion
- **WHEN** `game.bounty.completed` is received with `contract_id` matching an active contract with `xp_reward: 500`
- **THEN** a game run record with `xp: 500` and `source: "bounty"` SHALL be inserted

#### Scenario: Expired contract ignored
- **WHEN** `game.bounty.completed` is received but the contract's `valid_until` is in the past
- **THEN** no XP SHALL be awarded and an error SHALL be logged

### Requirement: Expired contracts are cleared from pack state
On startup, the system SHALL remove any bounty contracts whose `valid_until` timestamp is in the past from the active pack's in-memory contract list.

#### Scenario: Expired contracts pruned on load
- **WHEN** the pack is loaded and one contract has a `valid_until` timestamp before the current time
- **THEN** that contract SHALL not appear in the active pack's `BountyContracts` slice
