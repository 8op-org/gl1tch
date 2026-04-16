---
date: 2026-04-16
status: approved
topic: glitch run <workflow> --help, sourced from in-file metadata and AST walk
---

# Help From Workflow

## Problem

`.glitch` workflow files declare what they do (`:description`), what flags
they take (implicit `~param.X` references, or explicit `(arg ...)` forms for
plugin subcommands), and what positional input they expect (`~input`). None
of that reaches the CLI. `glitch run <workflow> --help` prints cobra's
generic run help — nothing about the workflow in front of you.

Two pains, one audience each:

1. **Authoring-time discovery.** You have ~20 workflows in `.glitch/workflows/`
   and each has different param conventions. Running one you haven't touched
   in weeks means opening the `.glitch` file to remember the param names.
2. **Publishing UX.** Workflows are starting to live inside plugins and be
   re-exposed as team commands. The "it's just a file" posture doesn't hold
   up when someone else needs to use the workflow you wrote.

A secondary goal lives on the same infrastructure: `UndefinedRefError`
currently reports `undefined reference "x" (did you mean "y"?)` with no file
or line. The fields `File`/`Line`/`Col` were dropped in the sexpr
interpolation merge (session notes
`docs/superpowers/specs/2026-04-16-sexpr-interpolation-design.md`) because
nothing populated them. Walking the AST to build help output gives us the
locations for free; threading them through the existing error type is a
small delta on top.

## Surface

### Invocation

```
glitch run <workflow> --help
```

`--help` on the `run` subcommand short-circuits before `pipeline.Run` and
writes the formatted help text to stdout, exit 0. No other invocation
surfaces in this spec: no `glitch help <workflow>`, no upgrade to
`glitch workspace workflow list`. Both are obvious follow-ups.

### Authoring: `arg` and `input` forms

Flag-style params use the existing `(arg ...)` form already supported by
plugin subcommands (`internal/plugin/args.go`). Top-level in the file,
outside `(workflow ...)`. Same syntax plugin authors already use.

```clojure
(arg "topic"
  :required    true
  :description "Topic of the new doc page — becomes the H1 and filename stem."
  :example     "batch comparison runs")

(arg "audience"
  :default     "developers"
  :description "Target reader profile."
  :example     "ops engineer")
```

Supported keywords on `arg`:

| Keyword        | Meaning                                                       | New in this spec |
|----------------|---------------------------------------------------------------|------------------|
| `:default`     | Default string value; omitted = required (unless `:type :flag`) | no             |
| `:type`        | `:string` (default), `:flag`, `:number`                       | no               |
| `:description` | One-line description shown in help                             | no               |
| `:required`    | Explicit required flag; normally implied by absence of `:default` | no           |
| `:example`     | Concrete example value, spliced into the generated usage line | **yes**          |

The positional input uses a new sibling form:

```clojure
(input
  :description "Free-form context passed as the first positional arg."
  :example     "fix the streaming API latency spike in EU-west")
```

Rules for `input`:

- Optional; at most one per file.
- Workflow-level (top of file, sibling of `arg` and `workflow`), not nested.
- Supported keywords: `:description`, `:example`. No `:required` (cobra
  already makes the positional optional), no `:default` (empty string is
  the existing baseline).
- Unlike `arg`, `input` takes no name; the form is fixed to bind to the
  `input` scope.

### Auto-extraction baseline

A workflow with no `(arg ...)` and no `(input ...)` still gets help output.
The help extractor walks every render-capable string in the workflow (same
fields `renderQuasi` processes today) and uses the existing `lexQuasi`
tokenizer to emit `ref` and `form` tokens. For each token:

- **Bare ref** `~param.X` → collect `X` as an implicit arg.
- **Inside a form** `~(or param.X "default")`, `~(upper param.X)` → walk the
  form AST; collect every atom matching `param.X`.
- **Deduplicate** by name.
- **Skip quote-first-arg forms.** `~(step diff)`, `~(stepfile ...)`, and
  `~(branch ...)` treat their first argument as a literal symbol (per the
  sexpr interpolation spec). The walker consults the existing `quoteFirstArg`
  map and skips those positions.

Any implicit entry is represented as an `ArgDef` with `Implicit: true` and
no description. Declared `(arg "X" ...)` entries replace implicit ones for
the same name. Declared entries with no matching reference anywhere in the
workflow produce a load-time warning:

```
warning: arg "X" declared in <file> is not referenced in any step
```

Transitive collection through `(call-workflow ...)` is **out of scope**. A
child workflow's args belong to its own help page.

### Help output

Cobra-idiomatic, rendered by `formatHelp(w *pipeline.Workflow) string`:

```
site-create-page - AI-generate a new doc page with gated verification

Usage:
  glitch run site-create-page [<input>] --set topic=<topic> [--set audience=<audience>]

Arguments:
  input       Free-form context passed as the first positional arg.
              Example: "fix the streaming API latency spike in EU-west"

Flags:
  topic       (required)                     Topic of the new doc page — becomes the H1 and filename stem.
                                             Example: --set topic="batch comparison runs"
  audience    (optional, default: developers) Target reader profile.
                                             Example: --set audience="ops engineer"

Defined in: .glitch/workflows/site-create-page.glitch:6
```

Column alignment via `text/tabwriter` (same dependency used by
`workspace workflow list`). When `Implicit: true`, the description line
reads `(undocumented — add (arg "<name>" :description "...") to annotate)`.
When the workflow has no `:description`, the top line degrades to the bare
workflow name with no trailing dash. `Defined in:` is always printed.

## Architecture

Three touch points, additive only.

```
.glitch file ──▶ sexpr parser ──▶ Workflow{ Steps, Args, Input, SourceFile }
                      │                          │
                      │                          ▼
                      │                 cmd/run.go: --help? ──yes──▶ formatHelp()
                      │                          │                         │
                      │                         no                         ▼
                      │                          ▼                    stdout
                      │                    pipeline.Run(...)
                      ▼
              Step.Line / Step.Col ──▶ UndefinedRefError{ File, Line, Col }
```

### 1. Parser changes (`internal/pipeline/sexpr.go`)

- Extend `LoadBytes`/`LoadFile` to recognise top-level `(arg ...)` and
  `(input ...)` forms before entering the workflow form. `arg` forms are
  parsed via the existing `plugin.ParseArgs` — pipeline already imports
  plugin, so no new import arrow. `input` is pipeline-local; its parser
  lives in `internal/pipeline/args.go` (new).
- After constructing `Workflow`, run a walk over every render-capable string
  in every `Step` and compound-form body to populate an implicit arg set.
  Merge with declared args; set `Implicit: true` on any not explicitly
  declared.
- Populate `Step.Line` and `Step.Col` from the corresponding `sexpr.Node`.
- Populate `Workflow.SourceFile` from the file path passed to `LoadFile`,
  or the name parameter passed to `LoadBytes`.

### 2. Struct changes

`internal/plugin/args.go` already owns `ArgDef` and `ParseArgs` — and
`internal/pipeline` already imports `internal/plugin` via `plugin_runner.go`,
so reusing `plugin.ArgDef` at the workflow layer adds no new import arrow.
We extend the existing type in place rather than moving it.

`internal/plugin/args.go`:

```go
type ArgDef struct {
    Name        string
    Type        string // "string", "flag", "number"
    Default     string
    Description string
    Example     string // new
    Required    bool
    Implicit    bool   // new: true when auto-extracted without a declaration
}
```

`internal/pipeline/types.go`:

```go
type Workflow struct {
    // ... existing fields ...
    Args       []plugin.ArgDef // declared + implicit
    Input      *InputDef       // optional; nil when no (input ...) and no ~input ref
    SourceFile string
}

type Step struct {
    // ... existing fields ...
    Line int
    Col  int
}

type InputDef struct {
    Description string
    Example     string
    Implicit    bool
}
```

`plugin.ParseArgs` is already callable from `internal/pipeline` (same import
that `plugin_runner.go` uses). The workflow parser calls it on the file
bytes before / alongside the existing workflow conversion pass. `InputDef`
is pipeline-local; `input` isn't a plugin concept.

### 3. Help command (`cmd/run.go`)

Cobra already auto-generates a `--help` flag for every command. The current
`runRun` function can check whether the cmd was invoked with `--help` by
using cobra's built-in flag — except we need workflow-specific help, not
cobra's generic help. Two options:

- Override `runCmd.SetHelpFunc` to a function that loads the workflow when
  the first positional arg is present and renders our custom help. Cobra
  calls this when it handles `--help`, before running `RunE`.
- Intercept manually in `runRun` by checking `cmd.Flags().Changed("help")`.

The former is the idiomatic cobra path; we use it. The help function
resolves the workflow path (same logic as `runRun`), calls
`pipeline.LoadFile`, and writes `formatHelp(w)` to `cmd.OutOrStdout()`.

### 4. Source-location threading (`internal/pipeline/scope.go`, `quasi.go`)

- `UndefinedRefError` regains its `File`/`Line`/`Col` fields. Field format
  in `Error()` becomes `path/to/file.glitch:17:5: undefined reference "x"
  (did you mean "y"?)` when locations are populated; falls back to the
  current format when they're zero.
- `render()` takes the current `Step` as context (via an optional argument
  or a small `RenderContext` struct). When the scope resolver returns an
  `UndefinedRefError`, the renderer stamps `Workflow.SourceFile`,
  `Step.Line`, `Step.Col` before returning.
- Intra-string column pointers (pointing at the exact column of `~param.typo`
  inside a multi-line prompt) are **out of scope**. Step-level locations are
  good enough for the immediate goal; finer grain pays back less than it
  costs.

## Error handling

Parse-time (loud — workflow fails to load):

- Unknown keyword on `(arg ...)` or `(input ...)`. Catches `:defalt` typos.
- Both `:required true` and `:default "x"` on the same `arg`.
- More than one `(input ...)` in a file.
- `(input "foo" ...)` — the form takes no name, unlike `arg`.

Load-time warnings (stderr, don't fail):

- Declared `arg` with no matching `~param.X` reference anywhere.
- Declared `input` with no `~input` reference.

No warning when an `arg` omits `:description`; authors upgrade incrementally.

Run-time (from source-location threading):

- `UndefinedRefError` for any `~param.X` with neither an `arg` declaration
  nor a runtime `--set` value — rendered with the source location.

## Testing

Five layers.

- **`internal/pipeline/sexpr_test.go` additions.** Parse `(arg ...)` and
  `(input ...)` forms; verify struct population; verify parse errors for
  every bad case in §Error handling. Verify `Step.Line`/`Col` are populated
  and stable across a representative workflow.
- **`internal/pipeline/helpdoc_test.go` (new).** Table-driven tests for
  implicit extraction: string with one `~param.x`, string with
  `~(or param.x "default")`, string with `~(step diff)` (skipped), triple-
  backtick with multiple refs, nested compound-form bodies (`par`, `map`,
  `map-resources`, `cond` branches). Merge test: declared `arg` entry
  replaces implicit one; declared but unreferenced produces the expected
  warning.
- **`internal/pipeline/help_format_test.go` (new).** Golden-output tests for
  `formatHelp(w)`. Cases: workflow with all fields populated, workflow with
  `:description` only (implicit args), workflow with no `:description` and
  no `arg` forms (truly bare), workflow with `(input ...)` only, mixed.
- **`cmd/run_test.go` additions.** Integration: `glitch run <name> --help`
  intercepts before `pipeline.Run`, writes rendered help to stdout, exit 0.
  Verify by capturing cobra's out stream.
- **`internal/pipeline/scope_test.go` update.** Verify
  `UndefinedRefError.File/Line/Col` are populated end-to-end when an
  unknown `~param.X` is encountered during render.

Smoke pack (`glitch smoke pack`) runs unchanged — help output doesn't affect
workflow semantics. Post-merge smoke run confirms the 24/24 baseline still
holds.

## Out of scope

- `glitch help <workflow>` as a top-level command. Follow-up.
- `glitch workspace workflow list` showing per-workflow arg summaries.
  Follow-up; one job per command.
- Transitive arg collection through `(call-workflow ...)`. Each workflow
  owns its own help.
- Intra-string column pointers inside multi-line prompts for
  `UndefinedRefError`. Step-level is enough.
- Type coercion on `--set` values. Today everything is a string; `:type`
  on `arg` stays cosmetic in help output until a concrete use case forces
  coercion.
- `:enum` and `:secret` keywords on `arg`. YAGNI until a workflow needs them.
- Markdown / HTML rendering of help. Terminal plain text only.
