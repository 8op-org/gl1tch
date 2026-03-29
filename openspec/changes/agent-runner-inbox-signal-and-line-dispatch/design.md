## Context

The switchboard runs agent jobs (via `CliAdapter.Execute()`) and tracks them as feed entries. When a job finishes, `jobDoneMsg`/`jobFailedMsg` is handled inside the switchboard's `Update()` loop — but nothing is published to the busd event bus. The inbox currently polls the store every 5 seconds and subscribes to `"pipeline.run.*"` bus events, but has no mechanism to learn about agent run completions in real time.

Line marking (`m` key) and cursor navigation are implemented in the feed. The `r` key is supposed to open the agent modal with marked content, but the prompt injection into the textarea needs to be verified and completed end-to-end.

## Goals / Non-Goals

**Goals:**
- Publish `agent.run.started`, `agent.run.completed`, `agent.run.failed` events to busd when the switchboard starts/finishes agent jobs
- Inbox subscribes to `"agent.run.*"` and refreshes in real time on completion
- Activity feed `r` key opens agent modal with marked line content pre-filled as the prompt

**Non-Goals:**
- Persisting agent runs to store (beyond what already exists)
- Retry or queueing logic for agent runs
- Changing the agent runner CLI adapter

## Decisions

### 1. Publish agent events from switchboard, not a separate runner

The agent job lifecycle lives entirely inside the switchboard `Update()` loop (via `jobDoneMsg`/`jobFailedMsg`). Publishing from the same site avoids introducing a second publisher and keeps event ordering consistent with feed mutations.

**Alternative considered**: Route through a dedicated `AgentRunner` struct that publishes. Rejected — adds indirection for a simple publish call.

### 2. Reuse busd client held by switchboard

The switchboard already holds a `busd.Client` for receiving pipeline events via `pipeline_bus.go`. Extend the same client to publish outbound agent events. A single `Publish(topic, payload)` call on `jobDoneMsg` / `jobFailedMsg` is sufficient.

**Alternative considered**: Open a new busd connection for publishing. Rejected — the switchboard's existing client is already connected and available.

### 3. New topic namespace `agent.run.*`

Use `agent.run.started`, `agent.run.completed`, `agent.run.failed` to mirror the existing `pipeline.run.*` pattern. This keeps topics discoverable and lets inbox use a wildcard subscription.

### 4. Inbox wildcard subscription

Inbox subscribes to `"agent.run.*"` with the same pattern used for `"pipeline.run.*"`. On any matching event, trigger a store refresh. This is the minimal change — no agent-specific message parsing needed.

### 5. Line dispatch: open modal with prefilled prompt

When `r` is pressed in the feed with marks active, open the agent runner modal and set the prompt textarea content to the concatenated marked lines. This mirrors how the signal board already injects context. The existing `openAgentModal(prompt)` path (if it exists) should be used; otherwise add a `promptPreset` field to the modal's init message.

## Risks / Trade-offs

- [Risk] Bus unavailable at publish time → Mitigation: publish is fire-and-forget; log error, do not affect job lifecycle
- [Risk] Agent run topic conflicts with future pipeline-agent integration → Mitigation: `agent.run.*` namespace is distinct from `pipeline.run.*`; easy to extend payload later
- [Risk] `r` key already bound to something else in feed context → Mitigation: confirm keybinding is free in feed-focused state; reassign if needed

## Open Questions

- Does the store already record agent runs with a `kind` field that distinguishes them from pipeline runs, or does inbox need a schema addition?
- Is `busd.Client.Publish()` already available on the switchboard, or does the client need a `Publish` method added?
