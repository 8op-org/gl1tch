package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/store"
)

func testServerWithStore(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "workflows"), 0o755)
	os.MkdirAll(filepath.Join(dir, "results"), 0o755)
	st, err := store.OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	srv := &Server{workspace: dir, store: st, mux: http.NewServeMux(), dev: true}
	srv.routes()
	return srv
}

func TestListRuns(t *testing.T) {
	srv := testServerWithStore(t)
	id, _ := srv.store.RecordRun(store.RunRecord{Kind: "workflow", Name: "hello", Input: ""})
	srv.store.FinishRun(id, "done", 0)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/runs", nil))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var runs []runEntry
	json.Unmarshal(w.Body.Bytes(), &runs)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Name != "hello" {
		t.Errorf("expected name hello, got %q", runs[0].Name)
	}
}

func TestGetRun(t *testing.T) {
	srv := testServerWithStore(t)
	id, _ := srv.store.RecordRun(store.RunRecord{Kind: "workflow", Name: "hello", Input: "test input"})
	srv.store.RecordStep(store.StepRecord{RunID: id, StepID: "step1", Prompt: "prompt", Output: "output", Model: "gpt-4", DurationMs: 1500})
	srv.store.FinishRun(id, "done", 0)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/runs/1", nil))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetResultDir(t *testing.T) {
	srv := testServerWithStore(t)
	resDir := filepath.Join(srv.workspace, "results", "test-repo")
	os.MkdirAll(resDir, 0o755)
	os.WriteFile(filepath.Join(resDir, "plan.md"), []byte("# Plan"), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/results/test-repo", nil))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var files []map[string]any
	json.Unmarshal(w.Body.Bytes(), &files)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestGetResultFile(t *testing.T) {
	srv := testServerWithStore(t)
	resDir := filepath.Join(srv.workspace, "results")
	os.WriteFile(filepath.Join(resDir, "output.md"), []byte("# Result"), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/results/output.md", nil))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/markdown" {
		t.Errorf("expected text/markdown, got %s", w.Header().Get("Content-Type"))
	}
}
