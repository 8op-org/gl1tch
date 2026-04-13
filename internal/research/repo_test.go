package research

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepoDir(t *testing.T) {
	dir := RepoDir("elastic", "observability-robots")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "share", "glitch", "repos", "elastic", "observability-robots")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
}

func TestEnsureRepoLocal(t *testing.T) {
	cwd, _ := os.Getwd()
	got, err := EnsureRepo("", "", cwd)
	if err != nil {
		t.Fatalf("EnsureRepo with existing path: %v", err)
	}
	if got != cwd {
		t.Fatalf("got %q, want %q", got, cwd)
	}
}

func TestParseRepoFromQuestion(t *testing.T) {
	cases := []struct {
		input    string
		wantOrg  string
		wantRepo string
	}{
		{"fix observability-robots#3928", "elastic", "observability-robots"},
		{"what's in elastic/ensemble", "elastic", "ensemble"},
		{"check the code", "", ""},
	}
	for _, tc := range cases {
		org, repo := ParseRepoFromQuestion(tc.input)
		if org != tc.wantOrg || repo != tc.wantRepo {
			t.Errorf("ParseRepoFromQuestion(%q) = %q/%q, want %q/%q", tc.input, org, repo, tc.wantOrg, tc.wantRepo)
		}
	}
}
