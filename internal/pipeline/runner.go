package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/ui"
	"golang.org/x/sync/errgroup"
)

const defaultOllamaURL = "http://localhost:11434"

// stepOutcome holds the result of executing a single step.
type stepOutcome struct {
	output     string
	tokensIn   int
	tokensOut  int
	cost       float64
	latencyMs  int64
	tier       int
	escalated  bool
	escChain   []int
	evalScores []int
	isLLM      bool
}

// runCtx bundles per-run state needed by executeStep and compound forms.
type runCtx struct {
	ctx              context.Context
	input            string
	params           map[string]string
	workspace        string
	defaultModel     string
	reg              *provider.ProviderRegistry
	providerResolver provider.ResolverFunc
	tiers            []provider.TierConfig
	evalThreshold    int
	steps            map[string]string
	prevStepID       string
	mu               sync.Mutex
	esURL            string
}

// stepsSnapshot returns a shallow copy of the steps map safe for concurrent reads.
func (r *runCtx) stepsSnapshot() map[string]string {
	r.mu.Lock()
	snap := make(map[string]string, len(r.steps))
	for k, v := range r.steps {
		snap[k] = v
	}
	r.mu.Unlock()
	return snap
}

// resolveESURL returns the ES URL to use, checking step override, then runCtx default, then fallback.
func resolveESURL(stepURL string, rctx *runCtx) string {
	if stepURL != "" {
		return stepURL
	}
	if rctx.esURL != "" {
		return rctx.esURL
	}
	return "http://localhost:9200"
}

// Result holds the output of a completed workflow run.
type Result struct {
	Workflow string
	Output   string            // output of the last step
	Steps    map[string]string // all step outputs keyed by step ID
}

// RunOpts holds optional dependencies for a workflow run.
type RunOpts struct {
	Telemetry        *esearch.Telemetry
	Issue            string
	ComparisonGroup  string
	ProviderResolver provider.ResolverFunc
	Tiers            []provider.TierConfig
	EvalThreshold    int
	SeedSteps        map[string]string // pre-computed step outputs; matching step IDs are skipped
	ESURL            string            // default ES URL from workspace config
	Workspace        string            // resolved workspace name for {{.workspace}} template var
}

// parseWorkflowName extracts issue number and comparison group from a workflow name.
// Convention: "issue-to-pr-local" → group="local", "3918-wrapper-curl-copilot" → issue="3918", group="copilot"
// The last hyphen-separated segment is the comparison group.
func parseWorkflowName(name string) (issue, compGroup string) {
	if idx := strings.LastIndex(name, "-"); idx > 0 {
		compGroup = name[idx+1:]
		name = name[:idx]
	}
	for i, c := range name {
		if c < '0' || c > '9' {
			if i > 0 {
				issue = name[:i]
			}
			break
		}
	}
	return
}

// Run executes a workflow with the given input string.
// Templates use {{.input}} for user input and {{step "id"}} for prior step output.
func Run(w *Workflow, input string, defaultModel string, params map[string]string, reg *provider.ProviderRegistry, opts ...RunOpts) (*Result, error) {
	steps := make(map[string]string) // step ID → output

	var tel *esearch.Telemetry
	var issue, compGroup string
	var providerResolver provider.ResolverFunc
	var tiers []provider.TierConfig
	var evalThreshold int
	if len(opts) > 0 {
		tel = opts[0].Telemetry
		issue = opts[0].Issue
		compGroup = opts[0].ComparisonGroup
		providerResolver = opts[0].ProviderResolver
		tiers = opts[0].Tiers
		evalThreshold = opts[0].EvalThreshold
	}
	if evalThreshold == 0 {
		evalThreshold = 4
	}
	if issue == "" || compGroup == "" {
		parsedIssue, parsedGroup := parseWorkflowName(w.Name)
		if issue == "" {
			issue = parsedIssue
		}
		if compGroup == "" {
			compGroup = parsedGroup
		}
	}
	// Fall back to params for issue if not derived from workflow name
	if issue == "" && params != nil {
		issue = params["issue"]
	}

	runID := esearch.NewRunID()

	// Accumulators for workflow run summary
	var totalTokensIn, totalTokensOut int64
	var totalCostUSD float64
	var totalLatencyMS int64
	var llmSteps int
	var lastLLMOutput string

	// Pre-populate steps from seed (data-gathering results shared across variants)
	if len(opts) > 0 && opts[0].SeedSteps != nil {
		for k, v := range opts[0].SeedSteps {
			steps[k] = v
		}
	}

	var esURL string
	if len(opts) > 0 && opts[0].ESURL != "" {
		esURL = opts[0].ESURL
	}

	var workspaceName string
	if len(opts) > 0 {
		workspaceName = opts[0].Workspace
	}

	rctx := &runCtx{
		ctx:              context.Background(),
		input:            input,
		params:           params,
		workspace:        workspaceName,
		defaultModel:     defaultModel,
		reg:              reg,
		providerResolver: providerResolver,
		tiers:            tiers,
		evalThreshold:    evalThreshold,
		steps:            steps,
		esURL:            esURL,
	}

	// runBareStep executes a single step and accumulates telemetry.
	runBareStep := func(step Step) error {
		if _, seeded := steps[step.ID]; seeded {
			ui.StepSDK(step.ID, "seeded, skipped")
			return nil
		}

		outcome, err := executeStep(rctx.ctx, rctx, step)
		if err != nil {
			return err
		}

		if outcome.isLLM {
			tokIn := int64(outcome.tokensIn)
			tokOut := int64(outcome.tokensOut)
			totalTokensIn += tokIn
			totalTokensOut += tokOut
			totalCostUSD += outcome.cost
			llmSteps++
			lastLLMOutput = outcome.output

			if tel != nil {
				prov := ""
				model := defaultModel
				if step.LLM != nil {
					prov = strings.ToLower(step.LLM.Provider)
					if step.LLM.Model != "" {
						model = step.LLM.Model
					}
				}
				reason := ""
				if outcome.escalated {
					reason = "eval"
				}
				tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
					RunID:            runID,
					Step:             fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
					Tier:             outcome.tier,
					Provider:         prov,
					Model:            model,
					TokensIn:         tokIn,
					TokensOut:        tokOut,
					TokensTotal:      tokIn + tokOut,
					CostUSD:          outcome.cost,
					LatencyMS:        outcome.latencyMs,
					Escalated:        outcome.escalated,
					EscalationReason: reason,
					EscalationChain:  outcome.escChain,
					EvalScores:       outcome.evalScores,
					FinalTier:        outcome.tier,
					WorkflowName:     w.Name,
					Issue:            issue,
					ComparisonGroup:  compGroup,
					Timestamp:        time.Now().UTC().Format(time.RFC3339),
				})
			}
		}
		return nil
	}

	// If Items is populated (sexpr with phases), walk the ordered item list.
	// Otherwise fall back to flat Steps for backward compat.
	if len(w.Items) > 0 {
		for _, item := range w.Items {
			if item.Step != nil {
				if err := runBareStep(*item.Step); err != nil {
					return nil, err
				}
			} else if item.Phase != nil {
				report, err := executePhase(rctx, *item.Phase)
				if err != nil {
					if report != nil {
						rctx.mu.Lock()
						rctx.steps["_verification_report"] = report.FormatReport()
						rctx.mu.Unlock()
					}
					return nil, err
				}
			}
		}
	} else {
		for i, step := range w.Steps {
			if i > 0 {
				rctx.prevStepID = w.Steps[i-1].ID
			}
			if err := runBareStep(step); err != nil {
				return nil, err
			}
		}
	}

	// Index workflow run summary
	if tel != nil {
		// Strip markdown bold/italic markers before checking review verdict
		stripped := strings.ReplaceAll(strings.ToUpper(lastLLMOutput), "*", "")
		reviewPass := strings.Contains(stripped, "OVERALL: PASS")

		// Parse per-criterion PASS/FAIL for confidence score
		passed, total := 0, 0
		for _, line := range strings.Split(stripped, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "OVERALL") {
				continue // skip the summary line
			}
			hasPass := strings.Contains(line, "PASS")
			hasFail := strings.Contains(line, "FAIL")
			if hasPass && !hasFail {
				passed++
				total++
			} else if hasFail {
				total++
			}
		}
		confidence := 0.0
		if total > 0 {
			confidence = float64(passed) / float64(total)
		}

		tel.IndexWorkflowRun(context.Background(), esearch.WorkflowRunDoc{
			RunID:           runID,
			WorkflowName:    w.Name,
			Issue:           issue,
			ComparisonGroup: compGroup,
			TotalSteps:      len(w.Steps),
			LLMSteps:        llmSteps,
			TotalTokensIn:   totalTokensIn,
			TotalTokensOut:  totalTokensOut,
			TotalCostUSD:    totalCostUSD,
			TotalLatencyMS:  totalLatencyMS,
			ReviewPass:      reviewPass,
			Confidence:      confidence,
			CriteriaPassed:  passed,
			CriteriaTotal:   total,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
		})

		// If this is a cross-review workflow, parse and index per-variant scores
		if strings.Contains(w.Name, "cross-review") {
			crossScores := ParseCrossReview(lastLLMOutput)
			iteration := ""
			if params != nil {
				iteration = params["iteration"]
			}
			for _, cs := range crossScores {
				conf := 0.0
				if cs.Total > 0 {
					conf = float64(cs.Passed) / float64(cs.Total)
				}
				tel.IndexCrossReview(context.Background(), esearch.CrossReviewDoc{
					RunID:        runID,
					Issue:        issue,
					Iteration:    iteration,
					Variant:      cs.Variant,
					Passed:       cs.Passed,
					Total:        cs.Total,
					Confidence:   conf,
					Winner:       cs.Winner,
					WorkflowName: w.Name,
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				})
			}
		}
	}

	// Determine last step output for Result.Output
	var lastOutput string
	if len(w.Items) > 0 {
		// Walk items in reverse to find the last step/phase output
		for i := len(w.Items) - 1; i >= 0; i-- {
			item := w.Items[i]
			if item.Step != nil {
				lastOutput = steps[item.Step.ID]
				break
			} else if item.Phase != nil {
				// Last step in the phase
				p := item.Phase
				if len(p.Steps) > 0 {
					lastOutput = steps[p.Steps[len(p.Steps)-1].ID]
					break
				}
			}
		}
	} else if len(w.Steps) > 0 {
		lastOutput = steps[w.Steps[len(w.Steps)-1].ID]
	}
	return &Result{
		Workflow: w.Name,
		Output:   lastOutput,
		Steps:    steps,
	}, nil
}

// PreRunSharedSteps executes only the (run ...) shell steps from a workflow and
// returns their outputs keyed by step ID. Used by batch to run data-gathering
// steps once and seed all variant runs with the results.
func PreRunSharedSteps(w *Workflow, params map[string]string) (map[string]string, error) {
	steps := make(map[string]string)
	rctx := &runCtx{
		ctx:    context.Background(),
		params: params,
		steps:  steps,
	}
	for i, step := range w.Steps {
		if step.Run == "" {
			continue
		}
		if i > 0 {
			rctx.prevStepID = w.Steps[i-1].ID
		}
		ui.StepShell(step.ID)
		outcome, err := executeStep(rctx.ctx, rctx, step)
		if err != nil {
			return nil, fmt.Errorf("shared step %s: %w", step.ID, err)
		}
		steps[step.ID] = outcome.output
	}
	return steps, nil
}

// CrossReviewScore holds a parsed per-variant score from a cross-review.
type CrossReviewScore struct {
	Variant string
	Passed  int
	Total   int
	Winner  bool
}

// ParseCrossReview extracts per-variant scores from cross-review LLM output.
// Supports two formats:
//
// Old format (PASS/FAIL):
//
//	--- LOCAL ---
//	1. Specificity — PASS — good
//	SCORE: 4/5
//	WINNER: LOCAL
//
// New format (numeric scores):
//
//	VARIANT: local
//	plan_completeness: 9/10
//	total: 36/40
//	WINNER: local
func ParseCrossReview(output string) []CrossReviewScore {
	upper := strings.ToUpper(strings.ReplaceAll(output, "*", ""))

	// Detect which format: new format uses "VARIANT:" headers
	if strings.Contains(upper, "\nVARIANT:") || strings.HasPrefix(upper, "VARIANT:") {
		return parseCrossReviewNumeric(output)
	}
	return parseCrossReviewPassFail(output)
}

// parseCrossReviewNumeric handles the new format with VARIANT: headers and N/M scores.
// A score >= 7 out of 10 counts as "passed". The total is the count of score lines.
func parseCrossReviewNumeric(output string) []CrossReviewScore {
	var scores []CrossReviewScore
	lines := strings.Split(output, "\n")

	// Find WINNER line (case-insensitive)
	winner := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "WINNER:") {
			winner = strings.TrimSpace(trimmed[len("WINNER:"):])
			break
		}
	}

	currentVariant := ""
	passed, total := 0, 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)

		// Detect VARIANT: header
		if strings.HasPrefix(upper, "VARIANT:") {
			// Save previous variant
			if currentVariant != "" && total > 0 {
				isWinner := strings.EqualFold(winner, currentVariant)
				scores = append(scores, CrossReviewScore{
					Variant: strings.ToLower(currentVariant),
					Passed:  passed,
					Total:   total,
					Winner:  isWinner,
				})
			}
			currentVariant = strings.TrimSpace(trimmed[len("VARIANT:"):])
			passed = 0
			total = 0
			continue
		}

		// Skip non-score lines
		if strings.HasPrefix(upper, "WINNER:") || strings.HasPrefix(upper, "REASON:") ||
			strings.HasPrefix(upper, "NOTES:") || strings.HasPrefix(upper, "TOTAL:") {
			continue
		}

		// Parse score lines like "plan_completeness: 9/10"
		if currentVariant != "" && strings.Contains(trimmed, "/") && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				scorePart := strings.TrimSpace(parts[1])
				numDenom := strings.SplitN(scorePart, "/", 2)
				if len(numDenom) == 2 {
					num, errN := strconv.Atoi(strings.TrimSpace(numDenom[0]))
					_, errD := strconv.Atoi(strings.TrimSpace(numDenom[1]))
					if errN == nil && errD == nil {
						total++
						if num >= 7 {
							passed++
						}
					}
				}
			}
		}
	}

	// Save last variant
	if currentVariant != "" && total > 0 {
		isWinner := strings.EqualFold(winner, currentVariant)
		scores = append(scores, CrossReviewScore{
			Variant: strings.ToLower(currentVariant),
			Passed:  passed,
			Total:   total,
			Winner:  isWinner,
		})
	}

	return scores
}

// parseCrossReviewPassFail handles the old format with --- VARIANT --- headers and PASS/FAIL lines.
func parseCrossReviewPassFail(output string) []CrossReviewScore {
	var scores []CrossReviewScore
	upper := strings.ToUpper(strings.ReplaceAll(output, "*", ""))
	lines := strings.Split(upper, "\n")

	// Find WINNER line
	winner := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "WINNER:") {
			winner = strings.TrimSpace(strings.TrimPrefix(line, "WINNER:"))
			break
		}
	}

	// Parse --- VARIANT --- sections
	currentVariant := ""
	passed, total := 0, 0
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect variant header: --- LOCAL --- or --- CLAUDE --- etc
		if strings.HasPrefix(line, "---") && strings.HasSuffix(line, "---") {
			// Save previous variant if any
			if currentVariant != "" && total > 0 {
				isWinner := strings.Contains(strings.ToUpper(winner), strings.ToUpper(currentVariant))
				scores = append(scores, CrossReviewScore{
					Variant: strings.ToLower(currentVariant),
					Passed:  passed,
					Total:   total,
					Winner:  isWinner,
				})
			}
			// Extract variant name
			name := strings.Trim(line, "- ")
			if name != "" {
				currentVariant = name
				passed = 0
				total = 0
			}
			continue
		}

		// Skip SCORE: lines (we count from PASS/FAIL)
		if strings.HasPrefix(line, "SCORE:") {
			continue
		}

		// Skip OVERALL and WINNER lines
		if strings.HasPrefix(line, "OVERALL") || strings.HasPrefix(line, "WINNER") {
			continue
		}

		// Count PASS/FAIL within current variant section
		if currentVariant != "" {
			hasPass := strings.Contains(line, "PASS")
			hasFail := strings.Contains(line, "FAIL")
			if hasPass && !hasFail {
				passed++
				total++
			} else if hasFail {
				total++
			}
		}
	}

	// Don't forget the last variant
	if currentVariant != "" && total > 0 {
		isWinner := strings.Contains(strings.ToUpper(winner), strings.ToUpper(currentVariant))
		scores = append(scores, CrossReviewScore{
			Variant: strings.ToLower(currentVariant),
			Passed:  passed,
			Total:   total,
			Winner:  isWinner,
		})
	}

	return scores
}

func render(tmpl string, data map[string]any, steps map[string]string) (string, error) {
	funcMap := template.FuncMap{
		"step": func(id string) string {
			return steps[id]
		},
		// stepfile writes step output to a temp file and returns the path.
		// Use in shell steps where inline content would break escaping:
		//   cat "{{stepfile "fetch-issue"}}"
		"stepfile": func(id string) string {
			content, ok := steps[id]
			if !ok {
				return ""
			}
			f, err := os.CreateTemp("", "glitch-step-*")
			if err != nil {
				return ""
			}
			f.WriteString(content)
			f.Close()
			return f.Name()
		},

		// String functions
		"split":      func(sep, s string) []string { return strings.Split(s, sep) },
		"join":       func(sep string, parts []string) string { return strings.Join(parts, sep) },
		"last":       func(s []string) string { if len(s) == 0 { return "" }; return s[len(s)-1] },
		"first":      func(s []string) string { if len(s) == 0 { return "" }; return s[0] },
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"trim":       strings.TrimSpace,
		"trimPrefix": func(prefix, s string) string { return strings.TrimPrefix(s, prefix) },
		"trimSuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
		"replace":    func(old, new, s string) string { return strings.ReplaceAll(s, old, new) },
		"truncate": func(n int, s string) string {
			runes := []rune(s)
			if len(runes) <= n {
				return s
			}
			return string(runes[:n])
		},
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
	}
	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// executeStep handles control-flow forms (retry, timeout, catch, cond, map)
// and delegates to runSingleStep for actual execution.
func executeStep(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	// Apply timeout before dispatching compound forms so (timeout "1s" (par ...)) works.
	if step.Timeout != "" {
		dur, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return nil, fmt.Errorf("step %s: invalid timeout %q: %w", step.ID, step.Timeout, err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dur)
		defer cancel()
	}

	switch step.Form {
	case "cond":
		return executeCond(ctx, rctx, step)
	case "map":
		return executeMap(ctx, rctx, step)
	case "par":
		return executePar(ctx, rctx, step)
	}

	maxAttempts := 1 + step.Retry
	var lastErr error
	for attempt := range maxAttempts {
		if attempt > 0 {
			ui.StepRetry(step.ID, attempt, step.Retry)
		}

		outcome, err := runSingleStep(ctx, rctx, step)
		if err == nil {
			// catch form: step succeeded, store output normally
			rctx.mu.Lock()
			rctx.steps[step.ID] = outcome.output
			rctx.mu.Unlock()
			return outcome, nil
		}
		lastErr = err

		// Check if context expired (timeout) — no point retrying
		if ctx.Err() != nil {
			break
		}
	}

	// All retries exhausted — try fallback if catch form
	if step.Form == "catch" && step.Fallback != nil {
		ui.StepFallback(step.ID, step.Fallback.ID)
		outcome, err := runSingleStep(ctx, rctx, *step.Fallback)
		if err != nil {
			return nil, fmt.Errorf("step %s fallback %s: %w", step.ID, step.Fallback.ID, err)
		}
		rctx.mu.Lock()
		rctx.steps[step.Fallback.ID] = outcome.output
		rctx.steps[step.ID] = outcome.output // catch step ID also gets the fallback output
		rctx.mu.Unlock()
		return outcome, nil
	}

	return nil, lastErr
}

// executeCond evaluates predicate branches in order and runs the first matching step.
func executeCond(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	for _, branch := range step.Branches {
		if branch.Pred == "else" {
			return executeStep(ctx, rctx, branch.Step)
		}
		// Render predicate template
		data := map[string]any{"input": rctx.input, "param": rctx.params, "workspace": rctx.workspace}
		rendered, err := render(branch.Pred, data, rctx.stepsSnapshot())
		if err != nil {
			return nil, fmt.Errorf("cond %s: template: %w", step.ID, err)
		}
		cmd := exec.CommandContext(ctx, "sh", "-c", rendered)
		if err := cmd.Run(); err == nil {
			// Predicate succeeded (exit 0)
			return executeStep(ctx, rctx, branch.Step)
		}
	}
	// No branch matched — return empty outcome
	rctx.mu.Lock()
	rctx.steps[step.ID] = ""
	rctx.mu.Unlock()
	return &stepOutcome{}, nil
}

// executeMap splits a prior step's output by newlines and runs the body step for each item.
func executeMap(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	snap := rctx.stepsSnapshot()
	source, ok := snap[step.MapOver]
	if !ok {
		return nil, fmt.Errorf("map %s: source step %q has no output", step.ID, step.MapOver)
	}

	lines := strings.Split(strings.TrimSpace(source), "\n")
	var outputs []string

	for idx, item := range lines {
		if item == "" {
			continue
		}
		// Clone the body step with a unique ID per iteration
		body := *step.MapBody
		body.ID = fmt.Sprintf("%s-%d", step.MapBody.ID, idx)

		// Make {{.item}} available in templates by injecting into params
		origParams := rctx.params
		mapParams := make(map[string]string, len(origParams)+1)
		for k, v := range origParams {
			mapParams[k] = v
		}
		mapParams["item"] = item
		mapParams["item_index"] = fmt.Sprintf("%d", idx)
		rctx.params = mapParams

		outcome, err := executeStep(ctx, rctx, body)
		rctx.params = origParams
		if err != nil {
			return nil, fmt.Errorf("map %s item %d: %w", step.ID, idx, err)
		}
		outputs = append(outputs, outcome.output)
	}

	combined := strings.Join(outputs, "\n")
	rctx.mu.Lock()
	rctx.steps[step.ID] = combined
	rctx.mu.Unlock()
	return &stepOutcome{output: combined}, nil
}

// executePar runs all child steps concurrently and waits for completion.
// Fail-fast: first error cancels all siblings via context.
func executePar(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	g, gctx := errgroup.WithContext(ctx)

	type parResult struct {
		outcome *stepOutcome
	}
	results := make([]parResult, len(step.ParSteps))

	for i, child := range step.ParSteps {
		g.Go(func() error {
			outcome, err := executeStep(gctx, rctx, child)
			if err != nil {
				return fmt.Errorf("par step %q: %w", child.ID, err)
			}
			results[i] = parResult{outcome: outcome}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var combined []string
	composite := &stepOutcome{}
	for _, r := range results {
		if r.outcome != nil {
			combined = append(combined, r.outcome.output)
			if r.outcome.isLLM {
				composite.isLLM = true
				composite.tokensIn += r.outcome.tokensIn
				composite.tokensOut += r.outcome.tokensOut
				composite.cost += r.outcome.cost
				if r.outcome.latencyMs > composite.latencyMs {
					composite.latencyMs = r.outcome.latencyMs
				}
			}
		}
	}

	composite.output = strings.Join(combined, "\n")
	rctx.mu.Lock()
	rctx.steps[step.ID] = composite.output
	rctx.mu.Unlock()

	return composite, nil
}

// runSingleStep executes one step (save, run, or llm) without control-flow wrappers.
func runSingleStep(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	// Snapshot steps map for safe concurrent reads during par execution.
	stepsSnap := rctx.stepsSnapshot()

	data := map[string]any{
		"input":     rctx.input,
		"param":     rctx.params,
		"workspace": rctx.workspace,
	}

	if step.Save != "" {
		rendered, err := render(step.Save, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		ui.StepSave(step.ID, rendered)
		sourceStep := step.SaveStep
		if sourceStep == "" && rctx.prevStepID != "" {
			sourceStep = rctx.prevStepID
		}
		content := stepsSnap[sourceStep]
		if err := os.MkdirAll(filepath.Dir(rendered), 0o755); err != nil {
			return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
		}
		if err := os.WriteFile(rendered, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("step %s: write: %w", step.ID, err)
		}
		out := fmt.Sprintf("saved %s to %s", sourceStep, rendered)
		return &stepOutcome{output: out}, nil
	}

	if step.Run != "" {
		rendered, err := render(step.Run, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		if !step.IsGate {
			if step.Hint != "" {
				ui.StepRunning(step.ID, step.Hint)
			} else {
				ui.StepShell(step.ID)
			}
		}
		out, err := provider.RunShellContext(ctx, rendered)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		return &stepOutcome{output: out}, nil
	}

	if step.LLM != nil {
		rendered, err := render(step.LLM.Prompt, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}

		if step.LLM.Skill != "" {
			skillContent, skillErr := loadSkill(step.LLM.Skill)
			if skillErr != nil {
				return nil, fmt.Errorf("step %s: skill %q: %w", step.ID, step.LLM.Skill, skillErr)
			}
			rendered = skillContent + "\n\n---\n\n" + rendered
		}

		prov := strings.ToLower(step.LLM.Provider)
		model := step.LLM.Model
		displayModel := step.LLM.Model // only show what the user explicitly set
		if model == "" {
			model = rctx.defaultModel
		}

		ui.StepLLM(step.ID, prov, displayModel)

		var out string
		var stepTier int
		var stepEscalated bool
		var stepEscalationChain []int
		var stepEvalScores []int
		var stepCost float64
		var stepTokensIn, stepTokensOut int
		llmStart := time.Now()

		useSmart := prov == "" && step.LLM.Tier == nil && len(rctx.tiers) > 0
		usePinned := step.LLM.Tier != nil && len(rctx.tiers) > 0

		if useSmart || usePinned {
			activeTiers := rctx.tiers
			if usePinned {
				tierIdx := *step.LLM.Tier
				if tierIdx >= 0 && tierIdx < len(rctx.tiers) {
					activeTiers = rctx.tiers[tierIdx : tierIdx+1]
				}
			}

			runner := provider.NewTieredRunner(activeTiers, rctx.reg)
			runner.Resolver = rctx.providerResolver
			runner.Log = ui.TierLog

			format := step.LLM.Format
			evalFn := func(evalModel, evalPrompt string) (provider.LLMResult, error) {
				return provider.RunOllamaWithResult(rctx.defaultModel, evalPrompt)
			}

			rr, llmErr := runner.RunSmart(ctx, rendered, format, rctx.evalThreshold, evalFn)
			if llmErr != nil {
				return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
			}
			out = rr.Response
			stepTier = rr.Tier
			if usePinned {
				stepTier = *step.LLM.Tier
			}
			stepEscalated = rr.Escalated
			stepEscalationChain = rr.EscalationChain
			stepEvalScores = rr.EvalScores
			stepCost = rr.CostUSD
			stepTokensIn = rr.TokensIn
			stepTokensOut = rr.TokensOut
		} else {
			switch prov {
			case "ollama", "":
				if model == "" {
					model = "qwen3:8b"
				}
				result, llmErr := provider.RunOllamaWithResult(model, rendered)
				if llmErr != nil {
					return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
				}
				out = result.Response
				stepTokensIn = result.TokensIn
				stepTokensOut = result.TokensOut
			case "lm-studio":
				lms := &provider.LMStudioProvider{
					BaseURL:      "http://localhost:1234",
					DefaultModel: "qwen3-8b",
				}
				result, llmErr := lms.Chat(model, rendered)
				if llmErr != nil {
					return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
				}
				out = result.Response
				stepTokensIn = result.TokensIn
				stepTokensOut = result.TokensOut
			default:
				var resolved bool

				if provider.IsAgent(prov) {
					agent := provider.KnownAgents[prov]
					result, llmErr := agent.Run(step.LLM.Model, rendered)
					if llmErr != nil {
						return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
					}
					resolved = true
					out = result.Response
					stepTokensIn = result.TokensIn
					stepTokensOut = result.TokensOut
					stepCost = result.CostUSD
				}

				if !resolved && rctx.providerResolver != nil {
					if fn, ok := rctx.providerResolver(prov); ok {
						resolved = true
						result, llmErr := fn(step.LLM.Model, rendered)
						if llmErr != nil {
							return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
						}
						out = result.Response
						stepTokensIn = result.TokensIn
						stepTokensOut = result.TokensOut
						stepCost = result.CostUSD
					}
				}

				if !resolved {
					result, provErr := rctx.reg.RunProviderWithResult(prov, model, rendered)
					if provErr != nil {
						return nil, fmt.Errorf("step %s: %w", step.ID, provErr)
					}
					out = result.Response
					stepTokensIn = result.TokensIn
					stepTokensOut = result.TokensOut
					stepCost = result.CostUSD
				}
			}
		}

		// Estimate cost if provider didn't report it
		if stepCost == 0 && (stepTokensIn+stepTokensOut) > 0 {
			stepCost = estimateCost(prov, model, stepTokensIn, stepTokensOut)
		}

		ui.StepLLMDone(step.ID, prov, displayModel, int64(stepTokensIn), int64(stepTokensOut), time.Since(llmStart))

		return &stepOutcome{
			output:     out,
			tokensIn:   stepTokensIn,
			tokensOut:  stepTokensOut,
			cost:       stepCost,
			latencyMs:  time.Since(llmStart).Milliseconds(),
			tier:       stepTier,
			escalated:  stepEscalated,
			escChain:   stepEscalationChain,
			evalScores: stepEvalScores,
			isLLM:      true,
		}, nil
	}

	// --- SDK forms ---

	if step.JsonPick != nil {
		from := stepsSnap[step.JsonPick.From]
		rendered, err := render(step.JsonPick.Expr, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		ui.StepSDK(step.ID, "json-pick")
		cmd := exec.CommandContext(ctx, "jq", rendered)
		cmd.Stdin = strings.NewReader(from)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("step %s: jq: %w", step.ID, err)
		}
		return &stepOutcome{output: strings.TrimSpace(string(out))}, nil
	}

	if step.Lines != "" {
		from := stepsSnap[step.Lines]
		lines := strings.Split(strings.TrimSpace(from), "\n")
		var parts []string
		for _, l := range lines {
			if l == "" {
				continue
			}
			b, _ := json.Marshal(l)
			parts = append(parts, string(b))
		}
		out := "[" + strings.Join(parts, ",") + "]"
		ui.StepSDK(step.ID, "lines")
		return &stepOutcome{output: out}, nil
	}

	if len(step.Merge) > 0 {
		merged := make(map[string]any)
		for _, id := range step.Merge {
			var obj map[string]any
			if err := json.Unmarshal([]byte(stepsSnap[id]), &obj); err != nil {
				return nil, fmt.Errorf("step %s: merge %q: %w", step.ID, id, err)
			}
			for k, v := range obj {
				merged[k] = v
			}
		}
		out, err := json.Marshal(merged)
		if err != nil {
			return nil, fmt.Errorf("step %s: merge marshal: %w", step.ID, err)
		}
		ui.StepSDK(step.ID, "merge")
		return &stepOutcome{output: string(out)}, nil
	}

	if step.HttpCall != nil {
		urlRendered, err := render(step.HttpCall.URL, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: url template: %w", step.ID, err)
		}
		var bodyReader io.Reader
		if step.HttpCall.Body != "" {
			bodyRendered, err := render(step.HttpCall.Body, data, stepsSnap)
			if err != nil {
				return nil, fmt.Errorf("step %s: body template: %w", step.ID, err)
			}
			bodyReader = strings.NewReader(bodyRendered)
		}
		method := step.HttpCall.Method
		if method == "" {
			method = "GET"
		}
		req, err := http.NewRequestWithContext(ctx, method, urlRendered, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("step %s: http request: %w", step.ID, err)
		}
		for k, v := range step.HttpCall.Headers {
			hv, err := render(v, data, stepsSnap)
			if err != nil {
				return nil, fmt.Errorf("step %s: header %q template: %w", step.ID, k, err)
			}
			req.Header.Set(k, hv)
		}
		ui.StepSDK(step.ID, "http-"+strings.ToLower(method))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("step %s: http: %w", step.ID, err)
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("step %s: http read body: %w", step.ID, err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("step %s: http %d: %s", step.ID, resp.StatusCode, strings.TrimSpace(string(respBody)))
		}
		return &stepOutcome{output: string(respBody)}, nil
	}

	if step.ReadFile != "" {
		rendered, err := render(step.ReadFile, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		ui.StepSDK(step.ID, "read-file")
		content, err := os.ReadFile(rendered)
		if err != nil {
			return nil, fmt.Errorf("step %s: read-file: %w", step.ID, err)
		}
		return &stepOutcome{output: string(content)}, nil
	}

	if step.WriteFile != nil {
		rendered, err := render(step.WriteFile.Path, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		content := stepsSnap[step.WriteFile.From]
		if err := os.MkdirAll(filepath.Dir(rendered), 0o755); err != nil {
			return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
		}
		if err := os.WriteFile(rendered, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("step %s: write-file: %w", step.ID, err)
		}
		ui.StepSDK(step.ID, "write-file")
		return &stepOutcome{output: rendered}, nil
	}

	if step.GlobPat != nil {
		pattern := step.GlobPat.Pattern
		if step.GlobPat.Dir != "" {
			dirRendered, err := render(step.GlobPat.Dir, data, stepsSnap)
			if err != nil {
				return nil, fmt.Errorf("step %s: dir template: %w", step.ID, err)
			}
			pattern = filepath.Join(dirRendered, pattern)
		}
		ui.StepSDK(step.ID, "glob")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("step %s: glob: %w", step.ID, err)
		}
		return &stepOutcome{output: strings.Join(matches, "\n")}, nil
	}

	if step.Search != nil {
		indexRendered, err := render(step.Search.IndexName, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: index template: %w", step.ID, err)
		}
		esURL := resolveESURL(step.Search.ESURL, rctx)
		es := esearch.NewClient(esURL)

		queryBody := map[string]any{}
		if step.Search.Query != "" {
			var q any
			if err := json.Unmarshal([]byte(step.Search.Query), &q); err != nil {
				return nil, fmt.Errorf("step %s: query parse: %w", step.ID, err)
			}
			queryBody["query"] = q
		}
		queryBody["size"] = step.Search.Size
		if len(step.Search.Fields) > 0 {
			queryBody["_source"] = step.Search.Fields
		}

		queryJSON, err := json.Marshal(queryBody)
		if err != nil {
			return nil, fmt.Errorf("step %s: query marshal: %w", step.ID, err)
		}

		ui.StepSDK(step.ID, "search")
		resp, err := es.Search(ctx, []string{indexRendered}, queryJSON)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}

		sources := make([]json.RawMessage, 0)
		for _, hit := range resp.Results {
			sources = append(sources, hit.Source)
		}
		out, err := json.Marshal(sources)
		if err != nil {
			return nil, fmt.Errorf("step %s: marshal results: %w", step.ID, err)
		}
		return &stepOutcome{output: string(out)}, nil
	}

	if step.Index != nil {
		indexRendered, err := render(step.Index.IndexName, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: index template: %w", step.ID, err)
		}
		docRendered, err := render(step.Index.Doc, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: doc template: %w", step.ID, err)
		}
		idRendered := step.Index.DocID
		if idRendered != "" {
			idRendered, err = render(idRendered, data, stepsSnap)
			if err != nil {
				return nil, fmt.Errorf("step %s: id template: %w", step.ID, err)
			}
		}

		docBytes := []byte(docRendered)

		if step.Index.EmbedField != "" {
			var docMap map[string]any
			if err := json.Unmarshal(docBytes, &docMap); err != nil {
				return nil, fmt.Errorf("step %s: parse doc for embedding: %w", step.ID, err)
			}
			fieldVal, ok := docMap[step.Index.EmbedField]
			if ok {
				text := fmt.Sprintf("%v", fieldVal)
				vec, err := provider.EmbedOllama(ctx, defaultOllamaURL, step.Index.EmbedModel, text)
				if err != nil {
					return nil, fmt.Errorf("step %s: embed: %w", step.ID, err)
				}
				docMap["embedding"] = vec
				docBytes, err = json.Marshal(docMap)
				if err != nil {
					return nil, fmt.Errorf("step %s: marshal embedded doc: %w", step.ID, err)
				}
			}
		}

		esURL := resolveESURL(step.Index.ESURL, rctx)
		es := esearch.NewClient(esURL)
		ui.StepSDK(step.ID, "index")
		resp, err := es.IndexDoc(ctx, indexRendered, idRendered, docBytes)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		out, _ := json.Marshal(resp)
		return &stepOutcome{output: string(out)}, nil
	}

	if step.Delete != nil {
		indexRendered, err := render(step.Delete.IndexName, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: index template: %w", step.ID, err)
		}
		esURL := resolveESURL(step.Delete.ESURL, rctx)
		es := esearch.NewClient(esURL)

		var q any
		if err := json.Unmarshal([]byte(step.Delete.Query), &q); err != nil {
			return nil, fmt.Errorf("step %s: query parse: %w", step.ID, err)
		}
		wrapped, _ := json.Marshal(map[string]any{"query": q})

		ui.StepSDK(step.ID, "delete")
		resp, err := es.DeleteByQuery(ctx, indexRendered, wrapped)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		out, _ := json.Marshal(resp)
		return &stepOutcome{output: string(out)}, nil
	}

	if step.Embed != nil {
		rendered, err := render(step.Embed.Input, data, stepsSnap)
		if err != nil {
			return nil, fmt.Errorf("step %s: input template: %w", step.ID, err)
		}
		ui.StepSDK(step.ID, "embed")
		vec, err := provider.EmbedOllama(ctx, defaultOllamaURL, step.Embed.Model, rendered)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		out, _ := json.Marshal(vec)
		return &stepOutcome{output: string(out)}, nil
	}

	if step.PluginCall != nil {
		ui.StepPlugin(step.ID, step.PluginCall.Plugin, step.PluginCall.Subcommand)
		pc := step.PluginCall

		// Render template expressions in plugin args (e.g., {{.param.repo}})
		renderedArgs := make(map[string]string, len(pc.Args))
		for k, v := range pc.Args {
			rendered, err := render(v, data, stepsSnap)
			if err != nil {
				return nil, fmt.Errorf("step %s: plugin arg %q: %w", step.ID, k, err)
			}
			renderedArgs[k] = rendered
		}

		// Search order: project-local, then user-global
		searchDirs := []string{".glitch/plugins"}
		if home, err := os.UserHomeDir(); err == nil {
			searchDirs = append(searchDirs, filepath.Join(home, ".config", "glitch", "plugins"))
		}

		for _, dir := range searchDirs {
			pluginDir := filepath.Join(dir, pc.Plugin)
			if _, err := os.Stat(pluginDir); err == nil {
				out, err := RunPluginSubcommand(dir, pc.Plugin, pc.Subcommand, renderedArgs, rctx.reg, RunOpts{Workspace: rctx.workspace})
				if err != nil {
					return nil, fmt.Errorf("step %s: %w", step.ID, err)
				}
				return &stepOutcome{output: out}, nil
			}
		}

		return nil, fmt.Errorf("step %s: plugin %q not found, searched: %s", step.ID, pc.Plugin, strings.Join(searchDirs, ", "))
	}

	return nil, fmt.Errorf("step %s: must have either 'run', 'llm', or 'save'", step.ID)
}

// GateResult holds the outcome of a single gate evaluation.
type GateResult struct {
	ID      string
	Passed  bool
	Detail  string
	Skipped bool
}

// VerificationReport is emitted when a phase exhausts its retry budget.
type VerificationReport struct {
	Phase    string
	Attempts int
	MaxRetry int
	Gates    []GateResult
}

// FormatReport returns a human-readable verification report.
func (vr *VerificationReport) FormatReport() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Phase %q FAILED (%d/%d attempts exhausted)\n\n", vr.Phase, vr.Attempts, vr.Attempts)
	b.WriteString("Gate results (final attempt):\n")
	for _, g := range vr.Gates {
		if g.Skipped {
			fmt.Fprintf(&b, "  %s: (skipped - prior gate failed)\n", g.ID)
		} else if g.Passed {
			fmt.Fprintf(&b, "  %s: PASS\n", g.ID)
		} else {
			detail := g.Detail
			if len(detail) > 200 {
				detail = detail[:200] + "..."
			}
			fmt.Fprintf(&b, "  %s: FAIL - %s\n", g.ID, detail)
		}
	}
	return b.String()
}

// executePhase runs all steps, then all gates. On gate failure, retries the
// phase up to p.Retries times. Returns nil report on success, non-nil on failure.
func executePhase(rctx *runCtx, p Phase) (*VerificationReport, error) {
	maxAttempts := 1 + p.Retries

	var lastGateResults []GateResult

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ui.PhaseStart(p.ID, attempt-1, p.Retries)

		// Run all work steps
		for _, step := range p.Steps {
			ui.PhaseStep(step.ID)
			outcome, err := executeStep(rctx.ctx, rctx, step)
			if err != nil {
				return nil, fmt.Errorf("phase %q step %q: %w", p.ID, step.ID, err)
			}
			rctx.mu.Lock()
			rctx.steps[step.ID] = outcome.output
			rctx.mu.Unlock()
		}

		// Flatten gates for result tracking (par blocks expand to children)
		var flatGates []Step
		for _, gate := range p.Gates {
			if gate.Form == "par" {
				flatGates = append(flatGates, gate.ParSteps...)
			} else {
				flatGates = append(flatGates, gate)
			}
		}
		lastGateResults = make([]GateResult, len(flatGates))
		allPassed := true
		flatIdx := 0

		for _, gate := range p.Gates {
			if gate.Form == "par" {
				// Run all par children concurrently
				ui.StepSDK("gates", "parallel")
				for _, child := range gate.ParSteps {
					ui.GateStart(child.ID)
				}
				_, execErr := executeStep(rctx.ctx, rctx, gate)
				if execErr != nil {
					for _, child := range gate.ParSteps {
						ui.GateFail(child.ID, execErr.Error())
						lastGateResults[flatIdx] = GateResult{ID: child.ID, Passed: false, Detail: execErr.Error()}
						flatIdx++
					}
					allPassed = false
					break
				}
				// Evaluate each par child gate
				for _, child := range gate.ParSteps {
					rctx.mu.Lock()
					output := rctx.steps[child.ID]
					rctx.mu.Unlock()
					passed, detail := evaluateGate(child, &stepOutcome{output: output}, nil)
					lastGateResults[flatIdx] = GateResult{ID: child.ID, Passed: passed, Detail: detail}
					if passed {
						ui.GatePass(child.ID)
					} else {
						ui.GateFail(child.ID, detail)
						allPassed = false
					}
					flatIdx++
				}
				if !allPassed {
					break
				}
			} else {
				// Sequential gate
				ui.GateStart(gate.ID)
				outcome, execErr := executeStep(rctx.ctx, rctx, gate)
				if execErr != nil {
					ui.GateFail(gate.ID, execErr.Error())
					rctx.mu.Lock()
					rctx.steps[gate.ID] = execErr.Error()
					rctx.mu.Unlock()
					lastGateResults[flatIdx] = GateResult{ID: gate.ID, Passed: false, Detail: execErr.Error()}
					allPassed = false
					flatIdx++
					break
				}
				rctx.mu.Lock()
				rctx.steps[gate.ID] = outcome.output
				rctx.mu.Unlock()
				passed, detail := evaluateGate(gate, outcome, nil)
				lastGateResults[flatIdx] = GateResult{ID: gate.ID, Passed: passed, Detail: detail}
				if passed {
					ui.GatePass(gate.ID)
				} else {
					ui.GateFail(gate.ID, detail)
					allPassed = false
					flatIdx++
					break
				}
				flatIdx++
			}
		}
		// Mark remaining gates as skipped
		for ; flatIdx < len(flatGates); flatIdx++ {
			lastGateResults[flatIdx] = GateResult{ID: flatGates[flatIdx].ID, Skipped: true}
		}

		if allPassed {
			return nil, nil
		}

		if attempt < maxAttempts {
			for _, gr := range lastGateResults {
				if !gr.Passed && !gr.Skipped {
					ui.GateRetry(gr.ID)
					break
				}
			}
		}
	}

	report := &VerificationReport{
		Phase:    p.ID,
		Attempts: maxAttempts,
		MaxRetry: p.Retries,
		Gates:    lastGateResults,
	}
	return report, fmt.Errorf("phase %q: all %d attempts exhausted\n%s", p.ID, maxAttempts, report.FormatReport())
}

// evaluateGate determines if a gate step passed or failed.
// Shell gates: execution error means failure, nil error means pass.
// LLM gates: parse output for PASS/FAIL verdict.
// Returns (passed, failureDetail).
func evaluateGate(step Step, outcome *stepOutcome, execErr error) (bool, string) {
	if execErr != nil {
		return false, execErr.Error()
	}

	if step.LLM != nil {
		upper := strings.ToUpper(strings.ReplaceAll(outcome.output, "*", ""))
		if strings.Contains(upper, "OVERALL: PASS") || strings.Contains(upper, "OVERALL PASS") {
			return true, ""
		}
		if strings.Contains(upper, "OVERALL: FAIL") || strings.Contains(upper, "OVERALL FAIL") {
			return false, outcome.output
		}
		var passed, failed int
		for _, line := range strings.Split(upper, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "OVERALL") {
				continue
			}
			hasPass := strings.Contains(line, "PASS")
			hasFail := strings.Contains(line, "FAIL")
			if hasPass && !hasFail {
				passed++
			} else if hasFail {
				failed++
			}
		}
		if failed > 0 {
			return false, outcome.output
		}
		if passed > 0 {
			return true, ""
		}
		return true, ""
	}

	return true, ""
}

// estimateCost approximates USD cost per token for known providers/models.
// Returns 0 for local/free models.
func estimateCost(prov, model string, tokensIn, tokensOut int) float64 {
	// Pricing per 1M tokens (input/output)
	type pricing struct{ in, out float64 }
	prices := map[string]pricing{
		"claude:sonnet":                  {3.0, 15.0},
		"claude:opus":                    {15.0, 75.0},
		"claude:haiku":                   {0.25, 1.25},
		"copilot:":                       {2.0, 8.0}, // estimate
		"openrouter:x-ai/grok-4.20":     {2.0, 10.0},
		"openrouter:google/gemma-4-31b-it:free": {0, 0},
	}

	key := prov + ":" + model
	if p, ok := prices[key]; ok {
		return (float64(tokensIn)*p.in + float64(tokensOut)*p.out) / 1_000_000
	}
	// Try provider-only match
	key = prov + ":"
	if p, ok := prices[key]; ok {
		return (float64(tokensIn)*p.in + float64(tokensOut)*p.out) / 1_000_000
	}
	return 0
}

// loadSkill resolves a skill name or path to its content.
//
// Resolution order:
//  1. Absolute or relative file path (if the string contains "/" or ends in ".md")
//  2. Skill name looked up in standard locations:
//     - .claude/skills/<name>/SKILL.md (project-local)
//     - ~/.config/glitch/skills/<name>/SKILL.md (user-global)
//     - skills/<name>/SKILL.md (gl1tch built-in)
func loadSkill(nameOrPath string) (string, error) {
	// Direct file path
	if strings.Contains(nameOrPath, "/") || strings.HasSuffix(nameOrPath, ".md") {
		data, err := os.ReadFile(nameOrPath)
		if err != nil {
			return "", fmt.Errorf("read skill file: %w", err)
		}
		return string(data), nil
	}

	// Search standard skill locations
	searchPaths := []string{
		filepath.Join(".claude", "skills", nameOrPath, "SKILL.md"),
		filepath.Join(".cursor", "skills", nameOrPath, "SKILL.md"),
	}

	// User-global location
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(home, ".config", "glitch", "skills", nameOrPath, "SKILL.md"),
		)
	}

	for _, p := range searchPaths {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("skill %q not found in: %s", nameOrPath, strings.Join(searchPaths, ", "))
}
