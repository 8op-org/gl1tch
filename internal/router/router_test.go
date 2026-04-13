package router

import (
	"strings"
	"testing"
)

func TestParseIssueRef_BareNumber(t *testing.T) {
	repo, issue, ok := ParseIssueRef("3442")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "" {
		t.Fatalf("expected empty repo, got %q", repo)
	}
	if issue != "3442" {
		t.Fatalf("expected 3442, got %q", issue)
	}
}

func TestParseIssueRef_ShortForm(t *testing.T) {
	repo, issue, ok := ParseIssueRef("observability-robots#3442")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "elastic/observability-robots" {
		t.Fatalf("expected elastic/observability-robots, got %q", repo)
	}
	if issue != "3442" {
		t.Fatalf("expected 3442, got %q", issue)
	}
}

func TestParseIssueRef_FullForm(t *testing.T) {
	repo, issue, ok := ParseIssueRef("elastic/ensemble#1281")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "elastic/ensemble" {
		t.Fatalf("expected elastic/ensemble, got %q", repo)
	}
	if issue != "1281" {
		t.Fatalf("expected 1281, got %q", issue)
	}
}

func TestParseIssueRef_Invalid(t *testing.T) {
	_, _, ok := ParseIssueRef("not an issue")
	if ok {
		t.Fatal("expected not ok")
	}
}

func TestMatchWorkOnIssue(t *testing.T) {
	tests := []struct {
		input     string
		wantMatch bool
		wantRef   string
	}{
		{"work on issue 3442", true, "3442"},
		{"work on issue observability-robots#3442", true, "observability-robots#3442"},
		{"work on issue elastic/ensemble#1281", true, "elastic/ensemble#1281"},
		{"what issues are open", false, ""},
		{"list prs", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, ok := MatchWorkOnIssue(tt.input)
			if ok != tt.wantMatch {
				t.Fatalf("match=%v, want %v", ok, tt.wantMatch)
			}
			if ok && ref != tt.wantRef {
				t.Fatalf("ref=%q, want %q", ref, tt.wantRef)
			}
		})
	}
}

func TestResolveRepo_FromGitRemote(t *testing.T) {
	repo, err := ResolveRepo("")
	if err != nil {
		t.Skipf("not in a git repo with github remote: %v", err)
	}
	if repo == "" {
		t.Fatal("expected non-empty repo")
	}
	if !strings.Contains(repo, "/") {
		t.Fatalf("expected owner/repo format, got %q", repo)
	}
}

func TestResolveRepo_WithExplicitRepo(t *testing.T) {
	repo, err := ResolveRepo("elastic/ensemble")
	if err != nil {
		t.Fatal(err)
	}
	if repo != "elastic/ensemble" {
		t.Fatalf("expected elastic/ensemble, got %q", repo)
	}
}
