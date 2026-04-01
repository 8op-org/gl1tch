## Why

GLITCH slash commands are invisible to new users — you have to already know `/help` exists to discover them. An inline autocomplete dropdown, triggered as soon as the user types `/`, makes the command surface instantly explorable without breaking the keyboard-driven flow.

## What Changes

- Typing `/` in the GLITCH chat input opens an autocomplete suggestion list inline above the input
- Continued typing filters the list by fuzzy-matching against command names and descriptions
- `Tab`, `↑`/`↓` navigate the list; `Enter` or `Tab` selects and inserts the completion
- `Esc` dismisses the list and restores normal input behaviour
- Commands that accept arguments (e.g. `/model <name>`, `/cwd <path>`) show a short usage hint in the suggestion row
- The suggestion list is dismissed automatically when the typed text no longer starts with `/`

## Capabilities

### New Capabilities

- `slash-command-autocomplete`: Inline dropdown suggestion list for slash commands, triggered by `/` prefix in the GLITCH chat input; supports fuzzy filtering, keyboard navigation, and selection-to-insert behaviour

### Modified Capabilities

<!-- none -->

## Impact

- `internal/switchboard/glitch_panel.go` — primary change site: add autocomplete state to `glitchChatPanel`, intercept keystrokes before the existing slash-command switch, render suggestion overlay
- `internal/modal/fuzzypicker.go` — reused for scoring logic (no modifications required, only imported)
- No new external dependencies; no API changes; no breaking changes to existing slash command behaviour
