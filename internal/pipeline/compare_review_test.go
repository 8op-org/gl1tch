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
