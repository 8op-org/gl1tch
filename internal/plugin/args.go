// internal/plugin/args.go
package plugin

import (
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// ArgDef describes a single argument declared in a .glitch file.
type ArgDef struct {
	Name        string
	Default     string // empty = required (unless type is flag)
	Type        string // "string", "flag", "number"
	Description string
	Example     string // concrete example value for help output
	Required    bool
	Implicit    bool // true when auto-extracted with no declaration
}

// ParseArgs parses a .glitch source file and extracts all (arg ...) forms.
//
// Each form has the shape:
//
//	(arg "name" :default "value" :type :flag :description "...")
//
// :type values come as keywords — the colon is stripped automatically via KeywordVal().
// Default type is "string". Required = true if no default AND type is not "flag".
func ParseArgs(src []byte) ([]ArgDef, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse args: %w", err)
	}

	var defs []ArgDef
	for _, node := range nodes {
		if !node.IsList() || len(node.Children) == 0 {
			continue
		}
		head := node.Children[0]
		if head.SymbolVal() != "arg" {
			continue
		}

		def, err := parseArgNode(node)
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}
	return defs, nil
}

// parseArgNode converts a single (arg ...) list node into an ArgDef.
func parseArgNode(node *sexpr.Node) (ArgDef, error) {
	children := node.Children // [0] = "arg" symbol
	if len(children) < 2 {
		return ArgDef{}, fmt.Errorf("line %d: arg form requires a name", node.Line)
	}

	def := ArgDef{
		Name: children[1].StringVal(),
		Type: "string",
	}
	if def.Name == "" {
		return ArgDef{}, fmt.Errorf("line %d: arg name must be a string", node.Line)
	}

	// Walk keyword/value pairs starting at index 2.
	requiredExplicit := false
	i := 2
	for i < len(children) {
		kw := children[i].KeywordVal()
		if kw == "" {
			i++
			continue
		}
		i++

		switch kw {
		case "default":
			if i < len(children) {
				def.Default = children[i].StringVal()
				i++
			}
		case "type":
			if i < len(children) {
				// Value may be a keyword (:flag) or a plain string ("flag").
				val := children[i].KeywordVal()
				if val == "" {
					val = children[i].StringVal()
				}
				if val != "" {
					def.Type = val
				}
				i++
			}
		case "description":
			if i < len(children) {
				def.Description = children[i].StringVal()
				i++
			}
		case "example":
			if i < len(children) {
				def.Example = children[i].StringVal()
				i++
			}
		case "required":
			if i < len(children) {
				val := strings.ToLower(children[i].StringVal())
				if val == "" {
					// Keyword value form: :required true  (rare; :required :true)
					val = children[i].KeywordVal()
				}
				def.Required = val == "true" || val == "t" || val == "yes"
				requiredExplicit = true
				i++
			}
		default:
			return ArgDef{}, fmt.Errorf("line %d: unknown keyword :%s on (arg \"%s\")", node.Line, kw, def.Name)
		}
	}

	// Mutual exclusion: :required true and :default "x" both set.
	if requiredExplicit && def.Required && def.Default != "" {
		return ArgDef{}, fmt.Errorf("line %d: (arg \"%s\") cannot set both :required and :default", node.Line, def.Name)
	}
	// Default Required inference when not explicitly set.
	if !requiredExplicit {
		def.Required = def.Default == "" && def.Type != "flag"
	}

	return def, nil
}

// BuildParams merges provided flag values with ArgDef defaults and validates
// that all required args are present.
func BuildParams(defs []ArgDef, provided map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(defs))

	for _, def := range defs {
		val, ok := provided[def.Name]
		switch {
		case ok:
			out[def.Name] = val
		case def.Type == "flag":
			// Flags default to empty string when not provided.
			out[def.Name] = ""
		case def.Default != "":
			out[def.Name] = def.Default
		case def.Required:
			return nil, fmt.Errorf("requires argument %q", def.Name)
		default:
			out[def.Name] = ""
		}
	}

	return out, nil
}
