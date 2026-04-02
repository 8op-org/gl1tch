## ADDED Requirements

### Requirement: Achievements table exists in SQLite
The store SHALL maintain an `achievements` table with columns: `id` (TEXT PK), `unlocked_at` (DATETIME), `run_id` (TEXT FK to game_runs).

#### Scenario: Schema migration on startup
- **WHEN** glitch starts and the `achievements` table does not exist
- **THEN** the store SHALL create the table without error and without affecting existing data

### Requirement: Achievement unlock persisted from scored payload
When a `game.run.scored` signal carries a non-empty `achievements` list, the game signal handler SHALL insert each achievement ID that is not already present in the `achievements` table.

#### Scenario: New achievement from run
- **WHEN** `game.run.scored` payload contains `achievements: ["speed-demon"]` and `speed-demon` is not in the table
- **THEN** a row with `id=speed-demon` and the current timestamp SHALL be inserted

#### Scenario: Duplicate achievement suppressed
- **WHEN** `game.run.scored` payload contains `achievements: ["speed-demon"]` and `speed-demon` already exists in the table
- **THEN** no new row is inserted and no error is returned

### Requirement: Achievement unlock notification in TUI feed
When a new achievement is persisted, the system SHALL emit a `game.achievement.unlocked` BUSD signal. The console signal handler for this topic SHALL render a styled notification line in the TUI feed.

#### Scenario: Unlock notification rendered
- **WHEN** `game.achievement.unlocked` is received with `achievement_id: "speed-demon"`
- **THEN** a styled line (e.g., `[ACHIEVEMENT UNLOCKED] speed-demon`) SHALL appear in the TUI feed output

### Requirement: GameStatsQuery returns unlocked achievement IDs
`GameStatsQuery` SHALL return a populated `UnlockedAchievementIDs []string` field containing all achievement IDs from the `achievements` table.

#### Scenario: Stats query includes achievements
- **WHEN** `GameStatsQuery` is called and two achievements exist in the table
- **THEN** the returned `GameStats.UnlockedAchievementIDs` SHALL contain both IDs
