package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractFiles(t *testing.T) {
	draft := "Here are the fixes.\n\n--- FILE: docs/teams/ci/macos/index.md ---\n# macOS Runners\n\nSome content here.\n--- END FILE ---\n\n--- FILE: docs/teams/ci/dependencies/updatecli.md ---\n# Updatecli\n\nMore content.\n--- END FILE ---\n"
	files := ExtractFiles(draft)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "docs/teams/ci/macos/index.md" {
		t.Fatalf("path: got %q", files[0].Path)
	}
	if !strings.Contains(files[0].Content, "macOS Runners") {
		t.Fatal("expected content")
	}
}

func TestIsSubstantive(t *testing.T) {
	if IsSubstantive("short answer") {
		t.Fatal("short answer should not be substantive")
	}
	if !IsSubstantive(strings.Repeat("x", 501)) {
		t.Fatal("long answer should be substantive")
	}
	if !IsSubstantive("--- FILE: foo.md ---\ncontent\n--- END FILE ---") {
		t.Fatal("answer with files should be substantive")
	}
}

func TestSaveLoopResult(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-run-001",
		Document: ResearchDocument{
			Source:    "github-issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/872",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "872"},
		},
		Goal:        GoalSummarize,
		Output:      "# Summary\n\nThis is a test summary.",
		ToolCalls: []ToolResult{
			{Tool: "grep_code", Output: "main.go:10: func main()"},
			{Tool: "read_file", Output: "package main\n\nimport \"fmt\""},
		},
		LLMCalls:    3,
		TokensIn:    1500,
		TokensOut:   800,
		CostUSD:     0.005,
		MaxTier:     1,
		Escalations: 0,
		Duration:    2 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	dir := filepath.Join(base, "elastic", "ensemble", "issue-872")

	// Verify summary.md exists
	if _, err := os.Stat(filepath.Join(dir, "summary.md")); err != nil {
		t.Fatal("summary.md not created")
	}

	// Verify run.json exists and has correct content
	runData, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal("run.json not created")
	}
	var meta runJSON
	if err := json.Unmarshal(runData, &meta); err != nil {
		t.Fatalf("run.json unmarshal: %v", err)
	}
	if meta.RunID != "test-run-001" {
		t.Fatalf("run_id: got %q", meta.RunID)
	}
	if meta.ToolCalls != 2 {
		t.Fatalf("tool_calls: got %d, want 2", meta.ToolCalls)
	}

	// Verify evidence files
	if _, err := os.Stat(filepath.Join(dir, "evidence", "001-grep_code.txt")); err != nil {
		t.Fatal("evidence/001-grep_code.txt not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "evidence", "002-read_file.txt")); err != nil {
		t.Fatal("evidence/002-read_file.txt not created")
	}
}

func TestRunJSON_StandardFields(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-std-001",
		Document: ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/50",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "50"},
		},
		Goal:     GoalSummarize,
		Output:   "summary text here, long enough to be substantive" + strings.Repeat(" content", 100),
		LLMCalls: 2,
		Duration: 5 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	dir := filepath.Join(base, "elastic", "ensemble", "issue-50")
	data, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatalf("read run.json: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if raw["repo"] != "elastic/ensemble" {
		t.Fatalf("repo: got %v", raw["repo"])
	}
	if raw["ref_type"] != "issue" {
		t.Fatalf("ref_type: got %v", raw["ref_type"])
	}
	if raw["ref_number"] != float64(50) {
		t.Fatalf("ref_number: got %v", raw["ref_number"])
	}
}

func TestSaveLoopResult_WritesReadme(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-readme-001",
		Document: ResearchDocument{
			Source:    "github_issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/42",
			Title:     "Fix flaky CI test",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "42"},
		},
		Goal:   GoalSummarize,
		Output: "# Summary\n\nThe CI test is flaky because of a race condition.\n\n## Recommendation\n\nAdd a mutex around the shared state.\n\n## Response Draft\n\nI investigated the flaky CI test and found a race condition in the shared state handler.",
		ToolCalls: []ToolResult{
			{Tool: "grep_code", Output: "found race"},
		},
		LLMCalls: 2,
		Duration: 3 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	dir := filepath.Join(base, "elastic", "ensemble", "issue-42")
	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal("README.md not created")
	}

	content := string(readme)
	if !strings.Contains(content, "repo: elastic/ensemble") {
		t.Fatal("README.md missing repo frontmatter")
	}
	if !strings.Contains(content, "ref: issue-42") {
		t.Fatal("README.md missing ref frontmatter")
	}
	if !strings.Contains(content, "title: \"Fix flaky CI test\"") {
		t.Fatal("README.md missing title frontmatter")
	}
	if !strings.Contains(content, "race condition") {
		t.Fatal("README.md missing output content")
	}
	if !strings.Contains(content, "001-grep_code.txt") {
		t.Fatal("README.md missing evidence index")
	}
}

func TestResultDir_IssuePrefix(t *testing.T) {
	result := LoopResult{
		Document: ResearchDocument{
			Source:   "github_issue",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "872"},
		},
	}
	got := resultDir("/base", result)
	want := "/base/elastic/ensemble/issue-872"
	if got != want {
		t.Fatalf("resultDir: got %q, want %q", got, want)
	}
}

func TestResultDir_PRPrefix(t *testing.T) {
	result := LoopResult{
		Document: ResearchDocument{
			Source:   "github_pr",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "100"},
		},
	}
	got := resultDir("/base", result)
	want := "/base/elastic/ensemble/pr-100"
	if got != want {
		t.Fatalf("resultDir: got %q, want %q", got, want)
	}
}

func TestResultDir_NoSourceFallback(t *testing.T) {
	result := LoopResult{
		Document: ResearchDocument{
			Source:   "text",
			Repo:     "elastic/ensemble",
			Metadata: map[string]string{"number": "50"},
		},
	}
	got := resultDir("/base", result)
	want := "/base/elastic/ensemble/issue-50"
	if got != want {
		t.Fatalf("resultDir: got %q, want %q", got, want)
	}
}

func TestSaveLoopResultImplement(t *testing.T) {
	base := filepath.Join(t.TempDir(), "results")

	result := LoopResult{
		RunID: "test-run-002",
		Document: ResearchDocument{
			Source:    "github-issue",
			SourceURL: "https://github.com/elastic/ensemble/issues/100",
			Repo:      "elastic/ensemble",
			Metadata:  map[string]string{"number": "100"},
		},
		Goal:   GoalImplement,
		Output: "# Implementation Plan\n\n1. Update config.go\n2. Add tests",
		ToolCalls: []ToolResult{
			{Tool: "read_file", Output: "some file content"},
		},
		LLMCalls: 2,
		Duration: 1 * time.Second,
	}

	if err := SaveLoopResult(base, result); err != nil {
		t.Fatalf("SaveLoopResult: %v", err)
	}

	dir := filepath.Join(base, "elastic", "ensemble", "issue-100")

	// Verify implementation/plan.md exists
	if _, err := os.Stat(filepath.Join(dir, "implementation", "plan.md")); err != nil {
		t.Fatal("implementation/plan.md not created")
	}

	// Verify summary.md does NOT exist
	if _, err := os.Stat(filepath.Join(dir, "summary.md")); !os.IsNotExist(err) {
		t.Fatal("summary.md should not exist for goal=implement")
	}
}
