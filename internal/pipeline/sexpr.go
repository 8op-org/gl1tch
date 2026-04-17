package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
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
	case "when":
		s, err := convertWhen(n, defs, false)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "when-not":
		s, err := convertWhen(n, defs, true)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "map", "each":
		s, err := convertMap(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "map-resources":
		s, err := convertMapResources(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "filter":
		s, err := convertFilter(n, defs)
		if err != nil {
			return nil, err
		}
		return []Step{s}, nil
	case "reduce":
		s, err := convertReduce(n, defs)
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
	case "->":
		return convertThread(n, defs)
	case "compare":
		s, err := convertCompare(n, defs)
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
	steps[0].Line = n.Line
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
	steps[0].Line = n.Line
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
	primary.Line = n.Line
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
	s.Line = n.Line
	return s, nil
}

// convertWhen: (when "pred" (step ...)) or (when-not "pred" (step ...))
func convertWhen(n *sexpr.Node, defs map[string]string, negate bool) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (when) needs predicate and body step", n.Line)
	}
	pred := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		// Body might be a compound form (map, par, etc.), not just a step
		head := children[1].Children[0].SymbolVal()
		if head == "" {
			head = children[1].Children[0].StringVal()
		}
		steps, formErr := convertForm(children[1], head, defs)
		if formErr != nil {
			return Step{}, fmt.Errorf("line %d: when body: %w", n.Line, err)
		}
		if len(steps) != 1 {
			return Step{}, fmt.Errorf("line %d: when body must be a single form", n.Line)
		}
		body = steps[0]
	}
	return Step{
		ID:       fmt.Sprintf("when-%d", n.Line),
		Form:     "when",
		WhenPred: pred,
		WhenBody: &body,
		WhenNot:  negate,
		Line:     n.Line,
	}, nil
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
		Line:    n.Line,
	}, nil
}

// convertFilter: (filter "step-id" (step "pred" ...))
func convertFilter(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (filter) needs source step ID and predicate step", n.Line)
	}
	source := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: filter body: %w", n.Line, err)
	}
	return Step{
		ID:         fmt.Sprintf("filter-%d", n.Line),
		Form:       "filter",
		FilterOver: source,
		FilterBody: &body,
		Line:       n.Line,
	}, nil
}

// convertReduce: (reduce "step-id" (step "fold" ...))
func convertReduce(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (reduce) needs source step ID and body step", n.Line)
	}
	source := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: reduce body: %w", n.Line, err)
	}
	return Step{
		ID:         fmt.Sprintf("reduce-%d", n.Line),
		Form:       "reduce",
		ReduceOver: source,
		ReduceBody: &body,
		Line:       n.Line,
	}, nil
}

// convertMapResources: (map-resources [:type "git"] (step "name" ...))
// Iterates over active workspace resources. Optional :type filter keeps only
// resources of the matching type. The trailing list form is the body step
// executed per resource with .resource.item bound to the current entry.
func convertMapResources(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	s := Step{
		ID:   fmt.Sprintf("map-resources-%d", n.Line),
		Form: "map-resources",
	}

	var bodyNode *sexpr.Node
	for i := 0; i < len(children); i++ {
		c := children[i]
		if c.IsAtom() && c.Atom.Type == sexpr.TokenKeyword {
			key := c.KeywordVal()
			if i+1 >= len(children) {
				return Step{}, fmt.Errorf("line %d: map-resources keyword :%s missing value", c.Line, key)
			}
			val := resolveVal(children[i+1], defs)
			switch key {
			case "type":
				s.MapResourcesType = val
			default:
				return Step{}, fmt.Errorf("line %d: map-resources: unknown keyword :%s", c.Line, key)
			}
			i++
			continue
		}
		if c.IsList() {
			// The body is the trailing list form (usually `step`).
			bodyNode = c
		}
	}
	if bodyNode == nil {
		return Step{}, fmt.Errorf("line %d: map-resources needs a body step", n.Line)
	}
	body, err := convertStep(bodyNode, defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: map-resources body: %w", n.Line, err)
	}
	s.MapResourcesBody = &body
	s.Line = n.Line
	return s, nil
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
		Line:     n.Line,
	}, nil
}

// convertThread: (-> form1 form2 form3 ...)
// Desugars into a sequence of steps where each form implicitly references the previous.
// SDK forms (search, index, etc.) are wrapped in auto-named steps.
// Collection forms (each/map, filter, reduce) have their source set to the previous step.
// (flatten) with no args becomes (flatten "prev-step-id").
func convertThread(n *sexpr.Node, defs map[string]string) ([]Step, error) {
	children := n.Children[1:] // skip "->"
	if len(children) < 2 {
		return nil, fmt.Errorf("line %d: (->) needs at least 2 forms", n.Line)
	}

	var steps []Step
	prevID := ""

	for i, child := range children {
		if !child.IsList() || len(child.Children) == 0 {
			return nil, fmt.Errorf("line %d: (->) children must be forms", child.Line)
		}
		head := child.Children[0].SymbolVal()
		if head == "" {
			head = child.Children[0].StringVal()
		}
		threadID := fmt.Sprintf("thread-%d-%d", n.Line, i)

		switch head {
		case "search", "index", "delete", "embed", "run", "llm",
			"json-pick", "pick", "lines", "merge", "http-get", "fetch",
			"http-post", "send", "read-file", "read", "write-file", "write", "glob", "plugin":
			// SDK/primitive form — wrap in a step
			wrappedNode := &sexpr.Node{
				Children: []*sexpr.Node{
					{Atom: &sexpr.Token{Type: sexpr.TokenSymbol, Val: "step"}},
					{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: threadID}},
					child,
				},
				Line: child.Line,
			}
			s, err := convertStep(wrappedNode, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = threadID

		case "flatten":
			// (flatten) with no args or (flatten "explicit") — auto-fill source if missing
			s := Step{ID: threadID, Flatten: prevID, Line: child.Line}
			if len(child.Children) >= 2 {
				s.Flatten = resolveVal(child.Children[1], defs)
			}
			steps = append(steps, s)
			prevID = threadID

		case "each", "map":
			// Rewrite source to prevID
			if prevID == "" {
				return nil, fmt.Errorf("line %d: (->) each/map has no preceding step", child.Line)
			}
			newChildren := make([]*sexpr.Node, 0, len(child.Children)+1)
			newChildren = append(newChildren, child.Children[0]) // "each"/"map"
			newChildren = append(newChildren, &sexpr.Node{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: prevID}})
			newChildren = append(newChildren, child.Children[1:]...) // body
			rewritten := &sexpr.Node{Children: newChildren, Line: child.Line}
			s, err := convertMap(rewritten, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		case "filter":
			if prevID == "" {
				return nil, fmt.Errorf("line %d: (->) filter has no preceding step", child.Line)
			}
			newChildren := make([]*sexpr.Node, 0, len(child.Children)+1)
			newChildren = append(newChildren, child.Children[0])
			newChildren = append(newChildren, &sexpr.Node{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: prevID}})
			newChildren = append(newChildren, child.Children[1:]...)
			rewritten := &sexpr.Node{Children: newChildren, Line: child.Line}
			s, err := convertFilter(rewritten, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		case "reduce":
			if prevID == "" {
				return nil, fmt.Errorf("line %d: (->) reduce has no preceding step", child.Line)
			}
			newChildren := make([]*sexpr.Node, 0, len(child.Children)+1)
			newChildren = append(newChildren, child.Children[0])
			newChildren = append(newChildren, &sexpr.Node{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: prevID}})
			newChildren = append(newChildren, child.Children[1:]...)
			rewritten := &sexpr.Node{Children: newChildren, Line: child.Line}
			s, err := convertReduce(rewritten, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		case "step":
			// Named step — just convert normally
			s, err := convertStep(child, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		default:
			return nil, fmt.Errorf("line %d: unknown form %q in (->)", child.Line, head)
		}
	}
	return steps, nil
}

// convertCompare: (compare [:id "name"] (branch "name" ...) ... [(review ...)])
func convertCompare(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	s := Step{
		ID:   fmt.Sprintf("compare-%d", n.Line),
		Form: "compare",
	}
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
			case "objective":
				s.CompareObjective = val
			}
			continue
		}
		break
	}
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
	if s.CompareObjective == "" {
		return s, fmt.Errorf("line %d: (compare) requires :objective — what is this comparison measuring?", n.Line)
	}
	s.Line = n.Line
	return s, nil
}

// convertBranch: (branch "name" (step ...) ...) or (branch "name" (llm ...))
func convertBranch(n *sexpr.Node, defs map[string]string) (CompareBranch, error) {
	children := n.Children[1:]
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
		switch head {
		case "llm", "run", "save":
			implicitStep := &sexpr.Node{
				Line: child.Line,
				Children: []*sexpr.Node{
					{Atom: &sexpr.Token{Type: sexpr.TokenSymbol, Val: "step"}, Line: child.Line},
					{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: name}, Line: child.Line},
					child,
				},
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
			formSteps, err := convertForm(child, head, defs)
			if err != nil {
				return CompareBranch{}, fmt.Errorf("line %d: branch %q: %w", child.Line, name, err)
			}
			steps = append(steps, formSteps...)
		}
	}
	return CompareBranch{Name: name, Steps: steps}, nil
}

// convertReview: (review [:criteria (...)] [:prompt "..."] [:model "..."])
func convertReview(n *sexpr.Node, defs map[string]string) (*ReviewConfig, error) {
	children := n.Children[1:]
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
		case "provider":
			r.Provider = resolveVal(valNode, defs)
		default:
			return nil, fmt.Errorf("line %d: unknown review keyword :%s (valid: :criteria, :prompt, :model, :provider)", child.Line, kw)
		}
	}
	return r, nil
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

	// Parse optional keywords before the body form (:hint "text")
	for len(children) > 0 {
		child := children[0]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			kw := child.KeywordVal()
			children = children[1:]
			if len(children) == 0 {
				return s, fmt.Errorf("line %d: keyword :%s missing value", child.Line, kw)
			}
			val := resolveVal(children[0], defs)
			children = children[1:]
			switch kw {
			case "hint":
				s.Hint = val
			}
			continue
		}
		break
	}

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
		case "json-pick", "pick":
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
		case "flatten":
			if len(child.Children) < 2 {
				return s, fmt.Errorf("line %d: (flatten) missing step ID", child.Line)
			}
			s.Flatten = resolveVal(child.Children[1], defs)
		case "merge":
			ids, err := convertMerge(child, defs)
			if err != nil {
				return s, err
			}
			s.Merge = ids
		case "http-get", "fetch":
			hc, err := convertHttpCall(child, "GET", defs)
			if err != nil {
				return s, err
			}
			s.HttpCall = hc
		case "http-post", "send":
			hc, err := convertHttpCall(child, "POST", defs)
			if err != nil {
				return s, err
			}
			s.HttpCall = hc
		case "read-file", "read":
			path, err := convertReadFile(child, defs)
			if err != nil {
				return s, err
			}
			s.ReadFile = path
		case "write-file", "write":
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
		case "search":
			sr, err := convertSearch(child, defs)
			if err != nil {
				return s, err
			}
			s.Search = sr
		case "index":
			idx, err := convertIndex(child, defs)
			if err != nil {
				return s, err
			}
			s.Index = idx
		case "delete":
			del, err := convertDelete(child, defs)
			if err != nil {
				return s, err
			}
			s.Delete = del
		case "embed":
			emb, err := convertEmbed(child, defs)
			if err != nil {
				return s, err
			}
			s.Embed = emb
		case "compare":
			cmp, err := convertCompare(child, defs)
			if err != nil {
				return s, err
			}
			s.Form = "compare"
			s.CompareBranches = cmp.CompareBranches
			s.CompareReview = cmp.CompareReview
			s.CompareObjective = cmp.CompareObjective
		case "plugin":
			pc, err := convertPluginCall(child, defs)
			if err != nil {
				return s, err
			}
			s.PluginCall = pc
		case "call-workflow":
			sub := child.Children[1:]
			if len(sub) < 1 {
				return s, fmt.Errorf("line %d: call-workflow needs workflow name", child.Line)
			}
			s.Form = "call-workflow"
			s.CallWorkflow = resolveVal(sub[0], defs)
			if s.CallSet == nil {
				s.CallSet = map[string]string{}
			}
			rest := sub[1:]
			for i := 0; i+1 < len(rest); i += 2 {
				if !(rest[i].IsAtom() && rest[i].Atom.Type == sexpr.TokenKeyword) {
					continue
				}
				key := rest[i].KeywordVal()
				val := resolveVal(rest[i+1], defs)
				switch key {
				case "input":
					s.CallInput = val
				case "set":
					if kv := splitKV(val); kv != nil {
						for k, v := range kv {
							s.CallSet[k] = v
						}
					}
				}
			}
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
	s.Line = n.Line
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

// nodeToJSON converts a sexpr node to JSON bytes.
// Map nodes (IsMap=true) become JSON objects, list nodes become arrays, atoms become strings.
func nodeToJSON(n *sexpr.Node) ([]byte, error) {
	if n.IsAtom() {
		return json.Marshal(n.StringVal())
	}
	if n.IsMap {
		// Map: children are alternating key, value
		result := make(map[string]json.RawMessage)
		for i := 0; i+1 < len(n.Children); i += 2 {
			key := n.Children[i].StringVal()
			val, err := nodeToJSON(n.Children[i+1])
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return json.Marshal(result)
	}
	// List: children become array
	var items []json.RawMessage
	for _, child := range n.Children {
		val, err := nodeToJSON(child)
		if err != nil {
			return nil, err
		}
		items = append(items, val)
	}
	return json.Marshal(items)
}

func convertSearch(n *sexpr.Node, defs map[string]string) (*SearchStep, error) {
	children := n.Children[1:] // skip "search"
	sr := &SearchStep{Size: 10}
	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			switch key {
			case "ndjson":
				sr.NDJSON = true
				i++
				continue
			}
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "index":
				sr.IndexName = resolveVal(val, defs)
			case "query":
				b, err := nodeToJSON(val)
				if err != nil {
					return nil, fmt.Errorf("line %d: query to JSON: %w", val.Line, err)
				}
				sr.Query = string(b)
			case "size":
				n, err := strconv.Atoi(resolveVal(val, defs))
				if err != nil {
					return nil, fmt.Errorf("line %d: :size must be an integer", val.Line)
				}
				sr.Size = n
			case "fields":
				if val.IsList() {
					for _, f := range val.Children {
						sr.Fields = append(sr.Fields, resolveVal(f, defs))
					}
				} else {
					sr.Fields = append(sr.Fields, resolveVal(val, defs))
				}
			case "es":
				sr.ESURL = resolveVal(val, defs)
			case "sort":
				b, err := nodeToJSON(val)
				if err != nil {
					return nil, fmt.Errorf("line %d: sort to JSON: %w", val.Line, err)
				}
				sr.Sort = string(b)
			default:
				return nil, fmt.Errorf("line %d: unknown search keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in search", child.Line)
	}
	if sr.IndexName == "" {
		return nil, fmt.Errorf("line %d: search missing :index", n.Line)
	}
	return sr, nil
}

func convertIndex(n *sexpr.Node, defs map[string]string) (*IndexStep, error) {
	children := n.Children[1:] // skip "index"
	idx := &IndexStep{}
	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			switch key {
			case "embed":
				// :embed is followed by sub-keywords :field, :model
				i++
				for i < len(children) {
					sub := children[i]
					if !sub.IsAtom() || sub.Atom.Type != sexpr.TokenKeyword {
						break
					}
					subKey := sub.KeywordVal()
					if subKey != "field" && subKey != "model" {
						break
					}
					i++
					if i >= len(children) {
						return nil, fmt.Errorf("line %d: embed :%s missing value", sub.Line, subKey)
					}
					val := children[i]
					switch subKey {
					case "field":
						idx.EmbedField = resolveVal(val, defs)
					case "model":
						idx.EmbedModel = resolveVal(val, defs)
					}
					i++
				}
				continue
			default:
				i++
				if i >= len(children) {
					return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
				}
				val := children[i]
				switch key {
				case "index":
					idx.IndexName = resolveVal(val, defs)
				case "doc":
					idx.Doc = resolveVal(val, defs)
				case "id":
					idx.DocID = resolveVal(val, defs)
				case "es":
					idx.ESURL = resolveVal(val, defs)
				case "upsert":
					v := strings.ToLower(resolveVal(val, defs))
					b := v != "false" && v != "0" && v != "no"
					idx.Upsert = &b
				default:
					return nil, fmt.Errorf("line %d: unknown index keyword :%s", child.Line, key)
				}
				i++
				continue
			}
		}
		return nil, fmt.Errorf("line %d: unexpected form in index", child.Line)
	}
	if idx.IndexName == "" {
		return nil, fmt.Errorf("line %d: index missing :index", n.Line)
	}
	if idx.Doc == "" {
		return nil, fmt.Errorf("line %d: index missing :doc", n.Line)
	}
	return idx, nil
}

func convertDelete(n *sexpr.Node, defs map[string]string) (*DeleteStep, error) {
	children := n.Children[1:] // skip "delete"
	del := &DeleteStep{}
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
			case "index":
				del.IndexName = resolveVal(val, defs)
			case "query":
				b, err := nodeToJSON(val)
				if err != nil {
					return nil, fmt.Errorf("line %d: query to JSON: %w", val.Line, err)
				}
				del.Query = string(b)
			case "es":
				del.ESURL = resolveVal(val, defs)
			default:
				return nil, fmt.Errorf("line %d: unknown delete keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in delete", child.Line)
	}
	if del.IndexName == "" {
		return nil, fmt.Errorf("line %d: delete missing :index", n.Line)
	}
	if del.Query == "" {
		return nil, fmt.Errorf("line %d: delete missing :query", n.Line)
	}
	return del, nil
}

func convertEmbed(n *sexpr.Node, defs map[string]string) (*EmbedStep, error) {
	children := n.Children[1:] // skip "embed"
	emb := &EmbedStep{}
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
			case "input":
				emb.Input = resolveVal(val, defs)
			case "model":
				emb.Model = resolveVal(val, defs)
			default:
				return nil, fmt.Errorf("line %d: unknown embed keyword :%s", child.Line, key)
			}
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in embed", child.Line)
	}
	if emb.Input == "" {
		return nil, fmt.Errorf("line %d: embed missing :input", n.Line)
	}
	if emb.Model == "" {
		return nil, fmt.Errorf("line %d: embed missing :model", n.Line)
	}
	return emb, nil
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

// splitKV parses a single "key=value" string into a map. Returns nil if the
// string lacks an "=" separator or the key is empty.
func splitKV(s string) map[string]string {
	if s == "" {
		return nil
	}
	if i := strings.Index(s, "="); i > 0 {
		return map[string]string{s[:i]: s[i+1:]}
	}
	return nil
}

// ParseSexprWorkflowFromFile loads and parses a .glitch workflow file.
func ParseSexprWorkflowFromFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseSexprWorkflow(data)
}
