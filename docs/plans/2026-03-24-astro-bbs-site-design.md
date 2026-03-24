# Astro BBS Site Migration — Design

**Date:** 2026-03-24
**Status:** Approved

## Summary

Migrate the orcai GitHub Pages site from hand-written HTML to Astro with Content Collections. Fixes nav copy-paste fragility, sysinfo box alignment, box-drawing character rendering, and adds changelog, ANSI gallery, pipeline reference, and plugin registry pages.

## Architecture

New Astro project lives at `site/` in the repo root (separate from current `docs/`). GitHub Actions deploys from `site/dist` to the `gh-pages` branch.

```
site/
  src/
    layouts/
      Base.astro               — html/head/body, loads bbs.css + bbs.js
      BBS.astro                — Base + Nav + Footer (all pages use this)
    components/
      Nav.astro                — fixed top bar, activePage prop, · separators
      Footer.astro             — [ ORCAI BBS ] [ MIT ] etc.
      SysinfoBox.astro         — index login panel, CSS-bordered table
      FeatureCard.astro        — feature grid cards
      AnsiLogo.astro           — color-cycled block logo
    pages/
      index.astro
      getting-started.astro
      plugins.astro
      pipelines.astro
      themes.astro             — ANSI gallery
      changelog.astro          — renders changelog collection
      registry/
        index.astro            — plugin registry list
        [slug].astro           — individual plugin page
    content/
      config.ts                — Content Collections schema
      changelog/               — one .md per release
      registry/                — one .md per plugin
      pipelines/               — .mdx files with code examples
  public/
    css/bbs.css                — unchanged (font swap only)
    js/bbs.js                  — unchanged
  astro.config.mjs
  package.json
```

## Content Collections Schema

```ts
// changelog
{ version: string, date: string }  // body = MarkdownContent

// registry
{ name: string, description: string, tier: 1 | 2,
  repo?: string, command?: string, capabilities: string[] }

// pipelines (MDX)
{ title: string, description: string, order: number }
```

## Key Design Decisions

### Font: JetBrains Mono
Replace `Share Tech Mono` + `VT323` with **JetBrains Mono** (Google Fonts). Full double-line box-drawing glyph coverage (`║ ═ ╔ ╗ ╚ ╝ ╠ ╣`). Single font, one `@import` change in `bbs.css`.

### Sysinfo Box
Rewritten as a `<table>` in `SysinfoBox.astro`. CSS `border: 1px solid var(--purple)` on the table element. `border-top: 1px solid var(--selbg)` divides info rows from key hints. No manual character padding — browser layout handles alignment. Decorative box chars used only in CSS `content:` strings, not in the DOM.

### Nav
`Nav.astro` receives `activePage: string` prop, applies `.active` class to matching link. Written once, composed into `BBS.astro` layout. Uses `·` (U+00B7) separators, which JetBrains Mono renders correctly.

### ANSI Gallery
`themes.astro` renders a grid of `<pre>` blocks from `public/ans/*.ans`. Ships with `welcome.ans`. "Submit a theme" link points to GitHub issues template.

### GitHub Actions
`gh-pages.yml` updated:
- Add Node.js setup step
- `cd site && npm ci && npm run build`
- `publish_dir: site/dist` (was `./docs`)

## Pages

| Page | Route | Source |
|------|-------|--------|
| Landing | `/` | `pages/index.astro` |
| Getting Started | `/getting-started` | `pages/getting-started.astro` |
| Plugins | `/plugins` | `pages/plugins.astro` |
| Pipeline Reference | `/pipelines` | `pages/pipelines.astro` + `content/pipelines/*.mdx` |
| Themes / ANSI Gallery | `/themes` | `pages/themes.astro` |
| Changelog | `/changelog` | `pages/changelog.astro` + `content/changelog/*.md` |
| Plugin Registry | `/registry` | `pages/registry/index.astro` + `content/registry/*.md` |
| Plugin Detail | `/registry/[slug]` | `pages/registry/[slug].astro` |

## Migration Notes

- `docs/` kept in place until Astro build is verified on `gh-pages` branch
- `docs/css/bbs.css` and `docs/js/bbs.js` copied to `site/public/` as starting point
- All existing ANSI art, color logic, hex canvas, typewriter — unchanged
- `_config.yml` and `.nojekyll` move to `site/public/`
