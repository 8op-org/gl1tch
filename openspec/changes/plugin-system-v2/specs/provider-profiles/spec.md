## ADDED Requirements

### Requirement: Bundled capability profiles ship embedded in binary
Orcai SHALL embed capability profiles for popular AI providers (claude, gemini, opencode, aider, goose, copilot) as YAML files via `//go:embed`. These profiles SHALL be loaded at startup without reading the filesystem.

#### Scenario: Bundled profile loaded for installed binary
- **WHEN** orcai starts and `claude` is found in PATH
- **THEN** the claude capability profile is active and available for session launching and model picking

#### Scenario: Bundled profile is dormant when binary not installed
- **WHEN** orcai starts and `gemini` is not found in PATH
- **THEN** the gemini profile is loaded but not presented as an available provider

### Requirement: User profiles override bundled profiles by name
Orcai SHALL scan `~/.config/orcai/providers/*.yaml` at startup. When a user profile shares a name with a bundled profile, the user profile SHALL take precedence for all fields.

#### Scenario: User overrides bundled model list
- **WHEN** `~/.config/orcai/providers/claude.yaml` exists with a custom `models` list
- **THEN** orcai uses the user-supplied model list instead of the bundled one

#### Scenario: User installs a community provider profile
- **WHEN** `~/.config/orcai/providers/aichat.yaml` is placed with a valid profile
- **THEN** orcai discovers it as an available provider if the binary is in PATH

### Requirement: Profile YAML schema covers deep integration fields
A valid provider profile YAML SHALL support: `name`, `binary`, `display_name`, `api_key_env`, `models` (list with `id`, `display`, `cost_input_per_1m`, `cost_output_per_1m`), and `session` (with `window_name` template, `launch_args`, `env` map).

#### Scenario: Model picker populates from profile
- **WHEN** a user initiates a new session for a provider
- **THEN** the model picker shows all models listed in that provider's profile

#### Scenario: Cost tracking uses profile rates
- **WHEN** a session ends and token counts are known
- **THEN** orcai computes session cost using `cost_input_per_1m` and `cost_output_per_1m` from the active model entry

### Requirement: Profile loading replaces hardcoded discovery lists
The `knownCLITools` slice in `internal/discovery` and the `adapterDefs` slice in `internal/bridge` SHALL be removed. Binary detection SHALL be driven by profile `binary` fields.

#### Scenario: No hardcoded provider names remain in discovery
- **WHEN** a new provider profile is added without code changes
- **THEN** that provider is discovered and available without recompiling orcai
