package research

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CritiqueLabel represents the grounding status of a claim.
type CritiqueLabel string

const (
	LabelGrounded   CritiqueLabel = "grounded"
	LabelPartial    CritiqueLabel = "partial"
	LabelUngrounded CritiqueLabel = "ungrounded"
)

// Critique holds the extracted claims from a draft.
type Critique struct {
	Claims []CritiqueClaim
}

// CritiqueClaim is a single factual claim with a grounding label.
type CritiqueClaim struct {
	Text  string        `json:"text"`
	Label CritiqueLabel `json:"label"`
}

// ParseCritique extracts a JSON array of {text, label} objects from raw LLM output.
func ParseCritique(raw string) (Critique, error) {
	start := strings.IndexByte(raw, '[')
	if start < 0 {
		return Critique{}, fmt.Errorf("parseCritique: no JSON array found")
	}
	depth := 0
	end := -1
	for i := start; i < len(raw); i++ {
		switch raw[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				end = i
				goto found
			}
		}
	}
	return Critique{}, fmt.Errorf("parseCritique: no matching closing bracket")
found:
	var claims []CritiqueClaim
	if err := json.Unmarshal([]byte(raw[start:end+1]), &claims); err != nil {
		return Critique{}, fmt.Errorf("parseCritique: %w", err)
	}
	// Normalise unknown labels to ungrounded.
	for i := range claims {
		switch claims[i].Label {
		case LabelGrounded, LabelPartial, LabelUngrounded:
			// ok
		default:
			claims[i].Label = LabelUngrounded
		}
	}
	return Critique{Claims: claims}, nil
}

// EvidenceCoverage computes (grounded + 0.5*partial) / total. Empty = 0.
func EvidenceCoverage(c Critique) float64 {
	if len(c.Claims) == 0 {
		return 0
	}
	var score float64
	for _, cl := range c.Claims {
		switch cl.Label {
		case LabelGrounded:
			score += 1.0
		case LabelPartial:
			score += 0.5
		}
	}
	return score / float64(len(c.Claims))
}

// CrossCapabilityAgree scores based on the number of distinct sources in a bundle.
func CrossCapabilityAgree(bundle EvidenceBundle) float64 {
	sources := bundle.Sources()
	switch {
	case len(sources) >= 2:
		return 1.0
	case len(sources) == 1:
		return 0.4
	default:
		return 0.0
	}
}

var decimalRe = regexp.MustCompile(`\d+\.\d+|\d+`)

// ParseJudgeScore extracts the first decimal number from raw and clamps to [0,1].
func ParseJudgeScore(raw string) (float64, error) {
	m := decimalRe.FindString(raw)
	if m == "" {
		return 0, fmt.Errorf("parseJudgeScore: no number found in: %s", raw)
	}
	v, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0, fmt.Errorf("parseJudgeScore: %w", err)
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return v, nil
}

// Composite computes the equal-weight average of non-nil signals in a Score.
func Composite(s Score) float64 {
	var sum float64
	var count int
	if s.EvidenceCoverage != nil {
		sum += *s.EvidenceCoverage
		count++
	}
	if s.CrossCapabilityAgree != nil {
		sum += *s.CrossCapabilityAgree
		count++
	}
	if s.JudgeScore != nil {
		sum += *s.JudgeScore
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// ComputeScore runs CCA, critique, and judge scoring, returning the Score and Critique.
func ComputeScore(ctx context.Context, llm LLMFn, question, draft string, bundle EvidenceBundle) (Score, Critique) {
	var s Score

	// CCA is free — no LLM needed.
	cca := CrossCapabilityAgree(bundle)
	s.CrossCapabilityAgree = Ptr(cca)

	// Critique → evidence coverage.
	critiqueRaw, err := llm(ctx, CritiquePrompt(draft, bundle))
	var crit Critique
	if err == nil {
		crit, err = ParseCritique(critiqueRaw)
	}
	if err == nil {
		ec := EvidenceCoverage(crit)
		s.EvidenceCoverage = Ptr(ec)
	}

	// Judge score.
	judgeRaw, err := llm(ctx, JudgePrompt(question, draft, bundle))
	if err == nil {
		if js, jerr := ParseJudgeScore(judgeRaw); jerr == nil {
			s.JudgeScore = Ptr(js)
		}
	}

	s.Composite = Composite(s)
	return s, crit
}
