package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func writeStubWorkflow(t *testing.T, path, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := []byte("(workflow \"" + name + "\"\n  (step \"main\"\n    (run \"echo hi\")))\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestLoadDir_SkipsHiddenDirs guards that dot-prefixed subdirectories are
// excluded from discovery, matching the common Unix convention. Issue #5.
func TestLoadDir_SkipsHiddenDirs(t *testing.T) {
	root := t.TempDir()

	writeStubWorkflow(t, filepath.Join(root, "visible.glitch"), "visible")
	writeStubWorkflow(t, filepath.Join(root, "nested", "also-visible.glitch"), "also-visible")
	writeStubWorkflow(t, filepath.Join(root, ".archive", "hidden.glitch"), "hidden")
	writeStubWorkflow(t, filepath.Join(root, ".archive", "batch-01", "deep-hidden.glitch"), "deep-hidden")

	got, err := LoadDir(root)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	want := map[string]bool{"visible": true, "also-visible": true}
	for name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("missing visible workflow %q in %v", name, keysOf(got))
		}
	}
	for _, unwanted := range []string{"hidden", "deep-hidden"} {
		if _, ok := got[unwanted]; ok {
			t.Errorf("workflow %q from hidden dir should have been skipped (got %v)", unwanted, keysOf(got))
		}
	}
}

// TestLoadDir_RootNameStartingWithDotNotSkipped guards the root directory
// itself is always walked, even when the caller happens to pass a path whose
// final segment begins with "." (e.g. ~/.config/glitch/workflows would be
// fine, but defensive check for .workflows).
func TestLoadDir_RootNameStartingWithDotNotSkipped(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, ".workflows")
	writeStubWorkflow(t, filepath.Join(root, "w.glitch"), "w")

	got, err := LoadDir(root)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if _, ok := got["w"]; !ok {
		t.Errorf("root dir with leading dot should still be walked; got %v", keysOf(got))
	}
}

func keysOf(m map[string]*Workflow) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
