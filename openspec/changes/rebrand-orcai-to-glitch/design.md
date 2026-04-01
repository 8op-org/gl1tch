## Context

The codebase currently uses "orcai" as both the Go module path (`github.com/adam-stokes/orcai`), the binary name (`orcai`), and the product name in all user-facing surfaces. GLITCH is the new public product identity. The BBS hacker aesthetic is preserved — this is a name migration, not a personality change.

Key constraint: the Go module path is referenced across all import paths in the codebase. A full module rename (`orcai` → `glitch`) is the highest-risk change and should be done last, after all user-facing surfaces are updated.

## Goals / Non-Goals

**Goals:**
- All user-facing text (site, README, CLI help, system prompts, UI labels) references GLITCH
- Binary name changes from `orcai` to `glitch`
- Go module path updated from `github.com/adam-stokes/orcai` → `github.com/adam-stokes/glitch`
- powerglove.dev reflected in site meta, README, install script
- GLITCH is framed as a personal hero — each user's own — not a corporate product

**Non-Goals:**
- Changing the BBS aesthetic, ANSI art, Dracula theme, or zer0c00l voice
- Renaming internal CSS classes (`.bbs-*`) immediately — these can migrate over time
- Changing the git repo name or GitHub org structure (separate concern)
- Updating `openspec/specs/` capability names (internal, not user-facing)

## Decisions

**Decision: User-facing surfaces first, module path last**
Rationale: Updating Go import paths touches every `.go` file. Doing it first creates a huge diff that obscures the actual UX changes. Ship the visible rebrand, then do the mechanical module rename as a follow-up commit. This also allows the site/README/CLI changes to be reviewed independently.

**Decision: Binary rename `orcai` → `glitch` in same pass as user-facing changes**
Rationale: The binary name IS user-facing — it's what users type. It belongs with the site/README/CLI pass, not the module rename pass.

**Decision: Keep `github.com/adam-stokes/orcai` module path for now, rename in follow-up**
Alternative considered: rename everything in one pass using `go mod edit` + `find/sed`. Risk: touches 59k+ lines, high merge conflict surface. The split approach keeps each commit reviewable.

**Decision: GLITCH identity is open, not prescriptive**
The product copy should frame GLITCH as belonging to the user — "your GLITCH" not "our GLITCH". System prompts should reinforce this. Each user's GLITCH is whatever they need it to be.

## Risks / Trade-offs

- **Install script breakage** → Mitigation: update `install.sh` binary name in same pass; test install path
- **Makefile / Taskfile build targets** → Mitigation: audit all `orcai` references in build files in the same pass
- **Module path rename is mechanical but large** → Mitigation: use `go mod edit -module` + `find . -name "*.go" | xargs sed` in a single isolated commit; run tests after
- **dist/ metadata files** → Mitigation: these are generated artifacts; update templates/generators, not dist output directly

## Migration Plan

**Pass 1 — User-facing surfaces (this change):**
1. Site copy: all Astro components, page titles, meta tags → GLITCH / powerglove.dev
2. README.md → GLITCH branding
3. CLI: `cmd/root.go` Use/Short/Long → `glitch` / GLITCH description
4. Binary name in Makefile, Taskfile, install.sh
5. System prompts in `internal/systemprompts/` → GLITCH
6. `apm.yml` package name
7. `dist/metadata.json` product name

**Pass 2 — Module path rename (follow-up, not in this change):**
1. `go mod edit -module github.com/adam-stokes/glitch`
2. `find . -name "*.go" | xargs sed -i 's|github.com/adam-stokes/orcai|github.com/adam-stokes/glitch|g'`
3. Run `go build ./...` and `go test ./...`

## Open Questions

- Does the GitHub repo get renamed from `orcai` → `glitch`? (affects module path and clone URLs)
- Does `nomoresecrets.dev` redirect to `powerglove.dev`? (DNS/redirect config, not in scope here)
