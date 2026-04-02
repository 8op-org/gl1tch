## Context

`internal/game/` is a complete gamification subsystem: token capture, XP computation, level table, Ollama-driven achievement evaluation and narration, and a pack system with an APM agent override mechanism. The default pack (`packs/cyberspace/pack.yaml`) is embedded at compile time. `APMWorldPackLoader` checks `~/.local/share/glitch/agents/` for a `kind: game-world` agent file first — so writing to that path is the correct extension point; no loader changes are needed.

The gaps: achievement thresholds are hardcoded guesses, `trace-ice` and `data-ice` have no triggering rules, `quest_events` always returns `[]`, the narrator has no awareness of long-term arc, and `ComputeXP()` treats all providers equally with a formula that never changes.

## Goals / Non-Goals

**Goals:**
- Tuner fires automatically after meaningful game events with zero user configuration
- Tuner uses local Ollama exclusively (two-call flow: analyze → generate)
- Evolved pack writes to APM agents dir; picked up on next run without restart
- `PackWeights` in the pack schema makes the XP formula runtime-tunable
- `glitch game tune` command for manual/cron use
- ICE classes `trace-ice` and `data-ice` get real trigger conditions
- Narrator style embeds user arc (level, streak, achievements, provider mix)

**Non-Goals:**
- Storing per-run achievement firing history (not in schema; not worth adding for this)
- Multi-user or shared game state
- Cloud sync of evolved packs
- Changing the SQLite schema (all needed data already exists)

## Decisions

### Decision: Two Ollama calls (analyze → generate), not one

One call asking Ollama to both analyze stats and produce a full YAML pack risks format confusion and makes the prompt unwieldy. Two calls — first returning structured JSON analysis, second taking that analysis to generate YAML — produces more reliable output and lets each call be focused.

*Alternative considered*: Single call with a complex prompt. Rejected: Ollama models are more reliable when tasks are decomposed.

### Decision: Trigger on game events, not arbitrary run count or wall-clock cron

Level-up, streak milestones (3, 7, 14, 30, 60, 90 days), and first unlock of any achievement are meaningful moments when recalibration is naturally motivated. A 7-day floor with ≥5 new runs ensures the game still evolves for steady users who don't level up frequently. A 1-day cooldown prevents thrash.

*Alternative considered*: `total_runs % 25`. Rejected: arbitrary, not meaningful.

*Alternative considered*: External cron entry only. Rejected: requires user configuration; not "just works."

### Decision: PackWeights in pack YAML, ComputeXP accepts weights param

Moving multipliers into the pack makes the XP formula evolve without recompile. `ComputeXP()` gets a `PackWeights` parameter; callers that don't have a pack pass `DefaultPackWeights()`. The default weights reproduce the current formula exactly — no behavioral change at rollout.

*Alternative considered*: Keep formula fixed in Go, tune only prompts. Rejected: user asked for all four dimensions of evolution including the formula.

### Decision: Validate before install; no-op on failure

The validate step checks: YAML parses cleanly, ≥5 achievements defined, all numeric thresholds > 0, weights within sane ranges (multipliers 0.1–5.0). On any failure the existing agent file is left untouched. The tuner logs the failure but does not surface it to the user — game is optional infrastructure.

### Decision: Tuner runs in a background goroutine, non-blocking

The scoring hook checks trigger conditions synchronously (fast: two int comparisons, one time check). If conditions are met, `go tuner.Tune(ctx)` is called. The user never waits for tuning. This matches how the existing `GameEngine.Evaluate` and `Narrate` calls are treated — game is optional, failures are silent.

### Decision: `glitch game tune` subcommand runs synchronously and prints diff

For manual use and cron use, synchronous output is more useful. The command prints a human-readable summary of what changed (achievement thresholds adjusted, new achievements added, ICE rules updated, weight deltas).

## Risks / Trade-offs

- **Ollama drift** → The tuner could produce progressively weirder packs if each run uses the previous tuner output as input. Mitigation: the analysis call always compares against the *embedded default* pack's achievement list as a baseline, not the last tuned pack. The generation call takes the evolved pack as starting point but the analysis anchors to defaults.

- **Achievement ID churn** → Tuner invents new achievement IDs; old IDs in the `achievements` table become orphaned. Mitigation: orphaned IDs are harmless (they just never fire). The tuner is instructed to preserve existing IDs and only add new ones.

- **Provider weight escalation** → If the tuner sees 100% Ollama runs it might push Ollama multiplier very high, making paid-provider runs feel disproportionately rewarding later. Mitigation: weight bounds check in validate (0.1–5.0); tuner prompt instructs relative balance, not absolute escalation.

- **ComputeXP signature change** → All existing callers break at compile time. Mitigation: `DefaultPackWeights()` provides a zero-config default; callers pass it until they have a pack reference. This is a clean breaking change detectable at compile time.

- **Tune latency visible in TUI** → Background goroutine makes an HTTP call to Ollama (2× for two calls). Could add 5–15s of background work. Mitigation: goroutine is fully detached; no TUI blocking.

## Migration Plan

1. Add `PackWeights` struct and `DefaultPackWeights()` to `pack.go`
2. Update `ComputeXP()` signature — fix all call sites (compiler-guided)
3. Add `weights:` section to `packs/cyberspace/pack.yaml`
4. Add `GameStatsQuery()` to `store/query.go`
5. Implement `internal/game/tuner.go`
6. Wire trigger check into the scoring hook (post-score, async)
7. Add `glitch game tune` subcommand
8. Update `game_rules` in default pack: ICE class rules for trace-ice/data-ice, quest_events logic

Rollback: remove `game-world-tuned.agent.md` from agents dir; `APMWorldPackLoader` falls back to embedded default automatically.

## Open Questions

- Should `glitch game tune --dry-run` be in scope for this change? (Would print proposed pack without writing.) Nice to have, not required for the feature to work.
- Streak milestone list: 3, 7, 14, 30, 60, 90 — correct cadence, or should it extend further?
