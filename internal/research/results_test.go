package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestSaveResults(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "results")
	result := Result{
		Draft: "--- FILE: docs/test.md ---\n# Test\n--- END FILE ---",
		Feedback: Feedback{
			Quality:    "good",
			Suggestion: "test suggestion",
		},
	}
	err := SaveResults(dir, result)
	if err != nil {
		t.Fatalf("SaveResults: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "drafts.md")); err != nil {
		t.Fatal("drafts.md not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "test.md")); err != nil {
		t.Fatal("extracted file not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "feedback.md")); err != nil {
		t.Fatal("feedback.md not created")
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
