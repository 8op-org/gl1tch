package pipeline

import (
	"fmt"
	"os"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// interpolate expands ~ref and ~(form) references in a string.
// Reuses the existing lexQuasi tokenizer from quasi.go.
func (ev *Evaluator) interpolate(env *Env, s string) (string, error) {
	if !strings.ContainsRune(s, '~') {
		return s, nil
	}

	parts, err := lexQuasi(s)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	for _, p := range parts {
		switch p.Kind {
		case partLiteral:
			out.WriteString(p.Literal)

		case partRef:
			v, err := ev.resolveRef(env, p.RefBase, p.RefPath)
			if err != nil {
				return "", err
			}
			out.WriteString(v)

		case partForm:
			nodes, err := sexpr.Parse([]byte(p.Form))
			if err != nil {
				return "", fmt.Errorf("interpolate form %q: %w", p.Form, err)
			}
			if len(nodes) == 0 {
				continue
			}
			val, err := ev.Eval(env, nodes[0])
			if err != nil {
				return "", fmt.Errorf("interpolate form %q: %w", p.Form, err)
			}
			out.WriteString(val.String())
		}
	}
	return out.String(), nil
}

// resolveRef resolves a ~ref with optional dotted path components.
// Resolution order:
//  1. Step lookup (no path)
//  2. Env lookup (no path)
//  3. Dotted paths: param.x, resource.name.field, input, workspace, env.VAR
//  4. Fallback: UndefinedRefError
func (ev *Evaluator) resolveRef(env *Env, base string, path []string) (string, error) {
	// No dotted path: try steps, then env
	if len(path) == 0 {
		ev.mu.Lock()
		if v, ok := ev.steps[base]; ok {
			ev.mu.Unlock()
			return v, nil
		}
		ev.mu.Unlock()

		if v, ok := env.Get(base); ok {
			return v.String(), nil
		}

		// Special bare symbols
		switch base {
		case "input":
			return ev.Input, nil
		case "workspace":
			return ev.Workspace, nil
		}

		return "", &UndefinedRefError{Symbol: base}
	}

	// Dotted paths
	switch base {
	case "param":
		if len(path) == 1 {
			if v, ok := ev.Params[path[0]]; ok {
				return v, nil
			}
			return "", &UndefinedRefError{Symbol: "param." + path[0]}
		}
		return "", &UndefinedRefError{Symbol: "param." + strings.Join(path, ".")}

	case "resource":
		if len(path) == 2 {
			if r, ok := ev.Resources[path[0]]; ok {
				if v, ok := r[path[1]]; ok {
					return v, nil
				}
			}
			return "", &UndefinedRefError{Symbol: "resource." + strings.Join(path, ".")}
		}
		return "", &UndefinedRefError{Symbol: "resource." + strings.Join(path, ".")}

	case "env":
		if len(path) == 1 {
			if v, ok := os.LookupEnv(path[0]); ok {
				return v, nil
			}
			return "", &UndefinedRefError{Symbol: "env." + path[0]}
		}
		return "", &UndefinedRefError{Symbol: "env." + strings.Join(path, ".")}

	case "input":
		return ev.Input, nil

	case "workspace":
		return ev.Workspace, nil
	}

	// Try step lookup with dotted name
	fullName := base + "." + strings.Join(path, ".")
	ev.mu.Lock()
	if v, ok := ev.steps[fullName]; ok {
		ev.mu.Unlock()
		return v, nil
	}
	ev.mu.Unlock()

	return "", &UndefinedRefError{Symbol: fullName}
}
