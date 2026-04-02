## ADDED Requirements

### Requirement: Pack defines MUD XP event table
The pack YAML format SHALL support a `mud_xp_events` map at the top level. Keys are BUSD topic names (e.g., `mud.room.entered`); values are XP amounts (positive integers).

#### Scenario: Pack YAML parses mud_xp_events
- **WHEN** a pack YAML contains `mud_xp_events: {mud.room.entered: 10, mud.espionage.talked: 25}`
- **THEN** `PackWeights.MUDXPEvents` SHALL be populated with those key-value pairs after loading

#### Scenario: Missing mud_xp_events is not an error
- **WHEN** a pack YAML has no `mud_xp_events` key
- **THEN** the pack SHALL load successfully with an empty `MUDXPEvents` map

### Requirement: Signal handler bridge awards XP for MUD events
The console signal handler setup SHALL register one handler per topic in the active pack's `MUDXPEvents` map. Each handler SHALL award XP by calling the game scoring path with a synthetic run record.

#### Scenario: Room entered awards XP
- **WHEN** `mud.room.entered` is received and the active pack maps it to `10` XP
- **THEN** a game run record with `xp: 10` and `source: "mud"` SHALL be inserted into `game_runs`

#### Scenario: XP deduped per signal instance
- **WHEN** the same `mud.room.entered` signal is received twice with the same signal ID
- **THEN** XP SHALL be awarded only once (idempotent on signal ID)

### Requirement: Cyberspace default pack includes MUD XP event defaults
The embedded `cyberspace/pack.yaml` SHALL include a `mud_xp_events` section with sensible defaults covering `mud.room.entered`, `mud.espionage.talked`, and `mud.hack.success`.

#### Scenario: Default pack has mud XP entries
- **WHEN** the default pack is loaded with no APM override
- **THEN** `PackWeights.MUDXPEvents` SHALL contain at least three entries
