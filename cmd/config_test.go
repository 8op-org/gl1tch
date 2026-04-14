package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_WithProviders(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte(`
default_model: qwen3:8b
default_provider: ollama
providers:
  openrouter:
    type: openai-compatible
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_API_KEY
    api_key: sk-or-fallback
    default_model: meta-llama/llama-4-scout:free
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfigFrom(cfgPath)
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("providers count = %d, want 1", len(cfg.Providers))
	}
	p := cfg.Providers["openrouter"]
	if p.Type != "openai-compatible" {
		t.Errorf("type = %q, want openai-compatible", p.Type)
	}
	if p.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("base_url = %q", p.BaseURL)
	}
	if p.APIKeyEnv != "OPENROUTER_API_KEY" {
		t.Errorf("api_key_env = %q", p.APIKeyEnv)
	}
	if p.APIKey != "sk-or-fallback" {
		t.Errorf("api_key = %q", p.APIKey)
	}
	if p.DefaultModel != "meta-llama/llama-4-scout:free" {
		t.Errorf("default_model = %q", p.DefaultModel)
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
