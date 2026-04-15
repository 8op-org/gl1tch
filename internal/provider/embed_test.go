package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbedOllama_ParsesResponse(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method

		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embedding": []float64{0.1, 0.2, 0.3},
		})
	}))
	defer srv.Close()

	vec, err := EmbedOllama(context.Background(), srv.URL, "nomic-embed-text", "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/embeddings" {
		t.Errorf("path = %q, want /api/embeddings", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["model"] != "nomic-embed-text" {
		t.Errorf("request model = %q, want nomic-embed-text", gotBody["model"])
	}
	if gotBody["prompt"] != "hello world" {
		t.Errorf("request prompt = %q, want hello world", gotBody["prompt"])
	}

	want := []float64{0.1, 0.2, 0.3}
	if len(vec) != len(want) {
		t.Fatalf("len(vec) = %d, want %d", len(vec), len(want))
	}
	for i, v := range want {
		if vec[i] != v {
			t.Errorf("vec[%d] = %f, want %f", i, vec[i], v)
		}
	}
}

func TestEmbedOllama_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("model not found"))
	}))
	defer srv.Close()

	_, err := EmbedOllama(context.Background(), srv.URL, "bad-model", "test")
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}
