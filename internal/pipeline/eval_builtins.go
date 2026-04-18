package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/sexpr"
)

// registerBuiltins installs all builtin functions into the given environment.
func (ev *Evaluator) registerBuiltins(env *Env) {
	b := func(name string, fn BuiltinFunc) {
		env.Set(name, &BuiltinVal{Name: name, Fn: fn})
	}

	// Shell execution
	b("sh", ev.builtinSh)
	b("run", ev.builtinSh)

	// Step reference
	b("ref", ev.builtinRef)

	// String concatenation
	b("str", ev.builtinStr)

	// LLM invocation
	b("llm", ev.builtinLLM)

	// File I/O
	b("save", ev.builtinSave)
	b("read-file", ev.builtinReadFile)
	b("read", ev.builtinReadFile)
	b("write-file", ev.builtinWriteFile)
	b("write", ev.builtinWriteFile)

	// Collections
	b("list", ev.builtinList)
	b("assoc", ev.builtinAssoc)
	b("count", ev.builtinCount)
	b("some", ev.builtinSome)
	b("every", ev.builtinEvery)
	b("set", ev.builtinSet)
	b("difference", ev.builtinDifference)
	b("flatten", ev.builtinFlatten)

	// Control
	b("assert", ev.builtinAssert)

	// Comparison
	b("<", ev.builtinLessThan)

	// Boolean
	b("not", ev.builtinNot)
	b("=", ev.builtinEq)

	// Debug output
	b("println", ev.builtinPrintln)

	// Short-circuit or
	b("or", ev.builtinOr)

	// File system
	b("glob", ev.builtinGlob)

	// Web
	b("websearch", ev.builtinWebsearch)
	b("http-get", ev.builtinHttpGet)
	b("fetch", ev.builtinHttpGet)
	b("http-post", ev.builtinHttpPost)
	b("send", ev.builtinHttpPost)

	// Workflow invocation
	b("call-workflow", ev.builtinCallWorkflow)

	// JSON / collection access
	b("json-pick", ev.builtinPick)
	b("pick", ev.builtinPick)

	// String operations
	b("replace", ev.builtinReplace)
	b("contains", ev.builtinContains)
	b("join", ev.builtinJoin)
	b("split", ev.builtinSplit)
	b("trim", ev.builtinTrim)
	b("upper", ev.builtinUpper)
	b("lower", ev.builtinLower)
	b("lines", ev.builtinLines)
	b("starts-with", ev.builtinStartsWith)
	b("ends-with", ev.builtinEndsWith)
	b("slice", ev.builtinSlice)
	b("regex-match", ev.builtinRegexMatch)
	b("regex-find", ev.builtinRegexFind)
}

// builtinSh: (sh "command") — execute a shell command
func (ev *Evaluator) builtinSh(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("sh: missing command argument")
	}
	cmdVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	cmd := cmdVal.String()

	output, err := provider.RunShell(cmd)
	if err != nil {
		return nil, fmt.Errorf("sh: %w", err)
	}
	return StringVal(strings.TrimRight(output, "\n")), nil
}

// builtinRef: (ref "step-id") — look up a step output
func (ev *Evaluator) builtinRef(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("ref: missing step id")
	}
	idVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	id := idVal.String()

	ev.mu.Lock()
	v, ok := ev.steps[id]
	ev.mu.Unlock()
	if !ok {
		return NilVal{}, nil
	}
	return StringVal(v), nil
}

// builtinStr: (str a b c...) — concatenate to string
func (ev *Evaluator) builtinStr(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	var sb strings.Builder
	for _, a := range args {
		v, err := ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
		sb.WriteString(v.String())
	}
	return StringVal(sb.String()), nil
}

// builtinLLM: (llm :prompt "..." :model "..." :provider "..." :skill "...")
func (ev *Evaluator) builtinLLM(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	var prompt, model, prov, skill string

	// Parse keyword arguments
	for i := 0; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				v, err := ev.Eval(env, args[i])
				if err != nil {
					return nil, err
				}
				switch kw {
				case "prompt":
					prompt = v.String()
				case "model":
					model = v.String()
				case "provider":
					prov = v.String()
				case "skill":
					skill = v.String()
				}
			}
			continue
		}
		// Positional: treat as prompt
		v, err := ev.Eval(env, n)
		if err != nil {
			return nil, err
		}
		if prompt == "" {
			prompt = v.String()
		}
	}

	if prompt == "" {
		return nil, fmt.Errorf("llm: missing prompt")
	}

	// Prepend skill context if provided
	if skill != "" {
		prompt = skill + "\n\n" + prompt
	}

	if model == "" {
		model = ev.DefaultModel
	}
	if prov == "" {
		prov = "ollama"
	}

	if ev.ProviderReg == nil {
		return StringVal("[llm stub: no provider registry]"), nil
	}

	result, err := ev.ProviderReg.RunProviderWithResult(prov, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	ev.mu.Lock()
	ev.TotalTokensIn += int64(result.TokensIn)
	ev.TotalTokensOut += int64(result.TokensOut)
	ev.TotalLatencyMS += result.Latency.Milliseconds()
	ev.TotalCostUSD += result.CostUSD
	ev.LLMSteps++
	ev.mu.Unlock()

	return StringVal(result.Response), nil
}

// builtinSave: (save :from "step-id" :to "path") or (save :content "..." :to "path")
func (ev *Evaluator) builtinSave(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	var content, to, from string

	for i := 0; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				v, err := ev.Eval(env, args[i])
				if err != nil {
					return nil, err
				}
				switch kw {
				case "from":
					from = v.String()
				case "content":
					content = v.String()
				case "to":
					to = v.String()
				}
			}
			continue
		}
	}

	if to == "" {
		return nil, fmt.Errorf("save: missing :to path")
	}

	if content == "" && from != "" {
		ev.mu.Lock()
		v, ok := ev.steps[from]
		ev.mu.Unlock()
		if !ok {
			return nil, fmt.Errorf("save: unknown step %q", from)
		}
		content = v
	}

	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return nil, fmt.Errorf("save: mkdir: %w", err)
	}
	if err := os.WriteFile(to, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return StringVal(to), nil
}

// builtinList: (list a b c...) — build a ListVal
func (ev *Evaluator) builtinList(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	vals := make(ListVal, 0, len(args))
	for _, a := range args {
		v, err := ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	return vals, nil
}

// builtinNot: (not val) — boolean negation
func (ev *Evaluator) builtinNot(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return BoolVal(true), nil
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	return BoolVal(!isTruthy(v)), nil
}

// builtinEq: (= a b) — string equality
func (ev *Evaluator) builtinEq(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("=: requires two arguments")
	}
	a, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	b, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	return BoolVal(a.String() == b.String()), nil
}

// builtinPrintln: (println args...) — debug output
func (ev *Evaluator) builtinPrintln(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		v, err := ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
		parts = append(parts, v.String())
	}
	fmt.Println(strings.Join(parts, " "))
	return NilVal{}, nil
}

// builtinOr: (or a b c...) — return first truthy, swallowing errors
func (ev *Evaluator) builtinOr(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	for _, a := range args {
		v, err := ev.Eval(env, a)
		if err != nil {
			continue // swallow errors
		}
		if isTruthy(v) {
			return v, nil
		}
	}
	return NilVal{}, nil
}

// builtinReadFile: (read-file "path" ...) — read one or more files
func (ev *Evaluator) builtinReadFile(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("read-file: missing path argument")
	}

	var parts []string
	for _, a := range args {
		v, err := ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(v.String())
		if err != nil {
			return nil, fmt.Errorf("read-file: %w", err)
		}
		parts = append(parts, string(data))
	}
	return StringVal(strings.Join(parts, "\n\n")), nil
}

// builtinWriteFile: (write-file :to "path" :content "..." or :from "step-id")
func (ev *Evaluator) builtinWriteFile(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	var content, to, from string

	for i := 0; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				v, err := ev.Eval(env, args[i])
				if err != nil {
					return nil, err
				}
				switch kw {
				case "to":
					to = v.String()
				case "content":
					content = v.String()
				case "from":
					from = v.String()
				}
			}
			continue
		}
	}

	if to == "" {
		return nil, fmt.Errorf("write-file: missing :to path")
	}

	if content == "" && from != "" {
		ev.mu.Lock()
		v, ok := ev.steps[from]
		ev.mu.Unlock()
		if !ok {
			return nil, fmt.Errorf("write-file: unknown step %q", from)
		}
		content = v
	}

	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return nil, fmt.Errorf("write-file: mkdir: %w", err)
	}
	if err := os.WriteFile(to, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("write-file: %w", err)
	}
	return StringVal(to), nil
}

// builtinGlob: (glob "pattern")
func (ev *Evaluator) builtinGlob(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("glob: missing pattern argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(v.String())
	if err != nil {
		return nil, fmt.Errorf("glob: %w", err)
	}
	vals := make(ListVal, len(matches))
	for i, m := range matches {
		vals[i] = StringVal(m)
	}
	return vals, nil
}

// builtinWebsearch: (websearch "query") — query SearXNG
func (ev *Evaluator) builtinWebsearch(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("websearch: missing query argument")
	}
	queryVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	query := queryVal.String()

	if ev.WebSearchURL == "" {
		return StringVal("[websearch stub: no WebSearchURL configured]"), nil
	}

	u, err := url.Parse(ev.WebSearchURL)
	if err != nil {
		return nil, fmt.Errorf("websearch: invalid URL: %w", err)
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("websearch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("websearch: read: %w", err)
	}

	// Parse JSON to extract results
	var data struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		// Return raw body if not JSON
		return StringVal(string(body)), nil
	}

	var sb strings.Builder
	for i, r := range data.Results {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s\n%s\n%s", r.Title, r.URL, r.Content))
	}
	return StringVal(sb.String()), nil
}

// builtinHttpGet: (http-get "url" :headers {...})
func (ev *Evaluator) builtinHttpGet(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("http-get: missing URL argument")
	}

	urlVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	targetURL := urlVal.String()

	// Parse optional keyword args
	headers := map[string]string{}
	for i := 1; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword && n.KeywordVal() == "headers" {
			if i+1 < len(args) {
				i++
				h := args[i]
				if h.IsList() && h.IsMap {
					for j := 0; j+1 < len(h.Children); j += 2 {
						k := h.Children[j].KeywordVal()
						if k == "" {
							k = h.Children[j].StringVal()
						}
						vv, err := ev.Eval(env, h.Children[j+1])
						if err != nil {
							return nil, err
						}
						headers[k] = vv.String()
					}
				}
			}
		}
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http-get: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http-get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http-get: read: %w", err)
	}
	return StringVal(string(body)), nil
}

// builtinHttpPost: (http-post "url" :body "..." :content-type "..." :headers {...})
func (ev *Evaluator) builtinHttpPost(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("http-post: missing URL argument")
	}

	urlVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	targetURL := urlVal.String()

	var reqBody, contentType string
	headers := map[string]string{}

	for i := 1; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				switch kw {
				case "body":
					v, err := ev.Eval(env, args[i])
					if err != nil {
						return nil, err
					}
					reqBody = v.String()
				case "content-type":
					v, err := ev.Eval(env, args[i])
					if err != nil {
						return nil, err
					}
					contentType = v.String()
				case "headers":
					h := args[i]
					if h.IsList() && h.IsMap {
						for j := 0; j+1 < len(h.Children); j += 2 {
							k := h.Children[j].KeywordVal()
							if k == "" {
								k = h.Children[j].StringVal()
							}
							vv, err := ev.Eval(env, h.Children[j+1])
							if err != nil {
								return nil, err
							}
							headers[k] = vv.String()
						}
					}
				}
			}
		}
	}

	if contentType == "" {
		contentType = "application/json"
	}

	req, err := http.NewRequest("POST", targetURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("http-post: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http-post: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http-post: read: %w", err)
	}
	return StringVal(string(body)), nil
}

// builtinCallWorkflow: (call-workflow "name" :input "..." :set {:key "val"})
func (ev *Evaluator) builtinCallWorkflow(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("call-workflow: missing workflow name")
	}

	nameVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	name := nameVal.String()

	var input string
	params := map[string]string{}

	for i := 1; i < len(args); i++ {
		n := args[i]
		if n.IsAtom() && n.Atom.Type == sexpr.TokenKeyword {
			kw := n.KeywordVal()
			if i+1 < len(args) {
				i++
				switch kw {
				case "input":
					v, err := ev.Eval(env, args[i])
					if err != nil {
						return nil, err
					}
					input = v.String()
				case "set":
					h := args[i]
					if h.IsList() && h.IsMap {
						for j := 0; j+1 < len(h.Children); j += 2 {
							k := h.Children[j].KeywordVal()
							if k == "" {
								k = h.Children[j].StringVal()
							}
							vv, err := ev.Eval(env, h.Children[j+1])
							if err != nil {
								return nil, err
							}
							params[k] = vv.String()
						}
					}
				}
			}
		}
	}

	// Find and read the workflow file
	wfPath := name
	if ev.WorkflowsDir != "" && !filepath.IsAbs(name) {
		// Try common extensions
		for _, ext := range []string{".glitch", ".yaml", ".yml"} {
			candidate := filepath.Join(ev.WorkflowsDir, name+ext)
			if _, err := os.Stat(candidate); err == nil {
				wfPath = candidate
				break
			}
		}
	}

	data, err := os.ReadFile(wfPath)
	if err != nil {
		return nil, fmt.Errorf("call-workflow: read %q: %w", wfPath, err)
	}

	// Cycle detection: check if this workflow is already on the call stack
	resolvedName := filepath.Base(wfPath)
	for _, ancestor := range ev.CallStack {
		if ancestor == resolvedName || ancestor == name {
			return nil, fmt.Errorf("call-workflow: cycle detected: %s already on call stack %v", name, ev.CallStack)
		}
	}

	// Create child evaluator
	child := NewEvaluator()
	child.Input = input
	child.DefaultModel = ev.DefaultModel
	child.Workspace = ev.Workspace
	child.Params = params
	child.Resources = ev.Resources
	child.ProviderReg = ev.ProviderReg
	child.ProviderResolver = ev.ProviderResolver
	child.Tiers = ev.Tiers
	child.EvalThreshold = ev.EvalThreshold
	child.ESURL = ev.ESURL
	child.WebSearchURL = ev.WebSearchURL
	child.WorkflowsDir = ev.WorkflowsDir
	child.StepRecorder = ev.StepRecorder
	child.CallStack = append(append([]string{}, ev.CallStack...), resolvedName)

	result, err := child.RunSource(data)
	if err != nil {
		return nil, fmt.Errorf("call-workflow %q: %w", name, err)
	}

	// Merge metrics
	ev.mu.Lock()
	ev.TotalTokensIn += child.TotalTokensIn
	ev.TotalTokensOut += child.TotalTokensOut
	ev.TotalLatencyMS += child.TotalLatencyMS
	ev.TotalCostUSD += child.TotalCostUSD
	ev.LLMSteps += child.LLMSteps
	ev.mu.Unlock()

	return result, nil
}

// builtinAssoc: (assoc :key val :key val ...) — build a MapVal from keyword-value pairs
func (ev *Evaluator) builtinAssoc(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	m := &MapVal{}
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return nil, fmt.Errorf("assoc: odd number of arguments — need key-value pairs")
		}
		keyNode := args[i]
		if !keyNode.IsAtom() || keyNode.Atom.Type != sexpr.TokenKeyword {
			return nil, fmt.Errorf("assoc: expected keyword, got %v", keyNode)
		}
		key := keyNode.KeywordVal()

		val, err := ev.Eval(env, args[i+1])
		if err != nil {
			return nil, err
		}
		m.Keys = append(m.Keys, key)
		m.Values = append(m.Values, val)
	}
	return m, nil
}

// builtinPick: (pick map :key) or (pick list index) or (pick json-string ".path")
func (ev *Evaluator) builtinPick(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return NilVal{}, nil
	}

	target, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}

	keyNode := args[1]

	// MapVal + keyword key
	if m, ok := target.(*MapVal); ok {
		if keyNode.IsAtom() && keyNode.Atom.Type == sexpr.TokenKeyword {
			v, found := m.Get(keyNode.KeywordVal())
			if !found {
				return NilVal{}, nil
			}
			return v, nil
		}
		keyVal, err := ev.Eval(env, keyNode)
		if err != nil {
			return nil, err
		}
		v, found := m.Get(keyVal.String())
		if !found {
			return NilVal{}, nil
		}
		return v, nil
	}

	// ListVal + numeric index
	if l, ok := target.(ListVal); ok {
		// Try raw atom value first (bare numbers are symbols, not defined vars)
		var idxStr string
		if keyNode.IsAtom() && keyNode.Atom.Type == sexpr.TokenSymbol {
			idxStr = keyNode.Atom.Val
		} else {
			keyVal, err := ev.Eval(env, keyNode)
			if err != nil {
				return nil, err
			}
			idxStr = keyVal.String()
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return NilVal{}, nil
		}
		if idx < 0 || idx >= len(l) {
			return NilVal{}, nil
		}
		return l[idx], nil
	}

	// Fallback: passthrough (existing behavior for json-pick placeholder)
	var result Value = target
	for _, a := range args[1:] {
		v, err := ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
		_ = v
	}
	return result, nil
}

// builtinReplace: (replace text old new [old new ...]) — multi-pair string replacement
func (ev *Evaluator) builtinReplace(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("replace: need at least text, old, new")
	}
	textVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	text := textVal.String()
	for i := 1; i+1 < len(args); i += 2 {
		oldVal, err := ev.Eval(env, args[i])
		if err != nil {
			return nil, err
		}
		newVal, err := ev.Eval(env, args[i+1])
		if err != nil {
			return nil, err
		}
		text = strings.ReplaceAll(text, oldVal.String(), newVal.String())
	}
	return StringVal(text), nil
}

// builtinContains: (contains text substr) or (contains list item)
func (ev *Evaluator) builtinContains(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("contains: need two arguments")
	}
	target, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	needle, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	if l, ok := target.(ListVal); ok {
		ns := needle.String()
		for _, item := range l {
			if item.String() == ns {
				return BoolVal(true), nil
			}
		}
		return BoolVal(false), nil
	}
	return BoolVal(strings.Contains(target.String(), needle.String())), nil
}

// builtinJoin: (join list sep) — join ListVal elements with separator
func (ev *Evaluator) builtinJoin(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("join: need list and separator")
	}
	listVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	sepVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	l, ok := listVal.(ListVal)
	if !ok {
		return nil, fmt.Errorf("join: first argument must be a list")
	}
	parts := make([]string, len(l))
	for i, v := range l {
		parts[i] = v.String()
	}
	return StringVal(strings.Join(parts, sepVal.String())), nil
}

// builtinSplit: (split text sep) — split string into ListVal
func (ev *Evaluator) builtinSplit(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("split: need text and separator")
	}
	textVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	sepVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	parts := strings.Split(textVal.String(), sepVal.String())
	vals := make(ListVal, len(parts))
	for i, p := range parts {
		vals[i] = StringVal(p)
	}
	return vals, nil
}

// builtinTrim: (trim text) — strings.TrimSpace
func (ev *Evaluator) builtinTrim(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("trim: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	return StringVal(strings.TrimSpace(v.String())), nil
}

// builtinUpper: (upper text) — strings.ToUpper
func (ev *Evaluator) builtinUpper(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("upper: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	return StringVal(strings.ToUpper(v.String())), nil
}

// builtinLower: (lower text) — strings.ToLower
func (ev *Evaluator) builtinLower(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("lower: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	return StringVal(strings.ToLower(v.String())), nil
}

// builtinLines: (lines text) — split on \n, return ListVal
func (ev *Evaluator) builtinLines(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("lines: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	parts := strings.Split(v.String(), "\n")
	vals := make(ListVal, len(parts))
	for i, p := range parts {
		vals[i] = StringVal(p)
	}
	return vals, nil
}

// builtinStartsWith: (starts-with text prefix) — return BoolVal
func (ev *Evaluator) builtinStartsWith(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("starts-with: need text and prefix")
	}
	textVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	prefixVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	return BoolVal(strings.HasPrefix(textVal.String(), prefixVal.String())), nil
}

// builtinEndsWith: (ends-with text suffix) — return BoolVal
func (ev *Evaluator) builtinEndsWith(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ends-with: need text and suffix")
	}
	textVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	suffixVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	return BoolVal(strings.HasSuffix(textVal.String(), suffixVal.String())), nil
}

// builtinSlice: (slice text start end) — substring extraction, bounds-safe
func (ev *Evaluator) builtinSlice(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("slice: need text, start, end")
	}
	textVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	// Handle bare numeric atoms (parsed as symbols) like pick does
	var startStr, endStr string
	if args[1].IsAtom() && args[1].Atom.Type == sexpr.TokenSymbol {
		startStr = args[1].Atom.Val
	} else {
		sv, err := ev.Eval(env, args[1])
		if err != nil {
			return nil, err
		}
		startStr = sv.String()
	}
	if args[2].IsAtom() && args[2].Atom.Type == sexpr.TokenSymbol {
		endStr = args[2].Atom.Val
	} else {
		sv, err := ev.Eval(env, args[2])
		if err != nil {
			return nil, err
		}
		endStr = sv.String()
	}
	text := textVal.String()
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return nil, fmt.Errorf("slice: invalid start: %w", err)
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		return nil, fmt.Errorf("slice: invalid end: %w", err)
	}
	// Bounds-safe clamping
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start > end {
		return StringVal(""), nil
	}
	return StringVal(text[start:end]), nil
}

// builtinRegexMatch: (regex-match pattern text) — all matches, returns ListVal
func (ev *Evaluator) builtinRegexMatch(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("regex-match: need pattern and text")
	}
	patVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	textVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(patVal.String())
	if err != nil {
		return nil, fmt.Errorf("regex-match: %w", err)
	}
	matches := re.FindAllStringSubmatch(textVal.String(), -1)
	vals := make(ListVal, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			vals = append(vals, StringVal(m[1]))
		} else {
			vals = append(vals, StringVal(m[0]))
		}
	}
	return vals, nil
}

// builtinRegexFind: (regex-find pattern text) — first match, returns StringVal or NilVal
func (ev *Evaluator) builtinRegexFind(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("regex-find: need pattern and text")
	}
	patVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	textVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(patVal.String())
	if err != nil {
		return nil, fmt.Errorf("regex-find: %w", err)
	}
	m := re.FindStringSubmatch(textVal.String())
	if m == nil {
		return NilVal{}, nil
	}
	if len(m) > 1 {
		return StringVal(m[1]), nil
	}
	return StringVal(m[0]), nil
}

// toList converts a Value to a ListVal. Strings are split on newlines.
func toList(v Value) ListVal {
	if l, ok := v.(ListVal); ok {
		return l
	}
	s := v.String()
	if s == "" {
		return ListVal{}
	}
	parts := strings.Split(s, "\n")
	result := make(ListVal, len(parts))
	for i, p := range parts {
		result[i] = StringVal(p)
	}
	return result
}

// applyValue calls a function value with a single pre-evaluated argument.
func (ev *Evaluator) applyValue(env *Env, fn Value, arg Value) (Value, error) {
	switch f := fn.(type) {
	case *FnVal:
		child := NewEnv(f.Env)
		if len(f.Params) > 0 {
			child.Set(f.Params[0], arg)
		}
		var result Value = NilVal{}
		var err error
		for _, body := range f.Body {
			result, err = ev.Eval(child, body)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("not callable: %v", fn)
	}
}

// builtinCount: (count list) or (count string) — returns length as StringVal
func (ev *Evaluator) builtinCount(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("count: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	switch val := v.(type) {
	case ListVal:
		return StringVal(strconv.Itoa(len(val))), nil
	case *MapVal:
		return StringVal(strconv.Itoa(len(val.Keys))), nil
	default:
		return StringVal(strconv.Itoa(len(v.String()))), nil
	}
}

// builtinSome: (some list pred) — true if any item matches predicate fn
func (ev *Evaluator) builtinSome(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("some: need list and predicate")
	}
	listVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	predVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	items := toList(listVal)
	for _, item := range items {
		result, err := ev.applyValue(env, predVal, item)
		if err != nil {
			return nil, err
		}
		if isTruthy(result) {
			return BoolVal(true), nil
		}
	}
	return BoolVal(false), nil
}

// builtinEvery: (every list pred) — true if all items match predicate fn
func (ev *Evaluator) builtinEvery(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("every: need list and predicate")
	}
	listVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	predVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	items := toList(listVal)
	for _, item := range items {
		result, err := ev.applyValue(env, predVal, item)
		if err != nil {
			return nil, err
		}
		if !isTruthy(result) {
			return BoolVal(false), nil
		}
	}
	return BoolVal(true), nil
}

// builtinSet: (set list) — deduplicate preserving order
func (ev *Evaluator) builtinSet(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("set: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	items := toList(v)
	seen := make(map[string]bool)
	result := make(ListVal, 0, len(items))
	for _, item := range items {
		key := item.String()
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result, nil
}

// builtinDifference: (difference set-a set-b) — items in a not in b
func (ev *Evaluator) builtinDifference(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("difference: need two lists")
	}
	aVal, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	bVal, err := ev.Eval(env, args[1])
	if err != nil {
		return nil, err
	}
	aItems := toList(aVal)
	bItems := toList(bVal)
	bSet := make(map[string]bool, len(bItems))
	for _, item := range bItems {
		bSet[item.String()] = true
	}
	result := make(ListVal, 0, len(aItems))
	for _, item := range aItems {
		if !bSet[item.String()] {
			result = append(result, item)
		}
	}
	return result, nil
}

// builtinFlatten: (flatten list-of-lists) — flatten one level
func (ev *Evaluator) builtinFlatten(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("flatten: missing argument")
	}
	v, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	outer, ok := v.(ListVal)
	if !ok {
		return v, nil
	}
	result := make(ListVal, 0)
	for _, item := range outer {
		if inner, ok := item.(ListVal); ok {
			result = append(result, inner...)
		} else {
			result = append(result, item)
		}
	}
	return result, nil
}

// builtinAssert: (assert condition message) — error if falsy
func (ev *Evaluator) builtinAssert(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("assert: need condition and message")
	}
	cond, err := ev.Eval(env, args[0])
	if err != nil {
		return nil, err
	}
	if !isTruthy(cond) {
		msgVal, err := ev.Eval(env, args[1])
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("assert: %s", msgVal.String())
	}
	return BoolVal(true), nil
}

// builtinLessThan: (< a b) — numeric less-than, falls back to string comparison
func (ev *Evaluator) builtinLessThan(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("<: need two arguments")
	}
	// Handle bare numeric atoms (parsed as symbols)
	var aStr, bStr string
	if args[0].IsAtom() && args[0].Atom.Type == sexpr.TokenSymbol {
		aStr = args[0].Atom.Val
	} else {
		av, err := ev.Eval(env, args[0])
		if err != nil {
			return nil, err
		}
		aStr = av.String()
	}
	if args[1].IsAtom() && args[1].Atom.Type == sexpr.TokenSymbol {
		bStr = args[1].Atom.Val
	} else {
		bv, err := ev.Eval(env, args[1])
		if err != nil {
			return nil, err
		}
		bStr = bv.String()
	}
	aF, aErr := strconv.ParseFloat(aStr, 64)
	bF, bErr := strconv.ParseFloat(bStr, 64)
	if aErr == nil && bErr == nil {
		return BoolVal(aF < bF), nil
	}
	return BoolVal(aStr < bStr), nil
}

// Ensure exec is available even if provider.RunShell signature changes.
var _ = exec.Command
