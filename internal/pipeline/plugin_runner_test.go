// internal/pipeline/plugin_runner_test.go
package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecutePluginCall_Basic creates a greeter plugin with manifest defs and a
// say subcommand that echoes the greeting + name arg.
func TestExecutePluginCall_Basic(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "greeter")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Manifest with a shared def
	manifest := []byte(`(plugin "greeter" :version "0.1.0")
(def greeting "hello")
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.glitch"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	// Subcommand: say.glitch
	sub := []byte(`(arg "name" :default "world")

(workflow "say"
  (step "greet"
    (run "echo {{.param.greeting}} {{.param.name}}")))
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "say.glitch"), sub, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunPluginSubcommand(root, "greeter", "say", map[string]string{"name": "alice"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hello alice") {
		t.Fatalf("expected output containing %q, got %q", "hello alice", out)
	}
}

// TestExecutePluginCall_MissingPlugin calls with a nonexistent plugin and expects an error.
func TestExecutePluginCall_MissingPlugin(t *testing.T) {
	root := t.TempDir()
	_, err := RunPluginSubcommand(root, "nonexistent", "run", nil)
	if err == nil {
		t.Fatal("expected error for missing plugin, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' in error, got: %v", err)
	}
}

// TestExecutePluginCall_MissingSubcommand creates a plugin dir but calls a
// subcommand that does not exist.
func TestExecutePluginCall_MissingSubcommand(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "myplugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a dummy subcommand so the plugin dir is valid
	dummy := []byte(`(workflow "hello"
  (step "s1"
    (run "echo hi")))
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "hello.glitch"), dummy, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := RunPluginSubcommand(root, "myplugin", "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for missing subcommand, got nil")
	}
	if !strings.Contains(err.Error(), "no subcommand") {
		t.Fatalf("expected 'no subcommand' in error, got: %v", err)
	}
}

// TestExecutePluginCall_RequiredArgMissing creates a plugin with a required arg
// and calls without providing it.
func TestExecutePluginCall_RequiredArgMissing(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "deployer")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Subcommand with a required arg (no :default)
	sub := []byte(`(arg "repo")

(workflow "deploy"
  (step "go"
    (run "echo deploying {{.param.repo}}")))
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "deploy.glitch"), sub, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := RunPluginSubcommand(root, "deployer", "deploy", nil)
	if err == nil {
		t.Fatal("expected error for missing required arg, got nil")
	}
	if !strings.Contains(err.Error(), "requires argument") {
		t.Fatalf("expected 'requires argument' in error, got: %v", err)
	}
}
