## REMOVED Requirements

### Requirement: Core widgets registered as cobra subcommands
**Reason**: `sysop`, `picker`, and `welcome` are dead entrypoints — they all resolve to `switchboard.Run()` or a deprecated stub. With a single switchboard entrypoint (`orcai` with no args), standalone widget subcommands add noise to `--help` without value.
**Migration**: Use `orcai` with no arguments to launch the switchboard. `orcai picker` users: use the agent runner overlay inside the switchboard (press `a`).

### Requirement: Widget dispatch invokes core widgets via exec
**Reason**: The exec dispatch through `orcai sysop` / `orcai picker` / `orcai welcome` is only meaningful if those subcommands exist. Removing the subcommands removes the need for the dispatch indirection.
**Migration**: No user-facing migration required; this was internal plumbing.

### Requirement: Core subcommands accept bus socket path via flag
**Reason**: The `--bus-socket` flag was carried by `sysop`, `picker`, and `welcome`, all of which are being removed.
**Migration**: The switchboard connects to the bus automatically via the bootstrap path; no flag is needed.

## ADDED Requirements

### Requirement: `orcai` with no arguments is the sole switchboard entry point
The orcai binary SHALL launch the switchboard via `bootstrap.Run()` when invoked with no subcommand. No separate `sysop` or `welcome` cobra subcommand SHALL be registered.

#### Scenario: No-arg invocation starts switchboard
- **WHEN** the user runs `orcai` with no arguments
- **THEN** bootstrap runs and the switchboard session is created or reattached

#### Scenario: `orcai sysop` is no longer valid
- **WHEN** the user runs `orcai sysop`
- **THEN** cobra returns an "unknown command" error

#### Scenario: `orcai welcome` is no longer valid
- **WHEN** the user runs `orcai welcome`
- **THEN** cobra returns an "unknown command" error
