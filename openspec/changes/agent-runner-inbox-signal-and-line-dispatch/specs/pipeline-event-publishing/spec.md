## ADDED Requirements

### Requirement: Agent run topic constants defined in topics package
The `internal/busd/topics` package SHALL define constants for agent run lifecycle topics: `AgentRunStarted = "agent.run.started"`, `AgentRunCompleted = "agent.run.completed"`, `AgentRunFailed = "agent.run.failed"`.

#### Scenario: Agent topic constants available for publisher and subscriber
- **WHEN** the switchboard publishes agent run events
- **THEN** it uses the `topics.AgentRunStarted`, `topics.AgentRunCompleted`, `topics.AgentRunFailed` constants
- **AND** the inbox subscribes using the same package constants or the `"agent.run.*"` wildcard
