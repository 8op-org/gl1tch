# Global Glitch Skill

**Date:** 2026-04-14
**Status:** Approved

## Problem

Glitch knowledge is split across two skills (`glitch-workflow`, `gl1tch-install`) and implicit project context. When working in other repos, agents don't know how to invoke glitch, what commands exist, or how to author workflows. There's no single reference that covers the full CLI.

## Solution

A single comprehensive skill at `~/.claude/skills/glitch/SKILL.md` that replaces both `glitch-workflow` and `gl1tch-install`. Covers: what glitch is, installation, CLI commands, ask routing, workflow authoring (s-expr + YAML), providers, batch runs, workspace model, observer/ES, research loop, and plugin system.

## Design

### Scope

The skill is a **reference document**, not a tutorial. It tells agents everything they need to:

1. **Use glitch** — run commands, query data, review PRs from any repo
2. **Author workflows** — create .glitch files following shell-first/LLM-last pattern
3. **Configure providers** — set up Ollama, Claude, OpenRouter, tiered escalation
4. **Run batch comparisons** — multi-variant evaluation runs
5. **Understand the workspace model** — `--workspace` flag for cross-repo work

### Structure

1. Skill metadata (broad trigger for any glitch mention)
2. What gl1tch is (one paragraph)
3. Installation (brew tap, shadow removal)
4. CLI command reference table
5. Ask routing (fast-path patterns, fallback chain)
6. Workspace model (`--workspace`, result directory structure, README.md rollup)
7. Workflow authoring (s-expr preferred, YAML legacy, template expressions, cardinal rule)
8. Workflow patterns (5 canonical patterns)
9. Provider & model reference (ollama, claude, copilot, openrouter, tiers)
10. Batch comparison runs (concept, naming, script pattern, results)
11. Observer & Elasticsearch (up/down, observe, index, indices)
12. Research loop (tool-use, goals, when it triggers)
13. Plugin system (naming convention, GoReleaser, brew tap)
14. Project reference (repo, packages, config paths)

### Replaces

- `~/.claude/skills/glitch-workflow/SKILL.md` — absorbed fully
- `~/.claude/skills/gl1tch-install/SKILL.md` — absorbed fully

Both should be deleted after the new skill ships.

### Trigger Description

Broad: fires on any mention of glitch, workflows, pipelines, batch runs, `glitch ask`, PR review via glitch, observer queries, or glitch installation.

## What does NOT change

- No code changes to glitch itself
- No changes to workflow file format
- No changes to provider configuration
