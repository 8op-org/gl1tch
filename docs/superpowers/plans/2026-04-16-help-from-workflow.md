# Help From Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. One subagent per task; review deferred until end-of-plan per user preference.

**Goal:** Make `glitch run <workflow> --help` print workflow-specific help sourced from in-file `(arg ...)` / `(input ...)` declarations plus auto-extracted `~param.X` references, and thread source locations into `UndefinedRefError`.

**Architecture:** One AST-walk pass inside the existing sexpr converter collects (a) top-level `(arg ...)` forms via the existing `plugin.ParseArgs`, (b) a new `(input ...)` form parser, and (c) implicit `~param.X` references found by walking render-capable strings through `lexQuasi`. The pipeline's `Workflow` struct grows `Args`, `Input`, `SourceFile` fields, and each `Step` grows `Line` / `Col`. A new `formatHelp(w *Workflow) string` renders cobra-idiomatic help. `UndefinedRefError` regains its `File` / `Line` / `Col` fields and is stamped by `render()` using the current step's position.

**Tech Stack:** Go, Cobra, existing `internal/sexpr` lexer/parser, existing `internal/pipeline/quasi.go` `lexQuasi` tokenizer, existing `internal/plugin/args.go` `ParseArgs`, `text/tabwriter`.

**Setup:** Work in a fresh worktree off `main` (currently at `6a689aa`). Create before starting Task 1:

```bash
cd /Users/stokes/Projects/gl1tch
git worktree add .worktrees/help-from-workflow -b feature/help-from-workflow main
cp -r internal/gui/dist .worktrees/help-from-workflow/internal/gui/dist
cd .worktrees/help-from-workflow
go build ./... && go test ./... # baseline must pass
```

**Review strategy:** Per user preference, no per-task review gate. All reviews deferred to the end-of-plan checkpoint (Task 12).

**Coverage strategy:** Per user preference, smoke pack (`glitch smoke pack`) is the coverage gate. Unit tests at each task are TDD-only — write the failing test that locks in the behavior, then implement. No speculative extra coverage.

---

## Task 1: Extend `plugin.ArgDef` with `Example`, `Implicit`, and strict parsing

**Files:**
- Modify: `internal/plugin/args.go`
- Test: `internal/plugin/args_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/plugin/args_test.go`:

```go
func TestParseArgs_ExampleKeyword(t *testing.T) {
	src := []byte(`(arg "topic" :required true :description "The topic" :example "batch comparison")`)
	defs, err := ParseArgs(src)
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("want 1 def, got %d", len(defs))
	}
	if defs[0].Example != "batch comparison" {
		t.Errorf("Example = %q, want %q", defs[0].Example, "batch comparison")
	}
	if defs[0].Implicit {
		t.Errorf("Implicit should default to false")
	}
}

func TestParseArgs_UnknownKeywordRejected(t *testing.T) {
	src := []byte(`(arg "topic" :defalt "oops")`)
	_, err := ParseArgs(src)
	if err == nil {
		t.Fatal("expected parse error for unknown keyword :defalt")
	}
}

func TestParseArgs_RequiredWithDefaultRejected(t *testing.T) {
	src := []byte(`(arg "topic" :required true :default "x")`)
	_, err := ParseArgs(src)
	if err == nil {
		t.Fatal("expected parse error when :required and :default both set")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/plugin/ -run TestParseArgs_ExampleKeyword -v
```

Expected: FAIL with "unknown field Example" or "undefined: Example".

- [ ] **Step 3: Add fields, `:example` case, and strict parsing**

In `internal/plugin/args.go`, update the struct:

```go
type ArgDef struct {
	Name        string
	Default     string // empty = required (unless type is flag)
	Type        string // "string", "flag", "number"
	Description string
	Example     string // concrete example value for help output
	Required    bool
	Implicit    bool // true when auto-extracted with no declaration
}
```

In `parseArgNode`, track whether `:required` was explicitly set so the
final `Required` computation can catch the mutual-exclusion case:

```go
// Add a local variable above the for loop.
requiredExplicit := false
```

In the `switch kw` block of `parseArgNode`:

1. Add a case after `"description"`:

```go
case "example":
	if i < len(children) {
		def.Example = children[i].StringVal()
		i++
	}
case "required":
	if i < len(children) {
		val := strings.ToLower(children[i].StringVal())
		if val == "" {
			// Keyword value form: :required true
			val = children[i].KeywordVal()
		}
		def.Required = val == "true" || val == "t" || val == "yes"
		requiredExplicit = true
		i++
	}
default:
	return ArgDef{}, fmt.Errorf("line %d: unknown keyword :%s on (arg \"%s\")", node.Line, kw, def.Name)
```

Import `strings` if not already imported.

Below the for loop, replace the existing required-computation with:

```go
// Mutual exclusion: :required true and :default "x" both set.
if requiredExplicit && def.Required && def.Default != "" {
	return ArgDef{}, fmt.Errorf("line %d: (arg \"%s\") cannot set both :required and :default", node.Line, def.Name)
}
// Default Required inference when not explicitly set.
if !requiredExplicit {
	def.Required = def.Default == "" && def.Type != "flag"
}
```

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/plugin/ -run TestParseArgs_ExampleKeyword -v
go test ./internal/plugin/... ./internal/pipeline/... # no regressions
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/args.go internal/plugin/args_test.go
git commit -m "feat(plugin): add Example and Implicit fields to ArgDef"
```

---

## Task 2: Populate `Step.Line` and `Step.Col` from sexpr nodes

**Files:**
- Modify: `internal/pipeline/types.go` (Step struct)
- Modify: `internal/pipeline/sexpr.go` (convertStep + compound-form converters)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/pipeline/sexpr_test.go`:

```go
func TestConvertStep_PopulatesLineCol(t *testing.T) {
	src := []byte(`
(workflow "pos"
  (step "a"
    (run "echo a"))
  (step "b"
    (run "echo b")))
`)
	w, err := LoadBytes(src, "pos.glitch")
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("want 2 steps, got %d", len(w.Steps))
	}
	if w.Steps[0].Line != 3 {
		t.Errorf("Steps[0].Line = %d, want 3", w.Steps[0].Line)
	}
	if w.Steps[1].Line != 5 {
		t.Errorf("Steps[1].Line = %d, want 5", w.Steps[1].Line)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestConvertStep_PopulatesLineCol -v
```

Expected: FAIL because `Step.Line` doesn't exist yet.

- [ ] **Step 3: Add `Line` / `Col` fields and populate**

In `internal/pipeline/types.go`, locate the `Step` struct and add:

```go
type Step struct {
	// ... existing fields ...
	Line int `yaml:"-"`
	Col  int `yaml:"-"`
}
```

In `internal/pipeline/sexpr.go`, at every point where a `Step` is constructed from a `*sexpr.Node`, set `Line` and `Col` from `n.Line` and `n.Col`. Minimum call sites: `convertStep`, `convertLLM`, `convertCond`, `convertWhen`, `convertMap`, `convertMapResources`, `convertFilter`, `convertReduce`, `convertPar`, `convertCatch`, `convertRetry`, `convertTimeout`, `convertCallWorkflow`, `convertCompare`. For each, immediately after `s := Step{...}` or equivalent, add `s.Line, s.Col = n.Line, n.Col`. When the returned step has a parent wrapper (retry/timeout etc.), set on the outermost step.

Grep before editing:

```bash
grep -n "Step{" internal/pipeline/sexpr.go | head -30
```

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestConvertStep_PopulatesLineCol -v
go test ./... # no regressions
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): populate Step.Line and Step.Col from sexpr nodes"
```

---

## Task 3: Add `Workflow.SourceFile` populated by `LoadFile` / `LoadBytes`

**Files:**
- Modify: `internal/pipeline/types.go` (Workflow struct)
- Modify: `internal/pipeline/types.go` (LoadFile function)
- Modify: `internal/pipeline/sexpr.go` (LoadBytes function)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/pipeline/sexpr_test.go`:

```go
func TestLoadBytes_PopulatesSourceFile(t *testing.T) {
	src := []byte(`(workflow "x" (step "s" (run "echo ok")))`)
	w, err := LoadBytes(src, "path/to/x.glitch")
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if w.SourceFile != "path/to/x.glitch" {
		t.Errorf("SourceFile = %q, want %q", w.SourceFile, "path/to/x.glitch")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestLoadBytes_PopulatesSourceFile -v
```

Expected: FAIL because `SourceFile` doesn't exist.

- [ ] **Step 3: Add field and populate**

In `internal/pipeline/types.go`, locate the `Workflow` struct and add:

```go
type Workflow struct {
	// ... existing fields ...
	SourceFile string `yaml:"-"`
}
```

In `internal/pipeline/sexpr.go`, at the top of `LoadBytes` after the workflow is constructed and before return, set `w.SourceFile = filename`.

In `internal/pipeline/types.go`, inside `LoadFile` (the YAML branch), set `w.SourceFile = path` if not already propagated via `LoadBytes`.

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestLoadBytes_PopulatesSourceFile -v
go test ./... # no regressions
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): populate Workflow.SourceFile on load"
```

---

## Task 4: Parse top-level `(arg ...)` forms into `Workflow.Args`

**Files:**
- Modify: `internal/pipeline/types.go` (Workflow struct)
- Modify: `internal/pipeline/sexpr.go` (LoadBytes)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/pipeline/sexpr_test.go`:

```go
func TestLoadBytes_CollectsArgForms(t *testing.T) {
	src := []byte(`
(arg "topic" :required true :description "Topic" :example "batch comparison")
(arg "audience" :default "developers" :description "Target audience")

(workflow "w"
  (step "s" (run "echo ~param.topic")))
`)
	w, err := LoadBytes(src, "w.glitch")
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if len(w.Args) != 2 {
		t.Fatalf("Args len = %d, want 2", len(w.Args))
	}
	if w.Args[0].Name != "topic" || w.Args[0].Example != "batch comparison" {
		t.Errorf("Args[0] = %+v", w.Args[0])
	}
	if w.Args[1].Name != "audience" || w.Args[1].Default != "developers" {
		t.Errorf("Args[1] = %+v", w.Args[1])
	}
	for _, a := range w.Args {
		if a.Implicit {
			t.Errorf("declared arg %q marked Implicit=true", a.Name)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestLoadBytes_CollectsArgForms -v
```

Expected: FAIL — `Workflow.Args` doesn't exist.

- [ ] **Step 3: Add field and invoke parser**

In `internal/pipeline/types.go`, add an import for `plugin` if missing:

```go
import "github.com/8op-org/gl1tch/internal/plugin"
```

Add to `Workflow` struct:

```go
type Workflow struct {
	// ... existing fields ...
	Args []plugin.ArgDef `yaml:"-"`
}
```

In `internal/pipeline/sexpr.go`, at the top of `LoadBytes` (before the sexpr parse loop), call `plugin.ParseArgs(src)` and store the result on `w.Args`. The existing parse loop that handles `(workflow ...)` already ignores top-level `(arg ...)` nodes because `convertWorkflow` only processes `(workflow ...)`, so no conflict. Example:

```go
func LoadBytes(src []byte, filename string) (*Workflow, error) {
	args, err := plugin.ParseArgs(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	// ... existing parse logic producing w ...

	w.Args = args
	w.SourceFile = filename
	return w, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestLoadBytes_CollectsArgForms -v
go test ./... # no regressions
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(pipeline): parse top-level (arg ...) forms into Workflow.Args"
```

---

## Task 5: Add `(input ...)` form with `InputDef` type and parser

**Files:**
- Create: `internal/pipeline/args.go`
- Modify: `internal/pipeline/types.go` (Workflow struct)
- Modify: `internal/pipeline/sexpr.go` (LoadBytes)
- Test: `internal/pipeline/args_test.go` (new)

- [ ] **Step 1: Write the failing test**

Create `internal/pipeline/args_test.go`:

```go
package pipeline

import "testing"

func TestParseInput_Basic(t *testing.T) {
	src := []byte(`(input :description "Free-form context" :example "fix latency spike")`)
	in, err := ParseInput(src)
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if in == nil {
		t.Fatal("ParseInput returned nil, want populated InputDef")
	}
	if in.Description != "Free-form context" {
		t.Errorf("Description = %q", in.Description)
	}
	if in.Example != "fix latency spike" {
		t.Errorf("Example = %q", in.Example)
	}
}

func TestParseInput_NoneReturnsNil(t *testing.T) {
	src := []byte(`(workflow "w" (step "s" (run "echo ok")))`)
	in, err := ParseInput(src)
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if in != nil {
		t.Errorf("expected nil InputDef when no (input ...), got %+v", in)
	}
}

func TestParseInput_RejectsMultiple(t *testing.T) {
	src := []byte(`(input :description "first") (input :description "second")`)
	_, err := ParseInput(src)
	if err == nil {
		t.Fatal("expected error for multiple (input ...) forms")
	}
}

func TestParseInput_RejectsPositionalName(t *testing.T) {
	src := []byte(`(input "foo" :description "bad")`)
	_, err := ParseInput(src)
	if err == nil {
		t.Fatal("expected error when (input ...) is given a name")
	}
}

func TestLoadBytes_WorkflowInputField(t *testing.T) {
	src := []byte(`
(input :description "Ctx" :example "e1")

(workflow "w" (step "s" (run "echo ~input")))
`)
	w, err := LoadBytes(src, "w.glitch")
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if w.Input == nil {
		t.Fatal("Workflow.Input is nil")
	}
	if w.Input.Description != "Ctx" || w.Input.Example != "e1" {
		t.Errorf("Input = %+v", w.Input)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestParseInput -v
```

Expected: FAIL because `ParseInput` doesn't exist.

- [ ] **Step 3: Implement `InputDef`, `ParseInput`, and wire into LoadBytes**

Create `internal/pipeline/args.go`:

```go
package pipeline

import (
	"fmt"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// InputDef describes the optional positional input for a workflow.
// At most one (input ...) form per file.
type InputDef struct {
	Description string
	Example     string
	Implicit    bool // true when auto-extracted from a ~input reference
}

// ParseInput extracts the single (input ...) form from a .glitch file.
// Returns (nil, nil) when no form is present. Returns an error when more
// than one (input ...) appears or when the form is given a name.
//
// Shape: (input :description "..." :example "...")
func ParseInput(src []byte) (*InputDef, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	var found *InputDef
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 {
			continue
		}
		if n.Children[0].SymbolVal() != "input" {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("line %d: only one (input ...) form allowed per file", n.Line)
		}

		// (input ...) takes no positional name. First child after the head
		// must be a keyword (:something), not a string.
		if len(n.Children) > 1 && !isKeyword(n.Children[1]) {
			return nil, fmt.Errorf("line %d: (input ...) takes no name, expected :keyword value pairs", n.Line)
		}

		def := &InputDef{}
		children := n.Children[1:]
		for i := 0; i < len(children); i++ {
			kw := children[i].KeywordVal()
			if kw == "" {
				continue
			}
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: (input ...) keyword :%s missing value", n.Line, kw)
			}
			switch kw {
			case "description":
				def.Description = children[i].StringVal()
			case "example":
				def.Example = children[i].StringVal()
			default:
				return nil, fmt.Errorf("line %d: (input ...) unknown keyword :%s", n.Line, kw)
			}
		}
		found = def
	}
	return found, nil
}

func isKeyword(n *sexpr.Node) bool {
	return n != nil && n.KeywordVal() != ""
}
```

In `internal/pipeline/types.go`, add to `Workflow`:

```go
Input *InputDef `yaml:"-"`
```

In `internal/pipeline/sexpr.go` `LoadBytes`, alongside `plugin.ParseArgs`:

```go
input, err := ParseInput(src)
if err != nil {
	return nil, fmt.Errorf("%s: %w", filename, err)
}

// ... existing logic ...

w.Input = input
```

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestParseInput -v
go test ./internal/pipeline/ -run TestLoadBytes_WorkflowInputField -v
go test ./... # no regressions
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/args.go internal/pipeline/args_test.go internal/pipeline/types.go internal/pipeline/sexpr.go
git commit -m "feat(pipeline): add (input ...) form and Workflow.Input field"
```

---

## Task 6: Auto-extract implicit `~param.X` and `~input` references via `lexQuasi`

**Files:**
- Create: `internal/pipeline/helpdoc.go`
- Test: `internal/pipeline/helpdoc_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/pipeline/helpdoc_test.go`:

```go
package pipeline

import (
	"sort"
	"testing"
)

func TestExtractImplicitRefs_FlatSteps(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", Run: "echo ~param.topic"},
			{ID: "b", Run: "echo ~param.audience ~param.topic"},
		},
	}
	params, usesInput := ExtractImplicitRefs(w)
	sort.Strings(params)
	want := []string{"audience", "topic"}
	if !equalStrings(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
	if usesInput {
		t.Errorf("usesInput = true, want false")
	}
}

func TestExtractImplicitRefs_DetectsInput(t *testing.T) {
	w := &Workflow{
		Steps: []Step{{ID: "a", Run: "echo ~input"}},
	}
	_, usesInput := ExtractImplicitRefs(w)
	if !usesInput {
		t.Errorf("usesInput = false, want true")
	}
}

func TestExtractImplicitRefs_InsideForms(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", Run: `echo ~(or param.topic "default") ~(upper param.audience)`},
		},
	}
	params, _ := ExtractImplicitRefs(w)
	sort.Strings(params)
	want := []string{"audience", "topic"}
	if !equalStrings(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
}

func TestExtractImplicitRefs_SkipsQuoteFirstArg(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", Run: `~(step diff) ~(stepfile foo)`},
		},
	}
	params, _ := ExtractImplicitRefs(w)
	if len(params) != 0 {
		t.Errorf("params should be empty (step/stepfile are quote-first-arg), got %v", params)
	}
}

func TestExtractImplicitRefs_LLMPrompt(t *testing.T) {
	w := &Workflow{
		Steps: []Step{
			{ID: "a", LLM: &LLMStep{Prompt: "Summarise ~param.topic for ~param.audience"}},
		},
	}
	params, _ := ExtractImplicitRefs(w)
	sort.Strings(params)
	if !equalStrings(params, []string{"audience", "topic"}) {
		t.Errorf("params = %v", params)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestExtractImplicitRefs -v
```

Expected: FAIL — `ExtractImplicitRefs` doesn't exist.

- [ ] **Step 3: Implement `ExtractImplicitRefs`**

Create `internal/pipeline/helpdoc.go`. Use the existing `lexQuasi` tokenizer and the `quoteFirstArg` map defined in `quasi.go`. For each render-capable string in every step (and nested step bodies), iterate tokens and collect `param.X` references. Walk form-token ASTs to catch `~(or param.x ...)` and similar.

```go
package pipeline

import (
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// ExtractImplicitRefs walks every render-capable string in a Workflow and
// returns (paramNames, usesInput). paramNames is the deduplicated, unordered
// set of param names referenced via ~param.X or within ~(form ...). usesInput
// is true iff any string references ~input or ~(... input ...). Quote-first-
// arg forms (step, stepfile, branch, itemfile) skip their first positional
// argument per the sexpr-interpolation spec.
func ExtractImplicitRefs(w *Workflow) (params []string, usesInput bool) {
	seen := map[string]struct{}{}
	visit := func(s string) {
		p, u := scanString(s)
		for _, name := range p {
			seen[name] = struct{}{}
		}
		if u {
			usesInput = true
		}
	}
	visitStep := func(step *Step) {
		if step == nil {
			return
		}
		visit(step.Run)
		if step.LLM != nil {
			visit(step.LLM.Prompt)
		}
		visit(step.Save)
		if step.HttpCall != nil {
			visit(step.HttpCall.URL)
			visit(step.HttpCall.Body)
			for _, v := range step.HttpCall.Headers {
				visit(v)
			}
		}
		if step.Search != nil {
			visit(step.Search.Query)
		}
		visit(step.CallInput)
		for _, v := range step.CallSet {
			visit(v)
		}
	}
	var recurse func(step *Step)
	recurse = func(step *Step) {
		if step == nil {
			return
		}
		visitStep(step)
		// Compound form bodies
		for _, b := range step.Branches {
			visit(b.Pred)
			recurse(&b.Step)
		}
		if step.WhenBody != nil {
			visit(step.WhenPred)
			recurse(step.WhenBody)
		}
		if step.MapBody != nil {
			recurse(step.MapBody)
		}
		if step.MapResourcesBody != nil {
			recurse(step.MapResourcesBody)
		}
		if step.FilterBody != nil {
			recurse(step.FilterBody)
		}
		if step.ReduceBody != nil {
			recurse(step.ReduceBody)
		}
		for i := range step.ParSteps {
			recurse(&step.ParSteps[i])
		}
		for _, cb := range step.CompareBranches {
			for i := range cb.Steps {
				recurse(&cb.Steps[i])
			}
		}
		if step.Fallback != nil {
			recurse(step.Fallback)
		}
	}
	for i := range w.Steps {
		recurse(&w.Steps[i])
	}

	params = make([]string, 0, len(seen))
	for k := range seen {
		params = append(params, k)
	}
	return params, usesInput
}

// scanString runs lexQuasi and walks the result for param / input refs.
func scanString(s string) (params []string, usesInput bool) {
	if !strings.Contains(s, "~") {
		return nil, false
	}
	tokens := lexQuasi(s)
	for _, tok := range tokens {
		switch tok.kind {
		case quasiRef:
			if tok.ref == "input" {
				usesInput = true
			} else if name, ok := paramOfPath(tok.ref); ok {
				params = append(params, name)
			}
		case quasiForm:
			// tok.formNode is the parsed AST; walk atoms, skip quote-first-arg positions.
			collectFromForm(tok.formNode, &params, &usesInput)
		}
	}
	return params, usesInput
}

// paramOfPath returns (name, true) if path looks like "param.<name>"; empty
// string otherwise.
func paramOfPath(path string) (string, bool) {
	if strings.HasPrefix(path, "param.") {
		name := strings.TrimPrefix(path, "param.")
		if name != "" && !strings.ContainsRune(name, '.') {
			return name, true
		}
	}
	return "", false
}

// collectFromForm walks a form AST node, collecting param references in every
// atom position except the first argument of quote-first-arg forms.
func collectFromForm(n *sexpr.Node, params *[]string, usesInput *bool) {
	if n == nil {
		return
	}
	if !n.IsList() || len(n.Children) == 0 {
		// Atom: check for input / param.X symbol
		if n != nil {
			sym := n.SymbolVal()
			if sym == "input" {
				*usesInput = true
				return
			}
			if name, ok := paramOfPath(sym); ok {
				*params = append(*params, name)
			}
		}
		return
	}
	head := n.Children[0].SymbolVal()
	skipIdx := -1
	if _, isQuote := quoteFirstArg[head]; isQuote {
		skipIdx = 1
	}
	for i, child := range n.Children {
		if i == 0 {
			continue
		}
		if i == skipIdx {
			continue
		}
		collectFromForm(child, params, usesInput)
	}
}
```

Notes for the implementer:
- `lexQuasi`, `quasiRef`, `quasiForm`, `quoteFirstArg` are defined in `internal/pipeline/quasi.go`. Use `grep -n lexQuasi internal/pipeline/quasi.go` to confirm exact symbol names. If the lexer exposes form-tokens as raw strings instead of parsed nodes, also run `sexpr.Parse([]byte(tok.rawForm))` before walking.
- The HTTP, Save, and other render-capable field lookups must match what `runner.go:render()` does today. If a new field gets added later that `render()` touches, this extractor must be updated.

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestExtractImplicitRefs -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/helpdoc.go internal/pipeline/helpdoc_test.go
git commit -m "feat(pipeline): auto-extract implicit ~param.X and ~input refs"
```

---

## Task 7: Merge declared `Args` with implicit extraction; warn on unreferenced

**Files:**
- Modify: `internal/pipeline/helpdoc.go`
- Modify: `internal/pipeline/sexpr.go` (LoadBytes — call merger)
- Test: `internal/pipeline/helpdoc_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/pipeline/helpdoc_test.go`:

```go
func TestMergeArgs_ImplicitFillsMissing(t *testing.T) {
	w := &Workflow{
		Args: []plugin.ArgDef{
			{Name: "topic", Description: "Topic", Required: true},
		},
		Steps: []Step{
			{ID: "a", Run: "echo ~param.topic ~param.audience"},
		},
	}
	warnings := MergeImplicitArgs(w)

	if len(w.Args) != 2 {
		t.Fatalf("Args len = %d, want 2 (declared topic + implicit audience)", len(w.Args))
	}

	var topic, audience *plugin.ArgDef
	for i := range w.Args {
		switch w.Args[i].Name {
		case "topic":
			topic = &w.Args[i]
		case "audience":
			audience = &w.Args[i]
		}
	}
	if topic == nil || topic.Implicit {
		t.Errorf("topic missing or marked implicit: %+v", topic)
	}
	if audience == nil || !audience.Implicit {
		t.Errorf("audience missing or not marked implicit: %+v", audience)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestMergeArgs_WarnsOnUnreferenced(t *testing.T) {
	w := &Workflow{
		Args: []plugin.ArgDef{
			{Name: "topic", Required: true},
			{Name: "ghost", Description: "orphan"},
		},
		Steps: []Step{
			{ID: "a", Run: "echo ~param.topic"},
		},
	}
	warnings := MergeImplicitArgs(w)
	if len(warnings) != 1 {
		t.Fatalf("want 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], `"ghost"`) {
		t.Errorf("warning should name the orphan arg, got %q", warnings[0])
	}
}

func TestMergeArgs_InputPopulatedFromRef(t *testing.T) {
	w := &Workflow{
		Steps: []Step{{ID: "a", Run: "echo ~input"}},
	}
	MergeImplicitArgs(w)
	if w.Input == nil {
		t.Fatal("Input should be populated from implicit ~input reference")
	}
	if !w.Input.Implicit {
		t.Errorf("implicit Input should have Implicit=true")
	}
}
```

Add import for `plugin` and `strings` to the test file.

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestMergeArgs -v
```

Expected: FAIL — `MergeImplicitArgs` doesn't exist.

- [ ] **Step 3: Implement the merger**

Append to `internal/pipeline/helpdoc.go`:

```go
// MergeImplicitArgs reconciles the declared Workflow.Args and Workflow.Input
// with references discovered by ExtractImplicitRefs. Implicit entries are
// appended for any referenced param with no declaration, and for ~input
// references without an (input ...) form. Returns warnings (one per
// declared arg/input with no matching reference).
func MergeImplicitArgs(w *Workflow) []string {
	params, usesInput := ExtractImplicitRefs(w)

	declared := map[string]bool{}
	for _, a := range w.Args {
		declared[a.Name] = true
	}

	referenced := map[string]bool{}
	for _, name := range params {
		referenced[name] = true
		if !declared[name] {
			w.Args = append(w.Args, plugin.ArgDef{Name: name, Implicit: true})
		}
	}

	var warnings []string
	for _, a := range w.Args {
		if !a.Implicit && !referenced[a.Name] {
			warnings = append(warnings, fmt.Sprintf(`arg "%s" declared but not referenced in any step`, a.Name))
		}
	}

	if w.Input == nil && usesInput {
		w.Input = &InputDef{Implicit: true}
	} else if w.Input != nil && !w.Input.Implicit && !usesInput {
		warnings = append(warnings, "input declared but no ~input reference in any step")
	}

	return warnings
}
```

Add imports for `fmt` and `github.com/8op-org/gl1tch/internal/plugin`.

In `internal/pipeline/sexpr.go` `LoadBytes`, after `w.Input = input`:

```go
for _, warn := range MergeImplicitArgs(w) {
	fmt.Fprintf(os.Stderr, "warning: %s: %s\n", filename, warn)
}
```

Make sure `os` is imported.

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestMergeArgs -v
go test ./... # no regressions
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/helpdoc.go internal/pipeline/helpdoc_test.go internal/pipeline/sexpr.go
git commit -m "feat(pipeline): merge declared args with implicit ref extraction"
```

---

## Task 8: Thread source locations into `UndefinedRefError`

**Files:**
- Modify: `internal/pipeline/scope.go` (UndefinedRefError struct + Error())
- Modify: `internal/pipeline/quasi.go` (render helpers take step ctx)
- Modify: `internal/pipeline/runner.go` (pass Step to render calls)
- Test: `internal/pipeline/scope_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/pipeline/scope_test.go`:

```go
func TestUndefinedRefError_HasSourceLocation(t *testing.T) {
	w := &Workflow{
		Name:       "t",
		SourceFile: "t.glitch",
		Steps: []Step{
			{ID: "s", Line: 3, Col: 2, Run: "echo ~param.missing"},
		},
	}
	_, err := Run(w, "", "", nil, nil, RunOpts{})
	if err == nil {
		t.Fatal("expected error for undefined ~param.missing")
	}
	var ue *UndefinedRefError
	if !errors.As(err, &ue) {
		t.Fatalf("error not UndefinedRefError: %v", err)
	}
	if ue.File != "t.glitch" {
		t.Errorf("File = %q, want %q", ue.File, "t.glitch")
	}
	if ue.Line != 3 {
		t.Errorf("Line = %d, want 3", ue.Line)
	}
	if ue.Col != 2 {
		t.Errorf("Col = %d, want 2", ue.Col)
	}
	msg := ue.Error()
	if !strings.Contains(msg, "t.glitch:3:2") {
		t.Errorf("Error() = %q, should contain t.glitch:3:2", msg)
	}
}
```

Add `errors` and `strings` to test imports if missing.

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestUndefinedRefError_HasSourceLocation -v
```

Expected: FAIL — fields don't exist or aren't populated.

- [ ] **Step 3: Resurrect fields and thread them through**

In `internal/pipeline/scope.go`:

```go
type UndefinedRefError struct {
	Symbol     string
	Suggestion string
	File       string
	Line       int
	Col        int
}

func (e *UndefinedRefError) Error() string {
	sug := ""
	if e.Suggestion != "" {
		sug = fmt.Sprintf(" (did you mean '%s'?)", e.Suggestion)
	}
	loc := ""
	if e.File != "" || e.Line != 0 {
		loc = fmt.Sprintf("%s:%d:%d: ", e.File, e.Line, e.Col)
	}
	return fmt.Sprintf("%sundefined reference '%s'%s", loc, e.Symbol, sug)
}
```

In `internal/pipeline/quasi.go`, locate the `render(s string, scope *Scope, steps map[string]string)` signature (or wherever the top-level render entrypoint lives). Add a stamping wrapper or extend the signature. A minimal approach: after `renderQuasi` returns an error, if it's an `UndefinedRefError` with zero location, stamp it from a passed-in context.

```go
// renderInStep is the entrypoint called from runner.go. It wraps render()
// and stamps source location onto any UndefinedRefError returned.
func renderInStep(s string, scope *Scope, steps map[string]string, w *Workflow, step *Step) (string, error) {
	out, err := render(s, scope, steps)
	if err != nil {
		var ue *UndefinedRefError
		if errors.As(err, &ue) && ue.File == "" {
			if w != nil {
				ue.File = w.SourceFile
			}
			if step != nil {
				ue.Line = step.Line
				ue.Col = step.Col
			}
		}
	}
	return out, err
}
```

In `internal/pipeline/runner.go`, find every call to `render(...)` and replace with `renderInStep(..., w, &step)`. Start with `runSingleStep` — that's the hottest call site. Use `grep -n 'render(' internal/pipeline/runner.go` to find them all.

The runner already has access to `w` (workflow) via `runCtx.workflow` (which is a string `w.Name` today — you may need to thread the full `*Workflow` through `runCtx` instead, or plumb `SourceFile` as a separate field). Minimal additive change: add `workflow *Workflow` alongside the existing `workflow string` field, and set it from `Run()` at context construction.

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestUndefinedRefError_HasSourceLocation -v
go test ./... # no regressions
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/scope.go internal/pipeline/quasi.go internal/pipeline/runner.go internal/pipeline/scope_test.go
git commit -m "feat(pipeline): thread source locations into UndefinedRefError"
```

---

## Task 9: Implement `formatHelp(w *Workflow) string`

**Files:**
- Create: `internal/pipeline/help_format.go`
- Test: `internal/pipeline/help_format_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/pipeline/help_format_test.go`:

```go
package pipeline

import (
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/plugin"
)

func TestFormatHelp_AllFields(t *testing.T) {
	w := &Workflow{
		Name:        "site-create-page",
		Description: "AI-generate a new doc page with gated verification",
		SourceFile:  ".glitch/workflows/site-create-page.glitch",
		Args: []plugin.ArgDef{
			{Name: "topic", Required: true, Description: "Topic of the page.", Example: "batch comparison"},
			{Name: "audience", Default: "developers", Description: "Target reader.", Example: "ops engineer"},
		},
		Input: &InputDef{Description: "Free-form context.", Example: "fix latency"},
		Steps: []Step{{ID: "s", Line: 6}},
	}
	out := FormatHelp(w)

	for _, want := range []string{
		"site-create-page - AI-generate a new doc page with gated verification",
		"glitch run site-create-page",
		"topic",
		"(required)",
		"Topic of the page.",
		`--set topic="batch comparison"`,
		"audience",
		"developers",
		"Free-form context.",
		".glitch/workflows/site-create-page.glitch",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatHelp output missing %q\n---\n%s", want, out)
		}
	}
}

func TestFormatHelp_NoDescription(t *testing.T) {
	w := &Workflow{Name: "bare", SourceFile: "bare.glitch"}
	out := FormatHelp(w)
	if strings.Contains(out, " - ") {
		t.Errorf("header should not have dash when :description is absent\n%s", out)
	}
	if !strings.Contains(out, "bare") {
		t.Errorf("workflow name missing from output\n%s", out)
	}
}

func TestFormatHelp_ImplicitMarker(t *testing.T) {
	w := &Workflow{
		Name: "x",
		Args: []plugin.ArgDef{{Name: "topic", Implicit: true}},
	}
	out := FormatHelp(w)
	if !strings.Contains(out, "undocumented") {
		t.Errorf("implicit args should be marked undocumented\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/pipeline/ -run TestFormatHelp -v
```

Expected: FAIL — `FormatHelp` doesn't exist.

- [ ] **Step 3: Implement `FormatHelp`**

Create `internal/pipeline/help_format.go`:

```go
package pipeline

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/8op-org/gl1tch/internal/plugin"
)

// FormatHelp renders cobra-idiomatic help text for a workflow. Output goes
// to stdout via cmd/run.go's SetHelpFunc hook.
func FormatHelp(w *Workflow) string {
	var b strings.Builder

	// Header
	if w.Description != "" {
		fmt.Fprintf(&b, "%s - %s\n\n", w.Name, w.Description)
	} else {
		fmt.Fprintf(&b, "%s\n\n", w.Name)
	}

	// Usage
	b.WriteString("Usage:\n")
	fmt.Fprintf(&b, "  glitch run %s%s%s\n\n", w.Name, usageInputFragment(w), usageArgsFragment(w.Args))

	// Arguments (positional input)
	if w.Input != nil {
		b.WriteString("Arguments:\n")
		tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
		desc := w.Input.Description
		if w.Input.Implicit {
			desc = "(undocumented — add (input :description \"...\") to annotate)"
		}
		fmt.Fprintf(tw, "  input\t%s\n", desc)
		if w.Input.Example != "" {
			fmt.Fprintf(tw, "\tExample: %q\n", w.Input.Example)
		}
		tw.Flush()
		b.WriteString("\n")
	}

	// Flags
	if len(w.Args) > 0 {
		b.WriteString("Flags:\n")
		tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
		for _, a := range w.Args {
			tag := flagTag(a)
			desc := a.Description
			if a.Implicit {
				desc = fmt.Sprintf(`(undocumented — add (arg "%s" :description "...") to annotate)`, a.Name)
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\n", a.Name, tag, desc)
			if a.Example != "" {
				fmt.Fprintf(tw, "\t\tExample: --set %s=%q\n", a.Name, a.Example)
			}
		}
		tw.Flush()
		b.WriteString("\n")
	}

	// Source
	if w.SourceFile != "" {
		line := 1
		if len(w.Steps) > 0 && w.Steps[0].Line > 0 {
			line = w.Steps[0].Line
		}
		fmt.Fprintf(&b, "Defined in: %s:%d\n", w.SourceFile, line)
	}

	return b.String()
}

func usageInputFragment(w *Workflow) string {
	if w.Input == nil {
		return ""
	}
	return " [<input>]"
}

func usageArgsFragment(args []plugin.ArgDef) string {
	if len(args) == 0 {
		return ""
	}
	var parts []string
	for _, a := range args {
		frag := fmt.Sprintf("--set %s=<%s>", a.Name, a.Name)
		if !a.Required {
			frag = "[" + frag + "]"
		}
		parts = append(parts, frag)
	}
	return " " + strings.Join(parts, " ")
}

func flagTag(a plugin.ArgDef) string {
	if a.Required {
		return "(required)"
	}
	if a.Default != "" {
		return fmt.Sprintf("(optional, default: %s)", a.Default)
	}
	return "(optional)"
}

// Suppress unused-import warnings when buffering; bytes.Buffer not used but
// kept for future streaming writes if output grows.
var _ = bytes.Buffer{}
```

Remove the `bytes` import if it genuinely isn't used — the above sketch keeps it for clarity but `strings.Builder` is sufficient.

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/pipeline/ -run TestFormatHelp -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/help_format.go internal/pipeline/help_format_test.go
git commit -m "feat(pipeline): FormatHelp renders cobra-idiomatic workflow help"
```

---

## Task 10: Wire cobra `SetHelpFunc` in `cmd/run.go`

**Files:**
- Modify: `cmd/run.go`
- Test: `cmd/run_test.go`

- [ ] **Step 1: Write the failing test**

Append to `cmd/run_test.go`:

```go
func TestRunCmd_HelpFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Scaffold a workflow with declared args.
	wfDir := filepath.Join(home, ".config", "glitch", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	src := `(arg "topic" :required true :description "Topic" :example "batch")

(workflow "demo" :description "demo workflow"
  (step "s" (run "echo ~param.topic")))
`
	if err := os.WriteFile(filepath.Join(wfDir, "demo.glitch"), []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"run", "demo", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"demo - demo workflow", "topic", "(required)", "--set topic"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing %q:\n%s", want, out)
		}
	}
}
```

Make sure `bytes` is imported.

- [ ] **Step 2: Run test to verify it fails**

```
go test ./cmd/ -run TestRunCmd_HelpFlag -v
```

Expected: FAIL — cobra prints its generic help, not workflow-specific.

- [ ] **Step 3: Wire `SetHelpFunc`**

In `cmd/run.go`, inside `init()`, after all flag definitions but before `rootCmd.AddCommand(runCmd)`:

```go
runCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
	// Fallback to cobra's default when no workflow name is present.
	if len(args) < 1 {
		cmd.Root().HelpFunc()(cmd, args)
		return
	}
	name := args[0]
	path, err := resolveWorkflowPath(name)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		return
	}
	w, err := pipeline.LoadFile(path)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "load %s: %v\n", name, err)
		return
	}
	fmt.Fprint(cmd.OutOrStdout(), pipeline.FormatHelp(w))
})
```

Note: cobra's `HelpFunc` receives the cmd and the args that came before `--help`. `args` here is the slice remaining after cobra parses flags — test it with `rootCmd.SetArgs([]string{"run", "demo", "--help"})` to confirm `args == ["demo"]` inside the hook.

- [ ] **Step 4: Run test to verify it passes**

```
go test ./cmd/ -run TestRunCmd_HelpFlag -v
go test ./... # no regressions
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/run.go cmd/run_test.go
git commit -m "feat(cmd): glitch run <workflow> --help prints workflow-specific help"
```

---

## Task 11: Self-check existing workflows parse without warnings

**Files:**
- None (verification task)

- [ ] **Step 1: Build and run a dry parse over every workflow in the repo**

```bash
go build -o /tmp/glitch-help ./
for f in .glitch/workflows/*.glitch examples/*.glitch test-workspace/workflows/*.glitch 2>/dev/null; do
  /tmp/glitch-help run "$(basename "$f" .glitch)" --help 2>&1 | head -3
done
```

Expected: each workflow renders help without a parse error. Warnings about unreferenced args are informational — expected only if a workflow declared an arg it doesn't use.

- [ ] **Step 2: Fix any unexpected parse errors**

If a workflow fails to load because of one of the loud parse errors (`unknown keyword`, etc.), the fix is in the workflow file, not the parser. Note each and fix inline. Commit fixes with a `fix(workflows):` prefix, one commit per fix.

- [ ] **Step 3: Commit any workflow corrections**

If no corrections needed, skip this step. Otherwise:

```bash
git add .glitch/workflows/... # or whichever path
git commit -m "fix(workflows): <specific change>"
```

---

## Task 12: End-of-plan review and smoke pack

**Files:**
- None (verification + merge)

- [ ] **Step 1: Run full test suite**

```bash
go test ./... 2>&1 | tail -25
```

Expected: all packages PASS.

- [ ] **Step 2: Invoke code review subagent**

Dispatch a review subagent with the spec (`docs/superpowers/specs/2026-04-16-help-from-workflow-design.md`) and the branch diff (`git log --oneline main..HEAD` + `git diff main...HEAD`). Reviewer checks:

- Every spec requirement has a corresponding implementation.
- `UndefinedRefError.File/Line/Col` populate end-to-end, not just at top-level steps.
- Auto-extraction handles every render-capable string the runtime renderer touches.
- Warnings go to stderr (not stdout, which is reserved for step output).
- No placeholder code or "TODO" markers left in production paths.

Capture findings; address any blockers inline with fix commits. Importants go in follow-up commits same branch.

- [ ] **Step 3: Run smoke pack as the coverage gate**

```bash
/tmp/glitch-help smoke pack 2>&1 | tail -30
```

Expected: 24/24 baseline. Anything less → bisect against `main` before merging. Per user's smoke-pack memory: this is the acceptance gate. A red smoke pack blocks merge.

- [ ] **Step 4: Merge to main**

When smoke pack is green and review is addressed:

```bash
cd /Users/stokes/Projects/gl1tch
git merge --no-ff feature/help-from-workflow -m "Merge branch 'feature/help-from-workflow'

Workflow-level --help driven by declared (arg ...) / (input ...) forms
plus auto-extracted ~param.X refs. Source-location threading for
UndefinedRefError folded in. Spec:
docs/superpowers/specs/2026-04-16-help-from-workflow-design.md"
```

**Do not push.** Per user's no-push-without-asking rule — merge locally only. Ask before any push.

- [ ] **Step 5: Prune worktree**

```bash
git worktree remove .worktrees/help-from-workflow
git worktree list
```
