## ADDED Requirements

### Requirement: Feedback loop panel clears on first send
The feedback loop panel in both builders SHALL clear its content and display a loading indicator when the user sends their first prompt/pipeline run request.

#### Scenario: First send clears the draft
- **WHEN** the user submits a prompt for the first time in a session
- **THEN** the feedback loop panel clears its draft content and shows a loading/waiting state

#### Scenario: Subsequent sends append to history
- **WHEN** the user submits a follow-up prompt after the first response has arrived
- **THEN** the new exchange is appended below the previous response in the feedback loop panel

### Requirement: Feedback loop panel shows streaming response
The feedback loop panel SHALL display the agent's response as it streams in, line by line.

#### Scenario: Streaming output appears incrementally
- **WHEN** the agent runner emits output lines
- **THEN** each line appears in the feedback loop panel in order as it arrives

### Requirement: ctrl+s saves from feedback loop context
When the user presses `ctrl+s` from any focus position in either builder, the current name and content SHALL be saved.

#### Scenario: Save works regardless of focus
- **WHEN** the user presses ctrl+s while the chat input, feedback loop, or sidebar has focus
- **THEN** the current draft is saved without requiring focus to move to a specific field

### Requirement: ctrl+r reinjects from feedback loop
When the user presses `ctrl+r`, the first prompt that was sent in the current session SHALL be re-injected into the test runner to start a fresh run.

#### Scenario: ctrl+r reruns with original prompt
- **WHEN** the user presses ctrl+r after at least one prompt has been sent
- **THEN** the RunnerPanel clears and a new test run begins with the first prompt content
