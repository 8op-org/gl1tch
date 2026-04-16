package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceFlag(t *testing.T) {
	// Reset for test
	workspacePath = ""

	// Create a real workspace dir so Resolve() succeeds and preserves the path.
	wsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(wsDir, "workspace.glitch"), []byte("(workspace \"flag-ws\")\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"--workspace", wsDir, "config", "show"})
	if err := rootCmd.Execute(); err != nil {
		// config show may fail, that's ok — we just check the flag was parsed
	}

	if workspacePath != wsDir {
		t.Fatalf("workspacePath: got %q, want %q", workspacePath, wsDir)
	}
}

func TestLoadWorkflows_Workspace(t *testing.T) {
	wsDir := t.TempDir()
	wfDir := filepath.Join(wsDir, "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "test-wf.yaml"), []byte("name: test-wf\ndescription: test workflow\nsteps: []\n"), 0o644)

	workspacePath = wsDir
	defer func() { workspacePath = "" }()

	workflows, err := loadWorkflows()
	if err != nil {
		t.Fatalf("loadWorkflows: %v", err)
	}
	if _, ok := workflows["test-wf"]; !ok {
		t.Fatal("expected test-wf workflow from workspace")
	}
}

func TestLoadWorkflows_ConfigOverride(t *testing.T) {
	customDir := t.TempDir()
	os.WriteFile(filepath.Join(customDir, "custom-wf.yaml"), []byte("name: custom-wf\ndescription: custom\nsteps: []\n"), 0o644)

	wsDir := t.TempDir()
	os.MkdirAll(filepath.Join(wsDir, "workflows"), 0o755)
	workspacePath = wsDir
	defer func() { workspacePath = "" }()

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.glitch")
	os.WriteFile(cfgPath, []byte(`(config :default-model "qwen3:8b" :workflows-dir "`+customDir+`")`+"\n"), 0o644)

	cfg, _ := loadConfigFrom(cfgPath)
	wfDir := resolveWorkflowsDir(cfg)
	if wfDir != customDir {
		t.Fatalf("resolveWorkflowsDir: got %q, want %q", wfDir, customDir)
	}
}
