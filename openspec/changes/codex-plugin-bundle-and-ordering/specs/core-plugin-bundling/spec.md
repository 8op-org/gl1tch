## ADDED Requirements

### Requirement: goreleaser builds core plugin binaries for all target platforms
The `.goreleaser.yml` in the `orcai` repo SHALL include a `builds` entry for each core plugin (`orcai-claude`, `orcai-github-copilot`, `orcai-codex`, `orcai-gemini`, `orcai-opencode`, `orcai-ollama`) using the `dir` field pointing to `../orcai-plugins/plugins/<name>`. Each plugin build SHALL use the same `goos`/`goarch` matrix as the `orcai` binary, excluding `windows/arm64`.

#### Scenario: Plugin binaries appear in release archives
- **WHEN** `goreleaser release` (or `goreleaser build`) completes
- **THEN** each target archive (e.g. `orcai_<ver>_linux_amd64.tar.gz`) contains `orcai-claude`, `orcai-github-copilot`, `orcai-codex`, `orcai-gemini`, `orcai-opencode`, and `orcai-ollama` alongside the `orcai` binary

#### Scenario: Windows archives omit arm64 plugin binaries
- **WHEN** the Windows archive is produced
- **THEN** only `amd64` plugin binaries are included; no `arm64` Windows binaries are present

### Requirement: Sidecar YAML files bundled in every release archive
The goreleaser `archives` configuration SHALL include `extra_files` entries (or a glob) that pulls the sidecar YAML for each core plugin (`claude.yaml`, `github-copilot.yaml`, `codex.yaml`, `gemini.yaml`, `opencode.yaml`, `ollama.yaml`) from `../orcai-plugins/plugins/*/` into the archive root or a `wrappers/` subdirectory.

#### Scenario: Sidecar YAMLs present in extracted archive
- **WHEN** a release archive is extracted
- **THEN** `codex.yaml`, `claude.yaml`, `github-copilot.yaml`, and other core sidecar files are present alongside the binaries

### Requirement: Release archive includes an install.sh script
Each release archive SHALL include an `install.sh` script that, when run from the extracted directory, copies all plugin binaries to `~/.local/bin` (falling back to `/usr/local/bin` if writable) and all sidecar YAML files to `~/.config/orcai/wrappers/`. The script SHALL be executable and print the destination of each installed file.

#### Scenario: install.sh installs binaries and sidecars
- **WHEN** `./install.sh` is run from the extracted archive directory
- **THEN** `orcai-codex`, `orcai-claude`, and other binaries are present in `~/.local/bin` and their sidecar YAMLs are present in `~/.config/orcai/wrappers/`

#### Scenario: install.sh is idempotent
- **WHEN** `./install.sh` is run twice
- **THEN** the second run overwrites with the same files and exits 0 without error
