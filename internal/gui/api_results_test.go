package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetResult_DirectoryWithoutTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	resultsDir := filepath.Join(dir, "results")
	subDir := filepath.Join(resultsDir, "elastic", "ensemble")
	os.MkdirAll(subDir, 0o755)
	os.WriteFile(filepath.Join(subDir, "data.json"), []byte(`{"ok":true}`), 0o644)

	srv := &Server{workspace: dir}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/results/{path...}", srv.handleGetResult)

	// Request WITHOUT trailing slash — should still return directory listing
	req := httptest.NewRequest("GET", "/api/results/elastic/ensemble", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var entries []struct {
		Name  string `json:"name"`
		IsDir bool   `json:"is_dir"`
	}
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "data.json" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestGetResult_DirectoryWithTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	resultsDir := filepath.Join(dir, "results")
	subDir := filepath.Join(resultsDir, "elastic")
	os.MkdirAll(subDir, 0o755)
	os.WriteFile(filepath.Join(subDir, "readme.md"), []byte("# hi"), 0o644)

	srv := &Server{workspace: dir}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/results/{path...}", srv.handleGetResult)

	req := httptest.NewRequest("GET", "/api/results/elastic/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var entries []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "readme.md" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}
