package research

import (
	"context"
	"encoding/json"
	"time"
)

type LLMFn func(ctx context.Context, prompt string) (string, error)

const DefaultLocalModel = "qwen2.5:7b"

type ResearchQuery struct {
	ID       string            `json:"id"`
	Question string            `json:"question"`
	Context  map[string]string `json:"context,omitempty"`
}

type Evidence struct {
	Source    string   `json:"source"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Refs      []string `json:"refs,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Truncated bool     `json:"truncated,omitempty"`
}

type EvidenceBundle struct {
	Items []Evidence `json:"items"`
}

// Add appends e to the bundle. Nil-safe.
func (b *EvidenceBundle) Add(e Evidence) {
	if b == nil {
		return
	}
	b.Items = append(b.Items, e)
}

// Len returns the number of items. Nil-safe.
func (b *EvidenceBundle) Len() int {
	if b == nil {
		return 0
	}
	return len(b.Items)
}

// Sources returns unique source strings in order of first appearance. Nil-safe.
func (b *EvidenceBundle) Sources() []string {
	if b == nil {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(b.Items))
	for _, item := range b.Items {
		if _, ok := seen[item.Source]; !ok {
			seen[item.Source] = struct{}{}
			out = append(out, item.Source)
		}
	}
	return out
}

type Budget struct {
	MaxIterations int
	MaxWallclock  time.Duration
}

func DefaultBudget() Budget {
	return Budget{MaxIterations: 3, MaxWallclock: 60 * time.Second}
}

type Reason string

const (
	ReasonAccepted       Reason = "accepted"
	ReasonBudgetExceeded Reason = "budget_exceeded"
	ReasonUnscored       Reason = "unscored"
)

// Score holds composite and optional sub-scores. Use Ptr/Float helpers for optional fields.
type Score struct {
	Composite            float64  `json:"composite"`
	EvidenceCoverage     *float64 `json:"evidence_coverage,omitempty"`
	CrossCapabilityAgree *float64 `json:"cross_capability_agreement,omitempty"`
	JudgeScore           *float64 `json:"judge_score,omitempty"`
}

// scoreAlias prevents infinite recursion in MarshalJSON.
type scoreAlias Score

// MarshalJSON uses the alias pattern to delegate to the default marshaller.
func (s Score) MarshalJSON() ([]byte, error) {
	return json.Marshal(scoreAlias(s))
}

// Ptr returns a pointer to v.
func Ptr(v float64) *float64 {
	return &v
}

// Float dereferences p; returns 0 if p is nil.
func Float(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// Feedback captures the paid model's assessment of evidence quality.
type Feedback struct {
	Quality    string   `json:"quality"`
	Missing    []string `json:"missing"`
	Useful     []string `json:"useful"`
	Suggestion string   `json:"suggestion"`
}

type Result struct {
	Query      ResearchQuery  `json:"query"`
	Draft      string         `json:"draft"`
	Bundle     EvidenceBundle `json:"bundle"`
	Score      Score          `json:"score"`
	Reason     Reason         `json:"reason"`
	Iterations int            `json:"iterations"`
	Feedback   Feedback       `json:"feedback,omitempty"`
}
