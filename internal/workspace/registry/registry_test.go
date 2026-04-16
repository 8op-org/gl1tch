package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryAddLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	if err := Add(Entry{Name: "alpha", Path: "/tmp/alpha"}); err != nil {
		t.Fatal(err)
	}
	if err := Add(Entry{Name: "beta", Path: "/tmp/beta"}); err != nil {
		t.Fatal(err)
	}
	if err := Add(Entry{Name: "alpha", Path: "/elsewhere"}); err == nil {
		t.Fatal("duplicate name should be rejected")
	}

	entries, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}

	raw, _ := os.ReadFile(filepath.Join(dir, ".config", "glitch", "workspaces.glitch"))
	if len(raw) == 0 {
		t.Fatal("registry file not written")
	}
}

func TestRegistryRemove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = Add(Entry{Name: "x", Path: "/a"})
	_ = Add(Entry{Name: "y", Path: "/b"})
	if err := Remove("x"); err != nil {
		t.Fatal(err)
	}
	entries, _ := List()
	if len(entries) != 1 || entries[0].Name != "y" {
		t.Fatalf("expected only y remaining, got %+v", entries)
	}
}
