## Why

ORCAI/ABBS served as the internal working name, but GLITCH is the product — the AI hero that users actually connect with. The rebrand establishes GLITCH as the focal point, with powerglove.dev as the public home, while preserving the BBS hacker aesthetic that defines the experience.

GLITCH belongs to everyone. He's your hero, not ours — each user's GLITCH is their own version of the AI they want in their terminal. The name and identity should reflect that openness.

## What Changes

- All user-facing product references change from "ORCAI" / "ORCAI ABBS" / "Agentic Bulletin Board System" → **GLITCH**
- Domain and URL references change from orcai.* → **powerglove.dev**
- Site copy, README, CLI help text, and UI labels updated to center GLITCH
- System prompts that introduce the assistant as "ORCAI" updated to GLITCH
- Go binary name (`orcai`) migrated to `glitch`
- Go module path (`orcai`) updated in go.mod and imports
- The BBS aesthetic, ANSI art, Dracula theme, and zer0c00l voice are **preserved entirely** — they are the world GLITCH lives in, not the brand being replaced
- Internal CSS class names (`.bbs-*`) and Go type names can migrate gradually — no big-bang rename required

## Capabilities

### New Capabilities

- `glitch-brand-identity`: GLITCH as the public product name — site copy, README, CLI, and UI all reference GLITCH; powerglove.dev as the canonical domain

### Modified Capabilities

- `agent-context-panel`: Any ORCAI references in assistant/agent-facing prompts or UI updated to GLITCH

## Impact

- `site/` — all Astro components, page copy, meta tags, titles
- `README.md` — primary user-facing doc
- `cmd/` — CLI root command name, help text, version output
- `go.mod` — module path (`github.com/*/orcai` → `github.com/*/glitch`)
- `internal/systemprompts/` — system prompts that name the product
- `main.go` — binary entrypoint
- `Makefile` / `Taskfile.yml` — build targets referencing binary name
- `install.sh` — installation script
- `apm.yml` — package manifest name
- `dist/` — build metadata
