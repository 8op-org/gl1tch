package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/pipeline/stdlib"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Severity indicates whether a diagnostic is a warning or an error.
type Severity int

const (
	SeverityWarning Severity = iota
	SeverityError
)

func (s Severity) String() string {
	if s == SeverityError {
		return "error"
	}
	return "warning"
}

// DiagKind classifies the type of diagnostic.
type DiagKind int

const (
	DiagUnknownSymbol DiagKind = iota
	DiagBadArity
	DiagUndefinedRef
	DiagShadowsBuiltin
	DiagUnreachable
)

// Diagnostic is a single static-analysis finding.
type Diagnostic struct {
	Kind     DiagKind
	Severity Severity
	Line     int
	Message  string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("line %d: %s: %s", d.Line, d.Severity, d.Message)
}

// knownSpecialForms lists every special form handled by evalSpecial in eval.go.
var knownSpecialForms = map[string]bool{
	"def": true, "fn": true, "let": true,
	"do": true, "begin": true,
	"if": true, "when": true, "when-not": true, "cond": true,
	"workflow": true, "step": true, "par": true,
	"retry": true, "timeout": true, "catch": true,
	"include": true,
	"map": true, "each": true, "filter": true, "reduce": true,
	"compare": true, "phase": true, "gate": true,
	"->": true,
	// Sub-forms of compare — recognised so they don't trigger unknown-symbol.
	"branch": true, "review": true,
}

// knownBuiltins lists every builtin registered in registerBuiltins (eval_builtins.go).
var knownBuiltins = map[string]bool{
	"sh": true, "run": true,
	"ref": true,
	"str": true,
	"llm": true,
	"save": true,
	"read-file": true, "read": true,
	"write-file": true, "write": true,
	"list": true,
	"not": true, "=": true,
	"println": true,
	"or": true,
	"glob": true,
	"websearch": true,
	"http-get": true, "fetch": true,
	"http-post": true, "send": true,
	"call-workflow": true,
	"json-pick": true, "pick": true,
}

// checkScope tracks names visible at a given point during the walk.
type checkScope struct {
	defs   map[string]bool
	steps  map[string]bool
	parent *checkScope
}

func newCheckScope(parent *checkScope) *checkScope {
	return &checkScope{
		defs:   make(map[string]bool),
		steps:  make(map[string]bool),
		parent: parent,
	}
}

func (cs *checkScope) isDefined(name string) bool {
	if cs.defs[name] {
		return true
	}
	if cs.steps[name] {
		return true
	}
	if cs.parent != nil {
		return cs.parent.isDefined(name)
	}
	return false
}

// Check walks a parsed AST and returns diagnostics for unknown forms,
// bad arity, undefined refs, and builtin shadowing — without evaluating.
func Check(nodes []*sexpr.Node) []Diagnostic {
	c := &checker{scope: newCheckScope(nil)}
	for _, n := range nodes {
		c.walk(n)
	}
	return c.diags
}

type checker struct {
	scope *checkScope
	diags []Diagnostic
}

func (c *checker) warn(line int, kind DiagKind, msg string) {
	c.diags = append(c.diags, Diagnostic{Kind: kind, Severity: SeverityWarning, Line: line, Message: msg})
}

func (c *checker) err(line int, kind DiagKind, msg string) {
	c.diags = append(c.diags, Diagnostic{Kind: kind, Severity: SeverityError, Line: line, Message: msg})
}

func (c *checker) walk(node *sexpr.Node) {
	if node.IsAtom() {
		c.checkAtom(node)
		return
	}
	if !node.IsList() || len(node.Children) == 0 {
		return
	}

	head := node.Children[0]
	args := node.Children[1:]

	// Skip keyword args (e.g. :description "...").
	if head.IsAtom() && head.Atom.Type == sexpr.TokenKeyword {
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	if !head.IsAtom() {
		c.walk(head)
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	name := head.SymbolVal()
	if name == "" {
		// String-headed list — walk children.
		for _, ch := range node.Children {
			c.walk(ch)
		}
		return
	}

	if knownSpecialForms[name] {
		c.checkSpecialForm(name, args, node)
		return
	}

	if knownBuiltins[name] {
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	if c.scope.isDefined(name) {
		for _, a := range args {
			c.walk(a)
		}
		return
	}

	c.err(node.Line, DiagUnknownSymbol, fmt.Sprintf("unknown function or form '%s'", name))
}

func (c *checker) checkAtom(node *sexpr.Node) {
	if node.Atom.Type == sexpr.TokenString {
		c.checkInterpolation(node)
	}
}

func (c *checker) checkInterpolation(node *sexpr.Node) {
	s := node.Atom.Val
	if !strings.ContainsRune(s, '~') {
		return
	}
	parts, err := lexQuasi(s)
	if err != nil {
		c.err(node.Line, DiagUndefinedRef, fmt.Sprintf("bad interpolation: %s", err))
		return
	}
	for _, p := range parts {
		if p.Kind != partRef {
			continue
		}
		full := p.RefBase
		if len(p.RefPath) > 0 {
			full += "." + strings.Join(p.RefPath, ".")
		}
		// Well-known ref namespaces are always valid.
		switch p.RefBase {
		case "param", "env", "resource", "input", "workspace":
			continue
		}
		if !c.scope.isDefined(p.RefBase) {
			c.warn(node.Line, DiagUndefinedRef, fmt.Sprintf("reference '~%s' may be undefined — not a known def or step at this point", full))
		}
	}
}

func (c *checker) checkSpecialForm(name string, args []*sexpr.Node, node *sexpr.Node) {
	switch name {
	case "def":
		if len(args) < 2 {
			c.err(node.Line, DiagBadArity, "def requires a name and a value")
			return
		}
		defName := args[0].SymbolVal()
		if defName == "" {
			defName = args[0].StringVal()
		}
		if knownBuiltins[defName] || knownSpecialForms[defName] {
			c.warn(node.Line, DiagShadowsBuiltin, fmt.Sprintf("def '%s' shadows a builtin", defName))
		}
		c.scope.defs[defName] = true
		for _, a := range args[1:] {
			c.walk(a)
		}

	case "step", "gate":
		if len(args) < 1 {
			c.err(node.Line, DiagBadArity, fmt.Sprintf("%s requires at least a step ID", name))
			return
		}
		stepID := args[0].StringVal()
		if stepID != "" {
			c.scope.steps[stepID] = true
		}
		for _, a := range args[1:] {
			c.walk(a)
		}

	case "let":
		if len(args) < 2 {
			c.err(node.Line, DiagBadArity, "let requires a binding list and a body")
			return
		}
		child := newCheckScope(c.scope)
		bindings := args[0]
		if bindings.IsList() {
			for i := 0; i+1 < len(bindings.Children); i += 2 {
				bName := bindings.Children[i].SymbolVal()
				if bName != "" {
					child.defs[bName] = true
				}
			}
		}
		saved := c.scope
		c.scope = child
		for _, a := range args[1:] {
			c.walk(a)
		}
		c.scope = saved

	case "fn":
		if len(args) < 2 {
			c.err(node.Line, DiagBadArity, "fn requires a parameter list and a body")
			return
		}
		child := newCheckScope(c.scope)
		params := args[0]
		if params.IsList() {
			for _, p := range params.Children {
				pName := p.SymbolVal()
				if pName != "" {
					child.defs[pName] = true
				}
			}
		}
		saved := c.scope
		c.scope = child
		for _, a := range args[1:] {
			c.walk(a)
		}
		c.scope = saved

	case "workflow":
		for _, a := range args {
			c.walk(a)
		}

	case "map", "each", "filter":
		child := newCheckScope(c.scope)
		child.defs["item"] = true
		child.defs["item_index"] = true
		saved := c.scope
		c.scope = child
		for _, a := range args {
			c.walk(a)
		}
		c.scope = saved

	case "reduce":
		child := newCheckScope(c.scope)
		child.defs["item"] = true
		child.defs["acc"] = true
		saved := c.scope
		c.scope = child
		for _, a := range args {
			c.walk(a)
		}
		c.scope = saved

	case "->":
		child := newCheckScope(c.scope)
		child.defs["_"] = true
		saved := c.scope
		c.scope = child
		for _, a := range args {
			c.walk(a)
		}
		c.scope = saved

	case "catch":
		if len(args) >= 2 {
			c.walk(args[0])
			child := newCheckScope(c.scope)
			child.defs["err"] = true
			saved := c.scope
			c.scope = child
			for _, a := range args[1:] {
				c.walk(a)
			}
			c.scope = saved
		} else {
			for _, a := range args {
				c.walk(a)
			}
		}

	case "include":
		if len(args) < 1 {
			c.err(node.Line, DiagBadArity, "include requires a path argument")
			return
		}
		path := args[0].StringVal()
		if strings.HasPrefix(path, "std/") {
			name := strings.TrimPrefix(path, "std/") + ".glitch"
			data, err := stdlib.FS.ReadFile(name)
			if err != nil {
				c.err(node.Line, DiagUnknownSymbol, fmt.Sprintf("unknown stdlib module '%s'", path))
				return
			}
			nodes, err := sexpr.Parse(data)
			if err != nil {
				c.err(node.Line, DiagUnknownSymbol, fmt.Sprintf("stdlib parse error in '%s': %s", path, err))
				return
			}
			for _, n := range nodes {
				if n.IsList() && len(n.Children) >= 2 {
					head := n.Children[0].SymbolVal()
					if head == "def" {
						defName := n.Children[1].SymbolVal()
						if defName == "" {
							defName = n.Children[1].StringVal()
						}
						if defName != "" {
							c.scope.defs[defName] = true
						}
					}
				}
			}
		}
		// Non-std includes can't be statically resolved — skip

	default:
		for _, a := range args {
			c.walk(a)
		}
	}
}
