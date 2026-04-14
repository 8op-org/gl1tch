package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIChat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
