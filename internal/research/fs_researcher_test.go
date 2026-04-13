package research

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFSResearcherName(t *testing.T) {
	f := &FSResearcher{}
	if f.Name() != "fs" {
		t.Fatalf("got %q, want %q", f.Name(), "fs")
	}
}

func TestFSResearcherGatherScan(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "docs"), 0o755)
	os.WriteFile(filepath.Join(dir, "docs", "test.md"), []byte("# Title\n\n> TBC\n\nSome content"), 0o644)

	f := &FSResearcher{RootPath: dir}
	q := ResearchQuery{Question: "what placeholders are in the docs?"}
	ev, err := f.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if !strings.Contains(ev.Body, "TBC") {
		t.Fatal("expected body to contain TBC scan results")
	}
}

func TestFSResearcherGatherTree(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src", "pkg"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "pkg", "main.go"), []byte("package main"), 0o644)

	f := &FSResearcher{RootPath: dir}
	q := ResearchQuery{Question: "what is the project structure?"}
	ev, err := f.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if !strings.Contains(ev.Body, "main.go") {
		t.Fatalf("expected tree output, got: %s", ev.Body)
	}
}
