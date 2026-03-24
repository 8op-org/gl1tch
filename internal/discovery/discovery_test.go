package discovery_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adam-stokes/orcai/internal/discovery"
)

func TestScanNative_Empty(t *testing.T) {
	dir := t.TempDir()
	plugins, err := discovery.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	for _, p := range plugins {
		if p.Type == discovery.TypeNative {
			t.Errorf("expected no native plugins in empty dir, got %+v", p)
		}
	}
}

func TestScanNative_FindsExecutable(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(pluginsDir, "orcai-test-plugin")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho hi"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	plugins, err := discovery.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	found := false
	for _, p := range plugins {
		if p.Name == "orcai-test-plugin" && p.Type == discovery.TypeNative {
			found = true
		}
	}
	if !found {
		t.Errorf("expected to find orcai-test-plugin, got %+v", plugins)
	}
}

func TestScanNative_SkipsNonExecutable(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(pluginsDir, "not-executable")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	plugins, err := discovery.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	for _, p := range plugins {
		if p.Name == "not-executable" {
			t.Errorf("should not have loaded non-executable file")
		}
	}
}

func TestNativePriorityOverCLI(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Create a native plugin named "claude" — it should shadow the CLI wrapper
	path := filepath.Join(pluginsDir, "claude")
	if err := os.WriteFile(path, []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	plugins, err := discovery.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	count := 0
	for _, p := range plugins {
		if p.Name == "claude" {
			count++
			if p.Type != discovery.TypeNative {
				t.Errorf("expected claude to be TypeNative, got %v", p.Type)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 claude plugin, got %d", count)
	}
}

func TestScanPipelines_FindsYAML(t *testing.T) {
	dir := t.TempDir()
	pipelinesDir := filepath.Join(dir, "pipelines")
	if err := os.MkdirAll(pipelinesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	yamlPath := filepath.Join(pipelinesDir, "my-pipeline.pipeline.yaml")
	content := "name: my-pipeline\nversion: \"1.0\"\nsteps: []\n"
	if err := os.WriteFile(yamlPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	plugins, err := discovery.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	found := false
	for _, p := range plugins {
		if p.Name == "my-pipeline" && p.Type == discovery.TypePipeline {
			found = true
		}
	}
	if !found {
		t.Error("expected to discover 'my-pipeline' as TypePipeline")
	}
}

func TestScanPipelines_IgnoresNonYAML(t *testing.T) {
	dir := t.TempDir()
	pipelinesDir := filepath.Join(dir, "pipelines")
	if err := os.MkdirAll(pipelinesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// A file that does NOT end in .pipeline.yaml
	if err := os.WriteFile(filepath.Join(pipelinesDir, "not-a-pipeline.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	plugins, err := discovery.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	for _, p := range plugins {
		if p.Type == discovery.TypePipeline {
			t.Errorf("expected no pipeline plugins, got %+v", p)
		}
	}
}
