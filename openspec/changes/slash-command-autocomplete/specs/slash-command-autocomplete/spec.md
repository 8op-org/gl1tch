## ADDED Requirements

### Requirement: Autocomplete overlay activates on slash prefix
The GLITCH chat input SHALL display an inline suggestion overlay whenever the input value starts with `/`.

#### Scenario: Overlay appears on slash keystroke
- **WHEN** the user types `/` into a focused, empty chat input
- **THEN** the suggestion overlay MUST appear above the input showing all available slash commands

#### Scenario: Overlay appears when input is set to slash-prefixed value
- **WHEN** the chat input value starts with `/`
- **THEN** `acActive` SHALL be `true` and `acSuggestions` SHALL be non-empty (assuming at least one command exists)

#### Scenario: Overlay dismisses when prefix is removed
- **WHEN** the user deletes characters until the input value no longer starts with `/`
- **THEN** the suggestion overlay SHALL be hidden

### Requirement: Suggestions are fuzzy-filtered in real time
The overlay SHALL filter and rank suggestions on every keystroke using the same fuzzy-scoring logic as the existing picker components.

#### Scenario: Full match ranks highest
- **WHEN** the user has typed `/model`
- **THEN** `/model` and `/models` SHALL both appear; `/model` (exact prefix) SHALL rank above `/models`

#### Scenario: Partial match filters list
- **WHEN** the user has typed `/th`
- **THEN** only commands whose name contains the substring (e.g. `/themes`) SHALL appear; non-matching commands SHALL be hidden

#### Scenario: No match hides overlay
- **WHEN** the user has typed `/zzz` (no matching command)
- **THEN** the suggestion overlay SHALL be hidden (or show a "no matches" state and not consume navigation keys)

### Requirement: Keyboard navigation selects suggestions
The user SHALL be able to navigate the suggestion list with keyboard keys without leaving the input field.

#### Scenario: Down arrow moves cursor down
- **WHEN** the suggestion overlay is active and the user presses the down-arrow key
- **THEN** the highlighted suggestion SHALL advance to the next item, wrapping to the first item after the last

#### Scenario: Up arrow moves cursor up
- **WHEN** the suggestion overlay is active and the user presses the up-arrow key
- **THEN** the highlighted suggestion SHALL move to the previous item, wrapping to the last item before the first

#### Scenario: Tab advances cursor (same as down arrow)
- **WHEN** the suggestion overlay is active and the user presses Tab
- **THEN** the highlighted suggestion SHALL advance exactly as it does for the down-arrow key

#### Scenario: Navigation keys are consumed by overlay
- **WHEN** the overlay is active and the user presses Tab, Up, or Down
- **THEN** those key events SHALL NOT be forwarded to the underlying textinput model

### Requirement: Enter or Tab on selected suggestion inserts command
The user SHALL be able to confirm a suggestion and have it inserted into the input.

#### Scenario: Enter confirms selected suggestion
- **WHEN** the suggestion overlay is active and the user presses Enter
- **THEN** the chat input value SHALL be set to the selected command followed by a single space, the cursor SHALL be positioned at the end, and the overlay SHALL be dismissed

#### Scenario: Tab confirms when cursor is on a suggestion
- **WHEN** the suggestion overlay is active, a suggestion is highlighted, and the user presses Tab
- **THEN** the same insertion behaviour as Enter SHALL apply

#### Scenario: Inserted command has trailing space
- **WHEN** any command is inserted via autocomplete
- **THEN** the input value SHALL be `"<command> "` (with one trailing space) so the user can immediately type arguments

### Requirement: Esc dismisses the overlay without changing input
The user SHALL be able to dismiss the suggestion overlay without selecting any command.

#### Scenario: Esc hides overlay and retains typed text
- **WHEN** the suggestion overlay is active and the user presses Esc
- **THEN** the overlay SHALL be hidden, the input value SHALL remain unchanged, and focus SHALL stay on the input

#### Scenario: Overlay does not reappear until next slash keypress after Esc
- **WHEN** the user has dismissed the overlay with Esc and the input still starts with `/`
- **THEN** the overlay SHALL remain hidden until the user modifies the input value

### Requirement: Suggestion rows display command name and usage hint
Each row in the suggestion overlay SHALL show the command name and a short usage/description hint.

#### Scenario: Command with no arguments shows description only
- **WHEN** `/help` appears in the suggestion list
- **THEN** its row SHALL render as `/help` with the hint `— this list` (or equivalent short description)

#### Scenario: Command with arguments shows usage hint
- **WHEN** `/model` appears in the suggestion list
- **THEN** its row SHALL render with a hint indicating it accepts a model name argument (e.g. `[name] — switch provider/model`)

#### Scenario: Hint text is truncated to fit panel width
- **WHEN** the panel width is too narrow to display the full hint
- **THEN** the hint text SHALL be truncated with an ellipsis rather than overflowing the panel
