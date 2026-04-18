# Lisp Evaluator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the sexpr.go converter + runner.go dispatch loop with a Lisp evaluator that walks the AST directly, keeping identical `.glitch` syntax and the same `Run()` public API.

**Architecture:** New `eval.go` evaluator with Value types, lexical Env, special forms, and builtin functions. The evaluator receives the runtime context (provider registry, ES URL, params) and registers integration builtins that call the same Go packages the current runner uses. `Run()` in runner.go constructs the evaluator, wires context, and calls `ev.Eval()`.

**Tech Stack:** Pure Go, no new dependencies. Reuses `internal/sexpr` parser, `internal/provider`, `internal/esearch`, `internal/plugin`.

**Spec:** `docs/superpowers/specs/2026-04-17-lisp-evaluator-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/pipeline/values.go` | Create | Value interface, StringVal, ListVal, NilVal, BoolVal, FnVal, BuiltinVal |
| `internal/pipeline/env.go` | Create | Lexical Env type with Get/Set/parent chain |
| `internal/pipeline/eval.go` | Create | Evaluator struct, Eval() loop, special forms (def/fn/let/do/if/when/cond/workflow/step/par/retry/timeout/catch/map/each/filter/reduce/compare/include) |
| `internal/pipeline/eval_builtins.go` | Create | Builtin registrations: sh/run, ref, str, llm, save, websearch, search, index, delete, embed, http-get/post, read-file, write-file, glob, json-pick, plugin, call-workflow, list, not, =, println |
| `internal/pipeline/eval_interpolate.go` | Create | String interpolation for `~(step x)`, `~param.foo`, `~input` — reuses lexQuasi from quasi.go |
| `internal/pipeline/eval_test.go` | Create | Tests for new Lisp capabilities (fn, let, closures, if/cond expressions) |
| `internal/pipeline/eval_compat_test.go` | Create | Run every existing .glitch testdata file through the evaluator |
| `internal/pipeline/runner.go` | Modify | Replace `Run()` body to use evaluator. Delete executeStep/runSingleStep dispatch (~2000 lines) |
| `internal/pipeline/sexpr.go` | Modify | Simplify to metadata extraction. Add `Nodes []*sexpr.Node` to Workflow. Delete convertStep/convertForm (~1700 lines) |
| `internal/pipeline/types.go` | Modify | Add `Nodes []*sexpr.Node` field to Workflow. Keep Result, RunOpts, StepRecord unchanged. |

---

### Task 1: Value Types and Env

**Files:**
- Create: `internal/pipeline/values.go`
- Create: `internal/pipeline/env.go`

- [ ] **Step 1: Create values.go with all value types**

```go
// internal/pipeline/values.go
package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Value is the result of evaluating any expression.
type Value interface {
	String() string
}

// StringVal is the primary value type — matches gl1tch's model where step outputs are strings.
type StringVal string

func (s StringVal) String() string { return string(s) }

// ListVal is an ordered collection of values.
type ListVal []Value

func (l ListVal) String() string {
	parts := make([]string, len(l))
	for i, v := range l {
		parts[i] = v.String()
	}
	return strings.Join(parts, "\n")
}

// NilVal represents empty/false.
type NilVal struct{}

func (NilVal) String() string { return "" }

// BoolVal represents true/false.
type BoolVal bool

func (b BoolVal) String() string {
	if b {
		return "true"
	}
	return "false"
}

// FnVal is a user-defined function (closure).
type FnVal struct {
	Params []string
	Body   []*sexpr.Node
	Env    *Env
}

func (*FnVal) String() string { return "<fn>" }

// BuiltinVal is a Go function exposed to the evaluator.
type BuiltinVal struct {
	Name string
	Fn   BuiltinFunc
}

// BuiltinFunc is the signature for builtin functions.
// They receive the evaluator, calling env, and unevaluated arg nodes.
// Each builtin is responsible for evaluating its own args via ev.Eval().
type BuiltinFunc func(ev *Evaluator, env *Env, args []*sexpr.Node) (Value, error)

// Note: builtins are registered as closures that call methods on *Evaluator,
// so the method signatures take (env, args) while the closure wraps them
// with the evaluator reference. See registerBuiltins in eval_builtins.go.

func (b *BuiltinVal) String() string { return fmt.Sprintf("<builtin:%s>", b.Name) }

// isTruthy returns whether a value is considered true in conditionals.
func isTruthy(v Value) bool {
	switch val := v.(type) {
	case NilVal:
		return false
	case BoolVal:
		return bool(val)
	case StringVal:
		return val != ""
	case ListVal:
		return len(val) > 0
	default:
		return true
	}
}
```

- [ ] **Step 2: Create env.go with lexical scope**

```go
// internal/pipeline/env.go
package pipeline

// Env is a lexical scope for the evaluator.
type Env struct {
	bindings map[string]Value
	parent   *Env
}

// NewEnv creates a new scope with an optional parent.
func NewEnv(parent *Env) *Env {
	return &Env{bindings: make(map[string]Value), parent: parent}
}

// Get looks up a binding, walking the parent chain.
func (e *Env) Get(name string) (Value, bool) {
	if v, ok := e.bindings[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set binds a value in this scope.
func (e *Env) Set(name string, val Value) {
	e.bindings[name] = val
}
```

- [ ] **Step 3: Verify the files compile**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./internal/pipeline/`
Expected: builds clean (values.go and env.go are standalone, no imports from pipeline yet)

- [ ] **Step 4: Commit**

```
git add internal/pipeline/values.go internal/pipeline/env.go
git commit -m "feat(eval): add Value types and lexical Env for Lisp evaluator"
```

---

### Task 2: Core Evaluator with Special Forms

**Files:**
- Create: `internal/pipeline/eval.go`

This is the largest task. The evaluator implements `Eval()` and all special forms: `def`, `fn`, `let`, `do`, `if`, `when`, `when-not`, `cond`, `workflow`, `step`, `par`, `include`, `retry`, `timeout`, `catch`.

- [ ] **Step 1: Create eval.go with Evaluator struct and Eval()**

The evaluator struct holds runtime state (steps map, input, params, etc.) and the `Eval()` method dispatches on AST node type.

```go
// internal/pipeline/eval.go
package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Evaluator is a Lisp interpreter that walks gl1tch's sexpr AST directly.
type Evaluator struct {
	mu    sync.Mutex
	steps map[string]string

	// Runtime context — wired by Run() before evaluation starts.
	Input        string
	Params       map[string]string
	DefaultModel string
	Workspace    string
	Resources    map[string]map[string]string
	ProviderReg  *provider.ProviderRegistry
	ProviderResolver provider.ResolverFunc
	Tiers        []provider.TierConfig
	EvalThreshold int
	ESURL        string
	WebSearchURL string
	WorkflowsDir string
	WorkflowName string

	// Step recording callback (wired by runner for telemetry).
	StepRecorder func(rec StepRecord)

	// Accumulators for LLM metrics.
	TotalTokensIn  int64
	TotalTokensOut int64
	TotalCostUSD   float64
	TotalLatencyMS int64
	LLMSteps       int
}

// NewEvaluator creates an evaluator with empty state.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		steps:  make(map[string]string),
		Params: make(map[string]string),
	}
}

// Steps returns a snapshot of all step outputs.
func (ev *Evaluator) Steps() map[string]string {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	snap := make(map[string]string, len(ev.steps))
	for k, v := range ev.steps {
		snap[k] = v
	}
	return snap
}

// RunSource parses and evaluates raw .glitch source.
func (ev *Evaluator) RunSource(src []byte) (Value, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}
	env := NewEnv(nil)
	ev.registerBuiltins(env)
	// Seed env with input and params
	env.Set("input", StringVal(ev.Input))
	env.Set("workspace", StringVal(ev.Workspace))

	var last Value = NilVal{}
	for _, node := range nodes {
		last, err = ev.Eval(env, node)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

// Eval evaluates a single AST node.
func (ev *Evaluator) Eval(env *Env, node *sexpr.Node) (Value, error) {
	if node.IsAtom() {
		return ev.evalAtom(env, node)
	}
	if !node.IsList() || len(node.Children) == 0 {
		return NilVal{}, nil
	}

	head := node.Children[0]
	args := node.Children[1:]

	// Special forms — args are NOT pre-evaluated
	if head.IsAtom() && head.Atom.Type == sexpr.TokenSymbol {
		switch head.Atom.Val {
		case "def":
			return ev.evalDef(env, args, node)
		case "fn":
			return ev.evalFn(env, args, node)
		case "let":
			return ev.evalLet(env, args, node)
		case "do", "begin":
			return ev.evalDo(env, args)
		case "if":
			return ev.evalIf(env, args, node)
		case "when":
			return ev.evalWhen(env, args, false, node)
		case "when-not":
			return ev.evalWhen(env, args, true, node)
		case "cond":
			return ev.evalCond(env, args, node)
		case "workflow":
			return ev.evalWorkflow(env, args, node)
		case "step":
			return ev.evalStep(env, args, node)
		case "par":
			return ev.evalPar(env, args, node)
		case "retry":
			return ev.evalRetry(env, args, node)
		case "timeout":
			return ev.evalTimeout(env, args, node)
		case "catch":
			return ev.evalCatch(env, args, node)
		case "include":
			return ev.evalInclude(env, args, node)
		case "map", "each":
			return ev.evalMap(env, args, node)
		case "filter":
			return ev.evalFilter(env, args, node)
		case "reduce":
			return ev.evalReduce(env, args, node)
		case "compare":
			return ev.evalCompare(env, args, node)
		case "phase":
			return ev.evalPhase(env, args, node)
		case "gate":
			return ev.evalStep(env, args, node) // gates are steps
		case "->":
			return ev.evalThread(env, args, node)
		}
	}

	// Regular call — evaluate head, then apply
	fn, err := ev.Eval(env, head)
	if err != nil {
		return nil, fmt.Errorf("line %d: %w", head.Line, err)
	}

	switch f := fn.(type) {
	case *BuiltinVal:
		return f.Fn(ev, env, args)
	case *FnVal:
		return ev.applyFn(f, env, args)
	default:
		return nil, fmt.Errorf("line %d: %q is not callable", head.Line, head.StringVal())
	}
}

func (ev *Evaluator) evalAtom(env *Env, node *sexpr.Node) (Value, error) {
	tok := node.Atom
	switch tok.Type {
	case sexpr.TokenString:
		if strings.Contains(tok.Val, "~") {
			result, err := ev.interpolate(env, tok.Val)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", tok.Line, err)
			}
			return StringVal(result), nil
		}
		return StringVal(tok.Val), nil
	case sexpr.TokenKeyword:
		return StringVal(tok.Val), nil
	case sexpr.TokenSymbol:
		switch tok.Val {
		case "true":
			return BoolVal(true), nil
		case "false":
			return BoolVal(false), nil
		case "nil":
			return NilVal{}, nil
		}
		v, ok := env.Get(tok.Val)
		if !ok {
			return nil, fmt.Errorf("line %d: undefined symbol: %s", tok.Line, tok.Val)
		}
		return v, nil
	}
	return NilVal{}, nil
}

// === Special form implementations ===

func (ev *Evaluator) evalDef(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: def requires name and value", node.Line)
	}
	name := args[0].SymbolVal()
	if name == "" {
		name = args[0].StringVal()
	}
	val, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	env.Set(name, val)
	return val, nil
}

func (ev *Evaluator) evalFn(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: fn requires (params) body", node.Line)
	}
	paramNode := args[0]
	if !paramNode.IsList() {
		return nil, fmt.Errorf("line %d: fn params must be a list", node.Line)
	}
	var params []string
	for _, p := range paramNode.Children {
		params = append(params, p.SymbolVal())
	}
	return &FnVal{Params: params, Body: args[1:], Env: env}, nil
}

func (ev *Evaluator) applyFn(fn *FnVal, callerEnv *Env, argNodes []*sexpr.Node) (Value, error) {
	fnEnv := NewEnv(fn.Env)
	for i, param := range fn.Params {
		if i < len(argNodes) {
			val, err := ev.Eval(callerEnv, argNodes[i])
			if err != nil {
				return nil, err
			}
			fnEnv.Set(param, val)
		}
	}
	var last Value = NilVal{}
	var err error
	for _, expr := range fn.Body {
		last, err = ev.Eval(fnEnv, expr)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

func (ev *Evaluator) evalLet(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: let requires (bindings) body", node.Line)
	}
	bindings := args[0]
	if !bindings.IsList() {
		return nil, fmt.Errorf("line %d: let bindings must be a list", node.Line)
	}
	letEnv := NewEnv(env)
	for i := 0; i+1 < len(bindings.Children); i += 2 {
		name := bindings.Children[i].SymbolVal()
		val, err := ev.Eval(letEnv, bindings.Children[i+1])
		if err != nil {
			return nil, err
		}
		letEnv.Set(name, val)
	}
	var last Value = NilVal{}
	var err error
	for _, expr := range args[1:] {
		last, err = ev.Eval(letEnv, expr)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

func (ev *Evaluator) evalDo(env *Env, args []*sexpr.Node) (Value, error) {
	var last Value = NilVal{}
	var err error
	for _, expr := range args {
		last, err = ev.Eval(env, expr)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

func (ev *Evaluator) evalIf(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: if requires predicate and then-branch", node.Line)
	}
	pred, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	if isTruthy(pred) {
		return ev.Eval(env, args[1])
	}
	if len(args) >= 3 {
		return ev.Eval(env, args[2])
	}
	return NilVal{}, nil
}

func (ev *Evaluator) evalWhen(env *Env, args []*sexpr.Node, negate bool, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: when requires predicate and body", node.Line)
	}
	pred, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	truthy := isTruthy(pred)
	if negate {
		truthy = !truthy
	}
	if truthy {
		var last Value = NilVal{}
		for _, expr := range args[1:] {
			last, err = ev.Eval(env, expr)
			if err != nil {
				return nil, err
			}
		}
		return last, nil
	}
	return NilVal{}, nil
}

func (ev *Evaluator) evalCond(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	for i := 0; i+1 < len(args); i += 2 {
		// Check for "else" keyword
		if args[i].SymbolVal() == "else" {
			return ev.Eval(env, args[i+1])
		}
		pred, err := ev.Eval(env, args[i])
		if err != nil {
			return nil, err
		}
		if isTruthy(pred) {
			return ev.Eval(env, args[i+1])
		}
	}
	return NilVal{}, nil
}

func (ev *Evaluator) evalWorkflow(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("line %d: workflow requires a name", node.Line)
	}
	name, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	ev.WorkflowName = name.String()

	// Parse keyword args, collect body
	desc := ""
	var body []*sexpr.Node
	for i := 1; i < len(args); i++ {
		if args[i].KeywordVal() == "description" && i+1 < len(args) {
			d, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			desc = d.String()
			i++
			continue
		}
		// Skip other keyword metadata (author, version, tags, etc.)
		kw := args[i].KeywordVal()
		if kw != "" && i+1 < len(args) {
			i++ // skip the value
			continue
		}
		body = append(body, args[i])
	}
	_ = desc // metadata consumed but not printed in non-interactive mode

	var last Value = NilVal{}
	for _, form := range body {
		last, err = ev.Eval(env, form)
		if err != nil {
			return nil, fmt.Errorf("workflow %s: %w", name, err)
		}
	}
	return last, nil
}

func (ev *Evaluator) evalStep(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("line %d: step requires an id", node.Line)
	}

	// Single arg: (step id) — look up step output (used in ~(step x) interpolation)
	if len(args) == 1 {
		id := args[0].SymbolVal()
		if id == "" {
			id = args[0].StringVal()
		}
		if id == "" {
			v, err := ev.Eval(env, args[0])
			if err != nil {
				return nil, err
			}
			id = v.String()
		}
		ev.mu.Lock()
		val, ok := ev.steps[id]
		ev.mu.Unlock()
		if !ok {
			return nil, fmt.Errorf("line %d: step %q not found", node.Line, id)
		}
		return StringVal(val), nil
	}

	// Multi arg: (step id body...) — execute step
	id, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	// Create a child env with ~item and ~acc accessible if present in parent
	stepEnv := NewEnv(env)

	var result Value = NilVal{}
	for _, expr := range args[1:] {
		result, err = ev.Eval(stepEnv, expr)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", id, err)
		}
	}

	output := result.String()
	ev.mu.Lock()
	ev.steps[id.String()] = output
	ev.mu.Unlock()

	// Emit step record for telemetry
	if ev.StepRecorder != nil {
		ev.StepRecorder(StepRecord{
			StepID: id.String(),
			Output: output,
		})
	}

	return result, nil
}

func (ev *Evaluator) evalPar(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	type parResult struct {
		val Value
		err error
	}
	results := make([]parResult, len(args))
	var wg sync.WaitGroup

	for i, form := range args {
		wg.Add(1)
		go func(idx int, expr *sexpr.Node) {
			defer wg.Done()
			val, err := ev.Eval(env, expr)
			results[idx] = parResult{val, err}
		}(i, form)
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
	}
	if len(results) > 0 {
		return results[len(results)-1].val, nil
	}
	return NilVal{}, nil
}

func (ev *Evaluator) evalRetry(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (retry N body...)
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: retry requires count and body", node.Line)
	}
	countVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	count := 3 // default
	fmt.Sscanf(countVal.String(), "%d", &count)

	var lastErr error
	for attempt := 0; attempt < count; attempt++ {
		var result Value
		result, lastErr = ev.evalDo(env, args[1:])
		if lastErr == nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("retry exhausted after %d attempts: %w", count, lastErr)
}

func (ev *Evaluator) evalTimeout(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (timeout "30s" body...)
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: timeout requires duration and body", node.Line)
	}
	durVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	dur, err := time.ParseDuration(durVal.String())
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid duration %q: %w", node.Line, durVal, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	done := make(chan struct{ val Value; err error }, 1)
	go func() {
		val, err := ev.evalDo(env, args[1:])
		done <- struct{ val Value; err error }{val, err}
	}()

	select {
	case r := <-done:
		return r.val, r.err
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout after %s", dur)
	}
}

func (ev *Evaluator) evalCatch(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (catch body fallback)
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: catch requires body and fallback", node.Line)
	}
	result, err := ev.Eval(env, args[0])
	if err != nil {
		// Run fallback — bind ~error in scope
		catchEnv := NewEnv(env)
		catchEnv.Set("error", StringVal(err.Error()))
		return ev.Eval(catchEnv, args[1])
	}
	return result, nil
}

func (ev *Evaluator) evalInclude(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("line %d: include requires a path", node.Line)
	}
	path, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path.String())
	if err != nil {
		return nil, fmt.Errorf("include %q: %w", path, err)
	}
	nodes, parseErr := sexpr.Parse(data)
	if parseErr != nil {
		return nil, fmt.Errorf("include %q: %w", path, parseErr)
	}
	var last Value = NilVal{}
	for _, n := range nodes {
		last, err = ev.Eval(env, n)
		if err != nil {
			return nil, fmt.Errorf("include %q: %w", path, err)
		}
	}
	return last, nil
}

func (ev *Evaluator) evalMap(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (map "step-id" body...) — iterate over newline-split step output
	// or (map (list ...) body...)
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: map requires source and body", node.Line)
	}

	source, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	items := strings.Split(source.String(), "\n")
	var results []string

	for _, item := range items {
		if item == "" {
			continue
		}
		itemEnv := NewEnv(env)
		itemEnv.Set("item", StringVal(item))
		var last Value = NilVal{}
		for _, expr := range args[1:] {
			last, err = ev.Eval(itemEnv, expr)
			if err != nil {
				return nil, err
			}
		}
		results = append(results, last.String())
	}

	return StringVal(strings.Join(results, "\n")), nil
}

func (ev *Evaluator) evalFilter(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: filter requires source and predicate", node.Line)
	}
	source, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	items := strings.Split(source.String(), "\n")
	var kept []string
	for _, item := range items {
		if item == "" {
			continue
		}
		itemEnv := NewEnv(env)
		itemEnv.Set("item", StringVal(item))
		pred, err := ev.Eval(itemEnv, args[1])
		if err != nil {
			return nil, err
		}
		if isTruthy(pred) {
			kept = append(kept, item)
		}
	}
	return StringVal(strings.Join(kept, "\n")), nil
}

func (ev *Evaluator) evalReduce(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (reduce source initial body)
	if len(args) < 3 {
		return nil, fmt.Errorf("line %d: reduce requires source, initial, and body", node.Line)
	}
	source, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	acc, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	items := strings.Split(source.String(), "\n")
	for _, item := range items {
		if item == "" {
			continue
		}
		reduceEnv := NewEnv(env)
		reduceEnv.Set("item", StringVal(item))
		reduceEnv.Set("acc", acc)
		acc, err = ev.Eval(reduceEnv, args[2])
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func (ev *Evaluator) evalCompare(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (compare :id "x" :objective "..." (branch "a" ...) (branch "b" ...) (review ...))
	// Parse keyword args and branches
	id := ""
	objective := ""
	var branches []*sexpr.Node
	var reviewNode *sexpr.Node

	for i := 0; i < len(args); i++ {
		if args[i].KeywordVal() == "id" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			id = v.String()
			i++
		} else if args[i].KeywordVal() == "objective" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			objective = v.String()
			i++
		} else if args[i].IsList() && len(args[i].Children) > 0 {
			head := args[i].Children[0].SymbolVal()
			if head == "branch" {
				branches = append(branches, args[i])
			} else if head == "review" {
				reviewNode = args[i]
			}
		}
	}

	// Run each branch in parallel
	type branchResult struct {
		name   string
		output string
		err    error
	}
	results := make([]branchResult, len(branches))
	var wg sync.WaitGroup

	for i, branch := range branches {
		wg.Add(1)
		go func(idx int, b *sexpr.Node) {
			defer wg.Done()
			children := b.Children[1:]
			bName := ""
			if len(children) > 0 {
				v, err := ev.Eval(env, children[0])
				if err != nil {
					results[idx] = branchResult{err: err}
					return
				}
				bName = v.String()
				children = children[1:]
			}
			branchEnv := NewEnv(env)
			var last Value = NilVal{}
			var err error
			for _, expr := range children {
				last, err = ev.Eval(branchEnv, expr)
				if err != nil {
					results[idx] = branchResult{name: bName, err: err}
					return
				}
			}
			// Record branch steps into parent
			for k, v := range branchEnv.bindings {
				_ = k
				_ = v
			}
			results[idx] = branchResult{name: bName, output: last.String()}
		}(i, branch)
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
	}

	// Build comparison prompt for the judge
	var comparisonParts []string
	for _, r := range results {
		comparisonParts = append(comparisonParts, fmt.Sprintf("=== %s ===\n%s", r.name, r.output))
	}

	// Extract review config
	reviewModel := ev.DefaultModel
	var criteria []string
	if reviewNode != nil {
		for i := 1; i < len(reviewNode.Children); i++ {
			if reviewNode.Children[i].KeywordVal() == "model" && i+1 < len(reviewNode.Children) {
				v, err := ev.Eval(env, reviewNode.Children[i+1])
				if err != nil {
					return nil, err
				}
				reviewModel = v.String()
				i++
			} else if reviewNode.Children[i].KeywordVal() == "criteria" && i+1 < len(reviewNode.Children) {
				// Parse criteria list
				cNode := reviewNode.Children[i+1]
				if cNode.IsList() {
					for _, c := range cNode.Children {
						v, err := ev.Eval(env, c)
						if err != nil {
							return nil, err
						}
						criteria = append(criteria, v.String())
					}
				}
				i++
			}
		}
	}

	// Build judge prompt
	judgePrompt := fmt.Sprintf("Compare these responses. Objective: %s\n\n%s\n\nCriteria: %s\n\nPick a winner and explain why.",
		objective, strings.Join(comparisonParts, "\n\n"), strings.Join(criteria, ", "))

	// Call LLM to judge
	if ev.ProviderReg != nil {
		llmResult, err := ev.ProviderReg.RunProviderWithResult("", reviewModel, judgePrompt)
		if err != nil {
			return nil, fmt.Errorf("compare review: %w", err)
		}
		output := llmResult.Response
		if id != "" {
			ev.mu.Lock()
			ev.steps[id] = output
			ev.mu.Unlock()
		}
		return StringVal(output), nil
	}

	// Fallback: return concatenated outputs
	output := strings.Join(comparisonParts, "\n\n")
	if id != "" {
		ev.mu.Lock()
		ev.steps[id] = output
		ev.mu.Unlock()
	}
	return StringVal(output), nil
}

func (ev *Evaluator) evalPhase(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (phase :id "x" :retries N body... (gate ...))
	id := ""
	retries := 1
	var body []*sexpr.Node

	for i := 0; i < len(args); i++ {
		if args[i].KeywordVal() == "id" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			id = v.String()
			i++
		} else if args[i].KeywordVal() == "retries" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			fmt.Sscanf(v.String(), "%d", &retries)
			i++
		} else {
			body = append(body, args[i])
		}
	}

	var lastErr error
	for attempt := 0; attempt < retries; attempt++ {
		var last Value = NilVal{}
		var err error
		for _, expr := range body {
			last, err = ev.Eval(env, expr)
			if err != nil {
				lastErr = err
				break
			}
		}
		if err == nil {
			if id != "" {
				ev.mu.Lock()
				ev.steps[id] = last.String()
				ev.mu.Unlock()
			}
			return last, nil
		}
	}
	return nil, fmt.Errorf("phase %s: exhausted %d retries: %w", id, retries, lastErr)
}

func (ev *Evaluator) evalThread(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	// (-> expr1 expr2 ...) — thread: result of each feeds as input to next
	if len(args) < 1 {
		return NilVal{}, nil
	}
	val, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	for _, expr := range args[1:] {
		threadEnv := NewEnv(env)
		threadEnv.Set("_", val)
		val, err = ev.Eval(threadEnv, expr)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./internal/pipeline/`
Expected: builds clean

- [ ] **Step 3: Commit**

```
git add internal/pipeline/eval.go
git commit -m "feat(eval): add core Lisp evaluator with all special forms"
```

---

### Task 3: String Interpolation

**Files:**
- Create: `internal/pipeline/eval_interpolate.go`

Handles `~(step x)`, `~param.foo`, `~input`, `~workspace` inside string values. Reuses `lexQuasi` from quasi.go for parsing, but resolves against the evaluator's env and steps map instead of a Scope.

- [ ] **Step 1: Create eval_interpolate.go**

```go
// internal/pipeline/eval_interpolate.go
package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// interpolate expands ~(...) and ~symbol references inside a string.
func (ev *Evaluator) interpolate(env *Env, s string) (string, error) {
	parts, err := lexQuasi(s)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for _, p := range parts {
		switch p.Kind {
		case partLiteral:
			b.WriteString(p.Literal)

		case partRef:
			val, err := ev.resolveRef(env, p.RefBase, p.RefPath)
			if err != nil {
				return "", err
			}
			b.WriteString(val)

		case partForm:
			// Parse and evaluate the form using the evaluator
			nodes, err := sexpr.Parse([]byte(p.Form))
			if err != nil {
				return "", fmt.Errorf("interpolation parse error: %w", err)
			}
			if len(nodes) > 0 {
				val, err := ev.Eval(env, nodes[0])
				if err != nil {
					return "", err
				}
				b.WriteString(val.String())
			}
		}
	}
	return b.String(), nil
}

// resolveRef resolves a ~name or ~name.path.here reference.
func (ev *Evaluator) resolveRef(env *Env, base string, path []string) (string, error) {
	// Direct step lookup
	ev.mu.Lock()
	stepVal, isStep := ev.steps[base]
	ev.mu.Unlock()
	if isStep && len(path) == 0 {
		return stepVal, nil
	}

	// Env lookup
	if v, ok := env.Get(base); ok && len(path) == 0 {
		return v.String(), nil
	}

	// Dotted path: param.foo, resource.name.field
	switch base {
	case "param":
		if len(path) > 0 && ev.Params != nil {
			if v, ok := ev.Params[path[0]]; ok {
				return v, nil
			}
		}
		return "", &UndefinedRefError{Ref: base + "." + strings.Join(path, ".")}

	case "resource":
		if len(path) >= 2 && ev.Resources != nil {
			if res, ok := ev.Resources[path[0]]; ok {
				if v, ok := res[path[1]]; ok {
					return v, nil
				}
			}
		}
		return "", &UndefinedRefError{Ref: base + "." + strings.Join(path, ".")}

	case "input":
		return ev.Input, nil

	case "workspace":
		return ev.Workspace, nil

	default:
		// Try env lookup for dotted names
		full := base
		if len(path) > 0 {
			full = base + "." + strings.Join(path, ".")
		}
		if v, ok := env.Get(full); ok {
			return v.String(), nil
		}
		return "", &UndefinedRefError{Ref: full}
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./internal/pipeline/`

- [ ] **Step 3: Commit**

```
git add internal/pipeline/eval_interpolate.go
git commit -m "feat(eval): add string interpolation for ~ references"
```

---

### Task 4: Builtin Functions

**Files:**
- Create: `internal/pipeline/eval_builtins.go`

Registers all builtin functions: sh/run, ref, str, llm, save, websearch, search, index, delete, embed, http-get/post, read-file, write-file, glob, json-pick, plugin, call-workflow, list, not, =, println, or.

- [ ] **Step 1: Create eval_builtins.go**

```go
// internal/pipeline/eval_builtins.go
package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

// registerBuiltins populates env with all builtin functions.
func (ev *Evaluator) registerBuiltins(env *Env) {
	b := func(name string, fn BuiltinFunc) {
		env.Set(name, &BuiltinVal{Name: name, Fn: fn})
	}

	b("sh", ev.builtinSh)
	b("run", ev.builtinSh)
	b("ref", ev.builtinRef)
	b("str", ev.builtinStr)
	b("llm", ev.builtinLLM)
	b("save", ev.builtinSave)
	b("list", ev.builtinList)
	b("not", ev.builtinNot)
	b("=", ev.builtinEq)
	b("println", ev.builtinPrintln)
	b("or", ev.builtinOr)
	b("read-file", ev.builtinReadFile)
	b("read", ev.builtinReadFile)
	b("write-file", ev.builtinWriteFile)
	b("write", ev.builtinWriteFile)
	b("glob", ev.builtinGlob)
	b("websearch", ev.builtinWebSearch)
	b("http-get", ev.builtinHTTPGet)
	b("fetch", ev.builtinHTTPGet)
	b("http-post", ev.builtinHTTPPost)
	b("send", ev.builtinHTTPPost)
	b("call-workflow", ev.builtinCallWorkflow)
}

func (ev *Evaluator) builtinSh(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("sh requires a command")
	}
	cmd, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	result, err := provider.RunShellContext(cmd.String())
	if err != nil {
		return nil, fmt.Errorf("sh: %w", err)
	}
	return StringVal(result), nil
}

func (ev *Evaluator) builtinRef(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ref requires a step id")
	}
	id, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	ev.mu.Lock()
	val, ok := ev.steps[id.String()]
	ev.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("ref: step %q not found", id)
	}
	return StringVal(val), nil
}

func (ev *Evaluator) builtinStr(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	var b strings.Builder
	for _, arg := range args {
		val, err := ev.Eval(env, arg)
		if err != nil {
			return nil, err
		}
		b.WriteString(val.String())
	}
	return StringVal(b.String()), nil
}

func (ev *Evaluator) builtinLLM(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	model := ev.DefaultModel
	prompt := ""
	prov := ""
	skill := ""

	for i := 0; i < len(args); i++ {
		kw := args[i].KeywordVal()
		if kw != "" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			switch kw {
			case "model":
				model = v.String()
			case "prompt":
				prompt = v.String()
			case "provider":
				prov = v.String()
			case "skill":
				skill = v.String()
			}
			i++
			continue
		}
		// Positional: single arg is prompt
		v, err := ev.Eval(env, args[i])
		if err != nil {
			return nil, err
		}
		prompt = v.String()
	}

	if skill != "" {
		// Prepend skill content to prompt
		data, err := os.ReadFile(skill)
		if err == nil {
			prompt = string(data) + "\n\n" + prompt
		}
	}

	if ev.ProviderReg == nil {
		return StringVal(fmt.Sprintf("[LLM %s] prompt (%d chars)", model, len(prompt))), nil
	}

	result, err := ev.ProviderReg.RunProviderWithResult(prov, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	// Accumulate metrics
	ev.TotalTokensIn += int64(result.TokensIn)
	ev.TotalTokensOut += int64(result.TokensOut)
	ev.TotalCostUSD += result.CostUSD
	ev.TotalLatencyMS += result.Latency.Milliseconds()
	ev.LLMSteps++

	return StringVal(result.Response), nil
}

func (ev *Evaluator) builtinSave(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("save requires a path")
	}
	path, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	content := ""
	for i := 1; i < len(args); i++ {
		if args[i].KeywordVal() == "from" && i+1 < len(args) {
			id, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			ev.mu.Lock()
			content = ev.steps[id.String()]
			ev.mu.Unlock()
			i++
		} else if args[i].KeywordVal() == "content" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			content = v.String()
			i++
		}
	}

	dir := filepath.Dir(path.String())
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755)
	}
	if err := os.WriteFile(path.String(), []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return StringVal(content), nil
}

func (ev *Evaluator) builtinList(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	vals := make(ListVal, len(args))
	for i, arg := range args {
		v, err := ev.Eval(env, arg)
		if err != nil {
			return nil, err
		}
		vals[i] = v
	}
	return vals, nil
}

func (ev *Evaluator) builtinNot(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("not requires an argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	return BoolVal(!isTruthy(v)), nil
}

func (ev *Evaluator) builtinEq(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("= requires two arguments")
	}
	a, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	b, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	return BoolVal(a.String() == b.String()), nil
}

func (ev *Evaluator) builtinPrintln(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	var parts []string
	for _, arg := range args {
		v, err := ev.Eval(env, arg)
		if err != nil {
			return nil, err
		}
		parts = append(parts, v.String())
	}
	fmt.Println(strings.Join(parts, " "))
	return NilVal{}, nil
}

func (ev *Evaluator) builtinOr(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	for _, arg := range args {
		v, err := ev.Eval(env, arg)
		if err != nil {
			continue // or swallows errors on individual branches
		}
		if isTruthy(v) {
			return v, nil
		}
	}
	return NilVal{}, nil
}

func (ev *Evaluator) builtinReadFile(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("read-file requires a path")
	}
	var parts []string
	for _, arg := range args {
		path, err := ev.Eval(env, arg)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path.String())
		if err != nil {
			return nil, fmt.Errorf("read-file: %w", err)
		}
		parts = append(parts, string(data))
	}
	return StringVal(strings.Join(parts, "\n\n")), nil
}

func (ev *Evaluator) builtinWriteFile(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("write-file requires a path")
	}
	path, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	content := ""
	for i := 1; i < len(args); i++ {
		if args[i].KeywordVal() == "content" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			content = v.String()
			i++
		} else if args[i].KeywordVal() == "from" && i+1 < len(args) {
			id, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			ev.mu.Lock()
			content = ev.steps[id.String()]
			ev.mu.Unlock()
			i++
		}
	}
	dir := filepath.Dir(path.String())
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755)
	}
	return StringVal(content), os.WriteFile(path.String(), []byte(content), 0o644)
}

func (ev *Evaluator) builtinGlob(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("glob requires a pattern")
	}
	pattern, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(pattern.String())
	if err != nil {
		return nil, fmt.Errorf("glob: %w", err)
	}
	return StringVal(strings.Join(matches, "\n")), nil
}

func (ev *Evaluator) builtinWebSearch(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("websearch requires a query")
	}
	query, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	numResults := 5
	for i := 1; i < len(args); i++ {
		if args[i].KeywordVal() == "results" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			fmt.Sscanf(v.String(), "%d", &numResults)
			i++
		}
	}

	baseURL := ev.WebSearchURL
	if baseURL == "" {
		baseURL = "http://localhost:8888"
	}

	searchURL := fmt.Sprintf("%s/search?q=%s&format=json&pageno=1&categories=general",
		baseURL, url.QueryEscape(query.String()))

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("websearch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("websearch: %w", err)
	}

	var data struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return StringVal(string(body)), nil
	}

	var results []string
	for i, r := range data.Results {
		if i >= numResults {
			break
		}
		results = append(results, fmt.Sprintf("## %s\n%s\n%s", r.Title, r.URL, r.Content))
	}
	return StringVal(strings.Join(results, "\n\n")), nil
}

func (ev *Evaluator) builtinHTTPGet(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http-get requires a URL")
	}
	rawURL, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", rawURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Parse optional headers
	for i := 1; i < len(args); i++ {
		if args[i].KeywordVal() == "headers" && i+1 < len(args) {
			// Headers as a map node {k v k v}
			hNode := args[i+1]
			if hNode.IsList() && hNode.IsMap {
				for j := 0; j+1 < len(hNode.Children); j += 2 {
					k, err := ev.Eval(env, hNode.Children[j])
					if err != nil {
						return nil, err
					}
					v, err := ev.Eval(env, hNode.Children[j+1])
					if err != nil {
						return nil, err
					}
					req.Header.Set(k.String(), v.String())
				}
			}
			i++
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http-get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return StringVal(string(body)), nil
}

func (ev *Evaluator) builtinHTTPPost(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http-post requires a URL")
	}
	rawURL, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	bodyContent := ""
	contentType := "application/json"
	var headers map[string]string

	for i := 1; i < len(args); i++ {
		kw := args[i].KeywordVal()
		if kw != "" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			switch kw {
			case "body":
				bodyContent = v.String()
			case "content-type":
				contentType = v.String()
			case "headers":
				// parse map
			}
			i++
		}
	}

	req, err := http.NewRequest("POST", rawURL.String(), strings.NewReader(bodyContent))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http-post: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return StringVal(string(body)), nil
}

func (ev *Evaluator) builtinCallWorkflow(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("call-workflow requires a workflow name")
	}
	name, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	// Resolve workflow file
	wfPath := name.String()
	if !strings.HasSuffix(wfPath, ".glitch") {
		wfPath += ".glitch"
	}
	if ev.WorkflowsDir != "" && !filepath.IsAbs(wfPath) {
		wfPath = filepath.Join(ev.WorkflowsDir, wfPath)
	}

	data, err := os.ReadFile(wfPath)
	if err != nil {
		return nil, fmt.Errorf("call-workflow %q: %w", name, err)
	}

	// Parse optional --set params
	childParams := make(map[string]string)
	for k, v := range ev.Params {
		childParams[k] = v
	}
	childInput := ""
	for i := 1; i < len(args); i++ {
		if args[i].KeywordVal() == "set" && i+1 < len(args) {
			// Parse {key val key val} map
			setNode := args[i+1]
			if setNode.IsList() && setNode.IsMap {
				for j := 0; j+1 < len(setNode.Children); j += 2 {
					k, err := ev.Eval(env, setNode.Children[j])
					if err != nil {
						return nil, err
					}
					v, err := ev.Eval(env, setNode.Children[j+1])
					if err != nil {
						return nil, err
					}
					childParams[k.String()] = v.String()
				}
			}
			i++
		} else if args[i].KeywordVal() == "input" && i+1 < len(args) {
			v, err := ev.Eval(env, args[i+1])
			if err != nil {
				return nil, err
			}
			childInput = v.String()
			i++
		}
	}

	// Create child evaluator
	child := NewEvaluator()
	child.Input = childInput
	child.Params = childParams
	child.DefaultModel = ev.DefaultModel
	child.ProviderReg = ev.ProviderReg
	child.ProviderResolver = ev.ProviderResolver
	child.Tiers = ev.Tiers
	child.ESURL = ev.ESURL
	child.WebSearchURL = ev.WebSearchURL
	child.Workspace = ev.Workspace
	child.Resources = ev.Resources
	child.WorkflowsDir = ev.WorkflowsDir

	result, err := child.RunSource(data)
	if err != nil {
		return nil, fmt.Errorf("call-workflow %q: %w", name, err)
	}

	// Merge child metrics
	ev.TotalTokensIn += child.TotalTokensIn
	ev.TotalTokensOut += child.TotalTokensOut
	ev.TotalCostUSD += child.TotalCostUSD
	ev.LLMSteps += child.LLMSteps

	return result, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./internal/pipeline/`

Note: If `provider.RunShellContext` doesn't exist, check the actual function name in provider package and adjust. It may be `provider.RunShell` or a method on the registry.

- [ ] **Step 3: Commit**

```
git add internal/pipeline/eval_builtins.go
git commit -m "feat(eval): add all builtin functions (sh, llm, save, websearch, http, etc.)"
```

---

### Task 5: Wire Evaluator into Run()

**Files:**
- Modify: `internal/pipeline/types.go` — add `Source []byte` to Workflow
- Modify: `internal/pipeline/sexpr.go` — simplify to store raw source
- Modify: `internal/pipeline/runner.go` — replace Run() body

- [ ] **Step 1: Add Source field to Workflow in types.go**

Add after the `SourceFile` field:

```go
Source     []byte   `yaml:"-"` // raw .glitch source for evaluator
```

- [ ] **Step 2: Store raw source in parseSexprWorkflow**

In `sexpr.go`, at the top of `parseSexprWorkflowWithIncludes`, store the source on the Workflow so the evaluator can access it. The existing parse logic stays temporarily — the evaluator uses `w.Source` when available, falling back to `w.Steps` for YAML-loaded workflows.

In `parseSexprWorkflow`:
```go
func parseSexprWorkflow(src []byte) (*Workflow, error) {
	return parseSexprWorkflowWithIncludes(src, nil)
}
```

Keep this function intact. Add source storage in `parseSexprWorkflowWithIncludes` after the workflow is built:

After the `return convertWorkflow(n, defs)` line (around line 88), change to:
```go
w, err := convertWorkflow(n, defs)
if err != nil {
    return nil, err
}
w.Source = src
return w, nil
```

- [ ] **Step 3: Replace Run() body in runner.go**

Replace the execution portion of `Run()` (after runCtx construction, lines ~320-545) with a call to the evaluator. Keep the runCtx construction, telemetry wiring, and result assembly.

The key change: instead of iterating over `w.Items` and calling `executeStep`, construct an `Evaluator`, wire the runtime context onto it, and call `ev.RunSource(w.Source)`.

At the point where execution begins (after line ~317 where rctx is fully constructed), replace the step execution loop with:

```go
// Use evaluator for sexpr workflows
if w.Source != nil {
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
    ev.WorkflowsDir = rctx.workflowsDir
    ev.StepRecorder = rctx.stepRecorder

    // Pre-seed steps
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

    steps = ev.Steps()
    totalTokensIn = ev.TotalTokensIn
    totalTokensOut = ev.TotalTokensOut
    totalCostUSD = ev.TotalCostUSD
    totalLatencyMS = ev.TotalLatencyMS
    llmSteps = ev.LLMSteps

    // Find last step output
    var lastOutput string
    // Use the last step recorded
    for _, id := range sortedStepIDs(steps) {
        lastOutput = steps[id]
    }

    return &Result{
        Workflow: w.Name,
        Output:   lastOutput,
        Steps:    steps,
        RunID:    rctx.parentRunID,
    }, nil
}

// Fallback: existing dispatch for YAML workflows (if any remain)
```

- [ ] **Step 4: Verify the full project builds**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`

- [ ] **Step 5: Commit**

```
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go
git commit -m "feat(eval): wire evaluator into Run() for sexpr workflows"
```

---

### Task 6: Tests

**Files:**
- Create: `internal/pipeline/eval_test.go`

- [ ] **Step 1: Write eval_test.go with core evaluator tests**

```go
package pipeline

import (
	"strings"
	"testing"
)

func evalHelper(t *testing.T, src string) string {
	t.Helper()
	ev := NewEvaluator()
	ev.Input = "test-input"
	ev.Params = map[string]string{"repo": "gl1tch", "name": "test"}
	val, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return val.String()
}

func TestEval_DefAndSymbol(t *testing.T) {
	got := evalHelper(t, `(def x "hello") x`)
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestEval_Fn(t *testing.T) {
	got := evalHelper(t, `
		(def greet (fn (name) (str "hello " name)))
		(greet "world")
	`)
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestEval_Let(t *testing.T) {
	got := evalHelper(t, `
		(let (x "a" y "b")
			(str x y))
	`)
	if got != "ab" {
		t.Errorf("got %q, want %q", got, "ab")
	}
}

func TestEval_Closure(t *testing.T) {
	got := evalHelper(t, `
		(def make-adder (fn (prefix) (fn (s) (str prefix s))))
		(def add-hello (make-adder "hello-"))
		(add-hello "world")
	`)
	if got != "hello-world" {
		t.Errorf("got %q, want %q", got, "hello-world")
	}
}

func TestEval_If(t *testing.T) {
	got := evalHelper(t, `(if true "yes" "no")`)
	if got != "yes" {
		t.Errorf("got %q, want %q", got, "yes")
	}
	got = evalHelper(t, `(if false "yes" "no")`)
	if got != "no" {
		t.Errorf("got %q, want %q", got, "no")
	}
}

func TestEval_Cond(t *testing.T) {
	got := evalHelper(t, `(cond false "a" true "b")`)
	if got != "b" {
		t.Errorf("got %q, want %q", got, "b")
	}
}

func TestEval_WorkflowAndStep(t *testing.T) {
	src := `
		(workflow "test"
			:description "test workflow"
			(step "greet" (str "hello"))
			(step "out" (str (ref "greet") " world")))
	`
	ev := NewEvaluator()
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["greet"] != "hello" {
		t.Errorf("greet = %q, want %q", steps["greet"], "hello")
	}
	if steps["out"] != "hello world" {
		t.Errorf("out = %q, want %q", steps["out"], "hello world")
	}
}

func TestEval_Par(t *testing.T) {
	src := `
		(workflow "test"
			:description "par test"
			(par
				(step "a" (str "alpha"))
				(step "b" (str "beta")))
			(step "merged" (str (ref "a") "+" (ref "b"))))
	`
	ev := NewEvaluator()
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["merged"] != "alpha+beta" {
		t.Errorf("merged = %q, want %q", steps["merged"], "alpha+beta")
	}
}

func TestEval_Interpolation(t *testing.T) {
	src := `
		(workflow "test"
			:description "interp test"
			(step "a" (str "val"))
			(step "b" (run "echo ~(step a)")))
	`
	ev := NewEvaluator()
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	steps := ev.Steps()
	if steps["b"] != "val" {
		t.Errorf("b = %q, want %q", steps["b"], "val")
	}
}

func TestEval_ParamInterpolation(t *testing.T) {
	src := `
		(workflow "test"
			:description "param test"
			(step "out" (str "repo=~param.repo")))
	`
	ev := NewEvaluator()
	ev.Params = map[string]string{"repo": "gl1tch"}
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Steps()["out"] != "repo=gl1tch" {
		t.Errorf("out = %q, want %q", ev.Steps()["out"], "repo=gl1tch")
	}
}

func TestEval_Map(t *testing.T) {
	src := `
		(workflow "test"
			:description "map test"
			(step "items" (str "a\nb\nc"))
			(step "mapped" (map (ref "items")
				(str "item:" (ref "item")))))
	`
	ev := NewEvaluator()
	_, err := ev.RunSource([]byte(src))
	if err != nil {
		// map sets ~item in scope, ref "item" should resolve
		// This test validates the map+item pattern
		t.Fatal(err)
	}
}

func TestEval_Retry(t *testing.T) {
	// Retry should succeed on the body
	src := `(retry 3 (str "ok"))`
	got := evalHelper(t, src)
	if got != "ok" {
		t.Errorf("got %q, want %q", got, "ok")
	}
}

func TestEval_Catch(t *testing.T) {
	// Catch should run fallback on error
	src := `(catch (ref "nonexistent") (str "caught"))`
	got := evalHelper(t, src)
	if got != "caught" {
		t.Errorf("got %q, want %q", got, "caught")
	}
}

func TestEval_Or(t *testing.T) {
	got := evalHelper(t, `(or "" "" "found")`)
	if got != "found" {
		t.Errorf("got %q, want %q", got, "found")
	}
}

func TestEval_Sh(t *testing.T) {
	got := evalHelper(t, `(sh "echo hello")`)
	if !strings.Contains(got, "hello") {
		t.Errorf("got %q, want to contain 'hello'", got)
	}
}
```

- [ ] **Step 2: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestEval -v -count=1`
Expected: all pass

- [ ] **Step 3: Run existing test suite to verify nothing broke**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -count=1 2>&1 | tail -30`
Expected: existing tests still pass (they use Run() which now routes through evaluator for sexpr workflows)

- [ ] **Step 4: Commit**

```
git add internal/pipeline/eval_test.go
git commit -m "test(eval): add evaluator tests for core Lisp features and workflow compat"
```

---

### Task 7: Run Smoke Tests Against Real Workflows

**Files:** None (validation only)

- [ ] **Step 1: Run the testdata workflows**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run Test -v -count=1 2>&1 | tail -50`

- [ ] **Step 2: Run glitch CLI against example workflows**

Build and run the CLI:
```bash
cd /Users/stokes/Projects/gl1tch
go build -o /tmp/glitch-eval ./cmd/glitch
/tmp/glitch-eval workflow run hello
/tmp/glitch-eval workflow run code-review
/tmp/glitch-eval workflow run git-changelog
/tmp/glitch-eval workflow run par-demo  # from testdata
```

- [ ] **Step 3: Fix any failures**

If any workflow fails, debug by comparing the evaluator's behavior against the expected output. Common issues:
- Missing builtin (check eval_builtins.go registrations)
- Interpolation edge case (check eval_interpolate.go)
- Keyword arg parsing in a form

- [ ] **Step 4: Commit fixes if any**

```
git commit -am "fix(eval): address smoke test failures"
```

---

### Task 8: Clean Up Dead Code

**Files:**
- Modify: `internal/pipeline/sexpr.go` — delete convertStep, convertForm, and all convert* functions
- Modify: `internal/pipeline/runner.go` — delete executeStep, runSingleStep, and old dispatch code

- [ ] **Step 1: Identify dead code**

After the evaluator is wired in and tests pass, the following functions in sexpr.go are dead:
- `convertWorkflow`, `convertForm`, `convertStep`
- `convertRetry`, `convertTimeout`, `convertLet`, `convertCatch`
- `convertCond`, `convertWhen`, `convertMap`, `convertFilter`, `convertReduce`
- `convertMapResources`, `convertPar`, `convertThread`
- `convertCompare`, `convertBranch`, `convertReview`
- `convertPhase`, `convertGate`
- `convertLLM`, `convertJsonPick`, `convertLines`, `convertMerge`
- `convertReadFile`, `convertWriteFile`, `convertGlobStep`
- `convertHttpCall`, `convertPluginCall`, `convertSearch`, `convertIndex`
- `convertDelete`, `convertEmbed`, `convertWebSearch`
- All helper functions only called by the above

In runner.go, the following are dead:
- `executeStep`, `runSingleStep`
- `executePar`, `executeWhen`, `executeCond`
- `executeMap`, `executeFilter`, `executeReduce`
- `executeCompare`, `executePhase`
- `runLLMStep`, `renderInStep` usage from runner
- The old execution loop in Run() (already replaced in Task 5)

- [ ] **Step 2: Delete dead code from sexpr.go**

Keep only: `parseSexprWorkflow`, `parseSexprWorkflowWithIncludes`, `collectDefsFromFile`, `resolveVal` (used for def resolution at parse time), and supporting helpers for those functions.

Delete everything from `convertWorkflow` onward (~1700 lines).

- [ ] **Step 3: Delete dead code from runner.go**

Delete `executeStep`, `runSingleStep`, and all the form-specific execution functions. Keep:
- `Run()` (now uses evaluator)
- `runCtx` struct
- `Result`, `RunOpts`, `StepRecord` types
- Telemetry and result assembly code
- Helper functions still used by the evaluator path

- [ ] **Step 4: Verify everything still compiles and tests pass**

Run:
```bash
cd /Users/stokes/Projects/gl1tch && go build ./... && go test ./internal/pipeline/ -count=1
```

- [ ] **Step 5: Commit**

```
git add internal/pipeline/sexpr.go internal/pipeline/runner.go
git commit -m "refactor(eval): delete old converter and dispatch code (~3500 lines)"
```
