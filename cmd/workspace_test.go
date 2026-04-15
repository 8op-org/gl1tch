package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/research"
)

func TestWorkspaceIntegration(t *testing.T) {
	wsDir := t.TempDir()

	// Create workspace structure
	wfDir := filepath.Join(wsDir, "workflows")
	os.MkdirAll(wfDir, 0o755)
	os.WriteFile(filepath.Join(wfDir, "test.yaml"), []byte("name: test\ndescription: test\nsteps: []\n"), 0o644)

	// Set workspace
	workspacePath = wsDir
	defer func() { workspacePath = "" }()

	// Verify workflow resolution
	wfs, err := loadWorkflows()
	if err != nil {
		t.Fatalf("loadWorkflows: %v", err)
	}
	if _, ok := wfs["test"]; !ok {
		t.Fatal("workspace workflow not found")
	}

	// Verify result path resolution
	rdir := resolveResultsDir()
	expected := filepath.Join(wsDir, "results")
	if rdir != expected {
		t.Fatalf("resolveResultsDir: got %q, want %q", rdir, expected)
	}

	// Verify SaveLoopResult writes to workspace
	result := research.LoopResult{
		RunID: "ws-test-001",
		Document: research.ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/99",
			Title:     "Test issue",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "99"},
		},
		Goal:   research.GoalSummarize,
		Output: "Test summary with enough content to be substantive." + strings.Repeat(" padding", 100),
	}

	if err := research.SaveLoopResult(rdir, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	// Check result landed in workspace (run-scoped dir with latest symlink)
	issueDir := filepath.Join(wsDir, "results", "elastic", "ensemble", "issue-99")
	latestDir, err := filepath.EvalSymlinks(filepath.Join(issueDir, "latest"))
	if err != nil {
		t.Fatalf("latest symlink: %v", err)
	}
	if _, err := os.Stat(filepath.Join(latestDir, "README.md")); err != nil {
		t.Fatal("README.md not in workspace results")
	}
	if _, err := os.Stat(filepath.Join(latestDir, "run.json")); err != nil {
		t.Fatal("run.json not in workspace results")
	}
}

func TestEnsureWorkspaceDir(t *testing.T) {
	wsDir := t.TempDir()

	if err := ensureWorkspaceDir(wsDir); err != nil {
		t.Fatalf("ensureWorkspaceDir: %v", err)
	}

	dotGlitch := filepath.Join(wsDir, ".glitch")
	info, err := os.Stat(dotGlitch)
	if err != nil {
		t.Fatalf(".glitch dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".glitch is not a directory")
	}

	gi, err := os.ReadFile(filepath.Join(dotGlitch, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	if string(gi) != "*\n" {
		t.Fatalf(".gitignore: got %q, want %q", string(gi), "*\n")
	}
}

func TestEnsureWorkspaceDir_Idempotent(t *testing.T) {
	wsDir := t.TempDir()

	if err := ensureWorkspaceDir(wsDir); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := ensureWorkspaceDir(wsDir); err != nil {
		t.Fatalf("second call: %v", err)
	}

	gi, _ := os.ReadFile(filepath.Join(wsDir, ".glitch", ".gitignore"))
	if string(gi) != "*\n" {
		t.Fatalf(".gitignore content wrong after second call: %q", string(gi))
	}
}
