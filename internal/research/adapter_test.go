package research

import (
	"testing"
)

func TestAdaptGitHubIssueURL(t *testing.T) {
	// Only test regex detection — don't actually call gh.
	url := "https://github.com/elastic/observability-robots/issues/3928"
	m := reGHIssue.FindStringSubmatch(url)
	if m == nil {
		t.Fatal("expected reGHIssue to match")
	}
	if m[1] != "elastic" || m[2] != "observability-robots" || m[3] != "3928" {
		t.Fatalf("unexpected captures: %v", m)
	}
}

func TestAdaptGitHubPRURL(t *testing.T) {
	url := "https://github.com/elastic/kibana/pull/12345"
	m := reGHPR.FindStringSubmatch(url)
	if m == nil {
		t.Fatal("expected reGHPR to match")
	}
	if m[1] != "elastic" || m[2] != "kibana" || m[3] != "12345" {
		t.Fatalf("unexpected captures: %v", m)
	}
}

func TestAdaptGoogleDocURL(t *testing.T) {
	url := "https://docs.google.com/document/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ"
	m := reGDoc.FindStringSubmatch(url)
	if m == nil {
		t.Fatal("expected reGDoc to match")
	}
	if m[1] != "1aBcDeFgHiJkLmNoPqRsTuVwXyZ" {
		t.Fatalf("unexpected doc ID: %s", m[1])
	}
}

func TestAdaptPlainText(t *testing.T) {
	doc, err := Adapt("What is the root cause of the flaky test?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Source != "text" {
		t.Fatalf("expected source=text, got %s", doc.Source)
	}
	if doc.Title != "What is the root cause of the flaky test?" {
		t.Fatalf("unexpected title: %s", doc.Title)
	}
}

func TestAdaptEmpty(t *testing.T) {
	_, err := Adapt("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestExtractLinks(t *testing.T) {
	text := `See https://github.com/elastic/kibana/issues/100 and
also https://github.com/elastic/kibana/pull/200 for context.
Duplicate: https://github.com/elastic/kibana/issues/100`

	links := extractLinks(text)
	if len(links) != 2 {
		t.Fatalf("expected 2 unique links, got %d: %v", len(links), links)
	}
	if links[0].Label != "elastic/kibana#100" {
		t.Fatalf("unexpected label: %s", links[0].Label)
	}
	if links[1].Label != "elastic/kibana#200 (PR)" {
		t.Fatalf("unexpected label: %s", links[1].Label)
	}
}

func TestExtractLinksEmpty(t *testing.T) {
	links := extractLinks("no links here")
	if links != nil {
		t.Fatalf("expected nil links, got %v", links)
	}
}

func TestLabelForURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/elastic/kibana/issues/42", "elastic/kibana#42"},
		{"https://github.com/elastic/kibana/pull/99", "elastic/kibana#99 (PR)"},
		{"https://github.com/elastic/kibana", "https://github.com/elastic/kibana"},
	}
	for _, tt := range tests {
		got := labelForURL(tt.url)
		if got != tt.want {
			t.Errorf("labelForURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
