package gui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

func TestListWorkspacesReturnsArray(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = registry.Add(registry.Entry{Name: "alpha", Path: "/tmp/alpha"})
	_ = registry.Add(registry.Entry{Name: "beta", Path: "/tmp/beta"})
	_ = registry.SetActive("alpha")

	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
	rec := httptest.NewRecorder()
	s.handleListWorkspaces(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.HasPrefix(strings.TrimSpace(body), "[") {
		t.Fatalf("expected JSON array, got %s", body)
	}
	if !strings.Contains(body, `"name":"alpha"`) || !strings.Contains(body, `"active":true`) {
		t.Fatalf("active workspace not marked: %s", body)
	}
}

func TestUseWorkspaceSetsActive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = registry.Add(registry.Entry{Name: "alpha", Path: "/tmp/alpha"})
	_ = registry.Add(registry.Entry{Name: "beta", Path: "/tmp/beta"})

	s := &Server{}
	body := strings.NewReader(`{"name":"beta"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces/use", body)
	rec := httptest.NewRecorder()
	s.handleUseWorkspace(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	active, _ := registry.GetActive()
	if active != "beta" {
		t.Fatalf("expected beta active, got %q", active)
	}
}
