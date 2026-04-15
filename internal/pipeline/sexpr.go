package pipeline

import (
	"fmt"
	"strconv"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// parseSexprWorkflow parses s-expression source into a Workflow.
func parseSexprWorkflow(src []byte) (*Workflow, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}

	// Collect defs
	defs := make(map[string]string)
	for _, n := range nodes {
		if n.IsList() && len(n.Children) >= 3 && n.Children[0].SymbolVal() == "def" {
			name := n.Children[1].SymbolVal()
			if name == "" {
				name = n.Children[1].StringVal()
			}
			val := resolveVal(n.Children[2], defs)
			defs[name] = val
		}
	}

	// Find workflow
	for _, n := range nodes {
		if n.IsList() && len(n.Children) > 0 && n.Children[0].StringVal() == "workflow" {
			return convertWorkflow(n, defs)
		}
	}
	return nil, fmt.Errorf("no (workflow ...) form found")
}

// resolveVal returns the string value of a node, substituting def bindings for symbols.
func resolveVal(n *sexpr.Node, defs map[string]string) string {
	if n.Atom != nil && n.Atom.Type == sexpr.TokenSymbol {
		if v, ok := defs[n.Atom.Val]; ok {
			return v
		}
	}
	return n.StringVal()
}

func convertWorkflow(n *sexpr.Node, defs map[string]string) (*Workflow, error) {
	children := n.Children[1:] // skip "workflow" symbol
	if len(children) == 0 {
		return nil, fmt.Errorf("line %d: workflow missing name", n.Line)
	}

	w := &Workflow{}

	// First child must be the name
	w.Name = children[0].StringVal()
	if w.Name == "" {
		return nil, fmt.Errorf("line %d: workflow name must be a string", children[0].Line)
	}
	children = children[1:]

	// Process remaining children: keywords for metadata, lists for steps/forms
	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "description":
				w.Description = resolveVal(val, defs)
			default:
				return nil, fmt.Errorf("line %d: unknown workflow keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].SymbolVal()
			if head == "" {
				head = child.Children[0].StringVal()
			}
			steps, err := convertForm(child, head, defs)
			if err != nil {
				return nil, err
			}
			w.Steps = append(w.Steps, steps...)
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in workflow", child.Line)
	}
	return w, nil
}

// convertForm dispatches a workflow-level list form to the appropriate converter.
func convertForm(n *sexpr.Node, head string, defs map[string]string) ([]Step, error) {
	switch head {
	case "step":
		s, err := convertStep(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "retry":
		s, err := convertRetry(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "timeout":
		s, err := convertTimeout(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "let":
		return convertLet(n, defs)
	case "catch":
		s, err := convertCatch(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "cond":
		s, err := convertCond(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "map":
		s, err := convertMap(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	default:
		return nil, fmt.Errorf("line %d: unknown form %q", n.Line, head)
	}
}

// convertRetry: (retry N (step ...)) or (retry N (timeout ...)) etc.
func convertRetry(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (retry) needs count and step", n.Line)
	}
	count, err := strconv.Atoi(resolveVal(children[0], defs))
	if err != nil {
		return Step{}, fmt.Errorf("line %d: (retry) count must be an integer", children[0].Line)
	}
	inner := children[1]
	head := inner.Children[0].SymbolVal()
	if head == "" {
		head = inner.Children[0].StringVal()
	}
	steps, err := convertForm(inner, head, defs)
	if err != nil {
		return Step{}, err
	}
	if len(steps) != 1 {
		return Step{}, fmt.Errorf("line %d: (retry) inner form must produce exactly one step", inner.Line)
	}
	steps[0].Retry = count
	return steps[0], nil
}

// convertTimeout: (timeout "30s" (step ...)) or (timeout "30s" (retry ...)) etc.
func convertTimeout(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (timeout) needs duration and step", n.Line)
	}
	dur := resolveVal(children[0], defs)
	inner := children[1]
	head := inner.Children[0].SymbolVal()
	if head == "" {
		head = inner.Children[0].StringVal()
	}
	steps, err := convertForm(inner, head, defs)
	if err != nil {
		return Step{}, err
	}
	if len(steps) != 1 {
		return Step{}, fmt.Errorf("line %d: (timeout) inner form must produce exactly one step", inner.Line)
	}
	steps[0].Timeout = dur
	return steps[0], nil
}

// convertLet: (let ((name val) ...) body...)
// Extends defs for the scope of the body, returns all body steps.
func convertLet(n *sexpr.Node, defs map[string]string) ([]Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return nil, fmt.Errorf("line %d: (let) needs bindings and body", n.Line)
	}
	bindings := children[0]
	if !bindings.IsList() {
		return nil, fmt.Errorf("line %d: (let) bindings must be a list", bindings.Line)
	}

	// Copy defs so let bindings are scoped
	scoped := make(map[string]string, len(defs))
	for k, v := range defs {
		scoped[k] = v
	}
	for _, b := range bindings.Children {
		if !b.IsList() || len(b.Children) < 2 {
			return nil, fmt.Errorf("line %d: let binding must be (name value)", b.Line)
		}
		name := b.Children[0].SymbolVal()
		if name == "" {
			name = b.Children[0].StringVal()
		}
		scoped[name] = resolveVal(b.Children[1], scoped)
	}

	// Convert body forms using scoped defs
	var steps []Step
	for _, child := range children[1:] {
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].SymbolVal()
			if head == "" {
				head = child.Children[0].StringVal()
			}
			bodySteps, err := convertForm(child, head, scoped)
			if err != nil {
				return nil, err
			}
			steps = append(steps, bodySteps...)
		}
	}
	return steps, nil
}

// convertCatch: (catch (step "primary" ...) (step "fallback" ...))
func convertCatch(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (catch) needs primary and fallback steps", n.Line)
	}
	primary, err := convertStep(children[0], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: catch primary: %w", n.Line, err)
	}
	fallback, err := convertStep(children[1], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: catch fallback: %w", n.Line, err)
	}
	primary.Form = "catch"
	primary.Fallback = &fallback
	return primary, nil
}

// convertCond: (cond (pred (step ...)) ... (else (step ...)))
func convertCond(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) == 0 {
		return Step{}, fmt.Errorf("line %d: (cond) needs at least one branch", n.Line)
	}

	s := Step{
		ID:   fmt.Sprintf("cond-%d", n.Line),
		Form: "cond",
	}

	for _, branch := range children {
		if !branch.IsList() || len(branch.Children) < 2 {
			return Step{}, fmt.Errorf("line %d: cond branch must be (predicate (step ...))", branch.Line)
		}
		pred := branch.Children[0]
		stepNode := branch.Children[1]

		var predStr string
		if pred.SymbolVal() == "else" {
			predStr = "else"
		} else {
			predStr = resolveVal(pred, defs)
		}

		step, err := convertStep(stepNode, defs)
		if err != nil {
			return Step{}, fmt.Errorf("line %d: cond branch step: %w", branch.Line, err)
		}
		s.Branches = append(s.Branches, CondBranch{Pred: predStr, Step: step})
	}
	return s, nil
}

// convertMap: (map "step-id" (step ...))
func convertMap(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (map) needs source step ID and body step", n.Line)
	}
	source := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: map body: %w", n.Line, err)
	}
	return Step{
		ID:      fmt.Sprintf("map-%d", n.Line),
		Form:    "map",
		MapOver: source,
		MapBody: &body,
	}, nil
}

func convertStep(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:] // skip "step"
	if len(children) == 0 {
		return Step{}, fmt.Errorf("line %d: step missing id", n.Line)
	}

	s := Step{}
	s.ID = children[0].StringVal()
	if s.ID == "" {
		return s, fmt.Errorf("line %d: step id must be a string", children[0].Line)
	}
	children = children[1:]

	for _, child := range children {
		if !child.IsList() || len(child.Children) == 0 {
			return s, fmt.Errorf("line %d: step body must be (run ...), (llm ...), or (save ...)", child.Line)
		}
		head := child.Children[0].StringVal()
		switch head {
		case "run":
			if len(child.Children) < 2 {
				return s, fmt.Errorf("line %d: (run) missing command", child.Line)
			}
			s.Run = resolveVal(child.Children[1], defs)
		case "llm":
			llm, err := convertLLM(child, defs)
			if err != nil {
				return s, err
			}
			s.LLM = llm
		case "save":
			if len(child.Children) < 2 {
				return s, fmt.Errorf("line %d: (save) missing path", child.Line)
			}
			s.Save = resolveVal(child.Children[1], defs)
			// Check for :from keyword
			rest := child.Children[2:]
			for j := 0; j < len(rest); j++ {
				if rest[j].IsAtom() && rest[j].Atom.Type == sexpr.TokenKeyword && rest[j].KeywordVal() == "from" {
					j++
					if j < len(rest) {
						s.SaveStep = resolveVal(rest[j], defs)
					}
				}
			}
		default:
			return s, fmt.Errorf("line %d: unknown step type %q", child.Line, head)
		}
	}
	return s, nil
}

func convertLLM(n *sexpr.Node, defs map[string]string) (*LLMStep, error) {
	children := n.Children[1:] // skip "llm"
	llm := &LLMStep{}

	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "prompt":
				llm.Prompt = resolveVal(val, defs)
			case "provider":
				llm.Provider = resolveVal(val, defs)
			case "model":
				llm.Model = resolveVal(val, defs)
			case "tier":
				n := 0
				valStr := resolveVal(val, defs)
				fmt.Sscanf(valStr, "%d", &n)
				llm.Tier = &n
			case "skill":
				llm.Skill = resolveVal(val, defs)
			case "format":
				llm.Format = resolveVal(val, defs)
			default:
				return nil, fmt.Errorf("line %d: unknown llm keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in llm", child.Line)
	}
	if llm.Prompt == "" {
		return nil, fmt.Errorf("line %d: llm missing :prompt", n.Line)
	}
	return llm, nil
}
