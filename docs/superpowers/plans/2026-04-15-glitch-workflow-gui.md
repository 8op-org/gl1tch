# glitch workflow gui — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `glitch workflow gui` command that starts a local web server for browsing, editing, running, and reviewing workflows in any gl1tch workspace.

**Architecture:** Go HTTP server (`internal/gui`) serves a Svelte SPA (`gui/`) via `go:embed` in prod or Vite proxy in dev. REST API imports `internal/pipeline` and `internal/store` directly. Kibana iframes embedded for per-run and per-workflow telemetry.

**Tech Stack:** Go stdlib `net/http`, Svelte 5, Vite, CodeMirror 6, marked, highlight.js, custom Lezer grammar for sexpr.

---

## File Structure

```
cmd/gui.go                      Cobra command — starts HTTP server
internal/gui/
  server.go                     Router, static file serving, go:embed
  api_workflows.go              GET/PUT /api/workflows, POST run
  api_runs.go                   GET /api/runs, GET /api/runs/:id
  api_results.go                GET /api/results/*path
  api_kibana.go                 GET /api/kibana/workflow/:name, /api/kibana/run/:id
gui/
  package.json                  Svelte + Vite + CodeMirror deps
  vite.config.js                Dev proxy to Go API
  src/
    App.svelte                  Router shell
    lib/
      api.js                    Fetch wrapper for /api/*
      sexpr.grammar             Lezer grammar for sexpr
      sexpr-lang.js             CodeMirror language support from grammar
      markdown.js               marked + highlight.js renderer
    routes/
      WorkflowList.svelte       Sidebar + workflow listing
      Editor.svelte             CodeMirror editor + save + run button
      RunDialog.svelte           Modal for --set params
      RunView.svelte            Run detail + step status + Kibana embed
      ResultsBrowser.svelte     File tree + rendered content
    app.css                     Dark theme styles
  index.html                    SPA entry point
```

---

### Task 1: Scaffold Svelte frontend with Vite

**Files:**
- Create: `gui/package.json`
- Create: `gui/vite.config.js`
- Create: `gui/index.html`
- Create: `gui/src/App.svelte`
- Create: `gui/src/app.css`

- [ ] **Step 1: Initialize gui/ directory**

```bash
cd ~/Projects/gl1tch
mkdir -p gui/src/lib gui/src/routes
```

- [ ] **Step 2: Create package.json**

```json
{
  "name": "glitch-gui",
  "private": true,
  "version": "0.0.1",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^5.0.0",
    "svelte": "^5.0.0",
    "vite": "^6.0.0"
  },
  "dependencies": {
    "@codemirror/lang-javascript": "^6.2.0",
    "@codemirror/state": "^6.5.0",
    "@codemirror/view": "^6.35.0",
    "@codemirror/language": "^6.10.0",
    "@lezer/generator": "^1.7.0",
    "@lezer/highlight": "^1.2.0",
    "@lezer/lr": "^1.4.0",
    "codemirror": "^6.0.0",
    "highlight.js": "^11.10.0",
    "marked": "^15.0.0",
    "svelte-spa-router": "^4.0.0"
  }
}
```

- [ ] **Step 3: Create vite.config.js with API proxy**

```js
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8374',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
```

- [ ] **Step 4: Create index.html entry point**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>gl1tch</title>
  <link rel="stylesheet" href="/src/app.css" />
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>
```

- [ ] **Step 5: Create src/main.js**

```js
import App from './App.svelte'

const app = new App({
  target: document.getElementById('app'),
})

export default app
```

- [ ] **Step 6: Create App.svelte with hash router**

```svelte
<script>
  import Router from 'svelte-spa-router'
  import WorkflowList from './routes/WorkflowList.svelte'
  import Editor from './routes/Editor.svelte'
  import RunView from './routes/RunView.svelte'
  import ResultsBrowser from './routes/ResultsBrowser.svelte'

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': Editor,
    '/run/:id': RunView,
    '/results': ResultsBrowser,
  }
</script>

<div class="shell">
  <nav class="topnav">
    <a href="#/">gl1tch</a>
    <a href="#/">Workflows</a>
    <a href="#/results">Results</a>
  </nav>
  <main>
    <Router {routes} />
  </main>
</div>
```

- [ ] **Step 7: Create app.css with dark theme**

```css
:root {
  --bg: #0d1117;
  --bg-surface: #161b22;
  --bg-hover: #1c2128;
  --border: #30363d;
  --text: #e6edf3;
  --text-muted: #8b949e;
  --accent: #58a6ff;
  --accent-hover: #79c0ff;
  --success: #3fb950;
  --danger: #f85149;
  --font-mono: 'SF Mono', 'Fira Code', monospace;
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-sans);
  font-size: 14px;
  line-height: 1.5;
}

.shell { display: flex; flex-direction: column; height: 100vh; }

.topnav {
  display: flex;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border);
}

.topnav a {
  color: var(--text-muted);
  text-decoration: none;
  font-size: 13px;
}
.topnav a:first-child {
  color: var(--accent);
  font-weight: 600;
  font-family: var(--font-mono);
  margin-right: 1rem;
}
.topnav a:hover { color: var(--text); }

main { flex: 1; overflow: auto; padding: 1rem; }

button {
  background: var(--bg-surface);
  color: var(--text);
  border: 1px solid var(--border);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}
button:hover { background: var(--bg-hover); }
button.primary { background: var(--accent); color: #000; border-color: var(--accent); }
button.primary:hover { background: var(--accent-hover); }

pre, code {
  font-family: var(--font-mono);
  font-size: 13px;
}
```

- [ ] **Step 8: Create placeholder route components**

Create minimal placeholder files for each route so the app compiles:

`gui/src/routes/WorkflowList.svelte`:
```svelte
<p>Workflows loading...</p>
```

`gui/src/routes/Editor.svelte`:
```svelte
<script>
  export let params = {}
</script>
<p>Editor: {params.name}</p>
```

`gui/src/routes/RunView.svelte`:
```svelte
<script>
  export let params = {}
</script>
<p>Run: {params.id}</p>
```

`gui/src/routes/ResultsBrowser.svelte`:
```svelte
<p>Results loading...</p>
```

`gui/src/routes/RunDialog.svelte`:
```svelte
<p>Run dialog placeholder</p>
```

- [ ] **Step 9: Verify frontend builds**

```bash
cd gui && npm install && npm run build
```

Expected: `gui/dist/` created with bundled assets.

- [ ] **Step 10: Commit**

```bash
git add gui/
git commit -m "feat(gui): scaffold Svelte + Vite frontend"
```

---

### Task 2: Go HTTP server with go:embed and dev mode

**Files:**
- Create: `internal/gui/server.go`
- Create: `cmd/gui.go`
- Modify: `cmd/workflow.go:19-24` (add gui subcommand)
- Modify: `.gitignore`

- [ ] **Step 1: Create internal/gui/server.go**

```go
package gui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/store"
)

//go:embed all:dist
var distFS embed.FS

// Server is the workflow GUI HTTP server.
type Server struct {
	addr      string
	workspace string
	dev       bool
	store     *store.Store
	mux       *http.ServeMux
}

// New creates a GUI server for the given workspace.
func New(addr, workspace string, dev bool) (*Server, error) {
	st, err := store.Open()
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	s := &Server{
		addr:      addr,
		workspace: workspace,
		dev:       dev,
		store:     st,
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/workflows", s.handleListWorkflows)
	s.mux.HandleFunc("GET /api/workflows/{name}", s.handleGetWorkflow)
	s.mux.HandleFunc("PUT /api/workflows/{name}", s.handlePutWorkflow)
	s.mux.HandleFunc("POST /api/workflows/{name}/run", s.handleRunWorkflow)
	s.mux.HandleFunc("GET /api/runs", s.handleListRuns)
	s.mux.HandleFunc("GET /api/runs/{id}", s.handleGetRun)
	s.mux.HandleFunc("GET /api/results/{path...}", s.handleGetResult)
	s.mux.HandleFunc("GET /api/kibana/workflow/{name}", s.handleKibanaWorkflow)
	s.mux.HandleFunc("GET /api/kibana/run/{id}", s.handleKibanaRun)

	if !s.dev {
		// Serve embedded frontend
		dist, _ := fs.Sub(distFS, "dist")
		fileServer := http.FileServer(http.FS(dist))
		s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			// SPA fallback: serve index.html for non-file paths
			if r.URL.Path != "/" {
				f, err := dist.(fs.ReadFileFS).ReadFile(r.URL.Path[1:])
				if err != nil {
					// Not a static file — serve index.html for SPA routing
					r.URL.Path = "/"
				}
				_ = f
			}
			fileServer.ServeHTTP(w, r)
		})
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	return srv.ListenAndServe()
}

// Close releases server resources.
func (s *Server) Close() error {
	return s.store.Close()
}

// workflowsDir returns the path to the workspace workflows directory.
func (s *Server) workflowsDir() string {
	return s.workspace + "/workflows"
}

// resultsDir returns the path to the workspace results directory.
func (s *Server) resultsDir() string {
	return s.workspace + "/results"
}
```

Note: The `//go:embed all:dist` directive requires `gui/dist/` to be copied into `internal/gui/dist/` at build time. We'll add a build script for this in Task 8.

- [ ] **Step 2: Create cmd/gui.go**

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/gui"
)

var (
	guiPort int
	guiDev  bool
)

func init() {
	guiCmd.Flags().IntVar(&guiPort, "port", 8374, "port to listen on")
	guiCmd.Flags().BoolVar(&guiDev, "dev", false, "dev mode (proxy to Vite on :5173)")
	workflowCmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "start the workflow management web UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws := workspacePath
		if ws == "" {
			var err error
			ws, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		ws, _ = filepath.Abs(ws)

		addr := fmt.Sprintf("127.0.0.1:%d", guiPort)
		srv, err := gui.New(addr, ws, guiDev)
		if err != nil {
			return err
		}
		defer srv.Close()

		fmt.Printf(">> gl1tch gui: http://%s\n", addr)
		if guiDev {
			fmt.Println(">> dev mode: frontend at http://127.0.0.1:5173")
		}
		return srv.ListenAndServe()
	},
}
```

- [ ] **Step 3: Add gui/dist/ and gui/node_modules/ to .gitignore**

Append to `.gitignore`:
```
gui/node_modules/
gui/dist/
internal/gui/dist/
```

- [ ] **Step 4: Verify it compiles**

```bash
cd ~/Projects/gl1tch && go build .
```

Expected: compiles (API handlers will be stubs returning 501 for now).

- [ ] **Step 5: Commit**

```bash
git add cmd/gui.go internal/gui/server.go .gitignore
git commit -m "feat(gui): add Go HTTP server skeleton with go:embed"
```

---

### Task 3: Workflow API endpoints

**Files:**
- Create: `internal/gui/api_workflows.go`

- [ ] **Step 1: Write test for listing workflows**

Create `internal/gui/api_workflows_test.go`:

```go
package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestListWorkflows(t *testing.T) {
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	os.MkdirAll(wfDir, 0o755)

	// Write a test workflow
	os.WriteFile(filepath.Join(wfDir, "hello.glitch"), []byte(`(workflow
  (name "hello")
  (description "test workflow")
  (step (id "greet") (run "echo hi"))
)`), 0o644)

	srv := &Server{workspace: dir, mux: http.NewServeMux()}
	srv.routes()

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var workflows []map[string]string
	json.Unmarshal(w.Body.Bytes(), &workflows)
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	if workflows[0]["name"] != "hello" {
		t.Errorf("expected name 'hello', got %q", workflows[0]["name"])
	}
}

func TestGetWorkflow(t *testing.T) {
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	os.MkdirAll(wfDir, 0o755)

	src := `(workflow
  (name "hello")
  (step (id "greet") (run "echo hi"))
)`
	os.WriteFile(filepath.Join(wfDir, "hello.glitch"), []byte(src), 0o644)

	srv := &Server{workspace: dir, mux: http.NewServeMux()}
	srv.routes()

	req := httptest.NewRequest("GET", "/api/workflows/hello.glitch", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["source"] != src {
		t.Errorf("source mismatch")
	}
}

func TestPutWorkflow(t *testing.T) {
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	os.MkdirAll(wfDir, 0o755)

	os.WriteFile(filepath.Join(wfDir, "hello.glitch"), []byte("old"), 0o644)

	srv := &Server{workspace: dir, mux: http.NewServeMux()}
	srv.routes()

	body := `{"source": "new content"}`
	req := httptest.NewRequest("PUT", "/api/workflows/hello.glitch",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(filepath.Join(wfDir, "hello.glitch"))
	if string(data) != "new content" {
		t.Errorf("file not updated: %q", string(data))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd ~/Projects/gl1tch && go test ./internal/gui/ -run TestListWorkflows -v
```

Expected: FAIL (handlers not implemented).

- [ ] **Step 3: Implement api_workflows.go**

```go
package gui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

type workflowEntry struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description"`
}

func (s *Server) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	dir := s.workflowsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var workflows []workflowEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".glitch" && ext != ".yaml" && ext != ".yml" {
			continue
		}
		wf, err := pipeline.LoadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		workflows = append(workflows, workflowEntry{
			Name:        wf.Name,
			File:        e.Name(),
			Description: wf.Description,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflows)
}

func (s *Server) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "..") {
		http.Error(w, "invalid name", 400)
		return
	}

	path := filepath.Join(s.workflowsDir(), name)
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	// Extract param references from source
	params := extractParams(string(data))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"name":   name,
		"source": string(data),
		"params": params,
	})
}

func (s *Server) handlePutWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "..") {
		http.Error(w, "invalid name", 400)
		return
	}

	var body struct {
		Source string `json:"source"`
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := json.Unmarshal(data, &body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	path := filepath.Join(s.workflowsDir(), name)
	if err := os.WriteFile(path, []byte(body.Source), 0o644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) handleRunWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "..") {
		http.Error(w, "invalid name", 400)
		return
	}

	var body struct {
		Params map[string]string `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	wf, err := pipeline.LoadFile(filepath.Join(s.workflowsDir(), name))
	if err != nil {
		http.Error(w, fmt.Sprintf("load workflow: %v", err), 400)
		return
	}

	// Run in background goroutine
	go func() {
		cfg := s.loadConfig()
		pipeline.Run(wf, "", cfg.DefaultModel, body.Params, cfg.ProviderReg, pipeline.RunOpts{
			Telemetry:        cfg.Telemetry,
			ProviderResolver: cfg.ProviderResolver,
			Tiers:            cfg.Tiers,
		})
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// extractParams finds {{.param.X}} references in workflow source.
func extractParams(source string) []string {
	var params []string
	seen := make(map[string]bool)
	for _, match := range paramRe.FindAllStringSubmatch(source, -1) {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			params = append(params, match[1])
		}
	}
	return params
}

var paramRe = regexp.MustCompile(`\{\{\.param\.(\w+)\}\}`)
```

Note: The `handleRunWorkflow` references `s.loadConfig()` — this is a helper we'll add to `server.go` that mirrors the config loading from `cmd/ask.go`. Exact implementation:

```go
// Add to server.go

type runConfig struct {
	DefaultModel     string
	ProviderReg      *provider.ProviderRegistry
	ProviderResolver provider.ResolverFunc
	Telemetry        *esearch.Telemetry
	Tiers            []provider.TierConfig
}

func (s *Server) loadConfig() runConfig {
	// Load from ~/.config/glitch/config.yaml — same as cmd layer
	cfg, _ := config.Load()
	var tel *esearch.Telemetry
	esClient := esearch.NewClient("http://localhost:9200")
	if err := esClient.Ping(context.Background()); err == nil {
		tel = esearch.NewTelemetry(esClient)
		tel.EnsureIndices(context.Background())
	}
	return runConfig{
		DefaultModel:     cfg.DefaultModel,
		ProviderReg:      s.providerReg,
		ProviderResolver: cfg.BuildProviderResolver(),
		Telemetry:        tel,
		Tiers:            cfg.Tiers,
	}
}
```

- [ ] **Step 4: Add missing imports to test file**

Add `"strings"` to the import block in `api_workflows_test.go`.

- [ ] **Step 5: Run tests**

```bash
cd ~/Projects/gl1tch && go test ./internal/gui/ -v
```

Expected: all 3 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/gui/api_workflows.go internal/gui/api_workflows_test.go
git commit -m "feat(gui): workflow list, get, put, and run API endpoints"
```

---

### Task 4: Runs and Results API endpoints

**Files:**
- Create: `internal/gui/api_runs.go`
- Create: `internal/gui/api_results.go`

- [ ] **Step 1: Write test for listing runs**

Create `internal/gui/api_runs_test.go`:

```go
package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	srv := &Server{workspace: dir, mux: http.NewServeMux()}
	// Open in-memory store for tests
	st, err := store.OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	srv.store = st
	srv.routes()
	return srv
}

func TestListRuns(t *testing.T) {
	srv := setupTestServer(t)

	// Insert a test run
	id, err := srv.store.RecordRun("workflow", "hello", "")
	if err != nil {
		t.Fatal(err)
	}
	srv.store.FinishRun(id, "done", 0)

	req := httptest.NewRequest("GET", "/api/runs", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var runs []map[string]any
	json.Unmarshal(w.Body.Bytes(), &runs)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/gui/ -run TestListRuns -v
```

- [ ] **Step 3: Implement api_runs.go**

```go
package gui

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type runEntry struct {
	ID         int64  `json:"id"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Input      string `json:"input"`
	Output     string `json:"output,omitempty"`
	ExitStatus int    `json:"exit_status"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at,omitempty"`
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.DB().Query(
		`SELECT id, kind, name, COALESCE(input,''), COALESCE(output,''),
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0)
		 FROM runs ORDER BY id DESC LIMIT 100`,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var runs []runEntry
	for rows.Next() {
		var r runEntry
		rows.Scan(&r.ID, &r.Kind, &r.Name, &r.Input, &r.Output,
			&r.ExitStatus, &r.StartedAt, &r.FinishedAt)
		runs = append(runs, r)
	}
	if runs == nil {
		runs = []runEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}

	var run runEntry
	err = s.store.DB().QueryRow(
		`SELECT id, kind, name, COALESCE(input,''), COALESCE(output,''),
		        COALESCE(exit_status,0), started_at, COALESCE(finished_at,0)
		 FROM runs WHERE id = ?`, id,
	).Scan(&run.ID, &run.Kind, &run.Name, &run.Input, &run.Output,
		&run.ExitStatus, &run.StartedAt, &run.FinishedAt)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	// Get steps for this run
	stepRows, _ := s.store.DB().Query(
		`SELECT step_id, COALESCE(model,''), COALESCE(duration_ms,0)
		 FROM steps WHERE run_id = ?`, id)
	defer stepRows.Close()

	type stepEntry struct {
		StepID     string `json:"step_id"`
		Model      string `json:"model"`
		DurationMs int64  `json:"duration_ms"`
	}
	var steps []stepEntry
	for stepRows.Next() {
		var s stepEntry
		stepRows.Scan(&s.StepID, &s.Model, &s.DurationMs)
		steps = append(steps, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"run":   run,
		"steps": steps,
	})
}
```

Note: This requires exposing `store.DB()` — add to `store.go`:

```go
// DB returns the underlying *sql.DB for direct queries.
func (s *Store) DB() *sql.DB {
	return s.db
}
```

- [ ] **Step 4: Implement api_results.go**

```go
package gui

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleGetResult(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if strings.Contains(path, "..") {
		http.Error(w, "invalid path", 400)
		return
	}

	fullPath := filepath.Join(s.resultsDir(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	// Directory listing
	if info.IsDir() {
		entries, _ := os.ReadDir(fullPath)
		type fileEntry struct {
			Name  string `json:"name"`
			IsDir bool   `json:"is_dir"`
			Size  int64  `json:"size"`
		}
		var files []fileEntry
		for _, e := range entries {
			info, _ := e.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			files = append(files, fileEntry{
				Name:  e.Name(),
				IsDir: e.IsDir(),
				Size:  size,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
		return
	}

	// File content
	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ext := filepath.Ext(fullPath)
	switch ext {
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".md":
		w.Header().Set("Content-Type", "text/markdown")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
	w.Write(data)
}
```

- [ ] **Step 5: Run all tests**

```bash
go test ./internal/gui/ -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/gui/api_runs.go internal/gui/api_runs_test.go internal/gui/api_results.go internal/store/store.go
git commit -m "feat(gui): runs list/detail and results browser API endpoints"
```

---

### Task 5: Kibana embed endpoints

**Files:**
- Create: `internal/gui/api_kibana.go`

- [ ] **Step 1: Implement Kibana URL generation**

```go
package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const defaultKibanaURL = "http://localhost:5601"

// handleKibanaWorkflow returns an iframe URL for aggregate telemetry for a workflow.
func (s *Server) handleKibanaWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	// Build Kibana Discover URL filtered by workflow_name
	filter := fmt.Sprintf(`(query:(match_phrase:(workflow_name:'%s')))`, name)
	iframeURL := fmt.Sprintf(
		"%s/app/discover#/?_g=(time:(from:now-24h,to:now))&_a=(dataView:glitch-llm-calls,filters:!(%s),columns:!(step_id,model,tokens_in,tokens_out,latency_ms,cost_usd))",
		defaultKibanaURL, url.PathEscape(filter),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  iframeURL,
		"type": "workflow",
		"name": name,
	})
}

// handleKibanaRun returns an iframe URL for a specific run's telemetry.
func (s *Server) handleKibanaRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	filter := fmt.Sprintf(`(query:(match_phrase:(run_id:'%s')))`, id)
	iframeURL := fmt.Sprintf(
		"%s/app/discover#/?_g=(time:(from:now-24h,to:now))&_a=(dataView:glitch-llm-calls,filters:!(%s),columns:!(step_id,model,tokens_in,tokens_out,latency_ms,cost_usd,escalated))",
		defaultKibanaURL, url.PathEscape(filter),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  iframeURL,
		"type": "run",
		"id":   id,
	})
}
```

- [ ] **Step 2: Verify compile**

```bash
go build .
```

- [ ] **Step 3: Commit**

```bash
git add internal/gui/api_kibana.go
git commit -m "feat(gui): Kibana iframe URL endpoints for workflow and run telemetry"
```

---

### Task 6: API client and sexpr CodeMirror language

**Files:**
- Create: `gui/src/lib/api.js`
- Create: `gui/src/lib/sexpr.grammar`
- Create: `gui/src/lib/sexpr-lang.js`
- Create: `gui/src/lib/markdown.js`

- [ ] **Step 1: Create API client**

`gui/src/lib/api.js`:

```js
const BASE = ''

async function request(path, opts = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(`${res.status}: ${text}`)
  }
  const ct = res.headers.get('content-type') || ''
  if (ct.includes('json')) return res.json()
  return res.text()
}

export const api = {
  listWorkflows: () => request('/api/workflows'),
  getWorkflow: (name) => request(`/api/workflows/${name}`),
  saveWorkflow: (name, source) =>
    request(`/api/workflows/${name}`, {
      method: 'PUT',
      body: JSON.stringify({ source }),
    }),
  runWorkflow: (name, params) =>
    request(`/api/workflows/${name}/run`, {
      method: 'POST',
      body: JSON.stringify({ params }),
    }),
  listRuns: () => request('/api/runs'),
  getRun: (id) => request(`/api/runs/${id}`),
  getResult: (path) => request(`/api/results/${path}`),
  getKibanaWorkflow: (name) => request(`/api/kibana/workflow/${name}`),
  getKibanaRun: (id) => request(`/api/kibana/run/${id}`),
}
```

- [ ] **Step 2: Create Lezer grammar for sexpr**

`gui/src/lib/sexpr.grammar`:

```
@top Program { expression* }

expression {
  List | Atom
}

List {
  "(" Keyword expression* ")"
}

Atom {
  String | Number | Symbol | TemplateExpr
}

@tokens {
  Keyword { std.asciiLetter (std.asciiLetter | std.digit | "-" | "_")* }
  Symbol { (std.asciiLetter | "_") (std.asciiLetter | std.digit | "-" | "_" | "." | "/")* }
  String { '"' (!["\\] | "\\" _)* '"' }
  Number { std.digit+ ("." std.digit+)? }
  TemplateExpr { "{{" ![}]* "}}" }
  LineComment { ";" ![\n]* }
  space { (" " | "\t" | "\n" | "\r")+ }
}

@skip { space | LineComment }
```

- [ ] **Step 3: Create CodeMirror language support**

`gui/src/lib/sexpr-lang.js`:

```js
import { parser } from './sexpr.grammar'
import { LRLanguage, LanguageSupport } from '@codemirror/language'
import { styleTags, tags } from '@lezer/highlight'

const sexprLanguage = LRLanguage.define({
  parser: parser.configure({
    props: [
      styleTags({
        Keyword: tags.keyword,
        String: tags.string,
        Number: tags.number,
        Symbol: tags.variableName,
        TemplateExpr: tags.special(tags.string),
        LineComment: tags.lineComment,
        '( )': tags.paren,
      }),
    ],
  }),
  languageData: {
    commentTokens: { line: ';' },
    closeBrackets: { brackets: ['(', '"'] },
  },
})

export function sexpr() {
  return new LanguageSupport(sexprLanguage)
}
```

Note: The grammar import requires a Lezer build step. Add to `package.json` scripts:

```json
"grammar": "lezer-generator src/lib/sexpr.grammar -o src/lib/sexpr.grammar.js"
```

And update the import in `sexpr-lang.js` to:
```js
import { parser } from './sexpr.grammar.js'
```

- [ ] **Step 4: Create markdown renderer**

`gui/src/lib/markdown.js`:

```js
import { marked } from 'marked'
import hljs from 'highlight.js/lib/core'
import json from 'highlight.js/lib/languages/json'
import bash from 'highlight.js/lib/languages/bash'
import yaml from 'highlight.js/lib/languages/yaml'
import python from 'highlight.js/lib/languages/python'
import go from 'highlight.js/lib/languages/go'
import 'highlight.js/styles/github-dark.css'

hljs.registerLanguage('json', json)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('python', python)
hljs.registerLanguage('go', go)

marked.setOptions({
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value
    }
    return hljs.highlightAuto(code).value
  },
})

export function renderMarkdown(src) {
  return marked.parse(src)
}
```

- [ ] **Step 5: Build grammar and verify**

```bash
cd gui && npx lezer-generator src/lib/sexpr.grammar -o src/lib/sexpr.grammar.js && npm run build
```

- [ ] **Step 6: Commit**

```bash
git add gui/src/lib/
git commit -m "feat(gui): API client, sexpr CodeMirror grammar, and markdown renderer"
```

---

### Task 7: Svelte route components

**Files:**
- Modify: `gui/src/routes/WorkflowList.svelte`
- Modify: `gui/src/routes/Editor.svelte`
- Modify: `gui/src/routes/RunDialog.svelte`
- Modify: `gui/src/routes/RunView.svelte`
- Modify: `gui/src/routes/ResultsBrowser.svelte`

- [ ] **Step 1: Implement WorkflowList.svelte**

```svelte
<script>
  import { onMount } from 'svelte'
  import { link } from 'svelte-spa-router'
  import { api } from '../lib/api.js'

  let workflows = []
  let error = null

  onMount(async () => {
    try {
      workflows = await api.listWorkflows()
    } catch (e) {
      error = e.message
    }
  })
</script>

<div class="workflow-list">
  <h2>Workflows</h2>
  {#if error}
    <p class="error">{error}</p>
  {:else if workflows.length === 0}
    <p class="muted">No workflows found in workspace.</p>
  {:else}
    <ul>
      {#each workflows as wf}
        <li>
          <a href="/workflow/{wf.file}" use:link>
            <span class="name">{wf.name}</span>
            {#if wf.description}
              <span class="desc">{wf.description}</span>
            {/if}
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .workflow-list ul { list-style: none; }
  .workflow-list li {
    border: 1px solid var(--border);
    border-radius: 4px;
    margin-bottom: 0.5rem;
  }
  .workflow-list a {
    display: block;
    padding: 0.75rem 1rem;
    text-decoration: none;
    color: var(--text);
  }
  .workflow-list a:hover { background: var(--bg-hover); }
  .name { font-family: var(--font-mono); color: var(--accent); }
  .desc { display: block; color: var(--text-muted); font-size: 12px; margin-top: 2px; }
  .error { color: var(--danger); }
  .muted { color: var(--text-muted); }
</style>
```

- [ ] **Step 2: Implement Editor.svelte**

```svelte
<script>
  import { onMount } from 'svelte'
  import { EditorView, basicSetup } from 'codemirror'
  import { EditorState } from '@codemirror/state'
  import { oneDark } from '@codemirror/theme-one-dark'
  import { sexpr } from '../lib/sexpr-lang.js'
  import { api } from '../lib/api.js'
  import RunDialog from './RunDialog.svelte'

  export let params = {}

  let editorEl
  let view
  let saving = false
  let saved = false
  let showRunDialog = false
  let workflowParams = []

  onMount(async () => {
    const data = await api.getWorkflow(params.name)
    workflowParams = data.params || []

    view = new EditorView({
      state: EditorState.create({
        doc: data.source,
        extensions: [basicSetup, oneDark, sexpr()],
      }),
      parent: editorEl,
    })
  })

  async function save() {
    saving = true
    const source = view.state.doc.toString()
    await api.saveWorkflow(params.name, source)
    saving = false
    saved = true
    setTimeout(() => (saved = false), 2000)
  }
</script>

<div class="editor-page">
  <div class="toolbar">
    <h2>{params.name}</h2>
    <div class="actions">
      <button on:click={save} disabled={saving}>
        {saving ? 'Saving...' : saved ? 'Saved' : 'Save'}
      </button>
      <button class="primary" on:click={() => (showRunDialog = true)}>
        Run
      </button>
    </div>
  </div>
  <div class="editor" bind:this={editorEl}></div>
</div>

{#if showRunDialog}
  <RunDialog
    name={params.name}
    params={workflowParams}
    on:close={() => (showRunDialog = false)}
  />
{/if}

<style>
  .editor-page { display: flex; flex-direction: column; height: calc(100vh - 60px); }
  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border);
    margin-bottom: 0.5rem;
  }
  .toolbar h2 { font-family: var(--font-mono); font-size: 14px; }
  .actions { display: flex; gap: 0.5rem; }
  .editor { flex: 1; overflow: auto; }
  .editor :global(.cm-editor) { height: 100%; }
</style>
```

- [ ] **Step 3: Implement RunDialog.svelte**

```svelte
<script>
  import { createEventDispatcher } from 'svelte'
  import { push } from 'svelte-spa-router'
  import { api } from '../lib/api.js'

  export let name
  export let params = []

  const dispatch = createEventDispatcher()
  let values = {}
  let running = false

  // Init empty values
  params.forEach((p) => (values[p] = ''))

  async function run() {
    running = true
    try {
      await api.runWorkflow(name, values)
      dispatch('close')
    } catch (e) {
      alert(e.message)
    }
    running = false
  }
</script>

<div class="overlay" on:click|self={() => dispatch('close')}>
  <div class="dialog">
    <h3>Run {name}</h3>
    {#if params.length > 0}
      {#each params as param}
        <label>
          <span>{param}</span>
          <input bind:value={values[param]} placeholder={param} />
        </label>
      {/each}
    {:else}
      <p class="muted">No parameters required.</p>
    {/if}
    <div class="actions">
      <button on:click={() => dispatch('close')}>Cancel</button>
      <button class="primary" on:click={run} disabled={running}>
        {running ? 'Starting...' : 'Start Run'}
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }
  .dialog {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    min-width: 400px;
    max-width: 500px;
  }
  h3 { font-family: var(--font-mono); margin-bottom: 1rem; }
  label { display: block; margin-bottom: 0.75rem; }
  label span { display: block; color: var(--text-muted); font-size: 12px; margin-bottom: 4px; }
  input {
    width: 100%;
    background: var(--bg);
    border: 1px solid var(--border);
    color: var(--text);
    padding: 0.5rem;
    border-radius: 4px;
    font-family: var(--font-mono);
    font-size: 13px;
  }
  .actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 1rem; }
  .muted { color: var(--text-muted); }
</style>
```

- [ ] **Step 4: Implement RunView.svelte**

```svelte
<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { renderMarkdown } from '../lib/markdown.js'

  export let params = {}

  let run = null
  let steps = []
  let kibanaURL = null
  let error = null

  onMount(async () => {
    try {
      const data = await api.getRun(params.id)
      run = data.run
      steps = data.steps || []

      const kibana = await api.getKibanaRun(params.id)
      kibanaURL = kibana.url
    } catch (e) {
      error = e.message
    }
  })

  function formatTime(ms) {
    if (!ms) return '—'
    return new Date(ms).toLocaleString()
  }

  function statusClass(exit) {
    return exit === 0 ? 'success' : 'fail'
  }
</script>

<div class="run-view">
  {#if error}
    <p class="error">{error}</p>
  {:else if !run}
    <p>Loading...</p>
  {:else}
    <div class="header">
      <h2>{run.name}</h2>
      <span class="badge {statusClass(run.exit_status)}">
        {run.exit_status === 0 ? 'PASS' : 'FAIL'}
      </span>
    </div>

    <div class="meta">
      <span>Started: {formatTime(run.started_at)}</span>
      <span>Finished: {formatTime(run.finished_at)}</span>
    </div>

    <h3>Steps</h3>
    <table>
      <thead>
        <tr><th>Step</th><th>Model</th><th>Duration</th></tr>
      </thead>
      <tbody>
        {#each steps as step}
          <tr>
            <td class="mono">{step.step_id}</td>
            <td>{step.model || '—'}</td>
            <td>{step.duration_ms ? `${(step.duration_ms / 1000).toFixed(1)}s` : '—'}</td>
          </tr>
        {/each}
      </tbody>
    </table>

    {#if run.output}
      <h3>Output</h3>
      <div class="output">{@html renderMarkdown(run.output)}</div>
    {/if}

    {#if kibanaURL}
      <h3>Telemetry</h3>
      <iframe src={kibanaURL} title="Kibana telemetry" class="kibana-frame"></iframe>
    {/if}
  {/if}
</div>

<style>
  .run-view h2 { font-family: var(--font-mono); }
  .header { display: flex; align-items: center; gap: 1rem; margin-bottom: 0.5rem; }
  .badge { padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
  .badge.success { background: var(--success); color: #000; }
  .badge.fail { background: var(--danger); color: #fff; }
  .meta { color: var(--text-muted); font-size: 12px; display: flex; gap: 1.5rem; margin-bottom: 1rem; }
  h3 { margin-top: 1.5rem; margin-bottom: 0.5rem; font-size: 14px; }
  table { width: 100%; border-collapse: collapse; }
  th, td { text-align: left; padding: 0.5rem; border-bottom: 1px solid var(--border); font-size: 13px; }
  th { color: var(--text-muted); font-weight: normal; }
  .mono { font-family: var(--font-mono); }
  .output { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 4px; padding: 1rem; }
  .kibana-frame { width: 100%; height: 400px; border: 1px solid var(--border); border-radius: 4px; margin-top: 0.5rem; }
  .error { color: var(--danger); }
</style>
```

- [ ] **Step 5: Implement ResultsBrowser.svelte**

```svelte
<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { renderMarkdown } from '../lib/markdown.js'
  import hljs from 'highlight.js/lib/core'

  let tree = []
  let currentPath = ''
  let fileContent = null
  let fileType = null
  let error = null

  onMount(() => loadDir(''))

  async function loadDir(path) {
    currentPath = path
    fileContent = null
    try {
      tree = await api.getResult(path || '.')
    } catch (e) {
      error = e.message
    }
  }

  async function openFile(name) {
    const path = currentPath ? `${currentPath}/${name}` : name
    try {
      const content = await api.getResult(path)
      fileContent = content
      fileType = name.split('.').pop()
    } catch (e) {
      error = e.message
    }
  }

  function navigate(name, isDir) {
    if (isDir) {
      const path = currentPath ? `${currentPath}/${name}` : name
      loadDir(path)
    } else {
      openFile(name)
    }
  }

  function goUp() {
    const parts = currentPath.split('/')
    parts.pop()
    loadDir(parts.join('/'))
  }

  function renderContent(content, type) {
    if (type === 'md') return renderMarkdown(content)
    if (type === 'json') {
      try {
        const formatted = JSON.stringify(JSON.parse(content), null, 2)
        return `<pre><code>${hljs.highlight(formatted, { language: 'json' }).value}</code></pre>`
      } catch {
        return `<pre><code>${content}</code></pre>`
      }
    }
    return `<pre><code>${content}</code></pre>`
  }
</script>

<div class="results">
  <div class="breadcrumb">
    <span>results/</span>
    {#if currentPath}
      <button class="link" on:click={goUp}>..</button>
      <span>{currentPath}</span>
    {/if}
  </div>

  <div class="split">
    <div class="file-tree">
      {#each tree as entry}
        <div
          class="entry"
          class:dir={entry.is_dir}
          on:click={() => navigate(entry.name, entry.is_dir)}
        >
          <span class="icon">{entry.is_dir ? '/' : ''}</span>
          {entry.name}
        </div>
      {/each}
    </div>

    <div class="preview">
      {#if fileContent !== null}
        <div class="rendered">{@html renderContent(fileContent, fileType)}</div>
      {:else}
        <p class="muted">Select a file to preview.</p>
      {/if}
    </div>
  </div>
</div>

<style>
  .results { height: calc(100vh - 60px); display: flex; flex-direction: column; }
  .breadcrumb {
    padding: 0.5rem 0;
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 13px;
    border-bottom: 1px solid var(--border);
    margin-bottom: 0.5rem;
  }
  .link { background: none; border: none; color: var(--accent); cursor: pointer; font-family: var(--font-mono); }
  .split { display: flex; flex: 1; gap: 1rem; overflow: hidden; }
  .file-tree { width: 250px; overflow-y: auto; border-right: 1px solid var(--border); padding-right: 1rem; }
  .entry {
    padding: 0.35rem 0.5rem;
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 13px;
    border-radius: 4px;
  }
  .entry:hover { background: var(--bg-hover); }
  .entry.dir { color: var(--accent); }
  .icon { display: inline-block; width: 1em; }
  .preview { flex: 1; overflow-y: auto; }
  .rendered { padding: 1rem; background: var(--bg-surface); border-radius: 4px; }
  .rendered :global(pre) { overflow-x: auto; }
  .muted { color: var(--text-muted); }
</style>
```

- [ ] **Step 6: Build and verify**

```bash
cd gui && npm run build
```

Expected: `gui/dist/` produced without errors.

- [ ] **Step 7: Commit**

```bash
git add gui/src/routes/
git commit -m "feat(gui): implement all route components — workflow list, editor, run, results"
```

---

### Task 8: Build script and go:embed wiring

**Files:**
- Create: `scripts/build-gui.sh`
- Modify: `internal/gui/server.go` (fix embed path)

- [ ] **Step 1: Create build script**

`scripts/build-gui.sh`:

```bash
#!/bin/bash
# Build the Svelte frontend and copy dist into internal/gui/ for go:embed.
set -euo pipefail

cd "$(dirname "$0")/.."

echo ">> Building frontend..."
cd gui
npm ci --silent
npm run grammar 2>/dev/null || true
npm run build
cd ..

echo ">> Copying dist to internal/gui/dist..."
rm -rf internal/gui/dist
cp -r gui/dist internal/gui/dist

echo ">> Building glitch binary..."
go build -o glitch .

echo ">> Done: ./glitch"
```

- [ ] **Step 2: Make executable**

```bash
chmod +x scripts/build-gui.sh
```

- [ ] **Step 3: Run full build**

```bash
./scripts/build-gui.sh
```

Expected: `./glitch` binary produced.

- [ ] **Step 4: Test the GUI**

```bash
./glitch --workspace ~/Projects/stokagent workflow gui
```

Expected: `>> gl1tch gui: http://127.0.0.1:8374` and opening that URL shows the workflow list.

- [ ] **Step 5: Commit**

```bash
git add scripts/build-gui.sh internal/gui/dist
git commit -m "feat(gui): build script — frontend + go:embed + single binary"
```

---

### Task 9: End-to-end smoke test

**Files:**
- Create: `internal/gui/smoke_test.go`

- [ ] **Step 1: Write smoke test**

```go
package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSmokeEndToEnd(t *testing.T) {
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	resDir := filepath.Join(dir, "results", "test")
	os.MkdirAll(wfDir, 0o755)
	os.MkdirAll(resDir, 0o755)

	// Create a workflow
	os.WriteFile(filepath.Join(wfDir, "smoke.glitch"), []byte(`(workflow
  (name "smoke")
  (description "smoke test")
  (step (id "hello") (run "echo hi"))
)`), 0o644)

	// Create a result file
	os.WriteFile(filepath.Join(resDir, "output.md"), []byte("# Result\nAll good."), 0o644)

	st, _ := store.OpenAt(filepath.Join(dir, "test.db"))
	defer st.Close()
	id, _ := st.RecordRun("workflow", "smoke", "")
	st.FinishRun(id, "done", 0)

	srv := &Server{workspace: dir, store: st, mux: http.NewServeMux()}
	srv.routes()

	// List workflows
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows", nil))
	if w.Code != 200 {
		t.Fatalf("list workflows: %d", w.Code)
	}
	var wfs []map[string]string
	json.Unmarshal(w.Body.Bytes(), &wfs)
	if len(wfs) != 1 || wfs[0]["name"] != "smoke" {
		t.Fatalf("unexpected workflows: %v", wfs)
	}

	// Get workflow source
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows/smoke.glitch", nil))
	if w.Code != 200 {
		t.Fatalf("get workflow: %d", w.Code)
	}

	// List runs
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/runs", nil))
	if w.Code != 200 {
		t.Fatalf("list runs: %d", w.Code)
	}

	// Get result directory
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/results/test", nil))
	if w.Code != 200 {
		t.Fatalf("list results: %d", w.Code)
	}

	// Get result file
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/results/test/output.md", nil))
	if w.Code != 200 {
		t.Fatalf("get result: %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/markdown" {
		t.Errorf("expected text/markdown, got %s", w.Header().Get("Content-Type"))
	}

	// Kibana endpoints (just verify they return JSON, not that Kibana is up)
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/kibana/workflow/smoke", nil))
	if w.Code != 200 {
		t.Fatalf("kibana workflow: %d", w.Code)
	}
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/kibana/run/1", nil))
	if w.Code != 200 {
		t.Fatalf("kibana run: %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd ~/Projects/gl1tch && go test ./internal/gui/ -v
```

Expected: all tests PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/gui/smoke_test.go
git commit -m "test(gui): end-to-end smoke test for all API endpoints"
```

---

Plan complete and saved to `docs/superpowers/plans/2026-04-15-glitch-workflow-gui.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?