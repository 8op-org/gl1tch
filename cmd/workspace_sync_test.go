package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace"
)

func TestWorkspaceSyncLocalOnly(t *testing.T) {
	ws := t.TempDir()
	target := t.TempDir()
	src := `(workspace "demo" (resource "notes" :type "local" :path "` + target + `"))` + "\n"
	_ = os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(src), 0o644)

	if err := runWorkspaceSync(ws, nil, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(ws, "resources", "notes")); err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	st, err := workspace.LoadResourceState(ws)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.Entries["notes"]; !ok {
		t.Fatal("state not recorded")
	}
}
