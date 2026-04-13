package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// Provider is a YAML-defined command template for an LLM tool.
type Provider struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// ProviderRegistry holds loaded provider definitions.
type ProviderRegistry struct {
	providers map[string]*Provider
}

// LoadProviders reads all .yaml files from dir, parses them into Provider
// structs, and returns a registry. If dir doesn't exist, returns an empty
// registry (not an error).
func LoadProviders(dir string) (*ProviderRegistry, error) {
	reg := &ProviderRegistry{providers: make(map[string]*Provider)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, fmt.Errorf("read provider dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		var p Provider
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		if p.Name == "" {
			continue
		}
		reg.providers[p.Name] = &p
	}
	return reg, nil
}

// RenderCommand looks up a provider by name and renders its command template
// with {{.prompt}} replaced by the given prompt. Returns an error listing
// available providers if name is not found.
func (r *ProviderRegistry) RenderCommand(name, prompt string) (string, error) {
	p, ok := r.providers[name]
	if !ok {
		available := make([]string, 0, len(r.providers))
		for k := range r.providers {
			available = append(available, k)
		}
		sort.Strings(available)
		return "", fmt.Errorf("provider %q not found; available: %s", name, strings.Join(available, ", "))
	}

	tmpl, err := template.New("cmd").Parse(p.Command)
	if err != nil {
		return "", fmt.Errorf("bad template for provider %q: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"prompt": prompt}); err != nil {
		return "", fmt.Errorf("render provider %q: %w", name, err)
	}
	return buf.String(), nil
}

// RunProvider looks up a provider by name and executes the command.
// If the command template contains {{.prompt}}, the prompt is rendered inline.
// Otherwise the prompt is piped via stdin (avoids shell escaping for long prompts).
func (r *ProviderRegistry) RunProvider(name, prompt string) (string, error) {
	p, ok := r.providers[name]
	if !ok {
		avail := make([]string, 0)
		for n := range r.providers {
			avail = append(avail, n)
		}
		return "", fmt.Errorf("provider %q not found (available: %s)", name, strings.Join(avail, ", "))
	}

	if strings.Contains(p.Command, "{{.prompt}}") {
		rendered, err := r.RenderCommand(name, prompt)
		if err != nil {
			return "", err
		}
		return RunShell(rendered)
	}
	return RunShellWithStdin(p.Command, prompt)
}

// RunShellWithStdin executes a shell command with prompt piped via stdin.
func RunShellWithStdin(command, stdin string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("shell: %w\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunShell executes a shell command and returns its stdout.
func RunShell(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("shell: %w\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunOllama sends a prompt to an Ollama model via the HTTP API.
func RunOllama(model, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama: %s\n%s", resp.Status, data)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("ollama: parse: %w", err)
	}
	return strings.TrimSpace(result.Response), nil
}

// RunOllamaWithResult sends a prompt to Ollama and returns a full LLMResult
// with token counts parsed from the Ollama response metadata.
func RunOllamaWithResult(model, prompt string) (LLMResult, error) {
	start := time.Now()

	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return LLMResult{}, fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResult{}, fmt.Errorf("ollama: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return LLMResult{}, fmt.Errorf("ollama: %s\n%s", resp.Status, data)
	}

	var raw struct {
		Response        string `json:"response"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return LLMResult{}, fmt.Errorf("ollama: parse: %w", err)
	}

	tokIn := raw.PromptEvalCount
	if tokIn == 0 {
		tokIn = EstimateTokens(prompt)
	}
	tokOut := raw.EvalCount
	if tokOut == 0 {
		tokOut = EstimateTokens(raw.Response)
	}

	return LLMResult{
		Provider:  "ollama",
		Model:     model,
		Response:  strings.TrimSpace(raw.Response),
		TokensIn:  tokIn,
		TokensOut: tokOut,
		Latency:   time.Since(start),
		CostUSD:   0,
	}, nil
}

