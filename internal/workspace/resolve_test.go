package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWorkspace(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "workspace.glitch"),
		[]byte("(workspace \""+name+"\")\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveEnvOverride(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "foo")
	_ = os.MkdirAll(ws, 0o755)
	writeWorkspace(t, ws, "foo-ws")

	got := Resolve(ResolveOpts{EnvPath: ws})
	if got.Name != "foo-ws" || got.Path != ws {
		t.Fatalf("env override failed: %+v", got)
	}
}

func TestResolveExplicitBeatsEnv(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	_ = os.MkdirAll(a, 0o755)
	_ = os.MkdirAll(b, 0o755)
	writeWorkspace(t, a, "ws-a")
	writeWorkspace(t, b, "ws-b")

	got := Resolve(ResolveOpts{ExplicitPath: a, EnvPath: b})
	if got.Name != "ws-a" {
		t.Fatalf("explicit should win: %+v", got)
	}
}

func TestResolveWalkUp(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(parent, "sub", "deep")
	_ = os.MkdirAll(child, 0o755)
	writeWorkspace(t, parent, "walked")

	got := Resolve(ResolveOpts{StartDir: child})
	if got.Name != "walked" || got.Path != parent {
		t.Fatalf("walk-up failed: %+v", got)
	}
}

func TestResolveNoneReturnsEmpty(t *testing.T) {
	got := Resolve(ResolveOpts{StartDir: t.TempDir()})
	if got.Name != "" || got.Path != "" {
		t.Fatalf("expected empty Resolved, got %+v", got)
	}
}
