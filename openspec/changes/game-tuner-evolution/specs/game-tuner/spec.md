## ADDED Requirements

### Requirement: Tuner fires automatically after meaningful game events
The tuner SHALL check trigger conditions after every scored run. Conditions that trigger a tune: (a) the run caused a level-up, (b) the run hit a streak milestone (3, 7, 14, 30, 60, or 90 consecutive days), (c) the run unlocked a new achievement. Additionally, the tuner SHALL fire if ≥7 days have elapsed since the last tune AND ≥5 runs have been scored since then. A 1-day cooldown SHALL prevent more than one tune per calendar day regardless of how many trigger conditions fire.

#### Scenario: Level-up triggers tune
- **WHEN** a scored run causes the player's level to increase
- **THEN** the tuner fires asynchronously after scoring completes

#### Scenario: Streak milestone triggers tune
- **WHEN** a scored run results in streak_days equal to 3, 7, 14, 30, 60, or 90
- **THEN** the tuner fires asynchronously after scoring completes

#### Scenario: New achievement triggers tune
- **WHEN** a scored run unlocks an achievement that was not previously in the achievements table
- **THEN** the tuner fires asynchronously after scoring completes

#### Scenario: 7-day floor triggers tune
- **WHEN** ≥7 days have elapsed since last_tuned_at AND ≥5 runs have been scored since last tune
- **THEN** the tuner fires asynchronously on the next scored run

#### Scenario: Cooldown suppresses rapid firing
- **WHEN** the tuner has already run today (same calendar date as now)
- **THEN** no additional tune fires even if trigger conditions are met

#### Scenario: Tuner failure is silent
- **WHEN** Ollama is unavailable or returns unparseable output
- **THEN** the existing agent file is left unchanged and no error surfaces to the user

### Requirement: Tuner uses two sequential Ollama calls
The tuner SHALL call local Ollama twice. The first call (analysis) SHALL provide the current pack rules and behavioral stats from SQLite, and return a structured JSON object identifying over-easy achievements (>70% estimated unlock rate based on metric distributions), unreachable achievements (<5% estimated unlock rate), uncovered behavioral patterns worth rewarding, conditions for trace-ice and data-ice, quest event triggers, and narrator arc notes. The second call (generation) SHALL take the analysis JSON and current pack, and return a complete valid pack YAML including evolved `game_rules`, `narrator_style`, and `weights` sections.

#### Scenario: Analysis call returns structured JSON
- **WHEN** the analysis Ollama call completes
- **THEN** its output parses as JSON with keys: calibrations, new_achievements, ice_rules, quest_rules, narrator_notes, weight_suggestions

#### Scenario: Generation call produces full YAML
- **WHEN** the generation Ollama call completes
- **THEN** its output parses as valid YAML containing game_rules, narrator_style, and weights fields

#### Scenario: Retry on malformed JSON from analysis call
- **WHEN** the analysis call returns non-JSON output
- **THEN** the tuner retries once with a stricter prompt before aborting

### Requirement: Tuner validates output before installing
The tuner SHALL validate generated pack YAML before writing it to the agents directory. Validation SHALL reject output that: fails YAML parsing, defines fewer than 5 achievements, contains numeric thresholds ≤ 0, or contains weight multipliers outside the range 0.1–5.0.

#### Scenario: Invalid YAML rejected
- **WHEN** the generated output fails yaml.Unmarshal
- **THEN** the existing agent file is not modified

#### Scenario: Too few achievements rejected
- **WHEN** the generated game_rules defines fewer than 5 achievements
- **THEN** the pack is rejected and existing file is not modified

#### Scenario: Out-of-range weight rejected
- **WHEN** any weight multiplier in the generated pack is outside [0.1, 5.0]
- **THEN** the pack is rejected and existing file is not modified

#### Scenario: Valid pack installed
- **WHEN** all validation checks pass
- **THEN** the pack is written to ~/.local/share/glitch/agents/game-world-tuned.agent.md

### Requirement: Tuner embeds user arc in narrator_style
The generated `narrator_style` SHALL incorporate the user's current level title, streak length, total runs, unlocked achievement IDs, and dominant provider. The narration character (Zero Cool / The Gibson) SHALL remain consistent; only the depth and stakes of the narration SHALL evolve as the player progresses.

#### Scenario: High-level player gets gravitas
- **WHEN** the player is level 8+ with a 14+ day streak
- **THEN** the generated narrator_style instructs the narrator to treat them as seasoned and speak with weight

#### Scenario: New player gets orientation
- **WHEN** the player has fewer than 10 total runs
- **THEN** the generated narrator_style instructs the narrator to be more explanatory and atmospheric

### Requirement: ICE classes trace-ice and data-ice have active trigger rules
The generated `game_rules` SHALL define conditions for all three ICE classes. `black-ice` SHALL trigger on step_failures > 0 (existing behavior). `trace-ice` SHALL trigger when cost_usd exceeds the user's average cost by a configurable multiplier. `data-ice` SHALL trigger when input_tokens exceeds the user's p90 input token count.

#### Scenario: trace-ice fires on high-cost run
- **WHEN** a run's cost_usd exceeds the threshold defined in game_rules
- **THEN** Ollama's Evaluate call returns ice_class: "trace-ice"

#### Scenario: data-ice fires on token spike
- **WHEN** a run's input_tokens exceeds the threshold defined in game_rules
- **THEN** Ollama's Evaluate call returns ice_class: "data-ice"

### Requirement: glitch game tune subcommand
The CLI SHALL expose a `glitch game tune` subcommand that runs the tuner synchronously, prints a human-readable summary of what changed (achievement thresholds adjusted, new achievements, ICE rules updated, weight deltas), and exits 0 on success or 1 on failure.

#### Scenario: Manual tune prints summary
- **WHEN** user runs `glitch game tune`
- **THEN** the command prints what changed and exits 0

#### Scenario: No Ollama prints error
- **WHEN** Ollama is not running and user runs `glitch game tune`
- **THEN** the command prints an error message and exits 1

### Requirement: Store exposes GameStatsQuery
The store SHALL provide a `GameStatsQuery(ctx, sinceDays int)` method returning aggregate behavioral stats over the specified window: total runs, average output/input ratio, p50 and p90 of output ratio, average and p90 cache read tokens, average cost per run, run count per provider, unlocked achievement IDs, and step failure rate (runs with XP=0 divided by total runs).

#### Scenario: Stats computed over window
- **WHEN** GameStatsQuery is called with sinceDays=30
- **THEN** it returns stats aggregated from score_events in the last 30 days only

#### Scenario: Empty window returns zero stats
- **WHEN** no score_events exist within the window
- **THEN** GameStatsQuery returns a zero-value GameStats struct with no error
