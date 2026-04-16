// internal/pipeline/builtins.go
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// callBuiltin dispatches a named builtin function with evaluated string args.
// Returns the result string or a clear error.
func callBuiltin(name string, args []string, scope *Scope) (string, error) {
	switch name {
	case "step":
		if len(args) == 0 {
			return "", fmt.Errorf("step: missing id argument")
		}
		v, ok := scope.steps[args[0]]
		if !ok {
			return "", fmt.Errorf("step: unknown step id %q", args[0])
		}
		return v, nil

	case "stepfile":
		if len(args) == 0 {
			return "", fmt.Errorf("stepfile: missing id")
		}
		content, ok := scope.steps[args[0]]
		if !ok {
			return "", fmt.Errorf("stepfile: unknown step id %q", args[0])
		}
		f, err := os.CreateTemp("", "glitch-step-*")
		if err != nil {
			return "", err
		}
		f.WriteString(content)
		f.Close()
		return f.Name(), nil

	case "itemfile":
		if !scope.hasItem {
			return "", fmt.Errorf("itemfile: no ~item in scope")
		}
		f, err := os.CreateTemp("", "glitch-item-*")
		if err != nil {
			return "", err
		}
		f.WriteString(scope.item)
		f.Close()
		return f.Name(), nil

	case "branch":
		if len(args) == 0 {
			return "", fmt.Errorf("branch: missing name")
		}
		for k, v := range scope.steps {
			if strings.HasSuffix(k, "/"+args[0]+"/__output") {
				return v, nil
			}
		}
		if v, ok := scope.steps[args[0]]; ok {
			return v, nil
		}
		return "", fmt.Errorf("branch: unknown branch %q", args[0])

	case "or":
		for _, a := range args {
			if a != "" {
				return a, nil
			}
		}
		return "", nil
	}

	// String helpers
	if v, ok, err := tryStringBuiltin(name, args); ok {
		return v, err
	}
	// JSON helpers
	if v, ok, err := tryJSONBuiltin(name, args); ok {
		return v, err
	}

	return "", fmt.Errorf("unknown builtin %q", name)
}

func tryStringBuiltin(name string, args []string) (string, bool, error) {
	// Implemented in Task 5.
	return "", false, nil
}

func tryJSONBuiltin(name string, args []string) (string, bool, error) {
	_ = json.Valid // keep import for next task
	return "", false, nil
}
