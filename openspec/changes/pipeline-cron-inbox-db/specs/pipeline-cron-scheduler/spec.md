## ADDED Requirements

### Requirement: `orcai cron` subcommand with start/stop/list/logs commands
The `orcai` binary SHALL expose a `cron` subcommand with four child commands: `start`, `stop`, `list`, and `logs`. `cron start` SHALL launch a named tmux session (`orcai-cron`) and exec `orcai cron run` inside it. `cron stop` SHALL kill the `orcai-cron` session. `cron list` SHALL print all scheduled entries from `~/.config/orcai/cron.yaml` with their next-fire time. `cron logs` SHALL tail the cron log file at `~/.local/share/orcai/cron.log`.

#### Scenario: start creates tmux session
- **WHEN** `orcai cron start` is called and no `orcai-cron` tmux session exists
- **THEN** a new tmux session named `orcai-cron` is created and `orcai cron run` is executing inside it

#### Scenario: start fails if session exists
- **WHEN** `orcai cron start` is called and an `orcai-cron` session already exists
- **THEN** the command exits with a non-zero status and prints an error suggesting `--force`

#### Scenario: start --force replaces existing session
- **WHEN** `orcai cron start --force` is called and an `orcai-cron` session exists
- **THEN** the existing session is killed and a new one is started

#### Scenario: stop kills the daemon session
- **WHEN** `orcai cron stop` is called and an `orcai-cron` session exists
- **THEN** the tmux session is killed and a confirmation message is printed

#### Scenario: list shows entries with next-fire time
- **WHEN** `orcai cron list` is called with a valid `cron.yaml`
- **THEN** each entry is printed with its name, cron expression, pipeline/agent target, and calculated next-fire time

### Requirement: Schedule configuration in `~/.config/orcai/cron.yaml`
The cron daemon SHALL load schedule entries from `~/.config/orcai/cron.yaml`. Each entry SHALL have a `name` (unique identifier), `schedule` (5-field cron expression), `kind` (`pipeline` or `agent`), `target` (pipeline file path or agent name), and optional `args` map. The daemon SHALL watch the file with `fsnotify` and reload entries on change without restarting.

#### Scenario: Valid cron.yaml is loaded
- **WHEN** `cron.yaml` contains a well-formed entry with a valid cron expression
- **THEN** the scheduler registers the entry and it fires at the next scheduled time

#### Scenario: Invalid cron expression is rejected at load
- **WHEN** `cron.yaml` contains an entry with a malformed cron expression
- **THEN** the entry is skipped, an error is logged via Charm `log`, and valid entries continue running

#### Scenario: File change triggers hot-reload
- **WHEN** `cron.yaml` is modified while the daemon is running
- **THEN** the scheduler removes all existing entries and re-registers from the updated file within 2 seconds

#### Scenario: Missing cron.yaml starts empty
- **WHEN** `orcai cron run` is started and `~/.config/orcai/cron.yaml` does not exist
- **THEN** the daemon starts with no entries and logs a warning; it begins watching for the file to be created

### Requirement: Cron daemon uses Charm `log` for structured output
The `orcai cron run` process SHALL use `github.com/charmbracelet/log` for all log output. Log lines SHALL include `time`, `level`, `msg`, and relevant fields (e.g., `name`, `target`, `duration`, `error`). Output SHALL be written to both stderr and a rotating log file at `~/.local/share/orcai/cron.log`.

#### Scenario: Successful run is logged
- **WHEN** a scheduled pipeline completes successfully
- **THEN** a log line at INFO level is emitted with `name`, `target`, `duration`, and `exit_status=0`

#### Scenario: Failed run is logged at error level
- **WHEN** a scheduled pipeline exits with a non-zero status or returns an error
- **THEN** a log line at ERROR level is emitted with `name`, `target`, `error`, and `exit_status`

### Requirement: Concurrent scheduled runs do not block each other
The scheduler SHALL execute each fired entry in its own goroutine. A slow or hung run SHALL NOT delay other scheduled entries from firing. Each goroutine SHALL have an independent context with a configurable timeout (default: none; set via `timeout` field in the entry).

#### Scenario: Two entries fire at the same time
- **WHEN** two entries have the same cron schedule and both fire simultaneously
- **THEN** both execute concurrently without either blocking the other

#### Scenario: Entry with timeout is cancelled
- **WHEN** an entry has `timeout: 5m` and the run exceeds 5 minutes
- **THEN** the run's context is cancelled, the run is terminated, and an ERROR log line is emitted with `reason=timeout`
