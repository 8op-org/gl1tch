# GUI Settings Page & Playwright Tests — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a dedicated Settings page to the gl1tch workflow GUI that lets users configure workflow defaults (model, provider, default params) and workspace config (name, Kibana URL, repos), with full Playwright E2E test coverage.

**Architecture:** Backend-first approach. Extend the workspace parser to support `(params ...)` and serialization, add `PUT /api/workspace` + `GET /api/providers` endpoints, then build the Settings Svelte page, wire RunDialog default pre-fill, and finish with Playwright tests.

**Tech Stack:** Go (net/http, sexpr parser), Svelte 5 (runes), Playwright, CodeMirror (not touched)

**Spec:** `docs/superpowers/specs/2026-04-15-gui-settings-and-playwright-design.md`

---

### Task 1: Extend workspace parser — Params field + parsing

**Files:**
- Modify: `internal/workspace/workspace.go:19-23` (Defaults struct)
- Modify: `internal/workspace/workspace.go:115-146` (convertDefaults function)
- Test: `internal/workspace/workspace_test.go`

- [ ] **Step 1: Write the failing test for params parsing**

Add to `internal/workspace/workspace_test.go`:

```go
func TestParseWorkspace_Params(t *testing.T) {
	src := []byte(`
(workspace "test"
  :description "test workspace"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    (params
      :repo "elastic/kibana"
      :results-dir "results/kibana")))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Defaults.Model != "qwen2.5:7b" {
		t.Errorf("Model: got %q, want %q", w.Defaults.Model, "qwen2.5:7b")
	}
	if len(w.Defaults.Params) != 2 {
		t.Fatalf("Params len: got %d, want 2", len(w.Defaults.Params))
	}
	if w.Defaults.Params["repo"] != "elastic/kibana" {
		t.Errorf("Params[repo]: got %q, want %q", w.Defaults.Params["repo"], "elastic/kibana")
	}
	if w.Defaults.Params["results-dir"] != "results/kibana" {
		t.Errorf("Params[results-dir]: got %q, want %q", w.Defaults.Params["results-dir"], "results/kibana")
	}
}

func TestParseWorkspace_NoParams(t *testing.T) {
	src := []byte(`
(workspace "test"
  (defaults :model "qwen2.5:7b"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.Params == nil {
		t.Fatal("Params should be initialized to empty map, not nil")
	}
	if len(w.Defaults.Params) != 0 {
		t.Errorf("Params should be empty, got %v", w.Defaults.Params)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -run TestParseWorkspace_Params -v`
Expected: FAIL — `w.Defaults.Params` field does not exist

- [ ] **Step 3: Add Params field to Defaults and update parser**

In `internal/workspace/workspace.go`, update the `Defaults` struct:

```go
type Defaults struct {
	Model         string
	Provider      string
	Elasticsearch string
	Params        map[string]string
}
```

Update `convertDefaults` to handle the `(params ...)` child form and initialize the map:

```go
func convertDefaults(n *sexpr.Node) (Defaults, error) {
	children := n.Children[1:] // skip "defaults" symbol
	d := Defaults{Params: map[string]string{}}

	i := 0
	for i < len(children) {
		child := children[i]

		// Keyword args: :model, :provider, :elasticsearch
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return Defaults{}, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "model":
				d.Model = val.StringVal()
			case "provider":
				d.Provider = val.StringVal()
			case "elasticsearch":
				d.Elasticsearch = val.StringVal()
			default:
				return Defaults{}, fmt.Errorf("line %d: unknown defaults keyword :%s", child.Line, key)
			}
			i++
			continue
		}

		// List form: (params :key "val" ...)
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].SymbolVal()
			if head == "" {
				head = child.Children[0].StringVal()
			}
			switch head {
			case "params":
				p, err := convertParams(child)
				if err != nil {
					return Defaults{}, err
				}
				d.Params = p
			default:
				return Defaults{}, fmt.Errorf("line %d: unknown defaults form %q", child.Line, head)
			}
			i++
			continue
		}

		return Defaults{}, fmt.Errorf("line %d: unexpected form in defaults", child.Line)
	}

	return d, nil
}

func convertParams(n *sexpr.Node) (map[string]string, error) {
	children := n.Children[1:] // skip "params" symbol
	params := map[string]string{}

	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			params[key] = children[i].StringVal()
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in params", child.Line)
	}

	return params, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -v`
Expected: ALL PASS (including existing tests)

- [ ] **Step 5: Commit**

```bash
git add internal/workspace/workspace.go internal/workspace/workspace_test.go
git commit -m "feat(workspace): add Params field to Defaults and parse (params ...) form"
```

---

### Task 2: Workspace serialization — Serialize function

**Files:**
- Create: `internal/workspace/serialize.go`
- Test: `internal/workspace/workspace_test.go`

- [ ] **Step 1: Write the failing test for serialization**

Add to `internal/workspace/workspace_test.go`:

```go
func TestSerialize_Full(t *testing.T) {
	w := &Workspace{
		Name:        "stokagent",
		Description: "Cross-repo research and automation",
		Owner:       "adam",
		Repos:       []string{"elastic/kibana", "elastic/ensemble"},
		Defaults: Defaults{
			Model:         "qwen2.5:7b",
			Provider:      "ollama",
			Elasticsearch: "http://localhost:9200",
			Params:        map[string]string{"repo": "elastic/kibana", "results-dir": "results/kibana"},
		},
	}

	data := Serialize(w)
	// Round-trip: parse the serialized output
	w2, err := ParseFile(data)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v\n%s", err, data)
	}
	if w2.Name != w.Name {
		t.Errorf("Name: got %q, want %q", w2.Name, w.Name)
	}
	if w2.Description != w.Description {
		t.Errorf("Description: got %q, want %q", w2.Description, w.Description)
	}
	if w2.Owner != w.Owner {
		t.Errorf("Owner: got %q, want %q", w2.Owner, w.Owner)
	}
	if len(w2.Repos) != len(w.Repos) {
		t.Fatalf("Repos len: got %d, want %d", len(w2.Repos), len(w.Repos))
	}
	for i, r := range w.Repos {
		if w2.Repos[i] != r {
			t.Errorf("Repos[%d]: got %q, want %q", i, w2.Repos[i], r)
		}
	}
	if w2.Defaults.Model != w.Defaults.Model {
		t.Errorf("Model: got %q, want %q", w2.Defaults.Model, w.Defaults.Model)
	}
	if w2.Defaults.Provider != w.Defaults.Provider {
		t.Errorf("Provider: got %q, want %q", w2.Defaults.Provider, w.Defaults.Provider)
	}
	if w2.Defaults.Elasticsearch != w.Defaults.Elasticsearch {
		t.Errorf("Elasticsearch: got %q, want %q", w2.Defaults.Elasticsearch, w.Defaults.Elasticsearch)
	}
	if w2.Defaults.Params["repo"] != w.Defaults.Params["repo"] {
		t.Errorf("Params[repo]: got %q, want %q", w2.Defaults.Params["repo"], w.Defaults.Params["repo"])
	}
	if w2.Defaults.Params["results-dir"] != w.Defaults.Params["results-dir"] {
		t.Errorf("Params[results-dir]: got %q, want %q", w2.Defaults.Params["results-dir"], w.Defaults.Params["results-dir"])
	}
}

func TestSerialize_Minimal(t *testing.T) {
	w := &Workspace{
		Name:  "minimal",
		Repos: []string{},
		Defaults: Defaults{Params: map[string]string{}},
	}
	data := Serialize(w)
	w2, err := ParseFile(data)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v\n%s", err, data)
	}
	if w2.Name != "minimal" {
		t.Errorf("Name: got %q, want %q", w2.Name, "minimal")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -run TestSerialize -v`
Expected: FAIL — `Serialize` not defined

- [ ] **Step 3: Implement Serialize function**

Create `internal/workspace/serialize.go`:

```go
package workspace

import (
	"fmt"
	"sort"
	"strings"
)

// Serialize writes a Workspace back to s-expression format.
func Serialize(w *Workspace) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("(workspace %q", w.Name))

	if w.Description != "" {
		b.WriteString(fmt.Sprintf("\n  :description %q", w.Description))
	}
	if w.Owner != "" {
		b.WriteString(fmt.Sprintf("\n  :owner %q", w.Owner))
	}

	if len(w.Repos) > 0 {
		b.WriteString("\n  (repos")
		for _, r := range w.Repos {
			b.WriteString(fmt.Sprintf("\n    %q", r))
		}
		b.WriteString(")")
	}

	hasDefaults := w.Defaults.Model != "" || w.Defaults.Provider != "" ||
		w.Defaults.Elasticsearch != "" || len(w.Defaults.Params) > 0
	if hasDefaults {
		b.WriteString("\n  (defaults")
		if w.Defaults.Model != "" {
			b.WriteString(fmt.Sprintf("\n    :model %q", w.Defaults.Model))
		}
		if w.Defaults.Provider != "" {
			b.WriteString(fmt.Sprintf("\n    :provider %q", w.Defaults.Provider))
		}
		if w.Defaults.Elasticsearch != "" {
			b.WriteString(fmt.Sprintf("\n    :elasticsearch %q", w.Defaults.Elasticsearch))
		}
		if len(w.Defaults.Params) > 0 {
			b.WriteString("\n    (params")
			// Sort keys for deterministic output
			keys := make([]string, 0, len(w.Defaults.Params))
			for k := range w.Defaults.Params {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				b.WriteString(fmt.Sprintf("\n      :%s %q", k, w.Defaults.Params[k]))
			}
			b.WriteString(")")
		}
		b.WriteString(")")
	}

	b.WriteString(")\n")
	return []byte(b.String())
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/workspace/serialize.go internal/workspace/workspace_test.go
git commit -m "feat(workspace): add Serialize function for writing workspace.glitch"
```

---

### Task 3: Backend — PUT /api/workspace + GET /api/providers endpoints

**Files:**
- Modify: `internal/gui/api_workspace.go:12-52` (response structs + GET handler + new PUT handler)
- Modify: `internal/gui/server.go:70-84` (routes)
- Modify: `internal/gui/api_kibana.go:11` (use workspace ES URL instead of hardcoded const)
- Test: `internal/gui/api_workspace_test.go`

- [ ] **Step 1: Write failing tests for PUT /api/workspace and GET /api/providers**

Replace `internal/gui/api_workspace_test.go` with:

```go
package gui

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetWorkspace(t *testing.T) {
	srv := testServer(t)
	src := `(workspace "test-ws" :description "a test" :owner "adam" (repos "elastic/ensemble") (defaults :model "qwen2.5:7b" :provider "ollama" :elasticsearch "http://es:9200" (params :repo "elastic/kibana")))`
	os.WriteFile(filepath.Join(srv.workspace, "workspace.glitch"), []byte(src), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workspace", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp workspaceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Name != "test-ws" {
		t.Errorf("expected name test-ws, got %v", resp.Name)
	}
	if resp.Owner != "adam" {
		t.Errorf("expected owner adam, got %v", resp.Owner)
	}
	if resp.Defaults.Elasticsearch != "http://es:9200" {
		t.Errorf("expected elasticsearch http://es:9200, got %v", resp.Defaults.Elasticsearch)
	}
	if resp.Defaults.Params["repo"] != "elastic/kibana" {
		t.Errorf("expected params.repo elastic/kibana, got %v", resp.Defaults.Params["repo"])
	}
}

func TestGetWorkspace_Missing(t *testing.T) {
	srv := testServer(t)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workspace", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200 even when missing, got %d", w.Code)
	}
	var resp workspaceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Name != "" {
		t.Errorf("expected empty name for missing workspace, got %v", resp.Name)
	}
}

func TestPutWorkspace(t *testing.T) {
	srv := testServer(t)

	body := `{"name":"updated","description":"new desc","owner":"bob","repos":["elastic/kibana"],"defaults":{"model":"gpt-4o","provider":"openai","elasticsearch":"http://es:9200","params":{"repo":"elastic/kibana"}}}`
	req := httptest.NewRequest("PUT", "/api/workspace", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the file was written and can be read back
	w2 := httptest.NewRecorder()
	srv.mux.ServeHTTP(w2, httptest.NewRequest("GET", "/api/workspace", nil))
	var resp workspaceResponse
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp.Name != "updated" {
		t.Errorf("expected name updated, got %v", resp.Name)
	}
	if resp.Defaults.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %v", resp.Defaults.Model)
	}
	if resp.Defaults.Params["repo"] != "elastic/kibana" {
		t.Errorf("expected params.repo elastic/kibana, got %v", resp.Defaults.Params["repo"])
	}
}

func TestPutWorkspace_EmptyName(t *testing.T) {
	srv := testServer(t)

	body := `{"name":"","repos":[],"defaults":{"params":{}}}`
	req := httptest.NewRequest("PUT", "/api/workspace", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400 for empty name, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetProviders(t *testing.T) {
	srv := testServer(t)
	// No provider registry configured — should return empty array
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/providers", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var names []string
	json.Unmarshal(w.Body.Bytes(), &names)
	if names == nil {
		t.Error("expected empty array, got nil")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/gui/ -run "TestPutWorkspace|TestGetProviders|TestGetWorkspace$" -v`
Expected: FAIL — `handlePutWorkspace`, `handleGetProviders` not defined; `workspaceResponse` missing fields

- [ ] **Step 3: Update workspaceResponse struct and GET handler**

In `internal/gui/api_workspace.go`, replace the entire file:

```go
package gui

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/8op-org/gl1tch/internal/workspace"
)

type workspaceResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Owner       string            `json:"owner"`
	Repos       []string          `json:"repos"`
	Defaults    workspaceDefaults `json:"defaults"`
}

type workspaceDefaults struct {
	Model         string            `json:"model,omitempty"`
	Provider      string            `json:"provider,omitempty"`
	Elasticsearch string            `json:"elasticsearch,omitempty"`
	Params        map[string]string `json:"params"`
}

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.workspace, "workspace.glitch")
	data, err := os.ReadFile(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaceResponse{Repos: []string{}, Defaults: workspaceDefaults{Params: map[string]string{}}})
		return
	}

	ws, err := workspace.ParseFile(data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	params := ws.Defaults.Params
	if params == nil {
		params = map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaceResponse{
		Name:        ws.Name,
		Description: ws.Description,
		Owner:       ws.Owner,
		Repos:       ws.Repos,
		Defaults: workspaceDefaults{
			Model:         ws.Defaults.Model,
			Provider:      ws.Defaults.Provider,
			Elasticsearch: ws.Defaults.Elasticsearch,
			Params:        params,
		},
	})
}

func (s *Server) handlePutWorkspace(w http.ResponseWriter, r *http.Request) {
	var req workspaceResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), 400)
		return
	}
	if req.Name == "" {
		http.Error(w, "workspace name is required", 400)
		return
	}

	repos := req.Repos
	if repos == nil {
		repos = []string{}
	}
	params := req.Defaults.Params
	if params == nil {
		params = map[string]string{}
	}

	ws := &workspace.Workspace{
		Name:        req.Name,
		Description: req.Description,
		Owner:       req.Owner,
		Repos:       repos,
		Defaults: workspace.Defaults{
			Model:         req.Defaults.Model,
			Provider:      req.Defaults.Provider,
			Elasticsearch: req.Defaults.Elasticsearch,
			Params:        params,
		},
	}

	data := workspace.Serialize(ws)
	path := filepath.Join(s.workspace, "workspace.glitch")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		http.Error(w, "write workspace: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetProviders(w http.ResponseWriter, r *http.Request) {
	names := []string{}
	if s.providerReg != nil {
		names = s.providerReg.Names()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}
```

- [ ] **Step 4: Add Names() method to ProviderRegistry**

In `internal/provider/provider.go`, add after the `ProviderRegistry` struct (around line 30):

```go
// Names returns a sorted list of registered provider names.
func (r *ProviderRegistry) Names() []string {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.providers))
	for k := range r.providers {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
```

- [ ] **Step 5: Register new routes**

In `internal/gui/server.go`, add after the `GET /api/workspace` line (line 83):

```go
	s.mux.HandleFunc("PUT /api/workspace", s.handlePutWorkspace)
	s.mux.HandleFunc("GET /api/providers", s.handleGetProviders)
```

- [ ] **Step 6: Wire Kibana URL from workspace instead of hardcoded const**

In `internal/gui/api_kibana.go`, replace the hardcoded `defaultKibanaURL` usage. Replace the file:

```go
package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/workspace"
)

const defaultKibanaURL = "http://localhost:5601"

// sanitizeKQL escapes single quotes for Kibana KQL filter values.
func sanitizeKQL(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

// kibanaBaseURL reads the workspace elasticsearch field and derives the Kibana
// URL. Falls back to defaultKibanaURL if not configured.
func (s *Server) kibanaBaseURL() string {
	path := filepath.Join(s.workspace, "workspace.glitch")
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultKibanaURL
	}
	ws, err := workspace.ParseFile(data)
	if err != nil || ws.Defaults.Elasticsearch == "" {
		return defaultKibanaURL
	}
	return ws.Defaults.Elasticsearch
}

func (s *Server) handleKibanaWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	base := s.kibanaBaseURL()

	filter := fmt.Sprintf(`(query:(match_phrase:(workflow_name:'%s')))`, sanitizeKQL(name))
	iframeURL := fmt.Sprintf(
		"%s/app/discover#/?_g=(time:(from:now-24h,to:now))&_a=(dataView:glitch-llm-calls,filters:!(%s),columns:!(step,model,tokens_in,tokens_out,latency_ms,cost_usd))",
		base, url.PathEscape(filter),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  iframeURL,
		"type": "workflow",
		"name": name,
	})
}

func (s *Server) handleKibanaRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	base := s.kibanaBaseURL()

	filter := fmt.Sprintf(`(query:(match_phrase:(run_id:'%s')))`, sanitizeKQL(id))
	iframeURL := fmt.Sprintf(
		"%s/app/discover#/?_g=(time:(from:now-24h,to:now))&_a=(dataView:glitch-llm-calls,filters:!(%s),columns:!(step,model,tokens_in,tokens_out,latency_ms,cost_usd,escalated))",
		base, url.PathEscape(filter),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  iframeURL,
		"type": "run",
		"id":   id,
	})
}
```

- [ ] **Step 7: Run all Go tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/gui/ ./internal/workspace/ ./internal/provider/ -v`
Expected: ALL PASS

- [ ] **Step 8: Commit**

```bash
git add internal/gui/api_workspace.go internal/gui/api_kibana.go internal/gui/server.go internal/gui/api_workspace_test.go internal/provider/provider.go
git commit -m "feat(gui): add PUT /api/workspace, GET /api/providers, wire Kibana URL from workspace"
```

---

### Task 4: Frontend — API functions and Sidebar navigation

**Files:**
- Modify: `gui/src/lib/api.js:37-39` (add workspace + providers functions)
- Modify: `gui/src/lib/components/Sidebar.svelte` (wire settings as nav link)
- Modify: `gui/src/App.svelte` (add /settings route)

- [ ] **Step 1: Add API functions**

In `gui/src/lib/api.js`, add at the end of the file:

```js
export function getWorkspace() { return request('/api/workspace'); }
export function updateWorkspace(data) {
  return request('/api/workspace', { method: 'PUT', body: JSON.stringify(data) });
}
export function getProviders() { return request('/api/providers'); }
```

- [ ] **Step 2: Update Sidebar to make Settings a real nav link**

In `gui/src/lib/components/Sidebar.svelte`, replace the `sidebar-footer` div (lines 45-48) with a proper nav link:

```svelte
  <div class="sidebar-footer">
    <a
      href="#/settings"
      class="nav-item"
      class:active={isActive('/settings', $location)}
    >
      <span class="nav-icon">{@html icon('settings')}</span>
      <span class="nav-label">Settings</span>
    </a>
  </div>
```

- [ ] **Step 3: Add Settings route to App.svelte**

Replace `gui/src/App.svelte`:

```svelte
<script>
  import Router from 'svelte-spa-router';
  import Sidebar from './lib/components/Sidebar.svelte';
  import WorkflowList from './routes/WorkflowList.svelte';
  import Editor from './routes/Editor.svelte';
  import RunList from './routes/RunList.svelte';
  import RunView from './routes/RunView.svelte';
  import ResultsBrowser from './routes/ResultsBrowser.svelte';
  import Settings from './routes/Settings.svelte';

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': Editor,
    '/runs': RunList,
    '/run/:id': RunView,
    '/results': ResultsBrowser,
    '/settings': Settings,
  };
</script>

<Sidebar />
<main class="main-area">
  <Router {routes} />
</main>
```

- [ ] **Step 4: Create placeholder Settings component**

Create `gui/src/routes/Settings.svelte` with a minimal placeholder so the route resolves (full implementation in Task 5):

```svelte
<script>
  import { icon } from '../lib/icons.js';
</script>

<div class="page-header">
  <h1>{@html icon('settings', 20)} Settings</h1>
</div>
<div class="page-content">
  <p class="text-muted">Loading...</p>
</div>
```

- [ ] **Step 5: Commit**

```bash
git add gui/src/lib/api.js gui/src/lib/components/Sidebar.svelte gui/src/App.svelte gui/src/routes/Settings.svelte
git commit -m "feat(gui): add settings route, sidebar nav link, and API functions"
```

---

### Task 5: Frontend — Settings page implementation

**Files:**
- Modify: `gui/src/routes/Settings.svelte` (full implementation)

- [ ] **Step 1: Implement the full Settings page**

Replace `gui/src/routes/Settings.svelte`:

```svelte
<script>
  import { getWorkspace, updateWorkspace, getProviders } from '../lib/api.js';
  import { icon } from '../lib/icons.js';

  let workspace = $state(null);
  let providers = $state([]);
  let saving = $state(false);
  let saveStatus = $state(null);
  let dirty = $state(false);
  let error = $state(null);

  // Snapshot of the original workspace JSON for dirty detection
  let originalJson = $state('');

  // Default param editing
  let newParamKey = $state('');
  let newParamVal = $state('');

  // New repo input
  let newRepo = $state('');

  $effect(() => {
    loadData();
  });

  async function loadData() {
    try {
      const [ws, prov] = await Promise.all([getWorkspace(), getProviders()]);
      workspace = ws;
      providers = prov;
      originalJson = JSON.stringify(ws);
      dirty = false;
      error = null;
    } catch (e) {
      error = e.message;
    }
  }

  function markDirty() {
    if (workspace) {
      dirty = JSON.stringify(workspace) !== originalJson;
    }
  }

  async function handleSave() {
    saving = true;
    saveStatus = null;
    try {
      await updateWorkspace(workspace);
      originalJson = JSON.stringify(workspace);
      dirty = false;
      saveStatus = 'saved';
      setTimeout(() => saveStatus = null, 2000);
    } catch (e) {
      saveStatus = 'error';
      error = e.message;
    } finally {
      saving = false;
    }
  }

  function addParam() {
    if (!newParamKey.trim()) return;
    workspace.defaults.params[newParamKey.trim()] = newParamVal;
    newParamKey = '';
    newParamVal = '';
    markDirty();
  }

  function removeParam(key) {
    delete workspace.defaults.params[key];
    workspace.defaults = { ...workspace.defaults, params: { ...workspace.defaults.params } };
    markDirty();
  }

  function addRepo() {
    if (!newRepo.trim()) return;
    workspace.repos = [...workspace.repos, newRepo.trim()];
    newRepo = '';
    markDirty();
  }

  function removeRepo(index) {
    workspace.repos = workspace.repos.filter((_, i) => i !== index);
    markDirty();
  }
</script>

<div class="page-header">
  <h1>{@html icon('settings', 20)} Settings</h1>
  <div class="flex gap-sm items-center">
    {#if saveStatus === 'saved'}
      <span class="status-pass" style="font-size:12px">Saved</span>
    {/if}
    {#if saveStatus === 'error'}
      <span class="status-fail" style="font-size:12px">Error saving</span>
    {/if}
    <button class="primary" disabled={!dirty || saving} onclick={handleSave}>
      {#if saving}Saving...{:else}{@html icon('save', 14)} Save{/if}
    </button>
  </div>
</div>

<div class="page-content settings-content">
  {#if error && !workspace}
    <p class="status-fail">{error}</p>
  {:else if !workspace}
    <p class="text-muted">Loading...</p>
  {:else}
    <!-- Workflow Defaults -->
    <section class="settings-section">
      <h2>Workflow Defaults</h2>

      <label class="settings-field">
        <span class="settings-label">Default Model</span>
        <input
          type="text"
          bind:value={workspace.defaults.model}
          oninput={markDirty}
          placeholder="e.g. qwen2.5:7b"
        />
      </label>

      <label class="settings-field">
        <span class="settings-label">Default Provider</span>
        {#if providers.length > 0}
          <select bind:value={workspace.defaults.provider} onchange={markDirty}>
            <option value="">— select —</option>
            {#each providers as p}
              <option value={p}>{p}</option>
            {/each}
          </select>
        {:else}
          <input
            type="text"
            bind:value={workspace.defaults.provider}
            oninput={markDirty}
            placeholder="e.g. ollama"
          />
        {/if}
      </label>

      <div class="settings-field">
        <span class="settings-label">Default Parameters</span>
        <div class="params-list">
          {#each Object.entries(workspace.defaults.params) as [key, val]}
            <div class="param-row">
              <span class="param-key">{key}</span>
              <input
                type="text"
                value={val}
                oninput={(e) => { workspace.defaults.params[key] = e.target.value; markDirty(); }}
              />
              <button class="danger" onclick={() => removeParam(key)} title="Remove">&times;</button>
            </div>
          {/each}
          <div class="param-row add-row">
            <input type="text" bind:value={newParamKey} placeholder="key" />
            <input type="text" bind:value={newParamVal} placeholder="value" />
            <button onclick={addParam} disabled={!newParamKey.trim()}>+</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Workspace Config -->
    <section class="settings-section">
      <h2>Workspace</h2>

      <label class="settings-field">
        <span class="settings-label">Workspace Name</span>
        <input
          type="text"
          bind:value={workspace.name}
          oninput={markDirty}
          placeholder="workspace name"
        />
      </label>

      <label class="settings-field">
        <span class="settings-label">Kibana URL</span>
        <input
          type="text"
          bind:value={workspace.defaults.elasticsearch}
          oninput={markDirty}
          placeholder="http://localhost:5601"
        />
      </label>

      <div class="settings-field">
        <span class="settings-label">Repositories</span>
        <div class="repos-list">
          {#each workspace.repos as repo, i}
            <div class="repo-row">
              <span class="mono">{repo}</span>
              <button class="danger" onclick={() => removeRepo(i)} title="Remove">&times;</button>
            </div>
          {/each}
          <div class="repo-row add-row">
            <input type="text" bind:value={newRepo} placeholder="owner/repo" />
            <button onclick={addRepo} disabled={!newRepo.trim()}>+</button>
          </div>
        </div>
      </div>
    </section>
  {/if}
</div>

<style>
  .settings-content {
    max-width: 640px;
  }

  .settings-section {
    margin-bottom: 32px;
  }
  .settings-section h2 {
    color: var(--neon-cyan);
    margin-bottom: 16px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
  }

  .settings-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 16px;
  }
  .settings-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .params-list, .repos-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .param-row, .repo-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .param-row input, .repo-row input {
    flex: 1;
  }
  .param-key {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--neon-cyan);
    min-width: 120px;
  }
  .param-row button, .repo-row button {
    padding: 4px 10px;
    font-size: 14px;
  }
  .add-row {
    opacity: 0.7;
  }
  .add-row:focus-within {
    opacity: 1;
  }
</style>
```

- [ ] **Step 2: Verify the page builds**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npm run build`
Expected: Build succeeds with no errors

- [ ] **Step 3: Commit**

```bash
git add gui/src/routes/Settings.svelte
git commit -m "feat(gui): implement Settings page with workflow defaults and workspace config"
```

---

### Task 6: Frontend — RunDialog default parameter pre-fill

**Files:**
- Modify: `gui/src/routes/RunDialog.svelte`

- [ ] **Step 1: Update RunDialog to fetch and apply workspace defaults**

Replace `gui/src/routes/RunDialog.svelte`:

```svelte
<script>
  import { untrack } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { runWorkflow, getWorkspace } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Modal from '../lib/components/Modal.svelte';

  let { name, params = [], autoParams = {}, onclose } = $props();
  let values = $state({});
  let running = $state(false);
  let error = $state(null);
  let workspaceDefaults = $state({});

  // Fetch workspace defaults on mount
  $effect(() => {
    getWorkspace().then(ws => {
      workspaceDefaults = ws.defaults?.params || {};
    }).catch(() => {});
  });

  // Merge autoParams keys into params so fields appear, and pre-fill values
  const allParams = $derived(() => {
    const set = new Set(params);
    for (const k of Object.keys(autoParams)) set.add(k);
    for (const k of Object.keys(workspaceDefaults)) set.add(k);
    return [...set];
  });

  // Pre-fill from workspace defaults first, then autoParams override
  // Priority: autoParams > workspace defaults > empty
  $effect(() => {
    const wd = workspaceDefaults;
    const ap = autoParams;
    untrack(() => {
      for (const [k, v] of Object.entries(wd)) {
        if (v && !values[k]) values[k] = v;
      }
      for (const [k, v] of Object.entries(ap)) {
        if (v) values[k] = v;
      }
    });
  });

  async function handleSubmit() {
    running = true; error = null;
    try { const result = await runWorkflow(name, values); onclose?.(); if (result.run_id) push(`/run/${result.run_id}`); } catch (e) { error = e.message; running = false; }
  }
</script>

<Modal title="Run {name}" {onclose}>
  {#if allParams().length > 0}
    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="flex flex-col gap-md">
      {#each allParams() as param}<label class="field"><span class="field-label">{param}</span><input type="text" bind:value={values[param]} placeholder={param} /></label>{/each}
      {#if error}<p class="status-fail" style="font-size:12px">{error}</p>{/if}
      <div class="flex justify-between" style="margin-top:8px">
        <button type="button" onclick={onclose}>Cancel</button>
        <button type="submit" class="primary" disabled={running}>{#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}</button>
      </div>
    </form>
  {:else}
    <div class="flex flex-col gap-md">
      <p class="text-muted">No parameters required.</p>
      {#if error}<p class="status-fail" style="font-size:12px">{error}</p>{/if}
      <div class="flex justify-between">
        <button onclick={onclose}>Cancel</button>
        <button class="primary" disabled={running} onclick={handleSubmit}>{#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}</button>
      </div>
    </div>
  {/if}
</Modal>

<style>
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field-label { font-family: var(--font-mono); font-size: 12px; color: var(--text-muted); }
</style>
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add gui/src/routes/RunDialog.svelte
git commit -m "feat(gui): pre-fill RunDialog params from workspace defaults"
```

---

### Task 7: Playwright tests — Settings page

**Files:**
- Create: `gui/e2e/settings.spec.js`
- Modify: `gui/e2e/gui.spec.js:122` (sidebar nav count)

- [ ] **Step 1: Create settings E2E test file**

Create `gui/e2e/settings.spec.js`:

```js
import { test, expect } from '@playwright/test'

// ── Settings page ──────────────────────────────────────────────────
test.describe('Settings page', () => {
  test('navigates to settings via sidebar', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await sidebar.hover()
    await page.locator('.sidebar-footer .nav-item', { hasText: 'Settings' }).click()
    await expect(page).toHaveURL(/\/#\/settings/)
  })

  test('sidebar settings link shows active state', async ({ page }) => {
    await page.goto('#/settings')
    await page.waitForTimeout(500)
    const settingsLink = page.locator('.sidebar-footer .nav-item')
    await expect(settingsLink).toHaveClass(/active/)
  })

  test('page loads without JS errors', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/settings')
    await page.waitForTimeout(2000)
    expect(errors).toEqual([])
  })

  test('shows Workflow Defaults section', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  })

  test('shows Workspace section', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workspace' })).toBeVisible({ timeout: 5000 })
  })

  test('displays current workspace name', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    const nameInput = page.locator('input[placeholder="workspace name"]')
    await expect(nameInput).toBeVisible()
    const val = await nameInput.inputValue()
    expect(val.length).toBeGreaterThan(0)
  })

  test('displays default model input', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await expect(page.locator('input[placeholder="e.g. qwen2.5:7b"]')).toBeVisible()
  })

  test('displays Kibana URL field', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workspace' })).toBeVisible({ timeout: 5000 })
    await expect(page.locator('input[placeholder="http://localhost:5601"]')).toBeVisible()
  })

  test('save button is disabled when no changes made', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button', { hasText: 'Save' })).toBeDisabled()
  })

  test('editing a field enables save button', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    const modelInput = page.locator('input[placeholder="e.g. qwen2.5:7b"]')
    await modelInput.fill('test-model-change')
    await expect(page.locator('button', { hasText: 'Save' })).toBeEnabled()
  })

  test('saving workspace config persists and reloads', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })

    // Change model
    const modelInput = page.locator('input[placeholder="e.g. qwen2.5:7b"]')
    const original = await modelInput.inputValue()
    await modelInput.fill('test-persist-model')
    await page.locator('button', { hasText: 'Save' }).click()
    await expect(page.locator('text=Saved')).toBeVisible({ timeout: 3000 })

    // Reload and verify
    await page.reload()
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await expect(modelInput).toHaveValue('test-persist-model')

    // Restore original
    await modelInput.fill(original)
    await page.locator('button', { hasText: 'Save' }).click()
    await expect(page.locator('text=Saved')).toBeVisible({ timeout: 3000 })
  })

  test('adding a default parameter shows key-value row', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await page.locator('input[placeholder="key"]').fill('test-param')
    await page.locator('input[placeholder="value"]').fill('test-value')
    await page.locator('.add-row button', { hasText: '+' }).first().click()
    await expect(page.locator('.param-key', { hasText: 'test-param' })).toBeVisible()
  })

  test('removing a default parameter removes the row', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    // Add a param first
    await page.locator('input[placeholder="key"]').fill('temp-param')
    await page.locator('input[placeholder="value"]').fill('temp-val')
    await page.locator('.add-row button', { hasText: '+' }).first().click()
    await expect(page.locator('.param-key', { hasText: 'temp-param' })).toBeVisible()
    // Remove it
    const row = page.locator('.param-row', { hasText: 'temp-param' })
    await row.locator('button.danger').click()
    await expect(page.locator('.param-key', { hasText: 'temp-param' })).not.toBeVisible()
  })

  test('page header shows Settings title with icon', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('.page-header h1')).toContainText('Settings')
    await expect(page.locator('.page-header h1 svg')).toBeVisible()
  })

  test('no JS errors during interaction', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    // Interact with various fields
    await page.locator('input[placeholder="e.g. qwen2.5:7b"]').fill('test')
    await page.locator('input[placeholder="http://localhost:5601"]').fill('http://test:5601')
    await page.locator('input[placeholder="key"]').fill('k')
    await page.locator('input[placeholder="value"]').fill('v')
    await page.locator('.add-row button', { hasText: '+' }).first().click()
    expect(errors).toEqual([])
  })
})
```

- [ ] **Step 2: Update existing sidebar test for nav count**

In `gui/e2e/gui.spec.js`, find line 122:

```js
    await expect(page.locator('.nav-item')).toHaveCount(3)
```

Change to:

```js
    await expect(page.locator('.nav-item')).toHaveCount(4)
```

- [ ] **Step 3: Build the GUI before running tests**

Run: `cd /Users/stokes/Projects/gl1tch && go build -o glitch . && cd gui && npm run build`
Expected: Both build successfully

- [ ] **Step 4: Run Playwright tests**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx playwright test`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add gui/e2e/settings.spec.js gui/e2e/gui.spec.js
git commit -m "test(gui): add Playwright tests for settings page, update nav count"
```

---

### Task 8: Final integration verification

- [ ] **Step 1: Run all Go tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./...`
Expected: ALL PASS

- [ ] **Step 2: Run all Playwright tests**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx playwright test`
Expected: ALL PASS

- [ ] **Step 3: Build final binary**

Run: `cd /Users/stokes/Projects/gl1tch && go build -o glitch .`
Expected: Build succeeds
