package plugin_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/adam-stokes/orcai/internal/plugin"
)

func TestCliAdapter_Name(t *testing.T) {
	a := plugin.NewCliAdapter("echo-tool", "A simple echo tool", "echo")
	if a.Name() != "echo-tool" {
		t.Errorf("expected 'echo-tool', got %q", a.Name())
	}
}

func TestCliAdapter_Execute(t *testing.T) {
	// cat reads stdin and echoes it to stdout — available on macOS/Linux.
	a := plugin.NewCliAdapter("cat-tool", "echoes input", "cat")
	var buf bytes.Buffer
	err := a.Execute(context.Background(), "hello world\n", nil, &buf)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("expected output to contain 'hello world', got %q", buf.String())
	}
}

func TestCliAdapter_Execute_ContextCancel(t *testing.T) {
	// sleep 10 will be cancelled by ctx before it finishes.
	a := plugin.NewCliAdapter("sleep-tool", "sleeps", "sleep", "10")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	var buf bytes.Buffer
	err := a.Execute(ctx, "", nil, &buf)
	if err == nil {
		t.Error("expected an error when context is cancelled")
	}
}
