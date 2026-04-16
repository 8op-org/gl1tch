// internal/pipeline/quasi.go
package pipeline

import (
	"fmt"
	"strings"
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
