(ns glitch.store
  "SQLite run/step recording.
   Schema identical to Go/Janet versions — existing .db files are compatible."
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
            [clojure.string :as str]))

;; Lazy pod loading — only loads SQLite pod on first actual DB operation.
;; Prevents segfault on quick-exit commands (version, up, mcp) that never
;; touch the store.
(def ^:private pod-loaded? (atom false))

(defn- ensure-pod! []
  (when (compare-and-set! pod-loaded? false true)
    (pods/load-pod 'org.babashka/go-sqlite3 "0.2.8")
    (require '[pod.babashka.go-sqlite3 :as sql])))

(defn- sql-execute! [db stmt-vec]
  (ensure-pod!)
  ((resolve 'pod.babashka.go-sqlite3/execute!) db stmt-vec))

(defn- sql-query [db stmt-vec]
  (ensure-pod!)
  ((resolve 'pod.babashka.go-sqlite3/query) db stmt-vec))

;; ---------------------------------------------------------------------------
;; Schema DDL — executed on every open to ensure tables exist.
;; ---------------------------------------------------------------------------

(def ^:private schema-stmts
  ["CREATE TABLE IF NOT EXISTS runs (
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
    )"
   "CREATE INDEX IF NOT EXISTS idx_runs_parent ON runs(parent_run_id)"
   "CREATE TABLE IF NOT EXISTS steps (
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
    )"
   "CREATE TABLE IF NOT EXISTS research_events (
      id              INTEGER PRIMARY KEY AUTOINCREMENT,
      query_id        TEXT NOT NULL,
      question        TEXT NOT NULL,
      researchers     TEXT NOT NULL,
      composite_score REAL,
      reason          TEXT,
      created_at      INTEGER NOT NULL
    )"])

;; ---------------------------------------------------------------------------
;; Open / Close
;; ---------------------------------------------------------------------------

(defn open
  "Open or create a glitch database at `path` (default ~/.local/share/glitch/glitch.db).
   Creates parent directories, executes schema DDL, returns the db path string."
  [& [path]]
  (let [path (or path
                 (str (System/getProperty "user.home")
                      "/.local/share/glitch/glitch.db"))
        parent (.getParentFile (io/file path))]
    (when (and parent (not (.exists parent)))
      (.mkdirs parent))
    (doseq [stmt schema-stmts]
      (sql-execute! path [stmt]))
    path))

(defn open-for-project
  "Open a project-scoped database at <project-root>/.glitch/glitch.db."
  [project-root]
  (open (str project-root "/.glitch/glitch.db")))

(defn close
  "No-op — the go-sqlite3 pod does not hold persistent connections."
  [_db]
  nil)

;; ---------------------------------------------------------------------------
;; Runs
;; ---------------------------------------------------------------------------

(defn record-run
  "Insert a run record. `rec` is a map with keys:
     :name :input :workflow-file :model :variant :workspace
     :parent-run-id :workflow-name
   Returns the integer row ID of the inserted row."
  [db rec]
  (let [now (quot (System/currentTimeMillis) 1000)
        result (sql-execute! db
                 ["INSERT INTO runs (name, input, workflow_file, model,
                                     variant, workspace, parent_run_id,
                                     workflow_name, started_at, kind)
                   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'workflow')"
                  (:name rec)
                  (:input rec)
                  (:workflow-file rec)
                  (:model rec)
                  (:variant rec)
                  (or (:workspace rec) "")
                  (:parent-run-id rec)
                  (:workflow-name rec)
                  now])]
    (:last-inserted-id result)))

(defn finish-run
  "Update a run with final output, exit status, and optional totals map
   with keys :tokens-in :tokens-out :cost."
  [db id output exit-status & [totals]]
  (let [now (quot (System/currentTimeMillis) 1000)]
    (sql-execute! db
      ["UPDATE runs SET output = ?, exit_status = ?,
                        tokens_in = ?, tokens_out = ?,
                        cost_usd = ?, finished_at = ?
        WHERE id = ?"
       output
       exit-status
       (or (:tokens-in totals) 0)
       (or (:tokens-out totals) 0)
       (or (:cost totals) 0)
       now
       id])))

(defn get-run
  "Fetch a single run by ID. Returns a map or nil."
  [db id]
  (first (sql-query db ["SELECT * FROM runs WHERE id = ?" id])))

(defn list-runs
  "List runs with optional filters.
   Options: :parent-id, :workflow, :limit (default 50)."
  [db & {:keys [parent-id workflow limit] :or {limit 50}}]
  (cond
    parent-id
    (sql-query db ["SELECT * FROM runs WHERE parent_run_id = ? ORDER BY id DESC LIMIT ?"
                   parent-id limit])

    workflow
    (sql-query db ["SELECT * FROM runs WHERE workflow_file = ? ORDER BY id DESC LIMIT ?"
                   workflow limit])

    :else
    (sql-query db ["SELECT * FROM runs ORDER BY id DESC LIMIT ?" limit])))

;; ---------------------------------------------------------------------------
;; Steps
;; ---------------------------------------------------------------------------

(defn record-step
  "Insert or replace a step record. `rec` is a map with keys:
     :run-id :step-id :prompt :output :model :duration :kind
     :exit :tokens-in :tokens-out :gate :artifacts"
  [db rec]
  (sql-execute! db
    ["INSERT OR REPLACE INTO steps
        (run_id, step_id, prompt, output, model,
         duration_ms, kind, exit_status,
         tokens_in, tokens_out, gate_passed, artifacts)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
     (:run-id rec)
     (:step-id rec)
     (:prompt rec)
     (:output rec)
     (:model rec)
     (:duration rec)
     (:kind rec)
     (:exit rec)
     (:tokens-in rec)
     (:tokens-out rec)
     (:gate rec)
     (:artifacts rec)]))

(defn get-steps
  "Fetch all steps for a run, ordered by id."
  [db run-id]
  (sql-query db ["SELECT * FROM steps WHERE run_id = ? ORDER BY id" run-id]))
