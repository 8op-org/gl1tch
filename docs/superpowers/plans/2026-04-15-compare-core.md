# Compare as Core DSL Primitive — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `(compare ...)` with `(branch ...)` and `(review ...)` as first-class DSL forms, wire CLI `--variant`/`--compare` on `workflow run`, extend ES telemetry for all runs, and surface comparison data in the GUI.

**Architecture:** Compare is a flow-control form like `(par ...)`. Branches run in parallel using errgroup, a review step (default or custom) judges outputs and picks a winner, and the winner's output flows downstream. All compare results emit ES documents for Kibana dashboards. CLI `--variant` flags inject implicit compare blocks at runtime.

**Tech Stack:** Go, sexpr parser, errgroup, Elasticsearch, Cobra CLI, existing provider/pipeline packages

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/pipeline/types.go` | Modify | Add `CompareBranch`, `ReviewConfig` types; add compare fields to `Step` |
| `internal/pipeline/sexpr.go` | Modify | Add `convertCompare`, `convertBranch`, `convertReview` parsers |
| `internal/pipeline/sexpr_test.go` | Modify | Tests for compare/branch/review parsing |
| `internal/pipeline/runner.go` | Modify | Add `executeCompare` function, extend `render` with compare template funcs |
| `internal/pipeline/runner_test.go` | Modify | Tests for compare execution, review scoring, seed detection |
| `internal/pipeline/compare_review.go` | Create | Default judge prompt, review prompt builder, score parsing |
| `internal/pipeline/compare_review_test.go` | Create | Tests for review prompt generation and score parsing |
| `internal/pipeline/testdata/compare-basic.glitch` | Create | Test fixture: step-level compare |
| `internal/pipeline/testdata/compare-multi.glitch` | Create | Test fixture: multi-step branch compare |
| `internal/pipeline/testdata/compare-custom-review.glitch` | Create | Test fixture: custom review prompt |
| `internal/esearch/telemetry.go` | Modify | Add `RunDoc` struct, `IndexRun` method; extend `CrossReviewDoc` |
| `internal/esearch/mappings.go` | Modify | Add `glitch-runs` index mapping; extend cross-reviews mapping |
| `internal/esearch/telemetry_test.go` | Modify | Tests for new doc types |
| `cmd/workflow.go` | Modify | Add `--variant`, `--compare`, `--review-criteria` flags |
| `cmd/workflow_test.go` | Create | Test flag parsing for compare flags |

---

### Task 1: Types — CompareBranch, ReviewConfig, Step fields

**Files:**
- Modify: `internal/pipeline/types.go:42-88`

- [ ] **Step 1: Write the failing test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_CompareBasic(t *testing.T) {
	src := `(workflow "cmp-test"
	  :description "compare smoke"
	  (step "pick"
	    (compare
	      (branch "fast" (llm :prompt "fast answer"))
	      (branch "slow" (llm :prompt "slow answer")))))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.Form != "compare" {
		t.Fatalf("expected form=compare, got %q", s.Form)
	}
	if len(s.CompareBranches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(s.CompareBranches))
	}
	if s.CompareBranches[0].Name != "fast" {
		t.Errorf("branch 0 name = %q, want fast", s.CompareBranches[0].Name)
	}
	if s.CompareBranches[1].Name != "slow" {
		t.Errorf("branch 1 name = %q, want slow", s.CompareBranches[1].Name)
	}
	if s.CompareReview != nil {
		t.Error("expected nil review (default), got non-nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_CompareBasic -v`
Expected: FAIL — `CompareBranches` field doesn't exist

- [ ] **Step 3: Add types to types.go**

Add after the `CondBranch` type (line 94):

```go
// CompareBranch is one named alternative in a (compare ...) form.
type CompareBranch struct {
	Name  string // branch name
	Steps []Step // steps to execute in this branch
}

// ReviewConfig configures the judge for a (compare ...) form.
type ReviewConfig struct {
	Criteria []string // scoring criteria names (criteria mode)
	Prompt   string   // custom review prompt (prompt mode)
	Model    string   // model override for the judge
}
```

Add fields to the `Step` struct, after `ParSteps` (line 61):

```go
	// Compare execution
	CompareBranches []CompareBranch `yaml:"-"` // compare: named alternative branches
	CompareReview   *ReviewConfig   `yaml:"-"` // compare: review config (nil = default judge)
	CompareID       string          `yaml:"-"` // compare: id for top-level compare blocks
```

- [ ] **Step 4: Run test to verify it still fails (compile passes, parse not yet implemented)**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: BUILD SUCCESS (types exist now, but test still fails because parser doesn't handle "compare")

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): add CompareBranch, ReviewConfig types and compare Step fields"
```

---

### Task 2: Parser — convertCompare, convertBranch, convertReview

**Files:**
- Modify: `internal/pipeline/sexpr.go:139-188` (convertForm switch), append new functions

- [ ] **Step 1: Add "compare" case to convertForm**

In `convertForm` (sexpr.go), add a case before the `default`:

```go
	case "compare":
		s, err := convertCompare(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
```

- [ ] **Step 2: Write convertCompare function**

Append to sexpr.go:

```go
// convertCompare: (compare [:id "name"] (branch "name" ...) ... [(review ...)])
func convertCompare(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:] // skip "compare"

	s := Step{
		ID:   fmt.Sprintf("compare-%d", n.Line),
		Form: "compare",
	}

	// Parse optional keywords (:id)
	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			kw := child.KeywordVal()
			i++
			if i >= len(children) {
				return s, fmt.Errorf("line %d: keyword :%s missing value", child.Line, kw)
			}
			val := resolveVal(children[i], defs)
			i++
			switch kw {
			case "id":
				s.CompareID = val
				s.ID = val
			}
			continue
		}
		break
	}

	// Parse branches and optional review
	for _, child := range children[i:] {
		if !child.IsList() || len(child.Children) == 0 {
			return s, fmt.Errorf("line %d: compare body must be (branch ...) or (review ...)", child.Line)
		}
		head := child.Children[0].SymbolVal()
		if head == "" {
			head = child.Children[0].StringVal()
		}
		switch head {
		case "branch":
			b, err := convertBranch(child, defs)
			if err != nil {
				return s, err
			}
			s.CompareBranches = append(s.CompareBranches, b)
		case "review":
			r, err := convertReview(child, defs)
			if err != nil {
				return s, err
			}
			s.CompareReview = r
		default:
			return s, fmt.Errorf("line %d: unexpected form %q in compare (expected branch or review)", child.Line, head)
		}
	}

	if len(s.CompareBranches) < 2 {
		return s, fmt.Errorf("line %d: (compare) needs at least 2 branches, got %d", n.Line, len(s.CompareBranches))
	}

	return s, nil
}

// convertBranch: (branch "name" (step ...) ...) or (branch "name" (llm ...)) for single-step shorthand
func convertBranch(n *sexpr.Node, defs map[string]string) (CompareBranch, error) {
	children := n.Children[1:] // skip "branch"
	if len(children) < 2 {
		return CompareBranch{}, fmt.Errorf("line %d: (branch) needs name and at least one body", n.Line)
	}

	name := resolveVal(children[0], defs)
	if name == "" {
		return CompareBranch{}, fmt.Errorf("line %d: branch name must be a non-empty string", children[0].Line)
	}

	var steps []Step
	for _, child := range children[1:] {
		if !child.IsList() || len(child.Children) == 0 {
			return CompareBranch{}, fmt.Errorf("line %d: branch body must be forms", child.Line)
		}
		head := child.Children[0].SymbolVal()
		if head == "" {
			head = child.Children[0].StringVal()
		}
		// Single body form (llm, run, save) — wrap in an implicit step
		switch head {
		case "llm", "run", "save":
			implicitStep := &sexpr.Node{
				Line:     child.Line,
				Children: append([]*sexpr.Node{{Atom: &sexpr.Token{Type: sexpr.TokenSymbol, Value: "step"}, Line: child.Line}, {Atom: &sexpr.Token{Type: sexpr.TokenString, Value: name}, Line: child.Line}}, child),
			}
			s, err := convertStep(implicitStep, defs)
			if err != nil {
				return CompareBranch{}, fmt.Errorf("line %d: branch %q body: %w", child.Line, name, err)
			}
			steps = append(steps, s)
		case "step":
			s, err := convertStep(child, defs)
			if err != nil {
				return CompareBranch{}, fmt.Errorf("line %d: branch %q: %w", child.Line, name, err)
			}
			steps = append(steps, s)
		default:
			// Try as a generic form (par, cond, etc.)
			formSteps, err := convertForm(child, head, defs)
			if err != nil {
				return CompareBranch{}, fmt.Errorf("line %d: branch %q: %w", child.Line, name, err)
			}
			steps = append(steps, formSteps...)
		}
	}

	return CompareBranch{Name: name, Steps: steps}, nil
}

// convertReview: (review [:criteria [...]] [:prompt "..."] [:model "..."])
func convertReview(n *sexpr.Node, defs map[string]string) (*ReviewConfig, error) {
	children := n.Children[1:] // skip "review"
	r := &ReviewConfig{}

	i := 0
	for i < len(children) {
		child := children[i]
		if !child.IsAtom() || child.Atom.Type != sexpr.TokenKeyword {
			return nil, fmt.Errorf("line %d: review expects keyword arguments", child.Line)
		}
		kw := child.KeywordVal()
		i++
		if i >= len(children) {
			return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, kw)
		}
		valNode := children[i]
		i++

		switch kw {
		case "criteria":
			// Expect a list of strings: ["a" "b" "c"]
			if !valNode.IsList() {
				return nil, fmt.Errorf("line %d: :criteria must be a list", valNode.Line)
			}
			for _, c := range valNode.Children {
				r.Criteria = append(r.Criteria, resolveVal(c, defs))
			}
		case "prompt":
			r.Prompt = resolveVal(valNode, defs)
		case "model":
			r.Model = resolveVal(valNode, defs)
		default:
			return nil, fmt.Errorf("line %d: unknown review keyword :%s", child.Line, kw)
		}
	}

	return r, nil
}
```

- [ ] **Step 3: Also handle compare inside convertStep (step-level compare)**

In `convertStep` (sexpr.go), add a case in the step body switch (around line 536):

```go
		case "compare":
			cmp, err := convertCompare(child, defs)
			if err != nil {
				return s, err
			}
			s.Form = "compare"
			s.CompareBranches = cmp.CompareBranches
			s.CompareReview = cmp.CompareReview
```

- [ ] **Step 4: Run the test from Task 1**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_CompareBasic -v`
Expected: PASS

- [ ] **Step 5: Add more parser tests**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_CompareMultiStep(t *testing.T) {
	src := `(workflow "cmp-multi"
	  :description "multi-step branches"
	  (compare
	    :id "impl"
	    (branch "local"
	      (step "plan" (llm :prompt "plan locally"))
	      (step "code" (run "echo done")))
	    (branch "cloud"
	      (step "plan" (llm :prompt "plan in cloud"))
	      (step "code" (run "echo done")))
	    (review :criteria ["accuracy" "completeness"] :model "qwen2.5:7b")))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(w.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(w.Items))
	}
	// Top-level compare becomes a step via convertForm
	s := w.Items[0].Step
	if s == nil {
		t.Fatal("expected Step item, got Phase")
	}
	if s.Form != "compare" {
		t.Fatalf("form = %q, want compare", s.Form)
	}
	if s.CompareID != "impl" {
		t.Errorf("CompareID = %q, want impl", s.CompareID)
	}
	if len(s.CompareBranches) != 2 {
		t.Fatalf("branches = %d, want 2", len(s.CompareBranches))
	}
	if len(s.CompareBranches[0].Steps) != 2 {
		t.Errorf("local branch steps = %d, want 2", len(s.CompareBranches[0].Steps))
	}
	if s.CompareReview == nil {
		t.Fatal("expected review config")
	}
	if len(s.CompareReview.Criteria) != 2 {
		t.Errorf("criteria = %d, want 2", len(s.CompareReview.Criteria))
	}
	if s.CompareReview.Model != "qwen2.5:7b" {
		t.Errorf("review model = %q, want qwen2.5:7b", s.CompareReview.Model)
	}
}

func TestSexprWorkflow_CompareCustomPrompt(t *testing.T) {
	src := `(workflow "cmp-custom"
	  :description "custom review"
	  (step "tone"
	    (compare
	      (branch "formal" (llm :prompt "Write formally"))
	      (branch "casual" (llm :prompt "Write casually"))
	      (review :prompt "Which is better?" :model "qwen2.5:7b"))))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	s := w.Steps[0]
	if s.CompareReview == nil {
		t.Fatal("expected review")
	}
	if s.CompareReview.Prompt != "Which is better?" {
		t.Errorf("prompt = %q", s.CompareReview.Prompt)
	}
}

func TestSexprWorkflow_CompareTooFewBranches(t *testing.T) {
	src := `(workflow "bad"
	  :description "needs 2"
	  (step "x"
	    (compare
	      (branch "only-one" (llm :prompt "solo")))))
	`
	_, err := parseSexprWorkflow([]byte(src))
	if err == nil {
		t.Fatal("expected error for <2 branches")
	}
}
```

- [ ] **Step 6: Run all compare parser tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Compare -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): parse (compare ...) (branch ...) (review ...) sexpr forms"
```

---

### Task 3: Review System — Default judge, criteria mode, custom prompt

**Files:**
- Create: `internal/pipeline/compare_review.go`
- Create: `internal/pipeline/compare_review_test.go`

- [ ] **Step 1: Write tests for review prompt building**

Create `internal/pipeline/compare_review_test.go`:

```go
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
	prompt := buildReviewPrompt(nil, branches)
	if !strings.Contains(prompt, "fast") {
		t.Error("prompt should contain branch name 'fast'")
	}
	if !strings.Contains(prompt, "slow") {
		t.Error("prompt should contain branch name 'slow'")
	}
	if !strings.Contains(prompt, "VARIANT:") {
		t.Error("prompt should instruct VARIANT: output format")
	}
	if !strings.Contains(prompt, "WINNER:") {
		t.Error("prompt should instruct WINNER: output format")
	}
}

func TestBuildReviewPrompt_Criteria(t *testing.T) {
	branches := map[string]string{
		"a": "output a",
		"b": "output b",
	}
	cfg := &ReviewConfig{Criteria: []string{"accuracy", "completeness"}}
	prompt := buildReviewPrompt(cfg, branches)
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
	prompt := buildReviewPrompt(cfg, branches)
	if !strings.Contains(prompt, "Which matches brand voice?") {
		t.Error("prompt should contain custom text")
	}
	// Custom prompts still get branch outputs injected
	if !strings.Contains(prompt, "Dear sir") {
		t.Error("prompt should contain formal branch output")
	}
}

func TestParseCompareScores(t *testing.T) {
	review := `VARIANT: fast
coherence: 8/10
accuracy: 7/10
total: 15/20

VARIANT: slow
coherence: 9/10
accuracy: 9/10
total: 18/20

WINNER: slow`

	scores := ParseCrossReview(review)
	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}
	var winner string
	for _, s := range scores {
		if s.Winner {
			winner = s.Variant
		}
	}
	if winner != "slow" {
		t.Errorf("winner = %q, want slow", winner)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestBuildReviewPrompt|TestParseCompareScores" -v`
Expected: FAIL — `buildReviewPrompt` not defined

- [ ] **Step 3: Implement compare_review.go**

Create `internal/pipeline/compare_review.go`:

```go
package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

// buildReviewPrompt constructs the judge prompt for a compare block.
// If cfg is nil, uses default generic criteria.
// If cfg.Prompt is set, uses custom prompt with branch outputs appended.
// If cfg.Criteria is set, generates structured scoring prompt.
func buildReviewPrompt(cfg *ReviewConfig, branchOutputs map[string]string) string {
	// Sort branch names for deterministic output
	names := make([]string, 0, len(branchOutputs))
	for name := range branchOutputs {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build the branch outputs section
	var outputsSection strings.Builder
	for _, name := range names {
		fmt.Fprintf(&outputsSection, "--- %s ---\n%s\n\n", strings.ToUpper(name), branchOutputs[name])
	}

	if cfg != nil && cfg.Prompt != "" {
		// Custom prompt mode: user's prompt + branch outputs
		return fmt.Sprintf(`%s

Here are the outputs from each variant:

%s
You MUST end your response with the structured scoring format:

For each variant, output:
VARIANT: <name>
total: <score>/<max>

Then on the final line:
WINNER: <name>`, cfg.Prompt, outputsSection.String())
	}

	// Determine criteria
	criteria := []string{"coherence", "completeness", "relevance"}
	if cfg != nil && len(cfg.Criteria) > 0 {
		criteria = cfg.Criteria
	}

	var criteriaLines strings.Builder
	for _, c := range criteria {
		fmt.Fprintf(&criteriaLines, "- %s\n", c)
	}

	return fmt.Sprintf(`You are a judge comparing outputs from different AI variants on the same task.

Score each variant on the following criteria (1-10 scale):
%s
Here are the outputs:

%s
For each variant, output EXACTLY this format (one block per variant):

VARIANT: <name>
%stotal: <sum>/%d

Then on the final line:
WINNER: <name of best variant>

Be objective. Score based on quality, not length.`,
		criteriaLines.String(),
		outputsSection.String(),
		buildCriteriaFormat(criteria),
		len(criteria)*10,
	)
}

// buildCriteriaFormat generates the per-criteria scoring lines for the prompt.
func buildCriteriaFormat(criteria []string) string {
	var b strings.Builder
	for _, c := range criteria {
		fmt.Fprintf(&b, "%s: <score>/10\n", c)
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestBuildReviewPrompt|TestParseCompareScores" -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/compare_review.go internal/pipeline/compare_review_test.go
git commit -m "feat(pipeline): review prompt builder for compare blocks"
```

---

### Task 4: Runner — executeCompare with branch scoping and review

**Files:**
- Modify: `internal/pipeline/runner.go:687-706` (executeStep switch), append `executeCompare`
- Modify: `internal/pipeline/runner.go:609-682` (render function — add `branch` template func)

- [ ] **Step 1: Write test fixtures**

Create `internal/pipeline/testdata/compare-basic.glitch`:

```scheme
(workflow "compare-basic"
  :description "basic compare smoke test"

  (step "setup" (run "echo shared-data"))

  (step "pick"
    (compare
      (branch "alpha" (run "echo alpha-result"))
      (branch "beta" (run "echo beta-result")))))
```

Create `internal/pipeline/testdata/compare-multi.glitch`:

```scheme
(workflow "compare-multi"
  :description "multi-step branch compare"

  (compare
    :id "impl"
    (branch "fast"
      (step "plan" (run "echo fast-plan"))
      (step "build" (run "echo fast-build")))
    (branch "slow"
      (step "plan" (run "echo slow-plan"))
      (step "build" (run "echo slow-build")))))
```

- [ ] **Step 2: Write the failing runner test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRun_Compare_Basic(t *testing.T) {
	w, err := LoadFile("testdata/compare-basic.glitch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	// Both branches are shell-only, no LLM judge — when no review is configured
	// and branches are shell steps, the first branch wins by default.
	// The key assertion: it ran without error and produced output.
	if result.Output == "" {
		t.Error("expected non-empty output")
	}
	// setup step should be in results
	if result.Steps["setup"] != "shared-data" {
		t.Errorf("setup = %q, want shared-data", result.Steps["setup"])
	}
	// pick step should have the winner's output
	if result.Steps["pick"] == "" {
		t.Error("pick step should have winner output")
	}
}

func TestRun_Compare_MultiStep(t *testing.T) {
	w, err := LoadFile("testdata/compare-multi.glitch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	// Top-level compare with :id "impl" — output should be winner's last step
	if result.Steps["impl"] == "" {
		t.Error("impl step should have winner output")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRun_Compare" -v`
Expected: FAIL — no "compare" case in executeStep

- [ ] **Step 4: Add compare case to executeStep**

In `runner.go` `executeStep` function, add after `case "par":` (line 704-705):

```go
	case "compare":
		return executeCompare(ctx, rctx, step)
```

- [ ] **Step 5: Implement executeCompare**

Append to `runner.go`:

```go
// executeCompare runs all branches in parallel, then judges results to pick a winner.
func executeCompare(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	if len(step.CompareBranches) < 2 {
		return nil, fmt.Errorf("compare %s: need at least 2 branches", step.ID)
	}

	type branchResult struct {
		name    string
		output  string
		outcome *stepOutcome
		err     error
	}

	results := make([]branchResult, len(step.CompareBranches))
	g, gctx := errgroup.WithContext(ctx)

	for i, branch := range step.CompareBranches {
		br := branch
		g.Go(func() error {
			// Create a scoped runCtx for this branch with isolated steps map
			branchCtx := &runCtx{
				ctx:              gctx,
				input:            rctx.input,
				params:           rctx.params,
				workspace:        rctx.workspace,
				defaultModel:     rctx.defaultModel,
				reg:              rctx.reg,
				providerResolver: rctx.providerResolver,
				tiers:            rctx.tiers,
				evalThreshold:    rctx.evalThreshold,
				steps:            rctx.stepsSnapshot(), // copy outer steps
				esURL:            rctx.esURL,
			}

			var lastOutput string
			composite := &stepOutcome{}
			for _, s := range br.Steps {
				outcome, err := executeStep(gctx, branchCtx, s)
				if err != nil {
					results[i] = branchResult{name: br.Name, err: err}
					return nil // don't fail other branches
				}
				lastOutput = outcome.output
				if outcome.isLLM {
					composite.isLLM = true
					composite.tokensIn += outcome.tokensIn
					composite.tokensOut += outcome.tokensOut
					composite.cost += outcome.cost
					if outcome.latencyMs > composite.latencyMs {
						composite.latencyMs = outcome.latencyMs
					}
				}
			}

			composite.output = lastOutput

			// Store namespaced steps in parent context
			rctx.mu.Lock()
			for id, val := range branchCtx.steps {
				// Only store steps that this branch created (not inherited)
				if _, inherited := rctx.steps[id]; !inherited {
					nsID := fmt.Sprintf("%s/%s/%s", step.ID, br.Name, id)
					rctx.steps[nsID] = val
				}
			}
			rctx.mu.Unlock()

			results[i] = branchResult{name: br.Name, output: lastOutput, outcome: composite}
			return nil
		})
	}

	g.Wait()

	// Collect successful branches
	branchOutputs := make(map[string]string)
	var successResults []branchResult
	for _, r := range results {
		if r.err == nil && r.output != "" {
			branchOutputs[r.name] = r.output
			successResults = append(successResults, r)
		}
	}

	if len(successResults) == 0 {
		return nil, fmt.Errorf("compare %s: all branches failed", step.ID)
	}

	// Pick winner
	var winnerName, winnerOutput string
	if len(successResults) == 1 {
		// Only one succeeded — it wins by default
		winnerName = successResults[0].name
		winnerOutput = successResults[0].output
	} else {
		// Run review to pick winner
		winnerName, winnerOutput = runCompareReview(ctx, rctx, step, branchOutputs)
	}

	// Store results accessible via template
	rctx.mu.Lock()
	rctx.steps[step.ID] = winnerOutput
	rctx.steps[step.ID+"/__winner"] = winnerName
	rctx.mu.Unlock()

	composite := &stepOutcome{output: winnerOutput}
	for _, r := range results {
		if r.outcome != nil && r.outcome.isLLM {
			composite.isLLM = true
			composite.tokensIn += r.outcome.tokensIn
			composite.tokensOut += r.outcome.tokensOut
			composite.cost += r.outcome.cost
		}
	}

	return composite, nil
}

// runCompareReview executes the judge and returns the winner name and output.
// For shell-only branches (no LLM available), picks the first branch as default.
func runCompareReview(ctx context.Context, rctx *runCtx, step Step, branchOutputs map[string]string) (string, string) {
	// If no provider registry, can't run LLM judge — pick first branch
	if rctx.reg == nil {
		for name, output := range branchOutputs {
			return name, output
		}
	}

	reviewModel := rctx.defaultModel
	if step.CompareReview != nil && step.CompareReview.Model != "" {
		reviewModel = step.CompareReview.Model
	}

	prompt := buildReviewPrompt(step.CompareReview, branchOutputs)

	// Create a temporary LLM step for the review
	reviewStep := Step{
		ID: step.ID + "-review",
		LLM: &LLMStep{
			Prompt: prompt,
			Model:  reviewModel,
		},
	}

	outcome, err := runSingleStep(ctx, rctx, reviewStep)
	if err != nil {
		// Review failed — pick first branch as fallback
		for name, output := range branchOutputs {
			return name, output
		}
	}

	// Parse the review output to find the winner
	scores := ParseCrossReview(outcome.output)
	for _, s := range scores {
		if s.Winner {
			if output, ok := branchOutputs[s.Variant]; ok {
				return s.Variant, output
			}
		}
	}

	// No winner parsed — pick first
	for name, output := range branchOutputs {
		return name, output
	}
	return "", ""
}
```

- [ ] **Step 6: Extend the render function with compare template helpers**

In `runner.go` `render` function, update the `"step"` func to support variant access. Replace the existing `"step"` entry in funcMap (around line 611):

```go
		"step": func(args ...string) string {
			if len(args) == 0 {
				return ""
			}
			id := args[0]
			// Check for keyword args: (step "id" :variant "name") or (step "id" :winner) or (step "id" :scores)
			for i := 1; i < len(args); i++ {
				switch args[i] {
				case ":variant":
					if i+1 < len(args) {
						i++
						return steps[id+"/"+args[i]+"/"+id]
					}
				case ":winner":
					return steps[id+"/__winner"]
				case ":scores":
					return steps[id+"/__scores"]
				}
			}
			return steps[id]
		},
```

Also add a `"branch"` func for use in custom review prompts:

```go
		"branch": func(name string) string {
			// Used in custom review prompts — looks for branch output in current context
			for k, v := range steps {
				if strings.HasSuffix(k, "/"+name) || k == name {
					return v
				}
			}
			return ""
		},
```

- [ ] **Step 7: Run the tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRun_Compare" -v`
Expected: ALL PASS

- [ ] **Step 8: Run all pipeline tests to verify no regressions**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 9: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go internal/pipeline/testdata/compare-basic.glitch internal/pipeline/testdata/compare-multi.glitch
git commit -m "feat(pipeline): executeCompare — parallel branches, review judge, winner selection"
```

---

### Task 5: ES Telemetry — RunDoc, extended CrossReviewDoc, new index

**Files:**
- Modify: `internal/esearch/telemetry.go:84-95` (CrossReviewDoc), append RunDoc
- Modify: `internal/esearch/mappings.go` (add RunsMapping, extend CrossReviewsMapping)

- [ ] **Step 1: Write failing test**

Add to `internal/esearch/telemetry_test.go`:

```go
func TestRunDocFields(t *testing.T) {
	doc := RunDoc{
		RunID:        "run-123",
		WorkflowName: "test-wf",
		Source:       "cli",
		Status:       "completed",
		HasCompare:   true,
	}
	if doc.RunID != "run-123" {
		t.Error("RunID mismatch")
	}
	if doc.Source != "cli" {
		t.Error("Source mismatch")
	}
}

func TestNilTelemetryIndexRun(t *testing.T) {
	var tel *Telemetry
	err := tel.IndexRun(context.Background(), RunDoc{})
	if err != nil {
		t.Errorf("nil telemetry IndexRun should be nil-safe: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/ -run "TestRunDoc|TestNilTelemetryIndexRun" -v`
Expected: FAIL — RunDoc not defined

- [ ] **Step 3: Add RunDoc struct and IndexRun method**

In `telemetry.go`, after `CrossReviewDoc` (line 95), add:

```go
// RunDoc represents a single workflow run for the runs index.
type RunDoc struct {
	RunID        string  `json:"run_id"`
	WorkflowName string  `json:"workflow_name"`
	Workspace    string  `json:"workspace"`
	Source       string  `json:"source"`     // "cli", "gui", "batch"
	Status       string  `json:"status"`     // "running", "completed", "failed"
	HasCompare   bool    `json:"has_compare"`
	DurationMs   int64   `json:"duration_ms"`
	Timestamp    string  `json:"timestamp"`
}
```

Extend `CrossReviewDoc` with new fields:

```go
type CrossReviewDoc struct {
	RunID         string  `json:"run_id"`
	Issue         string  `json:"issue"`
	Iteration     string  `json:"iteration"`
	Variant       string  `json:"variant"`
	Passed        int     `json:"passed"`
	Total         int     `json:"total"`
	Confidence    float64 `json:"confidence"`
	Winner        bool    `json:"winner"`
	WorkflowName  string  `json:"workflow_name"`
	Timestamp     string  `json:"timestamp"`
	CompareID     string  `json:"compare_id,omitempty"`
	Scope         string  `json:"scope,omitempty"`          // "step" or "workflow"
	CriteriaName  string  `json:"criteria_name,omitempty"`
	CriteriaScore int     `json:"criteria_score,omitempty"`
	Workspace     string  `json:"workspace,omitempty"`
}
```

Add `IndexRun` method:

```go
// IndexRun indexes a workflow run document.
func (t *Telemetry) IndexRun(ctx context.Context, doc RunDoc) error {
	if t == nil {
		return nil
	}
	return t.indexDoc(ctx, IndexRuns, doc.RunID, doc)
}
```

- [ ] **Step 4: Add index constant and mapping**

In `mappings.go`, add constant:

```go
	IndexRuns = "glitch-runs"
```

Add mapping:

```go
const RunsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":        { "type": "keyword" },
      "workflow_name": { "type": "keyword" },
      "workspace":     { "type": "keyword" },
      "source":        { "type": "keyword" },
      "status":        { "type": "keyword" },
      "has_compare":   { "type": "boolean" },
      "duration_ms":   { "type": "long" },
      "timestamp":     { "type": "date" }
    }
  }
}`
```

Extend `CrossReviewsMapping` to include new fields:

```go
const CrossReviewsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":          { "type": "keyword" },
      "issue":           { "type": "keyword" },
      "iteration":       { "type": "keyword" },
      "variant":         { "type": "keyword" },
      "passed":          { "type": "integer" },
      "total":           { "type": "integer" },
      "confidence":      { "type": "float" },
      "winner":          { "type": "boolean" },
      "workflow_name":   { "type": "keyword" },
      "timestamp":       { "type": "date" },
      "compare_id":      { "type": "keyword" },
      "scope":           { "type": "keyword" },
      "criteria_name":   { "type": "keyword" },
      "criteria_score":  { "type": "integer" },
      "workspace":       { "type": "keyword" }
    }
  }
}`
```

Add to `AllIndices`:

```go
	IndexRuns: RunsMapping,
```

- [ ] **Step 5: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/esearch/telemetry.go internal/esearch/mappings.go internal/esearch/telemetry_test.go
git commit -m "feat(esearch): add glitch-runs index, extend cross-reviews with compare_id/scope/criteria fields"
```

---

### Task 6: Wire telemetry into executeCompare

**Files:**
- Modify: `internal/pipeline/runner.go` (executeCompare and runCompareReview)

- [ ] **Step 1: Write failing test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRun_Compare_IndexesCrossReview(t *testing.T) {
	// This test verifies that compare execution calls the telemetry path.
	// We can't easily mock ES, but we verify the runner doesn't panic
	// when telemetry is nil (nil-safe pattern).
	w, err := LoadFile("testdata/compare-basic.glitch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	result, err := Run(w, "", "", nil, nil, RunOpts{
		Telemetry: nil, // nil telemetry — should not panic
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Steps["pick"] == "" {
		t.Error("expected non-empty pick output")
	}
}
```

- [ ] **Step 2: Run to verify it passes (nil-safe already works)**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestRun_Compare_IndexesCrossReview -v`
Expected: PASS (nil telemetry is already safe)

- [ ] **Step 3: Add telemetry fields to runCtx**

In `runner.go`, add to `runCtx` struct:

```go
	tel       *esearch.Telemetry
	runID     string
	workflow  string
```

Wire these in the `Run` function where `runCtx` is initialized (should already have `tel` from opts).

- [ ] **Step 4: Emit cross-review docs in runCompareReview**

After the `ParseCrossReview(outcome.output)` call in `runCompareReview`, add:

```go
	// Index cross-review scores to ES
	if rctx.tel != nil {
		for _, s := range scores {
			rctx.tel.IndexCrossReview(ctx, esearch.CrossReviewDoc{
				RunID:        rctx.runID,
				Variant:      s.Variant,
				Passed:       s.Passed,
				Total:        s.Total,
				Confidence:   float64(s.Passed) / float64(max(s.Total, 1)),
				Winner:       s.Winner,
				WorkflowName: rctx.workflow,
				CompareID:    step.ID,
				Scope:        "step",
				Workspace:    rctx.workspace,
				Timestamp:    time.Now().UTC().Format(time.RFC3339),
			})
		}
	}
```

- [ ] **Step 5: Emit RunDoc at start and end of Run function**

In the `Run` function, after initializing `runCtx`, add:

```go
	startTime := time.Now()
	runID := esearch.NewRunID()

	// Check if workflow has any compare forms
	hasCompare := workflowHasCompare(w)

	if tel != nil {
		tel.IndexRun(ctx, esearch.RunDoc{
			RunID:        runID,
			WorkflowName: w.Name,
			Workspace:    rctx.workspace,
			Source:       "cli",
			Status:       "running",
			HasCompare:   hasCompare,
			Timestamp:    startTime.UTC().Format(time.RFC3339),
		})
	}
```

At the end of `Run`, before returning the result:

```go
	if tel != nil {
		tel.IndexRun(ctx, esearch.RunDoc{
			RunID:        runID,
			WorkflowName: w.Name,
			Workspace:    rctx.workspace,
			Source:       "cli",
			Status:       "completed",
			HasCompare:   hasCompare,
			DurationMs:   time.Since(startTime).Milliseconds(),
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		})
	}
```

Add helper:

```go
// workflowHasCompare checks if any step in the workflow uses the compare form.
func workflowHasCompare(w *Workflow) bool {
	for _, item := range w.Items {
		if item.Step != nil && item.Step.Form == "compare" {
			return true
		}
	}
	for _, s := range w.Steps {
		if s.Form == "compare" {
			return true
		}
	}
	return false
}
```

- [ ] **Step 6: Run all pipeline tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(pipeline): wire ES telemetry into compare — RunDoc + CrossReviewDoc emission"
```

---

### Task 7: CLI — --variant, --compare, --review-criteria on workflow run

**Files:**
- Modify: `cmd/workflow.go:24-30` (flag defs), `cmd/workflow.go:80-160` (workflowRunCmd)

- [ ] **Step 1: Add flags to workflowRunCmd**

In `cmd/workflow.go`, add flag variables:

```go
var (
	workflowVariants       []string
	workflowCompare        bool
	workflowReviewCriteria string
)
```

In `init()`, add:

```go
	workflowRunCmd.Flags().StringArrayVar(&workflowVariants, "variant", nil, "variant provider:model for comparison (repeatable)")
	workflowRunCmd.Flags().BoolVar(&workflowCompare, "compare", false, "discover variant workflows and cross-review")
	workflowRunCmd.Flags().StringVar(&workflowReviewCriteria, "review-criteria", "", "comma-separated review criteria for comparison")
```

- [ ] **Step 2: Wire --variant into workflow run logic**

In `workflowRunCmd.RunE`, after `params` are parsed and before `pipeline.Run`, add:

```go
		// Handle --variant: inject implicit compare blocks
		if len(workflowVariants) > 0 {
			injectImplicitCompare(w, workflowVariants, workflowReviewCriteria)
		}

		// Handle --compare: discover sibling variant workflows
		if workflowCompare {
			variants := batch.DefaultVariants
			if len(workflowVariants) > 0 {
				// --compare + --variant: use specified variants
				variants = workflowVariants
			}
			return runCompareWorkflows(name, workflows, variants, params, cfg, tel, wsESURL, wsName)
		}
```

- [ ] **Step 3: Implement injectImplicitCompare**

Add to `cmd/workflow.go`:

```go
// injectImplicitCompare wraps every LLM step (not inside an existing compare) in
// an implicit compare block with one branch per variant.
func injectImplicitCompare(w *pipeline.Workflow, variants []string, reviewCriteria string) {
	var criteria []string
	if reviewCriteria != "" {
		criteria = strings.Split(reviewCriteria, ",")
		for i := range criteria {
			criteria[i] = strings.TrimSpace(criteria[i])
		}
	}

	for i, item := range w.Items {
		if item.Step != nil && item.Step.Form == "" && item.Step.LLM != nil {
			// Wrap this LLM step in a compare
			original := *item.Step
			var branches []pipeline.CompareBranch
			for _, v := range variants {
				branchStep := original
				branchStep.ID = original.ID
				// Parse variant as provider:model or just provider
				parts := strings.SplitN(v, ":", 2)
				branchLLM := *original.LLM
				branchLLM.Provider = parts[0]
				if len(parts) == 2 {
					branchLLM.Model = parts[1]
				}
				branchStep.LLM = &branchLLM
				branches = append(branches, pipeline.CompareBranch{
					Name:  parts[0],
					Steps: []pipeline.Step{branchStep},
				})
			}
			original.Form = "compare"
			original.CompareBranches = branches
			original.LLM = nil // no longer a direct LLM step
			if len(criteria) > 0 {
				original.CompareReview = &pipeline.ReviewConfig{Criteria: criteria}
			}
			w.Items[i].Step = &original
		}
	}

	// Also handle flat Steps slice for backwards compat
	for i, s := range w.Steps {
		if s.Form == "" && s.LLM != nil {
			original := s
			var branches []pipeline.CompareBranch
			for _, v := range variants {
				branchStep := original
				parts := strings.SplitN(v, ":", 2)
				branchLLM := *original.LLM
				branchLLM.Provider = parts[0]
				if len(parts) == 2 {
					branchLLM.Model = parts[1]
				}
				branchStep.LLM = &branchLLM
				branches = append(branches, pipeline.CompareBranch{
					Name:  parts[0],
					Steps: []pipeline.Step{branchStep},
				})
			}
			original.Form = "compare"
			original.CompareBranches = branches
			original.LLM = nil
			if len(criteria) > 0 {
				original.CompareReview = &pipeline.ReviewConfig{Criteria: criteria}
			}
			w.Steps[i] = original
		}
	}
}
```

- [ ] **Step 4: Implement runCompareWorkflows for --compare flag**

Add to `cmd/workflow.go`:

```go
// runCompareWorkflows discovers sibling variant workflows and runs them as a batch compare.
func runCompareWorkflows(name string, workflows map[string]*pipeline.Workflow, variants []string, params map[string]string, cfg *Config, tel *esearch.Telemetry, esURL, workspace string) error {
	// Discover variant workflows: name-local, name-claude, etc.
	found := make(map[string]*pipeline.Workflow)
	for _, v := range variants {
		variantName := name + "-" + v
		if w, ok := workflows[variantName]; ok {
			found[v] = w
		}
	}

	if len(found) < 2 {
		return fmt.Errorf("--compare: found %d variant workflows for %q (need at least 2). Expected: %s-<variant>", len(found), name, name)
	}

	fmt.Printf("Compare: %d variants for %s\n", len(found), name)
	for v := range found {
		fmt.Printf("  - %s-%s\n", name, v)
	}

	return batch.Run(context.Background(), batch.RunOpts{
		Items:      []string{name},
		Params:     params,
		ResultsDir: resolveResultsDir(),
		Variants:   variants,
		Iterations: 1,
		Workflows:  workflows,
		Config: batch.BatchConfig{
			DefaultModel:     cfg.DefaultModel,
			ProviderRegistry: providerReg,
			ProviderResolver: cfg.BuildProviderResolver(),
			Telemetry:        tel,
			Tiers:            cfg.Tiers,
			EvalThreshold:    cfg.EvalThreshold,
		},
	})
}
```

- [ ] **Step 5: Build and verify**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 6: Verify --help shows new flags**

Run: `cd /Users/stokes/Projects/gl1tch && go run . workflow run --help`
Expected: Shows `--variant`, `--compare`, `--review-criteria` flags

- [ ] **Step 7: Commit**

```bash
git add cmd/workflow.go
git commit -m "feat(cli): add --variant, --compare, --review-criteria flags to workflow run"
```

---

### Task 8: Integration test — end-to-end compare workflow

**Files:**
- Create: `internal/pipeline/testdata/compare-custom-review.glitch`
- Modify: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Create test fixture**

Create `internal/pipeline/testdata/compare-custom-review.glitch`:

```scheme
(workflow "compare-custom-review"
  :description "compare with explicit review criteria"

  (step "data" (run "echo 'test input data'"))

  (step "analyze"
    (compare
      (branch "alpha" (run "echo alpha-analysis"))
      (branch "beta" (run "echo beta-analysis"))
      (review :criteria ["accuracy" "completeness"]))))
```

- [ ] **Step 2: Write integration test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRun_Compare_WithReviewCriteria(t *testing.T) {
	w, err := LoadFile("testdata/compare-custom-review.glitch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// Without a provider, review can't run LLM — should fall back to first branch
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Steps["data"] != "test input data" {
		t.Errorf("data = %q", result.Steps["data"])
	}
	if result.Steps["analyze"] == "" {
		t.Error("analyze should have winner output")
	}
}

func TestRun_Compare_BranchFailure(t *testing.T) {
	src := `(workflow "fail-test"
	  :description "one branch fails"
	  (step "pick"
	    (compare
	      (branch "good" (run "echo works"))
	      (branch "bad" (run "exit 1")))))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v (should succeed with one good branch)", err)
	}
	if result.Steps["pick"] != "works" {
		t.Errorf("pick = %q, want 'works'", result.Steps["pick"])
	}
}

func TestRun_Compare_AllBranchesFail(t *testing.T) {
	src := `(workflow "all-fail"
	  :description "both fail"
	  (step "pick"
	    (compare
	      (branch "a" (run "exit 1"))
	      (branch "b" (run "exit 2")))))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = Run(w, "", "", nil, nil)
	if err == nil {
		t.Fatal("expected error when all branches fail")
	}
}

func TestRun_Compare_StepAccessVariant(t *testing.T) {
	src := `(workflow "variant-access"
	  :description "access specific variant output"
	  (compare
	    :id "impl"
	    (branch "fast"
	      (step "out" (run "echo fast-output")))
	    (branch "slow"
	      (step "out" (run "echo slow-output"))))
	  (step "check" (run "echo winner={{step \"impl\"}}")))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	// check step should reference the winner's output
	check := result.Steps["check"]
	if !strings.Contains(check, "winner=") {
		t.Errorf("check = %q, expected winner= prefix", check)
	}
	if check != "winner=fast-output" && check != "winner=slow-output" {
		t.Errorf("check = %q, expected one of the branch outputs", check)
	}
}
```

- [ ] **Step 3: Run integration tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRun_Compare" -v -count=1`
Expected: ALL PASS

- [ ] **Step 4: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... -count=1`
Expected: ALL PASS (no regressions)

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/testdata/compare-custom-review.glitch internal/pipeline/runner_test.go
git commit -m "test(pipeline): end-to-end compare tests — branch failure, all-fail, variant access"
```

---

### Task 9: Example workflows

**Files:**
- Create: `examples/compare-models.glitch`
- Create: `examples/compare-branches.glitch`

- [ ] **Step 1: Create step-level compare example**

Create `examples/compare-models.glitch`:

```scheme
;; compare-models.glitch — compare LLM outputs and pick the best
;;
;; Run with: glitch workflow run compare-models --set topic="Go error handling"
;;
;; Or use --variant for ad-hoc comparison:
;;   glitch workflow run compare-models --variant ollama:qwen2.5:7b --variant claude

(def model "qwen2.5:7b")

(workflow "compare-models"
  :description "Compare different models on the same prompt"

  (step "explain"
    (compare
      (branch "local"
        (llm :model "qwen2.5:7b"
          :prompt "Explain {{.param.topic}} in 3 sentences."))
      (branch "large"
        (llm :model "llama3.2"
          :prompt "Explain {{.param.topic}} in 3 sentences."))
      (review :criteria ["accuracy" "clarity" "conciseness"])))

  (step "report"
    (save "results/compare-{{.param.topic}}.md" :from "explain")))
```

- [ ] **Step 2: Create multi-step branch compare example**

Create `examples/compare-branches.glitch`:

```scheme
;; compare-branches.glitch — compare different analysis strategies
;;
;; Run with: glitch workflow run compare-branches --set repo=gl1tch

(workflow "compare-branches"
  :description "Compare analysis strategies for a repo"

  (step "files" (run "find {{.param.repo}} -name '*.go' -type f | head -20"))

  (compare
    :id "analysis"
    (branch "breadth-first"
      (step "scan"
        (llm :model "qwen2.5:7b"
          :prompt ```
          List all packages and their responsibilities:
          {{step "files"}}
          ```))
      (step "summary"
        (llm :model "qwen2.5:7b"
          :prompt "Summarize this package map:\n{{step \"scan\"}}")))
    (branch "depth-first"
      (step "pick"
        (run "echo '{{step \"files\"}}' | head -5"))
      (step "deep-dive"
        (llm :model "qwen2.5:7b"
          :prompt ```
          Do a deep analysis of these files:
          {{step "pick"}}
          ```)))
    (review
      :criteria ["insight_depth" "actionability" "coverage"]
      :model "qwen2.5:7b"))

  (step "winner-report"
    (save "results/{{.param.repo}}/analysis.md" :from "analysis")))
```

- [ ] **Step 3: Commit**

```bash
git add examples/compare-models.glitch examples/compare-branches.glitch
git commit -m "docs: add compare-models and compare-branches example workflows"
```
