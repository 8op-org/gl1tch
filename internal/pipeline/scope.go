// internal/pipeline/scope.go
package pipeline

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

type Scope struct {
	lets      map[string]string
	defs      map[string]string
	params    map[string]string
	item      string
	itemIdx   int
	hasItem   bool
	steps     map[string]string
	input     string
	hasInput  bool
	resources map[string]map[string]string
}

func NewScope() *Scope {
	return &Scope{
		lets:      map[string]string{},
		defs:      map[string]string{},
		params:    map[string]string{},
		steps:     map[string]string{},
		resources: map[string]map[string]string{},
	}
}

func (s *Scope) SetLet(name, val string)   { s.lets[name] = val }
func (s *Scope) SetDef(name, val string)   { s.defs[name] = val }
func (s *Scope) SetParam(name, val string) { s.params[name] = val }
func (s *Scope) SetInput(v string)         { s.input = v; s.hasInput = true }
func (s *Scope) SetItem(v string, idx int) { s.item = v; s.itemIdx = idx; s.hasItem = true }
func (s *Scope) SetSteps(st map[string]string) {
	s.steps = st
}
func (s *Scope) SetResources(r map[string]map[string]string) {
	s.resources = r
}

// Resolve looks up a bare symbol in precedence order: let, def, specials.
func (s *Scope) Resolve(name string) (string, error) {
	if v, ok := s.lets[name]; ok {
		return v, nil
	}
	if v, ok := s.defs[name]; ok {
		return v, nil
	}
	switch name {
	case "input":
		if s.hasInput {
			return s.input, nil
		}
	case "item":
		if s.hasItem {
			return s.item, nil
		}
	case "item_index":
		if s.hasItem {
			return fmt.Sprintf("%d", s.itemIdx), nil
		}
	}
	return "", &UndefinedRefError{Symbol: name, Suggestion: s.suggest(name)}
}

// ResolvePath resolves dotted paths like "param.repo" or "env.HOME".
func (s *Scope) ResolvePath(base string, path []string) (string, error) {
	switch base {
	case "param":
		if len(path) != 1 {
			return "", &UndefinedRefError{Symbol: "param." + joinDots(path), Suggestion: "param.x only supports one level"}
		}
		if v, ok := s.params[path[0]]; ok {
			return v, nil
		}
		return "", &UndefinedRefError{Symbol: "param." + path[0], Suggestion: s.suggestDottedNS("param", path[0])}
	case "env":
		if len(path) != 1 {
			return "", &UndefinedRefError{Symbol: "env." + joinDots(path)}
		}
		v, ok := os.LookupEnv(path[0])
		if !ok {
			return "", &UndefinedRefError{Symbol: "env." + path[0]}
		}
		return v, nil
	case "resource":
		// Two-level path: resource.<name>.<field>. Fails loud on missing
		// per spec: undefined refs surface as UndefinedRefError, consistent
		// with param/env/bare-symbol resolution. Use ~(or resource.x.y "")
		// if an empty fallback is genuinely desired.
		if len(path) != 2 {
			return "", &UndefinedRefError{Symbol: "resource." + joinDots(path), Suggestion: "resource.<name>.<field>"}
		}
		if r, ok := s.resources[path[0]]; ok {
			if v, ok := r[path[1]]; ok {
				return v, nil
			}
		}
		return "", &UndefinedRefError{Symbol: "resource." + joinDots(path), Suggestion: s.suggestDottedNS("resource", path[0])}
	}
	return "", &UndefinedRefError{Symbol: base + "." + joinDots(path)}
}

func joinDots(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "."
		}
		out += p
	}
	return out
}

// suggestDottedNS returns the closest key in the namespace named by base
// (e.g. "param" or "resource") as a fully-qualified "<base>.<key>" string,
// or "" when no close match exists. Empty string means "no suggestion" —
// callers should wire that into UndefinedRefError.Suggestion directly so
// the error formatter omits the "did you mean" clause.
func (s *Scope) suggestDottedNS(base, key string) string {
	var keys []string
	switch base {
	case "param":
		for k := range s.params {
			keys = append(keys, k)
		}
	case "resource":
		for k := range s.resources {
			keys = append(keys, k)
		}
	}
	best := ""
	bestDist := len(key) + 1
	for _, c := range keys {
		d := levenshtein(key, c)
		if d < bestDist && d <= 2 {
			bestDist = d
			best = c
		}
	}
	if best == "" {
		return ""
	}
	return base + "." + best
}

// suggest returns the closest known symbol by Levenshtein distance, or "".
func (s *Scope) suggest(name string) string {
	candidates := []string{}
	for k := range s.lets {
		candidates = append(candidates, k)
	}
	for k := range s.defs {
		candidates = append(candidates, k)
	}
	candidates = append(candidates, "input", "item", "item_index")

	sort.Strings(candidates)
	best := ""
	bestDist := len(name) + 1
	for _, c := range candidates {
		d := levenshtein(name, c)
		if d < bestDist && d <= 2 {
			bestDist = d
			best = c
		}
	}
	return best
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d1 := prev[j] + 1
			d2 := curr[j-1] + 1
			d3 := prev[j-1] + cost
			m := d1
			if d2 < m {
				m = d2
			}
			if d3 < m {
				m = d3
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

// UndefinedRefError signals a lookup miss during ~-interpolation.
type UndefinedRefError struct {
	Symbol     string
	Suggestion string
	File       string
	Line       int
	Col        int
}

func (e *UndefinedRefError) Error() string {
	sug := ""
	if e.Suggestion != "" {
		sug = fmt.Sprintf(" (did you mean '%s'?)", e.Suggestion)
	}
	loc := ""
	if e.File != "" || e.Line != 0 {
		loc = fmt.Sprintf("%s:%d:%d: ", e.File, e.Line, e.Col)
	}
	return fmt.Sprintf("%sundefined reference '%s'%s", loc, e.Symbol, sug)
}

// errorsAsStd wraps errors.As for test visibility.
func errorsAsStd(err error, target any) bool {
	return errors.As(err, target)
}
