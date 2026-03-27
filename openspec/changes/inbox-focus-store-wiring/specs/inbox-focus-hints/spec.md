## ADDED Requirements

### Requirement: inbox bottom-bar hints
When the inbox panel has focus, `viewBottomBar` MUST display inbox-specific key hints.

#### Scenario: inbox focused
- **WHEN** `m.inboxFocused == true`
- **THEN** bottom bar shows: `enter` open · `d` delete · `r` re-run · `tab` focus · `i` inbox

#### Scenario: inbox not focused
- **WHEN** `m.inboxFocused == false`
- **THEN** bottom bar is unchanged (existing launcher/signal/feed hints apply)
