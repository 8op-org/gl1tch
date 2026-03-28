## ADDED Requirements

### Requirement: Jump window renders as an in-process switchboard modal overlay
The switchboard TUI SHALL render the jump window as a modal overlay using the same overlay infrastructure as the help modal and theme picker. It SHALL NOT spawn a `tmux display-popup` or a separate orcai process.

#### Scenario: Jump modal opens on keypress
- **WHEN** the user presses the configured jump key inside the switchboard TUI
- **THEN** the jump window overlay appears over the switchboard, using the active theme palette and standard modal border

#### Scenario: Jump modal closes on escape or selection
- **WHEN** the user presses `esc` or selects a window in the jump modal
- **THEN** the modal is dismissed and the switchboard view is restored

#### Scenario: Jump modal inherits active theme
- **WHEN** the active theme is changed before or after opening the jump modal
- **THEN** the jump modal renders using the current active theme colors without requiring a restart

### Requirement: Bootstrap keybinding routes jump to switchboard window instead of spawning popup
The `^spc j` tmux keybinding SHALL send a key event to the switchboard window (e.g. `tmux send-keys -t orcai:0 J`) to trigger the in-process modal, rather than using `display-popup -E … orcai _jump`.

#### Scenario: Chord j triggers jump modal in switchboard
- **WHEN** the user presses `^spc j`
- **THEN** tmux sends the jump trigger key to the switchboard window, and the switchboard opens the jump modal overlay

### Requirement: `_jump` standalone dispatch is removed
The `_jump` case in `main.go` and the `jumpwindow.Run()` standalone entry SHALL be removed. The jump window SHALL only be accessible through the switchboard modal.

#### Scenario: `orcai _jump` is no longer a valid invocation
- **WHEN** the user runs `orcai _jump` directly
- **THEN** orcai exits with an error or falls through to the default cobra help
