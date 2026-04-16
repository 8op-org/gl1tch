package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/8op-org/gl1tch/internal/workspace"
)

func TestLoadConfigSexprRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := filepath.Join(dir, ".config", "glitch")
	_ = os.MkdirAll(cfgDir, 0o755)

	src := `(config
  :default-model "qwen2.5:7b"
  :default-provider "ollama"
  :eval-threshold 5
  :workflows-dir "/tmp/flows"
  (providers
    (provider "openrouter"
      :type "openai-compatible"
      :base-url "https://openrouter.ai/api/v1"
      :api-key-env "OPENROUTER_API_KEY"
      :default-model "gemma3:free")
    (provider "claude" :type "cli"))
  (tiers
    (tier :providers ("ollama") :model "qwen2.5:7b")
    (tier :providers ("openrouter") :model "gemma3:free")))
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.glitch"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfigFrom(filepath.Join(cfgDir, "config.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultModel != "qwen2.5:7b" || cfg.DefaultProvider != "ollama" || cfg.EvalThreshold != 5 || cfg.WorkflowsDir != "/tmp/flows" {
		t.Fatalf("top-level mismatch: %+v", cfg)
	}
	if len(cfg.Providers) != 2 {
		t.Fatalf("want 2 providers, got %d", len(cfg.Providers))
	}
	if cfg.Providers["openrouter"].Type != "openai-compatible" || cfg.Providers["openrouter"].BaseURL != "https://openrouter.ai/api/v1" || cfg.Providers["openrouter"].APIKeyEnv != "OPENROUTER_API_KEY" || cfg.Providers["openrouter"].DefaultModel != "gemma3:free" {
		t.Errorf("openrouter mismatch: %+v", cfg.Providers["openrouter"])
	}
	if cfg.Providers["claude"].Type != "cli" {
		t.Errorf("claude mismatch: %+v", cfg.Providers["claude"])
	}
	if len(cfg.Tiers) != 2 {
		t.Fatalf("want 2 tiers, got %d", len(cfg.Tiers))
	}
	if len(cfg.Tiers[0].Providers) != 1 || cfg.Tiers[0].Providers[0] != "ollama" || cfg.Tiers[0].Model != "qwen2.5:7b" {
		t.Errorf("tier 0 mismatch: %+v", cfg.Tiers[0])
	}
}

func TestLoadConfigMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := loadConfigFrom("/nonexistent/path/config.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultModel == "" || cfg.DefaultProvider == "" {
		t.Fatalf("expected defaults populated, got %+v", cfg)
	}
	if cfg.EvalThreshold == 0 {
		t.Fatal("expected default eval threshold")
	}
	if len(cfg.Tiers) == 0 {
		t.Fatal("expected default tiers")
	}
}

func TestSaveConfigWritesSexpr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.glitch")
	cfg := &Config{DefaultModel: "x", DefaultProvider: "y", EvalThreshold: 3, WorkflowsDir: "/tmp"}
	if err := saveConfigAt(cfg, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || data[0] != '(' {
		t.Fatalf("expected s-expr, got: %s", data)
	}
	cfg2, err := loadConfigFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.DefaultModel != "x" || cfg2.DefaultProvider != "y" || cfg2.EvalThreshold != 3 {
		t.Fatalf("round-trip mismatch: %+v", cfg2)
	}
}

func TestProviderConfig_ResolveAPIKey_EnvFirst(t *testing.T) {
	t.Setenv("TEST_KEY_ENV", "from-env")
	pc := ProviderConfig{
		APIKeyEnv: "TEST_KEY_ENV",
		APIKey:    "from-config",
	}
	key, err := pc.ResolveAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if key != "from-env" {
		t.Errorf("key = %q, want from-env", key)
	}
}

func TestProviderConfig_ResolveAPIKey_FallbackToConfig(t *testing.T) {
	pc := ProviderConfig{
		APIKeyEnv: "NONEXISTENT_VAR_12345",
		APIKey:    "from-config",
	}
	key, err := pc.ResolveAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if key != "from-config" {
		t.Errorf("key = %q, want from-config", key)
	}
}

func TestProviderConfig_ResolveAPIKey_NeitherSet(t *testing.T) {
	pc := ProviderConfig{}
	_, err := pc.ResolveAPIKey()
	if err == nil {
		t.Fatal("expected error when no API key configured")
	}
}

func TestApplyWorkspace(t *testing.T) {
	cfg := &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
	}

	ws := &workspace.Workspace{
		Defaults: workspace.Defaults{
			Model:    "llama3.2:3b",
			Provider: "lm-studio",
		},
	}

	ApplyWorkspace(ws, cfg)

	if cfg.DefaultModel != "llama3.2:3b" {
		t.Fatalf("DefaultModel: got %q, want llama3.2:3b", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "lm-studio" {
		t.Fatalf("DefaultProvider: got %q, want lm-studio", cfg.DefaultProvider)
	}
}

func TestApplyWorkspace_NilWorkspace(t *testing.T) {
	cfg := &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
	}

	ApplyWorkspace(nil, cfg)

	if cfg.DefaultModel != "qwen3:8b" {
		t.Fatalf("DefaultModel changed to %q", cfg.DefaultModel)
	}
}

func TestApplyWorkspace_PartialDefaults(t *testing.T) {
	cfg := &Config{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
	}

	ws := &workspace.Workspace{
		Defaults: workspace.Defaults{
			Model: "llama3.2:3b",
		},
	}

	ApplyWorkspace(ws, cfg)

	if cfg.DefaultModel != "llama3.2:3b" {
		t.Fatalf("DefaultModel: got %q", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "ollama" {
		t.Fatalf("DefaultProvider should stay ollama, got %q", cfg.DefaultProvider)
	}
}
