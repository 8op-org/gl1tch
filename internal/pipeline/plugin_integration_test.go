// internal/pipeline/plugin_integration_test.go
package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPluginIntegration_FullPath sets up a complete "echo-tool" plugin with a manifest,
// a say subcommand, and a count subcommand, then verifies all three execute correctly.
func TestPluginIntegration_FullPath(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "echo-tool")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// plugin.glitch — manifest with shared def
	manifest := []byte(`(plugin "echo-tool" :description "Echo utility" :version "1.0.0")
(def prefix "[echo]")
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.glitch"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	// say.glitch — uses manifest def via param injection
	say := []byte(`(arg "message" :default "hello")
(arg "loud" :type :flag)

(workflow "say"
  :description "Echo a message"
  (step "out"
    (run "echo {{.param.prefix}} {{.param.message}}")))
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "say.glitch"), say, 0o644); err != nil {
		t.Fatal(err)
	}

	// count.glitch — uses lines SDK form
	count := []byte(`(arg "items")

(workflow "count"
  :description "Count items"
  (step "list"
    (run "echo {{.param.items}}"))
  (step "result"
    (lines "list")))
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "count.glitch"), count, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("basic subcommand with manifest defs", func(t *testing.T) {
		out, err := RunPluginSubcommand(root, "echo-tool", "say", map[string]string{"message": "world"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "[echo] world") {
			t.Fatalf("expected output containing %q, got %q", "[echo] world", out)
		}
	})

	t.Run("default arg value", func(t *testing.T) {
		out, err := RunPluginSubcommand(root, "echo-tool", "say", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "[echo] hello") {
			t.Fatalf("expected output containing %q, got %q", "[echo] hello", out)
		}
	})

	t.Run("SDK form lines in plugin", func(t *testing.T) {
		out, err := RunPluginSubcommand(root, "echo-tool", "count", map[string]string{"items": "a b c"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "[") {
			t.Fatalf("expected JSON array output containing %q, got %q", "[", out)
		}
	})
}

// TestPluginIntegration_WorkflowInvokesPlugin sets up a "greet" plugin with a hello
// subcommand and verifies basic name arg substitution works end-to-end.
func TestPluginIntegration_WorkflowInvokesPlugin(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "greet")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	hello := []byte(`(arg "name" :default "world")
(workflow "hello"
  (step "out"
    (run "echo hello {{.param.name}}")))
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "hello.glitch"), hello, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunPluginSubcommand(root, "greet", "hello", map[string]string{"name": "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hello alice") {
		t.Fatalf("expected output containing %q, got %q", "hello alice", out)
	}
}
