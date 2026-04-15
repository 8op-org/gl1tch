# LM Studio Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add LM Studio as a first-class local LLM provider in gl1tch, with auto-download and load-state awareness.

**Architecture:** Standalone `LMStudioProvider` struct in `internal/provider/lmstudio.go` that uses OpenAI chat completions format but adds LM Studio-specific model management via the native `/api/v0/models` and `/api/v1/models/download` REST endpoints. Wired into the pipeline runner and tiered runner as `"lm-studio"`.

**Tech Stack:** Go, net/http, net/http/httptest, encoding/json

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/provider/lmstudio.go` | Create | LMStudioProvider struct, Chat(), model check, model download |
| `internal/provider/lmstudio_test.go` | Create | Unit tests with httptest mocks |
| `internal/provider/tokens.go` | Modify line 27 | Add `"lm-studio"` pricing entry |
| `internal/provider/tiers.go` | Modify lines 106-111 | Add `case "lm-studio"` in callProvider() |
| `internal/pipeline/runner.go` | Modify lines 778-786 | Add `case "lm-studio"` in provider switch |

---

### Task 1: LMStudioProvider struct and model check

**Files:**
- Create: `internal/provider/lmstudio.go`
- Create: `internal/provider/lmstudio_test.go`

- [ ] **Step 1: Write the failing test for model availability check**

```go
package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLMStudio_CheckModels_Found_Loaded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":    "qwen3-8b",
					"state": "loaded",
				},
			},
		})
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	exists, loaded, err := p.checkModels("qwen3-8b")
	if err != nil {
		t.Fatalf("checkModels: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if !loaded {
		t.Error("expected loaded=true")
	}
}

func TestLMStudio_CheckModels_Found_NotLoaded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":    "qwen3-8b",
					"state": "not-loaded",
				},
			},
		})
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	exists, loaded, err := p.checkModels("qwen3-8b")
	if err != nil {
		t.Fatalf("checkModels: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if loaded {
		t.Error("expected loaded=false")
	}
}

func TestLMStudio_CheckModels_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	exists, _, err := p.checkModels("missing-model")
	if err != nil {
		t.Fatalf("checkModels: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestLMStudio_CheckModels -v`
Expected: FAIL — `LMStudioProvider` not defined

- [ ] **Step 3: Write LMStudioProvider struct and checkModels**

Create `internal/provider/lmstudio.go`:

```go
package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LMStudioProvider calls LM Studio's OpenAI-compatible API with native model management.
type LMStudioProvider struct {
	BaseURL      string // default "http://localhost:1234"
	DefaultModel string // default "qwen3-8b"
}

// checkModels queries the native API to determine if a model exists locally and whether it's loaded.
func (p *LMStudioProvider) checkModels(model string) (exists, loaded bool, err error) {
	url := strings.TrimRight(p.BaseURL, "/") + "/api/v0/models"
	resp, err := http.Get(url)
	if err != nil {
		return false, false, fmt.Errorf("lm-studio: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, false, fmt.Errorf("lm-studio: read models: %w", err)
	}
	if resp.StatusCode != 200 {
		return false, false, fmt.Errorf("lm-studio: models: %s", resp.Status)
	}

	var result struct {
		Data []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, false, fmt.Errorf("lm-studio: parse models: %w", err)
	}

	for _, m := range result.Data {
		if m.ID == model || strings.Contains(m.ID, model) {
			return true, m.State == "loaded", nil
		}
	}
	return false, false, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestLMStudio_CheckModels -v`
Expected: 3 PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/provider/lmstudio.go internal/provider/lmstudio_test.go
git commit -m "feat(provider): add LMStudioProvider struct with model availability check"
```

---

### Task 2: Model auto-download

**Files:**
- Modify: `internal/provider/lmstudio.go`
- Modify: `internal/provider/lmstudio_test.go`

- [ ] **Step 1: Write the failing test for downloadModel**

Append to `internal/provider/lmstudio_test.go`:

```go
func TestLMStudio_DownloadModel_Success(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/models/download":
			var req struct{ Model string `json:"model"` }
			json.NewDecoder(r.Body).Decode(&req)
			if req.Model != "qwen3-8b" {
				t.Errorf("download model = %q, want qwen3-8b", req.Model)
			}
			json.NewEncoder(w).Encode(map[string]string{"job_id": "job-123"})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/models/download/status/job-123":
			callCount++
			status := "downloading"
			if callCount >= 2 {
				status = "completed"
			}
			json.NewEncoder(w).Encode(map[string]any{
				"status":     status,
				"downloaded": callCount * 500_000_000,
				"total_size": 1_000_000_000,
			})

		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	err := p.downloadModel("qwen3-8b")
	if err != nil {
		t.Fatalf("downloadModel: %v", err)
	}
}

func TestLMStudio_DownloadModel_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/models/download":
			json.NewEncoder(w).Encode(map[string]string{"job_id": "job-fail"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/models/download/status/job-fail":
			json.NewEncoder(w).Encode(map[string]any{"status": "failed"})
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	err := p.downloadModel("qwen3-8b")
	if err == nil {
		t.Fatal("expected error on failed download")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("error should mention failure: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestLMStudio_Download -v`
Expected: FAIL — `downloadModel` not defined

- [ ] **Step 3: Implement downloadModel**

Add to `internal/provider/lmstudio.go`:

```go
// downloadModel triggers a model download via LM Studio's native API and polls until complete.
func (p *LMStudioProvider) downloadModel(model string) error {
	base := strings.TrimRight(p.BaseURL, "/")

	// Trigger download
	reqBody, _ := json.Marshal(map[string]string{"model": model})
	resp, err := http.Post(base+"/api/v1/models/download", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("lm-studio: download request: %w", err)
	}
	defer resp.Body.Close()

	var dlResp struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dlResp); err != nil {
		return fmt.Errorf("lm-studio: parse download response: %w", err)
	}

	// Poll for completion
	statusURL := fmt.Sprintf("%s/api/v1/models/download/status/%s", base, dlResp.JobID)
	for {
		sResp, err := http.Get(statusURL)
		if err != nil {
			return fmt.Errorf("lm-studio: download status: %w", err)
		}

		var status struct {
			Status     string `json:"status"`
			Downloaded int64  `json:"downloaded"`
			TotalSize  int64  `json:"total_size"`
		}
		json.NewDecoder(sResp.Body).Decode(&status)
		sResp.Body.Close()

		switch status.Status {
		case "completed":
			return nil
		case "failed":
			return fmt.Errorf("lm-studio: download failed for model %q", model)
		default:
			pct := 0.0
			if status.TotalSize > 0 {
				pct = float64(status.Downloaded) / float64(status.TotalSize) * 100
			}
			fmt.Fprintf(os.Stderr, ">> lm-studio: downloading %s (%.0f%%)...\n", model, pct)
			time.Sleep(2 * time.Second)
		}
	}
}
```

Add `"os"` to the import block.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestLMStudio_Download -v`
Expected: 2 PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/provider/lmstudio.go internal/provider/lmstudio_test.go
git commit -m "feat(provider): add LM Studio model auto-download with progress polling"
```

---

### Task 3: Chat method

**Files:**
- Modify: `internal/provider/lmstudio.go`
- Modify: `internal/provider/lmstudio_test.go`

- [ ] **Step 1: Write the failing test for Chat**

Append to `internal/provider/lmstudio_test.go`:

```go
func TestLMStudio_Chat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/models":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "qwen3-8b", "state": "loaded"},
				},
			})
		case "/v1/chat/completions":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			var req struct {
				Model    string              `json:"model"`
				Messages []map[string]string `json:"messages"`
				Stream   bool                `json:"stream"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			if req.Model != "qwen3-8b" {
				t.Errorf("model = %q, want qwen3-8b", req.Model)
			}
			if req.Stream {
				t.Error("stream = true, want false")
			}
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"role": "assistant", "content": "hello world"}},
				},
				"usage": map[string]int{
					"prompt_tokens":     10,
					"completion_tokens": 2,
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	result, err := p.Chat("qwen3-8b", "say hello")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Response != "hello world" {
		t.Errorf("response = %q, want hello world", result.Response)
	}
	if result.Provider != "lm-studio" {
		t.Errorf("provider = %q, want lm-studio", result.Provider)
	}
	if result.TokensIn != 10 {
		t.Errorf("tokens_in = %d, want 10", result.TokensIn)
	}
	if result.TokensOut != 2 {
		t.Errorf("tokens_out = %d, want 2", result.TokensOut)
	}
	if result.CostUSD != 0 {
		t.Errorf("cost = %f, want 0", result.CostUSD)
	}
	if result.Latency <= 0 {
		t.Error("latency should be positive")
	}
}

func TestLMStudio_Chat_DefaultModel(t *testing.T) {
	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/models":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "qwen3-8b", "state": "loaded"},
				},
			})
		case "/v1/chat/completions":
			var req struct{ Model string }
			json.NewDecoder(r.Body).Decode(&req)
			gotModel = req.Model
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "ok"}},
				},
				"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1},
			})
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	_, err := p.Chat("", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if gotModel != "qwen3-8b" {
		t.Errorf("model = %q, want qwen3-8b", gotModel)
	}
}

func TestLMStudio_Chat_ServerDown(t *testing.T) {
	p := &LMStudioProvider{BaseURL: "http://127.0.0.1:19999", DefaultModel: "qwen3-8b"}
	_, err := p.Chat("qwen3-8b", "hello")
	if err == nil {
		t.Fatal("expected error when server is down")
	}
	if !strings.Contains(err.Error(), "lm-studio") {
		t.Errorf("error should mention lm-studio: %v", err)
	}
}

func TestLMStudio_Chat_ModelNotFound_TriggersDownload(t *testing.T) {
	downloadCalled := false
	modelsCallCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v0/models":
			modelsCallCount++
			if modelsCallCount == 1 {
				// First call: model not found
				json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
			} else {
				// After download: model exists
				json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{"id": "new-model", "state": "not-loaded"},
					},
				})
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/models/download":
			downloadCalled = true
			json.NewEncoder(w).Encode(map[string]string{"job_id": "job-1"})
		case r.URL.Path == "/api/v1/models/download/status/job-1":
			json.NewEncoder(w).Encode(map[string]any{"status": "completed"})
		case r.URL.Path == "/v1/chat/completions":
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "downloaded and ran"}},
				},
				"usage": map[string]int{"prompt_tokens": 5, "completion_tokens": 3},
			})
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b"}
	result, err := p.Chat("new-model", "hello")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if !downloadCalled {
		t.Error("expected download to be triggered")
	}
	if result.Response != "downloaded and ran" {
		t.Errorf("response = %q, want 'downloaded and ran'", result.Response)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run "TestLMStudio_Chat" -v`
Expected: FAIL — `Chat` method not defined

- [ ] **Step 3: Implement Chat method**

Add to `internal/provider/lmstudio.go`:

```go
// Chat sends a prompt to LM Studio's OpenAI-compatible chat completions endpoint.
// It checks model availability first, auto-downloading if needed, and logs load state.
func (p *LMStudioProvider) Chat(model, prompt string) (LLMResult, error) {
	if model == "" {
		model = p.DefaultModel
	}

	exists, loaded, err := p.checkModels(model)
	if err != nil {
		return LLMResult{}, err
	}

	if !exists {
		fmt.Fprintf(os.Stderr, ">> lm-studio: model %s not found, downloading...\n", model)
		if err := p.downloadModel(model); err != nil {
			return LLMResult{}, err
		}
		// Re-check after download
		exists, loaded, err = p.checkModels(model)
		if err != nil {
			return LLMResult{}, err
		}
		if !exists {
			return LLMResult{}, fmt.Errorf("lm-studio: model %q not found after download", model)
		}
	}

	if !loaded {
		fmt.Fprintf(os.Stderr, ">> lm-studio: loading %s, expect delay\n", model)
	}

	start := time.Now()

	reqBody, _ := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
	})

	base := strings.TrimRight(p.BaseURL, "/")
	req, err := http.NewRequest(http.MethodPost, base+"/v1/chat/completions", strings.NewReader(string(reqBody)))
	if err != nil {
		return LLMResult{}, fmt.Errorf("lm-studio: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LLMResult{}, fmt.Errorf("lm-studio: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResult{}, fmt.Errorf("lm-studio: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return LLMResult{}, fmt.Errorf("lm-studio: %s\n%s", resp.Status, data)
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
		return LLMResult{}, fmt.Errorf("lm-studio: parse: %w", err)
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = strings.TrimSpace(chatResp.Choices[0].Message.Content)
	}

	return LLMResult{
		Provider:  "lm-studio",
		Model:     model,
		Response:  content,
		TokensIn:  chatResp.Usage.PromptTokens,
		TokensOut: chatResp.Usage.CompletionTokens,
		Latency:   time.Since(start),
		CostUSD:   0,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run "TestLMStudio_Chat" -v`
Expected: 4 PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/provider/lmstudio.go internal/provider/lmstudio_test.go
git commit -m "feat(provider): add LM Studio Chat with auto-download and load-state logging"
```

---

### Task 4: Wire into tiered runner and pricing

**Files:**
- Modify: `internal/provider/tiers.go:106-111`
- Modify: `internal/provider/tokens.go:27`

- [ ] **Step 1: Add lm-studio case to callProvider in tiers.go**

In `internal/provider/tiers.go`, after the `if name == "ollama"` block (line 111), add:

```go
	if name == "lm-studio" {
		p := &LMStudioProvider{
			BaseURL:      "http://localhost:1234",
			DefaultModel: "qwen3-8b",
		}
		if model == "" {
			model = p.DefaultModel
		}
		return p.Chat(model, prompt)
	}
```

- [ ] **Step 2: Add lm-studio to pricing table in tokens.go**

In `internal/provider/tokens.go`, add after the `"ollama"` entry (line 27):

```go
	"lm-studio": {0.00, 0.00},
```

- [ ] **Step 3: Run existing tests to verify nothing broke**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -v`
Expected: All PASS

- [ ] **Step 4: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/provider/tiers.go internal/provider/tokens.go
git commit -m "feat(provider): wire lm-studio into tiered runner and pricing table"
```

---

### Task 5: Wire into pipeline runner

**Files:**
- Modify: `internal/pipeline/runner.go:778-786`

- [ ] **Step 1: Add lm-studio case to the provider switch**

In `internal/pipeline/runner.go`, change the switch block at line 778 from:

```go
			switch prov {
			case "ollama", "":
				if model == "" {
					model = "qwen3:8b"
				}
				result, llmErr := provider.RunOllamaWithResult(model, rendered)
				if llmErr != nil {
					return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
				}
				out = result.Response
				stepTokensIn = result.TokensIn
				stepTokensOut = result.TokensOut
```

To:

```go
			switch prov {
			case "ollama", "":
				if model == "" {
					model = "qwen3:8b"
				}
				result, llmErr := provider.RunOllamaWithResult(model, rendered)
				if llmErr != nil {
					return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
				}
				out = result.Response
				stepTokensIn = result.TokensIn
				stepTokensOut = result.TokensOut
			case "lm-studio":
				lms := &provider.LMStudioProvider{
					BaseURL:      "http://localhost:1234",
					DefaultModel: "qwen3-8b",
				}
				result, llmErr := lms.Chat(model, rendered)
				if llmErr != nil {
					return nil, fmt.Errorf("step %s: %w", step.ID, llmErr)
				}
				out = result.Response
				stepTokensIn = result.TokensIn
				stepTokensOut = result.TokensOut
```

- [ ] **Step 2: Run pipeline tests to verify nothing broke**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -20`
Expected: All packages PASS

- [ ] **Step 4: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add internal/pipeline/runner.go
git commit -m "feat(provider): wire lm-studio into pipeline runner"
```

---

### Task 6: End-to-end smoke test

- [ ] **Step 1: Verify go build succeeds**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: Clean build, no errors

- [ ] **Step 2: Verify full test suite passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... -count=1`
Expected: All PASS

- [ ] **Step 3: Manual smoke test (if LM Studio is running)**

Run: `cd /Users/stokes/Projects/gl1tch && echo '(step :id "test" (llm :provider "lm-studio" :model "qwen3-8b" :prompt "Say hello in one word"))' > /tmp/lms-test.glitch && go run . run /tmp/lms-test.glitch`
Expected: Single-word response from LM Studio

- [ ] **Step 4: Final commit if any fixups were needed**

```bash
cd /Users/stokes/Projects/gl1tch
git add -A
git commit -m "fix: lm-studio provider fixups from smoke test"
```
