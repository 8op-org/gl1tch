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
