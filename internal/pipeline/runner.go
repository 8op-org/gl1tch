package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
)

// runCtx bundles per-run state needed by the evaluator path.
type runCtx struct {
	parentRunID     int64
	workflowsDir    string
	callStack       []string
	childRunCreator func(parentID int64, workflowName string) (int64, error)
	stepRecorder    func(rec StepRecord)
}

// Result holds the output of a completed workflow run.
type Result struct {
	Workflow string
	Output   string            // output of the last step
	Steps    map[string]string // all step outputs keyed by step ID
	// RunID is the DB row id associated with this run. The pipeline package
	// does not own store.RecordRun; when the caller pre-creates a row and
	// passes its id via RunOpts.ParentRunID, that same value is echoed back
	// here so downstream callers (batch, GUI) can correlate the run.
	// Zero means unknown / no row created.
	RunID int64
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
	WebSearchURL     string            // default SearXNG URL from workspace config
	Workspace        string            // resolved workspace name for ~workspace references
	Resources        map[string]map[string]string // resource name → field → value (from active workspace)

	// call-workflow support
	WorkflowsDir string   // directory to resolve call-workflow targets
	ParentRunID  int64    // if non-zero, this run is a child of this parent
	CallStack    []string // workflow names already on the call stack (cycle guard)

	// ChildRunCreator is called before starting a nested workflow via call-workflow.
	// The callback should create a child run row in the store and return its id,
	// which is then used as ParentRunID for the child invocation — giving correct
	// per-level parent linkage in multi-level call-workflow trees.
	// When nil, call-workflow falls back to grandparent chaining
	// (child's ParentRunID = rctx.parentRunID).
	ChildRunCreator func(parentID int64, workflowName string) (int64, error)

	// StepRecorder, when non-nil, receives one StepRecord per completed
	// top-level workflow step. Callers wire this to store.RecordStep so the
	// steps table gets populated. Compound-form sub-steps are not reported
	// here — only the items in Workflow.Steps / Workflow.Items.
	StepRecorder func(rec StepRecord)
}

// StepRecord is a lightweight per-step record emitted to RunOpts.StepRecorder
// on each completed workflow step. Field set mirrors store.StepRecord so
// callers can forward directly. RunID is omitted because the caller's closure
// already has the parent run id bound.
type StepRecord struct {
	StepID     string
	Prompt     string
	Output     string
	Model      string
	DurationMs int64
	Kind       string
	ExitStatus *int
	TokensIn   int64
	TokensOut  int64
	Artifacts  []string
}

// Run executes a workflow with the given input string.
// All workflow execution is handled by the Lisp evaluator.
func Run(w *Workflow, input string, defaultModel string, params map[string]string, reg *provider.ProviderRegistry, opts ...RunOpts) (*Result, error) {
	var providerResolver provider.ResolverFunc
	var tiers []provider.TierConfig
	var evalThreshold int
	if len(opts) > 0 {
		providerResolver = opts[0].ProviderResolver
		tiers = opts[0].Tiers
		evalThreshold = opts[0].EvalThreshold
	}
	if evalThreshold == 0 {
		evalThreshold = 4
	}

	var esURL string
	if len(opts) > 0 && opts[0].ESURL != "" {
		esURL = opts[0].ESURL
	}

	var webSearchURL string
	if len(opts) > 0 && opts[0].WebSearchURL != "" {
		webSearchURL = opts[0].WebSearchURL
	}

	var workspaceName string
	if len(opts) > 0 {
		workspaceName = opts[0].Workspace
	}

	var resources map[string]map[string]string
	if len(opts) > 0 {
		resources = opts[0].Resources
	}
	if resources == nil {
		resources = map[string]map[string]string{}
	}

	rctx := &runCtx{}
	if len(opts) > 0 {
		rctx.workflowsDir = opts[0].WorkflowsDir
		rctx.parentRunID = opts[0].ParentRunID
		rctx.callStack = opts[0].CallStack
		rctx.childRunCreator = opts[0].ChildRunCreator
		rctx.stepRecorder = opts[0].StepRecorder
	}

	if w.Source == nil {
		return nil, fmt.Errorf("workflow %q has no source — only .glitch sexpr workflows are supported", w.Name)
	}

	ev := NewEvaluator()
	ev.Input = input
	ev.Params = params
	ev.DefaultModel = defaultModel
	ev.Workspace = workspaceName
	ev.Resources = resources
	ev.ProviderReg = reg
	ev.ProviderResolver = providerResolver
	ev.Tiers = tiers
	ev.EvalThreshold = evalThreshold
	ev.ESURL = esURL
	ev.WebSearchURL = webSearchURL
	ev.WorkflowName = w.Name
	ev.WorkflowsDir = rctx.workflowsDir
	ev.StepRecorder = rctx.stepRecorder
	if len(opts) > 0 {
		ev.CallStack = opts[0].CallStack
	}

	// Pre-seed steps from SeedSteps
	if len(opts) > 0 && opts[0].SeedSteps != nil {
		for k, v := range opts[0].SeedSteps {
			ev.mu.Lock()
			ev.steps[k] = v
			ev.mu.Unlock()
		}
	}

	_, err := ev.RunSource(w.Source)
	if err != nil {
		return nil, err
	}

	evSteps := ev.Steps()

	// Find last step output
	var lastOutput string
	for _, s := range evSteps {
		lastOutput = s
	}

	return &Result{
		Workflow: w.Name,
		Output:   lastOutput,
		Steps:    evSteps,
		RunID:    rctx.parentRunID,
	}, nil
}

// PreRunSharedSteps executes only the (run ...) shell steps from a workflow and
// returns their outputs keyed by step ID. Used by batch to run data-gathering
// steps once and seed all variant runs with the results.
//
// With the evaluator handling all execution, this is now a no-op that returns
// empty — the evaluator path in Run() handles seed steps via SeedSteps.
func PreRunSharedSteps(w *Workflow, params map[string]string) (map[string]string, error) {
	return make(map[string]string), nil
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

// render interpolates sexpr-level unquote (~name, ~param.x, ~(form)) in tmpl
// against the given scope. steps is merged into scope before rendering.
func render(tmpl string, scope *Scope, steps map[string]string) (string, error) {
	if steps != nil {
		scope.SetSteps(steps)
	}
	return renderQuasi(tmpl, scope)
}

// scopeFromData translates the legacy data map (with "param", "input" keys)
// into a *Scope. Transitional helper for call sites that still build a
// map[string]any.
func scopeFromData(data map[string]any) *Scope {
	s := NewScope()
	if p, ok := data["param"].(map[string]string); ok {
		for k, v := range p {
			s.SetParam(k, v)
		}
		if item, ok := p["item"]; ok {
			idx := 0
			if ix, ok := p["item_index"]; ok {
				fmt.Sscanf(ix, "%d", &idx)
			}
			s.SetItem(item, idx)
		}
	}
	if in, ok := data["input"].(string); ok {
		s.SetInput(in)
	}
	if defs, ok := data["def"].(map[string]string); ok {
		for k, v := range defs {
			s.SetDef(k, v)
		}
	}
	if r, ok := data["resource"].(map[string]map[string]string); ok {
		s.SetResources(r)
	}
	return s
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


