package pipeline

import (
	"strings"
	"testing"
)

func TestBuildReviewPrompt_Default(t *testing.T) {
	branches := map[string]string{
		"fast": "quick answer here",
		"slow": "detailed answer here",
	}
	prompt := buildReviewPrompt(nil, branches, "")
	if !strings.Contains(prompt, "FAST") {
		t.Error("prompt should contain branch name 'FAST'")
	}
	if !strings.Contains(prompt, "SLOW") {
		t.Error("prompt should contain branch name 'SLOW'")
	}
	if !strings.Contains(prompt, "VARIANT:") {
		t.Error("prompt should instruct VARIANT: output format")
	}
	if !strings.Contains(prompt, "WINNER:") {
		t.Error("prompt should instruct WINNER: output format")
	}
	if !strings.Contains(prompt, "coherence") {
		t.Error("default criteria should include coherence")
	}
}

func TestBuildReviewPrompt_Criteria(t *testing.T) {
	branches := map[string]string{
		"a": "output a",
		"b": "output b",
	}
	cfg := &ReviewConfig{Criteria: []string{"accuracy", "completeness"}}
	prompt := buildReviewPrompt(cfg, branches, "")
	if !strings.Contains(prompt, "accuracy") {
		t.Error("prompt should contain criterion 'accuracy'")
	}
	if !strings.Contains(prompt, "completeness") {
		t.Error("prompt should contain criterion 'completeness'")
	}
	if !strings.Contains(prompt, "/10") {
		t.Error("prompt should use /10 scoring")
	}
}

func TestBuildReviewPrompt_CustomPrompt(t *testing.T) {
	branches := map[string]string{
		"formal": "Dear sir",
		"casual": "Hey dude",
	}
	cfg := &ReviewConfig{Prompt: "Which matches brand voice?"}
	prompt := buildReviewPrompt(cfg, branches, "")
	if !strings.Contains(prompt, "Which matches brand voice?") {
		t.Error("prompt should contain custom text")
	}
	if !strings.Contains(prompt, "Dear sir") {
		t.Error("prompt should contain formal branch output")
	}
}

func TestBuildReviewPrompt_WithObjective(t *testing.T) {
	branches := map[string]string{
		"a": "output a",
		"b": "output b",
	}
	prompt := buildReviewPrompt(nil, branches, "find the most accurate model")
	if !strings.Contains(prompt, "find the most accurate model") {
		t.Error("prompt should contain objective")
	}
	if !strings.Contains(prompt, "achieves this objective") {
		t.Error("prompt should instruct scoring against objective")
	}
}

func TestBuildReviewPrompt_ObjectiveWithCustomPrompt(t *testing.T) {
	branches := map[string]string{
		"a": "output a",
		"b": "output b",
	}
	cfg := &ReviewConfig{Prompt: "Which is better?"}
	prompt := buildReviewPrompt(cfg, branches, "test objective")
	if !strings.Contains(prompt, "test objective") {
		t.Error("prompt should contain objective with custom prompt")
	}
}

func TestBuildReflectionPrompt(t *testing.T) {
	branches := map[string]string{
		"local": "local output",
		"cloud": "cloud output",
	}
	prompt := buildReflectionPrompt("find best JSON producer", "VARIANT: local\ntotal: 18/30\nVARIANT: cloud\ntotal: 27/30\nWINNER: cloud", branches, "cloud")
	if !strings.Contains(prompt, "find best JSON producer") {
		t.Error("reflection prompt should contain objective")
	}
	if !strings.Contains(prompt, "cloud") {
		t.Error("reflection prompt should contain winner")
	}
	if !strings.Contains(prompt, "FINDING:") {
		t.Error("reflection prompt should request FINDING format")
	}
	if !strings.Contains(prompt, "CONFIDENCE:") {
		t.Error("reflection prompt should request CONFIDENCE format")
	}
}

func TestBuildReflectionPrompt_TruncatesLongOutput(t *testing.T) {
	longOutput := strings.Repeat("x", 600)
	branches := map[string]string{
		"a": longOutput,
		"b": "short",
	}
	prompt := buildReflectionPrompt("test", "scores", branches, "a")
	if strings.Contains(prompt, longOutput) {
		t.Error("long branch output should be truncated")
	}
	if !strings.Contains(prompt, "...") {
		t.Error("truncated output should end with ...")
	}
}

func TestParseReflection(t *testing.T) {
	output := `FINDING: Cloud model produces valid JSON consistently.
MODEL_INSIGHT: Local model struggles with structured output.
CONFIDENCE: high
RECOMMENDATION: Use cloud for JSON tasks, local for prose.`

	r := ParseReflection(output)
	if r.Finding != "Cloud model produces valid JSON consistently." {
		t.Errorf("Finding = %q", r.Finding)
	}
	if r.Confidence != "high" {
		t.Errorf("Confidence = %q", r.Confidence)
	}
	if r.Recommendation != "Use cloud for JSON tasks, local for prose." {
		t.Errorf("Recommendation = %q", r.Recommendation)
	}
	if r.ModelInsight["_raw"] != "Local model struggles with structured output." {
		t.Errorf("ModelInsight._raw = %q", r.ModelInsight["_raw"])
	}
}

func TestParseReflection_Empty(t *testing.T) {
	r := ParseReflection("")
	if r.Finding != "" {
		t.Errorf("expected empty Finding, got %q", r.Finding)
	}
	if r.Confidence != "" {
		t.Errorf("expected empty Confidence, got %q", r.Confidence)
	}
	if len(r.ModelInsight) != 0 {
		t.Errorf("expected empty ModelInsight, got %v", r.ModelInsight)
	}
}

func TestParseReflection_PartialOutput(t *testing.T) {
	output := `FINDING: Something was learned.
Some extra noise here.
CONFIDENCE: medium`

	r := ParseReflection(output)
	if r.Finding != "Something was learned." {
		t.Errorf("Finding = %q", r.Finding)
	}
	if r.Confidence != "medium" {
		t.Errorf("Confidence = %q", r.Confidence)
	}
	if r.Recommendation != "" {
		t.Errorf("expected empty Recommendation, got %q", r.Recommendation)
	}
}
