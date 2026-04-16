package workspace

import (
	"fmt"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// Workspace represents a parsed workspace.glitch file.
type Workspace struct {
	Name        string
	Description string
	Owner       string
	Repos       []string
	Defaults    Defaults
}

// Defaults holds default model/provider settings for a workspace.
type Defaults struct {
	Model         string
	Provider      string
	Elasticsearch string
	Params        map[string]string
}

// ParseFile parses workspace.glitch source bytes into a Workspace.
func ParseFile(src []byte) (*Workspace, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}

	for _, n := range nodes {
		if n.IsList() && len(n.Children) > 0 && n.Children[0].StringVal() == "workspace" {
			return convertWorkspace(n)
		}
	}
	return nil, fmt.Errorf("no (workspace ...) form found")
}

func convertWorkspace(n *sexpr.Node) (*Workspace, error) {
	children := n.Children[1:] // skip "workspace" symbol
	if len(children) == 0 {
		return nil, fmt.Errorf("line %d: workspace missing name", n.Line)
	}

	w := &Workspace{
		Repos: []string{},
	}

	// First child must be the name string
	w.Name = children[0].StringVal()
	if w.Name == "" {
		return nil, fmt.Errorf("line %d: workspace name must be a string", children[0].Line)
	}
	children = children[1:]

	i := 0
	for i < len(children) {
		child := children[i]

		// Keyword args: :description, :owner
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "description":
				w.Description = val.StringVal()
			case "owner":
				w.Owner = val.StringVal()
			default:
				return nil, fmt.Errorf("line %d: unknown workspace keyword :%s", child.Line, key)
			}
			i++
			continue
		}

		// List forms: (repos ...), (defaults ...)
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].SymbolVal()
			if head == "" {
				head = child.Children[0].StringVal()
			}
			switch head {
			case "repos":
				for _, repo := range child.Children[1:] {
					val := repo.StringVal()
					if val == "" {
						return nil, fmt.Errorf("line %d: repos entries must be strings", repo.Line)
					}
					w.Repos = append(w.Repos, val)
				}
			case "defaults":
				d, err := convertDefaults(child)
				if err != nil {
					return nil, err
				}
				w.Defaults = d
			default:
				return nil, fmt.Errorf("line %d: unknown workspace form %q", child.Line, head)
			}
			i++
			continue
		}

		return nil, fmt.Errorf("line %d: unexpected form in workspace", child.Line)
	}

	return w, nil
}

func convertDefaults(n *sexpr.Node) (Defaults, error) {
	children := n.Children[1:] // skip "defaults" symbol
	d := Defaults{Params: map[string]string{}}

	i := 0
	for i < len(children) {
		child := children[i]

		// Keyword args: :model, :provider, :elasticsearch
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return Defaults{}, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "model":
				d.Model = val.StringVal()
			case "provider":
				d.Provider = val.StringVal()
			case "elasticsearch":
				d.Elasticsearch = val.StringVal()
			default:
				return Defaults{}, fmt.Errorf("line %d: unknown defaults keyword :%s", child.Line, key)
			}
			i++
			continue
		}

		// List form: (params :key "val" ...)
		if child.IsList() && len(child.Children) > 0 {
			head := child.Children[0].SymbolVal()
			if head == "" {
				head = child.Children[0].StringVal()
			}
			switch head {
			case "params":
				p, err := convertParams(child)
				if err != nil {
					return Defaults{}, err
				}
				d.Params = p
			default:
				return Defaults{}, fmt.Errorf("line %d: unknown defaults form %q", child.Line, head)
			}
			i++
			continue
		}

		return Defaults{}, fmt.Errorf("line %d: unexpected form in defaults", child.Line)
	}

	return d, nil
}

func convertParams(n *sexpr.Node) (map[string]string, error) {
	children := n.Children[1:] // skip "params" symbol
	params := map[string]string{}

	i := 0
	for i < len(children) {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			params[key] = children[i].StringVal()
			i++
			continue
		}
		return nil, fmt.Errorf("line %d: unexpected form in params", child.Line)
	}

	return params, nil
}
