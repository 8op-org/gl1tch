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

	// JSON
	b("json-pick", ev.builtinJsonPick)
	b("pick", ev.builtinJsonPick)
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

// builtinJsonPick: (json-pick "expr" data) — placeholder for jq integration
func (ev *Evaluator) builtinJsonPick(_ *Evaluator, env *Env, args []*sexpr.Node) (Value, error) {
	if len(args) == 0 {
		return NilVal{}, nil
	}
	// For now, evaluate and return the last arg (passthrough)
	var result Value = NilVal{}
	for _, a := range args {
		v, err := ev.Eval(env, a)
		if err != nil {
			return nil, err
		}
		result = v
	}
	return result, nil
}

// Ensure exec is available even if provider.RunShell signature changes.
var _ = exec.Command
