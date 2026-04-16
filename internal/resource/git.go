package resource

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// materializeGit clones or refreshes a git resource under ws/resources/<name>
// and returns the resolved commit SHA and the clone directory.
func materializeGit(ws string, r Resource, force bool) (string, string, error) {
	dir := filepath.Join(ws, "resources", r.Name)
	_, err := os.Stat(filepath.Join(dir, ".git"))
	switch {
	case err == nil && force:
		_ = os.RemoveAll(dir)
		fallthrough
	case os.IsNotExist(err) || force:
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return "", "", err
		}
		out, err := exec.Command("git", "clone", r.URL, dir).CombinedOutput()
		if err != nil {
			return "", "", fmt.Errorf("git clone %s: %v: %s", r.URL, err, out)
		}
	default:
		if out, err := exec.Command("git", "-C", dir, "fetch", "--tags", "origin").CombinedOutput(); err != nil {
			return "", "", fmt.Errorf("git fetch: %v: %s", err, out)
		}
	}
	if r.Ref != "" {
		if out, err := exec.Command("git", "-C", dir, "checkout", "-q", r.Ref).CombinedOutput(); err != nil {
			return "", "", fmt.Errorf("git checkout %s: %v: %s", r.Ref, err, out)
		}
	}
	shaOut, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", "", fmt.Errorf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(shaOut)), dir, nil
}
