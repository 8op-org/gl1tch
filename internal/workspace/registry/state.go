package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

func statePath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "state.glitch"), nil
}

func GetActive() (string, error) {
	p, err := statePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return "", err
	}
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "state" {
			continue
		}
		kids := n.Children[1:]
		for i := 0; i+1 < len(kids); i += 2 {
			if kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword && kids[i].KeywordVal() == "active-workspace" {
				return kids[i+1].StringVal(), nil
			}
		}
	}
	return "", nil
}

func SetActive(name string) error {
	p, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(fmt.Sprintf("(state :active-workspace %q)\n", name)), 0o644)
}
