# Agent Intuition Layer Design

**Date:** 2026-04-22
**Status:** Draft
**Problem:** Glitch has strong execution infrastructure (MCP, DSL, REPL, code intel) but agents don't know *when* to use it. The machinery exists; the judgment doesn't.

## Overview

Hybrid approach: a **smart skill** inside the agent's context handles the "should I use glitch?" decision. A new **advisory MCP tool** handles "what specifically should I use?" The skill triggers the tool. The tool's intelligence lives in a glitch workflow, not hardcoded rules.

Three components:
1. `glitch_advise` MCP tool -- advisory, never executes
2. Enhanced skill layer -- trigger conditions for when to consider glitch
3. Learning loop -- sessions and recommendations feed the workflow index

## 1. Advisory Tool: `glitch_advise`

New MCP tool. 8th tool in the surface.

### Interface

**Input:**
```json
{
  "task": "string, required -- natural language task description",
  "context": "string, optional -- repo, files, domain context"
}
```

**Output:**
```json
{
  "approach": "workflow | primitive | repl | none",
  "primitives": ["list of recommended DSL primitives"],
  "reasoning": "why this approach fits the task",
  "example": "concrete DSL snippet showing usage",
  "existing_workflows": ["paths to matching workflows from the index"]
}
```

### Implementation

`glitch_advise` calls an internal advisory workflow (`.glitch/workflows/advise.glitch`) that:

1. Searches the workflow index for task-shape matches (keyword + fuzzy)
2. Sends the task description + primitive catalog + matching workflows to the LLM
3. LLM returns structured recommendation
4. Handler formats and returns via MCP

The advisory workflow knows the full primitive catalog: `step`, `sh`, `llm`, `par`, `phase`, `gate`, `consensus`, `grounded?`, `investigate`, `validate`, `call-workflow`, `retry`, `with-timeout`, `json-extract`, `save`, `read-file`.

The LLM does the matching. No hardcoded rules, no keyword tables, no severity maps. Consistent with AI-first principle.

### What it doesn't do

- Never executes workflows or primitives
- Never modifies files or state
- Returns recommendation only; agent decides whether to follow it

## 2. Smart Skill Layer

Lives inside the agent's context as a skill/prompt. Replaces the "syntax reference" approach with a "decision support" approach.

### Trigger Conditions

The skill activates when the agent notices one of these signals in the task:

| Signal | Pattern | Example |
|--------|---------|---------|
| Repetition | Task will be done again or across multiple targets | "check all PRs", "audit these repos" |
| Confidence | Judgment where being wrong matters | "is this accurate?", "which is better?" |
| Multi-source | Needs information composed from multiple providers/tools | "compare what Claude and Copilot say about this" |
| Investigation | Uncertain facts, contradictions, structured reasoning | "figure out why this is failing", "is this finding real?" |

### Behavior When Triggered

1. Call `glitch_advise` with the task description
2. Incorporate the recommendation into the agent's reasoning (not surfaced to user)
3. Agent decides: use it, ignore it, or adapt it

### What it doesn't do

- No mandatory overhead -- if the task clearly isn't a glitch task, skill stays quiet
- No workflow-first gate -- agents don't check glitch before every action
- No auto-execution -- recommendation only, agent stays in control
- Thin implementation -- ~30-40 lines of trigger logic + a `glitch_advise` call

### Skill scope

The skill is agent-specific. For Claude Code, it's a superpowers-style skill. For Copilot, it would be an agent instruction. For other agents, it's whatever prompt injection mechanism they support. The trigger conditions and `glitch_advise` call are the same regardless.

## 3. Learning Loop

Makes `glitch_advise` get smarter over time without embeddings or ML.

### Session Tagging

When an agent calls `glitch_advise`, the recommendation is recorded in the active session:

```clojure
{:type :advise
 :task "the task description"
 :recommendation {:approach "workflow" :primitives ["grounded?"] ...}
 :followed? nil}  ;; updated later based on what the agent does
```

The `:followed?` field is updated when the session ends or is promoted:
- Agent subsequently called `glitch_run` or `glitch_eval` with the recommended primitive -> `true`
- Agent completed the task without using glitch -> `false`

### Index Enrichment

When a session is promoted to a workflow via `(promote)`, the advisory context travels with it:

```clojure
;; workflow index entry (existing fields + new)
{:path ".glitch/workflows/check-pr-accuracy.glitch"
 :description "Check PR summary accuracy against diff"
 :promoted-from "session-2026-04-22-143022"
 :tags ["pr" "accuracy" "grounding"]
 :task-shape "verify factual accuracy of text against source material"}  ;; NEW
```

The `:task-shape` field is a generalized description of what kind of task triggered this workflow's creation. Future `glitch_advise` calls match on task-shape in addition to description and tags.

### Feedback Storage

Stored in session EDN. No immediate use beyond informing future advisory workflow prompts. The advisory workflow can include recent feedback as context: "these recommendations were followed, these were ignored."

### What this doesn't do

- No embedding/vector store -- keyword + fuzzy match on EDN index
- No cross-machine sync -- local disk
- No automatic workflow creation -- agents explicitly `promote`
- No retraining -- feedback is context for the advisory LLM, not training data

## Architecture Summary

```
Agent Context (Inside)          MCP Server (Outside)
========================        ========================
                                
  Smart Skill                     glitch_advise
  - trigger detection     --->    - advisory workflow
  - calls glitch_advise           - index search
  - incorporates result           - LLM recommendation
                                  - session recording
                                
                                  glitch_run / glitch_eval
                                  - execution (existing)
                                  - session recording (existing)
                                
                                  Workflow Index (EDN)
                                  - descriptions
                                  - tags
                                  - task-shapes (NEW)
                                  - feedback signals (NEW)
```

## Constraints

- Advisory workflow uses the same provider/tiering as all other glitch workflows
- No new dependencies (no vector DB, no new services)
- Skill layer is per-agent but advisory tool is agent-agnostic
- `glitch_advise` is read-only; never mutates workflows or files
- Session recording already exists; this extends it, doesn't replace it

## Out of Scope

- Automatic workflow creation (agents must explicitly promote)
- Cross-machine workflow sync
- Embedding-based search (stays keyword + fuzzy)
- Mandatory workflow-first planning gate
- Point 4 from brainstorming (glitch-aware planning step) -- deferred, could become annoying overhead
