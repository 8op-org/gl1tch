## ADDED Requirements

### Requirement: `glitch game recap` subcommand exists
The `glitch game` command group SHALL include a `recap` subcommand. It SHALL accept an optional `--days N` flag (default 7). It SHALL fetch the last N days of `game_runs` from SQLite, format them as a JSON summary, and send them to Ollama with the active pack's narrator prompt requesting a story-arc narrative.

#### Scenario: Recap with default days
- **WHEN** user runs `glitch game recap` with no flags
- **THEN** the last 7 days of runs are fetched and a narrative is printed to stdout

#### Scenario: Recap with custom day range
- **WHEN** user runs `glitch game recap --days 14`
- **THEN** the last 14 days of runs are fetched for the narrative

#### Scenario: Recap with no runs in range
- **WHEN** no runs exist in the requested date range
- **THEN** a message "No runs recorded in the last N days." is printed and the command exits 0

### Requirement: Recap degrades gracefully when Ollama is unavailable
If the Ollama call fails or times out, `glitch game recap` SHALL print the raw stats table (total runs, total XP, achievements unlocked, average XP/run) instead of a narrative.

#### Scenario: Ollama unavailable falls back to stats table
- **WHEN** Ollama is not reachable and `glitch game recap` is run
- **THEN** a plain-text stats table is printed and the command exits 0 (not 1)

### Requirement: Recap uses pack narrator voice
The Ollama prompt for recap SHALL use the active pack's narrator style configuration, matching the narration tone used for individual run narrations.

#### Scenario: Recap prompt includes narrator style
- **WHEN** recap Ollama call is constructed
- **THEN** the system prompt SHALL include the pack's narrator style text
