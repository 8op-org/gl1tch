## ADDED Requirements

### Requirement: Agent run lifecycle events published to busd
The switchboard SHALL publish a busd event when an agent job starts, completes, or fails. Topic constants SHALL follow the `agent.run.*` namespace. The payload SHALL be JSON with at minimum `run_id`, `agent`, and `started_at` (on start) or `exit_status` and `duration_ms` (on completion/failure).

#### Scenario: Start event published when agent job begins
- **WHEN** the switchboard begins executing an agent job
- **THEN** an event with topic `agent.run.started` and payload `{"run_id": "<id>", "agent": "<name>"}` is published to busd

#### Scenario: Completion event published when agent job succeeds
- **WHEN** a `jobDoneMsg` is received for an agent job
- **THEN** an event with topic `agent.run.completed` and payload including `run_id` and `exit_status: 0` is published to busd

#### Scenario: Failure event published when agent job fails
- **WHEN** a `jobFailedMsg` is received for an agent job
- **THEN** an event with topic `agent.run.failed` and payload including `run_id` and `exit_status` (non-zero) is published to busd

#### Scenario: Bus unavailable does not affect job lifecycle
- **WHEN** busd is not running and an agent job completes
- **THEN** the publish error is silently ignored and the feed entry is updated normally

### Requirement: Inbox subscribes to agent run events
The inbox SHALL subscribe to the `"agent.run.*"` wildcard topic and trigger a store refresh when any matching event is received, causing the inbox list to update in real time without waiting for the polling interval.

#### Scenario: Inbox refreshes on agent completion
- **WHEN** an `agent.run.completed` event is received on the bus
- **THEN** the inbox refreshes its run list from the store within one event loop tick

#### Scenario: Inbox refreshes on agent failure
- **WHEN** an `agent.run.failed` event is received on the bus
- **THEN** the inbox refreshes its run list from the store
