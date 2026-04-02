## Why

The gamification system in `internal/game/` has hardcoded achievement thresholds, a static narrator, dead ICE classes, and an XP formula that never changes — meaning the game stagnates the moment you start using it. A self-evolving tuner closes the feedback loop: your actual usage patterns drive continuous improvement of the rules, weights, and narration without any user configuration.

## What Changes

- **New**: `internal/game/tuner.go` — `Tuner` struct with `Tune(ctx)` method; fires async after meaningful game events (level-up, streak milestone, new achievement, or 7-day floor with 5+ new runs); cooldown of 1 tune per day maximum
- **New**: `internal/store/query.go` — `GameStatsQuery()` returning aggregate stats (token ratios, cache rates, cost/run, provider mix, p50/p90 metrics, unlocked achievement IDs, step failure rate)
- **Modified**: `internal/game/pack.go` — `GameWorldPack` gains a `Weights` field (`PackWeights` struct)
- **Modified**: `internal/game/engine.go` — `ComputeXP()` accepts `PackWeights`; provider multipliers, streak multiplier, speed bonus scale, cache bonus rate all tunable
- **Modified**: `internal/game/packs/cyberspace/pack.yaml` — adds `weights:` section with sensible defaults; `game_rules` gains ICE class rules for `trace-ice` and `data-ice` (currently dead); `quest_events` logic is no longer always `[]`
- **New**: `glitch game tune` CLI subcommand — manual trigger; runs synchronously, prints what changed
- **Behavior**: After each scored run the tuner checks trigger conditions async; if conditions are met it calls Ollama twice (analyze → generate), validates output, writes `~/.local/share/glitch/agents/game-world-tuned.agent.md`; `APMWorldPackLoader` picks this up automatically on the next run

## Capabilities

### New Capabilities

- `game-tuner`: Self-evolving game pack tuner — reads behavioral stats from SQLite, calls local Ollama to analyze current rules against actual patterns, generates an evolved pack with calibrated achievement thresholds, new achievements, active ICE classes, quest event triggers, arc-aware narrator style, and evolved XP formula weights; writes output as a game-world agent file
- `game-pack-weights`: XP formula weights embedded in `GameWorldPack`; `ComputeXP()` accepts weights so provider multipliers, streak multiplier, speed bonus, and cache bonus rate are all runtime-tunable without recompile

### Modified Capabilities

- `pipeline-step-lifecycle`: Scoring hook that fires the tuner async after a run is scored (trigger condition check only — tuner runs in background goroutine)

## Impact

- `internal/game/` — new file `tuner.go`; `engine.go` and `pack.go` signature changes
- `internal/store/query.go` — new `GameStatsQuery()` method
- `internal/game/packs/cyberspace/pack.yaml` — schema addition (weights), content updates (ICE rules, quest events)
- All callers of `ComputeXP()` need to pass `PackWeights` (or a zero-value default)
- No external dependencies added — Ollama already in use, SQLite already in use
