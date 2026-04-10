# gl1tch Website One-Pager + Agent Skills

**Date:** 2026-04-10
**Status:** Approved

## Overview

Replace the legacy 8op.org site (gl1tch-legacy repo, Astro-based interactive terminal) with a clean one-pager in `site/index.html` inside the main `8op-org/gl1tch` repo. Zero build step, zero dependencies. Dark hacker aesthetic (Tokyo Night palette, monospace, neon accents). GitHub Pages deployment from `site/` directory, pointed at 8op.org.

## Architecture

Single `site/index.html` — all HTML, CSS, and JS inline. No framework, no build tool. A hex-rain canvas animation runs on the hero background (carried from legacy site). Dracula/Tokyo Night color scheme with teal and yellow accents.

## Page Structure (Story Arc)

### 1. Hero
- Large monospace "gl1tch" with glitch/distortion text effect
- Tagline: "your AI, your terminal, your rules"
- One-liner: "A composable CLI that chains shell commands and LLMs into workflows you own."
- Install: `brew install 8op-org/tap/glitch`
- GitHub CTA link
- Hex-rain canvas animation background

### 2. Ask Anything
- Terminal-style code block showing `glitch ask` routing examples
- Copy: "One command. gl1tch routes your question to the right workflow."
- Callouts: smart routing, local-first, zero config

### 3. Compose Workflows
- Split layout: YAML on left, explanation on right
- Copy: "Shell commands fetch the data. LLMs make sense of it. Chain them together in plain YAML."
- Real `github-pr-review` workflow example with shell + LLM steps
- Callouts: shell steps own data, LLM steps own reasoning, Go templates, workflows live in your repo

### 4. Extend with Plugins
- Copy: "Any binary on your PATH named glitch-* becomes a command."
- Terminal block showing `glitch plugin list` and usage
- Callouts: zero ceremony, any language, first-class args

### 5. Give Your Agents gl1tch
- Copy: "Install a skill and your AI coding agent can run gl1tch workflows directly in your project."
- Three sub-blocks:
  - **Claude Code:** `npx @anthropic-ai/superpower install gl1tch`
  - **Cursor / Copilot:** `.cursorrules` / agent config snippet
  - **Any agent:** Generic instruction block
- Note: "Skills let your agent compose and run workflows without you typing a thing."

### 6. Footer
- Install command, GitHub link, MIT license
- Attribution: "made with AI — guided by adam s†okes"

## Visual Design

- **Background:** Dark (#1a1b26), hex-rain canvas on hero
- **Typography:** JetBrains Mono (Google Fonts), monospace throughout
- **Colors:** Tokyo Night palette — teal (#7dcfff) for accents, yellow (#e0af68) for highlights, dim (#565f89) for secondary text, foreground (#c0caf5)
- **Code blocks:** Darker bg (#0f1017), teal border-left accent
- **Sections:** Generous vertical padding, subtle section dividers

## Deployment

- GitHub Pages from `site/` directory on main branch
- CNAME file for 8op.org domain
- No build step required

## Follow-up Tasks (Separate from this spec)

1. Clean up `8op-org/homebrew-tap` to only contain latest glitch formula
2. Update the gl1tch-plugin-release skill to match new tap structure
3. Create actual publishable skills for Claude Code, Cursor, and generic agents
