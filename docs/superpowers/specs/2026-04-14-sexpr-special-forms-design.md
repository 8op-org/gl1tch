# Sexpr Special Forms: par, when, each, par-each

**Date**: 2026-04-14
**Goal**: Add four control flow forms to the .glitch sexpr workflow format — parallel execution, conditionals, and iteration.

## Forms

### `(par ...)` — parallel step execution

All child steps run concurrently via goroutines. The par block waits for all to complete before the workflow continues. Each step writes its output to the shared step map (mutex-protected). Steps after the par block reference any parallel step output via `{{step "id"}}`.

```lisp
(par
  (step "analyze-local"
    (llm :provider "ollama" :model "qwen3-coder:30b" :prompt "..."))
  (step "analyze-claude"
    (llm :provider "claude" :model "sonnet" :prompt "..."))
  (step "analyze-copilot"
    (llm :provider "copilot" :model "gpt-5.2" :prompt "...")))

(step "compare"
  (llm :provider "ollama" :model "qwen3:8b"
    :prompt ```Compare: {{step "analyze-local"}} vs {{step "analyze-claude"}} vs {{step "analyze-copilot"}}```))
```

If any parallel step errors, the entire par block fails and the workflow stops.

### `(when ...)` — conditional execution

Evaluates a condition predicate. If true, runs child steps. If false, skips them (their IDs don't exist in the step map).

```lisp
(when (contains (step "review") "OVERALL: FAIL")
  (step "retry" (llm :provider "ollama" :model "qwen3.5:35b-a3b" :prompt "...")))

(when (empty (step "optional-data"))
  (step "fallback" (run "echo 'no data available'")))

(when (not (empty (step "results")))
  (step "save" (save "output.md" :from "results")))
```

**Condition predicates:**

| Predicate | Evaluates to true when |
|-----------|----------------------|
| `(contains (step "id") "text")` | Step output contains substring |
| `(empty (step "id"))` | Step output is empty string |
| `(not <predicate>)` | Inner predicate is false |

No else branch. Use two `when` blocks with opposite conditions if needed.

### `(each ...)` — sequential iteration

Iterates over a list of values, running child steps once per item. The binding variable is available in step IDs, run commands, prompts, and save paths.

**Inline list source:**
```lisp
(each model ["qwen3:8b" "qwen3-coder:30b" "qwen3.5:35b-a3b"]
  (step "bench-{{model}}"
    (llm :provider "ollama" :model model :prompt "Respond with your model name.")))
```

**Step output source (split on newlines):**
```lisp
(step "list-repos" (run "echo 'ensemble\noblt-cli\nkibana'"))
(each repo (step "list-repos")
  (step "index-{{repo}}" (run "glitch index ~/Projects/{{repo}}")))
```

Iterations run sequentially. Each iteration expands the binding variable in:
- Step IDs: `"bench-{{model}}"` → `"bench-qwen3:8b"`
- Run commands: the binding is available as a template variable and as a def-style symbol
- LLM prompts: same expansion
- Save paths: same expansion
- Model fields: symbol resolution applies (bare `model` resolves to current iteration value)

### `(par-each ...)` — parallel iteration

Identical to `(each ...)` but all iterations run concurrently. Same goroutine + mutex pattern as `(par ...)`.

```lisp
(par-each provider ["ollama" "claude" "copilot"]
  (step "run-{{provider}}"
    (llm :provider provider :model "default" :prompt "...")))
```

## Constraints

- **No nesting.** All four forms appear only as direct children of `workflow`, at the same level as `step`. They cannot contain each other.
- Par, when, each, and par-each bodies contain only `(step ...)` forms.
- This is enforced at parse time — the sexpr converter returns an error if nesting is detected.

## Data Model

### Current

```go
type Workflow struct {
    Name        string
    Description string
    Steps       []Step
}

type Step struct {
    ID       string
    Run      string
    LLM      *LLMStep
    Save     string
    SaveStep string
}
```

### New

```go
type Workflow struct {
    Name        string
    Description string
    Nodes       []ExecNode  // replaces Steps
}

type ExecNode struct {
    Step *Step       // plain step
    Par  []*Step     // parallel steps
    When *WhenNode   // conditional block
    Each *EachNode   // iteration block
}

type WhenNode struct {
    Cond Condition
    Body []*Step
}

type EachNode struct {
    Binding  string
    Source   EachSource
    Parallel bool       // true for par-each
    Body     []*Step
}

type Condition struct {
    Op      string      // "contains", "empty", "not"
    StepRef string      // step ID to test
    Arg     string      // substring for "contains"
    Inner   *Condition  // for "not"
}

type EachSource struct {
    Inline  []string    // literal list ["a" "b" "c"]
    StepRef string      // step ID — output split on newlines
}
```

### YAML backward compatibility

YAML `Unmarshal` still produces `[]Step`. The YAML loader wraps each step:

```go
for _, s := range yamlSteps {
    w.Nodes = append(w.Nodes, ExecNode{Step: &s})
}
```

Zero changes to YAML workflow files.

## Runner Changes

### File: `internal/pipeline/runner.go`

The runner's main loop changes from `for i, step := range w.Steps` to:

```go
for _, node := range w.Nodes {
    switch {
    case node.Step != nil:
        err = runStep(node.Step, ...)
    case node.Par != nil:
        err = runParallel(node.Par, ...)
    case node.When != nil:
        if evalCondition(node.When.Cond, steps) {
            for _, s := range node.When.Body {
                if err = runStep(s, ...); err != nil { break }
            }
        }
    case node.Each != nil:
        items := resolveEachSource(node.Each.Source, steps)
        for _, item := range items {
            for _, s := range node.Each.Body {
                expanded := expandStep(s, node.Each.Binding, item)
                if err = runStep(expanded, ...); err != nil { break }
            }
            if err != nil { break }
        }
    }
    if err != nil { return nil, err }
}
```

### `runStep` extraction

The current step execution logic (save/run/llm branches + telemetry) is extracted from the main loop body into a `runStep(*Step, ...) error` function. This is a pure refactor — no behavior change.

### `runParallel`

```go
func runParallel(parSteps []*Step, ...) error {
    var mu sync.Mutex
    var wg sync.WaitGroup
    var firstErr error
    for _, s := range parSteps {
        wg.Add(1)
        go func(s *Step) {
            defer wg.Done()
            if err := runStep(s, ...); err != nil {
                mu.Lock()
                if firstErr == nil { firstErr = err }
                mu.Unlock()
            }
        }(s)
    }
    wg.Wait()
    return firstErr
}
```

The shared `steps` map needs mutex protection for parallel writes. Use `sync.Mutex` around `steps[step.ID] = out` in `runStep`.

### `evalCondition`

```go
func evalCondition(cond Condition, steps map[string]string) bool {
    switch cond.Op {
    case "contains":
        return strings.Contains(steps[cond.StepRef], cond.Arg)
    case "empty":
        return strings.TrimSpace(steps[cond.StepRef]) == ""
    case "not":
        return !evalCondition(*cond.Inner, steps)
    }
    return false
}
```

### `expandStep`

Creates a copy of the step with the binding variable substituted into ID, Run, LLM.Prompt, LLM.Model, Save, and SaveStep fields:

```go
func expandStep(s *Step, binding, value string) *Step {
    r := strings.NewReplacer("{{"+binding+"}}", value)
    expanded := *s // shallow copy
    expanded.ID = r.Replace(s.ID)
    expanded.Run = r.Replace(s.Run)
    expanded.Save = r.Replace(s.Save)
    expanded.SaveStep = r.Replace(s.SaveStep)
    if s.LLM != nil {
        llm := *s.LLM
        llm.Prompt = r.Replace(s.LLM.Prompt)
        llm.Model = r.Replace(s.LLM.Model)
        llm.Provider = r.Replace(s.LLM.Provider)
        expanded.LLM = &llm
    }
    return &expanded
}
```

For `(par-each ...)`, expand all items first, then pass the expanded steps to `runParallel`.

## Parser Changes

### File: `internal/pipeline/sexpr.go`

The `convertWorkflow` function currently only handles `(step ...)` children. Add cases for `par`, `when`, `each`, `par-each`:

```go
head := child.Children[0].SymbolVal()
switch head {
case "step":
    step, err := convertStep(child, defs)
    w.Nodes = append(w.Nodes, ExecNode{Step: &step})
case "par":
    steps, err := convertParChildren(child, defs)
    w.Nodes = append(w.Nodes, ExecNode{Par: steps})
case "when":
    whenNode, err := convertWhen(child, defs)
    w.Nodes = append(w.Nodes, ExecNode{When: whenNode})
case "each":
    eachNode, err := convertEach(child, defs, false)
    w.Nodes = append(w.Nodes, ExecNode{Each: eachNode})
case "par-each":
    eachNode, err := convertEach(child, defs, true)
    w.Nodes = append(w.Nodes, ExecNode{Each: eachNode})
}
```

### Condition parsing

`(contains (step "review") "FAIL")` parses as a list with 3 children:
1. Symbol `contains`
2. List `(step "review")` → extract step ref
3. String `"FAIL"` → the argument

`(not (empty (step "id")))` parses as:
1. Symbol `not`
2. List `(empty (step "id"))` → recurse

### Inline list parsing

`["a" "b" "c"]` — the lexer already supports `[` and `]` via `TokenLBrace`/`TokenRBrace`. Wait — it uses `{`/`}` for maps. Square brackets aren't in the lexer.

**Option:** Use the existing list syntax `("a" "b" "c")` instead of `["a" "b" "c"]`. Avoids lexer changes. The parser distinguishes inline lists from step calls by checking if the first element is a string (not a symbol).

**Decision:** Use bare list syntax. Inline lists in `each` are just the remaining children after the binding variable and source:

```lisp
(each model ("qwen3:8b" "qwen3-coder:30b")
  (step ...))
```

The parser sees `(each <symbol> <source> <step>...)` where source is either `(step "id")` (list starting with symbol "step") or `("a" "b" "c")` (list starting with a string).

No lexer changes needed.

## Files Modified

| File | Change |
|------|--------|
| `internal/pipeline/types.go` | Add `ExecNode`, `WhenNode`, `EachNode`, `Condition`, `EachSource`. Change `Workflow.Steps` to `Workflow.Nodes`. YAML loading wraps `[]Step` → `[]ExecNode`. |
| `internal/pipeline/sexpr.go` | Add `convertPar`, `convertWhen`, `convertEach`, `convertCondition` functions. Update `convertWorkflow` to handle new head symbols. |
| `internal/pipeline/runner.go` | Extract `runStep`. Add `runParallel`, `evalCondition`, `expandStep`, `resolveEachSource`. Main loop dispatches on `ExecNode` fields. Mutex-protect steps map. |
| `internal/pipeline/runner_test.go` | Tests for each new form. |
| `internal/pipeline/sexpr_test.go` | Parser tests for each new form. |

## Not Included

- No nesting of special forms
- No `else` branch for `when`
- No new token types or lexer changes
- No changes to the research loop or ask command
- No YAML syntax for special forms (sexpr only)
