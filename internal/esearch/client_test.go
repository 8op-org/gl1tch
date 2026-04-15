package esearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() unexpected error: %v", err)
	}
}

func TestPingUnreachable(t *testing.T) {
	// Port 1 is typically unreachable / refused.
	c := NewClient("http://127.0.0.1:1")
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping() expected error for unreachable host, got nil")
	}
}

func TestSearch(t *testing.T) {
	indices := []string{"glitch-events", "glitch-summaries"}
	wantPath := "/" + strings.Join(indices, ",") + "/_search"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			http.Error(w, fmt.Sprintf("unexpected path %s", r.URL.Path), http.StatusBadRequest)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"hits": {
				"total": { "value": 2 },
				"hits": [
					{ "_index": "glitch-events",    "_source": {"type":"commit"} },
					{ "_index": "glitch-summaries",  "_source": {"scope":"daily"} }
				]
			}
		}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	query := json.RawMessage(`{"query":{"match_all":{}}}`)
	got, err := c.Search(context.Background(), indices, query)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if got.Total != 2 {
		t.Errorf("Total = %d, want 2", got.Total)
	}
	if len(got.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(got.Results))
	}
	if got.Results[0].Index != "glitch-events" {
		t.Errorf("Results[0].Index = %q, want %q", got.Results[0].Index, "glitch-events")
	}
}

func TestEnsureIndex(t *testing.T) {
	putCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			// Index does not exist.
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPut:
			putCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"acknowledged":true,"shards_acknowledged":true,"index":"test-index"}`)
		default:
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.EnsureIndex(context.Background(), "test-index", EventsMapping); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}
	if !putCalled {
		t.Error("EnsureIndex() did not issue a PUT request to create the index")
	}
}

func TestEnsureIndexAlreadyExists(t *testing.T) {
	putCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			// Index exists.
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			putCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.EnsureIndex(context.Background(), "test-index", EventsMapping); err != nil {
		t.Fatalf("EnsureIndex() error: %v", err)
	}
	if putCalled {
		t.Error("EnsureIndex() issued PUT when index already exists")
	}
}

func TestParseIndexStats(t *testing.T) {
	raw := `[
		{"index":"glitch-events","docs.count":"150","store.size":"1mb"},
		{"index":"glitch-llm-calls","docs.count":"42","store.size":"512kb"},
		{"index":".kibana_1","docs.count":"10","store.size":"100kb"}
	]`

	stats := parseIndexStats([]byte(raw))

	if len(stats) != 2 {
		t.Fatalf("expected 2 glitch indices, got %d", len(stats))
	}

	found := false
	for _, s := range stats {
		if s.Index == "glitch-events" && s.DocCount == "150" {
			found = true
		}
	}
	if !found {
		t.Error("expected glitch-events with 150 docs")
	}
}

func TestIndexDoc_BuildsCorrectRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/test-index/_doc/doc1" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]string
		json.Unmarshal(body, &got)
		if got["title"] != "hello" {
			t.Errorf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"_id":"doc1","_version":1,"result":"created"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.IndexDoc(context.Background(), "test-index", "doc1", json.RawMessage(`{"title":"hello"}`))
	if err != nil {
		t.Fatalf("IndexDoc() error: %v", err)
	}
	if resp.ID != "doc1" {
		t.Errorf("ID = %q, want %q", resp.ID, "doc1")
	}
	if resp.Version != 1 {
		t.Errorf("Version = %d, want 1", resp.Version)
	}
	if resp.Result != "created" {
		t.Errorf("Result = %q, want %q", resp.Result, "created")
	}
}

func TestIndexDoc_AutoID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/test-index/_doc" {
			t.Errorf("unexpected path %s, want /test-index/_doc", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"_id":"auto-generated-id","_version":1,"result":"created"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.IndexDoc(context.Background(), "test-index", "", json.RawMessage(`{"title":"hello"}`))
	if err != nil {
		t.Fatalf("IndexDoc() error: %v", err)
	}
	if resp.ID != "auto-generated-id" {
		t.Errorf("ID = %q, want %q", resp.ID, "auto-generated-id")
	}
}

func TestDeleteByQuery_BuildsCorrectRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/test-index/_delete_by_query" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "match_all") {
			t.Errorf("unexpected query body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"deleted":5}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	query := json.RawMessage(`{"query":{"match_all":{}}}`)
	resp, err := c.DeleteByQuery(context.Background(), "test-index", query)
	if err != nil {
		t.Fatalf("DeleteByQuery() error: %v", err)
	}
	if resp.Deleted != 5 {
		t.Errorf("Deleted = %d, want 5", resp.Deleted)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello…"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}
