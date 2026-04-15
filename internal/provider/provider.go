package provider

import (
	"bytes"
	"context"
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
func (r *ProviderRegistry) RenderCommand(name string, data map[string]string) (string, error) {
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
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render provider %q: %w", name, err)
	}
	return buf.String(), nil
}

// RunProvider looks up a provider by name and executes the command.
// If the command template contains {{.prompt}}, the prompt is rendered inline.
// Otherwise the prompt is piped via stdin (avoids shell escaping for long prompts).
func (r *ProviderRegistry) RunProvider(name, model, prompt string) (string, error) {
	p, ok := r.providers[name]
	if !ok {
		avail := make([]string, 0)
		for n := range r.providers {
			avail = append(avail, n)
		}
		return "", fmt.Errorf("provider %q not found (available: %s)", name, strings.Join(avail, ", "))
	}

	data := map[string]string{"prompt": prompt, "model": model}
	hasPromptTpl := strings.Contains(p.Command, "{{.prompt}}")
	hasModelTpl := strings.Contains(p.Command, "{{.model}}")

	if hasPromptTpl {
		// Both prompt and model rendered inline
		rendered, err := r.RenderCommand(name, data)
		if err != nil {
			return "", err
		}
		return RunShell(rendered)
	}
	if hasModelTpl {
		// Render model into command, pipe prompt via stdin
		rendered, err := r.RenderCommand(name, data)
		if err != nil {
			return "", err
		}
		return RunShellWithStdin(rendered, prompt)
	}
	return RunShellWithStdin(p.Command, prompt)
}

// RunProviderWithResult runs a provider and returns a structured LLMResult.
// It parses copilot-style footers to extract token counts and premium requests.
func (r *ProviderRegistry) RunProviderWithResult(name, model, prompt string) (LLMResult, error) {
	start := time.Now()
	raw, err := r.RunProvider(name, model, prompt)
	if err != nil {
		return LLMResult{}, err
	}

	result := LLMResult{
		Provider: name,
		Model:    model,
		Latency:  time.Since(start),
	}

	// Parse copilot-style footer: "Changes +N -N\nRequests N Premium\nTokens ..."
	if idx := strings.Index(raw, "\nChanges "); idx >= 0 {
		footer := raw[idx:]
		result.Response = strings.TrimSpace(raw[:idx])
		result.TokensIn, result.TokensOut, result.CostUSD = parseCopilotFooter(footer, name)
	} else {
		result.Response = raw
		result.TokensIn = EstimateTokens(prompt)
		result.TokensOut = EstimateTokens(raw)
		result.CostUSD = EstimateCost(name, result.TokensIn, result.TokensOut)
	}

	return result, nil
}

// parseCopilotFooter extracts token counts and premium request cost from
// copilot CLI footer output like:
//
//	Changes   +0 -0
//	Requests  1 Premium (4s)
//	Tokens    ↑ 16.6k • ↓ 21 • 15.7k (cached)
func parseCopilotFooter(footer, providerName string) (tokIn, tokOut int, cost float64) {
	for _, line := range strings.Split(footer, "\n") {
		line = strings.TrimSpace(line)

		// Parse "Requests  N Premium"
		if strings.HasPrefix(line, "Requests") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "Premium" && i > 0 {
					n := 0
					fmt.Sscanf(parts[i-1], "%d", &n)
					// GitHub Copilot premium request pricing
					cost = float64(n) * premiumRequestCost(providerName)
				}
			}
		}

		// Parse "Tokens  ↑ 16.6k • ↓ 21 • ..."
		if strings.HasPrefix(line, "Tokens") {
			// Find input tokens (after ↑) and output tokens (after ↓)
			// The line uses unicode arrows and k-suffixes
			parts := strings.Split(line, "•")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "↑") || strings.Contains(part, "\u2191") {
					tokIn = parseTokenCount(part)
				} else if strings.Contains(part, "↓") || strings.Contains(part, "\u2193") {
					tokOut = parseTokenCount(part)
				}
			}
		}
	}
	return
}

// parseTokenCount extracts a number from strings like "↑ 16.6k" or "↓ 21"
func parseTokenCount(s string) int {
	// Strip everything that's not a digit, dot, or 'k'
	var numStr string
	hasK := false
	for _, c := range s {
		if (c >= '0' && c <= '9') || c == '.' {
			numStr += string(c)
		} else if c == 'k' || c == 'K' {
			hasK = true
		}
	}
	if numStr == "" {
		return 0
	}
	var n float64
	fmt.Sscanf(numStr, "%f", &n)
	if hasK {
		n *= 1000
	}
	return int(n)
}

// premiumRequestCost returns the cost per premium request for a provider.
func premiumRequestCost(provider string) float64 {
	// GitHub Copilot Business: premium requests are ~$0.04 each
	// (based on $4/100 premium requests overage pricing)
	return 0.04
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
	return RunShellContext(context.Background(), command)
}

// RunShellContext executes a shell command with context support for cancellation.
func RunShellContext(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
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

