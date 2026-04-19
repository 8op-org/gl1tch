package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

// parseSexprWorkflow parses s-expression source into a Workflow.
func parseSexprWorkflow(src []byte) (*Workflow, error) {
	return parseSexprWorkflowWithIncludes(src, nil)
}

// parseSexprWorkflowWithIncludes parses s-expression source, processing
// (include "path") top-level forms to import (def ...) bindings from other
// files. The visited map tracks already-seen file paths to detect circular
// includes.
func parseSexprWorkflowWithIncludes(src []byte, visited map[string]bool) (w *Workflow, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	// Collect defs — first from includes, then from local defs (local wins)
	defs := make(map[string]string)

	// Process includes: extract only (def ...) bindings from included files
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) < 2 {
			continue
		}
		if n.Children[0].SymbolVal() != "include" {
			continue
		}
		path := n.Children[1].StringVal()
		if path == "" {
			return nil, fmt.Errorf("line %d: (include) path must be a string", n.Line)
		}
		if visited[path] {
			return nil, fmt.Errorf("line %d: circular include detected: %s", n.Line, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("line %d: include %q: %w", n.Line, path, err)
		}
		visited[path] = true
		includedDefs, err := collectDefsFromFile(data, visited)
		if err != nil {
			return nil, fmt.Errorf("line %d: include %q: %w", n.Line, path, err)
		}
		for k, v := range includedDefs {
			defs[k] = v
		}
	}

	// Collect local defs (override included ones)
	for _, n := range nodes {
		if n.IsList() && len(n.Children) >= 3 && n.Children[0].SymbolVal() == "def" {
			name := n.Children[1].SymbolVal()
			if name == "" {
				name = n.Children[1].StringVal()
			}
			val := resolveVal(n.Children[2], defs)
			defs[name] = val
		}
	}

	// Find workflow and extract metadata directly (no converter needed;
	// the evaluator handles execution from Source).
	for _, n := range nodes {
		if n.IsList() && len(n.Children) > 0 && n.Children[0].StringVal() == "workflow" {
			w := &Workflow{}
			if len(n.Children) > 1 {
				w.Name = n.Children[1].StringVal()
			}
			// Parse keyword metadata
			for i := 2; i < len(n.Children); i++ {
				kw := n.Children[i].KeywordVal()
				if kw == "" {
					break // body starts
				}
				if i+1 >= len(n.Children) {
					break
				}
				switch kw {
				case "description":
					w.Description = resolveVal(n.Children[i+1], defs)
				case "author":
					w.Author = resolveVal(n.Children[i+1], defs)
				case "version":
					w.Version = resolveVal(n.Children[i+1], defs)
				case "created":
					w.Created = resolveVal(n.Children[i+1], defs)
				case "tags":
					if n.Children[i+1].IsList() {
						for _, t := range n.Children[i+1].Children {
							w.Tags = append(w.Tags, resolveVal(t, defs))
						}
					}
				case "action":
					w.Actions = append(w.Actions, resolveVal(n.Children[i+1], defs))
				}
				i++ // skip value
			}
			w.Source = src
			return w, nil
		}
	}
	return nil, fmt.Errorf("no (workflow ...) form found")
}

// collectDefsFromFile parses a file's source and returns only its (def ...)
// bindings, recursively processing any (include ...) forms it contains.
func collectDefsFromFile(src []byte, visited map[string]bool) (map[string]string, error) {
	nodes, err := sexpr.Parse(src)
	if err != nil {
		return nil, err
	}

	defs := make(map[string]string)

	// Process nested includes first
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) < 2 {
			continue
		}
		if n.Children[0].SymbolVal() != "include" {
			continue
		}
		path := n.Children[1].StringVal()
		if path == "" {
			return nil, fmt.Errorf("line %d: (include) path must be a string", n.Line)
		}
		if visited[path] {
			return nil, fmt.Errorf("line %d: circular include detected: %s", n.Line, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("line %d: include %q: %w", n.Line, path, err)
		}
		visited[path] = true
		nested, err := collectDefsFromFile(data, visited)
		if err != nil {
			return nil, fmt.Errorf("line %d: include %q: %w", n.Line, path, err)
		}
		for k, v := range nested {
			defs[k] = v
		}
	}

	// Collect local defs
	for _, n := range nodes {
		if n.IsList() && len(n.Children) >= 3 && n.Children[0].SymbolVal() == "def" {
			name := n.Children[1].SymbolVal()
			if name == "" {
				name = n.Children[1].StringVal()
			}
			val := resolveVal(n.Children[2], defs)
			defs[name] = val
		}
	}

	return defs, nil
}

// resolveVal returns the string value of a node, substituting def bindings for symbols.
// If the node is a list (form), it evaluates known pure forms at parse time.
func resolveVal(n *sexpr.Node, defs map[string]string) string {
	// Symbol lookup
	if n.Atom != nil && n.Atom.Type == sexpr.TokenSymbol {
		if v, ok := defs[n.Atom.Val]; ok {
			return v
		}
	}

	// Form evaluation in def context
	if n.IsList() && len(n.Children) >= 2 {
		head := n.Children[0].SymbolVal()
		switch head {
		case "read-file", "read":
			return resolveReadFile(n, defs)
		case "glob":
			return resolveGlob(n, defs)
		case "->":
			return resolveThread(n, defs)
		}
	}

	return n.StringVal()
}

// resolveReadFile evaluates (read-file "path" ...) at parse time.
// Multiple paths are concatenated with \n\n separator.
func resolveReadFile(n *sexpr.Node, defs map[string]string) string {
	children := n.Children[1:]
	var parts []string
	for _, child := range children {
		path := resolveVal(child, defs)
		data, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("(read-file): %v", err))
		}
		parts = append(parts, string(data))
	}
	return strings.Join(parts, "\n\n")
}

// resolveGlob evaluates (glob "pattern") at parse time.
// Returns newline-separated list of matching file paths.
func resolveGlob(n *sexpr.Node, defs map[string]string) string {
	pattern := resolveVal(n.Children[1], defs)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		panic(fmt.Sprintf("(glob): %v", err))
	}
	sort.Strings(matches)
	return strings.Join(matches, "\n")
}

// resolveThread evaluates (-> form1 form2 ...) at parse time.
// Each form receives the output of the previous as implicit input.
func resolveThread(n *sexpr.Node, defs map[string]string) string {
	children := n.Children[1:] // skip "->"
	if len(children) < 2 {
		panic("(->): needs at least 2 forms")
	}

	// First form produces initial value
	val := resolveVal(children[0], defs)

	// Remaining forms transform the value
	for _, child := range children[1:] {
		val = applyDefForm(child, val, defs)
	}
	return val
}

// defListSep is a sentinel separator for internal list representation in def
// threading. Using a separator that won't appear in file content prevents
// (join) from splitting on newlines inside file contents.
const defListSep = "\x00\n\x00"

// applyDefForm applies a form to a value in def/thread context.
func applyDefForm(n *sexpr.Node, input string, defs map[string]string) string {
	if !n.IsList() || len(n.Children) == 0 {
		panic(fmt.Sprintf("line %d: expected form in (->), got atom", n.Line))
	}
	head := n.Children[0].SymbolVal()

	switch head {
	case "map":
		if len(n.Children) < 2 {
			panic("(map) in thread: missing form name")
		}
		formName := n.Children[1].SymbolVal()
		if formName == "" {
			formName = n.Children[1].StringVal()
		}
		// Split on sentinel separator if present, otherwise newlines (from glob)
		var items []string
		if strings.Contains(input, defListSep) {
			items = strings.Split(input, defListSep)
		} else {
			items = strings.Split(input, "\n")
		}
		var results []string
		for _, item := range items {
			if item == "" {
				continue
			}
			switch formName {
			case "read-file", "read":
				data, err := os.ReadFile(item)
				if err != nil {
					panic(fmt.Sprintf("(map read-file): %v", err))
				}
				results = append(results, string(data))
			default:
				panic(fmt.Sprintf("(map %s): unsupported form in def context", formName))
			}
		}
		return strings.Join(results, defListSep)

	case "join":
		sep := "\n"
		if len(n.Children) >= 2 {
			sep = resolveVal(n.Children[1], defs)
		}
		// Split on sentinel if present, otherwise newlines
		var items []string
		if strings.Contains(input, defListSep) {
			items = strings.Split(input, defListSep)
		} else {
			items = strings.Split(input, "\n")
		}
		var nonEmpty []string
		for _, l := range items {
			if l != "" {
				nonEmpty = append(nonEmpty, l)
			}
		}
		return strings.Join(nonEmpty, sep)

	case "lines":
		return input

	case "split":
		if len(n.Children) < 2 {
			panic("(split): missing separator")
		}
		sep := resolveVal(n.Children[1], defs)
		parts := strings.Split(input, sep)
		return strings.Join(parts, "\n")

	case "trim":
		return strings.TrimSpace(input)

	case "upper":
		return strings.ToUpper(input)

	case "lower":
		return strings.ToLower(input)

	case "replace":
		if len(n.Children) < 3 {
			panic("(replace): needs old and new strings")
		}
		old := resolveVal(n.Children[1], defs)
		newStr := resolveVal(n.Children[2], defs)
		return strings.ReplaceAll(input, old, newStr)

	case "filter":
		if len(n.Children) < 2 {
			panic("(filter): missing predicate")
		}
		pred := n.Children[1]
		var items []string
		if strings.Contains(input, defListSep) {
			items = strings.Split(input, defListSep)
		} else {
			items = strings.Split(input, "\n")
		}
		var kept []string
		for _, item := range items {
			if evalDefPredicate(pred, item, defs) {
				kept = append(kept, item)
			}
		}
		return strings.Join(kept, "\n")

	case "flatten":
		return input

	case "read-file", "read":
		data, err := os.ReadFile(strings.TrimSpace(input))
		if err != nil {
			panic(fmt.Sprintf("(read-file) in thread: %v", err))
		}
		return string(data)

	default:
		panic(fmt.Sprintf("unsupported form %q in def thread", head))
	}
}

// evalDefPredicate evaluates a predicate form against a string value.
func evalDefPredicate(n *sexpr.Node, val string, defs map[string]string) bool {
	if !n.IsList() || len(n.Children) == 0 {
		return false
	}
	head := n.Children[0].SymbolVal()
	switch head {
	case "contains":
		if len(n.Children) < 2 {
			return false
		}
		substr := resolveVal(n.Children[1], defs)
		return strings.Contains(val, substr)
	default:
		panic(fmt.Sprintf("unsupported predicate %q in filter", head))
	}
}


// ParseSexprWorkflowFromFile loads and parses a .glitch workflow file.
func ParseSexprWorkflowFromFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseSexprWorkflow(data)
}
