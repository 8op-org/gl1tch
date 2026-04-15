package store

const createSchema = `
CREATE TABLE IF NOT EXISTS runs (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  kind          TEXT NOT NULL,
  name          TEXT NOT NULL,
  input         TEXT,
  output        TEXT,
  exit_status   INTEGER,
  started_at    INTEGER NOT NULL,
  finished_at   INTEGER,
  metadata      TEXT,
  workflow_file TEXT,
  repo          TEXT,
  model         TEXT,
  tokens_in     INTEGER,
  tokens_out    INTEGER,
  cost_usd      REAL,
  variant       TEXT
);

CREATE TABLE IF NOT EXISTS steps (
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
  UNIQUE(run_id, step_id)
);

CREATE TABLE IF NOT EXISTS research_events (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  query_id        TEXT NOT NULL,
  question        TEXT NOT NULL,
  researchers     TEXT NOT NULL,
  composite_score REAL,
  reason          TEXT,
  created_at      INTEGER NOT NULL
);
`
