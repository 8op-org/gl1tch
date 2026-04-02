package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDecisionNodeEvaluate_success(t *testing.T) {
	// Build a fake Ollama server that returns a valid branch response.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/generate" {
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		inner, _ := json.Marshal(map[string]string{"branch": "yes"})
		resp := map[string]string{"response": string(inner)}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	d := &DecisionNode{
		Model:     "llama3",
		Prompt:    "should we proceed?",
		OllamaURL: srv.URL,
	}
	wctx := NewWorkflowContext()
	branch, err := d.Evaluate(context.Background(), wctx)
	if err != nil {
		t.Fatalf("Evaluate: unexpected error: %v", err)
	}
	if branch != "yes" {
		t.Errorf("branch = %q, want %q", branch, "yes")
	}
}

func TestDecisionNodeEvaluate_timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow Ollama — sleep longer than the decision timeout.
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &DecisionNode{
		Model:       "llama3",
		Prompt:      "choose",
		OllamaURL:   srv.URL,
		TimeoutSecs: 0, // triggers default; override for speed
	}
	// Use a context with a very short deadline to force timeout quickly.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wctx := NewWorkflowContext()
	_, err := d.Evaluate(ctx, wctx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestDecisionNodeEvaluate_missingBranchField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner, _ := json.Marshal(map[string]string{"result": "ok"}) // no "branch" key
		resp := map[string]string{"response": string(inner)}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	d := &DecisionNode{Model: "llama3", Prompt: "p", OllamaURL: srv.URL}
	wctx := NewWorkflowContext()
	_, err := d.Evaluate(context.Background(), wctx)
	if err == nil {
		t.Fatal("expected error for missing branch field, got nil")
	}
}

func TestDecisionNodeEvaluate_nonStringBranch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner, _ := json.Marshal(map[string]any{"branch": 42}) // wrong type
		resp := map[string]string{"response": string(inner)}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	d := &DecisionNode{Model: "llama3", Prompt: "p", OllamaURL: srv.URL}
	wctx := NewWorkflowContext()
	_, err := d.Evaluate(context.Background(), wctx)
	if err == nil {
		t.Fatal("expected error for non-string branch, got nil")
	}
}

func TestDecisionNodeEvaluate_HTTP503(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	d := &DecisionNode{Model: "llama3", Prompt: "p", OllamaURL: srv.URL}
	wctx := NewWorkflowContext()
	_, err := d.Evaluate(context.Background(), wctx)
	if err == nil {
		t.Fatal("expected error for HTTP 503, got nil")
	}
}
