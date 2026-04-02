## Why

The gl1tch game layer has significant dead code: achievements are evaluated by Ollama but thrown away, ICE classes and quest events are parsed but never consumed, and the MUD runs as a completely separate progression track with no XP or game tie-ins. Five partially-built systems need to be completed and five new mechanics added to make the game loop feel alive and worth playing every day.

## What Changes

- **Wire achievements**: stop discarding `EvaluateResult.Achievements`; persist to SQLite; surface unlock notifications in the TUI
- **Wire ICE classes**: emit `game.ice.encountered` BUSD signal when `ICEClass` is non-nil; display badge in score output
- **Wire quest events**: emit `game.quest.event` BUSD signal for each `QuestEvent` returned by Ollama
- **MUD ↔ game XP bridge**: translate `mud.room.entered`, `mud.espionage.talked`, and `mud.hack.success` signals into XP awards via the existing game scoring path
- **Achievement unlock notifications**: flash a styled notification in the TUI feed when achievements unlock
- **Bounty Boards**: tuner generates active contracts per MUD room; completing a contract awards an XP burst via a new `game.bounty.completed` signal
- **ICE Encounters**: non-nil `ICEClass` triggers a MUD-style combat event in the TUI; losing an ICE encounter breaks the active streak
- **Reputation Decay**: player MUD reputation decays by a configurable rate per day when no pipeline runs are recorded; resets on next run
- **Run Replay Narration**: `glitch game recap [--days N]` CLI subcommand; Ollama narrates the last N days (default 7) as a story arc
- **Personal Leaderboard**: `glitch game top` CLI subcommand; local SQLite table of personal bests (fastest run, highest XP, longest streak)

## Capabilities

### New Capabilities

- `game-achievement-persistence`: Store, query, and notify on achievement unlocks — SQLite schema, query path, and TUI notification
- `game-ice-signal`: Emit and handle ICE class signals — BUSD topic `game.ice.encountered`, score badge, ICE encounter TUI event
- `game-quest-signal`: Emit quest events from Ollama evaluation results onto the BUSD as `game.quest.event`
- `mud-xp-bridge`: Translate MUD BUSD signals into game XP awards using a configurable XP-per-event table in the pack
- `game-bounty-board`: Tuner-generated active contracts stored per MUD room; completion awards XP burst
- `game-ice-encounter`: ICE class triggers a TUI combat sequence; loss breaks streak
- `game-reputation-decay`: Daily decay of MUD reputation when no pipeline runs recorded; restored on next active run
- `game-recap`: `glitch game recap` CLI subcommand — Ollama story-arc narration of last N days
- `game-leaderboard`: `glitch game top` CLI subcommand — personal bests from local SQLite

### Modified Capabilities

- `pipeline-event-publishing`: `game.run.scored` payload must include `achievements []string` and `ice_class string` fields (currently hardcoded empty/nil)

## Impact

- `internal/game/engine.go` — new achievement unlock notification hook
- `internal/game/pack.go` — add `MUDXPEvents` table and `BountyContracts` to `PackWeights`
- `internal/game/tuner.go` — extend analysis/generate prompts to produce bounty contracts
- `internal/pipeline/runner.go` — stop hardcoding `Achievements: []string{}`; wire ICE and quest signal emission
- `internal/store/store.go` + `query.go` — achievements table, personal bests table, reputation decay column
- `internal/console/signal_handlers.go` — new handlers: `game.ice.encountered`, `game.quest.event`, MUD XP bridge handlers
- `cmd/game.go` — add `recap` and `top` subcommands
- `internal/game/packs/cyberspace/pack.yaml` — add `mud_xp_events` section and initial bounty contract templates
