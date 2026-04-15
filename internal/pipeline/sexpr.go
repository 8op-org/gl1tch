package pipeline

import (
	"fmt"
	"strconv"
	"strings"

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
	w.Tags = []string{}

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
			case "author":
				w.Author = resolveVal(val, defs)
			case "version":
				w.Version = resolveVal(val, defs)
			case "created":
				w.Created = resolveVal(val, defs)
			case "tags":
				if val.IsList() {
					for _, t := range val.Children {
						w.Tags = append(w.Tags, resolveVal(t, defs))
					}
				} else {
					w.Tags = append(w.Tags, resolveVal(val, defs))
				}
			case "action":
				w.Actions = append(w.Actions, resolveVal(val, defs))
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
			if head == "phase" {
				p, err := convertPhase(child, defs)
				if err != nil {
					return nil, err
				}
				w.Items = append(w.Items, WorkflowItem{Phase: &p})
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
				w.Items = append(w.Items, WorkflowItem{Step: &w.Steps[len(w.Steps)-len(steps)+idx]})
			}
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
	case "par":
		s, err := convertPar(n, defs)
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

// convertPar: (par (step ...) (step ...) ...)
// All children run concurrently. Minimum 2 children.
func convertPar(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:] // skip "par"
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (par) needs at least 2 children", n.Line)
	}

	var parSteps []Step
	for _, child := range children {
		if !child.IsList() || len(child.Children) == 0 {
			return Step{}, fmt.Errorf("line %d: (par) children must be forms", child.Line)
		}
		head := child.Children[0].SymbolVal()
		if head == "" {
			head = child.Children[0].StringVal()
		}
		if head == "gate" {
			g, err := convertGate(child, defs)
			if err != nil {
				return Step{}, fmt.Errorf("line %d: par child: %w", child.Line, err)
			}
			parSteps = append(parSteps, g)
		} else {
			steps, err := convertForm(child, head, defs)
			if err != nil {
				return Step{}, fmt.Errorf("line %d: par child: %w", child.Line, err)
			}
			parSteps = append(parSteps, steps...)
		}
	}

	return Step{
		ID:       fmt.Sprintf("par-%d", n.Line),
		Form:     "par",
		ParSteps: parSteps,
	}, nil
}

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

	i := 0
	for i < len(children) {
		child := children[i]
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
			case "par":
				parStep, err := convertPar(child, defs)
				if err != nil {
					return Phase{}, err
				}
				// All children must be the same kind (all gates or all steps).
				allGates := true
				allSteps := true
				for i := range parStep.ParSteps {
					if parStep.ParSteps[i].IsGate {
						allSteps = false
					} else {
						allGates = false
					}
				}
				if !allGates && !allSteps {
					return Phase{}, fmt.Errorf("line %d: (par) inside phase must contain all gates or all steps, not a mix", child.Line)
				}
				parStep.IsGate = allGates
				if allGates {
					p.Gates = append(p.Gates, parStep)
				} else {
					p.Steps = append(p.Steps, parStep)
				}
			default:
				return Phase{}, fmt.Errorf("line %d: unexpected form %q inside phase (expected step, gate, or par)", child.Line, head)
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
		case "json-pick":
			jp, err := convertJsonPick(child, defs)
			if err != nil {
				return s, err
			}
			s.JsonPick = jp
		case "lines":
			ref, err := convertLines(child, defs)
			if err != nil {
				return s, err
			}
			s.Lines = ref
		case "merge":
			ids, err := convertMerge(child, defs)
			if err != nil {
				return s, err
			}
			s.Merge = ids
		case "http-get":
			hc, err := convertHttpCall(child, "GET", defs)
			if err != nil {
				return s, err
			}
			s.HttpCall = hc
		case "http-post":
			hc, err := convertHttpCall(child, "POST", defs)
			if err != nil {
				return s, err
			}
			s.HttpCall = hc
		case "read-file":
			path, err := convertReadFile(child, defs)
			if err != nil {
				return s, err
			}
			s.ReadFile = path
		case "write-file":
			wf, err := convertWriteFile(child, defs)
			if err != nil {
				return s, err
			}
			s.WriteFile = wf
		case "glob":
			g, err := convertGlobStep(child, defs)
			if err != nil {
				return s, err
			}
			s.GlobPat = g
		case "plugin":
			pc, err := convertPluginCall(child, defs)
			if err != nil {
				return s, err
			}
			s.PluginCall = pc
		default:
			// github/prs → plugin "github", subcommand "prs"
			if parts := strings.SplitN(head, "/", 2); len(parts) == 2 {
				pc := &PluginCallStep{
					Plugin:     parts[0],
					Subcommand: parts[1],
					Args:       make(map[string]string),
				}
				for i := 1; i < len(child.Children); i++ {
					c := child.Children[i]
					if c.IsAtom() && c.Atom.Type == sexpr.TokenKeyword {
						key := c.KeywordVal()
						if i+1 >= len(child.Children) || (child.Children[i+1].IsAtom() && child.Children[i+1].Atom.Type == sexpr.TokenKeyword) {
							pc.Args[key] = "true"
						} else {
							i++
							pc.Args[key] = resolveVal(child.Children[i], defs)
						}
					}
				}
				s.PluginCall = pc
			} else {
				return s, fmt.Errorf("line %d: unknown step type %q", child.Line, head)
			}
		}
	}
	return s, nil
}

func convertJsonPick(n *sexpr.Node, defs map[string]string) (*JsonPickStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (json-pick) missing expression", n.Line)
	}
	jp := &JsonPickStep{Expr: resolveVal(children[0], defs)}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword && child.KeywordVal() == "from" {
			i++
			if i < len(children) {
				jp.From = resolveVal(children[i], defs)
			}
		}
	}
	return jp, nil
}

func convertLines(n *sexpr.Node, defs map[string]string) (string, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return "", fmt.Errorf("line %d: (lines) missing step ID", n.Line)
	}
	return resolveVal(children[0], defs), nil
}

func convertMerge(n *sexpr.Node, defs map[string]string) ([]string, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (merge) missing step IDs", n.Line)
	}
	ids := make([]string, len(children))
	for i, c := range children {
		ids[i] = resolveVal(c, defs)
	}
	return ids, nil
}

func convertReadFile(n *sexpr.Node, defs map[string]string) (string, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return "", fmt.Errorf("line %d: (read-file) missing path", n.Line)
	}
	return resolveVal(children[0], defs), nil
}

func convertWriteFile(n *sexpr.Node, defs map[string]string) (*WriteFileStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (write-file) missing path", n.Line)
	}
	wf := &WriteFileStep{Path: resolveVal(children[0], defs)}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword && child.KeywordVal() == "from" {
			i++
			if i < len(children) {
				wf.From = resolveVal(children[i], defs)
			}
		}
	}
	return wf, nil
}

func convertGlobStep(n *sexpr.Node, defs map[string]string) (*GlobStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (glob) missing pattern", n.Line)
	}
	g := &GlobStep{Pattern: resolveVal(children[0], defs)}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword && child.KeywordVal() == "dir" {
			i++
			if i < len(children) {
				g.Dir = resolveVal(children[i], defs)
			}
		}
	}
	return g, nil
}

func convertHttpCall(n *sexpr.Node, method string, defs map[string]string) (*HttpCallStep, error) {
	children := n.Children[1:]
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (%s) missing URL", n.Line, strings.ToLower("http-"+method))
	}
	hc := &HttpCallStep{
		Method:  method,
		URL:     resolveVal(children[0], defs),
		Headers: make(map[string]string),
	}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				break
			}
			val := children[i]
			switch key {
			case "body":
				hc.Body = resolveVal(val, defs)
			case "headers":
				if val.IsList() {
					src := val.Children
					for j := 0; j+1 < len(src); j += 2 {
						hc.Headers[src[j].StringVal()] = resolveVal(src[j+1], defs)
					}
				}
			}
		}
	}
	return hc, nil
}

func convertPluginCall(n *sexpr.Node, defs map[string]string) (*PluginCallStep, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return nil, fmt.Errorf("line %d: (plugin) needs name and subcommand", n.Line)
	}
	pc := &PluginCallStep{
		Plugin:     resolveVal(children[0], defs),
		Subcommand: resolveVal(children[1], defs),
		Args:       make(map[string]string),
	}
	for i := 2; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			// Check if next token is another keyword or end-of-list (flag mode)
			if i+1 >= len(children) || (children[i+1].IsAtom() && children[i+1].Atom.Type == sexpr.TokenKeyword) {
				pc.Args[key] = "true"
			} else {
				i++
				pc.Args[key] = resolveVal(children[i], defs)
			}
		}
	}
	return pc, nil
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
