// internal/plugin/manifest_test.go
package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func writePluginFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "plugin.glitch"), []byte(content), 0o644); err != nil {
		t.Fatalf("write plugin.glitch: %v", err)
	}
}

func TestLoadManifest_Full(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, `
(plugin "myplugin"
  :description "A test plugin"
  :version "1.2.3")

(def base "https://example.com")
(def endpoint "https://example.com/api")
`)

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m.Name != "myplugin" {
		t.Errorf("Name = %q, want %q", m.Name, "myplugin")
	}
	if m.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", m.Description, "A test plugin")
	}
	if m.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", m.Version, "1.2.3")
	}
	if m.Defs["base"] != "https://example.com" {
		t.Errorf("Defs[base] = %q, want %q", m.Defs["base"], "https://example.com")
	}
	if m.Defs["endpoint"] != "https://example.com/api" {
		t.Errorf("Defs[endpoint] = %q, want %q", m.Defs["endpoint"], "https://example.com/api")
	}
}

func TestLoadManifest_NoManifest(t *testing.T) {
	dir := t.TempDir()
	// No plugin.glitch file written — dir name is the basename of the temp dir.

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m.Name != filepath.Base(dir) {
		t.Errorf("Name = %q, want dir basename %q", m.Name, filepath.Base(dir))
	}
	if m.Description != "" {
		t.Errorf("Description = %q, want empty", m.Description)
	}
	if m.Version != "" {
		t.Errorf("Version = %q, want empty", m.Version)
	}
	if len(m.Defs) != 0 {
		t.Errorf("Defs = %v, want empty map", m.Defs)
	}
}

func TestLoadManifest_DefsOnly(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, `
(def host "localhost")
(def port "8080")
(def addr host)
`)

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	// No (plugin ...) form — name falls back to dir basename.
	if m.Name != filepath.Base(dir) {
		t.Errorf("Name = %q, want dir basename %q", m.Name, filepath.Base(dir))
	}
	if m.Defs["host"] != "localhost" {
		t.Errorf("Defs[host] = %q, want %q", m.Defs["host"], "localhost")
	}
	if m.Defs["port"] != "8080" {
		t.Errorf("Defs[port] = %q, want %q", m.Defs["port"], "8080")
	}
	// Symbol resolution: addr should resolve to the value of host.
	if m.Defs["addr"] != "localhost" {
		t.Errorf("Defs[addr] = %q, want %q (symbol resolved)", m.Defs["addr"], "localhost")
	}
}
