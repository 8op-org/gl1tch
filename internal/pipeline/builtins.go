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
	switch name {
	case "upper":
		if len(args) < 1 {
			return "", true, fmt.Errorf("upper: missing arg")
		}
		return strings.ToUpper(args[0]), true, nil
	case "lower":
		if len(args) < 1 {
			return "", true, fmt.Errorf("lower: missing arg")
		}
		return strings.ToLower(args[0]), true, nil
	case "trim":
		if len(args) < 1 {
			return "", true, fmt.Errorf("trim: missing arg")
		}
		return strings.TrimSpace(args[0]), true, nil
	case "trimPrefix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("trimPrefix: need (prefix s)")
		}
		return strings.TrimPrefix(args[1], args[0]), true, nil
	case "trimSuffix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("trimSuffix: need (suffix s)")
		}
		return strings.TrimSuffix(args[1], args[0]), true, nil
	case "replace":
		if len(args) < 3 {
			return "", true, fmt.Errorf("replace: need (old new s)")
		}
		return strings.ReplaceAll(args[2], args[0], args[1]), true, nil
	case "truncate":
		if len(args) < 2 {
			return "", true, fmt.Errorf("truncate: need (n s)")
		}
		n := 0
		fmt.Sscanf(args[0], "%d", &n)
		runes := []rune(args[1])
		if len(runes) <= n {
			return args[1], true, nil
		}
		return string(runes[:n]), true, nil
	case "contains":
		if len(args) < 2 {
			return "", true, fmt.Errorf("contains: need (haystack needle)")
		}
		return boolStr(strings.Contains(args[0], args[1])), true, nil
	case "hasPrefix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("hasPrefix: need (s prefix)")
		}
		return boolStr(strings.HasPrefix(args[0], args[1])), true, nil
	case "hasSuffix":
		if len(args) < 2 {
			return "", true, fmt.Errorf("hasSuffix: need (s suffix)")
		}
		return boolStr(strings.HasSuffix(args[0], args[1])), true, nil
	case "split":
		if len(args) < 2 {
			return "", true, fmt.Errorf("split: need (sep s)")
		}
		return strings.Join(strings.Split(args[1], args[0]), "\n"), true, nil
	case "join":
		if len(args) < 2 {
			return "", true, fmt.Errorf("join: need (sep s)")
		}
		lines := strings.Split(args[1], "\n")
		return strings.Join(lines, args[0]), true, nil
	case "first":
		if len(args) < 1 {
			return "", true, fmt.Errorf("first: missing arg")
		}
		lines := strings.Split(args[0], "\n")
		if len(lines) == 0 {
			return "", true, nil
		}
		return lines[0], true, nil
	case "last":
		if len(args) < 1 {
			return "", true, fmt.Errorf("last: missing arg")
		}
		lines := strings.Split(args[0], "\n")
		if len(lines) == 0 {
			return "", true, nil
		}
		return lines[len(lines)-1], true, nil
	}
	return "", false, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func tryJSONBuiltin(name string, args []string) (string, bool, error) {
	switch name {
	case "pick":
		if len(args) < 2 {
			return "", true, fmt.Errorf("pick: need (key json)")
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[1]), &obj); err != nil {
			return "", true, fmt.Errorf("pick: invalid JSON: %w", err)
		}
		parts := strings.Split(args[0], ".")
		var cur any = obj
		for _, p := range parts {
			m, ok := cur.(map[string]any)
			if !ok {
				return "", true, nil
			}
			cur = m[p]
		}
		switch v := cur.(type) {
		case string:
			return v, true, nil
		case nil:
			return "", true, nil
		default:
			b, _ := json.Marshal(v)
			return string(b), true, nil
		}
	case "assoc":
		if len(args) < 3 {
			return "", true, fmt.Errorf("assoc: need (key val json)")
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[2]), &obj); err != nil {
			return "", true, err
		}
		obj[args[0]] = args[1]
		b, err := json.Marshal(obj)
		if err != nil {
			return "", true, err
		}
		return string(b), true, nil
	}
	return "", false, nil
}
