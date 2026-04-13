package research

import (
	"errors"
	"sort"
	"sync"
)

var ErrDuplicateResearcher = errors.New("research: duplicate researcher name")

type Registry struct {
	mu          sync.RWMutex
	researchers map[string]Researcher
}

func NewRegistry() *Registry {
	return &Registry{
		researchers: make(map[string]Researcher),
	}
}

// Register adds researcher to the registry. Returns an error if researcher is
// nil, has an empty name, or a researcher with that name is already registered.
func (r *Registry) Register(researcher Researcher) error {
	if researcher == nil {
		return errors.New("research: cannot register nil researcher")
	}
	name := researcher.Name()
	if name == "" {
		return errors.New("research: cannot register researcher with empty name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.researchers[name]; exists {
		return ErrDuplicateResearcher
	}
	r.researchers[name] = researcher
	return nil
}

// Lookup returns the researcher registered under name and whether it was found.
func (r *Registry) Lookup(name string) (Researcher, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res, ok := r.researchers[name]
	return res, ok
}

// List returns all registered researchers sorted by Name().
func (r *Registry) List() []Researcher {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Researcher, 0, len(r.researchers))
	for _, res := range r.researchers {
		out = append(out, res)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name() < out[j].Name()
	})
	return out
}

// Names returns the names of all registered researchers, sorted.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.researchers))
	for name := range r.researchers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
