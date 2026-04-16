package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseReview(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.md")
	os.WriteFile(path, []byte(
		"1. Specificity — PASS — good\n2. Completeness — FAIL — missing\n3. Feasibility — PASS — ok\nOVERALL: FAIL\n",
	), 0o644)

	score, err := ParseReview(path)
	if err != nil {
		t.Fatal(err)
	}
	if score.Passed != 2 {
		t.Fatalf("expected 2 passed, got %d", score.Passed)
	}
	if score.Total != 3 {
		t.Fatalf("expected 3 total, got %d", score.Total)
	}
	if score.Pass {
		t.Fatal("expected overall FAIL")
	}
}

func TestGenerateManifest_PicksHighestConfidence(t *testing.T) {
	dir := t.TempDir()
	issueDir := filepath.Join(dir, "3642")

	// New layout: issueDir/children/<variant>-<iter>-<runid>/
	localDir := filepath.Join(issueDir, "children", "local-1-101")
	os.MkdirAll(localDir, 0o755)
	os.WriteFile(filepath.Join(localDir, "review.md"), []byte(
		"1. Specificity — PASS — good\n2. Completeness — FAIL — missing\n3. Feasibility — PASS — ok\n4. Testing — PASS — yes\n5. PR Quality — FAIL — weak\nOVERALL: FAIL\n",
	), 0o644)
	os.WriteFile(filepath.Join(localDir, "pr-title.txt"), []byte("Add telemetry infra"), 0o644)
	os.WriteFile(filepath.Join(localDir, "pr-body.md"), []byte("## Summary\nAdds telemetry"), 0o644)
	os.WriteFile(filepath.Join(localDir, "plan.md"), []byte("# Plan\nDo the thing"), 0o644)

	claudeDir := filepath.Join(issueDir, "children", "claude-1-102")
	os.MkdirAll(claudeDir, 0o755)
	os.WriteFile(filepath.Join(claudeDir, "review.md"), []byte(
		"1. Specificity — PASS — good\n2. Completeness — PASS — all covered\n3. Feasibility — PASS — ok\n4. Testing — PASS — yes\n5. PR Quality — PASS — strong\nOVERALL: PASS\n",
	), 0o644)
	os.WriteFile(filepath.Join(claudeDir, "pr-title.txt"), []byte("Implement telemetry"), 0o644)
	os.WriteFile(filepath.Join(claudeDir, "pr-body.md"), []byte("## Summary\nFull telemetry"), 0o644)
	os.WriteFile(filepath.Join(claudeDir, "plan.md"), []byte("# Plan\nComplete plan"), 0o644)

	manifest, err := GenerateManifest(issueDir, "3642", []string{"local", "claude"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.BestVariant != "claude" {
		t.Fatalf("expected winner claude, got %q", manifest.BestVariant)
	}
	if manifest.BestIteration != 1 {
		t.Fatalf("expected iteration 1, got %d", manifest.BestIteration)
	}
	if manifest.BestScore != 5 {
		t.Fatalf("expected score 5, got %d", manifest.BestScore)
	}
}

func TestWriteManifest(t *testing.T) {
	dir := t.TempDir()
	issueDir := filepath.Join(dir, "3642")
	os.MkdirAll(issueDir, 0o755)

	m := &Manifest{
		Issue:         "3642",
		BestVariant:   "claude",
		BestIteration: 1,
		BestScore:     5,
		BestTotal:     5,
		Scores: []IterationScores{
			{Iteration: 1, Variants: map[string]Score{
				"local":  {Passed: 3, Total: 5, Pass: false},
				"claude": {Passed: 5, Total: 5, Pass: true},
			}},
		},
	}

	err := WriteManifest(issueDir, m, []string{"local", "claude"}, 1)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(issueDir, "manifest.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !contains(content, "claude") {
		t.Fatal("manifest should mention claude as winner")
	}
	if !contains(content, "3/5") {
		t.Fatal("manifest should show local score 3/5")
	}
	if !contains(content, "5/5") {
		t.Fatal("manifest should show claude score 5/5")
	}
}

func TestResultPath_Convention(t *testing.T) {
	dir := t.TempDir()
	got := resultPath(filepath.Join(dir, "3920"), "claude", 1, 42)
	want := filepath.Join(dir, "3920", "children", "claude-1-42")
	if got != want {
		t.Fatalf("resultPath: got %q, want %q", got, want)
	}
}

func TestResultPath_ZeroRunID(t *testing.T) {
	dir := t.TempDir()
	got := resultPath(filepath.Join(dir, "100"), "local", 2, 0)
	want := filepath.Join(dir, "100", "children", "local-2-0")
	if got != want {
		t.Fatalf("resultPath: got %q, want %q", got, want)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
