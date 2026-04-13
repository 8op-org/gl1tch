package research

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const acceptThreshold = 0.7

// Loop is the core research loop engine.
type Loop struct {
	registry *Registry
	llm      LLMFn
	draftLLM LLMFn
	model    string
	logger   *slog.Logger
	events   EventSink
	hints    string
}

// NewLoop creates a Loop with the given registry and LLM function.
func NewLoop(reg *Registry, llm LLMFn) *Loop {
	return &Loop{
		registry: reg,
		llm:      llm,
		model:    DefaultLocalModel,
		logger:   slog.New(slog.NewTextHandler(&discardWriter{}, nil)),
		events:   nopSink{},
	}
}

// WithDraftLLM sets a separate LLM for drafting (defaults to the main LLM).
func (l *Loop) WithDraftLLM(fn LLMFn) *Loop {
	l.draftLLM = fn
	return l
}

// WithEventSink sets the event sink for the loop.
func (l *Loop) WithEventSink(sink EventSink) *Loop {
	l.events = sink
	return l
}

// WithLogger sets the logger.
func (l *Loop) WithLogger(log *slog.Logger) *Loop {
	l.logger = log
	return l
}

// WithHints sets additional hints for the plan prompt.
func (l *Loop) WithHints(hints string) *Loop {
	l.hints = hints
	return l
}

// Run executes the research loop.
func (l *Loop) Run(ctx context.Context, q ResearchQuery, budget Budget) (Result, error) {
	if l.registry == nil {
		return Result{}, errors.New("research: registry is nil")
	}
	if l.llm == nil {
		return Result{}, errors.New("research: llm is nil")
	}

	if q.ID == "" {
		q.ID = newQueryID()
	}
	if budget.MaxIterations <= 0 {
		budget.MaxIterations = 3
	}

	draftFn := l.llm
	if l.draftLLM != nil {
		draftFn = l.draftLLM
	}

	var (
		bestResult Result
		haveBest   bool
		bundle     EvidenceBundle
		picked     = make(map[string]struct{})
	)

	for iter := 1; iter <= budget.MaxIterations; iter++ {
		l.logger.Info("research iteration", "iter", iter, "query", q.Question)

		// Plan
		picks, err := l.plan(ctx, q.Question, picked)
		if err != nil {
			l.logger.Warn("plan failed", "err", err)
			if iter == 1 {
				return Result{
					Query:      q,
					Draft:      "I don't have enough evidence",
					Reason:     ReasonUnscored,
					Iterations: iter,
				}, nil
			}
			break
		}

		if len(picks) == 0 {
			if iter == 1 {
				return Result{
					Query:      q,
					Draft:      "I don't have enough evidence",
					Reason:     ReasonUnscored,
					Iterations: iter,
				}, nil
			}
			break
		}

		for _, p := range picks {
			picked[p] = struct{}{}
		}

		// Gather
		evidence := l.gather(ctx, q, picks, bundle)
		for _, e := range evidence {
			bundle.Add(e)
		}

		// Draft
		draftPrompt := DraftPrompt(q.Question, bundle)
		draft, err := draftFn(ctx, draftPrompt)
		if err != nil {
			l.logger.Warn("draft failed", "err", err)
			continue
		}

		// Score
		score, crit := ComputeScore(ctx, l.llm, q.Question, draft, bundle)

		result := Result{
			Query:      q,
			Draft:      draft,
			Bundle:     bundle,
			Score:      score,
			Iterations: iter,
		}

		if !haveBest || score.Composite > bestResult.Score.Composite {
			bestResult = result
			haveBest = true
		}

		if score.Composite >= acceptThreshold {
			bestResult.Reason = ReasonAccepted
			emitAttempt(l.events, q.Question, q.ID, iter, score, &bundle, ReasonAccepted)
			return bestResult, nil
		}

		emitAttempt(l.events, q.Question, q.ID, iter, score, &bundle, ReasonBudgetExceeded)

		// Extract ungrounded claims for refinement context (used in next iteration hints).
		var ungrounded []string
		for _, cl := range crit.Claims {
			if cl.Label == LabelUngrounded {
				ungrounded = append(ungrounded, cl.Text)
			}
		}
		if len(ungrounded) > 0 {
			l.logger.Info("ungrounded claims for refinement", "claims", strings.Join(ungrounded, "; "))
		}
	}

	return returnBest(q, bestResult, haveBest, budget.MaxIterations)
}

// plan calls the LLM with PlanPrompt, parses researcher names, validates against registry.
func (l *Loop) plan(ctx context.Context, question string, already map[string]struct{}) ([]string, error) {
	researchers := l.registry.List()
	if len(researchers) == 0 {
		return nil, fmt.Errorf("no researchers registered")
	}

	prompt := PlanPrompt(question, researchers, l.hints)
	raw, err := l.llm(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("plan LLM call: %w", err)
	}

	names, err := ParsePlan(raw)
	if err != nil {
		return nil, fmt.Errorf("plan parse: %w", err)
	}

	// Validate against registry and filter already picked.
	var valid []string
	for _, name := range names {
		if _, ok := l.registry.Lookup(name); !ok {
			l.logger.Warn("plan: unknown researcher", "name", name)
			continue
		}
		valid = append(valid, name)
	}

	return filterPicked(valid, already), nil
}

// gather runs selected researchers in parallel.
func (l *Loop) gather(ctx context.Context, q ResearchQuery, picks []string, prior EvidenceBundle) []Evidence {
	type result struct {
		evidence Evidence
		err      error
	}

	results := make([]result, len(picks))
	var wg sync.WaitGroup

	for i, name := range picks {
		r, ok := l.registry.Lookup(name)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(idx int, researcher Researcher) {
			defer wg.Done()
			ev, err := researcher.Gather(ctx, q, prior)
			results[idx] = result{evidence: ev, err: err}
		}(i, r)
	}
	wg.Wait()

	var out []Evidence
	for _, res := range results {
		if res.err != nil {
			l.logger.Warn("gather: researcher error", "err", res.err)
			continue
		}
		if res.evidence.Source != "" {
			out = append(out, res.evidence)
		}
	}
	return out
}

// returnBest returns the best result or a budget-exceeded result.
func returnBest(q ResearchQuery, best Result, haveBest bool, iters int) (Result, error) {
	if haveBest {
		best.Reason = ReasonBudgetExceeded
		return best, nil
	}
	return Result{
		Query:      q,
		Draft:      "I don't have enough evidence",
		Reason:     ReasonUnscored,
		Iterations: iters,
	}, nil
}

// filterPicked removes names that already appear in the already set.
func filterPicked(picks []string, already map[string]struct{}) []string {
	var out []string
	for _, p := range picks {
		if _, ok := already[p]; !ok {
			out = append(out, p)
		}
	}
	return out
}

// newQueryID generates a unique query ID.
func newQueryID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("q-%d-%x", time.Now().UnixNano(), b)
}

// discardWriter implements io.Writer and discards all data.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// Ensure discardWriter satisfies io.Writer at compile time.
var _ io.Writer = discardWriter{}
