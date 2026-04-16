package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveWorkflowPath_Workspace verifies workflow resolution looks in the
// active workspace's workflows/ directory first.
func TestResolveWorkflowPath_Workspace(t *testing.T) {
	wsDir := t.TempDir()
	wfDir := filepath.Join(wsDir, "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	wfFile := filepath.Join(wfDir, "hello.glitch")
	if err := os.WriteFile(wfFile, []byte(`(workflow "hello" (step "s" (run "echo hi")))`+"\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	prev := workspacePath
	prevPath := runPathFlag
	workspacePath = wsDir
	runPathFlag = ""
	defer func() {
		workspacePath = prev
		runPathFlag = prevPath
	}()

	got, err := resolveWorkflowPath("hello")
	if err != nil {
		t.Fatalf("resolveWorkflowPath: %v", err)
	}
	if got != wfFile {
		t.Fatalf("resolveWorkflowPath: got %q, want %q", got, wfFile)
	}
}

// TestResolveWorkflowPath_PathFlag honours --path when given.
func TestResolveWorkflowPath_PathFlag(t *testing.T) {
	dir := t.TempDir()
	wfFile := filepath.Join(dir, "explicit.glitch")
	if err := os.WriteFile(wfFile, []byte(`(workflow "explicit" (step "s" (run "echo hi")))`+"\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	prev := runPathFlag
	runPathFlag = wfFile
	defer func() { runPathFlag = prev }()

	got, err := resolveWorkflowPath("ignored-name")
	if err != nil {
		t.Fatalf("resolveWorkflowPath: %v", err)
	}
	if got != wfFile {
		t.Fatalf("resolveWorkflowPath: got %q, want %q", got, wfFile)
	}
}

// TestResolveWorkflowPath_NotFound errors when the workflow is missing in
// every candidate directory.
func TestResolveWorkflowPath_NotFound(t *testing.T) {
	// Point HOME at an empty dir so the global fallback resolves to nothing.
	home := t.TempDir()
	t.Setenv("HOME", home)

	prev := workspacePath
	prevPath := runPathFlag
	workspacePath = ""
	runPathFlag = ""
	defer func() {
		workspacePath = prev
		runPathFlag = prevPath
	}()

	if _, err := resolveWorkflowPath("nope-not-here"); err == nil {
		t.Fatal("expected error for missing workflow, got nil")
	}
}

// TestRunCmd_MissingWorkflow verifies `glitch run <missing>` returns a clear
// error rather than panicking.
func TestRunCmd_MissingWorkflow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GLITCH_WORKSPACE", "")

	// Reset module-level state so we don't leak from prior tests.
	prevWS := workspacePath
	prevMerged := mergedConfig
	prevRunPath := runPathFlag
	workspacePath = ""
	mergedConfig = nil
	runPathFlag = ""
	defer func() {
		workspacePath = prevWS
		mergedConfig = prevMerged
		runPathFlag = prevRunPath
	}()

	rootCmd.SetArgs([]string{"run", "definitely-not-a-workflow"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing workflow, got nil")
	}
}

func TestRunCmd_HelpFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Scaffold a workflow with declared args.
	wfDir := filepath.Join(home, ".config", "glitch", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	src := `(arg "topic" :required true :description "Topic" :example "batch")

(workflow "demo" :description "demo workflow"
  (step "s" (run "echo ~param.topic")))
`
	if err := os.WriteFile(filepath.Join(wfDir, "demo.glitch"), []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"run", "demo", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"demo - demo workflow", "topic", "(required)", "--set topic"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing %q:\n%s", want, out)
		}
	}
}
