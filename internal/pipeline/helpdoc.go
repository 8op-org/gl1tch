package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/plugin"
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
		if step.JsonPick != nil {
			visit(step.JsonPick.Expr)
		}
		visit(step.ReadFile)
		if step.WriteFile != nil {
			visit(step.WriteFile.Path)
		}
		if step.GlobPat != nil {
			visit(step.GlobPat.Dir)
			visit(step.GlobPat.Pattern)
		}
		if step.Search != nil {
			visit(step.Search.IndexName)
			visit(step.Search.Query)
			visit(step.Search.Sort)
		}
		if step.Index != nil {
			visit(step.Index.IndexName)
			visit(step.Index.Doc)
			visit(step.Index.DocID)
		}
		if step.Delete != nil {
			visit(step.Delete.IndexName)
			visit(step.Delete.Query)
		}
		if step.Embed != nil {
			visit(step.Embed.Input)
		}
		if step.PluginCall != nil {
			for _, v := range step.PluginCall.Args {
				visit(v)
			}
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
