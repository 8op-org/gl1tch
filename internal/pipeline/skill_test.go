package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkill_DirectPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my-skill.md")
	os.WriteFile(path, []byte("# Review checklist\n- Check errors"), 0o644)

	content, err := loadSkill(path)
	if err != nil {
		t.Fatalf("loadSkill(%q) error: %v", path, err)
	}
	if content != "# Review checklist\n- Check errors" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestLoadSkill_ByName(t *testing.T) {
	// Create a fake .claude/skills/test-skill/SKILL.md in a temp dir
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".claude", "skills", "test-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill content here"), 0o644)

	// loadSkill looks relative to cwd, so we chdir
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	content, err := loadSkill("test-skill")
	if err != nil {
		t.Fatalf("loadSkill(test-skill) error: %v", err)
	}
	if content != "skill content here" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestLoadSkill_NotFound(t *testing.T) {
	_, err := loadSkill("nonexistent-skill-xyz")
	if err == nil {
		t.Fatal("expected error for missing skill")
	}
}
