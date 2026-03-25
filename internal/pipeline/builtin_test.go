package pipeline

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuiltins(t *testing.T) {
	ctx := context.Background()

	t.Run("assert success with true", func(t *testing.T) {
		out, err := builtinAssert(ctx, map[string]any{"condition": "true"}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out["passed"] != true {
			t.Errorf("expected passed=true, got %v", out["passed"])
		}
	})

	t.Run("assert failure with false", func(t *testing.T) {
		_, err := builtinAssert(ctx, map[string]any{"condition": "false"}, nil)
		if err == nil {
			t.Error("expected error for false condition")
		}
	})

	t.Run("assert failure with custom message", func(t *testing.T) {
		_, err := builtinAssert(ctx, map[string]any{
			"condition": "false",
			"message":   "custom error",
		}, nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "custom error") {
			t.Errorf("expected 'custom error' in message, got: %v", err)
		}
	})

	t.Run("assert contains success", func(t *testing.T) {
		out, err := builtinAssert(ctx, map[string]any{
			"condition": "contains:hello",
			"value":     "hello world",
		}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out["passed"] != true {
			t.Errorf("expected passed=true")
		}
	})

	t.Run("assert contains failure", func(t *testing.T) {
		_, err := builtinAssert(ctx, map[string]any{
			"condition": "contains:missing",
			"value":     "hello world",
		}, nil)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("assert len gt success", func(t *testing.T) {
		out, err := builtinAssert(ctx, map[string]any{
			"condition": "len > 3",
			"value":     "hello",
		}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out["passed"] != true {
			t.Errorf("expected passed=true")
		}
	})

	t.Run("assert matches success", func(t *testing.T) {
		out, err := builtinAssert(ctx, map[string]any{
			"condition": "matches:^hell",
			"value":     "hello",
		}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out["passed"] != true {
			t.Errorf("expected passed=true")
		}
	})

	t.Run("set_data merges data", func(t *testing.T) {
		out, err := builtinSetData(ctx, map[string]any{
			"data": map[string]any{"key": "value", "num": 42},
		}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out["key"] != "value" {
			t.Errorf("expected key=value, got %v", out["key"])
		}
		if out["num"] != 42 {
			t.Errorf("expected num=42, got %v", out["num"])
		}
	})

	t.Run("set_data empty returns empty map", func(t *testing.T) {
		out, err := builtinSetData(ctx, map[string]any{}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty map, got %v", out)
		}
	})

	t.Run("set_data wrong type returns error", func(t *testing.T) {
		_, err := builtinSetData(ctx, map[string]any{"data": "not a map"}, nil)
		if err == nil {
			t.Error("expected error for non-map data arg")
		}
	})

	t.Run("log writes message to writer", func(t *testing.T) {
		var buf bytes.Buffer
		_, err := builtinLog(ctx, map[string]any{"message": "test log message"}, &buf)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !strings.Contains(buf.String(), "test log message") {
			t.Errorf("expected message in output, got %q", buf.String())
		}
	})

	t.Run("log nil writer no panic", func(t *testing.T) {
		out, err := builtinLog(ctx, map[string]any{"message": "hello"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != nil {
			t.Errorf("expected nil output")
		}
	})

	t.Run("sleep completes within timeout", func(t *testing.T) {
		start := time.Now()
		_, err := builtinSleep(ctx, map[string]any{"duration": "10ms"}, nil)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if elapsed < 10*time.Millisecond {
			t.Errorf("sleep too short: %v", elapsed)
		}
	})

	t.Run("sleep respects context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
		defer cancel()
		_, err := builtinSleep(cancelCtx, map[string]any{"duration": "1s"}, nil)
		if err == nil {
			t.Error("expected context cancellation error")
		}
	})

	t.Run("sleep invalid duration returns error", func(t *testing.T) {
		_, err := builtinSleep(ctx, map[string]any{"duration": "not-a-duration"}, nil)
		if err == nil {
			t.Error("expected error for invalid duration")
		}
	})

	t.Run("http_get success 2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok response"))
		}))
		defer srv.Close()

		out, err := builtinHTTPGet(ctx, map[string]any{"url": srv.URL}, nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out["status"] != 200 {
			t.Errorf("expected status 200, got %v", out["status"])
		}
		if out["body"] != "ok response" {
			t.Errorf("expected body 'ok response', got %v", out["body"])
		}
	})

	t.Run("http_get non-2xx returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		}))
		defer srv.Close()

		_, err := builtinHTTPGet(ctx, map[string]any{"url": srv.URL}, nil)
		if err == nil {
			t.Error("expected error for 404 response")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected '404' in error, got: %v", err)
		}
	})

	t.Run("http_get missing url returns error", func(t *testing.T) {
		_, err := builtinHTTPGet(ctx, map[string]any{}, nil)
		if err == nil {
			t.Error("expected error for missing url")
		}
	})

	t.Run("http_get invalid url returns error", func(t *testing.T) {
		_, err := builtinHTTPGet(ctx, map[string]any{"url": "://invalid"}, nil)
		if err == nil {
			t.Error("expected error for invalid url")
		}
	})
}

func TestExpandForEach(t *testing.T) {
	t.Run("json array", func(t *testing.T) {
		items := expandForEach(`["a","b","c"]`)
		if len(items) != 3 || items[0] != "a" || items[1] != "b" || items[2] != "c" {
			t.Errorf("unexpected items: %v", items)
		}
	})

	t.Run("newline separated", func(t *testing.T) {
		items := expandForEach("foo\nbar\nbaz")
		if len(items) != 3 || items[0] != "foo" || items[2] != "baz" {
			t.Errorf("unexpected items: %v", items)
		}
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		items := expandForEach("")
		if len(items) != 0 {
			t.Errorf("expected empty, got %v", items)
		}
	})

	t.Run("skips empty lines", func(t *testing.T) {
		items := expandForEach("a\n\nb\n\nc")
		if len(items) != 3 {
			t.Errorf("expected 3 items, got %v", items)
		}
	})
}
