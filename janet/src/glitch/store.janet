# SQLite run/step recording.
# Schema identical to Go version — existing .db files are compatible.

(import sqlite3 :as sql)

(defn- mkdir-p [dir]
  "Create directory and all parents. No-op if exists."
  (when (and dir (not= dir "") (not (os/stat dir)))
    (def parts (string/find-all "/" dir))
    (when (not (empty? parts))
      (mkdir-p (string/slice dir 0 (last parts))))
    (os/mkdir dir)))

(def- schema-stmts
  [`CREATE TABLE IF NOT EXISTS runs (
     id              INTEGER PRIMARY KEY AUTOINCREMENT,
     kind            TEXT NOT NULL DEFAULT 'workflow',
     name            TEXT NOT NULL,
     input           TEXT,
     output          TEXT,
     exit_status     INTEGER,
     started_at      INTEGER NOT NULL,
     finished_at     INTEGER,
     metadata        TEXT,
     workflow_file   TEXT,
     repo            TEXT,
     model           TEXT,
     tokens_in       INTEGER,
     tokens_out      INTEGER,
     cost_usd        REAL,
     variant         TEXT,
     workspace       TEXT NOT NULL DEFAULT '',
     parent_run_id   INTEGER REFERENCES runs(id),
     workflow_name   TEXT
   )`
   `CREATE INDEX IF NOT EXISTS idx_runs_parent ON runs(parent_run_id)`
   `CREATE TABLE IF NOT EXISTS steps (
     id          INTEGER PRIMARY KEY AUTOINCREMENT,
     run_id      INTEGER NOT NULL,
     step_id     TEXT NOT NULL,
     prompt      TEXT,
     output      TEXT,
     model       TEXT,
     duration_ms INTEGER,
     kind        TEXT,
     exit_status INTEGER,
     tokens_in   INTEGER,
     tokens_out  INTEGER,
     gate_passed INTEGER,
     artifacts   TEXT,
     UNIQUE(run_id, step_id)
   )`
   `CREATE TABLE IF NOT EXISTS research_events (
     id              INTEGER PRIMARY KEY AUTOINCREMENT,
     query_id        TEXT NOT NULL,
     question        TEXT NOT NULL,
     researchers     TEXT NOT NULL,
     composite_score REAL,
     reason          TEXT,
     created_at      INTEGER NOT NULL
   )`])

(defn open [&opt path]
  "Open or create a glitch database."
  (default path (string (os/getenv "HOME")
                        "/.local/share/glitch/glitch.db"))
  (def parts (string/find-all "/" path))
  (when (not (empty? parts))
    (def idx (last parts))
    (def dir (string/slice path 0 idx))
    (when (and (not= dir "") (not (os/stat dir)))
      (mkdir-p dir)))
  (def db (sql/open path))
  (each stmt schema-stmts
    (sql/eval db stmt))
  db)

(defn open-for-workspace [ws-path]
  "Open a workspace-scoped database."
  (open (string ws-path "/.glitch/glitch.db")))

(defn close [db]
  (sql/close db))

(defn record-run [db rec]
  "Insert a run record. Returns the row ID."
  (sql/eval db
    `INSERT INTO runs (name, input, workflow_file, model,
                       variant, workspace, parent_run_id,
                       workflow_name, started_at, kind)
     VALUES (:name, :input, :wf, :model,
             :variant, :workspace, :pid,
             :wfname, :now, 'workflow')`
    {:now (os/time)
     :name (rec :name)
     :input (rec :input)
     :wf (or (rec :workflow-file) nil)
     :model (or (rec :model) nil)
     :variant (or (rec :variant) nil)
     :workspace (or (rec :workspace) "")
     :pid (or (rec :parent-run-id) nil)
     :wfname (or (rec :workflow-name) nil)})
  (def rows (sql/eval db `SELECT last_insert_rowid() as id`))
  ((first rows) :id))

(defn finish-run [db id output exit-status &opt totals]
  "Update a run with final output."
  (default totals {})
  (sql/eval db
    `UPDATE runs SET output=:output, exit_status=:exit,
       tokens_in=:tin, tokens_out=:tout,
       cost_usd=:cost, finished_at=:now
     WHERE id=:id`
    {:id id :output output :exit exit-status
     :now (os/time)
     :tin (or (totals :tokens-in) 0)
     :tout (or (totals :tokens-out) 0)
     :cost (or (totals :cost) 0)}))

(defn record-step [db rec]
  "Insert or replace a step record."
  (sql/eval db
    `INSERT OR REPLACE INTO steps
       (run_id, step_id, prompt, output, model,
        duration_ms, kind, exit_status,
        tokens_in, tokens_out, gate_passed, artifacts)
     VALUES (:runid, :stepid, :prompt, :output, :model,
             :duration, :kind, :exit,
             :tin, :tout, :gate, :artifacts)`
    {:runid (rec :run-id)
     :stepid (rec :step-id)
     :prompt (or (rec :prompt) nil)
     :output (or (rec :output) nil)
     :model (or (rec :model) nil)
     :duration (or (rec :duration) nil)
     :kind (or (rec :kind) nil)
     :exit (or (rec :exit) nil)
     :tin (or (rec :tokens-in) nil)
     :tout (or (rec :tokens-out) nil)
     :gate (or (rec :gate) nil)
     :artifacts (or (rec :artifacts) nil)}))

(defn get-run [db id]
  "Fetch a single run by ID."
  (def rows (sql/eval db
    `SELECT * FROM runs WHERE id = :id` {:id id}))
  (when (> (length rows) 0) (first rows)))

(defn get-steps [db run-id]
  "Fetch all steps for a run."
  (sql/eval db
    `SELECT * FROM steps WHERE run_id = :rid ORDER BY id`
    {:rid run-id}))

(defn list-runs [db &named parent-id workflow limit]
  "List runs with optional filters."
  (default limit 50)
  (cond
    parent-id
      (sql/eval db
        `SELECT * FROM runs WHERE parent_run_id = :pid ORDER BY id DESC LIMIT :limit`
        {:pid parent-id :limit limit})
    workflow
      (sql/eval db
        `SELECT * FROM runs WHERE workflow_file = :wf ORDER BY id DESC LIMIT :limit`
        {:wf workflow :limit limit})
    (sql/eval db
      `SELECT * FROM runs ORDER BY id DESC LIMIT :limit`
      {:limit limit})))
