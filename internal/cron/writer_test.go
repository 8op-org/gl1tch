package cron

import (
	"path/filepath"
	"testing"
)

func TestWriteEntry_Create(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cron.yaml")

	entry := Entry{Name: "a", Schedule: "* * * * *", Kind: "pipeline", Target: "foo"}
	if err := writeEntry(path, entry); err != nil {
		t.Fatalf("writeEntry: %v", err)
	}
	entries, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "a" {
		t.Fatalf("expected 1 entry named 'a', got %v", entries)
	}
}

func TestWriteEntry_Append(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cron.yaml")

	for _, name := range []string{"a", "b"} {
		if err := writeEntry(path, Entry{Name: name, Schedule: "* * * * *", Kind: "pipeline", Target: name}); err != nil {
			t.Fatalf("writeEntry %s: %v", name, err)
		}
	}
	entries, _ := LoadConfigFrom(path)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestWriteEntry_Upsert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cron.yaml")

	writeEntry(path, Entry{Name: "a", Schedule: "* * * * *", Kind: "pipeline", Target: "old"}) //nolint:errcheck
	writeEntry(path, Entry{Name: "a", Schedule: "0 * * * *", Kind: "pipeline", Target: "new"}) //nolint:errcheck

	entries, _ := LoadConfigFrom(path)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after upsert, got %d", len(entries))
	}
	if entries[0].Target != "new" {
		t.Fatalf("expected updated target 'new', got %q", entries[0].Target)
	}
}

func TestRemoveEntry_Remove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cron.yaml")

	writeEntry(path, Entry{Name: "a", Schedule: "* * * * *", Kind: "pipeline", Target: "a"}) //nolint:errcheck
	writeEntry(path, Entry{Name: "b", Schedule: "* * * * *", Kind: "pipeline", Target: "b"}) //nolint:errcheck

	if err := removeEntry(path, "a"); err != nil {
		t.Fatalf("removeEntry: %v", err)
	}
	entries, _ := LoadConfigFrom(path)
	if len(entries) != 1 || entries[0].Name != "b" {
		t.Fatalf("expected only 'b', got %v", entries)
	}
}

func TestRemoveEntry_NoOp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cron.yaml")

	writeEntry(path, Entry{Name: "a", Schedule: "* * * * *", Kind: "pipeline", Target: "a"}) //nolint:errcheck

	if err := removeEntry(path, "nonexistent"); err != nil {
		t.Fatalf("removeEntry no-op: %v", err)
	}
	entries, _ := LoadConfigFrom(path)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry still, got %d", len(entries))
	}
}

func TestRemoveEntry_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")
	if err := removeEntry(path, "x"); err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}
