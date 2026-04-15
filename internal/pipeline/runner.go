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
	"text/template"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
)

// stepOutcome holds the result of executing a single step.
type stepOutcome struct {
	output     string
	tokensIn   int
	tokensOut  int
	cost       float64
	tier       int
	escalated  bool
	escChain   []int
	evalScores []int
	isLLM      bool
}

// runCtx bundles per-run state needed by executeStep and compound forms.
type runCtx struct {
	input            string
	params           map[string]string
	defaultModel     string
	reg              *provider.ProviderRegistry
	providerResolver provider.ResolverFunc
	tiers            []provider.TierConfig
	evalThreshold    int
	steps            map[string]string
	prevStepID       string
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

	rctx := &runCtx{
		input:            input,
		params:           params,
		defaultModel:     defaultModel,
		reg:              reg,
		providerResolver: providerResolver,
		tiers:            tiers,
		evalThreshold:    evalThreshold,
		steps:            steps,
	}

	for i, step := range w.Steps {
		if i > 0 {
			rctx.prevStepID = w.Steps[i-1].ID
		}

		outcome, err := executeStep(rctx, step)
		if err != nil {
			return nil, err
		}

		// Accumulate telemetry
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

	last := w.Steps[len(w.Steps)-1]
	return &Result{
		Workflow: w.Name,
		Output:   steps[last.ID],
		Steps:    steps,
	}, nil
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
func executeStep(rctx *runCtx, step Step) (*stepOutcome, error) {
	switch step.Form {
	case "cond":
		return executeCond(rctx, step)
	case "map":
		return executeMap(rctx, step)
	}

	// Determine timeout context
	ctx := context.Background()
	if step.Timeout != "" {
		dur, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return nil, fmt.Errorf("step %s: invalid timeout %q: %w", step.ID, step.Timeout, err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dur)
		defer cancel()
	}

	maxAttempts := 1 + step.Retry
	var lastErr error
	for attempt := range maxAttempts {
		if attempt > 0 {
			fmt.Printf("  > %s (retry %d/%d)\n", step.ID, attempt, step.Retry)
		}

		outcome, err := runSingleStep(ctx, rctx, step)
		if err == nil {
			// catch form: step succeeded, store output normally
			rctx.steps[step.ID] = outcome.output
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
		fmt.Printf("  > %s (fallback → %s)\n", step.ID, step.Fallback.ID)
		outcome, err := runSingleStep(context.Background(), rctx, *step.Fallback)
		if err != nil {
			return nil, fmt.Errorf("step %s fallback %s: %w", step.ID, step.Fallback.ID, err)
		}
		rctx.steps[step.Fallback.ID] = outcome.output
		rctx.steps[step.ID] = outcome.output // catch step ID also gets the fallback output
		return outcome, nil
	}

	return nil, lastErr
}

// executeCond evaluates predicate branches in order and runs the first matching step.
func executeCond(rctx *runCtx, step Step) (*stepOutcome, error) {
	for _, branch := range step.Branches {
		if branch.Pred == "else" {
			return executeStep(rctx, branch.Step)
		}
		// Render predicate template
		data := map[string]any{"input": rctx.input, "param": rctx.params}
		rendered, err := render(branch.Pred, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("cond %s: template: %w", step.ID, err)
		}
		cmd := exec.Command("sh", "-c", rendered)
		if err := cmd.Run(); err == nil {
			// Predicate succeeded (exit 0)
			return executeStep(rctx, branch.Step)
		}
	}
	// No branch matched — return empty outcome
	rctx.steps[step.ID] = ""
	return &stepOutcome{}, nil
}

// executeMap splits a prior step's output by newlines and runs the body step for each item.
func executeMap(rctx *runCtx, step Step) (*stepOutcome, error) {
	source, ok := rctx.steps[step.MapOver]
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

		outcome, err := executeStep(rctx, body)
		rctx.params = origParams
		if err != nil {
			return nil, fmt.Errorf("map %s item %d: %w", step.ID, idx, err)
		}
		outputs = append(outputs, outcome.output)
	}

	combined := strings.Join(outputs, "\n")
	rctx.steps[step.ID] = combined
	return &stepOutcome{output: combined}, nil
}

// runSingleStep executes one step (save, run, or llm) without control-flow wrappers.
func runSingleStep(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	data := map[string]any{
		"input": rctx.input,
		"param": rctx.params,
	}

	if step.Save != "" {
		rendered, err := render(step.Save, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		fmt.Printf("  > %s (save → %s)\n", step.ID, rendered)
		sourceStep := step.SaveStep
		if sourceStep == "" && rctx.prevStepID != "" {
			sourceStep = rctx.prevStepID
		}
		content := rctx.steps[sourceStep]
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
		rendered, err := render(step.Run, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		fmt.Printf("  > %s\n", step.ID)
		out, err := provider.RunShell(rendered)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.ID, err)
		}
		return &stepOutcome{output: out}, nil
	}

	if step.LLM != nil {
		rendered, err := render(step.LLM.Prompt, data, rctx.steps)
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

		fmt.Printf("  > %s\n", step.ID)

		prov := strings.ToLower(step.LLM.Provider)
		model := step.LLM.Model
		if model == "" {
			model = rctx.defaultModel
		}

		var out string
		var stepTier int
		var stepEscalated bool
		var stepEscalationChain []int
		var stepEvalScores []int
		var stepCost float64
		var stepTokensIn, stepTokensOut int

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

		return &stepOutcome{
			output:     out,
			tokensIn:   stepTokensIn,
			tokensOut:  stepTokensOut,
			cost:       stepCost,
			tier:       stepTier,
			escalated:  stepEscalated,
			escChain:   stepEscalationChain,
			evalScores: stepEvalScores,
			isLLM:      true,
		}, nil
	}

	// --- SDK forms ---

	if step.JsonPick != nil {
		from := rctx.steps[step.JsonPick.From]
		rendered, err := render(step.JsonPick.Expr, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		fmt.Printf("  > %s (json-pick)\n", step.ID)
		cmd := exec.CommandContext(ctx, "jq", rendered)
		cmd.Stdin = strings.NewReader(from)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("step %s: jq: %w", step.ID, err)
		}
		return &stepOutcome{output: strings.TrimSpace(string(out))}, nil
	}

	if step.Lines != "" {
		from := rctx.steps[step.Lines]
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
		fmt.Printf("  > %s (lines)\n", step.ID)
		return &stepOutcome{output: out}, nil
	}

	if len(step.Merge) > 0 {
		merged := make(map[string]any)
		for _, id := range step.Merge {
			var obj map[string]any
			if err := json.Unmarshal([]byte(rctx.steps[id]), &obj); err != nil {
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
		fmt.Printf("  > %s (merge)\n", step.ID)
		return &stepOutcome{output: string(out)}, nil
	}

	if step.HttpCall != nil {
		urlRendered, err := render(step.HttpCall.URL, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: url template: %w", step.ID, err)
		}
		var bodyReader io.Reader
		if step.HttpCall.Body != "" {
			bodyRendered, err := render(step.HttpCall.Body, data, rctx.steps)
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
			hv, err := render(v, data, rctx.steps)
			if err != nil {
				return nil, fmt.Errorf("step %s: header %q template: %w", step.ID, k, err)
			}
			req.Header.Set(k, hv)
		}
		fmt.Printf("  > %s (http-%s)\n", step.ID, strings.ToLower(method))
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
		rendered, err := render(step.ReadFile, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		fmt.Printf("  > %s (read-file)\n", step.ID)
		content, err := os.ReadFile(rendered)
		if err != nil {
			return nil, fmt.Errorf("step %s: read-file: %w", step.ID, err)
		}
		return &stepOutcome{output: string(content)}, nil
	}

	if step.WriteFile != nil {
		rendered, err := render(step.WriteFile.Path, data, rctx.steps)
		if err != nil {
			return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
		}
		content := rctx.steps[step.WriteFile.From]
		if err := os.MkdirAll(filepath.Dir(rendered), 0o755); err != nil {
			return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
		}
		if err := os.WriteFile(rendered, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("step %s: write-file: %w", step.ID, err)
		}
		fmt.Printf("  > %s (write-file)\n", step.ID)
		return &stepOutcome{output: rendered}, nil
	}

	if step.GlobPat != nil {
		pattern := step.GlobPat.Pattern
		if step.GlobPat.Dir != "" {
			dirRendered, err := render(step.GlobPat.Dir, data, rctx.steps)
			if err != nil {
				return nil, fmt.Errorf("step %s: dir template: %w", step.ID, err)
			}
			pattern = filepath.Join(dirRendered, pattern)
		}
		fmt.Printf("  > %s (glob)\n", step.ID)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("step %s: glob: %w", step.ID, err)
		}
		return &stepOutcome{output: strings.Join(matches, "\n")}, nil
	}

	if step.PluginCall != nil {
		fmt.Printf("  > %s (plugin %s %s)\n", step.ID, step.PluginCall.Plugin, step.PluginCall.Subcommand)
		pc := step.PluginCall

		// Search order: project-local, then user-global
		searchDirs := []string{".glitch/plugins"}
		if home, err := os.UserHomeDir(); err == nil {
			searchDirs = append(searchDirs, filepath.Join(home, ".config", "glitch", "plugins"))
		}

		for _, dir := range searchDirs {
			pluginDir := filepath.Join(dir, pc.Plugin)
			if _, err := os.Stat(pluginDir); err == nil {
				out, err := RunPluginSubcommand(dir, pc.Plugin, pc.Subcommand, pc.Args, rctx.reg)
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
