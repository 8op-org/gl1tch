## 1. CLI & Binary

- [x] 1.1 Update `cmd/root.go`: change `Use` to `glitch`, update `Short`/`Long` descriptions to reference GLITCH
- [x] 1.2 Update `Makefile`: rename build output target from `orcai` to `glitch`
- [x] 1.3 Update `Taskfile.yml`: rename binary references from `orcai` to `glitch`
- [x] 1.4 Update `install.sh`: change binary name from `orcai` to `glitch`
- [x] 1.5 Rename binary entrypoint symlink/alias in `bin/` if present

## 2. System Prompts

- [x] 2.1 Update `internal/systemprompts/defaults/pipeline-generator.md`: replace "orcai" with "glitch" in product references
- [x] 2.2 Update `internal/systemprompts/defaults/clarify.md`: replace `ORCAI_CLARIFY` protocol name with `GLITCH_CLARIFY`
- [x] 2.3 Update `internal/systemprompts/defaults/brain-write.md`: replace `~/.orcai/` path reference with `~/.glitch/`
- [x] 2.4 Update `internal/systemprompts/loader.go`: change config path from `~/.config/orcai/prompts/` to `~/.config/glitch/prompts/` and update warning messages

## 3. Site — Core Components

- [x] 3.1 Update `site/src/components/Footer.astro`: replace "ORCAI ABS" with "GLITCH" and update link
- [x] 3.2 Update `site/src/components/screens/AboutScreen.astro`: replace `// orcai — abbs //` header and all orcai references with GLITCH; use "your GLITCH" framing
- [x] 3.3 Update `site/src/layouts/Base.astro`: update page title, meta description, canonical URL to powerglove.dev / GLITCH
- [x] 3.4 Update `site/src/layouts/TerminalShell.astro`: replace any ORCAI/ABBS references
- [x] 3.5 Update `site/src/components/Nav.astro`: update product name/logo text
- [x] 3.6 Update `site/src/components/AnsiLogo.astro`: update product name if present
- [x] 3.7 Update `site/src/components/HelpOverlay.astro`: update product name references
- [x] 3.8 Update `site/src/components/SysinfoBox.astro`: update product name references

## 4. Site — Screen Content

- [x] 4.1 Update `site/src/components/screens/HomeScreen.astro`: replace all orcai/ABBS references with GLITCH; apply "your GLITCH" framing
- [x] 4.2 Update `site/src/components/screens/GettingStartedScreen.astro`: replace all `orcai` binary references with `glitch`; update config paths from `~/.config/orcai/` to `~/.config/glitch/`
- [x] 4.3 Update `site/src/components/screens/DocsScreen.astro`: replace `orcai pipeline run` with `glitch pipeline run` in all code examples
- [x] 4.4 Update `site/src/components/screens/ThemesScreen.astro`: replace "ship with orcai" with "ship with GLITCH"
- [x] 4.5 Update `site/src/components/screens/LabsScreen.astro`: replace all `orcai pipeline run` commands with `glitch pipeline run`
- [x] 4.6 Update `site/src/components/screens/PluginsScreen.astro`: replace orcai references

## 5. README & Docs

- [x] 5.1 Update `README.md`: replace all ORCAI/ABBS/orcai references with GLITCH/glitch; update domain to powerglove.dev; use "your GLITCH" framing in intro
- [x] 5.2 Update any files in `docs/` that reference ORCAI/ABBS

## 6. Config & Metadata

- [x] 6.1 Update `apm.yml`: change package name from `orcai` to `glitch`
- [x] 6.2 Update `dist/metadata.json`: update product name field
- [x] 6.3 Update `dist/config.yaml`: update any product name references
- [x] 6.4 Update `dist/artifacts.json`: update product name references

## 7. Verify

- [x] 7.1 Run `grep -ri "orcai\|ABBS\|agentic bulletin" site/ cmd/ README.md internal/systemprompts/ apm.yml dist/` — confirm zero user-facing hits remain
- [x] 7.2 Run `go build ./...` — confirm build succeeds with updated paths
- [x] 7.3 Run `./glitch --help` — confirm binary name and help text show GLITCH
