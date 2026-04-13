package research

import (
	"math"
	"testing"
)

func TestEvidenceCoverage(t *testing.T) {
	c := Critique{
		Claims: []CritiqueClaim{
			{Text: "claim1", Label: LabelGrounded},
			{Text: "claim2", Label: LabelPartial},
			{Text: "claim3", Label: LabelUngrounded},
		},
	}
	got := EvidenceCoverage(c)
	want := (1.0 + 0.5 + 0.0) / 3.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("EvidenceCoverage = %f, want %f", got, want)
	}
}

func TestEvidenceCoverageEmpty(t *testing.T) {
	got := EvidenceCoverage(Critique{})
	if got != 0 {
		t.Errorf("EvidenceCoverage(empty) = %f, want 0", got)
	}
}

func TestCrossCapabilityAgree(t *testing.T) {
	tests := []struct {
		name    string
		sources []string
		want    float64
	}{
		{"two sources", []string{"a", "b"}, 1.0},
		{"one source", []string{"a"}, 0.4},
		{"no sources", nil, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bundle EvidenceBundle
			for _, s := range tt.sources {
				bundle.Add(Evidence{Source: s, Body: "x"})
			}
			got := CrossCapabilityAgree(bundle)
			if got != tt.want {
				t.Errorf("CrossCapabilityAgree = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestParseCritique(t *testing.T) {
	raw := `Here are the claims: [{"text": "Go is fast", "label": "grounded"}, {"text": "Go has no GC", "label": "ungrounded"}]`
	crit, err := ParseCritique(raw)
	if err != nil {
		t.Fatalf("ParseCritique: %v", err)
	}
	if len(crit.Claims) != 2 {
		t.Fatalf("expected 2 claims, got %d", len(crit.Claims))
	}
	if crit.Claims[0].Label != LabelGrounded {
		t.Errorf("claim 0: got %q, want grounded", crit.Claims[0].Label)
	}
	if crit.Claims[1].Label != LabelUngrounded {
		t.Errorf("claim 1: got %q, want ungrounded", crit.Claims[1].Label)
	}
}

func TestParseCritiqueNormalisesUnknown(t *testing.T) {
	raw := `[{"text": "x", "label": "bogus"}]`
	crit, err := ParseCritique(raw)
	if err != nil {
		t.Fatalf("ParseCritique: %v", err)
	}
	if crit.Claims[0].Label != LabelUngrounded {
		t.Errorf("expected unknown label normalised to ungrounded, got %q", crit.Claims[0].Label)
	}
}

func TestParseJudgeScore(t *testing.T) {
	tests := []struct {
		raw     string
		want    float64
		wantErr bool
	}{
		{"0.85", 0.85, false},
		{"The score is 0.72 out of 1.0", 0.72, false},
		{"no number", 0, true},
		{"5.0", 1.0, false}, // clamped
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := ParseJudgeScore(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("ParseJudgeScore(%q) = %f, want %f", tt.raw, got, tt.want)
			}
		})
	}
}

func TestComposite(t *testing.T) {
	s := Score{
		EvidenceCoverage:     Ptr(0.8),
		CrossCapabilityAgree: Ptr(1.0),
		JudgeScore:           Ptr(0.9),
	}
	got := Composite(s)
	want := (0.8 + 1.0 + 0.9) / 3.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Composite = %f, want %f", got, want)
	}
}

func TestCompositePartial(t *testing.T) {
	s := Score{
		CrossCapabilityAgree: Ptr(1.0),
	}
	got := Composite(s)
	if got != 1.0 {
		t.Errorf("Composite with only CCA = %f, want 1.0", got)
	}
}
