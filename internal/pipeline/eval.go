package pipeline

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Evaluator walks an s-expression AST and evaluates it directly.
type Evaluator struct {
	mu    sync.Mutex
	steps map[string]string

	// Runtime context
	Input        string
	DefaultModel string
	Workspace    string
	WorkflowName string
	Params       map[string]string
	Resources    map[string]map[string]string

	// Provider integration
	ProviderReg      *provider.ProviderRegistry
	ProviderResolver provider.ResolverFunc
	Tiers            []provider.TierConfig
	EvalThreshold    int

	// External URLs
	ESURL        string
	WebSearchURL string
	WorkflowsDir string

	// Step recording callback
	StepRecorder func(rec StepRecord)

	// LLM metrics accumulators
	TotalTokensIn  int64
	TotalTokensOut int64
	TotalLatencyMS int64
	TotalCostUSD   float64
	LLMSteps       int
}

// NewEvaluator creates a new evaluator with default values.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		steps:     make(map[string]string),
		Params:    make(map[string]string),
		Resources: make(map[string]map[string]string),
	}
}

// Steps returns a snapshot of the current step outputs.
func (ev *Evaluator) Steps() map[string]string {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	out := make(map[string]string, len(ev.steps))
	for k, v := range ev.steps {
		out[k] = v
	}
	return out
}

// RunSource parses source bytes and evaluates all top-level forms.
func (ev *Evaluator) RunSource(src []byte) (Value, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	env := NewEnv(nil)
	ev.registerBuiltins(env)

	var result Value = NilVal{}
	for _, n := range nodes {
		result, err = ev.Eval(env, n)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Eval is the main dispatch for evaluating a single AST node.
func (ev *Evaluator) Eval(env *Env, node *sexpr.Node) (Value, error) {
	if node.IsAtom() {
		return ev.evalAtom(env, node)
	}
	if !node.IsList() || len(node.Children) == 0 {
		return NilVal{}, nil
	}

	head := node.Children[0]
	args := node.Children[1:]

	// Check for special forms (head is a symbol)
	if head.IsAtom() && head.Atom.Type == sexpr.TokenSymbol {
		name := head.Atom.Val
		if val, err, handled := ev.evalSpecial(env, name, args, node); handled {
			return val, err
		}
	}

	// Regular function call: evaluate head, then apply.
	headVal, err := ev.Eval(env, head)
	if err != nil {
		return nil, fmt.Errorf("line %d: %w", node.Line, err)
	}
	return ev.apply(env, headVal, args, node)
}

// evalAtom evaluates an atom node.
func (ev *Evaluator) evalAtom(env *Env, node *sexpr.Node) (Value, error) {
	tok := node.Atom
	switch tok.Type {
	case sexpr.TokenString:
		// Check for ~ interpolation
		if strings.ContainsRune(tok.Val, '~') {
			s, err := ev.interpolate(env, tok.Val)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", node.Line, err)
			}
			return StringVal(s), nil
		}
		return StringVal(tok.Val), nil

	case sexpr.TokenKeyword:
		return StringVal(node.KeywordVal()), nil

	case sexpr.TokenSymbol:
		sym := tok.Val
		switch sym {
		case "true":
			return BoolVal(true), nil
		case "false":
			return BoolVal(false), nil
		case "nil":
			return NilVal{}, nil
		}
		if v, ok := env.Get(sym); ok {
			return v, nil
		}
		return nil, fmt.Errorf("line %d: undefined symbol %q", node.Line, sym)
	}
	return NilVal{}, nil
}

// evalSpecial dispatches special forms. Returns (value, error, handled).
func (ev *Evaluator) evalSpecial(env *Env, name string, args []*sexpr.Node, node *sexpr.Node) (Value, error, bool) {
	switch name {
	case "def":
		v, err := ev.specialDef(env, args, node)
		return v, err, true
	case "fn":
		v, err := ev.specialFn(env, args, node)
		return v, err, true
	case "let":
		v, err := ev.specialLet(env, args, node)
		return v, err, true
	case "do", "begin":
		v, err := ev.specialDo(env, args)
		return v, err, true
	case "if":
		v, err := ev.specialIf(env, args, node)
		return v, err, true
	case "when":
		v, err := ev.specialWhen(env, args, false)
		return v, err, true
	case "when-not":
		v, err := ev.specialWhen(env, args, true)
		return v, err, true
	case "cond":
		v, err := ev.specialCond(env, args)
		return v, err, true
	case "workflow":
		v, err := ev.specialWorkflow(env, args, node)
		return v, err, true
	case "step":
		v, err := ev.specialStep(env, args, node)
		return v, err, true
	case "par":
		v, err := ev.specialPar(env, args)
		return v, err, true
	case "retry":
		v, err := ev.specialRetry(env, args, node)
		return v, err, true
	case "timeout":
		v, err := ev.specialTimeout(env, args, node)
		return v, err, true
	case "catch":
		v, err := ev.specialCatch(env, args, node)
		return v, err, true
	case "include":
		v, err := ev.specialInclude(env, args, node)
		return v, err, true
	case "map", "each":
		v, err := ev.specialMap(env, args, node)
		return v, err, true
	case "filter":
		v, err := ev.specialFilter(env, args, node)
		return v, err, true
	case "reduce":
		v, err := ev.specialReduce(env, args, node)
		return v, err, true
	case "compare":
		v, err := ev.specialCompare(env, args, node)
		return v, err, true
	case "phase":
		v, err := ev.specialPhase(env, args, node)
		return v, err, true
	case "gate":
		v, err := ev.specialStep(env, args, node)
		return v, err, true
	case "->":
		v, err := ev.specialThread(env, args)
		return v, err, true
	}
	return nil, nil, false
}

// specialDef: (def name val)
func (ev *Evaluator) specialDef(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: def requires name and value", node.Line)
	}
	name := args[0].SymbolVal()
	if name == "" {
		name = args[0].StringVal()
	}
	if name == "" {
		return nil, fmt.Errorf("line %d: def: first arg must be a symbol", node.Line)
	}
	val, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	env.Set(name, val)
	return val, nil
}

// specialFn: (fn (params...) body...)
func (ev *Evaluator) specialFn(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: fn requires params and body", node.Line)
	}
	paramNode := args[0]
	if !paramNode.IsList() {
		return nil, fmt.Errorf("line %d: fn: first arg must be a parameter list", node.Line)
	}
	params := make([]string, len(paramNode.Children))
	for i, p := range paramNode.Children {
		s := p.SymbolVal()
		if s == "" {
			s = p.StringVal()
		}
		params[i] = s
	}
	return &FnVal{
		Params: params,
		Body:   args[1:],
		Env:    env,
	}, nil
}

// specialLet: (let (x 1 y 2) body...)
func (ev *Evaluator) specialLet(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: let requires bindings and body", node.Line)
	}
	bindings := args[0]
	if !bindings.IsList() {
		return nil, fmt.Errorf("line %d: let: first arg must be a binding list", node.Line)
	}
	if len(bindings.Children)%2 != 0 {
		return nil, fmt.Errorf("line %d: let: bindings must be even (name value pairs)", node.Line)
	}
	child := NewEnv(env)
	for i := 0; i < len(bindings.Children); i += 2 {
		name := bindings.Children[i].SymbolVal()
		if name == "" {
			name = bindings.Children[i].StringVal()
		}
		val, err := ev.Eval(child, bindings.Children[i+1])
		if err != nil {
			return nil, err
		}
		child.Set(name, val)
	}
	return ev.specialDo(child, args[1:])
}

// specialDo: (do body...) — evaluate sequentially, return last.
func (ev *Evaluator) specialDo(env *Env, args []*sexpr.Node) (Value, error) {
	var result Value = NilVal{}
	var err error
	for _, a := range args {
		result, err = ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// specialIf: (if pred then else?)
func (ev *Evaluator) specialIf(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: if requires predicate and then branch", node.Line)
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

// specialWhen: (when pred body...) / (when-not pred body...)
func (ev *Evaluator) specialWhen(env *Env, args []*sexpr.Node, negate bool) (Value, error) {
	if len(args) < 2 {
		return NilVal{}, nil
	}
	pred, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	truth := isTruthy(pred)
	if negate {
		truth = !truth
	}
	if truth {
		return ev.specialDo(env, args[1:])
	}
	return NilVal{}, nil
}

// specialCond: (cond pred1 body1 pred2 body2 ... else bodyN)
func (ev *Evaluator) specialCond(env *Env, args []*sexpr.Node) (Value, error) {
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

// specialWorkflow: (workflow name :description "..." body...)
func (ev *Evaluator) specialWorkflow(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return NilVal{}, nil
	}

	// First arg is the workflow name
	nameNode := args[0]
	name := nameNode.SymbolVal()
	if name == "" {
		name = nameNode.StringVal()
	}
	if name != "" {
		ev.WorkflowName = name
	}

	// Parse keyword metadata, collect body forms
	body := make([]*sexpr.Node, 0, len(args))
	for i := 1; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			// Keyword: consume next arg as value (skip both)
			i++ // skip the keyword's value
			continue
		}
		body = append(body, n)
	}

	return ev.specialDo(env, body)
}

// specialStep: (step id body...) or (step id) for lookup
func (ev *Evaluator) specialStep(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("line %d: step requires an id", node.Line)
	}

	id := args[0].SymbolVal()
	if id == "" {
		id = args[0].StringVal()
	}
	if id == "" {
		// Evaluate to get the id
		v, err := ev.Eval(env, args[0])
		if err != nil {
			return nil, err
		}
		id = v.String()
	}

	// Single-arg form: lookup
	if len(args) == 1 {
		ev.mu.Lock()
		v, ok := ev.steps[id]
		ev.mu.Unlock()
		if !ok {
			return NilVal{}, nil
		}
		return StringVal(v), nil
	}

	// Multi-arg form: evaluate body, record output
	start := time.Now()
	result, err := ev.specialDo(env, args[1:])
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", id, err)
	}

	output := result.String()
	ev.mu.Lock()
	ev.steps[id] = output
	ev.mu.Unlock()

	// Record step
	if ev.StepRecorder != nil {
		dur := time.Since(start)
		ev.StepRecorder(StepRecord{
			StepID:     id,
			Output:     output,
			DurationMs: dur.Milliseconds(),
		})
	}

	return result, nil
}

// specialPar: (par form...) — run forms concurrently
func (ev *Evaluator) specialPar(env *Env, args []*sexpr.Node) (Value, error) {
	type parResult struct {
		idx int
		val Value
		err error
	}

	results := make([]parResult, len(args))
	var wg sync.WaitGroup
	wg.Add(len(args))

	for i, arg := range args {
		go func(idx int, node *sexpr.Node) {
			defer wg.Done()
			v, err := ev.Eval(env, node)
			results[idx] = parResult{idx: idx, val: v, err: err}
		}(i, arg)
	}
	wg.Wait()

	vals := make(ListVal, 0, len(args))
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		vals = append(vals, r.val)
	}
	return vals, nil
}

// specialRetry: (retry N body...)
func (ev *Evaluator) specialRetry(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: retry requires count and body", node.Line)
	}

	countVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	maxRetries, _ := strconv.Atoi(countVal.String())
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := ev.specialDo(env, args[1:])
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("retry exhausted after %d attempts: %w", maxRetries+1, lastErr)
}

// specialTimeout: (timeout "30s" body...)
func (ev *Evaluator) specialTimeout(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: timeout requires duration and body", node.Line)
	}

	durVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	dur, err := time.ParseDuration(durVal.String())
	if err != nil {
		return nil, fmt.Errorf("line %d: timeout: invalid duration %q: %w", node.Line, durVal.String(), err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	type result struct {
		val Value
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := ev.specialDo(env, args[1:])
		ch <- result{v, err}
	}()

	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return nil, fmt.Errorf("line %d: timeout after %s", node.Line, dur)
	}
}

// specialCatch: (catch body fallback)
func (ev *Evaluator) specialCatch(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: catch requires body and fallback", node.Line)
	}

	result, err := ev.Eval(env, args[0])
	if err == nil {
		return result, nil
	}

	// On error, bind "error" in a child env and run the fallback
	child := NewEnv(env)
	child.Set("error", StringVal(err.Error()))
	return ev.Eval(child, args[1])
}

// specialInclude: (include "path")
func (ev *Evaluator) specialInclude(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("line %d: include requires a file path", node.Line)
	}

	pathVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	path := pathVal.String()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("line %d: include %q: %w", node.Line, path, err)
	}

	nodes, err := sexpr.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("include %q: parse: %w", path, err)
	}

	var result Value = NilVal{}
	for _, n := range nodes {
		result, err = ev.Eval(env, n)
		if err != nil {
			return nil, fmt.Errorf("include %q: %w", path, err)
		}
	}
	return result, nil
}

// specialMap: (map source body...) — split source by newlines, eval body per item
func (ev *Evaluator) specialMap(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: map requires source and body", node.Line)
	}

	sourceVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	items := strings.Split(sourceVal.String(), "\n")

	results := make(ListVal, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		child := NewEnv(env)
		child.Set("item", StringVal(item))
		val, err := ev.specialDo(child, args[1:])
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return results, nil
}

// specialFilter: (filter source pred) — keep items where pred is truthy
func (ev *Evaluator) specialFilter(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("line %d: filter requires source and predicate", node.Line)
	}

	sourceVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	items := strings.Split(sourceVal.String(), "\n")

	results := make(ListVal, 0)
	for _, item := range items {
		if item == "" {
			continue
		}
		child := NewEnv(env)
		child.Set("item", StringVal(item))
		pred, err := ev.Eval(child, args[1])
		if err != nil {
			return nil, err
		}
		if isTruthy(pred) {
			results = append(results, StringVal(item))
		}
	}
	return results, nil
}

// specialReduce: (reduce source initial body) — accumulate with "item" and "acc" bound
func (ev *Evaluator) specialReduce(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("line %d: reduce requires source, initial, and body", node.Line)
	}

	sourceVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	items := strings.Split(sourceVal.String(), "\n")

	acc, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item == "" {
			continue
		}
		child := NewEnv(env)
		child.Set("item", StringVal(item))
		child.Set("acc", acc)
		acc, err = ev.Eval(child, args[2])
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

// specialCompare: (compare :id "x" :objective "..." (branch "name" steps...) (review ...))
func (ev *Evaluator) specialCompare(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	var id, objective string
	var branches []struct {
		name  string
		steps []*sexpr.Node
	}
	var reviewModel string
	var reviewCriteria []string

	// Parse keyword args and nested forms
	for i := 0; i < len(args); i++ {
		n := args[i]

		// Keyword args
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				val, err := ev.Eval(env, args[i])
				if err != nil {
					return nil, err
				}
				switch kw {
				case "id":
					id = val.String()
				case "objective":
					objective = val.String()
				}
			}
			continue
		}

		// Nested forms
		if n.IsList() && len(n.Children) > 0 {
			head := n.Children[0].SymbolVal()
			switch head {
			case "branch":
				if len(n.Children) >= 2 {
					bname := n.Children[1].StringVal()
					if bname == "" {
						bname = n.Children[1].SymbolVal()
					}
					branches = append(branches, struct {
						name  string
						steps []*sexpr.Node
					}{name: bname, steps: n.Children[2:]})
				}
			case "review":
				// Parse review keywords
				for j := 1; j < len(n.Children); j++ {
					c := n.Children[j]
					if c.IsAtom() && c.Atom.Type == sexpr.TokenKeyword {
						rkw := c.KeywordVal()
						if j+1 < len(n.Children) {
							j++
							switch rkw {
							case "model":
								rv, err := ev.Eval(env, n.Children[j])
								if err != nil {
									return nil, err
								}
								reviewModel = rv.String()
							case "criteria":
								cn := n.Children[j]
								if cn.IsList() {
									for _, cc := range cn.Children {
										s := cc.StringVal()
										if s == "" {
											s = cc.SymbolVal()
										}
										if s != "" {
											reviewCriteria = append(reviewCriteria, s)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	_ = id // used for recording

	// Run branches in parallel
	type branchResult struct {
		name   string
		output string
		err    error
	}
	branchResults := make([]branchResult, len(branches))
	var wg sync.WaitGroup
	wg.Add(len(branches))
	for i, br := range branches {
		go func(idx int, b struct {
			name  string
			steps []*sexpr.Node
		}) {
			defer wg.Done()
			child := NewEnv(env)
			val, err := ev.specialDo(child, b.steps)
			branchResults[idx] = branchResult{name: b.name, output: "", err: err}
			if err == nil {
				branchResults[idx].output = val.String()
			}
		}(i, br)
	}
	wg.Wait()

	// Check for errors
	for _, r := range branchResults {
		if r.err != nil {
			return nil, fmt.Errorf("compare branch %q: %w", r.name, r.err)
		}
	}

	// Build judge prompt
	if ev.ProviderReg == nil {
		// No provider: return concatenated outputs
		var sb strings.Builder
		for i, r := range branchResults {
			if i > 0 {
				sb.WriteString("\n---\n")
			}
			sb.WriteString(fmt.Sprintf("[%s]\n%s", r.name, r.output))
		}
		return StringVal(sb.String()), nil
	}

	// Build judge prompt
	var judgePrompt strings.Builder
	judgePrompt.WriteString(fmt.Sprintf("Objective: %s\n\n", objective))
	if len(reviewCriteria) > 0 {
		judgePrompt.WriteString("Criteria: " + strings.Join(reviewCriteria, ", ") + "\n\n")
	}
	for _, r := range branchResults {
		judgePrompt.WriteString(fmt.Sprintf("--- Branch: %s ---\n%s\n\n", r.name, r.output))
	}
	judgePrompt.WriteString("Compare the branches above against the objective. Which is best and why?")

	model := reviewModel
	if model == "" {
		model = ev.DefaultModel
	}

	result, err := ev.ProviderReg.RunProviderWithResult("ollama", model, judgePrompt.String())
	if err != nil {
		return nil, fmt.Errorf("compare review: %w", err)
	}

	ev.mu.Lock()
	ev.TotalTokensIn += int64(result.TokensIn)
	ev.TotalTokensOut += int64(result.TokensOut)
	ev.TotalLatencyMS += result.Latency.Milliseconds()
	ev.TotalCostUSD += result.CostUSD
	ev.LLMSteps++
	ev.mu.Unlock()

	return StringVal(result.Response), nil
}

// specialPhase: (phase :id "x" :retries N body...) — retriable unit
func (ev *Evaluator) specialPhase(env *Env, args []*sexpr.Node, node *sexpr.Node) (Value, error) {
	var retries int
	body := make([]*sexpr.Node, 0, len(args))

	for i := 0; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				switch kw {
				case "retries":
					rv, err := ev.Eval(env, args[i])
					if err != nil {
						return nil, err
					}
					retries, _ = strconv.Atoi(rv.String())
				case "id":
					// consume but don't need the value for execution
				}
			}
			continue
		}
		body = append(body, n)
	}

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		result, err := ev.specialDo(env, body)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("phase exhausted after %d attempts: %w", retries+1, lastErr)
}

// specialThread: (-> expr1 expr2...) — result of each feeds as "_" binding to next
func (ev *Evaluator) specialThread(env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return NilVal{}, nil
	}

	result, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	for _, arg := range args[1:] {
		child := NewEnv(env)
		child.Set("_", result)
		result, err = ev.Eval(child, arg)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// apply calls a function value with arguments.
func (ev *Evaluator) apply(env *Env, fn Value, argNodes []*sexpr.Node, node *sexpr.Node) (Value, error) {
	switch f := fn.(type) {
	case *BuiltinVal:
		return f.Fn(ev, env, argNodes)
	case *FnVal:
		return ev.applyFn(env, f, argNodes)
	default:
		return nil, fmt.Errorf("line %d: cannot call %T", node.Line, fn)
	}
}

// applyFn evaluates args in the caller env, binds them in a child of the closure env, and runs the body.
func (ev *Evaluator) applyFn(callerEnv *Env, fn *FnVal, argNodes []*sexpr.Node) (Value, error) {
	if len(argNodes) != len(fn.Params) {
		return nil, fmt.Errorf("fn expects %d args, got %d", len(fn.Params), len(argNodes))
	}

	// Evaluate args in the caller's environment
	argVals := make([]Value, len(argNodes))
	for i, a := range argNodes {
		v, err := ev.Eval(callerEnv, a)
		if err != nil {
			return nil, err
		}
		argVals[i] = v
	}

	// Bind in a child of the closure's captured environment
	child := NewEnv(fn.Env)
	for i, p := range fn.Params {
		child.Set(p, argVals[i])
	}

	return ev.specialDo(child, fn.Body)
}
