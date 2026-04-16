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

func setupTestWS(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GLITCH_WORKSPACE", "")
	ws := filepath.Join(home, "ws")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir ws: %v", err)
	}
	src := `(workspace "t" (resource "notes" :type "local" :path "/tmp"))` + "\n"
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(src), 0o644); err != nil {
		t.Fatalf("write workspace.glitch: %v", err)
	}
	return ws
}

func TestListResourcesEmptyOrPopulated(t *testing.T) {
	ws := setupTestWS(t)
	s := &Server{workspace: ws}
	req := httptest.NewRequest(http.MethodGet, "/api/workspace/resources", nil)
	rec := httptest.NewRecorder()
	s.handleListResources(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("not JSON array: %s", rec.Body.String())
	}
	if len(out) != 1 || out[0]["name"] != "notes" {
		t.Fatalf("unexpected resources: %+v", out)
	}
	if out[0]["type"] != "local" {
		t.Fatalf("expected type local, got %+v", out[0])
	}
}

func TestListResourcesNoWorkspace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GLITCH_WORKSPACE", "")
	s := &Server{workspace: ""}
	req := httptest.NewRequest(http.MethodGet, "/api/workspace/resources", nil)
	rec := httptest.NewRecorder()
	s.handleListResources(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("not JSON: %s", rec.Body.String())
	}
	if body["error"] == "" {
		t.Fatalf("expected error field, got %+v", body)
	}
}

func TestAddResourceLocal(t *testing.T) {
	ws := setupTestWS(t)
	// Reset workspace to no resources.
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"),
		[]byte(`(workspace "t")`+"\n"), 0o644); err != nil {
		t.Fatalf("write workspace.glitch: %v", err)
	}

	s := &Server{workspace: ws}
	target := t.TempDir()
	body := strings.NewReader(`{"input":"` + target + `","name":"data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/workspace/resources", body)
	rec := httptest.NewRecorder()
	s.handleAddResource(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "data")); err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	// Verify workspace.glitch was updated.
	data, err := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	if err != nil {
		t.Fatalf("read workspace: %v", err)
	}
	if !strings.Contains(string(data), `"data"`) {
		t.Fatalf("resource not written to workspace.glitch: %s", data)
	}
}

func TestAddResourceDuplicate(t *testing.T) {
	ws := setupTestWS(t)
	s := &Server{workspace: ws}
	body := strings.NewReader(`{"input":"/tmp","name":"notes"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/workspace/resources", body)
	rec := httptest.NewRecorder()
	s.handleAddResource(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddResourceNoWorkspace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GLITCH_WORKSPACE", "")
	s := &Server{workspace: ""}
	body := strings.NewReader(`{"input":"/tmp","name":"x"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/workspace/resources", body)
	rec := httptest.NewRecorder()
	s.handleAddResource(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRemoveResource(t *testing.T) {
	ws := setupTestWS(t)
	if err := os.MkdirAll(filepath.Join(ws, "resources"), 0o755); err != nil {
		t.Fatalf("mkdir resources: %v", err)
	}
	if err := os.Symlink("/tmp", filepath.Join(ws, "resources", "notes")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	s := &Server{workspace: ws}
	req := httptest.NewRequest(http.MethodDelete, "/api/workspace/resources/notes", nil)
	req.SetPathValue("name", "notes")
	rec := httptest.NewRecorder()
	s.handleRemoveResource(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "notes")); !os.IsNotExist(err) {
		t.Fatalf("symlink still present: %v", err)
	}
	// Verify workspace.glitch no longer contains the resource.
	data, _ := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	if strings.Contains(string(data), `"notes"`) {
		t.Fatalf("resource still in workspace.glitch: %s", data)
	}
}

func TestRemoveResourceNotFound(t *testing.T) {
	ws := setupTestWS(t)
	s := &Server{workspace: ws}
	req := httptest.NewRequest(http.MethodDelete, "/api/workspace/resources/missing", nil)
	req.SetPathValue("name", "missing")
	rec := httptest.NewRecorder()
	s.handleRemoveResource(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSyncResourceLocal(t *testing.T) {
	ws := setupTestWS(t)
	// Point the workspace's "notes" resource at a tmpdir that definitely exists.
	target := t.TempDir()
	src := `(workspace "t" (resource "notes" :type "local" :path "` + target + `"))` + "\n"
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(src), 0o644); err != nil {
		t.Fatalf("write workspace.glitch: %v", err)
	}

	s := &Server{workspace: ws}
	req := httptest.NewRequest(http.MethodPost, "/api/workspace/sync/notes", nil)
	req.SetPathValue("name", "notes")
	rec := httptest.NewRecorder()
	s.handleSyncResources(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "notes")); err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
}

func TestPinResourceRequiresFields(t *testing.T) {
	ws := setupTestWS(t)
	s := &Server{workspace: ws}
	body := strings.NewReader(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/workspace/pin", body)
	rec := httptest.NewRecorder()
	s.handlePinResource(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInferKindGUI(t *testing.T) {
	cases := map[string]string{
		"https://github.com/foo/bar":     "git",
		"git@github.com:foo/bar.git":     "git",
		"/tmp/foo":                       "local",
		"./foo":                          "local",
		"~/foo":                          "local",
		"elastic/kibana":                 "tracker",
		"":                               "",
	}
	for in, want := range cases {
		if got := inferKindGUI(in); got != want {
			t.Errorf("inferKindGUI(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInferResourceNameGUI(t *testing.T) {
	cases := map[string]string{
		"https://github.com/foo/bar.git": "bar",
		"/tmp/notes":                     "notes",
		"elastic/kibana":                 "kibana",
	}
	for in, want := range cases {
		if got := inferResourceNameGUI(in); got != want {
			t.Errorf("inferResourceNameGUI(%q) = %q, want %q", in, got, want)
		}
	}
}
