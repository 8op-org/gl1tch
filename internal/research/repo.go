package research

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var reRepoRef = regexp.MustCompile(`(?:([a-zA-Z0-9_.-]+)/)?([a-zA-Z0-9_.-]+)(?:#\d+)?`)

// RepoDir returns the standard clone path for a repo.
func RepoDir(org, repo string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "glitch", "repos", org, repo)
}

// EnsureRepo ensures a repo is available locally. Checks:
// 1. Explicit localPath (if provided and exists, use it)
// 2. ~/Projects/<repo>
// 3. ~/.local/share/glitch/repos/<org>/<repo> (clone if missing)
func EnsureRepo(org, repo, localPath string) (string, error) {
	if localPath != "" {
		if info, err := os.Stat(localPath); err == nil && info.IsDir() {
			return localPath, nil
		}
	}
	if home, err := os.UserHomeDir(); err == nil && repo != "" {
		candidate := filepath.Join(home, "Projects", repo)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	if org == "" || repo == "" {
		return "", fmt.Errorf("repo: cannot clone without org/repo")
	}
	dir := RepoDir(org, repo)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		cmd := exec.Command("git", "-C", dir, "pull", "--ff-only", "-q")
		cmd.Run()
		return dir, nil
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return "", fmt.Errorf("repo: mkdir: %w", err)
	}
	remote := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)
	cmd := exec.Command("git", "clone", "--depth=1", remote, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("repo: clone %s: %w\n%s", remote, err, out)
	}
	return dir, nil
}

// ParseRepoFromQuestion extracts org/repo from a question string.
func ParseRepoFromQuestion(question string) (org, repo string) {
	skip := map[string]bool{
		"the": true, "this": true, "that": true, "what": true, "fix": true,
		"check": true, "show": true, "find": true, "how": true, "why": true,
		"are": true, "there": true, "does": true, "have": true, "been": true,
		"code": true,
	}
	for _, word := range strings.Fields(question) {
		word = strings.Trim(word, "?.,!\"'")
		m := reRepoRef.FindStringSubmatch(word)
		if m == nil {
			continue
		}
		r := m[2]
		if skip[strings.ToLower(r)] || len(r) < 3 {
			continue
		}
		o := m[1]
		if o == "" {
			o = "elastic"
		}
		return o, r
	}
	return "", ""
}
