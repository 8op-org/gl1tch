package gui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/store"
)

func TestSmokeEndToEnd(t *testing.T) {
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	resDir := filepath.Join(dir, "results", "test-repo", "issue-123")
	os.MkdirAll(wfDir, 0o755)
	os.MkdirAll(resDir, 0o755)

	// Create a test workflow
	os.WriteFile(filepath.Join(wfDir, "smoke.glitch"), []byte(`(workflow "smoke" :description "smoke test" (step "hello" (run "echo hi")))`), 0o644)

	// Create result files
	os.WriteFile(filepath.Join(resDir, "plan.md"), []byte("# Plan\nAll good."), 0o644)
	os.WriteFile(filepath.Join(resDir, "classification.json"), []byte(`{"type":"bug"}`), 0o644)

	// Set up store with a test run
	st, err := store.OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	runID, err := st.RecordRun("workflow", "smoke", "test input")
	if err != nil {
		t.Fatal(err)
	}
	st.RecordStep(runID, "hello", "prompt", "output", "qwen3-8b", 2500)
	st.FinishRun(runID, "echo hi", 0)

	srv := &Server{workspace: dir, store: st, mux: http.NewServeMux(), dev: true}
	srv.routes()

	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		wantCode     int
		wantContains string
	}{
		{"list workflows", "GET", "/api/workflows", "", 200, "smoke"},
		{"get workflow", "GET", "/api/workflows/smoke.glitch", "", 200, "smoke"},
		{"save workflow", "PUT", "/api/workflows/smoke.glitch", `{"source":"(workflow \"smoke2\" (step \"s\" (run \"echo\")))"}`, 200, "saved"},
		{"list runs", "GET", "/api/runs", "", 200, "smoke"},
		{"get run", "GET", "/api/runs/1", "", 200, "hello"},
		{"list results dir", "GET", "/api/results/test-repo", "", 200, "issue-123"},
		{"get result subdir", "GET", "/api/results/test-repo/issue-123", "", 200, "plan.md"},
		{"get markdown file", "GET", "/api/results/test-repo/issue-123/plan.md", "", 200, "# Plan"},
		{"get json file", "GET", "/api/results/test-repo/issue-123/classification.json", "", 200, "bug"},
		{"kibana workflow", "GET", "/api/kibana/workflow/smoke", "", 200, "url"},
		{"kibana run", "GET", "/api/kibana/run/test-run-id", "", 200, "url"},
		{"not found workflow", "GET", "/api/workflows/nope.glitch", "", 404, ""},
		{"not found result", "GET", "/api/results/nope/nope.md", "", 404, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			srv.mux.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}
			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body missing %q: %s", tt.wantContains, w.Body.String())
			}
		})
	}

	// Path traversal: call handlers directly to bypass mux URL cleaning
	t.Run("path traversal blocked (workflow)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/workflows/..%2Fsecrets", nil)
		req.SetPathValue("name", "../secrets")
		w := httptest.NewRecorder()
		srv.handleGetWorkflow(w, req)
		if w.Code != 400 {
			t.Errorf("expected 400 for path traversal, got %d", w.Code)
		}
	})

	t.Run("path traversal blocked (result)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/results/..%2Fetc%2Fpasswd", nil)
		req.SetPathValue("path", "../etc/passwd")
		w := httptest.NewRecorder()
		srv.handleGetResult(w, req)
		if w.Code != 400 {
			t.Errorf("expected 400 for path traversal, got %d", w.Code)
		}
	})

	// Verify the PUT actually changed the file
	data, _ := os.ReadFile(filepath.Join(wfDir, "smoke.glitch"))
	if !strings.Contains(string(data), "smoke2") {
		t.Error("PUT did not persist the workflow change")
	}
}
