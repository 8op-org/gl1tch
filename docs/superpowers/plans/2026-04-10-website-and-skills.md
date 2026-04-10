# gl1tch Website, Agent Skills, and Brew Tap Cleanup — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a one-pager website at `site/index.html`, create installable agent skills for Claude Code / Cursor / generic agents, and clean up the homebrew tap to only serve current gl1tch.

**Architecture:** Single HTML file with inline CSS/JS for the site. Skills are markdown files published to the superpowers plugin registry. Brew tap cleanup removes legacy formulas and updates the install skill.

**Tech Stack:** HTML/CSS/JS (no build), GitHub Pages, superpowers skill format, Homebrew tap

---

### Task 1: Create site directory and index.html skeleton

**Files:**
- Create: `site/index.html`
- Create: `site/CNAME`

- [ ] **Step 1: Create site directory**

```bash
mkdir -p site
```

- [ ] **Step 2: Create CNAME for 8op.org**

```
8op.org
```

Write this to `site/CNAME`.

- [ ] **Step 3: Create index.html with full page structure**

Create `site/index.html` with the complete one-pager. This is a single file containing all HTML, CSS, and JS inline. Structure:

```
<!DOCTYPE html>
<html lang="en">
<head>
  - charset, viewport, title "gl1tch — your AI, your terminal, your rules"
  - Google Fonts: JetBrains Mono
  - All CSS inline in <style>
</head>
<body>
  <section id="hero">
    - "gl1tch" large monospace with glitch text animation (CSS keyframes)
    - Tagline: "your AI, your terminal, your rules"
    - Description: "A composable CLI that chains shell commands and LLMs into workflows you own."
    - Install: brew install 8op-org/tap/glitch (copyable code block)
    - GitHub CTA button linking to https://github.com/8op-org/gl1tch
    - <canvas id="hexbg"> for hex-rain animation behind content
  </section>

  <section id="ask">
    - Header: "Ask anything"
    - Copy: "One command. gl1tch routes your question to the right workflow — PR URLs go to review, natural language gets matched by your local LLM."
    - Terminal-style code block:
      $ glitch ask "review PR https://github.com/org/repo/pull/42"
      → routes to github-pr-review workflow
      $ glitch ask "what issues are open"
      → routes to github-issues workflow
      $ glitch ask "CI status"
      → routes to github-actions workflow
    - 3-column grid: Smart routing / Local-first / Zero config
  </section>

  <section id="workflows">
    - Header: "Compose workflows"
    - Copy: "Shell commands fetch the data. LLMs make sense of it. Chain them together in plain YAML."
    - Split layout: YAML left, callouts right
    - YAML example (use real syntax from the repo):
      name: github-pr-review
      steps:
        - id: fetch-pr
          run: gh pr view "{{.input}}" --json title,body,reviews
        - id: fetch-diff
          run: gh pr diff "{{.input}}"
        - id: review
          llm:
            provider: claude
            model: claude-haiku-4-5-20251001
            prompt: |
              PR: {{step "fetch-pr"}}
              Diff: {{step "fetch-diff"}}
              Review as a senior engineer. Flag bugs and security issues.
    - Right callouts: Shell steps own data / LLM steps own reasoning / Go templates / Your workflows, your repo
  </section>

  <section id="plugins">
    - Header: "Extend with plugins"
    - Copy: "Any binary on your PATH named glitch-* becomes a command. No registry, no config, no approval process."
    - Terminal block:
      $ glitch plugin list
        glitch-summarize
        glitch-deploy
      $ glitch summarize README.md
    - 3-column grid: Zero ceremony / Any language / First-class args
  </section>

  <section id="skills">
    - Header: "Give your agents gl1tch"
    - Copy: "Install a skill and your AI coding agent can run gl1tch workflows directly in your project."
    - 3 sub-blocks with headers and code blocks:
      Claude Code: npx @anthropic-ai/superpower install gl1tch
      Cursor / Copilot: .cursorrules snippet
      Any agent: generic instruction block
    - Note: "Skills let your agent compose and run workflows without you typing a thing. Author once, delegate forever."
  </section>

  <footer>
    - Install command / GitHub link / MIT
    - "made with AI — guided by adam s†okes"
  </footer>

  <script>
    - Hex-rain canvas animation (port from legacy site)
    - Copy-to-clipboard on install code blocks
  </script>
</body>
</html>
```

Color palette (CSS custom properties):
```css
:root {
  --bg: #1a1b26;
  --bg-dark: #0f1017;
  --fg: #c0caf5;
  --dim: #565f89;
  --teal: #7dcfff;
  --yellow: #e0af68;
  --pink: #f7768e;
  --border: #3d7a8a;
}
```

The glitch text effect uses CSS `@keyframes` with `clip-path` and `text-shadow` shifts — no JS needed.

Code blocks use `--bg-dark` background with a `border-left: 3px solid var(--teal)`.

Sections have `padding: 80px 0` with `max-width: 900px` centered.

The hex-rain canvas is positioned `fixed` behind the hero with low opacity (`0.07`), same implementation as legacy site's `initHexCanvas()`.

- [ ] **Step 4: Test locally**

```bash
cd site && python3 -m http.server 8080
# Open http://localhost:8080 in browser
# Verify: hero renders, hex animation runs, all sections visible, code blocks styled
```

- [ ] **Step 5: Commit**

```bash
git add site/index.html site/CNAME
git commit -m "feat(site): add one-pager website for 8op.org"
```

---

### Task 2: Create agent skills for Claude Code

**Files:**
- Create: `skills/claude-code/SKILL.md`

- [ ] **Step 1: Create the Claude Code skill**

Create `skills/claude-code/SKILL.md` — this is the skill users install via `npx @anthropic-ai/superpower install gl1tch`. It teaches Claude Code how to use glitch:

```markdown
---
name: gl1tch
description: Use the gl1tch CLI to automate GitHub tasks, run workflows, and compose shell+LLM pipelines. Invoke when the user asks to review PRs, triage issues, check CI, or run any glitch workflow.
---

# gl1tch — GitHub Automation CLI

## When to use

- User asks to review a PR, check issues, or see CI status
- User asks to run a glitch workflow
- User asks to create or edit a workflow YAML
- User mentions glitch, gl1tch, or references .glitch/workflows/

## Commands

| Command | What it does |
|---------|-------------|
| `glitch ask "<question or URL>"` | Routes to the best matching workflow |
| `glitch workflow list` | List available workflows |
| `glitch workflow run <name> [input]` | Run a specific workflow by name |
| `glitch config show` | Show current configuration |
| `glitch plugin list` | List installed plugins |

## Examples

```bash
# Review a PR
glitch ask "review PR https://github.com/org/repo/pull/42"

# List open issues
glitch ask "what issues are open"

# Check CI status
glitch ask "CI status"

# Run a specific workflow
glitch workflow run github-prs

# Run against a different repo
glitch ask -C ~/Projects/other-repo "show me open PRs"
```

## Workflow authoring

Workflows live in `.glitch/workflows/*.yaml`. Structure:

```yaml
name: my-workflow
description: What this workflow does
steps:
  - id: step-name
    run: shell command here    # shell step
  - id: llm-step
    llm:                       # LLM step
      provider: ollama         # or "claude"
      model: qwen2.5:7b
      prompt: |
        Use {{step "step-name"}} to reference prior step output.
        Use {{.input}} for the user's input.
```

Key rules:
- Shell steps fetch data (`gh`, `git`, `curl`, etc.)
- LLM steps reason about data — never use LLM for API calls
- Templates use Go text/template syntax: `{{.param.input}}` or `{{step "id"}}`
- Provider options: "ollama" (local, default) or "claude" (Anthropic API)

## Installation check

Before using glitch, verify it's installed:
```bash
glitch --version
```

If not installed: `brew install 8op-org/tap/glitch`
```

- [ ] **Step 2: Commit**

```bash
git add skills/claude-code/SKILL.md
git commit -m "feat(skills): add Claude Code agent skill"
```

---

### Task 3: Create agent config for Cursor / Copilot

**Files:**
- Create: `skills/cursor/.cursorrules`
- Create: `skills/generic/AGENTS.md`

- [ ] **Step 1: Create Cursor rules file**

Create `skills/cursor/.cursorrules`:

```markdown
# gl1tch — GitHub Automation CLI

This project uses gl1tch for GitHub automation workflows.

## Available commands

- `glitch ask "<question or URL>"` — routes to the best workflow automatically
- `glitch workflow list` — list available workflows
- `glitch workflow run <name> [input]` — run a named workflow
- `glitch plugin list` — list installed plugins

## When to use

When the user asks to review a PR, triage issues, check CI status, summarize
git activity, or automate any GitHub task, use the glitch CLI instead of
running gh commands manually.

## Examples

```bash
glitch ask "review PR https://github.com/org/repo/pull/42"
glitch ask "what issues are open"
glitch ask "CI status"
glitch workflow run github-prs
```

## Workflow authoring

Workflows live in `.glitch/workflows/*.yaml`:

```yaml
name: my-workflow
description: What this does
steps:
  - id: data
    run: gh pr list --json number,title
  - id: summary
    llm:
      provider: ollama
      prompt: |
        Summarize: {{step "data"}}
```

Shell steps own data fetching. LLM steps own reasoning. Never use LLM for API calls.
```

- [ ] **Step 2: Create generic AGENTS.md**

Create `skills/generic/AGENTS.md`:

```markdown
# gl1tch — GitHub Automation CLI

This project uses gl1tch for GitHub automation. Use these commands instead of
running raw gh/git commands when the user asks about PRs, issues, CI, or
git activity.

## Commands

- `glitch ask "<question or URL>"` — auto-routes to the best workflow
- `glitch workflow list` — list workflows
- `glitch workflow run <name> [input]` — run a workflow
- `glitch plugin list` — list plugins

## Workflow files

Location: `.glitch/workflows/*.yaml`

```yaml
name: example
steps:
  - id: fetch
    run: gh issue list --json number,title
  - id: analyze
    llm:
      provider: ollama
      prompt: "Categorize: {{step \"fetch\"}}"
```

Shell steps fetch data. LLM steps reason. Never call APIs from LLM steps.
```

- [ ] **Step 3: Commit**

```bash
git add skills/cursor/.cursorrules skills/generic/AGENTS.md
git commit -m "feat(skills): add Cursor and generic agent configs"
```

---

### Task 4: Update website skills section with real install commands

**Files:**
- Modify: `site/index.html` (skills section)

- [ ] **Step 1: Update the skills section**

After creating the actual skill files, update the skills section of `index.html` to reference the real file paths and install commands:

- Claude Code: `npx @anthropic-ai/superpower install gl1tch` (or link to `skills/claude-code/`)
- Cursor: "Copy `skills/cursor/.cursorrules` to your project root"
- Generic: "Copy `skills/generic/AGENTS.md` to your project root"
- Link to the GitHub repo skills directory for all three

- [ ] **Step 2: Commit**

```bash
git add site/index.html
git commit -m "feat(site): update skills section with real install paths"
```

---

### Task 5: Clean up homebrew tap

**Files:**
- Delete (in `8op-org/homebrew-tap`): `Formula/gl1tch-mud.rb`, `Formula/gl1tch-weather.rb`, `Formula/glitch-gamification.rb`, `Formula/glitch-notify.rb`
- Delete: `gl1tch-mud.rb` (root level, old v0.4.0 formula)
- Delete: `glitch.rb` (root level, old v0.4.0 formula)
- Keep: `Formula/glitch.rb` (current v0.5.0 — will be updated by next release)

- [ ] **Step 1: Clone the tap repo**

```bash
cd /tmp && gh repo clone 8op-org/homebrew-tap
cd /tmp/homebrew-tap
```

- [ ] **Step 2: Remove legacy formulas**

Delete all formulas except `Formula/glitch.rb`:

```bash
# Root-level legacy formulas
rm -f gl1tch-mud.rb glitch.rb

# Legacy plugin formulas
rm -f Formula/gl1tch-mud.rb Formula/gl1tch-weather.rb Formula/glitch-gamification.rb Formula/glitch-notify.rb
```

This leaves only `Formula/glitch.rb` (the current gl1tch v2 formula).

- [ ] **Step 3: Update README**

Update `README.md` to reflect the cleaned-up tap:

```markdown
# homebrew-tap

Homebrew tap for [gl1tch](https://github.com/8op-org/gl1tch).

## Install

```bash
brew install 8op-org/tap/glitch
```

## Formulas

| Formula | Description |
|---------|------------|
| `glitch` | gl1tch — your GitHub co-pilot |
```

- [ ] **Step 4: Commit and push**

```bash
git add -A
git commit -m "chore: remove legacy formulas, keep only glitch"
git push origin main
```

- [ ] **Step 5: Verify install still works**

```bash
brew update
brew reinstall glitch
glitch --version
```

---

### Task 6: Update gl1tch-install skill

**Files:**
- Modify: `~/.claude/skills/gl1tch-install/SKILL.md`

- [ ] **Step 1: Strip the install skill down to glitch only**

Update `~/.claude/skills/gl1tch-install/SKILL.md` — remove all legacy plugin references. The packages table should only have:

| Formula | Source repo |
|---------|------------|
| `glitch` | `8op-org/gl1tch` |

The execution section becomes:

```bash
# 1. Remove locally built shadows
rm -f ~/.local/bin/glitch
rm -f "$(go env GOPATH)/bin/glitch"

# 2. Tap
brew tap 8op-org/tap 2>/dev/null || true

# 3. Reinstall
brew reinstall glitch

# 4. Confirm
brew list --versions glitch
glitch --version
```

Remove all references to `gl1tch-notify`, `gl1tch-mud`, `gl1tch-weather`, `glitch-mattermost`, `glitch plugin install`, sidecar YAMLs, and wrappers.

Remove the notes about `gl1tch-notify` being darwin-only, `gl1tch-mud` web UI, etc.

- [ ] **Step 2: Commit (if skill is in a repo)**

The skill lives at `~/.claude/skills/gl1tch-install/SKILL.md` — this is a local file, not in the gl1tch repo. Just save it.

---

### Task 7: Update gl1tch-plugin-release skill

**Files:**
- Modify: `~/.claude/skills/gl1tch-plugin-release/SKILL.md`

- [ ] **Step 1: Update the release skill**

Key changes needed:
- Update the "Known Plugins" table — remove all legacy plugins (notify, mud, weather, mattermost)
- Remove BUSD integration section (legacy)
- Remove daemon/sidecar rules (legacy)
- Remove `glitch plugin install` step (v2 doesn't have this command — plugins are just binaries on PATH)
- Remove `glitch-plugin.yaml` manifest (legacy)
- Remove wrappers and sidecar references
- Update repo name convention: repos are under `8op-org` not `adam-stokes`
- Update module path convention: `github.com/8op-org/gl1tch-<name>`
- Update Go version to match current go.mod (1.26.1)
- Simplify the Makefile — no sidecar copy, just binary install
- Remove step 7 (plugin registration), step 8 (ecosystem page), step 9 (install skill update)
- Keep: GoReleaser config, GitHub Actions release workflow, HOMEBREW_TAP_GITHUB_TOKEN setup

- [ ] **Step 2: Save the updated skill**

The skill lives at `~/.claude/skills/gl1tch-plugin-release/SKILL.md` — local file.

---

### Task 8: Configure GitHub Pages deployment

- [ ] **Step 1: Enable GitHub Pages on 8op-org/gl1tch**

```bash
# Enable Pages from site/ directory on main branch
gh api repos/8op-org/gl1tch/pages -X POST -f source.branch=main -f source.path=/site 2>/dev/null || \
gh api repos/8op-org/gl1tch/pages -X PUT -f source.branch=main -f source.path=/site

# Set custom domain
gh api repos/8op-org/gl1tch/pages -X PUT -f cname=8op.org
```

- [ ] **Step 2: Update DNS (manual step)**

If 8op.org currently points to gl1tch-legacy's GitHub Pages, it needs to point to the gl1tch repo instead. This may require updating the DNS CNAME or the GitHub Pages custom domain config on the legacy repo (remove the custom domain there first).

```bash
# Remove custom domain from legacy repo
gh api repos/8op-org/gl1tch-legacy/pages -X PUT -f cname=""
```

- [ ] **Step 3: Verify deployment**

```bash
# Check Pages status
gh api repos/8op-org/gl1tch/pages --jq '.status'

# Once deployed, verify
curl -sI https://8op.org | head -5
```

- [ ] **Step 4: Commit any remaining changes**

```bash
git add -A
git commit -m "chore: configure GitHub Pages deployment"
git push origin main
```
