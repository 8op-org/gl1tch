## ADDED Requirements

### Requirement: `game_personal_bests` table in SQLite
The store SHALL maintain a `game_personal_bests` table with one row per metric. Tracked metrics: `fastest_run_ms` (INTEGER), `highest_xp` (INTEGER), `longest_streak` (INTEGER), `most_cache_tokens` (INTEGER), `lowest_cost_usd` (REAL). Each row has a `run_id` FK and `recorded_at` timestamp.

#### Scenario: Table created on startup
- **WHEN** glitch starts and the `game_personal_bests` table does not exist
- **THEN** the store SHALL create it without error

### Requirement: Personal bests updated after every scored run
After a `game.run.scored` event is processed, the store SHALL compare the run's metrics against the current personal bests and update any that are surpassed.

#### Scenario: New fastest run recorded
- **WHEN** a run completes with `duration_ms: 1200` and the current `fastest_run_ms` best is 2000
- **THEN** `fastest_run_ms` SHALL be updated to 1200 with the new `run_id`

#### Scenario: Existing best not overwritten when not surpassed
- **WHEN** a run completes with `xp: 50` and the current `highest_xp` best is 200
- **THEN** `highest_xp` SHALL remain 200

### Requirement: `glitch game top` subcommand renders personal bests
The `glitch game` command group SHALL include a `top` subcommand. It SHALL read `game_personal_bests` and render a styled table of all tracked metrics with their values, the date achieved, and the associated run ID.

#### Scenario: Top renders all metrics
- **WHEN** user runs `glitch game top` and personal bests exist for all metrics
- **THEN** a styled table with all 5 metrics, their values, and dates is printed to stdout

#### Scenario: Top with no data
- **WHEN** no scored runs have been recorded
- **THEN** `glitch game top` prints "No personal bests recorded yet." and exits 0
