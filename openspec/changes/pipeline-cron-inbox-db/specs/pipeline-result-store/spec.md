## ADDED Requirements

### Requirement: SQLite result store captures every pipeline and agent run
The `internal/store` package SHALL open a SQLite database at `~/.local/share/orcai/orcai.db` using `modernc.org/sqlite` (no CGo). It SHALL apply a schema migration on first open. Each completed pipeline or agent run SHALL be written as a row in the `runs` table with: `id`, `kind` (`pipeline`|`agent`), `name`, `started_at` (unix millis), `finished_at` (unix millis), `exit_status` (0=ok, non-zero=error, NULL=in-flight), `stdout`, `stderr`, and `metadata` (JSON blob).

#### Scenario: Run is recorded on completion
- **WHEN** a pipeline run completes (success or failure)
- **THEN** a row is inserted into `runs` with correct `name`, `kind`, `started_at`, `finished_at`, `exit_status`, `stdout`, and `stderr`

#### Scenario: In-flight run has NULL exit_status
- **WHEN** a pipeline run has started but not yet completed
- **THEN** a row exists with `exit_status = NULL` and `finished_at = NULL`

#### Scenario: Database is created on first open
- **WHEN** `store.Open()` is called and `orcai.db` does not exist
- **THEN** the file is created and the schema is applied without error

#### Scenario: Concurrent writes are safe
- **WHEN** two goroutines call `store.RecordRun()` simultaneously
- **THEN** both rows are written correctly; no data corruption occurs

### Requirement: WAL mode and single-writer queue for concurrency safety
The store SHALL enable SQLite WAL mode (`PRAGMA journal_mode=WAL`) on open. All write operations SHALL be serialized through an internal channel-based queue processed by a single goroutine, preventing write contention while allowing concurrent reads.

#### Scenario: WAL mode is set on open
- **WHEN** `store.Open()` is called
- **THEN** `PRAGMA journal_mode` returns `wal`

#### Scenario: Read does not block while write is queued
- **WHEN** a write is pending in the queue
- **THEN** a concurrent `store.QueryRuns()` call returns results without waiting for the write to complete

### Requirement: `db` built-in step type for pipeline access
The pipeline runner SHALL recognize a built-in step type `db` before dispatching to the plugin manager. A `db` step SHALL support two operations via the `op` field: `query` (SQL SELECT) and `exec` (SQL INSERT/UPDATE/DELETE). The `sql` field holds the statement; parameters are interpolated from the ExecutionContext using the existing `{{path}}` syntax. Query results SHALL be serialized as JSON and stored in the ExecutionContext at `<step-id>.out`.

#### Scenario: db query step stores results in context
- **WHEN** a pipeline step has `type: db`, `op: query`, and `sql: "SELECT * FROM runs WHERE kind = 'pipeline' LIMIT 5"`
- **THEN** the step executes the query and sets `<step-id>.out` to a JSON array of matching rows in the ExecutionContext

#### Scenario: db exec step returns rows affected
- **WHEN** a pipeline step has `type: db`, `op: exec`, and a valid INSERT statement
- **THEN** the step executes the statement and sets `<step-id>.rows_affected` in the ExecutionContext

#### Scenario: db step with interpolated SQL
- **WHEN** a `db` step's `sql` field contains `{{param.name}}` and `param.name` is set in the ExecutionContext
- **THEN** the placeholder is replaced before execution

#### Scenario: db step with invalid SQL logs error and fails the pipeline
- **WHEN** a `db` step's `sql` field contains a syntax error
- **THEN** the step returns an error, the pipeline halts (respecting retry settings), and the error is recorded in the run row

### Requirement: Result store retention policy with configurable pruning
The store SHALL support an `AutoPrune` option configured via `~/.config/orcai/config.yaml` under `store.retention`. Pruning SHALL delete rows older than `max_age_days` (default: 30) or when the total row count exceeds `max_rows` (default: 1000), whichever limit is reached first. Pruning SHALL run automatically after each new run is recorded.

#### Scenario: Old rows are pruned after a new run
- **WHEN** a new run is recorded and there are rows older than `max_age_days`
- **THEN** the old rows are deleted

#### Scenario: Row count cap is enforced
- **WHEN** total row count exceeds `max_rows` after a new insert
- **THEN** the oldest rows are deleted until count equals `max_rows`

#### Scenario: Pruning disabled when retention not configured
- **WHEN** no `store.retention` block exists in `config.yaml`
- **THEN** pruning uses the default limits (30 days, 1000 rows)
