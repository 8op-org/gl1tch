# Unified Site Workflow

Single entry point for all site operations: `glitch run site "your instructions here"`

## File Layout

```
.glitch/workflows/
  site.glitch           # router — classifies input, dispatches
  site-write.glitch     # writer — creates or updates a doc page
  site-gate.glitch      # LLM content gates using C.L.E.A.R.
  site-dev.glitch       # astro dev server (no LLM)
```

All workflows use `.glitch` extension with Janet syntax inside. Requires a one-line change in `janet/src/glitch/main.janet` to discover `.glitch` files alongside `.janet`.

## Extension Change

`main.janet` line 79 searches for `{name}.janet`. Add `.glitch` as the preferred extension:

```janet
(each ext [".glitch" ".janet"]
  (def p (string dir "/" wf-name ext))
  (when (and (os/stat p) (nil? found))
    (set found p)))
```

`.glitch` takes priority. `.janet` is fallback for backwards compat.

## Router (`site.glitch`)

Takes freeform input. One shell step lists existing pages, one LLM step classifies intent, one `call-workflow` dispatches.

### Classification

The router LLM receives:
- The user's freeform input
- List of existing page slugs (from `site/src/content/docs/*.md`)

It returns JSON:

```json
{ "action": "write-page", "slug": "getting-started", "instructions": "update to reflect janet syntax" }
```

Actions:
- `write-page` — create or update a specific page. Extracts slug and instructions.
- `write-all` — regenerate all doc pages. Iterates slugs and calls `site-write` for each.
- `dev` — start the astro dev server.

### Provider

Router uses default tiers (local model). Classification is cheap.

### Dispatch

```janet
(call-workflow "site-write" :set {"page" slug "instructions" instructions})
```

For `write-all`, the router loops sequentially over all existing slugs (each `call-workflow` is a separate LLM call with its own gate pass, so parallel would race on file writes and blow through token limits).

## Writer (`site-write.glitch`)

Receives params: `page` (slug), `instructions` (freeform).

### Steps

1. **check-existing** — shell: check if `site/src/content/docs/{page}.md` exists, cat it if so
2. **gather-context** — shell: cat `docs/site/*.md` stubs, `examples/*.glitch`, recent commits, list all current page slugs for cross-linking
3. **write** — LLM (copilot/Sonnet): generate full markdown body

   Prompt includes:
   - Voice rules (your framing, examples first, no internals)
   - Existing content if updating, or "new page" indicator
   - Repo context from gather step
   - User's instructions
   - Output format: raw markdown starting with `##`, no frontmatter

4. **gate** — `call-workflow site-gate` with generated content + slug + context
5. **save** — shell: inject YAML frontmatter and write to `site/src/content/docs/{slug}.md`

### Frontmatter

For existing pages: preserve current `title`, `order`, `description`.

For new pages: the write step also outputs these values, extracted from a JSON wrapper:

```json
{ "title": "Par Form", "order": 9, "description": "concurrent step execution", "content": "## ..." }
```

## Gate (`site-gate.glitch`)

Two sequential LLM gate steps using the C.L.E.A.R. method. Each returns structured JSON.

### Gate 1: Voice & Structure (C, L, R)

- **Clarity** — writing is explicit, easy to follow. Examples before explanation. No vague language.
- **Logic** — structure makes sense. `##` headings, correct cross-link slugs, logical flow.
- **Relevance** — stays on-topic for the slug. No scope creep. No internal implementation details (BubbleTea, tmux, SQLite, Go types, OTel).

### Gate 2: Accuracy (L, A, R)

- **Logic** — every command, flag, and code example exists in the repo context. No hallucinated features.
- **Actionability** — each issue includes a specific fix with enough detail to act on.
- **Relevance** — no references to dropped tools (Ollama, `glitch ask`). No internal types. Stays within what the repo supports.

### Empathy

Applied to prompt framing: gate reviews the content, not the writer. Issues phrased as "the content does X".

### Output Format

```json
{
  "pass": false,
  "issues": [
    { "category": "L", "line": "glitch ask --query ...", "fix": "glitch ask is decommissioned -- remove this example" },
    { "category": "C", "line": "Use the workflow engine to...", "fix": "too vague -- show the actual command first, then explain" }
  ]
}
```

### Failure Behavior

If either gate fails, the workflow errors with the issues list. Content is not written to disk. No retry loop — re-run with adjusted instructions.

### Provider

Both gates use `copilot` (Sonnet). Reviewer matches writer capability.

## Dev Server (`site-dev.glitch`)

No LLM. Shell only.

1. **ensure-dirs** — `mkdir -p site/generated`
2. **dev** — `cd site && npx astro dev --port 4322`

## Provider Summary

| Workflow | Provider | Reason |
|----------|----------|--------|
| site.glitch (router) | default tiers (local) | classification is cheap |
| site-write.glitch | copilot (Sonnet) | strong long-form writing |
| site-gate.glitch | copilot (Sonnet) | reviewer matches writer |
| site-dev.glitch | none | shell only |

## Usage Examples

```bash
# Update a page
glitch run site "update getting-started to use janet syntax examples"

# Create a new page
glitch run site "add a page about the par form"

# Regenerate everything
glitch run site "refresh all doc pages"

# Dev server
glitch run site "start dev server"
```
