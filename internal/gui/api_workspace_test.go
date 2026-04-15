package gui

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetWorkspace(t *testing.T) {
	srv := testServer(t)
	src := `(workspace "test-ws" :description "a test" :owner "adam" (repos "elastic/ensemble") (defaults :model "qwen2.5:7b" :provider "ollama"))`
	os.WriteFile(filepath.Join(srv.workspace, "workspace.glitch"), []byte(src), 0o644)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workspace", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != "test-ws" {
		t.Errorf("expected name test-ws, got %v", resp["name"])
	}
	if resp["owner"] != "adam" {
		t.Errorf("expected owner adam, got %v", resp["owner"])
	}
}

func TestGetWorkspace_Missing(t *testing.T) {
	srv := testServer(t)

	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/workspace", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200 even when missing, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != nil && resp["name"] != "" {
		t.Errorf("expected empty name for missing workspace, got %v", resp["name"])
	}
}
