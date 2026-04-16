package pipeline

import (
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// ExtractImplicitRefs walks every render-capable string in a Workflow and
// returns (paramNames, usesInput). paramNames is the deduplicated, unordered
// set of param names referenced via ~param.X or within ~(form ...). usesInput
// is true iff any string references ~input or ~(... input ...). Quote-first-
// arg forms (step, stepfile, branch) skip their first positional argument per
// the sexpr-interpolation spec.
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
	tokens, err := lexQuasi(s)
	if err != nil {
		return nil, false
	}
	for _, tok := range tokens {
		switch tok.Kind {
		case partRef:
			// bare ~input
			if tok.RefBase == "input" && len(tok.RefPath) == 0 {
				usesInput = true
			} else if tok.RefBase == "param" && len(tok.RefPath) == 1 {
				// ~param.X
				params = append(params, tok.RefPath[0])
			}
		case partForm:
			// tok.Form is raw "(... )" text; parse and walk
			nodes, err := sexpr.Parse([]byte(tok.Form))
			if err != nil || len(nodes) == 0 {
				continue
			}
			collectFromForm(nodes[0], &params, &usesInput)
		}
	}
	return params, usesInput
}

// collectFromForm walks a form AST node, collecting param references in every
// atom position except the first argument of quote-first-arg forms.
func collectFromForm(n *sexpr.Node, params *[]string, usesInput *bool) {
	if n == nil {
		return
	}
	if !n.IsList() || len(n.Children) == 0 {
		// leaf atom
		sym := n.SymbolVal()
		if sym == "input" {
			*usesInput = true
			return
		}
		if strings.HasPrefix(sym, "param.") {
			name := strings.TrimPrefix(sym, "param.")
			if name != "" && !strings.ContainsRune(name, '.') {
				*params = append(*params, name)
			}
		}
		return
	}
	head := n.Children[0].SymbolVal()
	skipIdx := -1
	if quoteFirstArg[head] {
		skipIdx = 1
	}
	for i, child := range n.Children {
		if i == 0 {
			// skip the head symbol itself
			continue
		}
		if i == skipIdx {
			// skip the quoted-first-arg literal
			continue
		}
		collectFromForm(child, params, usesInput)
	}
}
