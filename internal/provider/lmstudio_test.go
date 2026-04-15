package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLMStudio_CheckModels_Found_Loaded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/models" {
			t.Errorf("path = %q, want /api/v0/models", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "qwen3-8b", "state": "loaded"},
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
		t.Error("exists = false, want true")
	}
	if !loaded {
		t.Error("loaded = false, want true")
	}
}

func TestLMStudio_CheckModels_Found_NotLoaded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "qwen3-8b", "state": "not-loaded"},
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
		t.Error("exists = false, want true")
	}
	if loaded {
		t.Error("loaded = true, want false")
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
	exists, loaded, err := p.checkModels("qwen3-8b")
	if err != nil {
		t.Fatalf("checkModels: %v", err)
	}
	if exists {
		t.Error("exists = true, want false")
	}
	if loaded {
		t.Error("loaded = true, want false")
	}
}

func TestLMStudio_DownloadModel_Success(t *testing.T) {
	var pollCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/models/download":
			var req struct{ Model string `json:"model"` }
			json.NewDecoder(r.Body).Decode(&req)
			if req.Model != "qwen3-8b" {
				t.Errorf("download model = %q, want qwen3-8b", req.Model)
			}
			json.NewEncoder(w).Encode(map[string]string{"job_id": "job-123"})

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/models/download/status/"):
			n := pollCount.Add(1)
			if n == 1 {
				json.NewEncoder(w).Encode(map[string]any{
					"status":   "downloading",
					"progress": 50.0,
				})
			} else {
				json.NewEncoder(w).Encode(map[string]any{
					"status":   "completed",
					"progress": 100.0,
				})
			}

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b", PollInterval: 1 * time.Millisecond}
	err := p.downloadModel("qwen3-8b")
	if err != nil {
		t.Fatalf("downloadModel: %v", err)
	}
}

func TestLMStudio_DownloadModel_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/models/download":
			json.NewEncoder(w).Encode(map[string]string{"job_id": "job-456"})

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/models/download/status/"):
			json.NewEncoder(w).Encode(map[string]any{
				"status": "failed",
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b", PollInterval: 1 * time.Millisecond}
	err := p.downloadModel("qwen3-8b")
	if err == nil {
		t.Fatal("expected error on failed download")
	}
	if !strings.Contains(err.Error(), "lm-studio") {
		t.Errorf("error should contain lm-studio: %v", err)
	}
}

func TestLMStudio_Chat_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v0/models":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "qwen3-8b", "state": "loaded"},
				},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions":
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
					"completion_tokens": 5,
				},
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
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
	if result.TokensOut != 5 {
		t.Errorf("tokens_out = %d, want 5", result.TokensOut)
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
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v0/models":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "qwen3-8b", "state": "loaded"},
				},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions":
			var req struct{ Model string `json:"model"` }
			json.NewDecoder(r.Body).Decode(&req)
			gotModel = req.Model
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "ok"}},
				},
				"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1},
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
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
	_, err := p.Chat("", "hello")
	if err == nil {
		t.Fatal("expected error when server is down")
	}
	if !strings.Contains(err.Error(), "lm-studio") {
		t.Errorf("error should contain lm-studio: %v", err)
	}
}

func TestLMStudio_Chat_ModelNotFound_TriggersDownload(t *testing.T) {
	var modelCheckCount atomic.Int32
	var downloadCalled atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v0/models":
			n := modelCheckCount.Add(1)
			if n == 1 {
				// First check: model not found
				json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{},
				})
			} else {
				// Second check after download: model available
				json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{"id": "qwen3-8b", "state": "loaded"},
					},
				})
			}

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/models/download":
			downloadCalled.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"job_id": "job-789"})

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/models/download/status/"):
			json.NewEncoder(w).Encode(map[string]any{
				"status":   "completed",
				"progress": 100.0,
			})

		case r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions":
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "downloaded and ready"}},
				},
				"usage": map[string]int{"prompt_tokens": 5, "completion_tokens": 3},
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	p := &LMStudioProvider{BaseURL: srv.URL, DefaultModel: "qwen3-8b", PollInterval: 1 * time.Millisecond}
	result, err := p.Chat("qwen3-8b", "hello")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if downloadCalled.Load() == 0 {
		t.Error("expected downloadModel to be called")
	}
	if result.Response != "downloaded and ready" {
		t.Errorf("response = %q, want 'downloaded and ready'", result.Response)
	}
}
