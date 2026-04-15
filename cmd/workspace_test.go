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

	// Check result landed in workspace
	resultDir := filepath.Join(wsDir, "results", "elastic", "ensemble", "issue-99")
	if _, err := os.Stat(filepath.Join(resultDir, "README.md")); err != nil {
		t.Fatal("README.md not in workspace results")
	}
	if _, err := os.Stat(filepath.Join(resultDir, "run.json")); err != nil {
		t.Fatal("run.json not in workspace results")
	}
}
