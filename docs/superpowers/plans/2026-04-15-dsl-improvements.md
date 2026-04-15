# DSL Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rename DSL forms for readability, add first-class ES + embed forms, and add template string functions — eliminating curl/python boilerplate from knowledge pipelines.

**Architecture:** The sexpr DSL has a 3-part pattern: types (types.go) → parser (sexpr.go) → executor (runner.go). Each new form follows this pattern. ES forms use the existing `esearch.Client`. Embedding calls Ollama's `/api/embeddings` endpoint directly. Template functions are registered in the existing `render()` function.

**Tech Stack:** Go, `internal/pipeline` package, `internal/esearch` package, `internal/provider` package, `text/template`

**Worktree:** `/Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements` (branch: `feature/dsl-improvements`)

---

### Task 1: Template String Functions

Add string manipulation functions to the template `FuncMap` in `render()`. This is the simplest change and immediately useful.

**Files:**
- Modify: `internal/pipeline/runner.go:580-611` (render function)
- Create: `internal/pipeline/render_test.go`

- [ ] **Step 1: Write tests for template functions**

Create `internal/pipeline/render_test.go`:

```go
package pipeline

import (
	"testing"
)

func TestRender_Split(t *testing.T) {
	out, err := render(`{{split "/" "elastic/ensemble" | last}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ensemble" {
		t.Fatalf("got %q, want %q", out, "ensemble")
	}
}

func TestRender_First(t *testing.T) {
	out, err := render(`{{split "/" "elastic/ensemble" | first}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "elastic" {
		t.Fatalf("got %q, want %q", out, "elastic")
	}
}

func TestRender_Join(t *testing.T) {
	out, err := render(`{{split "/" "a/b/c" | join "-"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "a-b-c" {
		t.Fatalf("got %q, want %q", out, "a-b-c")
	}
}

func TestRender_Upper(t *testing.T) {
	out, err := render(`{{upper "hello"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "HELLO" {
		t.Fatalf("got %q, want %q", out, "HELLO")
	}
}

func TestRender_Lower(t *testing.T) {
	out, err := render(`{{lower "HELLO"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello" {
		t.Fatalf("got %q, want %q", out, "hello")
	}
}

func TestRender_Trim(t *testing.T) {
	out, err := render(`{{trim "  hello  "}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello" {
		t.Fatalf("got %q, want %q", out, "hello")
	}
}

func TestRender_TrimPrefix(t *testing.T) {
	out, err := render(`{{trimPrefix "refs/" "refs/heads/main"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "heads/main" {
		t.Fatalf("got %q, want %q", out, "heads/main")
	}
}

func TestRender_TrimSuffix(t *testing.T) {
	out, err := render(`{{trimSuffix ".git" "github.com/foo/bar.git"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "github.com/foo/bar" {
		t.Fatalf("got %q, want %q", out, "github.com/foo/bar")
	}
}

func TestRender_Replace(t *testing.T) {
	out, err := render(`{{replace "/" "-" "elastic/ensemble"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "elastic-ensemble" {
		t.Fatalf("got %q, want %q", out, "elastic-ensemble")
	}
}

func TestRender_Truncate(t *testing.T) {
	out, err := render(`{{truncate 5 "hello world"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello" {
		t.Fatalf("got %q, want %q", out, "hello")
	}
}

func TestRender_Truncate_ShortString(t *testing.T) {
	out, err := render(`{{truncate 100 "short"}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "short" {
		t.Fatalf("got %q, want %q", out, "short")
	}
}

func TestRender_Contains(t *testing.T) {
	out, err := render(`{{if contains "fix" "bugfix-123"}}yes{{else}}no{{end}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "yes" {
		t.Fatalf("got %q, want %q", out, "yes")
	}
}

func TestRender_HasPrefix(t *testing.T) {
	out, err := render(`{{if hasPrefix "feat" "feat/login"}}yes{{else}}no{{end}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "yes" {
		t.Fatalf("got %q, want %q", out, "yes")
	}
}

func TestRender_HasSuffix(t *testing.T) {
	out, err := render(`{{if hasSuffix ".go" "main.go"}}yes{{else}}no{{end}}`, map[string]any{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "yes" {
		t.Fatalf("got %q, want %q", out, "yes")
	}
}

func TestRender_PipeChain(t *testing.T) {
	data := map[string]any{
		"param": map[string]string{"repo": "elastic/ensemble"},
	}
	out, err := render(`{{split "/" .param.repo | last}}`, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ensemble" {
		t.Fatalf("got %q, want %q", out, "ensemble")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/ -run TestRender_ -v 2>&1 | head -30`
Expected: FAIL — functions not registered yet

- [ ] **Step 3: Add template functions to render()**

In `internal/pipeline/runner.go`, replace the `funcMap` in the `render` function (line 581-601) with:

```go
func render(tmpl string, data map[string]any, steps map[string]string) (string, error) {
	funcMap := template.FuncMap{
		"step": func(id string) string {
			return steps[id]
		},
		// stepfile writes step output to a temp file and returns the path.
		// Use in shell steps where inline content would break escaping:
		//   cat "{{stepfile "fetch-issue"}}"
		"stepfile": func(id string) string {
			content, ok := steps[id]
			if !ok {
				return ""
			}
			f, err := os.CreateTemp("", "glitch-step-*")
			if err != nil {
				return ""
			}
			f.WriteString(content)
			f.Close()
			return f.Name()
		},
		// String functions
		"split":      func(sep, s string) []string { return strings.Split(s, sep) },
		"join":       func(sep string, parts []string) string { return strings.Join(parts, sep) },
		"last":       func(s []string) string { if len(s) == 0 { return "" }; return s[len(s)-1] },
		"first":      func(s []string) string { if len(s) == 0 { return "" }; return s[0] },
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"trim":       strings.TrimSpace,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"replace":    func(old, new, s string) string { return strings.ReplaceAll(s, old, new) },
		"truncate": func(n int, s string) string {
			runes := []rune(s)
			if len(runes) <= n {
				return s
			}
			return string(runes[:n])
		},
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
	}
	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/ -run TestRender_ -v`
Expected: All PASS

- [ ] **Step 5: Run full test suite to check no regressions**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/... ./internal/sexpr/...`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/pipeline/runner.go internal/pipeline/render_test.go
git commit -m "feat(dsl): add template string functions (split, join, upper, lower, trim, etc.)"
```

---

### Task 2: Form Renames (Aliases)

Add alias cases in the parser so new names resolve to existing converters. Old names keep working.

**Files:**
- Modify: `internal/pipeline/sexpr.go:138-184` (convertForm switch) and `535-640` (convertStep switch)
- Modify: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write tests for renamed forms**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_AliasEach(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "data" (run "echo line1\nline2"))
  (each "data"
    (step "item" (run "echo {{.param.item}}"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	// each is an alias for map — should parse to the same form
	found := false
	for _, item := range w.Items {
		if item.Step != nil && item.Step.Form == "map" {
			found = true
			if item.Step.MapOver != "data" {
				t.Fatalf("expected MapOver %q, got %q", "data", item.Step.MapOver)
			}
		}
	}
	if !found {
		t.Fatal("expected each to parse as map form")
	}
}

func TestSexprWorkflow_AliasPick(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1" (run "echo '{\"a\":1}'"))
  (step "s2" (pick ".a" :from "s1")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.JsonPick == nil {
		t.Fatal("expected pick to parse as json-pick")
	}
	if s.JsonPick.Expr != ".a" {
		t.Fatalf("expected expr %q, got %q", ".a", s.JsonPick.Expr)
	}
}

func TestSexprWorkflow_AliasFetch(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1" (fetch "http://example.com")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected fetch to parse as http-get")
	}
	if s.HttpCall.Method != "GET" {
		t.Fatalf("expected GET, got %q", s.HttpCall.Method)
	}
}

func TestSexprWorkflow_AliasSend(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1" (send "http://example.com" :body "{\"x\":1}")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected send to parse as http-post")
	}
	if s.HttpCall.Method != "POST" {
		t.Fatalf("expected POST, got %q", s.HttpCall.Method)
	}
}

func TestSexprWorkflow_AliasRead(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1" (read "path/to/file.txt")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].ReadFile != "path/to/file.txt" {
		t.Fatalf("expected read to parse as read-file, got ReadFile=%q", w.Steps[0].ReadFile)
	}
}

func TestSexprWorkflow_AliasWrite(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "gen" (run "echo hello"))
  (step "s1" (write "out.txt" :from "gen")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[1].WriteFile == nil {
		t.Fatal("expected write to parse as write-file")
	}
	if w.Steps[1].WriteFile.Path != "out.txt" {
		t.Fatalf("expected path %q, got %q", "out.txt", w.Steps[1].WriteFile.Path)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/ -run TestSexprWorkflow_Alias -v 2>&1 | head -30`
Expected: FAIL — unknown form errors

- [ ] **Step 3: Add alias cases in convertForm()**

In `internal/pipeline/sexpr.go`, update the `convertForm` switch (line 139):

Change:
```go
	case "map":
```
To:
```go
	case "map", "each":
```

- [ ] **Step 4: Add alias cases in convertStep()**

In `internal/pipeline/sexpr.go`, update the `convertStep` switch (line 535):

Change:
```go
	case "json-pick":
```
To:
```go
	case "json-pick", "pick":
```

Change:
```go
	case "http-get":
```
To:
```go
	case "http-get", "fetch":
```

Change:
```go
	case "http-post":
```
To:
```go
	case "http-post", "send":
```

Change:
```go
	case "read-file":
```
To:
```go
	case "read-file", "read":
```

Change:
```go
	case "write-file":
```
To:
```go
	case "write-file", "write":
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/ -run TestSexprWorkflow_Alias -v`
Expected: All PASS

- [ ] **Step 6: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/... ./internal/sexpr/...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add human-friendly aliases (each, pick, fetch, send, read, write)"
```

---

### Task 3: ES Client Extensions (IndexDoc + DeleteByQuery)

Add single-document indexing and delete-by-query to `esearch.Client`. These are needed by the new DSL forms.

**Files:**
- Modify: `internal/esearch/client.go`
- Modify: `internal/esearch/client_test.go`

- [ ] **Step 1: Write tests for IndexDoc and DeleteByQuery**

Add to `internal/esearch/client_test.go`:

```go
func TestIndexDoc_BuildsCorrectRequest(t *testing.T) {
	// Test that IndexDoc constructs the right URL and method.
	// We use a test server to capture the request.
	var gotMethod, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(200)
		w.Write([]byte(`{"_id":"doc1","_version":1,"result":"created"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.IndexDoc(context.Background(), "test-index", "doc1", json.RawMessage(`{"title":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "PUT" {
		t.Fatalf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/test-index/_doc/doc1" {
		t.Fatalf("path = %q, want /test-index/_doc/doc1", gotPath)
	}
	if gotBody != `{"title":"hello"}` {
		t.Fatalf("body = %q, want %q", gotBody, `{"title":"hello"}`)
	}
	if resp.ID != "doc1" {
		t.Fatalf("response ID = %q, want doc1", resp.ID)
	}
}

func TestIndexDoc_AutoID(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(201)
		w.Write([]byte(`{"_id":"auto-id","_version":1,"result":"created"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.IndexDoc(context.Background(), "test-index", "", json.RawMessage(`{"title":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" {
		t.Fatalf("method = %q, want POST (auto-id)", gotMethod)
	}
	if gotPath != "/test-index/_doc" {
		t.Fatalf("path = %q, want /test-index/_doc", gotPath)
	}
	if resp.ID != "auto-id" {
		t.Fatalf("response ID = %q, want auto-id", resp.ID)
	}
}

func TestDeleteByQuery_BuildsCorrectRequest(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(200)
		w.Write([]byte(`{"deleted":5}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.DeleteByQuery(context.Background(), "test-index", json.RawMessage(`{"query":{"term":{"type":"old"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/test-index/_delete_by_query" {
		t.Fatalf("path = %q, want /test-index/_delete_by_query", gotPath)
	}
	if gotBody != `{"query":{"term":{"type":"old"}}}` {
		t.Fatalf("body = %q", gotBody)
	}
	if resp.Deleted != 5 {
		t.Fatalf("deleted = %d, want 5", resp.Deleted)
	}
}
```

Make sure the test file has these imports at the top:

```go
import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/esearch/ -run "TestIndexDoc|TestDeleteByQuery" -v 2>&1 | head -20`
Expected: FAIL — methods don't exist

- [ ] **Step 3: Implement IndexDoc and DeleteByQuery**

Add to `internal/esearch/client.go` before the `truncate` function:

```go
// IndexDocResponse holds the ES response from indexing a single document.
type IndexDocResponse struct {
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// IndexDoc indexes a single document. If docID is empty, ES auto-generates an ID (POST).
// If docID is provided, uses PUT for upsert semantics.
func (c *Client) IndexDoc(ctx context.Context, index, docID string, doc json.RawMessage) (*IndexDocResponse, error) {
	var method, path string
	if docID != "" {
		method = http.MethodPut
		path = "/" + index + "/_doc/" + docID
	} else {
		method = http.MethodPost
		path = "/" + index + "/_doc"
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(doc))
	if err != nil {
		return nil, fmt.Errorf("index doc: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("index doc: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("index doc: read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("index doc: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var result IndexDocResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("index doc: decode: %w", err)
	}
	return &result, nil
}

// DeleteByQueryResponse holds the ES response from a delete-by-query operation.
type DeleteByQueryResponse struct {
	Deleted int `json:"deleted"`
}

// DeleteByQuery deletes documents matching query from the given index.
func (c *Client) DeleteByQuery(ctx context.Context, index string, query json.RawMessage) (*DeleteByQueryResponse, error) {
	path := "/" + index + "/_delete_by_query"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("delete by query: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("delete by query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("delete by query: read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("delete by query: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var result DeleteByQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("delete by query: decode: %w", err)
	}
	return &result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/esearch/ -run "TestIndexDoc|TestDeleteByQuery" -v`
Expected: All PASS

- [ ] **Step 5: Run full esearch tests**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/esearch/...`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/esearch/client.go internal/esearch/client_test.go
git commit -m "feat(esearch): add IndexDoc and DeleteByQuery methods"
```

---

### Task 4: Ollama Embedding Function

Add an `EmbedOllama` function in the provider package that calls Ollama's `/api/embeddings` endpoint.

**Files:**
- Create: `internal/provider/embed.go`
- Create: `internal/provider/embed_test.go`

- [ ] **Step 1: Write test for EmbedOllama**

Create `internal/provider/embed_test.go`:

```go
package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbedOllama_ParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Fatalf("path = %q, want /api/embeddings", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Fatalf("method = %q, want POST", r.Method)
		}

		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		if req["model"] != "nomic-embed-text" {
			t.Fatalf("model = %q, want nomic-embed-text", req["model"])
		}
		if req["prompt"] != "hello world" {
			t.Fatalf("prompt = %q, want hello world", req["prompt"])
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{
			"embedding": []float64{0.1, 0.2, 0.3},
		})
	}))
	defer srv.Close()

	vec, err := EmbedOllama(srv.URL, "nomic-embed-text", "hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 3 {
		t.Fatalf("vector length = %d, want 3", len(vec))
	}
	if vec[0] != 0.1 || vec[1] != 0.2 || vec[2] != 0.3 {
		t.Fatalf("vector = %v, want [0.1 0.2 0.3]", vec)
	}
}

func TestEmbedOllama_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("model not found"))
	}))
	defer srv.Close()

	_, err := EmbedOllama(srv.URL, "bad-model", "hello")
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/provider/ -run TestEmbedOllama -v 2>&1 | head -10`
Expected: FAIL — function doesn't exist

- [ ] **Step 3: Implement EmbedOllama**

Create `internal/provider/embed.go`:

```go
package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbedOllama calls the Ollama /api/embeddings endpoint and returns the embedding vector.
// baseURL is the Ollama API base (e.g. "http://localhost:11434").
func EmbedOllama(baseURL, model, text string) ([]float64, error) {
	body, _ := json.Marshal(map[string]string{
		"model":  model,
		"prompt": text,
	})

	resp, err := http.Post(baseURL+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed ollama: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embed ollama: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("embed ollama: %s\n%s", resp.Status, data)
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("embed ollama: parse: %w", err)
	}
	return result.Embedding, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/provider/ -run TestEmbedOllama -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/provider/embed.go internal/provider/embed_test.go
git commit -m "feat(provider): add EmbedOllama for Ollama /api/embeddings endpoint"
```

---

### Task 5: Workspace ES URL Config

Add an `Elasticsearch` field to the workspace config so ES forms can resolve their URL from workspace settings.

**Files:**
- Modify: `internal/workspace/workspace.go`
- Modify: `internal/workspace/workspace_test.go`

- [ ] **Step 1: Write test for ES URL in workspace config**

Add to `internal/workspace/workspace_test.go`:

```go
func TestParseWorkspace_Elasticsearch(t *testing.T) {
	src := []byte(`
(workspace "test"
  :description "test workspace"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://es.internal:9200"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.Elasticsearch != "http://es.internal:9200" {
		t.Fatalf("elasticsearch = %q, want http://es.internal:9200", w.Defaults.Elasticsearch)
	}
}

func TestParseWorkspace_ElasticsearchDefault(t *testing.T) {
	src := []byte(`
(workspace "test"
  (defaults :model "qwen2.5:7b"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.Elasticsearch != "" {
		t.Fatalf("elasticsearch should be empty by default, got %q", w.Defaults.Elasticsearch)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/workspace/ -run TestParseWorkspace_Elasticsearch -v 2>&1 | head -20`
Expected: FAIL — unknown keyword

- [ ] **Step 3: Add Elasticsearch field to Defaults**

In `internal/workspace/workspace.go`, update the `Defaults` struct (line 19):

```go
// Defaults holds default model/provider settings for a workspace.
type Defaults struct {
	Model         string
	Provider      string
	Elasticsearch string
}
```

And in `convertDefaults` (line 129), add the case:

```go
			case "elasticsearch":
				d.Elasticsearch = val.StringVal()
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/workspace/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/workspace/workspace.go internal/workspace/workspace_test.go
git commit -m "feat(workspace): add elasticsearch URL to workspace defaults"
```

---

### Task 6: DSL Types for ES + Embed Forms

Add the new step types and fields to `types.go`.

**Files:**
- Modify: `internal/pipeline/types.go:42-80`

- [ ] **Step 1: Add new types and fields**

In `internal/pipeline/types.go`, add new fields to the `Step` struct after line 79 (`GlobPat`):

```go
	// ES forms
	Search *SearchStep `yaml:"-"`
	Index  *IndexStep  `yaml:"-"`
	Delete *DeleteStep `yaml:"-"`

	// Embedding
	Embed *EmbedStep `yaml:"-"`
```

Add the new type definitions after the existing `GlobStep` type (find it near the end of the file):

```go
// SearchStep queries Elasticsearch and returns hits as JSON.
type SearchStep struct {
	IndexName string
	Query     string   // raw JSON query body
	Size      int      // max hits (default 10)
	Fields    []string // _source field filter
	ESURL     string   // override ES URL (empty = workspace default)
}

// IndexStep indexes a single document into Elasticsearch.
type IndexStep struct {
	IndexName     string
	Doc           string // template-rendered JSON document
	DocID         string // optional explicit _id
	ESURL         string
	EmbedField    string // field in doc to embed (empty = no embedding)
	EmbedProvider string
	EmbedModel    string
}

// DeleteStep deletes documents matching a query from Elasticsearch.
type DeleteStep struct {
	IndexName string
	Query     string // raw JSON query body
	ESURL     string
}

// EmbedStep generates an embedding vector from text.
type EmbedStep struct {
	Input    string // template-rendered text to embed
	Provider string
	Model    string
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go build ./internal/pipeline/...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/pipeline/types.go
git commit -m "feat(dsl): add SearchStep, IndexStep, DeleteStep, EmbedStep types"
```

---

### Task 7: Parser — search, index, delete, embed Converters

Add sexpr→Step conversion for the four new forms.

**Files:**
- Modify: `internal/pipeline/sexpr.go`
- Modify: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write parser tests**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Search(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "q" (search :index "my-index" :query {"term" {"type" "doc"}} :size 50 :fields ("title" "content"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if s.Search.IndexName != "my-index" {
		t.Fatalf("index = %q, want my-index", s.Search.IndexName)
	}
	if s.Search.Size != 50 {
		t.Fatalf("size = %d, want 50", s.Search.Size)
	}
	if len(s.Search.Fields) != 2 || s.Search.Fields[0] != "title" || s.Search.Fields[1] != "content" {
		t.Fatalf("fields = %v, want [title content]", s.Search.Fields)
	}
}

func TestSexprWorkflow_SearchDefaultSize(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "q" (search :index "my-index" :query {"term" {"type" "doc"}})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].Search.Size != 10 {
		t.Fatalf("default size = %d, want 10", w.Steps[0].Search.Size)
	}
}

func TestSexprWorkflow_SearchWithES(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "q" (search :index "my-index" :query {"match_all" {}} :es "http://remote:9200")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].Search.ESURL != "http://remote:9200" {
		t.Fatalf("es url = %q, want http://remote:9200", w.Steps[0].Search.ESURL)
	}
}

func TestSexprWorkflow_Index(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "idx" (index :index "my-index" :doc "{{step \"prev\"}}" :id "doc-1")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Index == nil {
		t.Fatal("expected index step")
	}
	if s.Index.IndexName != "my-index" {
		t.Fatalf("index = %q", s.Index.IndexName)
	}
	if s.Index.DocID != "doc-1" {
		t.Fatalf("doc id = %q", s.Index.DocID)
	}
}

func TestSexprWorkflow_IndexWithEmbed(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "idx" (index :index "my-index" :doc "{}" :embed :field "content" :provider "ollama" :model "nomic-embed-text")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Index.EmbedField != "content" {
		t.Fatalf("embed field = %q, want content", s.Index.EmbedField)
	}
	if s.Index.EmbedProvider != "ollama" {
		t.Fatalf("embed provider = %q, want ollama", s.Index.EmbedProvider)
	}
	if s.Index.EmbedModel != "nomic-embed-text" {
		t.Fatalf("embed model = %q, want nomic-embed-text", s.Index.EmbedModel)
	}
}

func TestSexprWorkflow_Delete(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "del" (delete :index "my-index" :query {"term" {"type" "old"}})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Delete == nil {
		t.Fatal("expected delete step")
	}
	if s.Delete.IndexName != "my-index" {
		t.Fatalf("index = %q", s.Delete.IndexName)
	}
}

func TestSexprWorkflow_Embed(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "vec" (embed :input "hello world" :provider "ollama" :model "nomic-embed-text")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Embed == nil {
		t.Fatal("expected embed step")
	}
	if s.Embed.Input != "hello world" {
		t.Fatalf("input = %q", s.Embed.Input)
	}
	if s.Embed.Provider != "ollama" {
		t.Fatalf("provider = %q", s.Embed.Provider)
	}
	if s.Embed.Model != "nomic-embed-text" {
		t.Fatalf("model = %q", s.Embed.Model)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/ -run "TestSexprWorkflow_(Search|Index|Delete|Embed)" -v 2>&1 | head -20`
Expected: FAIL — unknown step types

- [ ] **Step 3: Implement converter functions**

Add to `internal/pipeline/sexpr.go`, after the existing `convertGlobStep` function:

```go
func convertSearch(n *sexpr.Node, defs map[string]string) (*SearchStep, error) {
	children := n.Children[1:]
	s := &SearchStep{Size: 10}

	i := 0
	for i < len(children) {
		child := children[i]
		if !child.IsAtom() || child.Atom.Type != sexpr.TokenKeyword {
			return nil, fmt.Errorf("line %d: (search) expected keyword, got %v", child.Line, child)
		}
		kw := child.KeywordVal()
		i++
		if i >= len(children) {
			return nil, fmt.Errorf("line %d: (search) keyword :%s missing value", child.Line, kw)
		}
		val := children[i]
		switch kw {
		case "index":
			s.IndexName = resolveVal(val, defs)
		case "query":
			b, err := nodeToJSON(val)
			if err != nil {
				return nil, fmt.Errorf("line %d: (search) :query: %w", val.Line, err)
			}
			s.Query = string(b)
		case "size":
			n, err := strconv.Atoi(resolveVal(val, defs))
			if err != nil {
				return nil, fmt.Errorf("line %d: (search) :size must be integer", val.Line)
			}
			s.Size = n
		case "fields":
			if !val.IsList() {
				return nil, fmt.Errorf("line %d: (search) :fields must be a list", val.Line)
			}
			for _, f := range val.Children {
				s.Fields = append(s.Fields, resolveVal(f, defs))
			}
		case "es":
			s.ESURL = resolveVal(val, defs)
		default:
			return nil, fmt.Errorf("line %d: (search) unknown keyword :%s", child.Line, kw)
		}
		i++
	}

	if s.IndexName == "" {
		return nil, fmt.Errorf("line %d: (search) missing :index", n.Line)
	}
	if s.Query == "" {
		return nil, fmt.Errorf("line %d: (search) missing :query", n.Line)
	}
	return s, nil
}

func convertIndex(n *sexpr.Node, defs map[string]string) (*IndexStep, error) {
	children := n.Children[1:]
	s := &IndexStep{}

	i := 0
	for i < len(children) {
		child := children[i]
		if !child.IsAtom() || child.Atom.Type != sexpr.TokenKeyword {
			return nil, fmt.Errorf("line %d: (index) expected keyword", child.Line)
		}
		kw := child.KeywordVal()
		i++
		if kw == "embed" {
			// :embed is followed by sub-keywords :field, :provider, :model
			for i < len(children) {
				sub := children[i]
				if !sub.IsAtom() || sub.Atom.Type != sexpr.TokenKeyword {
					break
				}
				subKw := sub.KeywordVal()
				if subKw != "field" && subKw != "provider" && subKw != "model" {
					break // not an embed sub-keyword, back to main loop
				}
				i++
				if i >= len(children) {
					return nil, fmt.Errorf("line %d: (index) :embed :%s missing value", sub.Line, subKw)
				}
				subVal := resolveVal(children[i], defs)
				switch subKw {
				case "field":
					s.EmbedField = subVal
				case "provider":
					s.EmbedProvider = subVal
				case "model":
					s.EmbedModel = subVal
				}
				i++
			}
			continue
		}
		if i >= len(children) {
			return nil, fmt.Errorf("line %d: (index) keyword :%s missing value", child.Line, kw)
		}
		val := children[i]
		switch kw {
		case "index":
			s.IndexName = resolveVal(val, defs)
		case "doc":
			s.Doc = resolveVal(val, defs)
		case "id":
			s.DocID = resolveVal(val, defs)
		case "es":
			s.ESURL = resolveVal(val, defs)
		default:
			return nil, fmt.Errorf("line %d: (index) unknown keyword :%s", child.Line, kw)
		}
		i++
	}

	if s.IndexName == "" {
		return nil, fmt.Errorf("line %d: (index) missing :index", n.Line)
	}
	if s.Doc == "" {
		return nil, fmt.Errorf("line %d: (index) missing :doc", n.Line)
	}
	return s, nil
}

func convertDelete(n *sexpr.Node, defs map[string]string) (*DeleteStep, error) {
	children := n.Children[1:]
	s := &DeleteStep{}

	i := 0
	for i < len(children) {
		child := children[i]
		if !child.IsAtom() || child.Atom.Type != sexpr.TokenKeyword {
			return nil, fmt.Errorf("line %d: (delete) expected keyword", child.Line)
		}
		kw := child.KeywordVal()
		i++
		if i >= len(children) {
			return nil, fmt.Errorf("line %d: (delete) keyword :%s missing value", child.Line, kw)
		}
		val := children[i]
		switch kw {
		case "index":
			s.IndexName = resolveVal(val, defs)
		case "query":
			b, err := nodeToJSON(val)
			if err != nil {
				return nil, fmt.Errorf("line %d: (delete) :query: %w", val.Line, err)
			}
			s.Query = string(b)
		case "es":
			s.ESURL = resolveVal(val, defs)
		default:
			return nil, fmt.Errorf("line %d: (delete) unknown keyword :%s", child.Line, kw)
		}
		i++
	}

	if s.IndexName == "" {
		return nil, fmt.Errorf("line %d: (delete) missing :index", n.Line)
	}
	if s.Query == "" {
		return nil, fmt.Errorf("line %d: (delete) missing :query", n.Line)
	}
	return s, nil
}

func convertEmbed(n *sexpr.Node, defs map[string]string) (*EmbedStep, error) {
	children := n.Children[1:]
	s := &EmbedStep{}

	i := 0
	for i < len(children) {
		child := children[i]
		if !child.IsAtom() || child.Atom.Type != sexpr.TokenKeyword {
			return nil, fmt.Errorf("line %d: (embed) expected keyword", child.Line)
		}
		kw := child.KeywordVal()
		i++
		if i >= len(children) {
			return nil, fmt.Errorf("line %d: (embed) keyword :%s missing value", child.Line, kw)
		}
		val := children[i]
		switch kw {
		case "input":
			s.Input = resolveVal(val, defs)
		case "provider":
			s.Provider = resolveVal(val, defs)
		case "model":
			s.Model = resolveVal(val, defs)
		default:
			return nil, fmt.Errorf("line %d: (embed) unknown keyword :%s", child.Line, kw)
		}
		i++
	}

	if s.Input == "" {
		return nil, fmt.Errorf("line %d: (embed) missing :input", n.Line)
	}
	if s.Provider == "" {
		return nil, fmt.Errorf("line %d: (embed) missing :provider", n.Line)
	}
	if s.Model == "" {
		return nil, fmt.Errorf("line %d: (embed) missing :model", n.Line)
	}
	return s, nil
}
```

Now add a `nodeToJSON` helper function. Check if it already exists first — if not, add it:

```go
// nodeToJSON converts a sexpr map node ({...}) to a JSON byte slice.
func nodeToJSON(n *sexpr.Node) ([]byte, error) {
	if n.IsAtom() {
		// Bare string value
		return json.Marshal(n.StringVal())
	}

	// Map node: children are alternating key, value pairs
	if n.Atom == nil && len(n.Children) > 0 && n.Children[0].IsAtom() {
		// Check if it looks like a map (even-length children, string keys)
		// vs a list
		isMap := true
		for i := 0; i < len(n.Children); i += 2 {
			if i+1 >= len(n.Children) {
				isMap = false
				break
			}
			if !n.Children[i].IsAtom() {
				isMap = false
				break
			}
		}
		if isMap && len(n.Children)%2 == 0 {
			result := make(map[string]json.RawMessage)
			for i := 0; i < len(n.Children); i += 2 {
				key := n.Children[i].StringVal()
				val, err := nodeToJSON(n.Children[i+1])
				if err != nil {
					return nil, err
				}
				result[key] = val
			}
			return json.Marshal(result)
		}
	}

	// List node
	var items []json.RawMessage
	for _, child := range n.Children {
		val, err := nodeToJSON(child)
		if err != nil {
			return nil, err
		}
		items = append(items, val)
	}
	return json.Marshal(items)
}
```

**Important:** Check if `nodeToJSON` or a similar function already exists in sexpr.go before adding. If a `{...}` map syntax is already handled (for http-get headers), look at how `convertHttpCall` parses map nodes and reuse that pattern.

- [ ] **Step 4: Wire converters into convertStep switch**

In `internal/pipeline/sexpr.go`, add these cases in the `convertStep` switch (after the `glob` case, before the `plugin` case):

```go
		case "search":
			sr, err := convertSearch(child, defs)
			if err != nil {
				return s, err
			}
			s.Search = sr
		case "index":
			idx, err := convertIndex(child, defs)
			if err != nil {
				return s, err
			}
			s.Index = idx
		case "delete":
			del, err := convertDelete(child, defs)
			if err != nil {
				return s, err
			}
			s.Delete = del
		case "embed":
			emb, err := convertEmbed(child, defs)
			if err != nil {
				return s, err
			}
			s.Embed = emb
```

Also add `"strconv"` to the imports in sexpr.go if not already present.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/ -run "TestSexprWorkflow_(Search|Index|Delete|Embed)" -v`
Expected: All PASS

- [ ] **Step 6: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/... ./internal/sexpr/...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add search, index, delete, embed form parsers"
```

---

### Task 8: Runner — Execute ES + Embed Forms

Wire the new form types into `runSingleStep` so they actually execute.

**Files:**
- Modify: `internal/pipeline/runner.go`
- Modify: `internal/pipeline/runner.go` (RunOpts / runCtx for ES URL)

- [ ] **Step 1: Add ESURL to RunOpts and runCtx**

In `internal/pipeline/runner.go`, add to `RunOpts` (line 72):

```go
	ESURL string // default ES URL from workspace config
```

And add to `runCtx` (line 40):

```go
	esURL string
```

In the `Run` function where `rctx` is constructed (around line 153), add:

```go
	var esURL string
	if len(opts) > 0 && opts[0].ESURL != "" {
		esURL = opts[0].ESURL
	}
```

And set it on the rctx:

```go
	rctx := &runCtx{
		// ... existing fields ...
		esURL: esURL,
	}
```

Do the same for the second rctx construction (around line 362 for the non-Items path).

- [ ] **Step 2: Add a resolveESURL helper**

Add to `internal/pipeline/runner.go`:

```go
// resolveESURL returns the ES URL to use, checking step override, then runCtx default, then fallback.
func resolveESURL(stepURL string, rctx *runCtx) string {
	if stepURL != "" {
		return stepURL
	}
	if rctx.esURL != "" {
		return rctx.esURL
	}
	return "http://localhost:9200"
}
```

- [ ] **Step 3: Add execution logic in runSingleStep**

In `internal/pipeline/runner.go`, add these blocks in `runSingleStep` before the `PluginCall` check (before line 1145):

```go
	if step.Search != nil {
		indexRendered, err := render(step.Search.IndexName, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: index template: %w", step.ID, err)
		}
		esURL := resolveESURL(step.Search.ESURL, rctx)
		es := esearch.NewClient(esURL)

		// Build the full query body with size and _source
		queryBody := map[string]any{}
		if step.Search.Query != "" {
			var q any
			if err := json.Unmarshal([]byte(step.Search.Query), &q); err != nil {
				return nil, fmt.Errorf("step %s: query parse: %w", step.ID, err)
			}
			queryBody["query"] = q
		}
		queryBody["size"] = step.Search.Size
		if len(step.Search.Fields) > 0 {
			queryBody["_source"] = step.Search.Fields
		}

		queryJSON, err := json.Marshal(queryBody)
		if err != nil {
			return nil, fmt.Errorf("step %s: query marshal: %w", step.ID, err)
		}

		ui.StepSDK(step.ID, "search")
		resp, err := es.Search(ctx, []string{indexRendered}, queryJSON)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}

		// Return just the _source objects as a JSON array
		var sources []json.RawMessage
		for _, hit := range resp.Results {
			sources = append(sources, hit.Source)
		}
		out, err := json.Marshal(sources)
		if err != nil {
			return nil, fmt.Errorf("step %s: marshal results: %w", step.ID, err)
		}
		return &stepOutcome{output: string(out)}, nil
	}

	if step.Index != nil {
		indexRendered, err := render(step.Index.IndexName, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: index template: %w", step.ID, err)
		}
		docRendered, err := render(step.Index.Doc, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: doc template: %w", step.ID, err)
		}
		idRendered := step.Index.DocID
		if idRendered != "" {
			idRendered, err = render(idRendered, data, stepsSnap)
			if err != nil {
				return nil, fmt.Errorf("step %s: id template: %w", step.ID, err)
			}
		}

		docBytes := []byte(docRendered)

		// Handle embedding if configured
		if step.Index.EmbedField != "" {
			var docMap map[string]any
			if err := json.Unmarshal(docBytes, &docMap); err != nil {
				return nil, fmt.Errorf("step %s: parse doc for embedding: %w", step.ID, err)
			}
			fieldVal, ok := docMap[step.Index.EmbedField]
			if ok {
				text := fmt.Sprintf("%v", fieldVal)
				vec, err := provider.EmbedOllama("http://localhost:11434", step.Index.EmbedModel, text)
				if err != nil {
					return nil, fmt.Errorf("step %s: embed: %w", step.ID, err)
				}
				docMap["embedding"] = vec
				docBytes, err = json.Marshal(docMap)
				if err != nil {
					return nil, fmt.Errorf("step %s: marshal embedded doc: %w", step.ID, err)
				}
			}
		}

		esURL := resolveESURL(step.Index.ESURL, rctx)
		es := esearch.NewClient(esURL)
		ui.StepSDK(step.ID, "index")
		resp, err := es.IndexDoc(ctx, indexRendered, idRendered, docBytes)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		out, _ := json.Marshal(resp)
		return &stepOutcome{output: string(out)}, nil
	}

	if step.Delete != nil {
		indexRendered, err := render(step.Delete.IndexName, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: index template: %w", step.ID, err)
		}
		esURL := resolveESURL(step.Delete.ESURL, rctx)
		es := esearch.NewClient(esURL)

		ui.StepSDK(step.ID, "delete")
		resp, err := es.DeleteByQuery(ctx, indexRendered, json.RawMessage(step.Delete.Query))
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		out, _ := json.Marshal(resp)
		return &stepOutcome{output: string(out)}, nil
	}

	if step.Embed != nil {
		rendered, err := render(step.Embed.Input, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: input template: %w", step.ID, err)
		}
		ui.StepSDK(step.ID, "embed")
		vec, err := provider.EmbedOllama("http://localhost:11434", step.Embed.Model, rendered)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		out, _ := json.Marshal(vec)
		return &stepOutcome{output: string(out)}, nil
	}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go build ./internal/pipeline/...`
Expected: No errors

- [ ] **Step 5: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go test ./internal/pipeline/... ./internal/sexpr/... ./internal/esearch/...`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add internal/pipeline/runner.go
git commit -m "feat(dsl): execute search, index, delete, embed forms in runner"
```

---

### Task 9: Wire Workspace ES URL Through to Runner

Pass the workspace ES URL from the CLI layer into `RunOpts.ESURL`.

**Files:**
- Modify: `cmd/workflow.go` (where `pipeline.Run` is called)

- [ ] **Step 1: Find where pipeline.Run is called with RunOpts**

Look at `cmd/workflow.go:109` where `esClient := esearch.NewClient("http://localhost:9200")` is set. The workspace is likely loaded nearby.

- [ ] **Step 2: Pass ES URL into RunOpts**

Where `pipeline.RunOpts` is constructed in `cmd/workflow.go`, add:

```go
ESURL: wsDefaults.Elasticsearch, // from workspace config; empty string falls through to default
```

The exact field name depends on how the workspace is accessed in the command. If the workspace `Defaults` struct is available as e.g. `ws.Defaults`, use `ws.Defaults.Elasticsearch`.

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add cmd/workflow.go
git commit -m "feat(dsl): pass workspace ES URL to pipeline runner"
```

---

### Task 10: Update Documentation

Update the workflow syntax docs to use new form names and document new forms.

**Files:**
- Modify: `docs/site/workflow-syntax.md`

- [ ] **Step 1: Add new form documentation**

Add sections for `search`, `index`, `delete`, `embed` forms with syntax and examples. Add the alias table showing old→new names. Document the template string functions.

- [ ] **Step 2: Update existing examples to use new names**

Replace `json-pick` → `pick`, `http-get` → `fetch`, `http-post` → `send`, `read-file` → `read`, `write-file` → `write`, `map` → `each` in examples throughout the doc.

- [ ] **Step 3: Commit**

```bash
cd /Users/stokes/Projects/gl1tch/.worktrees/dsl-improvements
git add docs/site/workflow-syntax.md
git commit -m "docs: update workflow syntax with new form names, ES forms, template functions"
```
