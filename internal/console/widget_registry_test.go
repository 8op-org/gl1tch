package console

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWidgetSidecar(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeWidgetSidecar: %v", err)
	}
}

func TestLoadWidgetRegistry_Empty(t *testing.T) {
	reg := LoadWidgetRegistry(t.TempDir())
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if reg.FindByTrigger("/anything") != nil {
		t.Error("expected nil from empty registry")
	}
}

func TestLoadWidgetRegistry_NonexistentDir(t *testing.T) {
	reg := LoadWidgetRegistry("/nonexistent/dir/for/test")
	if reg == nil {
		t.Fatal("expected non-nil registry even for missing dir")
	}
}

func TestLoadWidgetRegistry_SkipsNoModeBlock(t *testing.T) {
	dir := t.TempDir()
	writeWidgetSidecar(t, dir, "plain.yaml", `
name: plain
command: echo
kind: tool
`)
	reg := LoadWidgetRegistry(dir)
	if len(reg.widgets) != 0 {
		t.Errorf("expected 0 widgets, got %d", len(reg.widgets))
	}
}

func TestLoadWidgetRegistry_LoadsModeBlock(t *testing.T) {
	dir := t.TempDir()
	writeWidgetSidecar(t, dir, "widget.yaml", `
name: my-widget
command: my-binary
kind: tool
mode:
  trigger: /widget
  logo: WIDGET
  speaker: WDGT
  exit_command: quit
`)
	reg := LoadWidgetRegistry(dir)
	cfg := reg.FindByTrigger("/widget")
	if cfg == nil {
		t.Fatal("expected to find /widget trigger")
	}
	if cfg.Schema.Mode.Logo != "WIDGET" {
		t.Errorf("expected logo 'WIDGET', got %q", cfg.Schema.Mode.Logo)
	}
}

func TestLoadWidgetRegistry_DuplicateTriggerFirstWins(t *testing.T) {
	dir := t.TempDir()
	writeWidgetSidecar(t, dir, "a.yaml", `
name: first
command: first-bin
mode:
  trigger: /dup
  logo: FIRST
  speaker: FST
  exit_command: quit
`)
	writeWidgetSidecar(t, dir, "b.yaml", `
name: second
command: second-bin
mode:
  trigger: /dup
  logo: SECOND
  speaker: SND
  exit_command: quit
`)
	reg := LoadWidgetRegistry(dir)
	if len(reg.widgets) != 1 {
		t.Fatalf("expected 1 widget (first wins), got %d", len(reg.widgets))
	}
	cfg := reg.FindByTrigger("/dup")
	if cfg == nil {
		t.Fatal("expected /dup to be found")
	}
	if cfg.Schema.Mode.Logo != "FIRST" {
		t.Errorf("expected first loaded logo 'FIRST', got %q", cfg.Schema.Mode.Logo)
	}
}

func TestLoadWidgetRegistry_SkipsMissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	writeWidgetSidecar(t, dir, "bad.yaml", `
name: bad
command: bin
mode:
  trigger: /bad
  logo: BAD
  # speaker and exit_command missing
`)
	reg := LoadWidgetRegistry(dir)
	if len(reg.widgets) != 0 {
		t.Errorf("expected 0 widgets due to missing required fields, got %d", len(reg.widgets))
	}
}

func TestWidgetRegistry_FindByTrigger_Unknown(t *testing.T) {
	dir := t.TempDir()
	writeWidgetSidecar(t, dir, "w.yaml", `
name: w
command: bin
mode:
  trigger: /known
  logo: L
  speaker: SPK
  exit_command: quit
`)
	reg := LoadWidgetRegistry(dir)
	if reg.FindByTrigger("/unknown") != nil {
		t.Error("expected nil for unknown trigger")
	}
}

func TestWidgetRegistry_AllSignalTopics_Deduplicated(t *testing.T) {
	dir := t.TempDir()
	writeWidgetSidecar(t, dir, "a.yaml", `
name: a
command: bin-a
mode:
  trigger: /a
  logo: A
  speaker: A
  exit_command: quit
signals:
  - topic: mud.*
    handler: companion
  - topic: shared.*
    handler: log
`)
	writeWidgetSidecar(t, dir, "b.yaml", `
name: b
command: bin-b
mode:
  trigger: /b
  logo: B
  speaker: B
  exit_command: quit
signals:
  - topic: shared.*
    handler: companion
`)
	reg := LoadWidgetRegistry(dir)
	topics := reg.AllSignalTopics()
	// shared.* appears in both but should be deduplicated
	seen := map[string]int{}
	for _, t2 := range topics {
		seen[t2]++
	}
	if seen["mud.*"] != 1 {
		t.Errorf("expected mud.* once, got %d", seen["mud.*"])
	}
	if seen["shared.*"] != 1 {
		t.Errorf("expected shared.* once (deduplicated), got %d", seen["shared.*"])
	}
}
