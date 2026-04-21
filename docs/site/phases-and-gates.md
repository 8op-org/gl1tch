# Phases & Gates

Phases group steps into stages with pass/fail verification checkpoints. Gates are assertions that must pass before a phase completes. If a gate fails, the whole phase retries.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it, start there.

## Quick example

```glitch
(phase "verify" :retries 1
  (step "check"
    (run "curl -sf https://my-site.dev/new-page"))
  (gate "not-empty" (run "test -s output.html")))
```

The phase runs both steps. If the gate exits non-zero, the whole phase retries (up to 1 retry). If it still fails, the workflow errors.

## How phases work

A `(phase "name" ...)` groups steps into a named block. Without gates, a phase is just a label for logging and observability. Add `(gate ...)` assertions to enforce correctness:

```
phase starts → run all steps → run all gates
  if all gates pass → phase complete
  if any gate fails → retry entire phase (if retries remain)
  if retries exhausted → throw error
```

### Syntax

```glitch
(phase "name" [:retries N]
  (step ...)
  (step ...)
  (gate "gate-id" (run "assertion command"))
  (gate "gate-id" (call-workflow "verification-workflow")))
```

| Part | Required | Description |
|------|----------|-------------|
| `"name"` | yes | Phase identifier, shown in logs and run records |
| `:retries` | no | Max retries on gate failure (default 0 — no retries) |
| `(step ...)` | yes | One or more steps to execute |
| `(gate ...)` | no | Assertions that must pass (exit 0) |

## Gates

A gate is a step whose output is treated as a boolean — truthy means pass, falsy or exception means fail. Gates run after all steps in the phase.

### Shell gates

The most common form. Exit 0 means pass:

```glitch
(gate "file-exists" (run "test -f output/report.md"))
(gate "valid-json" (run "jq . < output/data.json"))
(gate "no-errors" (run "! grep -q ERROR build.log"))
```

### Workflow gates

Call a separate workflow as a gate. The child workflow's final output is the gate value:

```glitch
(gate "hallucinations" (call-workflow "gate-hallucinations"))
(gate "syntax"         (call-workflow "gate-syntax"))
```

This keeps verification logic reusable across phases. Write the gate workflow once, use it everywhere.

## Real-world example: site deployment

This workflow from the gl1tch site pipeline verifies generated pages before deploying:

````glitch
(workflow "site-deploy"
  :description "Verify and deploy site changes"

  (step "build"
    (run "cd site && npm run build"))

  ;; Content verification — retry once if gates fail
  (phase "verify" :retries 1
    (step "check-pages"
      (run "find site/dist -name '*.html' | head -20"))
    (gate "hallucinations" (call-workflow "gate-hallucinations"))
    (gate "syntax"         (call-workflow "gate-syntax"))
    (gate "structure"      (call-workflow "gate-structure"))
    (gate "links"          (call-workflow "gate-links"))
    (gate "sidebar"        (call-workflow "gate-sidebar")))

  ;; Smoke tests — no retries, fail fast
  (phase "smoke" :retries 0
    (gate "playwright" (run "cd site && npx playwright test --reporter=line 2>&1")))

  (step "deploy"
    (run "cd site && npm run deploy")))
````

Two phases with different retry strategies:

- **verify** retries once — LLM-generated content sometimes needs a second pass
- **smoke** doesn't retry — if Playwright tests fail, something is structurally wrong

## Composing with other control flow

Phases work with `retry`, `timeout`, and `catch`:

```glitch
;; Timeout the entire phase
(timeout "120s"
  (phase "heavy-verify" :retries 2
    (step "analyze" (llm :prompt "..."))
    (gate "quality" (run "test $(wc -w < output.md) -gt 100"))))
```

```glitch
;; Catch phase failure and run fallback
(catch
  (phase "strict-verify" :retries 1
    (gate "perfect" (run "diff expected.json actual.json")))
  (step "fallback"
    (run "echo 'verification relaxed — manual review needed'")))
```

## When to use phases

**Use phases when** you need verification checkpoints — content quality gates, build smoke tests, or any assertion that should trigger a retry.

**Use plain steps when** your workflow is linear and doesn't need verification. Phases add structure; don't add them where simple sequential steps work fine.

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the full form reference
- [Plugins](/docs/plugins) — reusable data-gathering subcommands
