# Daily Driver Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `glitch ask` a daily driver with a three-tier fallback chain: workflow routing → research loop → one-shot LLM.

**Architecture:** Four new packages (`esearch`, `store`, `research`, `observer`) ported and trimmed from gl1tch-legacy. YAML-backed researchers run via `pipeline.Run()`. ES provides persistent activity indexing. SQLite stores run history and research hints.

**Tech Stack:** Go 1.26, Elasticsearch 8.17, SQLite (modernc.org/sqlite), Cobra CLI, Ollama (qwen2.5:7b)

---

## File Structure

```
gl1tch/
├── docker-compose.yml                    (new — ES + Kibana)
├── cmd/
│   ├── ask.go                            (modify — three-tier fallback chain)
│   ├── observe.go                        (new — direct ES query command)
│   ├── up.go                             (new — docker lifecycle)
│   └── research_helpers.go               (new — buildResearchLoop, loadResearchers)
├── internal/
│   ├── esearch/
│   │   ├── client.go                     (new — raw HTTP ES client)
│   │   ├── mappings.go                   (new — index schemas for 5 indices)
│   │   └── client_test.go               (new)
│   ├── store/
│   │   ├── store.go                      (new — SQLite open/close/writer)
│   │   ├── schema.go                     (new — 3 tables, no migrations)
│   │   └── store_test.go                (new)
│   ├── research/
│   │   ├── types.go                      (new — Query, Evidence, Score, Result, Budget)
│   │   ├── registry.go                   (new — concurrent-safe researcher registry)
│   │   ├── researcher.go                 (new — Researcher interface)
│   │   ├── yaml_researcher.go            (new — workflow-backed researcher adapter)
│   │   ├── es_researcher.go              (new — ES activity + code query researchers)
│   │   ├── loop.go                       (new — 5-stage plan→gather→draft→critique→score)
│   │   ├── prompts.go                    (new — prompt templates as Go strings)
│   │   ├── score.go                      (new — composite scoring, critique parsing)
│   │   ├── hints.go                      (new — jaccard hint provider from SQLite)
│   │   ├── events.go                     (new — JSONL event sink for learning)
│   │   ├── loop_test.go                  (new)
│   │   ├── score_test.go                 (new)
│   │   └── registry_test.go             (new)
│   ├── observer/
│   │   ├── query.go                      (new — NL → ES → synthesized answer)
│   │   └── query_test.go                (new)
│   └── indexer/
│       └── indexer.go                    (modify — use esearch.Client)
└── researchers/                          (new — default YAML researcher files)
    ├── git-log.yaml
    ├── github-prs.yaml
    └── github-issues.yaml
```

---

### Task 1: Docker Compose + `glitch up/down`

**Files:**
- Create: `docker-compose.yml`
- Create: `cmd/up.go`

- [ ] **Step 1: Create docker-compose.yml**

```yaml
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.17.0
    container_name: glitch-es
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - xpack.security.http.ssl.enabled=false
      - ES_JAVA_OPTS=-Xms512m -Xmx512m
    ports:
      - "9200:9200"
    volumes:
      - glitch-es-data:/usr/share/elasticsearch/data
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 10

  kibana:
    image: docker.elastic.co/kibana/kibana:8.17.0
    container_name: glitch-kibana
    depends_on:
      elasticsearch:
        condition: service_healthy
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
      - SERVER_NAME=glitch-kibana
      - SERVER_HOST=0.0.0.0
      - XPACK_SECURITY_ENABLED=false
      - XPACK_REPORTING_ENABLED=false
      - XPACK_ENCRYPTEDSAVEDOBJECTS_ENCRYPTIONKEY=glitch_local_dev_encryption_key_32b
      - XPACK_REPORTING_ENCRYPTIONKEY=glitch_local_dev_reporting_key_32bytes_x
      - XPACK_SECURITY_ENCRYPTIONKEY=glitch_local_dev_security_key_32bytes_xx
    ports:
      - "5601:5601"
    volumes:
      - glitch-kibana-data:/usr/share/kibana/data
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:5601/api/status || exit 1"]
      interval: 15s
      timeout: 5s
      retries: 20
      start_period: 60s

volumes:
  glitch-es-data:
  glitch-kibana-data:
```

- [ ] **Step 2: Create cmd/up.go**

```go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "start elasticsearch and kibana",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dockerCompose("up", "-d")
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "stop elasticsearch and kibana",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dockerCompose("down")
	},
}

func dockerCompose(args ...string) error {
	composeFile, err := findComposeFile()
	if err != nil {
		return err
	}
	fullArgs := append([]string{"compose", "-f", composeFile}, args...)
	c := exec.Command("docker", fullArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func findComposeFile() (string, error) {
	// 1. Next to the binary
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "docker-compose.yml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	// 2. Source tree (go run / dev mode)
	_, thisFile, _, _ := runtime.Caller(0)
	candidate := filepath.Join(filepath.Dir(thisFile), "..", "docker-compose.yml")
	if abs, err := filepath.Abs(candidate); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	// 3. Current directory
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		return "docker-compose.yml", nil
	}
	return "", fmt.Errorf("docker-compose.yml not found; run from the gl1tch repo root or install with 'glitch up'")
}
```

- [ ] **Step 3: Test manually**

Run: `go build -o glitch . && ./glitch up`
Expected: ES + Kibana containers start. `curl http://localhost:9200` returns cluster info.

Run: `./glitch down`
Expected: Containers stop.

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml cmd/up.go
git commit -m "feat: add docker-compose and glitch up/down commands"
```

---

### Task 2: `internal/esearch` — ES Client

**Files:**
- Create: `internal/esearch/client.go`
- Create: `internal/esearch/mappings.go`
- Create: `internal/esearch/client_test.go`

- [ ] **Step 1: Write client_test.go**

```go
package esearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"cluster_name":"glitch"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestPingUnreachable(t *testing.T) {
	c := NewClient("http://127.0.0.1:1") // nothing listening
	if err := c.Ping(context.Background()); err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

func TestSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/glitch-events,glitch-code/_search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		resp := `{"hits":{"total":{"value":1},"hits":[{"_index":"glitch-events","_source":{"message":"test"}}]}}`
		w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	query := json.RawMessage(`{"query":{"match_all":{}}}`)
	result, err := c.Search(context.Background(), []string{"glitch-events", "glitch-code"}, query)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 result, got %d", result.Total)
	}
	if result.Results[0].Index != "glitch-events" {
		t.Fatalf("expected index glitch-events, got %s", result.Results[0].Index)
	}
}

func TestEnsureIndex(t *testing.T) {
	created := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "HEAD":
			w.WriteHeader(404) // index doesn't exist
		case "PUT":
			created = true
			w.Write([]byte(`{"acknowledged":true}`))
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.EnsureIndex(context.Background(), "test-index", EventsMapping)
	if err != nil {
		t.Fatalf("EnsureIndex: %v", err)
	}
	if !created {
		t.Fatal("expected index to be created")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/...`
Expected: FAIL — package doesn't exist yet.

- [ ] **Step 3: Create internal/esearch/mappings.go**

```go
package esearch

// Index names used by gl1tch.
const (
	IndexEvents    = "glitch-events"
	IndexSummaries = "glitch-summaries"
	IndexPipelines = "glitch-pipelines"
	IndexInsights  = "glitch-insights"
)

// EventsMapping is the schema for glitch-events.
const EventsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "type":         { "type": "keyword" },
      "source":       { "type": "keyword" },
      "repo":         { "type": "keyword" },
      "author":       { "type": "keyword" },
      "message":      { "type": "text" },
      "body":         { "type": "text" },
      "metadata":     { "type": "object", "enabled": false },
      "timestamp":    { "type": "date" }
    }
  }
}`

// SummariesMapping is the schema for glitch-summaries.
const SummariesMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "scope":     { "type": "keyword" },
      "date":      { "type": "date" },
      "summary":   { "type": "text" },
      "timestamp": { "type": "date" }
    }
  }
}`

// PipelinesMapping is the schema for glitch-pipelines.
const PipelinesMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "name":      { "type": "keyword" },
      "status":    { "type": "keyword" },
      "stdout":    { "type": "text" },
      "model":     { "type": "keyword" },
      "provider":  { "type": "keyword" },
      "timestamp": { "type": "date" }
    }
  }
}`

// InsightsMapping is the schema for glitch-insights.
const InsightsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "type":           { "type": "keyword" },
      "pattern":        { "type": "text" },
      "recommendation": { "type": "text" },
      "timestamp":      { "type": "date" }
    }
  }
}`

// AllIndices returns index name → mapping for bootstrap.
func AllIndices() map[string]string {
	return map[string]string{
		IndexEvents:    EventsMapping,
		IndexSummaries: SummariesMapping,
		IndexPipelines: PipelinesMapping,
		IndexInsights:  InsightsMapping,
	}
}
```

- [ ] **Step 4: Create internal/esearch/client.go**

```go
package esearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a thin Elasticsearch HTTP client.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a client pointing at the given ES base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Ping checks whether ES is reachable.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("es ping: status %d", resp.StatusCode)
	}
	return nil
}

// SearchResult is one hit from a search response.
type SearchResult struct {
	Index  string          `json:"_index"`
	Source json.RawMessage `json:"_source"`
}

// SearchResponse is the parsed ES search response.
type SearchResponse struct {
	Total   int
	Results []SearchResult
}

// Search executes a query against the given indices.
func (c *Client) Search(ctx context.Context, indices []string, query json.RawMessage) (*SearchResponse, error) {
	url := fmt.Sprintf("%s/%s/_search", c.baseURL, strings.Join(indices, ","))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(query))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("es search %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var raw struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []SearchResult `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("es search decode: %w", err)
	}
	return &SearchResponse{
		Total:   raw.Hits.Total.Value,
		Results: raw.Hits.Hits,
	}, nil
}

// BulkDoc is one document for bulk indexing.
type BulkDoc struct {
	ID   string
	Body json.RawMessage
}

// BulkIndex sends a batch of documents to ES via the bulk API.
func (c *Client) BulkIndex(ctx context.Context, index string, docs []BulkDoc) error {
	var buf bytes.Buffer
	for _, doc := range docs {
		meta := map[string]any{"index": map[string]any{"_index": index}}
		if doc.ID != "" {
			meta["index"].(map[string]any)["_id"] = doc.ID
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')
		buf.Write(doc.Body)
		buf.WriteByte('\n')
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/_bulk", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es bulk %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return nil
}

// EnsureIndex creates an index with the given mapping if it doesn't exist.
func (c *Client) EnsureIndex(ctx context.Context, index, mapping string) error {
	// Check if index exists
	req, err := http.NewRequestWithContext(ctx, "HEAD", c.baseURL+"/"+index, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil // already exists
	}

	// Create it
	req, err = http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/"+index, strings.NewReader(mapping))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es create index %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/esearch/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/esearch/
git commit -m "feat: add esearch package — thin ES HTTP client with index management"
```

---

### Task 3: `internal/store` — SQLite

**Files:**
- Create: `internal/store/schema.go`
- Create: `internal/store/store.go`
- Create: `internal/store/store_test.go`

- [ ] **Step 1: Add modernc.org/sqlite dependency**

Run: `cd /Users/stokes/Projects/gl1tch && go get modernc.org/sqlite`

- [ ] **Step 2: Write store_test.go**

```go
package store

import (
	"path/filepath"
	"testing"
)

func TestOpenAndClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := OpenAt(path)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()
}

func TestRecordRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := OpenAt(path)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	id, err := s.RecordRun("workflow", "github-pr-review", "https://github.com/foo/bar/pull/1")
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive run ID, got %d", id)
	}

	if err := s.FinishRun(id, "review output", 0); err != nil {
		t.Fatalf("FinishRun: %v", err)
	}
}

func TestRecordStep(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := OpenAt(path)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	id, _ := s.RecordRun("research", "ask", "what broke?")
	err = s.RecordStep(id, "plan", "pick researchers", "github-prs", "qwen2.5:7b", 150)
	if err != nil {
		t.Fatalf("RecordStep: %v", err)
	}
}

func TestRecordResearchEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := OpenAt(path)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	err = s.RecordResearchEvent(ResearchEvent{
		QueryID:        "q-123",
		Question:       "what PRs are open?",
		Researchers:    `["github-prs"]`,
		CompositeScore: 0.85,
		Reason:         "accepted",
	})
	if err != nil {
		t.Fatalf("RecordResearchEvent: %v", err)
	}
}

func TestSimilarResearchEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := OpenAt(path)
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	// Seed some events
	s.RecordResearchEvent(ResearchEvent{
		QueryID: "q-1", Question: "what PRs are open in ensemble",
		Researchers: `["github-prs"]`, CompositeScore: 0.9, Reason: "accepted",
	})
	s.RecordResearchEvent(ResearchEvent{
		QueryID: "q-2", Question: "show me recent commits",
		Researchers: `["git-log"]`, CompositeScore: 0.8, Reason: "accepted",
	})

	events, err := s.SimilarResearchEvents("what PRs merged this week", 5)
	if err != nil {
		t.Fatalf("SimilarResearchEvents: %v", err)
	}
	// Should find the PR-related event as most similar
	if len(events) == 0 {
		t.Fatal("expected at least one similar event")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/store/...`
Expected: FAIL — package doesn't exist yet.

- [ ] **Step 4: Create internal/store/schema.go**

```go
package store

const createSchema = `
CREATE TABLE IF NOT EXISTS runs (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  kind        TEXT NOT NULL,
  name        TEXT NOT NULL,
  input       TEXT,
  output      TEXT,
  exit_status INTEGER,
  started_at  INTEGER NOT NULL,
  finished_at INTEGER,
  metadata    TEXT
);

CREATE TABLE IF NOT EXISTS steps (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id      INTEGER NOT NULL,
  step_id     TEXT NOT NULL,
  prompt      TEXT,
  output      TEXT,
  model       TEXT,
  duration_ms INTEGER,
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
```

- [ ] **Step 5: Create internal/store/store.go**

```go
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store is the SQLite-backed run history and research event store.
type Store struct {
	db *sql.DB
}

// ResearchEvent is one record in the research_events table.
type ResearchEvent struct {
	QueryID        string
	Question       string
	Researchers    string // JSON array
	CompositeScore float64
	Reason         string
}

// Open opens (or creates) the store at the default path.
func Open() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".local", "share", "glitch", "glitch.db")
	return OpenAt(path)
}

// OpenAt opens (or creates) the store at the given path.
func OpenAt(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(createSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("store schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close shuts down the store.
func (s *Store) Close() error {
	return s.db.Close()
}

// RecordRun inserts a new run and returns its ID.
func (s *Store) RecordRun(kind, name, input string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO runs (kind, name, input, started_at) VALUES (?, ?, ?, ?)`,
		kind, name, input, time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FinishRun updates a run with output and exit status.
func (s *Store) FinishRun(id int64, output string, exitStatus int) error {
	_, err := s.db.Exec(
		`UPDATE runs SET output = ?, exit_status = ?, finished_at = ? WHERE id = ?`,
		output, exitStatus, time.Now().Unix(), id,
	)
	return err
}

// RecordStep inserts a step record for a run.
func (s *Store) RecordStep(runID int64, stepID, prompt, output, model string, durationMs int64) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO steps (run_id, step_id, prompt, output, model, duration_ms) VALUES (?, ?, ?, ?, ?, ?)`,
		runID, stepID, prompt, output, model, durationMs,
	)
	return err
}

// RecordResearchEvent inserts a research event for the hints provider.
func (s *Store) RecordResearchEvent(evt ResearchEvent) error {
	_, err := s.db.Exec(
		`INSERT INTO research_events (query_id, question, researchers, composite_score, reason, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		evt.QueryID, evt.Question, evt.Researchers, evt.CompositeScore, evt.Reason, time.Now().Unix(),
	)
	return err
}

// SimilarResearchEvents returns research events with questions similar to the
// given question, using token jaccard similarity. Returns up to limit results
// sorted by composite_score descending.
func (s *Store) SimilarResearchEvents(question string, limit int) ([]ResearchEvent, error) {
	queryTokens := tokenise(question)
	if len(queryTokens) == 0 {
		return nil, nil
	}

	rows, err := s.db.Query(
		`SELECT query_id, question, researchers, composite_score, reason FROM research_events ORDER BY created_at DESC LIMIT 200`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		evt ResearchEvent
		sim float64
	}
	var matches []scored
	for rows.Next() {
		var evt ResearchEvent
		if err := rows.Scan(&evt.QueryID, &evt.Question, &evt.Researchers, &evt.CompositeScore, &evt.Reason); err != nil {
			continue
		}
		pastTokens := tokenise(evt.Question)
		sim := jaccard(queryTokens, pastTokens)
		if sim >= 0.2 {
			matches = append(matches, scored{evt: evt, sim: sim})
		}
	}

	// Sort by composite_score * similarity
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			si := matches[i].sim * matches[i].evt.CompositeScore
			sj := matches[j].sim * matches[j].evt.CompositeScore
			if sj > si {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}
	out := make([]ResearchEvent, len(matches))
	for i, m := range matches {
		out[i] = m.evt
	}
	return out, nil
}

// tokenise lowercases and splits on non-alphanumeric, drops stopwords and
// 1-char tokens, does trailing-s plural collapse.
func tokenise(s string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		tok := strings.ToLower(cur.String())
		cur.Reset()
		if len(tok) <= 1 {
			return
		}
		if len(tok) >= 3 && tok[len(tok)-1] == 's' {
			tok = tok[:len(tok)-1]
		}
		if stopwords[tok] {
			return
		}
		tokens = append(tokens, tok)
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		isAlnum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if isAlnum {
			cur.WriteByte(c)
		} else {
			flush()
		}
	}
	flush()
	seen := make(map[string]struct{}, len(tokens))
	uniq := tokens[:0]
	for _, t := range tokens {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		uniq = append(uniq, t)
	}
	return uniq
}

var stopwords = map[string]bool{
	"the": true, "is": true, "are": true, "was": true, "were": true,
	"this": true, "that": true, "what": true, "who": true, "when": true,
	"where": true, "why": true, "how": true, "do": true, "does": true,
	"did": true, "you": true, "me": true, "my": true, "of": true,
	"to": true, "in": true, "on": true, "at": true, "and": true,
	"or": true, "for": true, "from": true, "as": true, "by": true,
	"with": true, "an": true, "be": true, "have": true, "has": true,
	"had": true, "it": true, "its": true, "any": true, "some": true,
	"all": true, "no": true, "yes": true, "yet": true,
}

func jaccard(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(a))
	for _, t := range a {
		set[t] = struct{}{}
	}
	intersect := 0
	for _, t := range b {
		if _, ok := set[t]; ok {
			intersect++
		}
	}
	union := len(a) + len(b) - intersect
	if union == 0 {
		return 0
	}
	return float64(intersect) / float64(union)
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/store/...`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/store/
git commit -m "feat: add store package — SQLite for run history and research hints"
```

---

### Task 4: `internal/research` — Types, Registry, Researcher Interface

**Files:**
- Create: `internal/research/types.go`
- Create: `internal/research/researcher.go`
- Create: `internal/research/registry.go`
- Create: `internal/research/registry_test.go`

- [ ] **Step 1: Write registry_test.go**

```go
package research

import (
	"context"
	"testing"
)

type stubResearcher struct {
	name     string
	describe string
	evidence Evidence
}

func (s *stubResearcher) Name() string    { return s.name }
func (s *stubResearcher) Describe() string { return s.describe }
func (s *stubResearcher) Gather(_ context.Context, _ ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	return s.evidence, nil
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	reg := NewRegistry()
	r := &stubResearcher{name: "git-log", describe: "recent commits"}
	if err := reg.Register(r); err != nil {
		t.Fatalf("Register: %v", err)
	}
	got, ok := reg.Lookup("git-log")
	if !ok {
		t.Fatal("expected to find git-log")
	}
	if got.Name() != "git-log" {
		t.Fatalf("expected git-log, got %s", got.Name())
	}
}

func TestRegistryDuplicate(t *testing.T) {
	reg := NewRegistry()
	r := &stubResearcher{name: "git-log", describe: "recent commits"}
	reg.Register(r)
	if err := reg.Register(r); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubResearcher{name: "b-researcher", describe: "b"})
	reg.Register(&stubResearcher{name: "a-researcher", describe: "a"})
	list := reg.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if list[0].Name() != "a-researcher" {
		t.Fatalf("expected sorted, got %s first", list[0].Name())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/research/...`
Expected: FAIL

- [ ] **Step 3: Create internal/research/researcher.go**

```go
package research

import "context"

// Researcher is anything the loop can ask for evidence.
type Researcher interface {
	Name() string
	Describe() string
	Gather(ctx context.Context, q ResearchQuery, prior EvidenceBundle) (Evidence, error)
}
```

- [ ] **Step 4: Create internal/research/types.go**

```go
package research

import (
	"encoding/json"
	"time"
)

// LLMFn is the seam between the loop and the provider system.
type LLMFn func(ctx context.Context, prompt string) (string, error)

// DefaultLocalModel is the hard default for generation.
const DefaultLocalModel = "qwen2.5:7b"

// ResearchQuery is the input to one research call.
type ResearchQuery struct {
	ID       string            `json:"id"`
	Question string            `json:"question"`
	Context  map[string]string `json:"context,omitempty"`
}

// Evidence is one piece of information returned by a Researcher.
type Evidence struct {
	Source    string   `json:"source"`
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Refs     []string `json:"refs,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Truncated bool    `json:"truncated,omitempty"`
}

// EvidenceBundle is the accumulated set of Evidence values.
type EvidenceBundle struct {
	Items []Evidence `json:"items"`
}

func (b *EvidenceBundle) Add(e Evidence) {
	if b == nil {
		return
	}
	b.Items = append(b.Items, e)
}

func (b *EvidenceBundle) Len() int {
	if b == nil {
		return 0
	}
	return len(b.Items)
}

func (b *EvidenceBundle) Sources() []string {
	if b == nil {
		return nil
	}
	seen := make(map[string]struct{}, len(b.Items))
	out := make([]string, 0, len(b.Items))
	for _, it := range b.Items {
		if _, ok := seen[it.Source]; ok {
			continue
		}
		seen[it.Source] = struct{}{}
		out = append(out, it.Source)
	}
	return out
}

// Budget caps how much work one research call may do.
type Budget struct {
	MaxIterations  int
	MaxWallclock   time.Duration
}

func DefaultBudget() Budget {
	return Budget{
		MaxIterations: 3,
		MaxWallclock:  60 * time.Second,
	}
}

// Reason explains why the loop returned its Result.
type Reason string

const (
	ReasonAccepted       Reason = "accepted"
	ReasonBudgetExceeded Reason = "budget_exceeded"
	ReasonUnscored       Reason = "unscored"
)

// Score is the per-signal breakdown.
type Score struct {
	Composite            float64  `json:"composite"`
	EvidenceCoverage     *float64 `json:"evidence_coverage,omitempty"`
	CrossCapabilityAgree *float64 `json:"cross_capability_agreement,omitempty"`
	JudgeScore           *float64 `json:"judge_score,omitempty"`
}

func (s Score) MarshalJSON() ([]byte, error) {
	type alias Score
	return json.Marshal(alias(s))
}

func Ptr(v float64) *float64 { return &v }

func Float(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// Result is the output of one research call.
type Result struct {
	Query      ResearchQuery  `json:"query"`
	Draft      string         `json:"draft"`
	Bundle     EvidenceBundle `json:"bundle"`
	Score      Score          `json:"score"`
	Reason     Reason         `json:"reason"`
	Iterations int            `json:"iterations"`
}
```

- [ ] **Step 5: Create internal/research/registry.go**

```go
package research

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrDuplicateResearcher = errors.New("research: duplicate researcher name")

// Registry is the process-wide registry of Researchers.
type Registry struct {
	mu          sync.RWMutex
	researchers map[string]Researcher
}

func NewRegistry() *Registry {
	return &Registry{researchers: make(map[string]Researcher)}
}

func (r *Registry) Register(researcher Researcher) error {
	if researcher == nil {
		return errors.New("research: cannot register nil researcher")
	}
	name := researcher.Name()
	if name == "" {
		return errors.New("research: researcher Name() must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.researchers[name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateResearcher, name)
	}
	r.researchers[name] = researcher
	return nil
}

func (r *Registry) Lookup(name string) (Researcher, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	researcher, ok := r.researchers[name]
	return researcher, ok
}

func (r *Registry) List() []Researcher {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Researcher, 0, len(r.researchers))
	for _, researcher := range r.researchers {
		out = append(out, researcher)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.researchers))
	for name := range r.researchers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/research/...`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/research/
git commit -m "feat: add research types, researcher interface, and registry"
```

---

### Task 5: Research Loop — Prompts + Score + Loop Engine

**Files:**
- Create: `internal/research/prompts.go`
- Create: `internal/research/score.go`
- Create: `internal/research/events.go`
- Create: `internal/research/loop.go`
- Create: `internal/research/loop_test.go`
- Create: `internal/research/score_test.go`

- [ ] **Step 1: Write loop_test.go**

```go
package research

import (
	"context"
	"fmt"
	"testing"
)

// scriptedLLM returns canned responses in order.
func scriptedLLM(responses ...string) LLMFn {
	i := 0
	return func(_ context.Context, _ string) (string, error) {
		if i >= len(responses) {
			return "", fmt.Errorf("scriptedLLM: exhausted after %d calls", len(responses))
		}
		r := responses[i]
		i++
		return r, nil
	}
}

func TestLoopHappyPath(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubResearcher{
		name:     "github-prs",
		describe: "open pull requests",
		evidence: Evidence{Source: "github-prs", Title: "PR #42", Body: "fix: broken auth"},
	})

	llm := scriptedLLM(
		`["github-prs"]`,             // plan
		"The auth was fixed in PR #42", // draft
		`[{"text":"PR #42 fixed auth","label":"grounded"}]`, // critique
		"0.9", // judge
	)
	loop := NewLoop(reg, llm)

	result, err := loop.Run(context.Background(), ResearchQuery{Question: "what happened?"}, DefaultBudget())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Draft == "" {
		t.Fatal("expected non-empty draft")
	}
	if result.Reason != ReasonAccepted {
		t.Fatalf("expected accepted, got %s", result.Reason)
	}
}

func TestLoopEmptyPlanShortCircuits(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubResearcher{name: "git-log", describe: "commits"})

	llm := scriptedLLM(`[]`) // empty plan
	loop := NewLoop(reg, llm)

	result, err := loop.Run(context.Background(), ResearchQuery{Question: "nothing relevant"}, DefaultBudget())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Reason != ReasonUnscored {
		t.Fatalf("expected unscored, got %s", result.Reason)
	}
}

func TestLoopRejectsNilLLM(t *testing.T) {
	reg := NewRegistry()
	loop := NewLoop(reg, nil)
	_, err := loop.Run(context.Background(), ResearchQuery{Question: "test"}, DefaultBudget())
	if err == nil {
		t.Fatal("expected error for nil LLM")
	}
}

func TestParsePlanTolerant(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{`["git-log","github-prs"]`, 2},
		{`Some preamble: ["git-log"]`, 1},
		{`[\"git-log\"]`, 1},
		{`[git-log, github-prs]`, 2},
		{`[]`, 0},
	}
	for _, tc := range cases {
		names, err := ParsePlan(tc.input)
		if err != nil && tc.want > 0 {
			t.Errorf("ParsePlan(%q): %v", tc.input, err)
			continue
		}
		if len(names) != tc.want {
			t.Errorf("ParsePlan(%q): got %d names, want %d", tc.input, len(names), tc.want)
		}
	}
}
```

- [ ] **Step 2: Write score_test.go**

```go
package research

import "testing"

func TestEvidenceCoverage(t *testing.T) {
	c := Critique{Claims: []CritiqueClaim{
		{Text: "claim 1", Label: LabelGrounded},
		{Text: "claim 2", Label: LabelPartial},
		{Text: "claim 3", Label: LabelUngrounded},
	}}
	got := EvidenceCoverage(c)
	want := (1.0 + 0.5 + 0.0) / 3.0
	if got != want {
		t.Fatalf("EvidenceCoverage: got %f, want %f", got, want)
	}
}

func TestCrossCapabilityAgree(t *testing.T) {
	bundle := EvidenceBundle{Items: []Evidence{
		{Source: "git-log"}, {Source: "github-prs"},
	}}
	if CrossCapabilityAgree(bundle) != 1.0 {
		t.Fatal("expected 1.0 for 2 sources")
	}

	single := EvidenceBundle{Items: []Evidence{{Source: "git-log"}}}
	if CrossCapabilityAgree(single) != 0.4 {
		t.Fatal("expected 0.4 for 1 source")
	}
}

func TestParseCritique(t *testing.T) {
	raw := `[{"text":"PR #42 fixed auth","label":"grounded"},{"text":"commit abc123","label":"ungrounded"}]`
	c, err := ParseCritique(raw)
	if err != nil {
		t.Fatalf("ParseCritique: %v", err)
	}
	if len(c.Claims) != 2 {
		t.Fatalf("expected 2 claims, got %d", len(c.Claims))
	}
	if c.Claims[0].Label != LabelGrounded {
		t.Fatalf("expected grounded, got %s", c.Claims[0].Label)
	}
}

func TestParseJudgeScore(t *testing.T) {
	v, err := ParseJudgeScore("0.85")
	if err != nil {
		t.Fatalf("ParseJudgeScore: %v", err)
	}
	if v != 0.85 {
		t.Fatalf("expected 0.85, got %f", v)
	}

	_, err = ParseJudgeScore("no number here")
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/research/...`
Expected: FAIL — prompts.go, score.go, events.go, loop.go don't exist yet.

- [ ] **Step 4: Create internal/research/prompts.go**

```go
package research

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PlanPrompt builds the prompt that asks the LLM to pick researchers.
func PlanPrompt(question string, researchers []Researcher, hints string) string {
	var b strings.Builder
	b.WriteString(`You are the planning stage of a research loop. Your job is to pick which
researchers should gather evidence for a user's question. You do NOT answer
the question yourself. You ONLY pick researcher names from the list below.

Question:
`)
	b.WriteString(strings.TrimSpace(question))
	b.WriteString("\n\nAvailable researchers (you may only pick names from this list):\n")
	for _, r := range researchers {
		fmt.Fprintf(&b, "- %s: %s\n", r.Name(), strings.TrimSpace(r.Describe()))
	}
	if hints != "" {
		fmt.Fprintf(&b, "\nBrain hints (past calls for similar questions):\n%s\n", hints)
	}
	b.WriteString(`
Rules:
1. Output ONLY a JSON array of researcher names. No prose, no markdown, no explanation.
2. Pick only names that appear verbatim in the list above.
3. If no researcher fits the question, output [].
4. NEVER invent a researcher name.

Output (JSON array only):
`)
	return b.String()
}

// DraftPrompt builds the prompt that asks the LLM to answer using evidence.
func DraftPrompt(question string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString(`You are the drafting stage of a research loop. Answer the user's question
using ONLY the evidence below. You may not invent facts or prior knowledge.

Question:
`)
	b.WriteString(strings.TrimSpace(question))
	b.WriteString("\n\nEvidence:\n")
	if bundle.Len() == 0 {
		b.WriteString("(no evidence was gathered)\n")
	} else {
		for i, ev := range bundle.Items {
			fmt.Fprintf(&b, "[%d] source=%s\n", i+1, ev.Source)
			if ev.Title != "" {
				fmt.Fprintf(&b, "    title: %s\n", ev.Title)
			}
			if len(ev.Refs) > 0 {
				fmt.Fprintf(&b, "    refs: %s\n", strings.Join(ev.Refs, ", "))
			}
			if ev.Body != "" {
				fmt.Fprintf(&b, "    body:\n      %s\n", strings.TrimSpace(ev.Body))
			}
		}
	}
	b.WriteString(`
Rules:
1. Cite specific identifiers (PR numbers, commit SHAs, file paths, dates, URLs)
   ONLY when they appear verbatim in the evidence above. Never invent them.
2. If the evidence contains nothing relevant, reply with exactly:
   "I don't have enough evidence to answer that."
3. Do not say "you should run" any command.
4. Be concise. Lead with the answer, not preamble.

Answer:
`)
	return b.String()
}

// CritiquePrompt builds the prompt that labels claims as grounded/partial/ungrounded.
func CritiquePrompt(draft string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString(`Extract the factual claims from this draft and label each one against the evidence.

Draft:
`)
	b.WriteString(strings.TrimSpace(draft))
	b.WriteString("\n\nEvidence:\n")
	for i, ev := range bundle.Items {
		fmt.Fprintf(&b, "[%d] source=%s", i+1, ev.Source)
		if ev.Title != "" {
			fmt.Fprintf(&b, " title=%s", ev.Title)
		}
		b.WriteByte('\n')
	}
	b.WriteString(`
Labels: "grounded", "partial", "ungrounded"
Output ONLY a JSON array: [{"text": "...", "label": "grounded|partial|ungrounded"}, ...]

Output (JSON array only):
`)
	return b.String()
}

// JudgePrompt asks for a 0.0-1.0 score of how well the draft answers the question.
func JudgePrompt(question, draft string, bundle EvidenceBundle) string {
	var b strings.Builder
	b.WriteString(`Read the question, draft answer, and evidence. Return ONE number between 0.0
and 1.0 representing how well the draft answers the question using only the evidence.

Question: `)
	b.WriteString(strings.TrimSpace(question))
	b.WriteString("\n\nDraft:\n")
	b.WriteString(strings.TrimSpace(draft))
	b.WriteString("\n\nEvidence:\n")
	for i, ev := range bundle.Items {
		fmt.Fprintf(&b, "[%d] source=%s title=%s\n", i+1, ev.Source, ev.Title)
	}
	b.WriteString("\nScore (single number, 0.0 to 1.0):\n")
	return b.String()
}

// ParsePlan extracts researcher names from planner output.
func ParsePlan(raw string) ([]string, error) {
	if names, err := parsePlanJSON(raw); err == nil {
		return names, nil
	}
	// Try stripping backslash escapes
	if unescaped := strings.ReplaceAll(raw, `\"`, `"`); unescaped != raw {
		if names, err := parsePlanJSON(unescaped); err == nil {
			return names, nil
		}
	}
	// Try bare identifiers
	if names := parsePlanBare(raw); len(names) > 0 {
		return names, nil
	}
	return parsePlanJSON(raw) // surface original error
}

func parsePlanJSON(raw string) ([]string, error) {
	start := strings.Index(raw, "[")
	if start < 0 {
		return nil, fmt.Errorf("research: no JSON array in: %q", truncatePlan(raw, 200))
	}
	depth, inStr, esc, end := 0, false, false, -1
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inStr {
			if esc { esc = false; continue }
			if c == '\\' { esc = true; continue }
			if c == '"' { inStr = false }
			continue
		}
		switch c {
		case '"': inStr = true
		case '[': depth++
		case ']':
			depth--
			if depth == 0 { end = i + 1 }
		}
		if end > 0 { break }
	}
	if end < 0 {
		return nil, fmt.Errorf("research: unbalanced JSON array: %q", truncatePlan(raw, 200))
	}
	var names []string
	if err := json.Unmarshal([]byte(raw[start:end]), &names); err != nil {
		return nil, fmt.Errorf("research: not a string array: %v", err)
	}
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" { continue }
		if _, dup := seen[n]; dup { continue }
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}

func parsePlanBare(raw string) []string {
	start := strings.Index(raw, "[")
	if start < 0 { return nil }
	end := strings.Index(raw[start:], "]")
	if end < 0 { end = len(raw) } else { end += start }
	body := raw[start+1 : end]
	var tokens []string
	var cur strings.Builder
	flush := func() {
		t := strings.TrimSpace(cur.String())
		cur.Reset()
		if t != "" { tokens = append(tokens, t) }
	}
	for i := 0; i < len(body); i++ {
		c := body[i]
		isToken := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if isToken { cur.WriteByte(c) } else { flush() }
	}
	flush()
	if len(tokens) == 0 { return nil }
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if _, dup := seen[t]; dup { continue }
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func truncatePlan(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n] + "..."
}
```

- [ ] **Step 5: Create internal/research/score.go**

```go
package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type CritiqueLabel string

const (
	LabelGrounded   CritiqueLabel = "grounded"
	LabelPartial    CritiqueLabel = "partial"
	LabelUngrounded CritiqueLabel = "ungrounded"
)

type Critique struct {
	Claims []CritiqueClaim `json:"claims"`
}

type CritiqueClaim struct {
	Text  string        `json:"text"`
	Label CritiqueLabel `json:"label"`
}

func ParseCritique(raw string) (Critique, error) {
	start := strings.Index(raw, "[")
	if start < 0 {
		return Critique{}, fmt.Errorf("research: critique has no JSON array: %q", truncatePlan(raw, 200))
	}
	depth, inStr, esc, end := 0, false, false, -1
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inStr {
			if esc { esc = false; continue }
			if c == '\\' { esc = true; continue }
			if c == '"' { inStr = false }
			continue
		}
		switch c {
		case '"': inStr = true
		case '[': depth++
		case ']':
			depth--
			if depth == 0 { end = i + 1 }
		}
		if end > 0 { break }
	}
	if end < 0 {
		return Critique{}, fmt.Errorf("research: critique has unbalanced array")
	}
	var raws []struct {
		Text  string `json:"text"`
		Label string `json:"label"`
	}
	if err := json.Unmarshal([]byte(raw[start:end]), &raws); err != nil {
		return Critique{}, fmt.Errorf("research: critique not {text,label} array: %v", err)
	}
	out := Critique{Claims: make([]CritiqueClaim, 0, len(raws))}
	for _, r := range raws {
		text := strings.TrimSpace(r.Text)
		if text == "" { continue }
		out.Claims = append(out.Claims, CritiqueClaim{Text: text, Label: normaliseLabel(r.Label)})
	}
	return out, nil
}

func normaliseLabel(s string) CritiqueLabel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "grounded": return LabelGrounded
	case "partial": return LabelPartial
	default: return LabelUngrounded
	}
}

func EvidenceCoverage(c Critique) float64 {
	if len(c.Claims) == 0 { return 0 }
	var sum float64
	for _, claim := range c.Claims {
		switch claim.Label {
		case LabelGrounded: sum += 1.0
		case LabelPartial: sum += 0.5
		}
	}
	return sum / float64(len(c.Claims))
}

func CrossCapabilityAgree(bundle EvidenceBundle) float64 {
	switch n := len(bundle.Sources()); {
	case n >= 2: return 1.0
	case n == 1: return 0.4
	default: return 0.0
	}
}

func ParseJudgeScore(raw string) (float64, error) {
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if (c >= '0' && c <= '9') || c == '.' {
			end := i
			for end < len(raw) {
				cc := raw[end]
				if (cc >= '0' && cc <= '9') || cc == '.' { end++ } else { break }
			}
			var v float64
			n, err := fmt.Sscanf(raw[i:end], "%f", &v)
			if err == nil && n == 1 {
				if v < 0 || v > 1 {
					return 0, fmt.Errorf("research: judge score out of range: %v", v)
				}
				return v, nil
			}
			i = end
		}
	}
	return 0, fmt.Errorf("research: judge output has no score: %q", truncatePlan(raw, 200))
}

// Composite computes an equal-weight average of non-nil signals.
func Composite(s Score) float64 {
	var sum float64
	var n int
	for _, p := range []*float64{s.EvidenceCoverage, s.CrossCapabilityAgree, s.JudgeScore} {
		if p != nil {
			sum += *p
			n++
		}
	}
	if n == 0 { return 0 }
	return sum / float64(n)
}

// ComputeScore runs critique + judge and returns a Score.
func ComputeScore(ctx context.Context, llm LLMFn, question, draft string, bundle EvidenceBundle) (Score, Critique) {
	out := Score{}
	var crit Critique

	// Signal 1: cross-capability agreement (free)
	cca := CrossCapabilityAgree(bundle)
	out.CrossCapabilityAgree = Ptr(cca)

	// Signal 2: evidence coverage via critique
	if llm != nil {
		raw, err := llm(ctx, CritiquePrompt(draft, bundle))
		if err == nil {
			parsed, perr := ParseCritique(raw)
			if perr == nil {
				crit = parsed
				ec := EvidenceCoverage(parsed)
				out.EvidenceCoverage = Ptr(ec)
			}
		}
	}

	// Signal 3: judge pass
	if llm != nil {
		raw, err := llm(ctx, JudgePrompt(question, draft, bundle))
		if err == nil {
			js, perr := ParseJudgeScore(raw)
			if perr == nil {
				out.JudgeScore = Ptr(js)
			}
		}
	}

	out.Composite = Composite(out)
	return out, crit
}
```

- [ ] **Step 6: Create internal/research/events.go**

```go
package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type EventType string

const (
	EventTypeAttempt EventType = "research_attempt"
	EventTypeScore   EventType = "research_score"
)

type Event struct {
	Type      EventType      `json:"type"`
	Timestamp string         `json:"timestamp"`
	QueryID   string         `json:"query_id,omitempty"`
	Question  string         `json:"question,omitempty"`
	Iteration int            `json:"iteration,omitempty"`
	Reason    Reason         `json:"reason,omitempty"`
	Score     Score          `json:"score,omitempty"`
	Bundle    *EvidenceBundle `json:"bundle,omitempty"`
}

// EventSink is what the loop emits to.
type EventSink interface {
	Emit(Event) error
}

type nopSink struct{}
func (nopSink) Emit(Event) error { return nil }

// MemoryEventSink is an in-memory sink for tests.
type MemoryEventSink struct {
	mu     sync.Mutex
	events []Event
}

func NewMemoryEventSink() *MemoryEventSink { return &MemoryEventSink{} }
func (m *MemoryEventSink) Emit(ev Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, ev)
	return nil
}
func (m *MemoryEventSink) Events() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}

// FileEventSink appends JSONL events to a file.
type FileEventSink struct {
	Path string
	mu   sync.Mutex
}

func NewFileEventSink(path string) *FileEventSink {
	if path == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			path = filepath.Join(home, ".glitch", "research_events.jsonl")
		} else {
			path = filepath.Join(".glitch", "research_events.jsonl")
		}
	}
	return &FileEventSink{Path: path}
}

func (f *FileEventSink) Emit(ev Event) error {
	if ev.Timestamp == "" {
		ev.Timestamp = time.Now().Format(time.RFC3339)
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	fh, err := os.OpenFile(f.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fmt.Fprintf(fh, "%s\n", data)
	return err
}

func emitAttempt(sink EventSink, q ResearchQuery, iter int, score Score, bundle EvidenceBundle, reason Reason) {
	if sink == nil { return }
	now := time.Now().Format(time.RFC3339)
	bundleCopy := bundle
	_ = sink.Emit(Event{
		Type: EventTypeAttempt, Timestamp: now, QueryID: q.ID, Question: q.Question,
		Iteration: iter, Reason: reason, Score: score, Bundle: &bundleCopy,
	})
	_ = sink.Emit(Event{
		Type: EventTypeScore, Timestamp: now, QueryID: q.ID,
		Iteration: iter, Reason: reason, Score: score,
	})
}
```

- [ ] **Step 7: Create internal/research/loop.go**

```go
package research

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Loop is the research-loop driver: plan → gather → draft → critique → score.
type Loop struct {
	registry *Registry
	llm      LLMFn
	draftLLM LLMFn
	model    string
	logger   *slog.Logger
	events   EventSink
	hints    string // pre-computed hint string for this call
}

func NewLoop(reg *Registry, llm LLMFn) *Loop {
	return &Loop{
		registry: reg,
		llm:      llm,
		model:    DefaultLocalModel,
		logger:   slog.New(slog.NewTextHandler(discardWriter{}, nil)),
		events:   nopSink{},
	}
}

func (l *Loop) WithDraftLLM(fn LLMFn) *Loop {
	cp := *l
	cp.draftLLM = fn
	return &cp
}

func (l *Loop) WithEventSink(sink EventSink) *Loop {
	if sink == nil { l.events = nopSink{} } else { l.events = sink }
	return l
}

func (l *Loop) WithLogger(log *slog.Logger) *Loop {
	if log != nil { l.logger = log }
	return l
}

func (l *Loop) WithHints(hints string) *Loop {
	l.hints = hints
	return l
}

// Run executes the research loop.
func (l *Loop) Run(ctx context.Context, q ResearchQuery, budget Budget) (Result, error) {
	if l.registry == nil {
		return Result{}, errors.New("research: Loop has nil registry")
	}
	if l.llm == nil {
		return Result{}, errors.New("research: Loop has nil llm")
	}
	if budget.MaxIterations <= 0 {
		budget = DefaultBudget()
	}
	if q.ID == "" {
		q.ID = newQueryID()
	}

	deadline := time.Now().Add(budget.MaxWallclock)
	if budget.MaxWallclock <= 0 {
		deadline = time.Now().Add(60 * time.Second)
	}

	const acceptThreshold = 0.7
	var (
		best       Result
		haveBest   bool
		extraNeeds []string
		bundleSoFar EvidenceBundle
		picksSoFar = make(map[string]struct{})
		iter       int
	)

	for iter = 1; iter <= budget.MaxIterations; iter++ {
		if ctx.Err() != nil || time.Now().After(deadline) {
			return returnBest(q, best, haveBest, iter-1), nil
		}

		// Plan
		augmented := q.Question
		if len(extraNeeds) > 0 {
			augmented += "\n\nAdditional evidence needs: " + strings.Join(extraNeeds, "; ")
		}
		picks, err := l.plan(ctx, augmented)
		if err != nil {
			return Result{}, fmt.Errorf("plan: %w", err)
		}
		picks = filterPicked(picks, picksSoFar)
		l.logger.Info("research plan", "iter", iter, "picks", picks)

		if len(picks) == 0 {
			if !haveBest {
				return Result{Query: q, Draft: "I don't have enough evidence to answer that.", Reason: ReasonUnscored}, nil
			}
			return returnBest(q, best, haveBest, iter-1), nil
		}

		// Gather (parallel)
		gathered := l.gather(ctx, q, picks, bundleSoFar)
		for _, ev := range gathered {
			bundleSoFar.Add(ev)
		}
		for _, p := range picks {
			picksSoFar[p] = struct{}{}
		}

		if bundleSoFar.Len() == 0 {
			return Result{Query: q, Draft: "I don't have enough evidence to answer that.", Reason: ReasonUnscored}, nil
		}

		// Draft
		draftLLM := l.llm
		if l.draftLLM != nil {
			draftLLM = l.draftLLM
		}
		draft, err := draftLLM(ctx, DraftPrompt(q.Question, bundleSoFar))
		if err != nil {
			return Result{}, fmt.Errorf("draft: %w", err)
		}

		// Score
		score, crit := ComputeScore(ctx, l.llm, q.Question, draft, bundleSoFar)
		emitAttempt(l.events, q, iter, score, bundleSoFar, ReasonAccepted)

		result := Result{
			Query: q, Draft: draft, Bundle: bundleSoFar, Score: score, Iterations: iter,
		}
		if !haveBest || score.Composite > best.Score.Composite {
			best = result
			haveBest = true
		}

		if score.Composite >= acceptThreshold {
			best.Reason = ReasonAccepted
			return best, nil
		}

		// Extract ungrounded claims for refinement
		extraNeeds = extraNeeds[:0]
		for _, claim := range crit.Claims {
			if claim.Label == LabelUngrounded {
				extraNeeds = append(extraNeeds, claim.Text)
			}
		}
	}

	return returnBest(q, best, haveBest, iter-1), nil
}

func (l *Loop) plan(ctx context.Context, question string) ([]string, error) {
	prompt := PlanPrompt(question, l.registry.List(), l.hints)
	raw, err := l.llm(ctx, prompt)
	if err != nil {
		return nil, err
	}
	names, err := ParsePlan(raw)
	if err != nil {
		return nil, err
	}
	// Validate against registry
	valid := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := l.registry.Lookup(n); ok {
			valid = append(valid, n)
		}
	}
	return valid, nil
}

func (l *Loop) gather(ctx context.Context, q ResearchQuery, picks []string, prior EvidenceBundle) []Evidence {
	type result struct {
		ev  Evidence
		err error
	}
	ch := make(chan result, len(picks))
	var wg sync.WaitGroup
	for _, name := range picks {
		r, ok := l.registry.Lookup(name)
		if !ok { continue }
		wg.Add(1)
		go func(r Researcher) {
			defer wg.Done()
			ev, err := r.Gather(ctx, q, prior)
			ch <- result{ev: ev, err: err}
		}(r)
	}
	go func() { wg.Wait(); close(ch) }()

	var out []Evidence
	for res := range ch {
		if res.err != nil {
			l.logger.Warn("gather error", "err", res.err)
			continue
		}
		if res.ev.Body != "" {
			out = append(out, res.ev)
		}
	}
	return out
}

func returnBest(q ResearchQuery, best Result, haveBest bool, iters int) Result {
	if !haveBest {
		return Result{Query: q, Draft: "I don't have enough evidence to answer that.", Reason: ReasonBudgetExceeded}
	}
	best.Reason = ReasonBudgetExceeded
	best.Iterations = iters
	return best
}

func filterPicked(picks []string, already map[string]struct{}) []string {
	out := make([]string, 0, len(picks))
	for _, p := range picks {
		if _, done := already[p]; !done {
			out = append(out, p)
		}
	}
	return out
}

func newQueryID() string {
	b := make([]byte, 4)
	cryptoRand.Read(b)
	return fmt.Sprintf("q-%d-%x", time.Now().UnixNano(), b)
}

type discardWriter struct{}
func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
```

- [ ] **Step 8: Run tests**

Run: `go test ./internal/research/...`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add internal/research/
git commit -m "feat: add research loop — plan/gather/draft/critique/score with prompt templates"
```

---

### Task 6: YAML Researcher + ES Researchers

**Files:**
- Create: `internal/research/yaml_researcher.go`
- Create: `internal/research/es_researcher.go`
- Create: `researchers/git-log.yaml`
- Create: `researchers/github-prs.yaml`
- Create: `researchers/github-issues.yaml`

- [ ] **Step 1: Create internal/research/yaml_researcher.go**

```go
package research

import (
	"context"
	"fmt"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
)

// YAMLResearcher wraps a workflow YAML file as a Researcher.
type YAMLResearcher struct {
	workflow *pipeline.Workflow
	reg      *provider.ProviderRegistry
}

// NewYAMLResearcher creates a researcher from a loaded workflow.
func NewYAMLResearcher(w *pipeline.Workflow, reg *provider.ProviderRegistry) *YAMLResearcher {
	return &YAMLResearcher{workflow: w, reg: reg}
}

func (y *YAMLResearcher) Name() string    { return y.workflow.Name }
func (y *YAMLResearcher) Describe() string { return y.workflow.Description }

func (y *YAMLResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	result, err := pipeline.Run(y.workflow, q.Question, "", nil, y.reg)
	if err != nil {
		return Evidence{}, fmt.Errorf("yaml researcher %s: %w", y.workflow.Name, err)
	}
	return Evidence{
		Source: y.workflow.Name,
		Title:  y.workflow.Name,
		Body:   result.Output,
	}, nil
}

// LoadResearchers loads YAML researcher files from a directory and registers them.
func LoadResearchers(dir string, reg *Registry, providerReg *provider.ProviderRegistry) error {
	workflows, err := pipeline.LoadDir(dir)
	if err != nil {
		return nil // directory might not exist, that's ok
	}
	for _, w := range workflows {
		r := NewYAMLResearcher(w, providerReg)
		if err := reg.Register(r); err != nil {
			continue // skip duplicates
		}
	}
	return nil
}
```

- [ ] **Step 2: Create internal/research/es_researcher.go**

```go
package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/esearch"
)

// ESActivityResearcher queries glitch-events for recent activity.
type ESActivityResearcher struct {
	es *esearch.Client
}

func NewESActivityResearcher(es *esearch.Client) *ESActivityResearcher {
	return &ESActivityResearcher{es: es}
}

func (r *ESActivityResearcher) Name() string { return "es-activity" }
func (r *ESActivityResearcher) Describe() string {
	return "indexed activity from git commits, GitHub PRs, issues, and pipeline runs"
}

func (r *ESActivityResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	query := json.RawMessage(fmt.Sprintf(`{
		"size": 20,
		"sort": [{"timestamp": {"order": "desc", "unmapped_type": "date"}}],
		"query": {
			"multi_match": {
				"query": %q,
				"fields": ["message", "body", "summary", "name"],
				"type": "best_fields"
			}
		}
	}`, q.Question))

	indices := []string{esearch.IndexEvents, esearch.IndexPipelines, esearch.IndexSummaries}
	results, err := r.es.Search(ctx, indices, query)
	if err != nil {
		return Evidence{}, fmt.Errorf("es-activity: %w", err)
	}

	var body strings.Builder
	for _, hit := range results.Results {
		fmt.Fprintf(&body, "[%s] %s\n", hit.Index, string(hit.Source))
	}
	return Evidence{
		Source: "es-activity",
		Title:  fmt.Sprintf("%d indexed events", results.Total),
		Body:   body.String(),
	}, nil
}

// ESCodeResearcher queries glitch-code-* for relevant source code.
type ESCodeResearcher struct {
	es *esearch.Client
}

func NewESCodeResearcher(es *esearch.Client) *ESCodeResearcher {
	return &ESCodeResearcher{es: es}
}

func (r *ESCodeResearcher) Name() string { return "es-code" }
func (r *ESCodeResearcher) Describe() string {
	return "indexed source code from repositories — functions, types, and file contents"
}

func (r *ESCodeResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	query := json.RawMessage(fmt.Sprintf(`{
		"size": 10,
		"query": {
			"bool": {
				"should": [
					{"match": {"symbols": %q}},
					{"match": {"content": %q}}
				],
				"minimum_should_match": 1
			}
		}
	}`, q.Question, q.Question))

	// Search all code indices
	results, err := r.es.Search(ctx, []string{"glitch-code-*"}, query)
	if err != nil {
		return Evidence{}, fmt.Errorf("es-code: %w", err)
	}

	var body strings.Builder
	for _, hit := range results.Results {
		fmt.Fprintf(&body, "[%s] %s\n", hit.Index, string(hit.Source))
	}
	return Evidence{
		Source: "es-code",
		Title:  fmt.Sprintf("%d code chunks", results.Total),
		Body:   body.String(),
	}, nil
}
```

- [ ] **Step 3: Create researcher YAML files**

Create `researchers/git-log.yaml`:
```yaml
name: git-log
description: recent commit history for the current repo

steps:
  - id: gather
    run: git log --oneline -50
```

Create `researchers/github-prs.yaml`:
```yaml
name: github-prs
description: open and recently merged pull requests

steps:
  - id: gather
    run: gh pr list --limit 20 --state all --json number,title,state,author,updatedAt,baseRefName
```

Create `researchers/github-issues.yaml`:
```yaml
name: github-issues
description: open issues and recently closed issues

steps:
  - id: gather
    run: gh issue list --limit 20 --state all --json number,title,state,author,updatedAt,labels
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/research/...`
Expected: PASS (new files compile, existing tests still pass)

- [ ] **Step 5: Commit**

```bash
git add internal/research/yaml_researcher.go internal/research/es_researcher.go researchers/
git commit -m "feat: add YAML and ES researchers for the research loop"
```

---

### Task 7: `internal/observer` — Query Engine

**Files:**
- Create: `internal/observer/query.go`
- Create: `internal/observer/query_test.go`

- [ ] **Step 1: Write query_test.go**

```go
package observer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/8op-org/gl1tch/internal/esearch"
)

func TestDefaultQuery(t *testing.T) {
	q := defaultQuery("what broke today")
	data, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty query")
	}
}

func TestAnswerWithMockES(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{"hits":{"total":{"value":1},"hits":[{"_index":"glitch-events","_source":{"message":"fix auth bug","source":"git","timestamp":"2026-04-13T10:00:00Z"}}]}}`
		w.Write([]byte(resp))
	}))
	defer srv.Close()

	es := esearch.NewClient(srv.URL)
	// Can't test full Answer without Ollama, but we can test search path
	_ = es
}

func TestExtractJSON(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`{"query":{}}`, `{"query":{}}`},
		{"```json\n{\"query\":{}}\n```", `{"query":{}}`},
		{`some text {"query":{}} more text`, `{"query":{}}`},
	}
	for _, tc := range cases {
		got := extractJSON(tc.input)
		if got != tc.want {
			t.Errorf("extractJSON(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Create internal/observer/query.go**

```go
// Package observer implements the query engine that bridges natural language
// questions to Elasticsearch and synthesizes answers via Ollama.
package observer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
)

const defaultModel = "qwen2.5:7b"

// QueryEngine translates natural language → ES query → synthesized answer.
type QueryEngine struct {
	es    *esearch.Client
	model string
}

func NewQueryEngine(es *esearch.Client, model string) *QueryEngine {
	if model == "" {
		model = defaultModel
	}
	return &QueryEngine{es: es, model: model}
}

func allIndices() []string {
	return []string{
		esearch.IndexEvents,
		esearch.IndexSummaries,
		esearch.IndexPipelines,
		esearch.IndexInsights,
	}
}

// Answer processes a question and returns a synthesized answer.
func (q *QueryEngine) Answer(ctx context.Context, question string) (string, error) {
	results, err := q.searchWithFallback(ctx, question)
	if err != nil {
		return "", fmt.Errorf("search: %w", err)
	}
	if results == nil || results.Total == 0 {
		return "I don't have any indexed data matching that question yet. Make sure ES is running and data has been indexed.", nil
	}
	return q.synthesize(ctx, question, results)
}

func (q *QueryEngine) searchWithFallback(ctx context.Context, question string) (*esearch.SearchResponse, error) {
	esQuery, err := q.generateQuery(ctx, question)
	if err != nil {
		esQuery = defaultQuery(question)
	}
	queryJSON, _ := json.Marshal(esQuery)
	results, err := q.es.Search(ctx, allIndices(), queryJSON)
	if err != nil {
		fallbackJSON, _ := json.Marshal(defaultQuery(question))
		results, err = q.es.Search(ctx, allIndices(), fallbackJSON)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func defaultQuery(question string) map[string]any {
	return map[string]any{
		"size": 20,
		"sort": []map[string]any{{"timestamp": map[string]any{"order": "desc", "unmapped_type": "date"}}},
		"query": map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{"multi_match": map[string]any{
						"query":  question,
						"fields": []string{"message", "body", "summary", "pattern", "stdout", "name"},
						"type":   "best_fields",
					}},
				},
				"minimum_should_match": 1,
			},
		},
	}
}

func (q *QueryEngine) generateQuery(ctx context.Context, question string) (map[string]any, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	weekAgo := now.Add(-7 * 24 * time.Hour).Format("2006-01-02")

	prompt := fmt.Sprintf(`You are an Elasticsearch query generator. Return ONLY valid JSON.

Indices: glitch-events (type, source, repo, author, message, body, timestamp), glitch-summaries (scope, date, summary, timestamp), glitch-pipelines (name, status, stdout, model, provider, timestamp), glitch-insights (type, pattern, recommendation, timestamp)

Today: %s  Week ago: %s

Question: "%s"

Rules:
- Include: "sort": [{"timestamp": {"order": "desc", "unmapped_type": "date"}}]
- Default size: 20
- Use multi_match for free-text search

JSON:`, today, weekAgo, question)

	resp, err := ollamaGenerate(ctx, q.model, prompt)
	if err != nil {
		return nil, err
	}
	jsonStr := extractJSON(resp)
	var query map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &query); err != nil {
		return nil, err
	}
	return query, nil
}

func (q *QueryEngine) synthesize(ctx context.Context, question string, results *esearch.SearchResponse) (string, error) {
	context := formatResults(results)
	prompt := fmt.Sprintf(`Based on the following indexed data, answer the question concisely.

Question: %s

Observed data (%d results):
%s

Rules:
- Be direct and specific — cite repos, commits, timestamps
- If data doesn't contain relevant info, say "I don't have data on that"
- Only reference information in the observed data above
- Be concise`, question, results.Total, context)

	return ollamaGenerate(ctx, q.model, prompt)
}

func formatResults(results *esearch.SearchResponse) string {
	if results == nil || len(results.Results) == 0 {
		return "(no results found)"
	}
	var sb strings.Builder
	for i, r := range results.Results {
		if i >= 15 { break }
		fmt.Fprintf(&sb, "\n[%s] %s\n", r.Index, string(r.Source))
	}
	return sb.String()
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.SplitN(s, "\n", 2)
		if len(lines) > 1 { s = lines[1] }
		if idx := strings.LastIndex(s, "```"); idx >= 0 { s = s[:idx] }
	}
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func ollamaGenerate(ctx context.Context, model, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model": model, "prompt": prompt, "stream": false,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11434/api/generate", bytes.NewReader(body))
	if err != nil { return "", err }
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return "", fmt.Errorf("ollama unavailable: %w", err) }
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama %d: %s", resp.StatusCode, b)
	}
	var result struct{ Response string `json:"response"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Response, nil
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/observer/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/observer/
git commit -m "feat: add observer query engine — NL to ES to synthesized answer"
```

---

### Task 8: Wire It All Together — `cmd/ask.go` Fallback Chain + `cmd/observe.go`

**Files:**
- Modify: `cmd/ask.go`
- Create: `cmd/observe.go`
- Create: `cmd/research_helpers.go`

- [ ] **Step 1: Create cmd/research_helpers.go**

```go
package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/research"
	"github.com/8op-org/gl1tch/internal/store"
)

// buildResearchLoop assembles the research loop with all available researchers.
// Returns nil if no researchers could be loaded.
func buildResearchLoop() *research.Loop {
	reg := research.NewRegistry()

	// 1. Load YAML researchers from ~/.config/glitch/researchers/
	if home, err := os.UserHomeDir(); err == nil {
		researcherDir := filepath.Join(home, ".config", "glitch", "researchers")
		research.LoadResearchers(researcherDir, reg, providerReg)
	}

	// 2. Load YAML researchers from .glitch/researchers/ (project-local)
	research.LoadResearchers(".glitch/researchers", reg, providerReg)

	// 3. Add ES researchers if ES is reachable
	es := esearch.NewClient("http://localhost:9200")
	if err := es.Ping(context.Background()); err == nil {
		reg.Register(research.NewESActivityResearcher(es))
		reg.Register(research.NewESCodeResearcher(es))
	}

	if len(reg.Names()) == 0 {
		return nil
	}

	llm := func(ctx context.Context, prompt string) (string, error) {
		return providerRunOllama("qwen2.5:7b", prompt)
	}
	loop := research.NewLoop(reg, llm)

	// Wire hints from store if available
	if st, err := store.Open(); err == nil {
		defer st.Close()
		events, _ := st.SimilarResearchEvents("", 0)
		_ = events // hints will be wired in a future iteration
	}

	return loop
}

// providerRunOllama wraps the existing provider.RunOllama to match LLMFn signature.
func providerRunOllama(model, prompt string) (string, error) {
	return providerPkg.RunOllama(model, prompt)
}
```

Wait — `providerPkg` doesn't exist. The current code uses `provider.RunOllama` as a package-level function. Let me check.

Actually, looking at the current code: `provider.RunOllama(model, prompt)` is already a package-level function. The research_helpers.go should just use that directly.

- [ ] **Step 1 (revised): Create cmd/research_helpers.go**

```go
package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
)

// buildResearchLoop assembles the research loop with all available researchers.
// Returns nil if no researchers could be loaded.
func buildResearchLoop() *research.Loop {
	reg := research.NewRegistry()

	// 1. Load YAML researchers from ~/.config/glitch/researchers/
	if home, err := os.UserHomeDir(); err == nil {
		researcherDir := filepath.Join(home, ".config", "glitch", "researchers")
		research.LoadResearchers(researcherDir, reg, providerReg)
	}

	// 2. Load YAML researchers from .glitch/researchers/ (project-local)
	research.LoadResearchers(".glitch/researchers", reg, providerReg)

	// 3. Add ES researchers if ES is reachable
	es := esearch.NewClient("http://localhost:9200")
	if err := es.Ping(context.Background()); err == nil {
		reg.Register(research.NewESActivityResearcher(es))
		reg.Register(research.NewESCodeResearcher(es))
	}

	if len(reg.Names()) == 0 {
		return nil
	}

	llm := func(ctx context.Context, prompt string) (string, error) {
		return provider.RunOllama("qwen2.5:7b", prompt)
	}

	loop := research.NewLoop(reg, llm)
	return loop
}
```

- [ ] **Step 2: Modify cmd/ask.go — add fallback chain**

Replace the existing `askCmd` RunE with the three-tier fallback:

```go
// In cmd/ask.go, replace the RunE function body with:

var askCmd = &cobra.Command{
	Use:   "ask [input]",
	Short: "route a question or URL to the best workflow",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetPath != "" {
			if err := os.Chdir(targetPath); err != nil {
				return fmt.Errorf("chdir %s: %w", targetPath, err)
			}
		}

		input := strings.Join(args, " ")
		ctx := cmd.Context()

		// Tier 1: workflow routing
		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}

		w, resolved, params := router.Match(input, workflows, "")
		if w != nil {
			fmt.Printf(">> %s\n", w.Name)
			result, err := pipeline.Run(w, resolved, "", params, providerReg)
			if err != nil {
				return err
			}
			fmt.Println(result.Output)
			return nil
		}

		// Tier 2: research loop
		if loop := buildResearchLoop(); loop != nil {
			fmt.Fprintln(os.Stderr, ">> researching...")
			result, err := loop.Run(ctx, research.ResearchQuery{Question: input}, research.DefaultBudget())
			if err == nil && result.Draft != "" {
				fmt.Println(result.Draft)
				return nil
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "research: %v\n", err)
			}
		}

		// Tier 3: one-shot fallback
		fmt.Fprintln(os.Stderr, ">> asking ollama...")
		answer, err := provider.RunOllama("qwen2.5:7b", input)
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	},
}
```

The import block in ask.go needs to add `research` and `provider`:

```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
	"github.com/8op-org/gl1tch/internal/router"
)
```

- [ ] **Step 3: Create cmd/observe.go**

```go
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/observer"
)

func init() {
	rootCmd.AddCommand(observeCmd)
}

var observeCmd = &cobra.Command{
	Use:   "observe [question]",
	Short: "query indexed activity via elasticsearch",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := strings.Join(args, " ")

		es := esearch.NewClient("http://localhost:9200")
		if err := es.Ping(context.Background()); err != nil {
			return fmt.Errorf("elasticsearch is not running — start with: glitch up")
		}

		engine := observer.NewQueryEngine(es, "")
		answer, err := engine.Answer(cmd.Context(), question)
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	},
}
```

- [ ] **Step 4: Build and verify**

Run: `go build -o glitch .`
Expected: Compiles successfully.

Run: `./glitch ask --help`
Expected: Shows help text.

Run: `./glitch observe --help`
Expected: Shows help text.

- [ ] **Step 5: Commit**

```bash
git add cmd/ask.go cmd/observe.go cmd/research_helpers.go
git commit -m "feat: wire three-tier fallback chain in ask + add observe command"
```

---

### Task 9: Refactor Indexer to Use `esearch.Client`

**Files:**
- Modify: `internal/indexer/indexer.go`

- [ ] **Step 1: Read current indexer.go to identify raw HTTP calls**

The current indexer has its own `ensureIndex`, `bulkIndex`, and HTTP client code. Replace those with calls to `esearch.Client`.

- [ ] **Step 2: Refactor indexer.go**

Update `IndexRepo` to accept an `*esearch.Client` parameter instead of a raw `esURL` string. Replace the internal `ensureIndex` and `bulkIndex` functions with `client.EnsureIndex` and `client.BulkIndex`.

Key changes:
- Change `IndexRepo(root, esURL string)` to `IndexRepo(root string, es *esearch.Client)`
- Remove the internal `ensureIndex` and `bulkIndex` functions
- Use `es.EnsureIndex` and `es.BulkIndex` instead
- Update `cmd/index.go` to create an `esearch.Client` and pass it

- [ ] **Step 3: Update cmd/index.go**

```go
// In cmd/index.go, update the RunE to:
esClient := esearch.NewClient("http://localhost:9200")
return indexer.IndexRepo(path, esClient)
```

- [ ] **Step 4: Run existing indexer tests**

Run: `go test ./internal/indexer/...`
Expected: PASS (tests that don't hit ES should still pass; HTTP-dependent tests may need mock updates)

- [ ] **Step 5: Commit**

```bash
git add internal/indexer/ cmd/index.go
git commit -m "refactor: migrate indexer to use shared esearch.Client"
```

---

### Task 10: Integration Smoke Test

- [ ] **Step 1: Ensure ES is running**

Run: `./glitch up`
Wait for: `curl -s http://localhost:9200` returns cluster info.

- [ ] **Step 2: Index a repo**

Run: `./glitch index .`
Expected: "Done: N files processed, M chunks indexed"

- [ ] **Step 3: Test the observe command**

Run: `./glitch observe "what code is in this repo?"`
Expected: Synthesized answer from ES code index.

- [ ] **Step 4: Test ask with workflow routing (tier 1)**

Run: `./glitch ask "https://github.com/8op-org/gl1tch/pull/1"`
Expected: Routes to github-pr-review workflow.

- [ ] **Step 5: Test ask with research loop (tier 2)**

First, copy default researchers to config:
```bash
mkdir -p ~/.config/glitch/researchers
cp researchers/*.yaml ~/.config/glitch/researchers/
```

Run: `./glitch ask "what recent commits have been made?"`
Expected: ">> researching..." then a grounded answer citing actual commits.

- [ ] **Step 6: Test ask one-shot fallback (tier 3)**

With no researchers loaded and no workflow match:
Run: `./glitch ask "what is a goroutine?"`
Expected: ">> asking ollama..." then a general answer.

- [ ] **Step 7: Commit final state**

```bash
git add -A
git commit -m "feat: daily driver integration — ES + research loop + observe working end-to-end"
```
