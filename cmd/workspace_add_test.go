package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace"
)

func TestWorkspaceAddLocal(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(`(workspace "demo")`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	if err := runWorkspaceAdd(ws, target, "notes", "", ""); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	w, err := workspace.ParseFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Resources) != 1 || w.Resources[0].Name != "notes" || w.Resources[0].Type != "local" {
		t.Fatalf("resource not recorded: %+v", w.Resources)
	}
	link := filepath.Join(ws, "resources", "notes")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
}

func TestWorkspaceAddRejectsPathTraversal(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(`(workspace "demo")`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runWorkspaceAdd(ws, "/tmp", "../evil", "", ""); err == nil {
		t.Fatal("expected error for name with path traversal")
	}
	// Confirm nothing was written outside resources/.
	if _, err := os.Lstat(filepath.Join(ws, "..", "evil")); err == nil {
		t.Fatal("traversal escaped the workspace")
	}
}

func TestWorkspaceRm(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "workspace.glitch"),
		[]byte(`(workspace "demo" (resource "notes" :type "local" :path "/tmp"))`+"\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(ws, "resources"), 0o755)
	_ = os.Symlink("/tmp", filepath.Join(ws, "resources", "notes"))

	if err := runWorkspaceRm(ws, "notes"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	w, _ := workspace.ParseFile(data)
	if len(w.Resources) != 0 {
		t.Fatalf("resource not removed: %+v", w.Resources)
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "notes")); !os.IsNotExist(err) {
		t.Fatalf("symlink not removed: %v", err)
	}
}
