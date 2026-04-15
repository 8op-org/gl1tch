package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "workflows"), 0o755)
	os.MkdirAll(filepath.Join(dir, "results"), 0o755)
	srv := &Server{workspace: dir, mux: http.NewServeMux(), dev: true}
	srv.routes()
	return srv
}

func TestListWorkflows(t *testing.T) {
	srv := testServer(t)
	wfDir := filepath.Join(srv.workspace, "workflows")
	os.WriteFile(filepath.Join(wfDir, "hello.glitch"), []byte("(workflow \"hello\" :description \"test\" (step \"s1\" (run \"echo hi\")))"), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var wfs []workflowEntry
	json.Unmarshal(w.Body.Bytes(), &wfs)
	if len(wfs) != 1 {
		t.Fatalf("expected 1, got %d", len(wfs))
	}
	if wfs[0].Name != "hello" {
		t.Errorf("expected name hello, got %q", wfs[0].Name)
	}
}

func TestGetWorkflow(t *testing.T) {
	srv := testServer(t)
	src := "(workflow \"hello\" (step \"s1\" (run \"echo {{.param.repo}}\")))"
	os.WriteFile(filepath.Join(srv.workspace, "workflows", "hello.glitch"), []byte(src), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows/hello.glitch", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["source"] != src {
		t.Error("source mismatch")
	}
	params := resp["params"].([]any)
	if len(params) != 1 || params[0] != "repo" {
		t.Errorf("expected [repo], got %v", params)
	}
}

func TestPutWorkflow(t *testing.T) {
	srv := testServer(t)
	os.WriteFile(filepath.Join(srv.workspace, "workflows", "hello.glitch"), []byte("old"), 0o644)

	body := `{"source": "new content"}`
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("PUT", "/api/workflows/hello.glitch", strings.NewReader(body)))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	data, _ := os.ReadFile(filepath.Join(srv.workspace, "workflows", "hello.glitch"))
	if string(data) != "new content" {
		t.Errorf("file not updated: %q", string(data))
	}
}

func TestPathTraversalBlocked(t *testing.T) {
	srv := testServer(t)
	// Direct handler call to bypass mux URL cleaning
	req := httptest.NewRequest("GET", "/api/workflows/..%2Fsecrets", nil)
	req.SetPathValue("name", "../secrets")
	w := httptest.NewRecorder()
	srv.handleGetWorkflow(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400 for path traversal, got %d", w.Code)
	}
}

func TestListWorkflowsEmpty(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows", nil))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var wfs []workflowEntry
	json.Unmarshal(w.Body.Bytes(), &wfs)
	if len(wfs) != 0 {
		t.Errorf("expected empty list, got %d", len(wfs))
	}
}

func TestListWorkflows_WithMetadata(t *testing.T) {
	srv := testServer(t)
	wfDir := filepath.Join(srv.workspace, "workflows")
	src := `(workflow "pr-review" :description "Review PRs" :tags ("review" "ci") :author "adam" :version "1.0" :created "2026-04-01" (step "s1" (run "echo hi")))`
	os.WriteFile(filepath.Join(wfDir, "pr-review.glitch"), []byte(src), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workflows", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var wfs []workflowEntry
	json.Unmarshal(w.Body.Bytes(), &wfs)
	if len(wfs) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(wfs))
	}
	if len(wfs[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(wfs[0].Tags))
	}
	if wfs[0].Author != "adam" {
		t.Errorf("expected author adam, got %q", wfs[0].Author)
	}
	if wfs[0].Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", wfs[0].Version)
	}
}
