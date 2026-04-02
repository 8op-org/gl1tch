## Context

The gl1tch game layer was built incrementally. Each piece — XP scoring, Ollama achievement evaluation, ICE classification, quest events, the MUD signal bus — was added in isolation. The result is a system where all the wiring exists but the outputs never connect: achievements are evaluated and dropped, ICE class is parsed and ignored, quest events are parsed and ignored, and the MUD runs a completely parallel progression track. The five wiring gaps are code that already exists but has dead outputs. The five new features (bounty boards, ICE encounters, reputation decay, recap, leaderboard) are net-new but designed to slot into the existing BUSD/SQLite/Ollama patterns already established.

**Constraints:**
- All LLM calls use local Ollama — no external AI APIs
- BUSD (Bus of Distributed Signals) is the integration layer between components
- SQLite is the persistence layer (`internal/store/`)
- Shell pipeline steps own data fetching; LLM steps own formatting
- tmux required for MUD session management

## Goals / Non-Goals

**Goals:**
- All Ollama `EvaluateResult` fields (Achievements, ICEClass, QuestEvents) reach their final destination
- MUD activity (room discovery, NPC talk, hack success) contributes XP to the game progression track
- Achievement unlocks are visible in the TUI at the moment they happen
- Two new CLI subcommands: `glitch game recap` and `glitch game top`
- Tuner produces bounty contracts that MUD rooms can host
- ICE class triggers a TUI combat sequence with streak stakes
- Reputation decay creates daily-use pressure linked to pipeline activity

**Non-Goals:**
- Multiplayer or cloud sync of any kind
- External leaderboards or score sharing
- Modifying the MUD binary directly (all integration is signal-bus based)
- Replacing Ollama with any external LLM service

## Decisions

### 1. Achievements: persist on `game.run.scored`, notify via existing feed

The fix is surgical: `runner.go` passes `result.Achievements` through to the scored payload. A new `signal_handlers.go` handler subscribes to `game.run.scored` and for each achievement ID not already in the achievements table, inserts it and emits a `game.achievement.unlocked` event. The TUI feed already renders styled output from signal handlers — no new rendering path needed.

**Alternatives considered:**
- Emit `game.achievement.unlocked` directly from runner.go — rejected because the runner shouldn't own persistence logic
- Batch achievements into a separate Ollama call — unnecessary, evaluation already returns them

### 2. ICE class: signal emission in runner.go, badge in score output, encounter as separate goroutine

When `result.ICEClass != ""`, runner.go emits `game.ice.encountered` with the class name and run ID. The score handler already renders a styled score card — extend it with a conditional ICE badge line. The ICE encounter (streak risk) fires as a background goroutine triggered by the signal handler, giving the user a short window (configurable, default 30s) to respond via `glitch game ice` before it auto-resolves as a loss.

**Alternatives considered:**
- ICE encounter blocks the pipeline run — rejected, too disruptive to workflow
- ICE encounter is purely cosmetic — rejected, streak risk is the mechanic that gives it teeth

### 3. Quest events: fire-and-forget BUSD signals

`runner.go` loops `result.QuestEvents` and emits one `game.quest.event` signal per event with the event type and payload. Initial consumers are just the log handler. Plugins can subscribe independently. No game-engine logic in the core runner.

### 4. MUD ↔ game XP bridge: pack-defined XP table, signal handler bridge

The pack YAML gains a `mud_xp_events` map: `mud_signal_topic → xp_amount`. A new signal handler in `signal_handlers.go` subscribes to each topic in the map and calls the existing `ComputeXP` path with a synthetic `TokenUsage` (zeroed tokens, `SourceMUD` provider) so the XP multiplier chain still applies. This keeps the pack as the single source of truth for XP weights.

**Alternatives considered:**
- Hardcode XP values in signal_handlers.go — rejected, pack should own all weights
- Give MUD its own XP engine — rejected, duplicates scoring logic

### 5. Bounty Boards: tuner extends analysis to produce contracts; contracts stored in pack state

The tuner's Ollama analysis prompt gains a section asking for 3-5 bounty contracts (objective + xp_reward + room_id). Contracts are written into the evolved pack YAML under a `bounty_contracts` key. The MUD reads this file on startup (or via a BUSD refresh signal) and attaches contracts to their target rooms. Completion fires `game.bounty.completed` → signal handler awards XP burst.

### 6. ICE Encounter mechanics: streak penalty on timeout or explicit loss

ICE encounter is a short-lived TUI sub-mode triggered by the `game.ice.encountered` signal handler. The handler writes a pending encounter to a SQLite `ice_encounters` table with a deadline timestamp. `glitch game ice` reads the pending encounter and presents a simple "fight or jack out" choice. Timeout = loss. Loss decrements streak by 1 (not to zero). Win = clear the encounter, no effect.

The deadline duration is controlled by `ice_encounter.timeout_hours` in the pack (default 24h). The MUD plugin is not required for ICE encounters — they are triggered by pipeline runs, not MUD activity. The 24h default accommodates unattended or CI-adjacent runs where the player is not at the terminal. When the MUD plugin is active and a session is live, operators can set `timeout_hours: 1` (or shorter) for more pressure.

### 7. Reputation Decay: daily cron-style check on `glitch` startup

On each `glitch` startup, the store checks the last pipeline run timestamp. For each full day with no run, MUD reputation for all factions decays by a configurable `decay_per_day` value (default 2). Decay is capped at a floor (default 10, to prevent total lockout). This requires a `last_run_at` column (already derivable from run history) and the decay rate in the pack.

### 8. Recap: `glitch game recap` calls Ollama with last N days of stats

Fetch last N days of `game_runs` from SQLite (shell step), format as a JSON summary, pass to Ollama with the pack's narrator prompt asking for a story-arc narrative. Output is printed to stdout — no persistence. Uses the same Ollama client as the existing narration path.

### 9. Leaderboard: `game_personal_bests` table, `glitch game top` renders it

A single SQLite table updated after every scored run: one row per metric (fastest_run_ms, highest_xp, longest_streak, most_cache_tokens, lowest_cost_usd). `glitch game top` reads and renders with lipgloss-styled output.

## Risks / Trade-offs

- **ICE encounter UX friction** → Mitigation: auto-resolve timeout is configurable; default 30s is enough to notice without blocking work
- **Bounty contract staleness** (tuner runs infrequently) → Mitigation: contracts have a `valid_until` timestamp; expired contracts are silently cleared from rooms
- **Reputation decay on first startup after a vacation** → Mitigation: decay is capped per-startup check (max 7 days of decay applied at once, configurable)
- **MUD XP bridge double-counting** (if MUD emits multiple signals per action) → Mitigation: each bridge signal handler enforces a per-run dedup key in SQLite before awarding XP
- **Ollama model availability for recap** → Mitigation: graceful degradation — if Ollama is unreachable, print the raw stats table instead of a narrative

## Migration Plan

1. Apply SQLite schema migrations (achievements table, personal_bests table, ice_encounters table, last_run_at column) on first startup — existing `store.go` migration pattern handles this
2. Deploy all changes atomically — no staged rollout needed, all changes are additive
3. Rollback: revert the binary; old schema columns are ignored by old code, achievements/bests tables are additive-only

## Open Questions

- Should ICE encounter loss also penalize XP (in addition to streak)? — defaulting to streak-only for now, easier to tune up than down
- Should `game recap` output be copyable to a file (--output flag)? — out of scope for now, can add later
