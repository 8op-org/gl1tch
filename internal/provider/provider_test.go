package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProviders_ReadsYAMLFiles(t *testing.T) {
	dir := t.TempDir()

	// Write two provider YAML files.
	err := os.WriteFile(filepath.Join(dir, "claude.yaml"), []byte("name: claude\ncommand: claude -p \"{{.prompt}}\"\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(dir, "gpt.yaml"), []byte("name: gpt\ncommand: gpt --query \"{{.prompt}}\"\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	reg, err := LoadProviders(dir)
	if err != nil {
		t.Fatalf("LoadProviders: %v", err)
	}

	if len(reg.providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(reg.providers))
	}
	if reg.providers["claude"] == nil {
		t.Fatal("missing provider: claude")
	}
	if reg.providers["gpt"] == nil {
		t.Fatal("missing provider: gpt")
	}
	if reg.providers["claude"].Command != `claude -p "{{.prompt}}"` {
		t.Fatalf("unexpected command: %s", reg.providers["claude"].Command)
	}
}

func TestLoadProviders_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	reg, err := LoadProviders(dir)
	if err != nil {
		t.Fatalf("LoadProviders: %v", err)
	}
	if len(reg.providers) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(reg.providers))
	}
}

func TestLoadProviders_MissingDir(t *testing.T) {
	reg, err := LoadProviders("/tmp/does-not-exist-glitch-test-dir")
	if err != nil {
		t.Fatalf("expected no error for missing dir, got: %v", err)
	}
	if len(reg.providers) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(reg.providers))
	}
}

func TestRenderProviderCommand(t *testing.T) {
	reg := &ProviderRegistry{
		providers: map[string]*Provider{
			"claude": {
				Name:    "claude",
				Command: `claude -p "{{.prompt}}"`,
			},
		},
	}

	got, err := reg.RenderCommand("claude", "hello world")
	if err != nil {
		t.Fatalf("RenderCommand: %v", err)
	}
	want := `claude -p "hello world"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderProviderCommand_NotFound(t *testing.T) {
	reg := &ProviderRegistry{
		providers: map[string]*Provider{
			"claude": {Name: "claude", Command: "claude"},
		},
	}

	_, err := reg.RenderCommand("gpt", "hello")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
	if !strings.Contains(err.Error(), "gpt") {
		t.Fatalf("error should mention requested provider: %v", err)
	}
	if !strings.Contains(err.Error(), "claude") {
		t.Fatalf("error should list available providers: %v", err)
	}
}

func TestRunProvider_ExecsCommand(t *testing.T) {
	reg := &ProviderRegistry{
		providers: map[string]*Provider{
			"echo": {
				Name:    "echo",
				Command: `echo "{{.prompt}}"`,
			},
		},
	}

	got, err := reg.RunProvider("echo", "hello world")
	if err != nil {
		t.Fatalf("RunProvider: %v", err)
	}
	if got != "hello world" {
		t.Fatalf("got %q, want %q", got, "hello world")
	}
}
