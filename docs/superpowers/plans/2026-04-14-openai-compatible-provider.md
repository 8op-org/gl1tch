# OpenAI-Compatible Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a native OpenAI-compatible HTTP client so glitch can use OpenRouter (and any `/v1/chat/completions` endpoint) in pipeline steps and tiered escalation.

**Architecture:** New `OpenAICompatibleProvider` in `internal/provider/openai.go` speaks the OpenAI chat completions protocol. Config struct in `cmd/config.go` gains a `Providers` map for named provider definitions. Pipeline runner and tiered runner resolve provider names against this map before falling back to the shell-template registry.

**Tech Stack:** Go stdlib `net/http`, `encoding/json`; existing `provider.LLMResult`; `gopkg.in/yaml.v3` for config.

---

### Task 1: OpenAI-compatible client — types and Chat method

**Files:**
- Create: `internal/provider/openai.go`
- Create: `internal/provider/openai_test.go`

- [ ] **Step 1: Write the failing test for Chat() with a mock HTTP server**

```go
package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIChat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request structure
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("auth header = %q, want Bearer test-key", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content-type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		var req struct {
			Model    string              `json:"model"`
			Messages []map[string]string `json:"messages"`
			Stream   bool                `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "meta-llama/llama-4-scout:free" {
			t.Errorf("model = %q, want meta-llama/llama-4-scout:free", req.Model)
		}
		if req.Stream {
			t.Error("stream = true, want false")
		}
		if len(req.Messages) != 1 || req.Messages[0]["role"] != "user" || req.Messages[0]["content"] != "hello" {
			t.Errorf("messages = %v, want [{role:user content:hello}]", req.Messages)
		}

		w.Header().Set("x-openrouter-cost", "0.00042")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "world"}},
			},
			"usage": map[string]int{
				"prompt_tokens":     5,
				"completion_tokens": 1,
			},
		})
	}))
	defer srv.Close()

	p := &OpenAICompatibleProvider{
		Name:    "openrouter",
		BaseURL: srv.URL,
		APIKey:  "test-key",
	}

	result, err := p.Chat("meta-llama/llama-4-scout:free", "hello")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Response != "world" {
		t.Errorf("response = %q, want world", result.Response)
	}
	if result.Provider != "openrouter" {
		t.Errorf("provider = %q, want openrouter", result.Provider)
	}
	if result.Model != "meta-llama/llama-4-scout:free" {
		t.Errorf("model = %q, want meta-llama/llama-4-scout:free", result.Model)
	}
	if result.TokensIn != 5 {
		t.Errorf("tokens_in = %d, want 5", result.TokensIn)
	}
	if result.TokensOut != 1 {
		t.Errorf("tokens_out = %d, want 1", result.TokensOut)
	}
	if result.CostUSD < 0.00041 || result.CostUSD > 0.00043 {
		t.Errorf("cost = %f, want ~0.00042", result.CostUSD)
	}
	if result.Latency <= 0 {
		t.Error("latency should be positive")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestOpenAIChat_Success -v`
Expected: FAIL — `OpenAICompatibleProvider` not defined

- [ ] **Step 3: Write minimal implementation**

Create `internal/provider/openai.go`:

```go
package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OpenAICompatibleProvider calls any OpenAI-compatible chat completions API.
type OpenAICompatibleProvider struct {
	Name         string
	BaseURL      string // e.g. "https://openrouter.ai/api/v1"
	APIKey       string
	DefaultModel string
}

// Chat sends a prompt to the chat completions endpoint and returns a structured LLMResult.
func (p *OpenAICompatibleProvider) Chat(model, prompt string) (LLMResult, error) {
	if model == "" {
		model = p.DefaultModel
	}

	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: marshal: %w", err)
	}

	start := time.Now()

	url := strings.TrimRight(p.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return LLMResult{}, fmt.Errorf("openai: %s\n%s", resp.Status, data)
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return LLMResult{}, fmt.Errorf("openai: parse: %w", err)
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = strings.TrimSpace(chatResp.Choices[0].Message.Content)
	}

	cost := 0.0
	if h := resp.Header.Get("x-openrouter-cost"); h != "" {
		cost, _ = strconv.ParseFloat(h, 64)
	}

	return LLMResult{
		Provider:  p.Name,
		Model:     model,
		Response:  content,
		TokensIn:  chatResp.Usage.PromptTokens,
		TokensOut: chatResp.Usage.CompletionTokens,
		Latency:   time.Since(start),
		CostUSD:   cost,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestOpenAIChat_Success -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/openai.go internal/provider/openai_test.go
git commit -m "feat: add OpenAI-compatible provider client with Chat method"
```

---

### Task 2: Error handling and edge cases for Chat

**Files:**
- Modify: `internal/provider/openai_test.go`

- [ ] **Step 1: Write failing tests for error paths**

Append to `internal/provider/openai_test.go`:

```go
func TestOpenAIChat_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()

	p := &OpenAICompatibleProvider{Name: "test", BaseURL: srv.URL, APIKey: "k"}
	_, err := p.Chat("model", "hello")
	if err == nil {
		t.Fatal("expected error on 429")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error should mention status code: %v", err)
	}
}

func TestOpenAIChat_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{},
			"usage":   map[string]int{"prompt_tokens": 5, "completion_tokens": 0},
		})
	}))
	defer srv.Close()

	p := &OpenAICompatibleProvider{Name: "test", BaseURL: srv.URL, APIKey: "k"}
	result, err := p.Chat("model", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "" {
		t.Errorf("response = %q, want empty", result.Response)
	}
}

func TestOpenAIChat_DefaultModel(t *testing.T) {
	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Model string }
		json.NewDecoder(r.Body).Decode(&req)
		gotModel = req.Model
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
			"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1},
		})
	}))
	defer srv.Close()

	p := &OpenAICompatibleProvider{
		Name: "test", BaseURL: srv.URL, APIKey: "k",
		DefaultModel: "fallback-model",
	}
	_, err := p.Chat("", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if gotModel != "fallback-model" {
		t.Errorf("model = %q, want fallback-model", gotModel)
	}
}

func TestOpenAIChat_NoCostHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No x-openrouter-cost header
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
			"usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 5},
		})
	}))
	defer srv.Close()

	p := &OpenAICompatibleProvider{Name: "test", BaseURL: srv.URL, APIKey: "k"}
	result, err := p.Chat("model", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if result.CostUSD != 0 {
		t.Errorf("cost = %f, want 0 when no header", result.CostUSD)
	}
}
```

- [ ] **Step 2: Run tests to verify they pass** (implementation already handles these)

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestOpenAIChat -v`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git add internal/provider/openai_test.go
git commit -m "test: add error and edge case tests for OpenAI-compatible client"
```

---

### Task 3: Config struct — add Providers map

**Files:**
- Modify: `cmd/config.go`

- [ ] **Step 1: Write the failing test**

Create `cmd/config_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_WithProviders(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte(`
default_model: qwen3:8b
default_provider: ollama
providers:
  openrouter:
    type: openai-compatible
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_API_KEY
    api_key: sk-or-fallback
    default_model: meta-llama/llama-4-scout:free
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfigFrom(cfgPath)
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("providers count = %d, want 1", len(cfg.Providers))
	}
	p := cfg.Providers["openrouter"]
	if p.Type != "openai-compatible" {
		t.Errorf("type = %q, want openai-compatible", p.Type)
	}
	if p.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("base_url = %q", p.BaseURL)
	}
	if p.APIKeyEnv != "OPENROUTER_API_KEY" {
		t.Errorf("api_key_env = %q", p.APIKeyEnv)
	}
	if p.APIKey != "sk-or-fallback" {
		t.Errorf("api_key = %q", p.APIKey)
	}
	if p.DefaultModel != "meta-llama/llama-4-scout:free" {
		t.Errorf("default_model = %q", p.DefaultModel)
	}
}

func TestProviderConfig_ResolveAPIKey_EnvFirst(t *testing.T) {
	t.Setenv("TEST_KEY_ENV", "from-env")
	pc := ProviderConfig{
		APIKeyEnv: "TEST_KEY_ENV",
		APIKey:    "from-config",
	}
	key, err := pc.ResolveAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if key != "from-env" {
		t.Errorf("key = %q, want from-env", key)
	}
}

func TestProviderConfig_ResolveAPIKey_FallbackToConfig(t *testing.T) {
	pc := ProviderConfig{
		APIKeyEnv: "NONEXISTENT_VAR_12345",
		APIKey:    "from-config",
	}
	key, err := pc.ResolveAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if key != "from-config" {
		t.Errorf("key = %q, want from-config", key)
	}
}

func TestProviderConfig_ResolveAPIKey_NeitherSet(t *testing.T) {
	pc := ProviderConfig{}
	_, err := pc.ResolveAPIKey()
	if err == nil {
		t.Fatal("expected error when no API key configured")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run TestLoadConfig_WithProviders -v`
Expected: FAIL — `ProviderConfig` not defined, `loadConfigFrom` not defined

- [ ] **Step 3: Add ProviderConfig and update Config struct**

Edit `cmd/config.go` — add `ProviderConfig` struct and `Providers` field:

```go
// Add after the Config struct definition:

// ProviderConfig defines a named LLM provider endpoint.
type ProviderConfig struct {
	Type         string `yaml:"type"`                    // "openai-compatible"
	BaseURL      string `yaml:"base_url"`
	APIKeyEnv    string `yaml:"api_key_env,omitempty"`   // env var name to check first
	APIKey       string `yaml:"api_key,omitempty"`       // fallback literal key
	DefaultModel string `yaml:"default_model,omitempty"`
}

// ResolveAPIKey returns the API key, checking the environment variable first.
func (pc *ProviderConfig) ResolveAPIKey() (string, error) {
	if pc.APIKeyEnv != "" {
		if v := os.Getenv(pc.APIKeyEnv); v != "" {
			return v, nil
		}
	}
	if pc.APIKey != "" {
		return pc.APIKey, nil
	}
	return "", fmt.Errorf("no API key: set %s env var or api_key in config", pc.APIKeyEnv)
}
```

Add `Providers` field to `Config`:

```go
type Config struct {
	DefaultModel    string                    `yaml:"default_model"`
	DefaultProvider string                    `yaml:"default_provider"`
	Tiers           []provider.TierConfig     `yaml:"tiers,omitempty"`
	Providers       map[string]ProviderConfig `yaml:"providers,omitempty"`
}
```

Extract `loadConfigFrom` so tests can point to a temp file:

```go
func loadConfigFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{
			DefaultModel:    "qwen3:8b",
			DefaultProvider: "ollama",
			Tiers:           provider.DefaultTiers(),
		}, nil
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Tiers) == 0 {
		cfg.Tiers = provider.DefaultTiers()
	}
	return &cfg, nil
}

func loadConfig() (*Config, error) {
	return loadConfigFrom(configPath())
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./cmd/ -run "TestLoadConfig_WithProviders|TestProviderConfig_ResolveAPIKey" -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/config.go cmd/config_test.go
git commit -m "feat: add ProviderConfig map to config for named LLM endpoints"
```

---

### Task 4: Wire OpenAI-compatible providers into TieredRunner

**Files:**
- Modify: `internal/provider/tiers.go`
- Modify: `internal/provider/tiers_test.go`

The `TieredRunner` needs to accept a resolver function so it can look up OpenAI-compatible providers by name. We pass a callback rather than importing `cmd` (avoids circular dependency).

- [ ] **Step 1: Write the failing test**

Append to `internal/provider/tiers_test.go`:

```go
func TestTieredRunner_WithResolver(t *testing.T) {
	called := false
	fakeProvider := func(model, prompt string) (LLMResult, error) {
		called = true
		return LLMResult{
			Provider: "openrouter",
			Model:    model,
			Response: "resolved-response",
			TokensIn: 10, TokensOut: 5,
		}, nil
	}

	resolver := func(name string) (func(model, prompt string) (LLMResult, error), bool) {
		if name == "openrouter" {
			return fakeProvider, true
		}
		return nil, false
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"openrouter"}, Model: "test-model"},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	result, err := runner.Run(context.Background(), "hello", func(s string) EscalationReason {
		return "" // accept anything
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("resolver provider was not called")
	}
	if result.Response != "resolved-response" {
		t.Errorf("response = %q, want resolved-response", result.Response)
	}
	if result.Provider != "openrouter" {
		t.Errorf("provider = %q, want openrouter", result.Provider)
	}
}

func TestTieredRunner_ResolverFallsBackToRegistry(t *testing.T) {
	// Resolver returns false → falls back to registry → provider not found → error
	resolver := func(name string) (func(model, prompt string) (LLMResult, error), bool) {
		return nil, false
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"unknown"}},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	_, err := runner.Run(context.Background(), "hello", func(s string) EscalationReason {
		return ""
	})
	if err == nil {
		t.Fatal("expected error when resolver and registry both miss")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestTieredRunner_WithResolver -v`
Expected: FAIL — `Resolver` field not defined on `TieredRunner`

- [ ] **Step 3: Add Resolver field and update callProvider**

Edit `internal/provider/tiers.go`:

Add `Resolver` field to `TieredRunner`:

```go
// ProviderFunc calls an LLM provider and returns a result.
type ProviderFunc func(model, prompt string) (LLMResult, error)

// ResolverFunc looks up a provider by name. Returns the call function and true if found.
type ResolverFunc func(name string) (ProviderFunc, bool)

type TieredRunner struct {
	tiers    []TierConfig
	reg      *ProviderRegistry
	Resolver ResolverFunc
}
```

Update `callProvider` to check Resolver before the registry:

```go
func (tr *TieredRunner) callProvider(name, model, prompt string) (LLMResult, error) {
	start := time.Now()
	_ = start // used below

	if name == "ollama" {
		if model == "" {
			model = "qwen3:8b"
		}
		return RunOllamaWithResult(model, prompt)
	}

	// Check resolver (openai-compatible providers from config)
	if tr.Resolver != nil {
		if fn, ok := tr.Resolver(name); ok {
			return fn(model, prompt)
		}
	}

	raw, err := tr.reg.RunProvider(name, model, prompt)
	if err != nil {
		return LLMResult{}, err
	}

	tokIn := EstimateTokens(prompt)
	tokOut := EstimateTokens(raw)
	return LLMResult{
		Provider:  name,
		Model:     model,
		Response:  raw,
		TokensIn:  tokIn,
		TokensOut: tokOut,
		Latency:   time.Since(start),
		CostUSD:   EstimateCost(name, tokIn, tokOut),
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestTieredRunner -v`
Expected: All PASS (including existing tests)

- [ ] **Step 5: Commit**

```bash
git add internal/provider/tiers.go internal/provider/tiers_test.go
git commit -m "feat: add Resolver to TieredRunner for openai-compatible providers"
```

---

### Task 5: Wire OpenAI-compatible providers into pipeline runner

**Files:**
- Modify: `internal/pipeline/runner.go`

The pipeline runner's `Run` function needs to resolve `provider: openrouter` in LLM steps. We add a `ProviderResolver` field to `RunOpts` (same `ResolverFunc` type from the provider package).

- [ ] **Step 1: Write the failing test**

Append to `internal/pipeline/runner_test.go`:

```go
func TestRun_OpenAICompatibleProvider(t *testing.T) {
	called := false
	resolver := func(name string) (provider.ProviderFunc, bool) {
		if name == "openrouter" {
			return func(model, prompt string) (provider.LLMResult, error) {
				called = true
				return provider.LLMResult{
					Provider: "openrouter",
					Model:    model,
					Response: "llm-output",
					TokensIn: 10, TokensOut: 5,
				}, nil
			}, true
		}
		return nil, false
	}

	w := &Workflow{
		Name: "test-openai",
		Steps: []Step{
			{
				ID: "ask",
				LLM: &LLMStep{
					Provider: "openrouter",
					Model:    "meta-llama/llama-4-scout:free",
					Prompt:   "say hello",
				},
			},
		},
	}

	reg := &provider.ProviderRegistry{}
	result, err := Run(w, "", "qwen3:8b", nil, reg, RunOpts{
		ProviderResolver: resolver,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Fatal("resolver was not called for openrouter provider")
	}
	if result.Output != "llm-output" {
		t.Errorf("output = %q, want llm-output", result.Output)
	}
}
```

Note: `ProviderRegistry` may need to be exported or the test may need adjustment depending on existing test patterns. Check `internal/pipeline/runner_test.go` for how `reg` is constructed in existing tests.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestRun_OpenAICompatibleProvider -v`
Expected: FAIL — `ProviderResolver` field not defined on `RunOpts`

- [ ] **Step 3: Add ProviderResolver to RunOpts and wire into the switch**

Edit `internal/pipeline/runner.go`:

Add to `RunOpts`:

```go
type RunOpts struct {
	Telemetry        *esearch.Telemetry
	Issue            string
	ComparisonGroup  string
	ProviderResolver provider.ResolverFunc
}
```

In the LLM step switch block (around line 141), add a resolver check between the `case "ollama", ""` and the `default`:

```go
			switch prov {
			case "ollama", "":
				// ... existing ollama code unchanged ...
			default:
				// Try resolver first (openai-compatible providers from config)
				var resolved bool
				if opts[0].ProviderResolver != nil {
					if fn, ok := opts[0].ProviderResolver(prov); ok {
						resolved = true
						result, llmErr := fn(model, rendered)
						if llmErr != nil {
							err = llmErr
						} else {
							out = result.Response
							tokIn := int64(result.TokensIn)
							tokOut := int64(result.TokensOut)
							totalTokensIn += tokIn
							totalTokensOut += tokOut
							totalCostUSD += result.CostUSD
							totalLatencyMS += result.Latency.Milliseconds()
							llmSteps++
							lastLLMOutput = out
							if tel != nil {
								tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
									RunID:           runID,
									Step:            fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
									Tier:            1,
									Provider:        prov,
									Model:           model,
									TokensIn:        tokIn,
									TokensOut:       tokOut,
									TokensTotal:     tokIn + tokOut,
									CostUSD:         result.CostUSD,
									LatencyMS:       result.Latency.Milliseconds(),
									WorkflowName:    w.Name,
									Issue:           issue,
									ComparisonGroup: compGroup,
									Timestamp:       time.Now().UTC().Format(time.RFC3339),
								})
							}
						}
					}
				}
				if !resolved {
					// Fall back to shell-template registry
					result, provErr := reg.RunProviderWithResult(prov, model, rendered)
					// ... existing code unchanged ...
				}
			}
```

Note: Guard the `opts[0]` access — if `len(opts) == 0`, `ProviderResolver` is nil. The simplest approach: extract the resolver once at the top of `Run`:

```go
var providerResolver provider.ResolverFunc
if len(opts) > 0 {
	// ... existing opts extraction ...
	providerResolver = opts[0].ProviderResolver
}
```

Then use `providerResolver` in the switch.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat: wire openai-compatible provider resolver into pipeline runner"
```

---

### Task 6: Build the resolver in cmd layer and pass it through

**Files:**
- Modify: Where `Run()` is called from CLI commands (find the cobra command that invokes `pipeline.Run`)

- [ ] **Step 1: Find the call site**

Run: `cd /Users/stokes/Projects/gl1tch && grep -rn "pipeline.Run(" cmd/`

This will show where `pipeline.Run` is called. The resolver needs to be built from `Config.Providers` and passed as `RunOpts.ProviderResolver`.

- [ ] **Step 2: Write the resolver builder function**

Add to `cmd/config.go`:

```go
// BuildProviderResolver creates a ResolverFunc from the config's Providers map.
func (cfg *Config) BuildProviderResolver() provider.ResolverFunc {
	return func(name string) (provider.ProviderFunc, bool) {
		pc, ok := cfg.Providers[name]
		if !ok || pc.Type != "openai-compatible" {
			return nil, false
		}
		key, err := pc.ResolveAPIKey()
		if err != nil {
			return nil, false
		}
		p := &provider.OpenAICompatibleProvider{
			Name:         name,
			BaseURL:      pc.BaseURL,
			APIKey:       key,
			DefaultModel: pc.DefaultModel,
		}
		return p.Chat, true
	}
}
```

- [ ] **Step 3: Pass resolver at call sites**

At each `pipeline.Run(...)` call, add `ProviderResolver: cfg.BuildProviderResolver()` to the `RunOpts`. The exact edit depends on Step 1's output — the pattern is:

```go
// Before:
pipeline.Run(w, input, cfg.DefaultModel, params, reg, pipeline.RunOpts{
	Telemetry: tel,
})

// After:
pipeline.Run(w, input, cfg.DefaultModel, params, reg, pipeline.RunOpts{
	Telemetry:        tel,
	ProviderResolver: cfg.BuildProviderResolver(),
})
```

- [ ] **Step 4: Similarly wire the resolver into TieredRunner construction**

Wherever `NewTieredRunner` is called, set `Resolver`:

```go
tr := provider.NewTieredRunner(cfg.Tiers, reg)
tr.Resolver = cfg.BuildProviderResolver()
```

- [ ] **Step 5: Run all tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -20`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/
git commit -m "feat: build and pass provider resolver from config to runner and tiers"
```

---

### Task 7: Manual integration test

**Files:** None (verification only)

- [ ] **Step 1: Create an OpenRouter API key**

Go to https://openrouter.ai/keys and create one.

- [ ] **Step 2: Set the environment variable**

```bash
export OPENROUTER_API_KEY=sk-or-v1-...
```

- [ ] **Step 3: Add OpenRouter to glitch config**

```bash
cat >> ~/.config/glitch/config.yaml << 'EOF'
providers:
  openrouter:
    type: openai-compatible
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_API_KEY
    default_model: meta-llama/llama-4-scout:free
EOF
```

- [ ] **Step 4: Create a test workflow**

Write `.glitch/workflows/test-openrouter.yaml`:

```yaml
name: test-openrouter
description: Smoke test for OpenRouter integration
steps:
  - id: ask
    llm:
      provider: openrouter
      model: meta-llama/llama-4-scout:free
      prompt: "Say hello in exactly 5 words."
```

- [ ] **Step 5: Run it**

```bash
glitch pipeline run test-openrouter
```

Expected: A 5-word greeting from Llama via OpenRouter.

- [ ] **Step 6: Test tier escalation**

Update `~/.config/glitch/config.yaml` tiers to include `openrouter`:

```yaml
tiers:
  - providers: [ollama]
    model: qwen3:8b
  - providers: [openrouter]
    model: meta-llama/llama-4-scout:free
  - providers: [copilot, claude]
```

Stop Ollama, run a tiered workflow, and verify it escalates to OpenRouter.

- [ ] **Step 7: Clean up test workflow**

```bash
rm .glitch/workflows/test-openrouter.yaml
```
