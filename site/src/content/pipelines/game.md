---
title: "Game System"
description: "Every pipeline you run earns XP, builds streaks, and may trigger an ICE encounter. Here's how it works."
order: 11
---

gl1tch turns your pipeline runs into a persistent game. The more you run, the more XP you earn. Build streaks. Hit thresholds. Unlock achievements. Occasionally, ICE detects you.

## How Scoring Works

Every time a pipeline completes, gl1tch scores the run. The score is based on:

- **Speed** — faster runs score higher
- **Cache efficiency** — hits on cached tokens add a bonus
- **Streak multiplier** — consecutive days of runs stack your multiplier
- **Retry penalty** — steps that fail and retry cost points
- **Provider weights** — some providers score differently based on the active world pack

You don't configure any of this. It happens automatically in the background.

## XP, Levels, and Streaks

Each scored run adds XP to your total. XP accumulates into levels. Your streak counts consecutive calendar days with at least one run.

Check your standing with `glitch game top`:

```bash
glitch game top
```

```
── Personal Bests ──────────────────────────────
  METRIC                    VALUE           DATE
  ──────────────────────────────────────────────
  Fastest Run               420ms           2026-03-15
  Highest XP (single run)   1840 XP         2026-03-22
  Longest Streak            14 days         2026-03-01
  Most Cache Tokens         84200           2026-03-18
  Lowest Cost (non-zero)    $0.000012       2026-03-20
```

## Recap

`glitch game recap` generates a short cyberpunk story arc narrating your last N days of runs. It uses your stats — runs, XP, achievements, streak — as the source material and asks your local Ollama model to narrate it.

```bash
glitch game recap
glitch game recap --days 14
```

No Ollama? It falls back to a plain stats table.

## ICE Encounters

When you cross certain thresholds or trip specific conditions, ICE spawns. You have until the deadline to resolve it.

Check for an active encounter:

```bash
glitch game ice
```

```
[ICE DETECTED] Black ICE — Class IV
Deadline: 2026-04-04T08:00:00Z

  [1] fight  — attempt to defeat the ICE
  [2] jack-out — disconnect safely (loss)

Choice:
```

**Fight** to keep your streak intact. **Jack out** to disengage safely — but your streak takes a hit.

## The World Pack

gl1tch maintains a "world pack" — a set of scoring weights and narrative rules that shape how the game feels. Over time, gl1tch analyzes your run patterns and tunes the world pack to fit how you actually work.

To trigger the tuner manually:

```bash
glitch game tune
```

The tuner calls Ollama with your last 30 days of stats, generates an evolved world pack, and writes it to `~/.local/share/glitch/agents/game-world-tuned.agent.md`. The next run picks it up automatically.

```
Running game tuner...
  Stats window: last 30 days, 47 runs
  Current pack: neon-sprawl-v2

Game pack evolved in 8s
  Old pack: neon-sprawl-v2
  New pack: neon-sprawl-v3

Weight changes:
  speed_bonus_scale: 1.2000 → 1.4500 (+0.2500)
  streak_multiplier: 1.5000 → 1.7000 (+0.2000)

Game rules: 42 lines (unchanged length)

Pack written to ~/.local/share/glitch/agents/game-world-tuned.agent.md
```

The tuner also runs automatically after runs that cross score thresholds — you don't need to trigger it manually unless you want to force a refresh.

## See Also

- [CLI Reference](/docs/pipelines/cli-reference#glitch-game) — `glitch game` subcommands
- [Architecture](/docs/pipelines/architecture) — how run data is stored
