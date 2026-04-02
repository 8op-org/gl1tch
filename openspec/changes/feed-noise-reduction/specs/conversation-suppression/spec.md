## ADDED Requirements

### Requirement: Unsolicited narration is suppressed during active conversation
The deck SHALL suppress unsolicited `glitchNarrationMsg` and `glitchRunEventMsg` delivery when a conversation is active. A conversation is active when the GL1TCH panel is streaming a response, the intent router is running, OR the user submitted a message less than 30 seconds ago. Direct responses to user input SHALL never be suppressed.

#### Scenario: Narration dropped while streaming
- **WHEN** a `glitchNarrationMsg` arrives while `glitchChat.streaming == true`
- **THEN** the message is dropped and no text appears in the feed

#### Scenario: Narration dropped while routing
- **WHEN** a `glitchNarrationMsg` arrives while `glitchChat.routing == true`
- **THEN** the message is dropped and no text appears in the feed

#### Scenario: Narration dropped within silence window
- **WHEN** the user submitted a message 10 seconds ago and a `glitchNarrationMsg` arrives
- **THEN** the message is dropped (10s < 30s silence window)

#### Scenario: Narration delivered after silence window expires
- **WHEN** 31 seconds have elapsed since the last user message and no streaming or routing is active
- **THEN** a `glitchNarrationMsg` is delivered normally to the feed

#### Scenario: Direct reply never suppressed
- **WHEN** the user types a message and gl1tch streams a direct reply
- **THEN** the reply appears regardless of any suppression state

### Requirement: glitchChatPanel exposes active state
The `glitchChatPanel` type SHALL expose an `IsActive() bool` method returning `true` when `streaming || routing`.

#### Scenario: IsActive true during stream
- **WHEN** `glitchChatPanel.streaming == true`
- **THEN** `IsActive()` returns `true`

#### Scenario: IsActive false when idle
- **WHEN** both `streaming` and `routing` are `false`
- **THEN** `IsActive()` returns `false`
