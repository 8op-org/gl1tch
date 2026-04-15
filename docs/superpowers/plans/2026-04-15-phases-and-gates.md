# Phases & Gates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `(phase ...)` and `(gate ...)` forms to the glitch workflow sexpr syntax so workflows can declare verified, retriable units of work.

**Architecture:** The Workflow struct gains a new `Items []WorkflowItem` field — an ordered sequence of bare steps and phases. The runner walks Items if present, falling back to the existing Steps field for backward compat. Phases group steps and gates; gates are steps with `IsGate=true` whose output is evaluated as pass/fail. A new `executePhase()` function handles the retry loop.

**Tech Stack:** Go, existing sexpr parser, existing pipeline runner

---

### Task 1: Add Phase and WorkflowItem types to types.go

**Files:**
- Modify: `internal/pipeline/types.go:12-17`
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Phase(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase "gather"
    (step "fetch" (run "echo data")))

  (phase "verify" :retries 2
    (step "process" (run "echo processed"))
    (gate "check" (run "test -f output.txt"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(w.Items))
	}

	// Phase 1: gather
	p0 := w.Items[0].Phase
	if p0 == nil {
		t.Fatal("expected item 0 to be a phase")
	}
	if p0.ID != "gather" {
		t.Fatalf("expected phase id %q, got %q", "gather", p0.ID)
	}
	if p0.Retries != 0 {
		t.Fatalf("expected retries 0, got %d", p0.Retries)
	}
	if len(p0.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(p0.Steps))
	}
	if p0.Steps[0].ID != "fetch" {
		t.Fatalf("expected step id %q, got %q", "fetch", p0.Steps[0].ID)
	}
	if len(p0.Gates) != 0 {
		t.Fatalf("expected 0 gates, got %d", len(p0.Gates))
	}

	// Phase 2: verify
	p1 := w.Items[1].Phase
	if p1 == nil {
		t.Fatal("expected item 1 to be a phase")
	}
	if p1.ID != "verify" {
		t.Fatalf("expected phase id %q, got %q", "verify", p1.ID)
	}
	if p1.Retries != 2 {
		t.Fatalf("expected retries 2, got %d", p1.Retries)
	}
	if len(p1.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(p1.Steps))
	}
	if p1.Steps[0].ID != "process" {
		t.Fatalf("expected step id %q, got %q", "process", p1.Steps[0].ID)
	}
	if len(p1.Gates) != 1 {
		t.Fatalf("expected 1 gate, got %d", len(p1.Gates))
	}
	if p1.Gates[0].ID != "check" {
		t.Fatalf("expected gate id %q, got %q", "check", p1.Gates[0].ID)
	}
	if !p1.Gates[0].IsGate {
		t.Fatal("expected gate to have IsGate=true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Phase -v`
Expected: FAIL — `w.Items` field doesn't exist, `Phase` type doesn't exist, `IsGate` field doesn't exist.

- [ ] **Step 3: Add types to types.go**

Add after the `Workflow` struct definition in `internal/pipeline/types.go`:

```go
// WorkflowItem is a union type for the ordered sequence of workflow elements.
// Exactly one of Step or Phase is non-nil.
type WorkflowItem struct {
	Step  *Step
	Phase *Phase
}

// Phase groups steps and gates into a retriable unit of work.
type Phase struct {
	ID      string
	Retries int
	Steps   []Step // work steps
	Gates   []Step // verification steps (IsGate = true)
}
```

Add the `IsGate` field to the `Step` struct:

```go
IsGate bool `yaml:"-"` // true for gate steps inside a phase
```

Add the `Items` field to the `Workflow` struct:

```go
Items []WorkflowItem `yaml:"-"` // ordered sequence of bare steps and phases (sexpr only)
```

- [ ] **Step 4: Run test to verify it still fails (types exist but parser doesn't populate them)**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Phase -v`
Expected: FAIL — `w.Items` is empty because parser doesn't handle `phase` or `gate` yet.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): add Phase, WorkflowItem, IsGate types for phase/gate support"
```

---

### Task 2: Parse phase and gate forms in sexpr.go

**Files:**
- Modify: `internal/pipeline/sexpr.go:50-101` (convertWorkflow), `internal/pipeline/sexpr.go:103-147` (convertForm)

- [ ] **Step 1: Add `convertPhase` and `convertGate` functions to sexpr.go**

Add to `internal/pipeline/sexpr.go`:

```go
// convertPhase: (phase "name" [:retries N] (step ...) ... (gate ...) ...)
func convertPhase(n *sexpr.Node, defs map[string]string) (Phase, error) {
	children := n.Children[1:] // skip "phase"
	if len(children) == 0 {
		return Phase{}, fmt.Errorf("line %d: (phase) missing name", n.Line)
	}

	p := Phase{}
	p.ID = children[0].StringVal()
	if p.ID == "" {
		return Phase{}, fmt.Errorf("line %d: phase name must be a string", children[0].Line)
	}
	children = children[1:]

	// Process keywords and children
	i := 0
	for i < len(children) {
		child := children[i]
		// Keyword args
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return Phase{}, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "retries":
				n, err := strconv.Atoi(resolveVal(val, defs))
				if err != nil {
					return Phase{}, fmt.Errorf("line %d: :retries must be an integer", val.Line)
				}
				p.Retries = n
			default:
				return Phase{}, fmt.Errorf("line %d: unknown phase keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		// Child forms: step or gate
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].SymbolVal()
			if head == "" {
				head = child.Children[0].StringVal()
			}
			switch head {
			case "gate":
				g, err := convertGate(child, defs)
				if err != nil {
					return Phase{}, err
				}
				p.Gates = append(p.Gates, g)
			case "step":
				s, err := convertStep(child, defs)
				if err != nil {
					return Phase{}, err
				}
				p.Steps = append(p.Steps, s)
			default:
				return Phase{}, fmt.Errorf("line %d: unexpected form %q inside phase (expected step or gate)", child.Line, head)
			}
			i++
			continue
		}
		return Phase{}, fmt.Errorf("line %d: unexpected form in phase", child.Line)
	}
	return p, nil
}

// convertGate: (gate "name" (run ...) | (llm ...))
// Structurally identical to convertStep but sets IsGate = true.
func convertGate(n *sexpr.Node, defs map[string]string) (Step, error) {
	s, err := convertStep(n, defs)
	if err != nil {
		return Step{}, err
	}
	s.IsGate = true
	return s, nil
}
```

- [ ] **Step 2: Update convertWorkflow to handle phase forms and populate Items**

In `convertWorkflow` in `internal/pipeline/sexpr.go`, update the loop that processes children. Replace the block that calls `convertForm` with:

```go
if child.IsList() && len(child.Children) > 0 {
	head := child.Children[0].SymbolVal()
	if head == "" {
		head = child.Children[0].StringVal()
	}
	if head == "phase" {
		p, err := convertPhase(child, defs)
		if err != nil {
			return nil, err
		}
		w.Items = append(w.Items, WorkflowItem{Phase: &p})
		// Also flatten phase steps into w.Steps for backward compat
		w.Steps = append(w.Steps, p.Steps...)
		w.Steps = append(w.Steps, p.Gates...)
		i++
		continue
	}
	if head == "gate" {
		return nil, fmt.Errorf("line %d: (gate) must be inside a (phase)", child.Line)
	}
	steps, err := convertForm(child, head, defs)
	if err != nil {
		return nil, err
	}
	w.Steps = append(w.Steps, steps...)
	for idx := range steps {
		w.Items = append(w.Items, WorkflowItem{Step: &steps[idx]})
	}
	i++
	continue
}
```

Also update the existing non-phase path — bare steps parsed via `convertForm` need to be added to Items too. The key change is that every step/form that was previously just appended to `w.Steps` also gets appended to `w.Items` as a `WorkflowItem{Step: &step}`.

- [ ] **Step 3: Run the phase test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Phase -v`
Expected: PASS

- [ ] **Step 4: Run all existing tests to verify no regressions**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All existing tests PASS (bare-step workflows still work).

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/sexpr.go
git commit -m "feat(pipeline): parse phase and gate forms in sexpr parser"
```

---

### Task 3: Add parser edge-case tests

**Files:**
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write tests for mixed bare steps and phases**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_MixedStepsAndPhases(t *testing.T) {
	src := []byte(`
(workflow "mixed"
  (step "setup" (run "echo start"))
  (phase "core" :retries 1
    (step "work" (run "echo working"))
    (gate "check" (run "test -f done.txt")))
  (step "cleanup" (run "echo done")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(w.Items))
	}
	if w.Items[0].Step == nil || w.Items[0].Step.ID != "setup" {
		t.Fatal("expected item 0 to be bare step 'setup'")
	}
	if w.Items[1].Phase == nil || w.Items[1].Phase.ID != "core" {
		t.Fatal("expected item 1 to be phase 'core'")
	}
	if w.Items[2].Step == nil || w.Items[2].Step.ID != "cleanup" {
		t.Fatal("expected item 2 to be bare step 'cleanup'")
	}
}

func TestSexprWorkflow_GateOutsidePhase(t *testing.T) {
	src := []byte(`
(workflow "bad"
  (gate "orphan" (run "echo fail")))
`)
	_, err := parseSexprWorkflow(src)
	if err == nil {
		t.Fatal("expected error for gate outside phase")
	}
	if !strings.Contains(err.Error(), "must be inside") {
		t.Fatalf("expected 'must be inside' error, got: %v", err)
	}
}

func TestSexprWorkflow_PhaseWithLLMGate(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase "analyze" :retries 1
    (step "gen"
      (llm :prompt "generate something"))
    (gate "review"
      (llm :tier 2 :prompt "review: {{step \"gen\"}}"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	p := w.Items[0].Phase
	if p == nil {
		t.Fatal("expected phase")
	}
	if len(p.Gates) != 1 {
		t.Fatalf("expected 1 gate, got %d", len(p.Gates))
	}
	g := p.Gates[0]
	if g.LLM == nil {
		t.Fatal("expected LLM gate")
	}
	if g.LLM.Tier == nil || *g.LLM.Tier != 2 {
		t.Fatalf("expected tier 2, got %v", g.LLM.Tier)
	}
	if !g.IsGate {
		t.Fatal("expected IsGate=true")
	}
}

func TestSexprWorkflow_PhaseNoName(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase
    (step "s" (run "echo"))))
`)
	_, err := parseSexprWorkflow(src)
	if err == nil {
		t.Fatal("expected error for phase without name")
	}
}

func TestSexprWorkflow_NoPhases_ItemsPopulated(t *testing.T) {
	src := []byte(`
(workflow "old-style"
  (step "a" (run "echo a"))
  (step "b" (run "echo b")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 2 {
		t.Fatalf("expected 2 items for bare-step workflow, got %d", len(w.Items))
	}
	if w.Items[0].Step == nil || w.Items[0].Step.ID != "a" {
		t.Fatal("expected item 0 to be step 'a'")
	}
	if w.Items[1].Step == nil || w.Items[1].Step.ID != "b" {
		t.Fatal("expected item 1 to be step 'b'")
	}
}
```

- [ ] **Step 2: Add `"strings"` import if not already present in sexpr_test.go**

Check the imports at the top of `sexpr_test.go`. Add `"strings"` to the import block if missing.

- [ ] **Step 3: Run all parser tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow -v`
Expected: All PASS

- [ ] **Step 4: Commit**

```bash
git add internal/pipeline/sexpr_test.go
git commit -m "test(pipeline): add parser tests for phase/gate edge cases"
```

---

### Task 4: Add gate pass/fail evaluation to runner

**Files:**
- Modify: `internal/pipeline/runner.go`
- Test: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write the failing test for gate evaluation**

Add to `internal/pipeline/runner_test.go`:

```go
func TestEvaluateGate_ShellPass(t *testing.T) {
	outcome := &stepOutcome{output: "all good", isLLM: false}
	pass, detail := evaluateGate(Step{ID: "check", Run: "true", IsGate: true}, outcome, nil)
	if !pass {
		t.Fatalf("expected shell gate to pass, got fail: %s", detail)
	}
}

func TestEvaluateGate_LLMPass(t *testing.T) {
	outcome := &stepOutcome{output: "Criterion 1: PASS\nCriterion 2: PASS\nOVERALL: PASS", isLLM: true}
	pass, _ := evaluateGate(Step{ID: "review", IsGate: true, LLM: &LLMStep{}}, outcome, nil)
	if !pass {
		t.Fatal("expected LLM gate with OVERALL: PASS to pass")
	}
}

func TestEvaluateGate_LLMFail(t *testing.T) {
	outcome := &stepOutcome{output: "Criterion 1: PASS\nCriterion 2: FAIL\nOVERALL: FAIL", isLLM: true}
	pass, _ := evaluateGate(Step{ID: "review", IsGate: true, LLM: &LLMStep{}}, outcome, nil)
	if pass {
		t.Fatal("expected LLM gate with OVERALL: FAIL to fail")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestEvaluateGate -v`
Expected: FAIL — `evaluateGate` doesn't exist.

- [ ] **Step 3: Implement evaluateGate in runner.go**

Add to `internal/pipeline/runner.go`:

```go
// evaluateGate determines if a gate step passed or failed.
// Shell gates: the step already ran; err from executeStep means failure.
// This function is called only when the step succeeded (no exec error).
// For LLM gates, parse output for PASS/FAIL verdict.
// Returns (passed, failureDetail).
func evaluateGate(step Step, outcome *stepOutcome, execErr error) (bool, string) {
	// Shell gate: execution error = fail
	if execErr != nil {
		return false, execErr.Error()
	}

	// LLM gate: parse for PASS/FAIL
	if step.LLM != nil {
		upper := strings.ToUpper(strings.ReplaceAll(outcome.output, "*", ""))
		if strings.Contains(upper, "OVERALL: PASS") || strings.Contains(upper, "OVERALL PASS") {
			return true, ""
		}
		if strings.Contains(upper, "OVERALL: FAIL") || strings.Contains(upper, "OVERALL FAIL") {
			return false, outcome.output
		}
		// No explicit verdict — count PASS vs FAIL lines
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
		// No PASS/FAIL lines found — treat as pass (gate output is informational)
		return true, ""
	}

	// Shell gate that executed successfully = pass
	return true, ""
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestEvaluateGate -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(pipeline): add evaluateGate for gate pass/fail evaluation"
```

---

### Task 5: Implement executePhase in runner.go

**Files:**
- Modify: `internal/pipeline/runner.go`
- Test: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write the failing test for executePhase — gates pass**

Add to `internal/pipeline/runner_test.go`:

```go
func TestExecutePhase_AllGatesPass(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")

	p := Phase{
		ID:      "test-phase",
		Retries: 0,
		Steps: []Step{
			{ID: "create", Run: fmt.Sprintf("touch %s", marker)},
		},
		Gates: []Step{
			{ID: "verify", Run: fmt.Sprintf("test -f %s", marker), IsGate: true},
		},
	}

	rctx := &runCtx{
		steps:  make(map[string]string),
		params: map[string]string{},
	}

	report, err := executePhase(rctx, p)
	if err != nil {
		t.Fatalf("expected phase to pass, got error: %v", err)
	}
	if report != nil {
		t.Fatalf("expected nil report on success, got: %v", report)
	}
}

func TestExecutePhase_GateFailsNoRetry(t *testing.T) {
	p := Phase{
		ID:      "test-phase",
		Retries: 0,
		Steps: []Step{
			{ID: "work", Run: "echo done"},
		},
		Gates: []Step{
			{ID: "check", Run: "false", IsGate: true},
		},
	}

	rctx := &runCtx{
		steps:  make(map[string]string),
		params: map[string]string{},
	}

	report, err := executePhase(rctx, p)
	if err == nil {
		t.Fatal("expected error when gate fails with no retries")
	}
	if report == nil {
		t.Fatal("expected verification report")
	}
	if report.Phase != "test-phase" {
		t.Fatalf("expected phase %q, got %q", "test-phase", report.Phase)
	}
	if report.Gates[0].Passed {
		t.Fatal("expected gate 'check' to fail")
	}
}

func TestExecutePhase_GateFailsThenPassesOnRetry(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")

	// Shell gate that fails first time (file doesn't exist),
	// but the step creates it on each run. Gate checks for TWO files.
	// First run: step creates file1. Gate checks file1 AND file2 → fail.
	// Second run: step creates file2. Gate checks file1 AND file2 → pass.
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	counter := filepath.Join(dir, "counter.txt")

	// Step: on first call creates file1, on second creates file2
	stepCmd := fmt.Sprintf("if [ ! -f %s ]; then touch %s && touch %s; else touch %s; fi", counter, counter, file1, file2)
	gateCmd := fmt.Sprintf("test -f %s && test -f %s", file1, file2)
	_ = marker // unused, using file1/file2 instead

	p := Phase{
		ID:      "retry-phase",
		Retries: 1,
		Steps: []Step{
			{ID: "create", Run: stepCmd},
		},
		Gates: []Step{
			{ID: "verify", Run: gateCmd, IsGate: true},
		},
	}

	rctx := &runCtx{
		steps:  make(map[string]string),
		params: map[string]string{},
	}

	report, err := executePhase(rctx, p)
	if err != nil {
		t.Fatalf("expected phase to pass on retry, got error: %v", err)
	}
	if report != nil {
		t.Fatalf("expected nil report on success, got: %v", report)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestExecutePhase -v`
Expected: FAIL — `executePhase` doesn't exist, `VerificationReport` doesn't exist.

- [ ] **Step 3: Implement executePhase and VerificationReport**

Add to `internal/pipeline/runner.go`:

```go
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
		if attempt > 1 {
			fmt.Printf("\n>>> Phase %q retry %d/%d\n", p.ID, attempt-1, p.Retries)
		} else {
			fmt.Printf("\n>>> Phase %q\n", p.ID)
		}

		// Run all work steps
		for _, step := range p.Steps {
			fmt.Printf("  > %s\n", step.ID)
			outcome, err := executeStep(rctx, step)
			if err != nil {
				return nil, fmt.Errorf("phase %q step %q: %w", p.ID, step.ID, err)
			}
			rctx.steps[step.ID] = outcome.output
		}

		// Run gates in order, stop on first failure
		lastGateResults = make([]GateResult, len(p.Gates))
		allPassed := true

		for gi, gate := range p.Gates {
			fmt.Printf("  > gate %s\n", gate.ID)
			outcome, execErr := executeStep(rctx, gate)

			if execErr != nil {
				// Shell gate execution failure
				rctx.steps[gate.ID] = execErr.Error()
				lastGateResults[gi] = GateResult{ID: gate.ID, Passed: false, Detail: execErr.Error()}
				allPassed = false
				// Mark remaining gates as skipped
				for ski := gi + 1; ski < len(p.Gates); ski++ {
					lastGateResults[ski] = GateResult{ID: p.Gates[ski].ID, Skipped: true}
				}
				break
			}

			rctx.steps[gate.ID] = outcome.output
			passed, detail := evaluateGate(gate, outcome, nil)
			lastGateResults[gi] = GateResult{ID: gate.ID, Passed: passed, Detail: detail}

			if !passed {
				allPassed = false
				// Mark remaining gates as skipped
				for ski := gi + 1; ski < len(p.Gates); ski++ {
					lastGateResults[ski] = GateResult{ID: p.Gates[ski].ID, Skipped: true}
				}
				break
			}
		}

		if allPassed {
			return nil, nil
		}

		if attempt < maxAttempts {
			// Log which gate failed for feedback
			for _, gr := range lastGateResults {
				if !gr.Passed && !gr.Skipped {
					fmt.Printf("  > gate %s FAILED, retrying phase\n", gr.ID)
					break
				}
			}
		}
	}

	// All retries exhausted
	report := &VerificationReport{
		Phase:    p.ID,
		Attempts: maxAttempts,
		MaxRetry: p.Retries,
		Gates:    lastGateResults,
	}
	return report, fmt.Errorf("phase %q: all %d attempts exhausted\n%s", p.ID, maxAttempts, report.FormatReport())
}
```

- [ ] **Step 4: Run executePhase tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestExecutePhase -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(pipeline): implement executePhase with gate evaluation and retry"
```

---

### Task 6: Wire executePhase into the main Run() function

**Files:**
- Modify: `internal/pipeline/runner.go:88-210` (the `Run` function)
- Test: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write the failing integration test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRun_WithPhases(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "output.txt")

	w := &Workflow{
		Name: "phase-test",
		Items: []WorkflowItem{
			{Phase: &Phase{
				ID: "gather",
				Steps: []Step{
					{ID: "fetch", Run: fmt.Sprintf("echo hello > %s && echo hello", outFile)},
				},
			}},
			{Phase: &Phase{
				ID: "verify",
				Steps: []Step{
					{ID: "process", Run: "echo processed"},
				},
				Gates: []Step{
					{ID: "check", Run: fmt.Sprintf("test -f %s", outFile), IsGate: true},
				},
			}},
		},
	}

	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps["fetch"] == "" {
		t.Fatal("expected fetch step output")
	}
	if result.Steps["process"] == "" {
		t.Fatal("expected process step output")
	}
}

func TestRun_WithPhases_GateFails(t *testing.T) {
	w := &Workflow{
		Name: "fail-test",
		Items: []WorkflowItem{
			{Phase: &Phase{
				ID: "work",
				Steps: []Step{
					{ID: "do-thing", Run: "echo working"},
				},
				Gates: []Step{
					{ID: "check", Run: "false", IsGate: true},
				},
			}},
		},
	}

	_, err := Run(w, "", "", nil, nil)
	if err == nil {
		t.Fatal("expected error when gate fails")
	}
	if !strings.Contains(err.Error(), "all") || !strings.Contains(err.Error(), "exhausted") {
		t.Fatalf("expected exhaustion error, got: %v", err)
	}
}

func TestRun_BareStepsStillWork(t *testing.T) {
	// Verify backward compat: workflow with only Steps (no Items) still works
	w := &Workflow{
		Name: "bare-test",
		Steps: []Step{
			{ID: "a", Run: "echo alpha"},
			{ID: "b", Run: "echo beta"},
		},
	}

	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Steps["a"]) != "alpha" {
		t.Fatalf("expected 'alpha', got %q", result.Steps["a"])
	}
	if strings.TrimSpace(result.Steps["b"]) != "beta" {
		t.Fatalf("expected 'beta', got %q", result.Steps["b"])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRun_WithPhases|TestRun_BareStepsStillWork" -v`
Expected: FAIL — `Run()` doesn't know about Items/phases yet.

- [ ] **Step 3: Update Run() to walk Items when present**

In `internal/pipeline/runner.go`, modify the `Run` function. Replace the step iteration loop (the `for i, step := range w.Steps` block starting around line 148) with logic that checks for Items first:

```go
	// If Items is populated (sexpr with phases), walk the ordered item list.
	// Otherwise fall back to flat Steps for backward compat (YAML workflows,
	// old sexpr workflows, and programmatically constructed workflows).
	if len(w.Items) > 0 {
		for _, item := range w.Items {
			if item.Step != nil {
				step := *item.Step
				if _, seeded := steps[step.ID]; seeded {
					fmt.Printf("  > %s (seeded, skipped)\n", step.ID)
					continue
				}

				outcome, err := executeStep(rctx, step)
				if err != nil {
					return nil, err
				}
				rctx.steps[step.ID] = outcome.output
				accumulateTelemetry(step, outcome, w, runID, issue, compGroup, defaultModel, tel,
					&totalTokensIn, &totalTokensOut, &totalCostUSD, &llmSteps, &lastLLMOutput)
			} else if item.Phase != nil {
				report, err := executePhase(rctx, *item.Phase)
				if err != nil {
					if report != nil {
						// Store the verification report as the workflow output
						rctx.steps["_verification_report"] = report.FormatReport()
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

			if _, seeded := steps[step.ID]; seeded {
				fmt.Printf("  > %s (seeded, skipped)\n", step.ID)
				continue
			}

			outcome, err := executeStep(rctx, step)
			if err != nil {
				return nil, err
			}
			accumulateTelemetry(step, outcome, w, runID, issue, compGroup, defaultModel, tel,
				&totalTokensIn, &totalTokensOut, &totalCostUSD, &llmSteps, &lastLLMOutput)
		}
	}
```

Extract the telemetry accumulation into a helper to avoid duplication:

```go
func accumulateTelemetry(step Step, outcome *stepOutcome, w *Workflow, runID, issue, compGroup, defaultModel string, tel *esearch.Telemetry,
	totalTokensIn, totalTokensOut *int64, totalCostUSD *float64, llmSteps *int, lastLLMOutput *string) {
	if outcome.isLLM {
		tokIn := int64(outcome.tokensIn)
		tokOut := int64(outcome.tokensOut)
		*totalTokensIn += tokIn
		*totalTokensOut += tokOut
		*totalCostUSD += outcome.cost
		*llmSteps++
		*lastLLMOutput = outcome.output

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
}
```

Also update the `Result` construction at the end of `Run()` to use `rctx.steps` (it already does — the steps map is shared).

- [ ] **Step 4: Run all tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS — new phase tests and all existing tests.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(pipeline): wire executePhase into Run() with backward-compat fallback"
```

---

### Task 7: End-to-end test with a real .glitch workflow file

**Files:**
- Create: `internal/pipeline/testdata/phase-gate.glitch`
- Test: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Create the test workflow file**

Run: `mkdir -p /Users/stokes/Projects/gl1tch/internal/pipeline/testdata`

Create `internal/pipeline/testdata/phase-gate.glitch`:

```scheme
(workflow "phase-gate-test"
  :description "end-to-end phase and gate test"

  (phase "gather"
    (step "data" (run "echo 'hello world'")))

  (phase "process" :retries 1
    (step "transform" (run "echo 'TRANSFORMED: hello world'"))
    (gate "not-empty" (run "test -n \"$(echo 'TRANSFORMED: hello world')\"")))

  (step "done" (run "echo finished")))
```

- [ ] **Step 2: Write the end-to-end test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRun_EndToEnd_PhaseGateWorkflow(t *testing.T) {
	w, err := LoadFile("testdata/phase-gate.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "phase-gate-test" {
		t.Fatalf("expected name %q, got %q", "phase-gate-test", w.Name)
	}

	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(result.Steps["data"]) != "hello world" {
		t.Fatalf("expected data='hello world', got %q", result.Steps["data"])
	}
	if !strings.Contains(result.Steps["transform"], "TRANSFORMED") {
		t.Fatalf("expected transform output, got %q", result.Steps["transform"])
	}
	if strings.TrimSpace(result.Steps["done"]) != "finished" {
		t.Fatalf("expected done='finished', got %q", result.Steps["done"])
	}
}
```

- [ ] **Step 3: Run the end-to-end test**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestRun_EndToEnd_PhaseGateWorkflow -v`
Expected: PASS

- [ ] **Step 4: Run the full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/testdata/phase-gate.glitch internal/pipeline/runner_test.go
git commit -m "test(pipeline): add end-to-end test for phase/gate workflow"
```

---

### Task 8: Run full project tests and verify no regressions

**Files:**
- No file changes — verification only

- [ ] **Step 1: Run the full project test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -30`
Expected: All packages PASS (or pre-existing failures unrelated to this change).

- [ ] **Step 2: Build the binary**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./cmd/glitch`
Expected: Clean build, no errors.

- [ ] **Step 3: Verify a workflow without phases still runs**

Run: `cd /Users/stokes/Projects/gl1tch && echo '(workflow "smoke" (step "hi" (run "echo ok")))' > /tmp/smoke.glitch && go run ./cmd/glitch workflow run /tmp/smoke.glitch`
Expected: Runs and prints "ok" — backward compat confirmed.

- [ ] **Step 4: Commit the plan as done**

No code to commit — this is a verification step. If all passes, the feature is complete.
