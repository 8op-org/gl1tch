package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

func TestWorkspaceInitCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	target := filepath.Join(dir, "wsx")
	if err := runWorkspaceInit(target, "wsx"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "workspace.glitch")); err != nil {
		t.Fatalf("workspace.glitch missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "workflows")); err != nil {
		t.Fatalf("workflows/ missing: %v", err)
	}
	entries, _ := registry.List()
	if len(entries) != 1 || entries[0].Name != "wsx" {
		t.Fatalf("registry not updated: %+v", entries)
	}
}

func TestWorkspaceInitRejectsExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	target := filepath.Join(dir, "existing")
	_ = os.MkdirAll(target, 0o755)
	_ = os.WriteFile(filepath.Join(target, "workspace.glitch"), []byte(`(workspace "x")`), 0o644)
	if err := runWorkspaceInit(target, "existing"); err == nil {
		t.Fatal("expected error when workspace.glitch already exists")
	}
}

func TestWorkspaceRegisterRejectsDuplicate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	_ = os.MkdirAll(a, 0o755)
	_ = os.MkdirAll(b, 0o755)
	_ = os.WriteFile(filepath.Join(a, "workspace.glitch"), []byte(`(workspace "shared")`), 0o644)
	_ = os.WriteFile(filepath.Join(b, "workspace.glitch"), []byte(`(workspace "shared")`), 0o644)
	if err := runWorkspaceRegister(a, ""); err != nil {
		t.Fatal(err)
	}
	if err := runWorkspaceRegister(b, ""); err == nil {
		t.Fatal("expected duplicate-name error")
	}
}

func TestWorkspaceUnregisterRemoves(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = registry.Add(registry.Entry{Name: "keep", Path: "/k"})
	_ = registry.Add(registry.Entry{Name: "drop", Path: "/d"})
	if err := runWorkspaceUnregister("drop"); err != nil {
		t.Fatal(err)
	}
	entries, _ := registry.List()
	if len(entries) != 1 || entries[0].Name != "keep" {
		t.Fatalf("unregister failed: %+v", entries)
	}
}
