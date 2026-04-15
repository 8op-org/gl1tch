# Skill Creator Guide

Skills are small, focused playbooks for agent sessions. They turn “what I keep doing by hand” into a reusable, reviewable `SKILL.md`. A good skill captures intent, constraints, and the exact steps that make the result reliable.

The workflow is straightforward: do the work manually in an agent session, notice the parts you repeat, then formalize that routine with `skill-creator`. Don’t try to automate everything. Capture the stable pieces: the commands you always run, the checks you always apply, and the output format you always end up with.

The output is a production-ready `SKILL.md` you commit under `.github/skills/`. That’s the handoff point. Once it’s there, anyone on the team can reuse it without re-discovering your conventions from chat history.

This guide is organized as five concrete use cases. Each one maps to a core design pattern. Pick the pattern that matches your workflow, follow the example, then adapt it to your repo.

## The five core patterns

- **Tool Wrapper:** Encapsulate a CLI/API tool with an explicit intent → command map, plus prerequisite checks and failure handling.
- **Generator:** Produce a standardized artifact from a template or schema, using a stepwise workflow to gather required inputs.
- **Reviewer:** Evaluate (and optionally edit) input against an explicit checklist of quality criteria.
- **Inversion:** Run a mandatory clarification interview first, then act only after requirements are concrete.
- **Pipeline:** Orchestrate a multi-phase workflow with sequential gates; each stage must complete before the next begins.

```mermaid
flowchart LR
  A[Manual Work] --> B[Agent Session]
  B --> C[skill-creator]
  C --> D[SKILL.md]
  D --> E[Team Reuse]
```

## How to Use This Guide

1. Start with the use case closest to your repetitive workflow.
2. Copy the structure. Replace the domain details (commands, checks, artifacts).
3. Write constraints like you mean them. “Must” and “must not” are the skill’s guardrails.
4. Design for handoff. Assume the next developer has zero context.
5. Commit the `SKILL.md` to `.github/skills/` and iterate through review.

## Use cases (one per pattern)

- [01-wrapper-curl.md](./01-wrapper-curl.md)
- [02-generator-mermaid.md](./02-generator-mermaid.md)
- [03-reviewer-kiss.md](./03-reviewer-kiss.md)
- [04-inversion-invitations.md](./04-inversion-invitations.md)
- [05-pipeline-pr-checks.md](./05-pipeline-pr-checks.md)
