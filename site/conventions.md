# gl1tch Site Conventions

Rules for generating and editing 8op.org content. Every site workflow
injects this file into LLM prompts. Violations are caught by gate scripts.

## Interpolation Syntax

gl1tch uses tilde-based unquote interpolation. The old Go template
syntax (double-brace) is dead — never generate it.

| Form | Example | Purpose |
|------|---------|---------|
| `~(step name)` | `~(step gather)` | Insert a previous step's output |
| `~(stepfile name)` | `~(stepfile diff)` | Write step output to tempfile, return path |
| `~param.key` | `~param.repo` | Reference a `--set key=value` parameter |
| `~input` | `~input` | Reference user input |
| `~(fn args)` | `~(split "/" "a/b/c")` | String function |
| `~(fn args \| fn2)` | `~(trim " x " \| upper)` | Pipe threading |

**NEVER use:** `{{step "name"}}`, `{{.param.key}}`, `{{.input}}` — these are
old Go template syntax and do not work.

## Tone & Voice

- "your" framing throughout — never say "the user"
- Examples before explanation — show real code first, explain second
- No internal implementation details: BubbleTea, tmux, SQLite, Go types,
  OTel, lipgloss, or internal package names
- Every code example must come from CONTEXT provided, not from training data
- Do NOT invent commands, flags, or features not present in the context

## Decommissioned Features

- `glitch ask` — removed, do not mention in any content

## Frontmatter

| Content type | Required fields |
|-------------|-----------------|
| Docs | `title`, `order`, `description` |
| Labs | `title`, `slug`, `description`, `date` |

## Code Blocks

- Workflow examples use ````glitch` fence
- Shell examples use ````bash` fence
- All workflow code must use current interpolation syntax (see table above)
- Triple backticks inside glitch code delimit multiline strings

## CLI Commands

Only reference commands that exist. Current valid commands:

- `glitch workflow run <name>`, `glitch workflow list`
- `glitch run <name>` (alias for workflow run)
- `glitch observe`, `glitch up`, `glitch down`
- `glitch workspace init|use|list|status|gui|register|unregister|add|rm|sync|pin|workflow`
- `glitch plugin list`, `glitch plugin`
- `glitch config show`, `glitch config set`
- `glitch index`, `glitch version`, `glitch --help`

## Content Format (EDN)

All generated content MUST be valid EDN. Output raw EDN, never markdown.

### Doc schema:
{:title "Page Title"
 :description "One-line description"
 :order 1
 :sections [{:heading "Section" :level 2
             :body [[:p "text"]
                    [:code {:lang "bash"} "command"]
                    [:code "inline-code"]
                    [:a {:href "/docs/other"} "link"]
                    [:diagram {:type :flowchart :direction :lr}
                     [:node :a "Start"] [:node :b "End"] [:edge :a :b]]]}]}

### Lab schema:
{:title "Lab Title" :slug "slug" :description "desc"
 :date "2026-04-21" :duration "15min"
 :steps [{:heading "Step" :body [[:p "text"]]}]}

### Changelog schema:
{:date "2026-04-21"
 :entries [{:type :feat :summary "what changed"}]}
