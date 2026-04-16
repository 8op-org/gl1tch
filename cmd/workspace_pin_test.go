package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkspacePinPreservesComments verifies that `runWorkspacePin` does not
// strip comments from workspace.glitch. The previous implementation called
// workspace.Serialize(), which regenerated the file from scratch and lost
// every comment on the first pin.
func TestWorkspacePinPreservesComments(t *testing.T) {
	ws := t.TempDir()
	src := `(workspace "t"
  ;; keep me
  (resource "local-thing" :type "local" :path "/tmp"))
`
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runWorkspacePin(ws, "local-thing", "newref"); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), ";; keep me") {
		t.Fatalf("comment lost:\n%s", out)
	}
	if !strings.Contains(string(out), `:ref "newref"`) {
		t.Fatalf("ref not updated:\n%s", out)
	}
}

// TestWorkspacePinUnknownResource returns a useful error rather than silently
// doing nothing when the named resource is not present.
func TestWorkspacePinUnknownResource(t *testing.T) {
	ws := t.TempDir()
	src := `(workspace "t")` + "\n"
	if err := os.WriteFile(filepath.Join(ws, "workspace.glitch"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runWorkspacePin(ws, "ghost", "x"); err == nil {
		t.Fatal("expected error for unknown resource")
	}
}
