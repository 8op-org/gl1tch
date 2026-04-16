package resource

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initUpstream creates a tiny local git repo with one commit so we can
// clone it without hitting the network.
func initUpstream(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	must := func(c *exec.Cmd) {
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	must(exec.Command("git", "init", "-q", "-b", "main"))
	must(exec.Command("git", "-c", "user.email=t@t", "-c", "user.name=t",
		"commit", "-q", "--allow-empty", "-m", "init"))
	return dir
}

func TestSyncGitResource(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	upstream := initUpstream(t)
	ws := t.TempDir()

	r := Resource{Name: "up", Kind: KindGit, URL: upstream, Ref: "main"}
	result, err := Sync(ws, r)
	if err != nil {
		t.Fatal(err)
	}
	if result.Pin == "" {
		t.Fatal("expected pin SHA populated")
	}
	if _, err := os.Stat(filepath.Join(ws, "resources", "up", ".git")); err != nil {
		t.Fatalf("clone not materialized: %v", err)
	}
}

func TestSyncLocalSymlink(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(t.TempDir(), "data")
	_ = os.MkdirAll(target, 0o755)

	r := Resource{Name: "data", Kind: KindLocal, Path: target}
	if _, err := Sync(ws, r); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ws, "resources", "data")
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("not a symlink: %v", err)
	}
	if got != target {
		t.Fatalf("symlink target mismatch: got %q want %q", got, target)
	}
}

func TestSyncTracker(t *testing.T) {
	ws := t.TempDir()
	r := Resource{Name: "bug", Kind: KindTracker, Repo: "org/repo"}
	res, err := Sync(ws, r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != KindTracker {
		t.Fatalf("expected tracker, got %s", res.Kind)
	}
	if _, err := os.Stat(filepath.Join(ws, "resources", "bug")); !os.IsNotExist(err) {
		t.Fatalf("tracker should not materialize on disk: err=%v", err)
	}
}
