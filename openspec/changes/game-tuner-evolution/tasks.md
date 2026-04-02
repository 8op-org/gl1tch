## 1. PackWeights — struct, defaults, pack YAML

- [x] 1.1 Add `PackWeights` struct to `internal/game/pack.go` with fields: BaseMultiplier, CacheBonusRate, SpeedBonusCap, SpeedBonusScale, RetryPenalty, StreakMultiplier, ProviderWeights map[string]float64
- [x] 1.2 Add `DefaultPackWeights()` function returning values that reproduce the current ComputeXP formula exactly
- [x] 1.3 Add `Weights PackWeights` field to `GameWorldPack` struct
- [x] 1.4 Update `DefaultWorldPackLoader.ActivePack()` to populate Weights from parsed YAML; fall back to DefaultPackWeights() if weights section absent
- [x] 1.5 Update `parseGameWorldAgent()` to parse weights section from frontmatter
- [x] 1.6 Add `weights:` section to `internal/game/packs/cyberspace/pack.yaml` matching DefaultPackWeights values

## 2. ComputeXP — accept PackWeights, fix all call sites

- [x] 2.1 Update `ComputeXP(usage TokenUsage, retryCount int, weights PackWeights)` signature in `internal/game/engine.go`
- [x] 2.2 Apply weights in ComputeXP: BaseMultiplier on base, CacheBonusRate on cache bonus, SpeedBonusCap/Scale on speed bonus, RetryPenalty per retry, ProviderWeights multiplier on Final, StreakMultiplier on Final when streak > 1
- [x] 2.3 Find all callers of ComputeXP and update them to pass `pack.ActivePack().Weights` or `DefaultPackWeights()`
- [x] 2.4 Update `engine_test.go` to pass weights to ComputeXP; add test cases for provider multiplier and streak multiplier

## 3. Default pack — ICE rules and quest events

- [x] 3.1 Update `game_rules` in `packs/cyberspace/pack.yaml`: add trace-ice condition (cost_usd > threshold), add data-ice condition (input_tokens > threshold), keep black-ice on step_failures > 0
- [x] 3.2 Add basic quest_event triggers to `game_rules`: streak milestones, XP thresholds, first achievement unlock
- [x] 3.3 Verify existing `engine_ollama_test.go` still passes with updated pack

## 4. Store — GameStatsQuery

- [x] 4.1 Add `GameStats` struct to `internal/store/store.go`: TotalRuns, AvgOutputRatio, P50OutputRatio, P90OutputRatio, AvgCacheReadTokens, P90CacheReadTokens, AvgCostUSD, ProviderRunCounts map[string]int64, UnlockedAchievementIDs []string, StepFailureRate float64, RunsSinceDate int64
- [x] 4.2 Implement `GameStatsQuery(ctx context.Context, sinceDays int) (GameStats, error)` in `internal/store/query.go` using SQL aggregates over score_events and a join with achievements
- [x] 4.3 Add tests in `internal/store/` covering: empty window returns zero struct, stats aggregate correctly over window, provider mix populated correctly

## 5. Tuner — core implementation

- [x] 5.1 Create `internal/game/tuner.go` with `Tuner` struct holding: `*store.Store`, `*GameEngine`, `packLoader WorldPackLoader`, `agentPath string`
- [x] 5.2 Implement `NewTuner(store, engine, loader)` that resolves agentPath to `~/.local/share/glitch/agents/game-world-tuned.agent.md`
- [x] 5.3 Implement `ShouldTune(ctx, stats GameStats, lastTunedAt time.Time, prevLevel, newLevel int, prevStreak, newStreak int, newAchievements []string) bool` with all trigger conditions and 1-day cooldown
- [x] 5.4 Implement `buildAnalysisPrompt(stats GameStats, currentPack GameWorldPack) string` — includes current game_rules, behavioral stats, instructs Ollama to return JSON analysis
- [x] 5.5 Implement `buildGenerationPrompt(analysisJSON string, currentPack GameWorldPack) string` — instructs Ollama to produce complete pack YAML with evolved rules, narrator with arc context, and weights
- [x] 5.6 Implement `Tune(ctx context.Context, stats GameStats, score GameRunScoredPayload) error` — calls Ollama twice, retries analysis on JSON parse failure, validates, writes agent file
- [x] 5.7 Implement `validate(yamlBytes []byte) error` — checks: parses as valid YAML, ≥5 achievements in game_rules, all numeric thresholds > 0, weights in [0.1, 5.0]
- [x] 5.8 Add `tuner_test.go`: test ShouldTune trigger conditions, cooldown, validate rejection cases

## 6. Post-score hook — wire tuner into scoring flow

- [x] 6.1 Locate where XP is scored and user_score is updated (find the scoring call site that produces GameRunScoredPayload)
- [x] 6.2 Instantiate `Tuner` at app startup alongside `GameEngine` and pass it through to the scoring path
- [x] 6.3 After scoring, call `store.GameStatsQuery(ctx, 30)` and check `tuner.ShouldTune(...)` with detached context
- [x] 6.4 If ShouldTune returns true, launch `go tuner.Tune(context.Background(), stats, payload)` — non-blocking
- [x] 6.5 Persist `last_tuned_at` and `runs_since_tune` counter — decide storage: new columns in user_score table or a simple flat file in glitch config dir (flat file preferred to avoid schema migration)

## 7. CLI subcommand — glitch game tune

- [x] 7.1 Create `glitch game tune` subcommand (locate existing game subcommand or create new one under the appropriate cmd directory)
- [x] 7.2 Run tuner synchronously, capturing what changed by diffing old vs new pack
- [x] 7.3 Print human-readable summary: achievement thresholds adjusted, new achievements added, ICE rules changed, weight deltas
- [x] 7.4 Exit 1 with clear error message if Ollama unavailable or validation fails

## 8. Integration verification

- [x] 8.1 Run `go build ./...` — confirm no compilation errors
- [x] 8.2 Run `go test ./internal/game/... ./internal/store/...` — confirm all tests pass
- [x] 8.3 Manually run `glitch game tune` and verify agent file is written to correct path
- [x] 8.4 Confirm `APMWorldPackLoader` picks up the new agent file on next run (check pack name in game narration output)
