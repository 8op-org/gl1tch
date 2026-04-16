---
date: 2026-04-16
status: approved
topic: sexpr-level unquote replacing Go text/template in .glitch workflows
---

# Sexpr Interpolation for .glitch Workflows

## Problem

`.glitch` workflow files are a Clojure-flavored s-expression DSL, but string
interpolation inside them uses Go's `text/template`. That grafts a second,
non-Lispy grammar onto every prompt body, shell command, save path, HTTP body,
and ES index name. Four concrete pains:

1. **Visual noise.** `{{.param.repo}}` and `{{step "diff"}}` clutter otherwise
   clean multiline prose.
2. **Silent-on-missing.** `runner.go:891` sets `missingkey=zero`, so a typo
   like `{{.param.issu}}` renders as empty string. Prompts ship to the LLM
   with unfilled holes and no error.
3. **Dot-prefix gotcha.** `{{param.x}}` is missing the leading dot, which Go's
   template engine treats as a literal — no error, no warning, the template
   stays in the output.
4. **Two grammars.** Function composition inside templates (`| pick "k"`) uses
   Go's pipeline syntax, while the same helpers outside would naturally be
   called with sexpr syntax. Authors context-switch between the two.

The goal of this spec is to replace Go `text/template` with a Clojure-style
unquote mechanism that lives inside the s-expression grammar, so composition
uses one syntax across the whole DSL.

## Surface

Interpolation is only active inside triple-backtick strings. Plain
`"double-quoted"` strings are literal, no unquote.

### Forms

- `~name` — resolve a bare symbol. Looks up `let` bindings first, then `def`
  bindings, then special bindings (`input`, `item`, `item_index`).
- `~param.x`, `~env.x` — dotted-path access on map-like scopes. `param` holds
  `--set key=value` runtime params. `env` is a new scope exposing process
  environment variables.
- `~(form)` — evaluate any s-expression form. Result is stringified via the
  same stringifier used for step output and spliced in.
- `\~` — literal tilde, for cases like `~/home` inside a shell command
  embedded in a triple-backtick.

### Functions replacing template builtins

These move from `text/template` funcs to s-expression built-ins, callable
anywhere in the DSL (not only inside strings):

- Step access: `(step id)`, `(stepfile id)`, `(itemfile)`. Step IDs are bare
  symbols, not quoted strings.
- JSON helpers: `(pick :key json)`, `(assoc :key val json)`.
- String helpers: `split`, `join`, `first`, `last`, `upper`, `lower`, `trim`,
  `trimPrefix`, `trimSuffix`, `replace`, `truncate`, `contains`, `hasPrefix`,
  `hasSuffix`.
- Branch access: `(branch name)`.
- Conditional: `(or a b c …)` — returns the first truthy argument; the escape
  hatch for genuinely optional values (`~(or param.maybe "default")`).

### Example before and after

Before:

```clojure
(step "review"
  (llm :prompt ```
    Review this diff: {{step "diff"}}
    Repo: {{.param.repo}}
    Title: {{.param.item | pick "title" | truncate 50}}
    Optional: {{.param.notes}}
    ```))
```

After:

```clojure
(step "review"
  (llm :prompt ```
    Review this diff: ~(step diff)
    Repo: ~param.repo
    Title: ~(truncate 50 (pick :title param.item))
    Optional: ~(or param.notes "none provided")
    ```))
```

## Internals

### Lexer

`internal/sexpr/lexer.go` already handles triple-backtick with lex-time
dedent (`lexMultilineString`, lines 124-190). Extend that function: while
scanning the body, recognize three new patterns:

- `\~` — emit a literal `~` to the fragment buffer, consume the backslash.
- `~` followed by an identifier start — flush current literal fragment, read
  a symbol-or-dotted-path token (`~param.x` captures `param.x` as a single
  reference token), emit a `TokenQuasiRef` part.
- `~(` — flush current literal fragment, recursively invoke the main sexpr
  reader starting at the `(`, consume until matching `)`, emit a
  `TokenQuasiForm` part carrying the parsed form.

Emit a single `TokenQuasiString` whose value is the ordered list of parts.
Dedent behavior is unchanged — it still applies to literal fragments at lex
time.

### Parser

`internal/sexpr/parser.go` translates the token into a new AST node:

```go
type QuasiStringNode struct {
    Parts []QuasiPart
    Pos   Position
}

type QuasiPart struct {
    Literal *string // one of these is set
    Form    *Node
    Ref     *RefPath // e.g. {Base: "param", Path: []string{"x"}}
}
```

`Node` grows a `QuasiString *QuasiStringNode` variant alongside the existing
`Atom`, `List`, etc.

### Runtime

`internal/pipeline/runner.go` today wires 19 template call sites through
`render()`, which invokes `text/template`. The rewrite:

1. Delete the `funcMap`, `template.New`, `Option("missingkey=zero")`, and
   `template.Execute` plumbing (lines 752-903).
2. Replace with `renderQuasi(node *QuasiStringNode, env *Scope) (string, error)`
   that walks `Parts`, evaluates `Form` and `Ref` parts against `env`, and
   concatenates. Stringification uses the same rules as today's step output
   (strings as-is, other values via `fmt.Sprint`).
3. The 19 call sites still route through a single `render()` entry point —
   they just pass the pre-parsed `QuasiStringNode` instead of a raw string.
4. Move the 11 string helpers and `pick`/`assoc` from template funcs to the
   s-expression evaluator's built-in function table. They become first-class
   sexpr functions usable anywhere.

### Scope resolution

One `Scope` struct threads through evaluation. Resolution order for bare
symbols:

1. `let` lexical bindings (innermost to outermost).
2. `def` top-level bindings.
3. Special bindings, only where contextually valid:
   - `input` — always available.
   - `item`, `item_index` — only inside a `(map …)` body. Outside that body,
     references to `~item` fail with `UndefinedRefError`.
4. If nothing matches, return `UndefinedRefError{Symbol, File, Line, Col}`.

Dotted paths are a separate resolver. `param.x` reads from the workflow's
`--set` params map. `env.x` reads from `os.Getenv`. Missing dotted-path keys
also raise `UndefinedRefError` — same fail-loud behavior as bare symbols.

## Error behavior

`UndefinedRefError` fails the step immediately, not the whole workflow. The
error message points at the exact ref:

```
step "review" failed: undefined reference 'param.issu' at workflows/pr-review.glitch:23:14
  suggestion: did you mean 'param.issue'?
```

Suggestion uses Levenshtein distance against available names in the current
scope; omitted if no close match.

The escape hatch for genuinely optional refs is `~(or form default)`. `or`
returns the first truthy argument, where truthy means non-empty string and
non-nil. This stays in the sexpr language (not a separate sigil), so there's
no second grammar to learn.

## Non-goals

- **No `~@` splice in v1.** No current workflow uses splice, step output is
  already newline-delimited, and adding splice later is non-breaking (today a
  `~@` would be a parse error inside a quasi-string, so no collision). Ship
  when someone has the real need.
- **No backwards compatibility shim for `{{…}}`.** Per project policy (no
  migrations pre-1.0), we hard-break. Every `.glitch` file gets rewritten in
  the same PR as the DSL change.
- **No interpolation inside `"double-quoted"` strings.** Only triple-backtick
  is quasi-quoted. One rule, no surprise.
- **No `{{if}}` / `{{range}}` equivalents in strings.** Workflow-level
  control already lives at the sexpr layer via `cond`, `when`, `map`. Adding
  string-level conditionals would duplicate that.

## Migration

Existing workflow files (`~/.config/glitch/workflows/`, `.glitch/workflows/`,
workspace `<ws>/workflows/`, and all golden files in the repo) are rewritten
in the same PR. Mechanical substitutions cover the common cases:

| Old                            | New                                |
| ------------------------------ | ---------------------------------- |
| `{{.input}}`                   | `~input`                           |
| `{{.param.X}}`                 | `~param.X`                         |
| `{{step "X"}}`                 | `~(step X)`                        |
| `{{stepfile "X"}}`             | `~(stepfile X)`                    |
| `{{itemfile}}`                 | `~(itemfile)`                      |
| `{{.param.item}}`              | `~item`                            |
| `{{.param.item_index}}`        | `~item_index`                      |
| `{{X \| pick "k"}}`            | `~(pick :k X)`                     |
| `{{X \| upper}}`               | `~(upper X)`                       |
| `{{X \| truncate 50}}`         | `~(truncate 50 X)`                 |
| Chained pipes `{{X \| a \| b}}` | Nested calls `~(b (a X))`         |

The rewrite is performed by a one-shot script (`scripts/rewrite-quasi.go`)
checked in alongside the change PR for review reproducibility, then deleted
at merge. This is not a migration in the "carry legacy behavior forward"
sense — it runs once, never again, and no code path in the shipped binary
knows about the old syntax. Edge cases (nested template conditionals, if any
exist) fall out as parse errors in the new grammar and need manual fixes.
Grep of the current workflow corpus shows no `{{if}}`, `{{range}}`, or
`{{with}}` usage, so the mechanical pass is expected to cover 100% of
files.

## Testing

Five layers:

1. **Lexer unit tests** — `internal/sexpr/lexer_test.go` grows a `TestQuasiString`
   table: literal-only, single `~sym`, single `~(form)`, mixed, escaped
   `\~`, nested parens inside `~(…)`, dedent interaction, multiline edge
   cases, unterminated quasi-string.
2. **Parser unit tests** — round-trip quasi-strings through parse + print,
   verify `QuasiStringNode` part ordering.
3. **Runner unit tests** — one test per call-site type: shell command, LLM
   prompt, save path, save from, HTTP URL / body / headers, read/write-file
   paths, glob dir, ES index / doc / id, embed input, plugin args, cond
   predicate, when predicate, call-workflow input. 19 call sites, one focused
   test each.
4. **Scope resolver tests** — let shadows def, def shadows specials, dotted
   paths on missing maps, suggestion distance, `or` short-circuit.
5. **Golden output tests** — for each rewritten `.glitch` workflow in the
   repo corpus, capture the expected rendered output of each step given
   fixed inputs and assert the new evaluator produces it. These are not
   parity tests against the old path (the old path is deleted in this PR),
   they're forward-looking golden files committed alongside the rewrite.

Then the full smoke pack (baseline 24/24 against ensemble, kibana,
oblt-cli, observability-robots) before merge. Smoke pack regressions block
the merge regardless of unit test state.

## Risks

- **Lexer complexity.** Nested sexpr inside a multiline string requires
  careful delimiter tracking — `~(upper (pick :k param.x))` must balance
  parens correctly while also respecting the closing triple-backtick.
  Mitigated by reusing the existing sexpr reader recursively from within
  the string lexer, so paren balancing logic is shared with the main parser.
- **Helper semantics drift.** The 11 string helpers move from Go template
  funcs to sexpr built-ins. Signatures change (positional args instead of
  pipe target), so semantics must be verified case-by-case. Golden parity
  test covers this.
- **Performance.** Each quasi-string now parses at workflow-load time
  instead of at render time. This is a win — today every render re-parses
  the template. No performance regression expected.
- **Triple-backtick shell commands with literal `~/`.** A workflow that
  writes `(run ```cp x ~/dest```)` changes meaning — `~/dest` now tries to
  resolve a ref named `/dest`, fails with undefined-ref error. Fix: either
  escape as `\~/dest` or switch that command to a `"double-quoted"`
  string. Grep of current corpus: zero such cases. Low-risk.

## Downstream enables

This change is not about help output, but it enables it. Once `~param.X`
references are first-class AST nodes, a static walk of a workflow file can
enumerate every `--set` key the workflow consumes, every step it depends
on, and every env var it reads. A follow-up spec for `glitch run <workflow>
--help` (and the equivalent GUI surface) will consume that AST walk — with
today's `text/template` strings, the same feature would require regex
scraping of prompt bodies. Interpolation-first unblocks metadata-extraction
work that follows.

## Work decomposition

This spec covers one coherent change. An implementation plan will break it
into:

1. Lexer extension (`TokenQuasiString`, escape handling, recursive reader).
2. Parser AST node (`QuasiStringNode`, `QuasiPart`, `RefPath`).
3. Scope resolver and `UndefinedRefError` with suggestions.
4. Runtime rewrite of `render()` and deletion of `text/template` wiring.
5. Migration of the 11 string helpers + `pick`/`assoc` to sexpr built-ins.
6. Rewrite tool for existing `.glitch` files.
7. Hard cut-over: delete old path, rewrite all workflow corpora, run smoke
   pack.

Each is independently testable. The writing-plans skill will convert this
into a step-by-step plan next.
