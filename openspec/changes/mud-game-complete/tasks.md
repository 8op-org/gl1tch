## 1. SQLite Schema Migrations

- [ ] 1.1 Add `achievements` table migration to `internal/store/store.go` (id TEXT PK, unlocked_at DATETIME, run_id TEXT)
- [ ] 1.2 Add `ice_encounters` table migration (id TEXT PK, ice_class TEXT, run_id TEXT, deadline DATETIME, resolved BOOL, outcome TEXT)
- [ ] 1.3 Add `game_personal_bests` table migration (metric TEXT PK, value REAL, run_id TEXT, recorded_at DATETIME)
- [ ] 1.4 Verify all three tables are created in the existing migration chain and do not break existing startup

## 2. Store Query Layer

- [ ] 2.1 Add `InsertAchievement(id, runID string) error` to `internal/store/store.go` — insert-or-ignore
- [ ] 2.2 Add `HasAchievement(id string) (bool, error)` to `internal/store/store.go`
- [ ] 2.3 Update `GameStatsQuery` in `internal/store/query.go` to populate `UnlockedAchievementIDs` from the achievements table
- [ ] 2.4 Add `InsertOrUpdatePersonalBest(metric string, value float64, runID string) error` to `internal/store/store.go`
- [ ] 2.5 Add `GetPersonalBests() ([]PersonalBest, error)` to `internal/store/store.go`
- [ ] 2.6 Add `InsertICEEncounter(id, iceClass, runID string, deadline time.Time) error` to `internal/store/store.go`
- [ ] 2.7 Add `ResolveICEEncounter(id, outcome string) error` and `GetPendingICEEncounter() (*ICEEncounter, error)` to `internal/store/store.go`
- [ ] 2.8 Add `AutoResolveExpiredEncounters(applyPenalty func()) error` to `internal/store/store.go` — called on startup

## 3. Pack Schema Extensions

- [ ] 3.1 Add `MUDXPEvents map[string]int` to `PackWeights` struct in `internal/game/pack.go`
- [ ] 3.2 Add `BountyContracts []BountyContract` to pack struct; define `BountyContract` type (id, description, objective_type, objective_value, xp_reward, room_id, valid_until)
- [ ] 3.3 Add `ReputationDecay` struct to pack (decay_per_day int, floor int, max_decay_days int) with defaults
- [ ] 3.4 Update pack YAML loader to parse `mud_xp_events`, `bounty_contracts`, and `reputation_decay` fields
- [ ] 3.5 Add startup pruning: filter expired `BountyContracts` from in-memory pack on load
- [ ] 3.6 Update `internal/game/packs/cyberspace/pack.yaml` — add `mud_xp_events` with defaults for `mud.room.entered` (10 XP), `mud.espionage.talked` (25 XP), `mud.hack.success` (50 XP); add `reputation_decay` defaults

## 4. Pipeline Runner Wiring

- [ ] 4.1 In `internal/pipeline/runner.go`, replace `Achievements: []string{}` with `Achievements: result.Achievements` (use empty slice if nil)
- [ ] 4.2 Add `ICEClass: result.ICEClass` to the scored payload struct
- [ ] 4.3 After scoring, if `result.ICEClass != ""`, emit `game.ice.encountered` signal with `ice_class`, `run_id`, `pipeline_name`
- [ ] 4.4 After scoring, loop `result.QuestEvents` and emit one `game.quest.event` signal per entry with `event_type`, `payload`, `run_id`

## 5. Signal Handlers

- [ ] 5.1 Add `game.achievement.unlocked` handler in `internal/console/signal_handlers.go` — render styled unlock notification in TUI feed
- [ ] 5.2 Add `game.ice.encountered` handler — insert ICE encounter row via store, log to signal log
- [ ] 5.3 Add `game.quest.event` handler — append to plugin signals log (extends existing log handler)
- [ ] 5.4 Add score handler update: when `game.run.scored` payload has non-empty `ice_class`, render ICE badge line in score card
- [ ] 5.5 Add score handler update: when `game.run.scored` has non-empty `achievements`, emit `game.achievement.unlocked` for each new one (checking HasAchievement before insert)
- [ ] 5.6 Register MUD XP bridge handlers on startup — iterate active pack's `MUDXPEvents` map; for each topic, register a handler that awards XP via a synthetic run record with dedup on signal ID
- [ ] 5.7 Add `game.bounty.completed` handler — validate contract not expired, award `xp_reward` via synthetic run record, mark contract completed in SQLite

## 6. Tuner Extensions

- [ ] 6.1 Extend tuner analysis prompt in `internal/game/tuner.go` to request 3–5 bounty contracts (include room IDs from cyberspace world)
- [ ] 6.2 Extend tuner generate prompt to include bounty contracts in evolved pack YAML output
- [ ] 6.3 Extend tuner validation to check that any `bounty_contracts` entries have required fields and positive `xp_reward`

## 7. ICE Encounter CLI

- [ ] 7.1 Add `ice` subcommand to `cmd/game.go` — reads pending encounter from store, presents fight/jack-out prompt, records outcome
- [ ] 7.2 Wire streak decrement on loss: call existing streak-update path with `-1` delta, bounded at 0
- [ ] 7.3 Add startup hook in `main.go` (or store init) to call `AutoResolveExpiredEncounters` with streak penalty callback

## 8. Reputation Decay

- [ ] 8.1 Add `ApplyReputationDecay(pack *Pack) error` to store — computes inactive days from last run timestamp, applies decay per faction per day up to `max_decay_days`, respects `floor`
- [ ] 8.2 Call `ApplyReputationDecay` on glitch startup after pack is loaded
- [ ] 8.3 Write unit test covering: 1-day decay, cap at max_decay_days, floor clamping, no decay on same-day run

## 9. Personal Bests Tracking

- [ ] 9.1 After each `game.run.scored` is processed, call `InsertOrUpdatePersonalBest` for: fastest_run_ms, highest_xp, most_cache_tokens, lowest_cost_usd
- [ ] 9.2 After each streak update, call `InsertOrUpdatePersonalBest` for `longest_streak` if current streak exceeds stored best

## 10. Recap CLI

- [ ] 10.1 Add `recap` subcommand to `cmd/game.go` with `--days` flag (default 7)
- [ ] 10.2 Implement fetch: query `game_runs` for last N days, format as JSON summary (total runs, total XP, achievements, avg XP/run, best streak)
- [ ] 10.3 Implement Ollama call: send JSON summary + pack narrator style prompt, stream output to stdout
- [ ] 10.4 Implement graceful degradation: on Ollama error, print plain-text stats table and exit 0

## 11. Leaderboard CLI

- [ ] 11.1 Add `top` subcommand to `cmd/game.go`
- [ ] 11.2 Query `game_personal_bests` and render a lipgloss-styled table (metric | value | date achieved)
- [ ] 11.3 Handle empty state: print "No personal bests recorded yet." and exit 0

## 12. Tests

- [ ] 12.1 Unit test `InsertAchievement` + `HasAchievement` idempotency in `internal/store/gamestats_test.go`
- [ ] 12.2 Unit test personal bests update logic — new best overwrites, non-best does not
- [ ] 12.3 Unit test reputation decay: 1-day, cap, floor, same-day no-op
- [ ] 12.4 Unit test pack loading with `mud_xp_events` and `bounty_contracts` fields
- [ ] 12.5 Unit test tuner validation accepts packs with `bounty_contracts`
