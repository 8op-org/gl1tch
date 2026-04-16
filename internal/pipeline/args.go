package pipeline

import (
	"fmt"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// InputDef describes the optional positional input for a workflow.
// At most one (input ...) form per file.
type InputDef struct {
	Description string
	Example     string
	Implicit    bool // true when auto-extracted from a ~input reference
}

// ParseInput extracts the single (input ...) form from a .glitch file.
// Returns (nil, nil) when no form is present. Returns an error when more
// than one (input ...) appears or when the form is given a name.
//
// Shape: (input :description "..." :example "...")
func ParseInput(src []byte) (*InputDef, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	var found *InputDef
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 {
			continue
		}
		if n.Children[0].SymbolVal() != "input" {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("line %d: only one (input ...) form allowed per file", n.Line)
		}

		// (input ...) takes no positional name. First child after the head
		// must be a keyword (:something), not a string.
		if len(n.Children) > 1 && !isKeyword(n.Children[1]) {
			return nil, fmt.Errorf("line %d: (input ...) takes no name, expected :keyword value pairs", n.Line)
		}

		def := &InputDef{}
		children := n.Children[1:]
		for i := 0; i < len(children); i++ {
			kw := children[i].KeywordVal()
			if kw == "" {
				continue
			}
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: (input ...) keyword :%s missing value", n.Line, kw)
			}
			switch kw {
			case "description":
				def.Description = children[i].StringVal()
			case "example":
				def.Example = children[i].StringVal()
			default:
				return nil, fmt.Errorf("line %d: (input ...) unknown keyword :%s", n.Line, kw)
			}
		}
		found = def
	}
	return found, nil
}

func isKeyword(n *sexpr.Node) bool {
	return n != nil && n.KeywordVal() != ""
}
