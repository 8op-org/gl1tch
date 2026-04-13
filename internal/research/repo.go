package research

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// reGitHubURL matches https://github.com/org/repo (with optional path)
var reGitHubURL = regexp.MustCompile(`https?://github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)`)

// reExplicitRepo matches org/repo (requires the slash)
var reExplicitRepo = regexp.MustCompile(`([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)`)

// reIssueRepo matches repo#number (requires the hash+number)
var reIssueRepo = regexp.MustCompile(`([a-zA-Z0-9_.-]+)#(\d+)`)

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
// Prioritizes explicit org/repo patterns, then repo#number patterns.
// Single bare words are never matched — requires a slash or hash.
func ParseRepoFromQuestion(question string) (org, repo string) {
	// Pass 0: GitHub URL (e.g., "https://github.com/elastic/ensemble/issues/872")
	if m := reGitHubURL.FindStringSubmatch(question); m != nil {
		return m[1], m[2]
	}
	// Pass 1: look for explicit org/repo (e.g., "elastic/observability-robots")
	for _, word := range strings.Fields(question) {
		word = strings.Trim(word, "?.,!\"'")
		if m := reExplicitRepo.FindStringSubmatch(word); m != nil {
			return m[1], m[2]
		}
	}
	// Pass 2: look for repo#number (e.g., "observability-robots#3928")
	for _, word := range strings.Fields(question) {
		word = strings.Trim(word, "?.,!\"'")
		if m := reIssueRepo.FindStringSubmatch(word); m != nil {
			return "elastic", m[1] // default org
		}
	}
	return "", ""
}
