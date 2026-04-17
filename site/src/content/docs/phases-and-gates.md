---
title: "Phases & Gates"
order: 9
description: "Structured execution stages with pass/fail verification checkpoints. A phase groups steps. Gates are shell-based assertions that must pass before the phase completes."
---

Phases group steps into named stages. Gates are verification checkpoints — shell commands that must exit 0 before the phase is considered complete. If a gate fails, the whole phase retries.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it yet, start there — phases and gates use the same step forms, templates, and control flow.

## Quick example

```glitch
(phase "build" :retries 2
  (step "compile"
    (run "make build"))
  (gate "binary-exists"
    (run "test -f ./bin/app")))
```

That phase runs the `compile` step, then checks that the binary actually exists. If the gate fails, the entire phase retries — up to 2 times.

## The `(phase ...)` form

A phase wraps one or more steps and zero or more gates into a named execution stage:

```glitch
(phase "name" :retries N
  (step "work" (run "..."))
  (gate "check" (run "test -f output.json")))
```

| Part | What it does |
|------|-------------|
| `"name"` | Identifier for the phase — shows up in logs and error messages |
| `:retries N` | How many times to retry the entire phase if any gate fails (default: 0) |
| `(step ...)` | Regular steps — run in order, just like outside a phase |
| `(gate ...)` | Verification assertions — run AFTER all steps complete |

Steps inside a phase execute sequentially, same as top-level steps. The difference is what happens after: gates run, and if any gate fails, the whole phase (steps + gates) reruns from the top.

## The `(gate ...)` form

A gate is a shell command that must exit 0. Gates run after all steps in the phase have completed:

```glitch
(gate "check-name"
  (run "test -f output.json"))
```

Gates are assertions, not data steps. They don't produce output that other steps reference — they verify that prior steps did what they were supposed to do. Think of them as post-conditions.

Multiple gates run in order. If any gate fails, the phase retries (if retries remain) or the workflow fails:

```glitch
(phase "validate" :retries 1
  (step "generate" (run "python3 generate.py"))
  (gate "output-exists" (run "test -f output.json"))
  (gate "output-valid" (run "python3 -c 'import json; json.load(open(\"output.json\"))'")))
```

Here, `output-exists` checks the file was created, then `output-valid` checks it parses as JSON. Both must pass.

## Real example: site verification gates

This is from `site-create-page.glitch` — the workflow that generates documentation pages for the gl1tch website. After AI generates a page and rebuilds the site JSON, two phases verify the output:

````glitch
;; site-create-page.glitch — AI generates a doc page, gates verify it
;;
;; Run with: glitch workflow run site-create-page --set topic="batch comparison runs"

(workflow "site-create-page"
  :description "AI-generate a new doc page with gated verification"

  ;; ── Gather context (shell, 0 tokens) ──────────────
  (step "existing-docs"
    (run "for f in docs/site/*.md; do echo '=== '$f' ==='; head -5 \"$f\"; echo; done"))

  (step "examples"
    (run "for f in examples/*.glitch; do echo '=== '$f' ==='; cat \"$f\"; echo; done"))

  ;; ... more context-gathering steps ...

  ;; ── Generate page with LLM ───────────────────────
  (step "generate"
    (llm :tier 0 :prompt ```
      You are a technical writer for gl1tch (8op.org).
      ...
      ```))

  ;; ── Save and rebuild ─────────────────────────────
  (step "save-stub"
    (run "cat '~(stepfile generate)' | python3 scripts/save-generated-page.py"))

  (step "rebuild-json"
    (run "python3 scripts/stubs-to-json.py > site/generated/docs.json && python3 scripts/split-docs.py"))

  ;; ── Verify ───────────────────────────────────────
  (phase "verify" :retries 1
    (gate "no-hallucinations"
      (run "python3 scripts/gate-hallucinations.py"))
    (gate "stub-coverage"
      (run "python3 scripts/gate-coverage.py"))
    (gate "structure-and-tone"
      (run "python3 scripts/gate-structure.py")))

  ;; ── Page tests ───────────────────────────────────
  (phase "page-tests" :retries 0
    (gate "playwright"
      (run "bash scripts/gate-playwright.sh")))

  (step "done"
    (run "echo 'Page created and tested. Run: glitch workflow run site-dev to preview.'")))
````

Two phases, two different verification strategies:

- **`verify`** has `:retries 1` — if the hallucination checker, coverage checker, or tone checker fails, the whole phase reruns once. This gives the pipeline a second chance after transient issues.
- **`page-tests`** has `:retries 0` — Playwright tests either pass or the workflow fails immediately. No point retrying a structural test failure.

Notice that phases can contain only gates (no steps). The `verify` phase is pure assertion — it checks work done by earlier steps outside the phase. The `page-tests` phase runs a single Playwright gate. Both patterns are valid.

## Composing with other control flow

Phases nest with the same wrapper forms you use for steps. Wrap a phase in `timeout` to cap how long verification can take:

```glitch
(timeout "120s"
  (phase "verify" :retries 2
    (step "generate-report" (run "python3 report.py"))
    (gate "report-valid" (run "python3 validate_report.py"))))
```

Wrap in `catch` to provide a fallback when verification fails after all retries:

```glitch
(catch
  (phase "strict-check" :retries 1
    (step "build" (run "make release"))
    (gate "checksum" (run "sha256sum -c checksums.txt")))
  (step "fallback"
    (run "echo 'Checksum verification failed — using debug build'")))
```

Use `retry` around a phase for an additional retry layer on top of the phase's own `:retries`:

```glitch
(retry 2
  (timeout "60s"
    (phase "flaky-integration" :retries 1
      (step "test" (run "bash integration-test.sh"))
      (gate "logs-clean" (run "! grep -q ERROR test.log")))))
```

That gives you: phase retries once internally, timeout caps each attempt at 60 seconds, and the outer retry gives 2 more attempts if the whole thing still fails.

## When to use phases vs plain steps

**Use phases when you need verification checkpoints.** If your workflow generates output that needs to be validated — files exist, JSON parses, tests pass, no hallucinations detected — wrap those steps in a phase with gates.

**Use plain steps for simple linear workflows.** A workflow that gathers data, sends it to an LLM, and saves the result doesn't need phases. Steps already fail the workflow if they exit non-zero.

Rules of thumb:

- **AI-generated output** — use gates. LLMs can produce malformed output. Gates catch it, retries fix it.
- **Multi-step builds** — use gates. Verify artifacts exist and are valid before moving on.
- **Data gathering** — plain steps are fine. `curl` and `gh` either work or they don't. Use `retry` or `catch` for flaky APIs instead.
- **Single-step validation** — just use `retry`. You don't need a phase for one step that might fail.

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — complete reference for all forms, including phase and gate
- [Plugins](/docs/plugins) — package reusable subcommands and compose them into workflows
- [Batch Runs](/docs/batch-runs) — run the same workflow across multiple providers and compare results
