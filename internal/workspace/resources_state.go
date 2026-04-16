package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// ResourceState holds per-resource fetched timestamps for a workspace.
// Persisted at <workspace>/.glitch/resources.glitch (never committed).
type ResourceState struct {
	Entries map[string]time.Time
}

func resourceStatePath(ws string) string {
	return filepath.Join(ws, ".glitch", "resources.glitch")
}

func LoadResourceState(ws string) (ResourceState, error) {
	data, err := os.ReadFile(resourceStatePath(ws))
	if err != nil {
		if os.IsNotExist(err) {
			return ResourceState{Entries: map[string]time.Time{}}, nil
		}
		return ResourceState{}, err
	}
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return ResourceState{}, err
	}
	out := ResourceState{Entries: map[string]time.Time{}}
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "resources" {
			continue
		}
		for _, c := range n.Children[1:] {
			if !c.IsList() || len(c.Children) < 2 || c.Children[0].SymbolVal() != "resource-state" {
				continue
			}
			name := c.Children[1].StringVal()
			kids := c.Children[2:]
			for i := 0; i+1 < len(kids); i += 2 {
				if kids[i].IsAtom() && kids[i].Atom.Type == sexpr.TokenKeyword && kids[i].KeywordVal() == "fetched" {
					if t, err := time.Parse(time.RFC3339, kids[i+1].StringVal()); err == nil {
						out.Entries[name] = t
					}
				}
			}
		}
	}
	return out, nil
}

func SaveResourceState(ws string, st ResourceState) error {
	path := resourceStatePath(ws)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("(resources")
	names := make([]string, 0, len(st.Entries))
	for n := range st.Entries {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		b.WriteString(fmt.Sprintf("\n  (resource-state %q :fetched %q)", n, st.Entries[n].UTC().Format(time.RFC3339)))
	}
	b.WriteString(")\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
