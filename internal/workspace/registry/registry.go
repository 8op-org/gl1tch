package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

type Entry struct {
	Name string
	Path string
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "glitch"), nil
}

func registryPath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "workspaces.glitch"), nil
}

func List() ([]Entry, error) {
	p, err := registryPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "workspaces" {
			continue
		}
		for _, child := range n.Children[1:] {
			if !child.IsList() || len(child.Children) < 2 || child.Children[0].SymbolVal() != "workspace" {
				continue
			}
			e := Entry{Name: child.Children[1].StringVal()}
			kids := child.Children[2:]
			for i := 0; i+1 < len(kids); i += 2 {
				if kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword && kids[i].KeywordVal() == "path" {
					e.Path = kids[i+1].StringVal()
				}
			}
			out = append(out, e)
		}
	}
	return out, nil
}

func Add(e Entry) error {
	if e.Name == "" || e.Path == "" {
		return fmt.Errorf("registry: name and path required")
	}
	cur, err := List()
	if err != nil {
		return err
	}
	for _, x := range cur {
		if x.Name == e.Name {
			return fmt.Errorf("registry: name %q already registered at %s", x.Name, x.Path)
		}
	}
	cur = append(cur, e)
	return Save(cur)
}

func Remove(name string) error {
	cur, err := List()
	if err != nil {
		return err
	}
	out := cur[:0]
	for _, x := range cur {
		if x.Name != name {
			out = append(out, x)
		}
	}
	return Save(out)
}

func Save(entries []Entry) error {
	p, err := registryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("(workspaces")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("\n  (workspace %q :path %q)", e.Name, e.Path))
	}
	b.WriteString(")\n")
	return os.WriteFile(p, []byte(b.String()), 0o644)
}

func Find(name string) (Entry, bool, error) {
	entries, err := List()
	if err != nil {
		return Entry{}, false, err
	}
	for _, e := range entries {
		if e.Name == name {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}
