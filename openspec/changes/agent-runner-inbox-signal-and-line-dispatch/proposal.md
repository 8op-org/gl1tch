## Why

Agent runs launched from the switchboard complete silently — the inbox never receives a notification because no bus events are published when an agent job finishes. Additionally, the activity feed's line-marking and agent-dispatch flow needs to be wired up so marked output lines are injected as the prompt for a new agent run, completing the feedback loop from observed output to new action.

## What Changes

- Agent run lifecycle (start, complete, fail) publishes busd events so the inbox receives real-time notifications
- New busd topic constants: `agent.run.started`, `agent.run.completed`, `agent.run.failed`
- Inbox subscribes to `"agent.run.*"` alongside existing `"pipeline.run.*"`
- Store records agent runs so inbox list is populated on restart
- Activity feed line-marking (`m`) and agent dispatch (`r`) creates a new agent run with marked lines pre-injected into the prompt textarea

## Capabilities

### New Capabilities

- `agent-run-bus-events`: Agent run lifecycle events emitted to busd when an agent job starts, completes, or fails — enabling inbox and any future subscriber to react in real-time
- `activity-feed-line-dispatch`: Marked lines in the activity feed can be dispatched to a new agent run with the line content pre-filled as the prompt, completing the observe → act loop

### Modified Capabilities

- `pipeline-event-publishing`: Extend event publishing pattern to cover agent runs, not just pipeline runs

## Impact

- `internal/busd/topics/topics.go` — new `AgentRun*` topic constants
- `internal/switchboard/switchboard.go` — publish bus events on `jobDoneMsg`/`jobFailedMsg`; wire up `r` key to open agent modal with marked content injected
- `internal/inbox/model.go` — subscribe to `"agent.run.*"`, refresh on agent completion events
- `internal/store/` — ensure agent runs are recorded so inbox persists across restarts
