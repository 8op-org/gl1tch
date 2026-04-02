package console

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignalHandlerRegistry_UnknownHandler_Drops(t *testing.T) {
	reg := BuildSignalHandlerRegistry(nil, nil)
	// Should not panic; unknown handler is silently dropped.
	reg.Dispatch("nonexistent", "some.topic", `{"key":"val"}`)
}

func TestSignalHandlerRegistry_LogHandler_WritesLine(t *testing.T) {
	// Override HOME to a temp dir so the log file lands there.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	reg := BuildSignalHandlerRegistry(nil, nil)
	reg.Dispatch("log", "test.topic", `{"hello":"world"}`)

	logPath := filepath.Join(tmpHome, ".local", "share", "glitch", "plugin-signals.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected log file at %s, got error: %v", logPath, err)
	}
	content := string(data)
	if !strings.Contains(content, "test.topic") {
		t.Errorf("expected log line to contain 'test.topic', got: %q", content)
	}
	if !strings.Contains(content, `{"hello":"world"}`) {
		t.Errorf("expected log line to contain payload, got: %q", content)
	}
}

func TestSignalHandlerRegistry_LogHandler_AppendsMultiple(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	reg := BuildSignalHandlerRegistry(nil, nil)
	reg.Dispatch("log", "topic.one", "payload-one")
	reg.Dispatch("log", "topic.two", "payload-two")

	logPath := filepath.Join(tmpHome, ".local", "share", "glitch", "plugin-signals.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected log file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "topic.one") || !strings.Contains(content, "topic.two") {
		t.Errorf("expected both log lines, got: %q", content)
	}
}
