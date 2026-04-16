// internal/pipeline/quasi.go
package pipeline

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

type quasiPartKind int

const (
	partLiteral quasiPartKind = iota
	partRef
	partForm
)

type quasiPart struct {
	Kind    quasiPartKind
	Literal string   // for partLiteral
	RefBase string   // for partRef: bare symbol or dotted base
	RefPath []string // for partRef: dotted trailing components
	Form    string   // for partForm: raw "( ... )" source including outer parens
	Line    int
	Col     int
}

// lexQuasi scans a raw string for ~... interpolations.
// Returns an ordered list of literal/ref/form parts. No-tilde strings return
// a single literal part.
func lexQuasi(src string) ([]quasiPart, error) {
	var parts []quasiPart
	var lit strings.Builder
	flushLit := func() {
		if lit.Len() > 0 {
			parts = append(parts, quasiPart{Kind: partLiteral, Literal: lit.String()})
			lit.Reset()
		}
	}

	i := 0
	for i < len(src) {
		ch := src[i]
		if ch == '\\' && i+1 < len(src) && src[i+1] == '~' {
			lit.WriteByte('~')
			i += 2
			continue
		}
		if ch != '~' {
			lit.WriteByte(ch)
			i++
			continue
		}
		// ch == '~'
		if i+1 >= len(src) {
			lit.WriteByte('~')
			i++
			continue
		}
		next := src[i+1]
		if next == '(' {
			end, err := findFormEnd(src, i+1)
			if err != nil {
				return nil, err
			}
			flushLit()
			parts = append(parts, quasiPart{Kind: partForm, Form: src[i+1 : end+1]})
			i = end + 1
			continue
		}
		if isRefStartByte(next) {
			j := i + 1
			for j < len(src) && isRefByte(src[j]) {
				j++
			}
			ref := src[i+1 : j]
			base, path := splitDotted(ref)
			flushLit()
			parts = append(parts, quasiPart{Kind: partRef, RefBase: base, RefPath: path})
			i = j
			continue
		}
		// Lone ~ not followed by a ref start — treat as literal.
		lit.WriteByte('~')
		i++
	}
	flushLit()
	return parts, nil
}

func isRefStartByte(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isRefByte(ch byte) bool {
	return isRefStartByte(ch) || (ch >= '0' && ch <= '9') || ch == '.' || ch == '-'
}

func splitDotted(ref string) (base string, path []string) {
	if !strings.Contains(ref, ".") {
		return ref, nil
	}
	parts := strings.Split(ref, ".")
	return parts[0], parts[1:]
}

// findFormEnd finds the matching ")" for the "(" at src[start].
// Tracks string literals to avoid miscounting parens inside "..." .
func findFormEnd(src string, start int) (int, error) {
	if src[start] != '(' {
		return 0, fmt.Errorf("findFormEnd: expected '(' at %d", start)
	}
	depth := 0
	i := start
	inString := false
	for i < len(src) {
		ch := src[i]
		if inString {
			if ch == '\\' && i+1 < len(src) {
				i += 2
				continue
			}
			if ch == '"' {
				inString = false
			}
			i++
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, nil
			}
		}
		i++
	}
	return 0, fmt.Errorf("unterminated ~( form starting at byte %d", start)
}

// quoteFirstArg lists builtins whose first argument is a bare-symbol name
// taken literally rather than evaluated. For `(step diff)`, `diff` is the
// step ID, not a value to resolve.
var quoteFirstArg = map[string]bool{
	"step":     true,
	"stepfile": true,
	"branch":   true,
}

// evalForm parses and evaluates a single sexpr form against the scope.
// Returns the stringified result.
func evalForm(src string, scope *Scope) (string, error) {
	nodes, err := sexpr.Parse([]byte(src))
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", src, err)
	}
	if len(nodes) != 1 {
		return "", fmt.Errorf("expected one form, got %d", len(nodes))
	}
	return evalNode(nodes[0], scope)
}

func evalNode(n *sexpr.Node, scope *Scope) (string, error) {
	if n.IsAtom() {
		return evalAtom(n, scope)
	}
	if !n.IsList() || len(n.Children) == 0 {
		return "", nil
	}
	head := n.Children[0]
	name := head.SymbolVal()
	if name == "" {
		name = head.StringVal()
	}
	if name == "" {
		return "", fmt.Errorf("line %d: invalid form head", n.Line)
	}
	args := make([]string, 0, len(n.Children)-1)
	for i, c := range n.Children[1:] {
		if i == 0 && quoteFirstArg[name] {
			sym := c.SymbolVal()
			if sym == "" {
				sym = c.StringVal()
			}
			args = append(args, sym)
			continue
		}
		v, err := evalNode(c, scope)
		if err != nil {
			return "", err
		}
		args = append(args, v)
	}
	return callBuiltin(name, args, scope)
}

// isNumericLiteral reports whether sym looks like an integer or decimal literal
// (optional leading -, digits, optional single decimal point with digits).
// These are returned as-is rather than passed to scope.Resolve.
func isNumericLiteral(sym string) bool {
	if sym == "" {
		return false
	}
	i := 0
	if sym[0] == '-' || sym[0] == '+' {
		i = 1
		if i >= len(sym) {
			return false
		}
	}
	sawDigit := false
	sawDot := false
	for ; i < len(sym); i++ {
		ch := sym[i]
		if ch >= '0' && ch <= '9' {
			sawDigit = true
			continue
		}
		if ch == '.' && !sawDot {
			sawDot = true
			continue
		}
		return false
	}
	return sawDigit
}

func evalAtom(n *sexpr.Node, scope *Scope) (string, error) {
	switch n.Atom.Type {
	case sexpr.TokenString:
		return n.Atom.Val, nil
	case sexpr.TokenKeyword:
		return n.KeywordVal(), nil
	case sexpr.TokenSymbol:
		sym := n.Atom.Val
		if isNumericLiteral(sym) {
			return sym, nil
		}
		if strings.Contains(sym, ".") {
			parts := strings.Split(sym, ".")
			return scope.ResolvePath(parts[0], parts[1:])
		}
		return scope.Resolve(sym)
	}
	return "", fmt.Errorf("line %d: unsupported atom type", n.Line)
}

// renderQuasi interpolates ~name, ~param.x, and ~(form) references in src
// against the given scope. Strings without "~" pass through unchanged.
// Undefined refs return an error.
func renderQuasi(src string, scope *Scope) (string, error) {
	if !strings.ContainsRune(src, '~') {
		return src, nil
	}
	parts, err := lexQuasi(src)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, p := range parts {
		switch p.Kind {
		case partLiteral:
			out.WriteString(p.Literal)
		case partRef:
			var v string
			if len(p.RefPath) == 0 {
				v, err = scope.Resolve(p.RefBase)
			} else {
				v, err = scope.ResolvePath(p.RefBase, p.RefPath)
			}
			if err != nil {
				return "", err
			}
			out.WriteString(v)
		case partForm:
			v, err := evalForm(p.Form, scope)
			if err != nil {
				return "", err
			}
			out.WriteString(v)
		}
	}
	return out.String(), nil
}
