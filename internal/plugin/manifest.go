// internal/plugin/manifest.go
package plugin

import (
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Manifest holds metadata and shared definitions parsed from a plugin.glitch file.
type Manifest struct {
	Name        string
	Description string
	Version     string
	Defs        map[string]string
}

// LoadManifest reads and parses the plugin.glitch file in dir.
// If no file exists, it returns a default manifest with Name set to the
// directory's base name and an empty Defs map.
func LoadManifest(dir string) (*Manifest, error) {
	m := &Manifest{
		Name: filepath.Base(dir),
		Defs: make(map[string]string),
	}

	data, err := os.ReadFile(filepath.Join(dir, "plugin.glitch"))
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}

	nodes, err := sexpr.Parse(data)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if !node.IsList() || len(node.Children) == 0 {
			continue
		}
		head := node.Children[0].SymbolVal()
		switch head {
		case "plugin":
			parsePluginForm(node.Children[1:], m)
		case "def":
			parseDefForm(node.Children[1:], m)
		}
	}

	return m, nil
}

// parsePluginForm handles (plugin "name" :key "val" ...).
// Children passed in are everything after the "plugin" symbol.
func parsePluginForm(args []*sexpr.Node, m *Manifest) {
	if len(args) == 0 {
		return
	}
	idx := 0
	// First positional arg may be the plugin name string.
	if args[0].IsAtom() && args[0].Atom.Type == sexpr.TokenString {
		m.Name = args[0].StringVal()
		idx = 1
	}
	// Remaining args are keyword/value pairs.
	for i := idx; i+1 < len(args); i += 2 {
		k := args[i].KeywordVal()
		v := args[i+1].StringVal()
		switch k {
		case "description":
			m.Description = v
		case "version":
			m.Version = v
		}
	}
}

// parseDefForm handles (def name "value") where value may reference earlier defs.
// Children passed in are everything after the "def" symbol.
func parseDefForm(args []*sexpr.Node, m *Manifest) {
	if len(args) < 2 {
		return
	}
	name := args[0].SymbolVal()
	if name == "" {
		return
	}
	val := resolveValue(args[1], m.Defs)
	m.Defs[name] = val
}

// resolveValue returns the string value of a node, substituting any symbol
// that matches an earlier def with its resolved value.
func resolveValue(n *sexpr.Node, defs map[string]string) string {
	if n.IsAtom() {
		if n.Atom.Type == sexpr.TokenString {
			return n.StringVal()
		}
		if n.Atom.Type == sexpr.TokenSymbol {
			if v, ok := defs[n.SymbolVal()]; ok {
				return v
			}
			return n.SymbolVal()
		}
	}
	return ""
}
