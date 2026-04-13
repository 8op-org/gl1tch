package research

import "testing"

func TestParseFeedback(t *testing.T) {
	raw := `Here are the fixed files.

--- FEEDBACK ---
- evidence_quality: good
- missing: ["directory listing for docs/teams/ci/macos/", "sibling file orka.md"]
- useful: ["repo-scan found all TBC placeholders"]
- suggestion: "for doc fixes, include sibling files in the same directory"
`
	fb, err := ParseFeedback(raw)
	if err != nil {
		t.Fatalf("ParseFeedback: %v", err)
	}
	if fb.Quality != "good" {
		t.Fatalf("quality: got %q, want %q", fb.Quality, "good")
	}
	if len(fb.Missing) != 2 {
		t.Fatalf("missing: got %d, want 2", len(fb.Missing))
	}
	if fb.Suggestion == "" {
		t.Fatal("expected non-empty suggestion")
	}
}

func TestParseFeedbackNoSection(t *testing.T) {
	raw := "Just a plain answer with no feedback section."
	fb, err := ParseFeedback(raw)
	if err != nil {
		t.Fatalf("ParseFeedback: %v", err)
	}
	if fb.Quality != "" {
		t.Fatalf("expected empty quality, got %q", fb.Quality)
	}
}

func TestSplitDraftAndFeedback(t *testing.T) {
	raw := "The fixes are:\n\nblah blah\n\n--- FEEDBACK ---\n- evidence_quality: adequate\n- suggestion: test"
	draft, fb := SplitDraftAndFeedback(raw)
	if draft != "The fixes are:\n\nblah blah" {
		t.Fatalf("draft: got %q", draft)
	}
	if fb == "" {
		t.Fatal("expected non-empty feedback section")
	}
}
