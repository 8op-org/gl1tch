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
