package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_WorkspaceStore(t *testing.T) {
	wsDir := t.TempDir()
	os.MkdirAll(filepath.Join(wsDir, "workflows"), 0o755)

	s, err := New(":0", wsDir, false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.store.Close()

	dbPath := filepath.Join(wsDir, ".glitch", "glitch.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected workspace DB at %s: %v", dbPath, err)
	}
}
