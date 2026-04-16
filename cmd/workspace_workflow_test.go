package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceWorkflowNew(t *testing.T) {
	ws := t.TempDir()
	if err := runWorkspaceWorkflowNew(ws, "hello"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ws, "workflows", "hello.glitch")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("workflow not written: %v", err)
	}
	if !strings.Contains(string(data), `(workflow "hello"`) {
		t.Fatalf("wrong skeleton: %s", data)
	}
}

func TestWorkspaceWorkflowNewRejectsExisting(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "workflows"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "workflows", "x.glitch"), []byte("(workflow \"x\")"), 0o644)
	if err := runWorkspaceWorkflowNew(ws, "x"); err == nil {
		t.Fatal("expected error on existing")
	}
}
