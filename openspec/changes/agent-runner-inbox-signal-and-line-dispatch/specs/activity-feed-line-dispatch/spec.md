## ADDED Requirements

### Requirement: Marked lines dispatched to agent runner with prompt pre-filled
When lines are marked in the activity feed and the user presses `r`, the agent runner modal SHALL open with the marked line content pre-filled in the prompt textarea. The user SHALL be able to edit the prompt before submitting.

#### Scenario: Pressing r with marked lines opens agent modal with content
- **WHEN** one or more lines are marked in the activity feed (via `m` key)
- **AND** the user presses `r`
- **THEN** the agent runner modal opens with the concatenated marked line content pre-filled in the prompt textarea

#### Scenario: Pressing r with no marked lines opens empty agent modal
- **WHEN** no lines are marked in the activity feed
- **AND** the user presses `r`
- **THEN** the agent runner modal opens with an empty prompt textarea

#### Scenario: Marks are cleared after dispatch
- **WHEN** the user submits the agent modal after dispatching marked lines
- **THEN** all line marks in the feed are cleared

### Requirement: Marked lines are visually distinct in the feed
Lines marked with `m` SHALL be visually highlighted in the activity feed so the user can see which lines will be dispatched. The cursor line and marked lines SHALL use distinct highlight styles.

#### Scenario: Marked line shown with highlight
- **WHEN** a line is marked with `m`
- **THEN** the line is rendered with a distinct background or foreground color in the feed

#### Scenario: Unmarking removes highlight
- **WHEN** `m` is pressed again on an already-marked line
- **THEN** the highlight is removed and the line is no longer included in dispatch
