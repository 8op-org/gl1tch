## ADDED Requirements

### Requirement: Unsolicited narration is rate-limited per run
The deck SHALL allow at most 1 unsolicited narration delivery per completed run. If 2 or more runs complete within a 60-second window (busy mode), all unsolicited narration SHALL be suppressed until the window expires. The run completion window resets automatically after 60 seconds of inactivity.

#### Scenario: First narration after a run is allowed
- **WHEN** one run has completed in the last 60 seconds and conversation is not active
- **THEN** a `glitchNarrationMsg` is delivered normally

#### Scenario: Busy mode suppresses narration
- **WHEN** 2 or more runs have completed in the last 60 seconds
- **THEN** any `glitchNarrationMsg` is dropped regardless of conversation state

#### Scenario: Run window resets after 60 seconds
- **WHEN** 61 seconds have elapsed since the run window started
- **THEN** `recentRunCount` resets and narration is no longer suppressed by busy mode

#### Scenario: Run completions increment the counter
- **WHEN** a `topics.RunCompleted` or `topics.RunFailed` bus event arrives
- **THEN** `recentRunCount` is incremented (and window started if not already running)

### Requirement: Rate limit applies only to unsolicited narration
Direct GL1TCH replies to user messages SHALL NOT be subject to the rate limit. Only `glitchNarrationMsg` and `glitchRunEventMsg` deliveries that originate from background events (game scoring, pipeline completion) are gated.

#### Scenario: User prompt reply unaffected by busy mode
- **WHEN** 3 runs complete in rapid succession and the user types a message
- **THEN** gl1tch responds to the user message normally
