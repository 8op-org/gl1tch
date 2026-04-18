# Lisp Evaluator: Replace Runner with Direct AST Interpretation

**Date:** 2026-04-17
**Status:** Approved

## Summary

Replace the two-stage workflow execution pipeline (sexpr.go converter + runner.go dispatch loop, ~4500 lines) with a Lisp evaluator (~800-1000 lines) that walks the existing sexpr AST directly. The evaluator reuses the existing parser, provider integrations, and ES client unchanged. `.glitch` files run with identical syntax.

## Motivation

The current runner converts parsed AST nodes into intermediate `Step` structs (80+ fields), then dispatches execution via a large switch-case loop. This architecture:

- Makes every new form require changes in two places (converter + runner)
- Prevents user-defined functions, closures, or let-bindings
- Encodes all control flow in Go switch cases rather than composable forms
- Results in 4500+ lines for what is fundamentally "evaluate an s-expression"

A Lisp evaluator collapses both stages into one: parse AST, evaluate directly. New forms are added by registering a builtin function or special form — no struct fields, no converter code.

## What Changes

### New Files

- **`internal/pipeline/eval.go`** (~800-1000 lines) — Lisp evaluator with Eval() loop, special forms, and builtin registration
- **`internal/pipeline/values.go`** (~80 lines) — Value interface and types: StringVal, ListVal, NilVal, BoolVal, FnVal, BuiltinVal, MapVal
- **`internal/pipeline/env.go`** (~40 lines) — Lexical scope (Env struct with Get/Set/parent chain)

### Modified Files

- **`internal/pipeline/types.go`** — Add `Nodes []*sexpr.Node` field to Workflow struct. Keep Result, RunOpts, StepRecord unchanged (public API stable). The Step struct and its 80+ fields remain temporarily for any code that still references them but are no longer populated by the evaluator path.
- **`internal/pipeline/sexpr.go`** — Simplify to only extract workflow metadata (name, description, args, input) and store raw body nodes in `Workflow.Nodes`. Remove all `convertStep`/`convertForm` functions (~1700 lines deleted).
- **`internal/pipeline/runner.go`** — Replace the body of `Run()` to construct an evaluator, register builtins with the runtime context (provider registry, ES URL, params, etc.), and call `ev.Run()`. The `runCtx` struct, step recording, telemetry wiring, and result assembly stay in runner.go. The `executeStep`/`runSingleStep` dispatch chain (~2000 lines) is deleted.

### Unchanged

- `internal/sexpr/` — parser, lexer, AST, token types (no changes)
- `internal/provider/` — provider registry, RunProviderWithResult (called from llm builtin)
- `internal/esearch/` — ES client (called from search/index/delete builtins)
- `internal/plugin/` — plugin system (called from plugin builtin)
- All `.glitch` workflow files — syntax is identical
- `quasi.go` — `~(step x)` interpolation logic, reused by the evaluator

## Evaluator Architecture

### Value Types

All values implement `Value` interface with `String() string`:

```
StringVal  — primary value type (matches current model where everything is a string)
ListVal    — []Value, used by map/filter/reduce
NilVal     — empty/false
BoolVal    — true/false
FnVal      — closure: params []string, body []*Node, captured Env
BuiltinVal — Go function: func(env *Env, args []*Node) (Value, error)
MapVal     — map[string]Value, for structured data (resources, params)
```

### Lexical Environment

```go
type Env struct {
    bindings map[string]Value
    parent   *Env
}
```

Lookup walks the parent chain. `def` binds in current scope. `let` creates a child scope. `fn` captures its defining scope (closure).

### Special Forms (unevaluated args)

| Form | Behavior |
|------|----------|
| `def` | Bind name to evaluated value in current scope |
| `fn` | Create closure with param list and body |
| `let` | Create child scope with sequential bindings, eval body |
| `do` | Eval forms sequentially, return last |
| `if` | Eval predicate, branch |
| `when` / `when-not` | Conditional eval of body |
| `cond` | Multi-branch conditional (pred body pred body ...) |
| `workflow` | Extract metadata, eval body forms, return Result |
| `step` | Eval body, record output in steps map. Single-arg form returns step output (for `~(step x)` compat) |
| `par` | Eval children concurrently via goroutines |
| `retry` | Re-eval body up to N times on error |
| `timeout` | Eval body with context deadline |
| `catch` | Eval body, run fallback on error |
| `include` | Parse and eval another .glitch file |
| `map` / `each` | Iterate: split prior step output by newlines, eval body per item |
| `filter` | Keep items where predicate is truthy |
| `reduce` | Accumulate over items |
| `compare` | Run branches in parallel, judge with LLM |
| `phase` | Retriable unit with gates |
| `gate` | Verification step inside phase |

### Builtin Functions (args evaluated before call)

| Builtin | Integration |
|---------|------------|
| `sh` / `run` | `exec.Command` shell execution |
| `ref` | Steps map lookup |
| `str` | String concatenation |
| `llm` | `provider.ProviderRegistry.RunProviderWithResult()` |
| `save` | Write step output to file |
| `websearch` | SearXNG HTTP call |
| `search` | `esearch.Client.Search()` |
| `index` | `esearch.Client.Index()` |
| `delete` | `esearch.Client.Delete()` |
| `embed` | `esearch.Client.Embed()` |
| `http-get` / `fetch` | HTTP GET |
| `http-post` / `send` | HTTP POST |
| `read-file` / `read` | `os.ReadFile` |
| `write-file` / `write` | `os.WriteFile` |
| `glob` | `filepath.Glob` |
| `json-pick` / `pick` | jq expression via gojq |
| `plugin` | Plugin subcommand invocation |
| `call-workflow` | Parse + eval nested workflow file |
| `list` | Build a list value |
| `not` / `=` | Boolean operations |
| `println` | Debug output |

### String Interpolation

When a `TokenString` contains `~`, the evaluator expands it before returning the value:

- `~(step x)` — parse the parenthesized form, eval it (step lookup)
- `~(ref x)` — same as above
- `~param.foo` — look up in params map
- `~input` — look up input value
- `~workspace` — look up workspace name

This reuses the pattern from quasi.go but implemented directly in the evaluator's string handling.

### Concurrency Model (par)

`par` spawns one goroutine per child form. The evaluator's `Env` is read-safe (parent chain is immutable once created). Step results are recorded via the mutex-protected steps map. Each goroutine gets the same `env` reference (safe for reads) and writes only to `ev.steps` (mutex-protected).

### Entry Point Wiring

```go
// runner.go — simplified Run()
func Run(w *Workflow, input string, defaultModel string, 
         params map[string]string, reg *provider.ProviderRegistry, 
         opts ...RunOpts) (*Result, error) {
    
    ev := NewEvaluator()
    
    // Wire runtime context
    ev.Input = input
    ev.Params = params
    ev.DefaultModel = defaultModel
    ev.ProviderReg = reg
    // ... wire ESURL, WebSearchURL, Resources, Telemetry, etc.
    
    // Register integration builtins
    ev.RegisterProviderBuiltins()
    ev.RegisterESBuiltins()
    ev.RegisterPluginBuiltins()
    ev.RegisterHTTPBuiltins()
    
    // Evaluate the workflow AST
    val, err := ev.Run(w.Source)
    if err != nil {
        return nil, err
    }
    
    return &Result{
        Workflow: w.Name,
        Output:   val.String(),
        Steps:    ev.Steps(),
    }, nil
}
```

## New Capabilities

Available immediately without syntax changes:

- **User-defined functions**: `(def greet (fn (name) (str "hello " name)))`
- **Closures**: Functions capture their defining scope
- **Let bindings**: `(let (x (sh "date")) (str "Today: " x))`
- **If/cond anywhere**: Not just at step level — inside any expression
- **Composable forms**: Pass functions as arguments, build abstractions
- **Include with eval**: Included files can define functions, not just defs

## Testing Strategy

- All existing `.glitch` test files in `internal/pipeline/testdata/` must pass unchanged
- All existing `_test.go` files that test via `Run()` must pass (Result shape unchanged)
- Add `eval_test.go` exercising new capabilities: fn, let, closures, if/cond in expressions
- Add `eval_compat_test.go` running every example workflow through the evaluator

## Migration

1. Add eval.go, values.go, env.go alongside existing runner.go
2. Simplify sexpr.go to metadata extraction + raw node storage  
3. Replace Run() body to use evaluator
4. Delete dead code (convertStep, convertForm, executeStep, runSingleStep)
5. Run full test suite
6. Remove unused Step struct fields in a follow-up

No flag, no side-by-side mode. Clean replacement.
