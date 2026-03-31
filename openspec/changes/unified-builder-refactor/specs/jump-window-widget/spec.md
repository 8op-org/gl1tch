## ADDED Requirements

### Requirement: orcai widget jump-window is a standalone subcommand
The CLI SHALL expose `orcai widget jump-window` as a subcommand that launches the jump window TUI as a standalone alt-screen program.

#### Scenario: Subcommand opens jump window
- **WHEN** the user runs `orcai widget jump-window`
- **THEN** the terminal enters alt-screen and the jump window TUI is displayed in non-embedded mode

### Requirement: orcai widget is a command group
The CLI SHALL expose `orcai widget` as a cobra command group that namespaces reusable TUI widget subcommands.

#### Scenario: widget group lists subcommands
- **WHEN** the user runs `orcai widget --help`
- **THEN** the help output lists `jump-window` as an available subcommand
