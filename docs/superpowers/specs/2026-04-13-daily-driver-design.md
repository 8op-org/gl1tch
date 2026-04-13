# gl1tch Daily Driver: ES + Research Loop + Store

**Date:** 2026-04-13
**Status:** Approved
**Scope:** Code search + activity observer (Option A)
**Approach:** Port & adapt from legacy (Approach 1), workflow-backed researchers

---

## Goal

Make `glitch ask` a daily driver by adding a research loop that gathers real evidence before answering, backed by Elasticsearch for persistent activity history and SQLite for run tracking and learning hints.

The user types one command. The backend decides whether to route to a workflow, run a research loop, or fall back to a one-shot LLM call. Minimal surface, powerful backend.

## Architecture

```
glitch ask "what broke in ensemble this week?"

┌─────────────────────────────────────────────────┐
│ cmd/ask.go — fallback chain                     │
│                                                 │
│  1. router.Match() → workflow found? run it.    │
│  2. research.Loop.Run() → gather evidence,      │
│     draft, critique, score → grounded answer    │
│  3. provider.RunOllama() → one-shot fallback    │
└─────────────────────────────────────────────────┘
         │                    │
         ▼                    ▼
┌──────────────┐    ┌──────────────────┐
│ pipeline/    │    │ research/        │
│ (existing)   │    │ Loop + Registry  │
│ runs workflows│   │ + Researchers    │
└──────────────┘    └──────────────────┘
                         │         │
                    ┌────┘         └────┐
                    ▼                   ▼
            ┌─────────────┐    ┌──────────────┐
            │ YAML         │    │ esearch/     │
            │ Researchers  │    │ ES Client    │
            │ (git, gh)   │    │ ~5 indices   │
            └─────────────┘    └──────────────┘
                                      │
                               ┌──────┴──────┐
                               │ Elasticsearch│
                               │ (Docker)     │
                               └─────────────┘

            ┌─────────────┐
            │ store/       │
            │ SQLite       │
            │ runs + hints │
            └─────────────┘
```

## New Packages

### `internal/esearch` — ES Client

Thin HTTP client, no `go-elasticsearch` SDK dependency. Matches the existing indexer's raw HTTP approach.

```go
type Client struct {
    baseURL string       // default http://localhost:9200
    http    *http.Client
}

func (c *Client) Search(ctx context.Context, indices []string, query json.RawMessage) (*SearchResponse, error)
func (c *Client) BulkIndex(ctx context.Context, index string, docs []BulkDoc) error
func (c *Client) EnsureIndex(ctx context.Context, index, mapping string) error
func (c *Client) Ping(ctx context.Context) error
```

**Indices (5):**

| Index | Contents | Writer |
|-------|----------|--------|
| `glitch-code-<repo>` | Chunked source code + symbols | `glitch index` |
| `glitch-events` | Git commits, GitHub PRs/issues/reviews | ES researchers during gather |
| `glitch-summaries` | Synthesized observer answers | Observer after answering |
| `glitch-pipelines` | Workflow run results | Pipeline runner on completion |
| `glitch-insights` | Research loop drafts + scores | Research loop on completion |

**Graceful degradation:** If ES is unreachable, `Ping()` fails, ES researchers are excluded from the registry, and the observe command returns a clear "ES offline" message. Shell researchers + workflow routing still work.

### `internal/research` — Research Loop

Ported from legacy. 5-stage cycle: plan → gather → draft → critique → score.

**Core types:**

```go
type LLMFn func(ctx context.Context, prompt string) (string, error)

type Researcher interface {
    Name() string
    Describe() string
    Gather(ctx context.Context, q ResearchQuery, prior EvidenceBundle) (Evidence, error)
}

type Registry struct { /* concurrent-safe map[string]Researcher */ }

type Loop struct {
    registry *Registry
    llm      LLMFn
    draftLLM LLMFn          // optional separate provider for drafts
    hints    HintsProvider   // biases planner from past successes
    events   EventSink       // JSONL log for learning
}
```

**The cycle:**

1. **Plan** — Local LLM (qwen2.5:7b) picks researchers from the registry menu. Hints from past successful combos bias the pick.
2. **Gather** — Selected researchers run in parallel. YAML researchers via `pipeline.Run()`, ES researchers via `esearch.Client`.
3. **Draft** — LLM writes answer grounded in evidence. Forbidden from inventing identifiers.
4. **Critique** — LLM labels each claim as grounded/partial/ungrounded.
5. **Score** — Composite confidence from evidence coverage + cross-source agreement + critique ratio. If below threshold and budget remains, refine with ungrounded claims fed back to planner.

**Budget:** Max 3 iterations, 60s wallclock.

**Researchers are workflow-backed YAML files:**

Located in `~/.config/glitch/researchers/`. Loaded the same way `loadWorkflows()` works.

```yaml
# researchers/git-log.yaml
name: git-log
description: recent commit history for the current repo
steps:
  - id: gather
    run: git log --oneline -50
```

```yaml
# researchers/github-prs.yaml
name: github-prs
description: open and recently merged pull requests
steps:
  - id: gather
    run: gh pr list --limit 20 --json number,title,state,author,updatedAt,baseRefName
```

Only ES researchers (`es-activity`, `es-code`) are Go-native — they need the `esearch.Client` for query construction. Everything else is YAML.

**No performance difference vs legacy.** Legacy ran shell researchers via `exec.Command("sh", "-c", ...)` — same as `pipeline.Run()` under the hood. YAML template parsing adds ~1ms, invisible next to 500ms+ `gh` CLI calls.

**Wiring to current provider system:**

```go
llmFn := func(ctx context.Context, prompt string) (string, error) {
    return provider.RunOllama("qwen2.5:7b", prompt)
}
loop := research.NewLoop(registry, llmFn)
```

### `internal/store` — SQLite

Run history + research event hints. No migrations pre-1.0.

**Location:** `~/.local/share/glitch/glitch.db`

**Schema (3 tables):**

```sql
CREATE TABLE runs (
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

CREATE TABLE steps (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id      INTEGER NOT NULL,
  step_id     TEXT NOT NULL,
  prompt      TEXT,
  output      TEXT,
  model       TEXT,
  duration_ms INTEGER,
  UNIQUE(run_id, step_id)
);

CREATE TABLE research_events (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  query_id        TEXT NOT NULL,
  question        TEXT NOT NULL,
  researchers     TEXT NOT NULL,
  composite_score REAL,
  reason          TEXT,
  created_at      INTEGER NOT NULL
);
```

**Writer interface (write-only for callers):**

```go
type Writer interface {
    RecordRun(kind, name, input string) (runID int64, err error)
    FinishRun(runID int64, output string, exitStatus int) error
    RecordStep(runID int64, stepID, prompt, output, model string, durationMs int64) error
    RecordResearchEvent(evt ResearchEvent) error
}
```

**Hints provider:** Reads `research_events`, finds past questions with token jaccard similarity, returns researcher combos that scored highest. <100ms — one SQLite query + string comparison.

### `internal/observer` — Query Engine

Bridges natural language → ES query → synthesized answer.

```go
type QueryEngine struct {
    es    *esearch.Client
    model string
}

func (q *QueryEngine) Answer(ctx context.Context, question string) (string, error)
```

Flow: generate ES query via LLM → search → fallback to default query on error → synthesize answer from results via LLM.

## Modified Packages

### `cmd/` — New Commands

**`ask.go` — fallback chain:**

```
Tier 1: router.Match() → workflow
Tier 2: research.Loop.Run() → grounded answer
Tier 3: provider.RunOllama() → one-shot
```

`buildResearchLoop()` loads researcher YAMLs, adds ES researchers if ES reachable, wires LLMFn + hints, returns nil if no researchers available.

**`observe.go` — direct ES query:**

```
glitch observe "what happened today?"
```

Skips routing and research. Straight to observer.QueryEngine.Answer().

**`up.go` / `down.go` — Docker lifecycle:**

```
glitch up    — docker compose up -d
glitch down  — docker compose down
```

Uses bundled docker-compose.yml.

### `internal/indexer` — Shared ES Client

Refactored to use `esearch.Client` instead of its own raw HTTP calls. Same functionality, shared connection.

## Docker Compose

ES + Kibana only. No APM server.

```yaml
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.17.0
    container_name: glitch-es
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - ES_JAVA_OPTS=-Xms512m -Xmx512m
    ports:
      - "9200:9200"
    volumes:
      - glitch-es-data:/usr/share/elasticsearch/data

  kibana:
    image: docker.elastic.co/kibana/kibana:8.17.0
    container_name: glitch-kibana
    depends_on:
      elasticsearch:
        condition: service_healthy
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
      - XPACK_SECURITY_ENABLED=false
    ports:
      - "5601:5601"
    volumes:
      - glitch-kibana-data:/usr/share/kibana/data

volumes:
  glitch-es-data:
  glitch-kibana-data:
```

## File Layout (New/Modified)

```
gl1tch/
├── docker-compose.yml              (new)
├── cmd/
│   ├── ask.go                      (modified — fallback chain)
│   ├── observe.go                  (new)
│   ├── up.go                       (new)
│   └── ...
├── internal/
│   ├── esearch/
│   │   ├── client.go               (new — HTTP client)
│   │   ├── mappings.go             (new — index schemas)
│   │   └── client_test.go          (new)
│   ├── research/
│   │   ├── loop.go                 (ported — 5-stage engine)
│   │   ├── types.go                (ported — Query, Evidence, Score, Result)
│   │   ├── registry.go             (ported — researcher registry)
│   │   ├── researcher.go           (ported — interface)
│   │   ├── yaml_researcher.go      (new — workflow-backed researcher)
│   │   ├── es_researcher.go        (new — ES activity + code query)
│   │   ├── score.go                (ported — composite scoring)
│   │   ├── hints.go                (ported — jaccard hint provider)
│   │   ├── prompts.go              (ported — plan/draft/critique templates)
│   │   ├── events.go               (ported — event sink for learning)
│   │   └── loop_test.go            (ported + new)
│   ├── store/
│   │   ├── store.go                (new — open/close/writer)
│   │   ├── schema.go               (new — 3 tables, no migrations)
│   │   └── store_test.go           (new)
│   ├── observer/
│   │   ├── query.go                (ported — natural language → ES → answer)
│   │   └── query_test.go           (new)
│   ├── indexer/
│   │   └── indexer.go              (modified — use esearch.Client)
│   └── ...
└── researchers/                     (new — default YAML researchers, copied to ~/.config/glitch/researchers/)
    ├── git-log.yaml
    ├── github-prs.yaml
    └── github-issues.yaml
```

## What's NOT Included

- Game scoring, XP, achievements, streaks
- Message bus (busd)
- OpenTelemetry / APM
- Desktop/TUI interception
- Plugin manager
- Workspaces / multi-workspace scoping
- Chat history indexing
- Brain RAG / vector embeddings
- Attention classification system
- Clarification system
- Drafts / prompt builder

These can be added later as workflows or Go packages. The architecture doesn't preclude them.
