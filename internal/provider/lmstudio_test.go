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
