package research

import (
	"context"
	"strings"
	"testing"
)

func TestGitResearcherName(t *testing.T) {
	g := &GitResearcher{}
	if g.Name() != "git" {
		t.Fatalf("got %q, want %q", g.Name(), "git")
	}
}

func TestGitResearcherGather(t *testing.T) {
	g := &GitResearcher{}
	q := ResearchQuery{Question: "what recent commits were made?"}
	ev, err := g.Gather(context.Background(), q, EvidenceBundle{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if ev.Source != "git" {
		t.Fatalf("source: got %q, want %q", ev.Source, "git")
	}
	if ev.Body == "" {
		t.Fatal("expected non-empty body")
	}
	if !strings.ContainsAny(ev.Body, "0123456789abcdef") {
		t.Fatal("expected git log output with commit hashes")
	}
}
