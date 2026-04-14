package pipeline

import (
	"fmt"

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

	// Process remaining children: keywords for metadata, lists for steps
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
		if child.IsList() && len(child.Children) > 0 && child.Children[0].StringVal() == "step" {
			step, err := convertStep(child, defs)
			if err != nil {
				return nil, err
			}
			w.Steps = append(w.Steps, step)
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in workflow", child.Line)
	}
	return w, nil
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
