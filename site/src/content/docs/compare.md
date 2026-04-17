---
title: "Compare Runs"
order: 6
description: "Run the same task through different models or strategies, then let a neutral judge score the results against your criteria."
---

Run the same task through different models or strategies, then let a neutral judge score the results. The `(compare ...)` form is a language-level feature — it lives inside your workflow alongside regular steps.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it yet, start there — compare uses the same steps, templates, and control flow.

## Compare models

The simplest case: send the same prompt to two models and see which one answers better.

````glitch
;; compare-models.glitch — compare LLM outputs and pick the best
;;
;; Run with: glitch workflow run compare-models --set topic="Go error handling"
;;
;; Or use --variant for ad-hoc comparison:
;;   glitch workflow run compare-models --variant ollama:qwen2.5:7b --variant claude

(def model "qwen2.5:7b")

(workflow "compare-models"
  :description "Compare different models on the same prompt"

  (step "explain"
    (compare
      (branch "local"
        (llm :model "qwen2.5:7b"
          :prompt "Explain ~param.topic in 3 sentences."))
      (branch "large"
        (llm :model "llama3.2"
          :prompt "Explain ~param.topic in 3 sentences."))
      (review :criteria ("accuracy" "clarity" "conciseness"))))

  (step "report"
    (save "results/compare-~param.topic.md" :from "explain")))
````

What happens when you run this:

1. Both branches execute — `"local"` sends the prompt to `qwen2.5:7b`, `"large"` sends it to `llama3.2`
2. The `(review ...)` block scores each branch on your three criteria
3. The winning branch's output becomes the step result, available downstream as `~(step explain)`
4. The `(save ...)` step writes the winner to disk

## Compare strategies

Branches don't have to use different models. You can compare entirely different approaches to the same problem — different prompts, different pipelines, different numbers of steps.

````glitch
;; compare-branches.glitch — compare different analysis strategies
;;
;; Run with: glitch workflow run compare-branches --set repo=gl1tch

(workflow "compare-branches"
  :description "Compare analysis strategies for a repo"

  (step "files" (run "find ~param.repo -name '*.go' -type f | head -20"))

  (compare
    :id "analysis"
    (branch "breadth-first"
      (step "scan"
        (llm :model "qwen2.5:7b"
          :prompt ```
          List all packages and their responsibilities:
          ~(step files)
          ```))
      (step "summary"
        (llm :model "qwen2.5:7b"
          :prompt "Summarize this package map:\n~(step scan)")))
    (branch "depth-first"
      (step "pick"
        (run "echo '~(step files)' | head -5"))
      (step "deep-dive"
        (llm :model "qwen2.5:7b"
          :prompt ```
          Do a deep analysis of these files:
          ~(step pick)
          ```)))
    (review
      :criteria ("insight_depth" "actionability" "coverage")
      :model "qwen2.5:7b"))

  (step "winner-report"
    (save "results/~param.repo/analysis.md" :from "analysis")))
````

Notice the differences from the first example:

- **`:id "analysis"`** — names the compare block so you can reference it later with `~(step analysis)` or `:from "analysis"`
- **Multi-step branches** — `"breadth-first"` runs two LLM steps in sequence; `"depth-first"` runs a shell step then an LLM step
- **`:model` on `(review ...)`** — pins the judge to a specific model instead of using the default

## How the compare form works

```
(compare
  [:id "name"]
  (branch "label" step...)
  (branch "label" step...)
  (review :criteria ("criterion-1" "criterion-2") [:model "model"]))
```

| Part | Required | Description |
|------|----------|-------------|
| `:id` | no | Name for the compare block (defaults to the enclosing step ID) |
| `(branch "label" ...)` | yes, 2+ | Named execution path — contains one or more steps |
| `(review ...)` | yes | Scoring block — defines what the judge evaluates |
| `:criteria` | yes (on review) | List of criteria the judge scores each branch against |
| `:model` | no (on review) | Model for the judge — defaults to your local model |

Each branch runs independently. The review step sees all branch outputs and scores them. The highest-scoring branch's final step output becomes the compare block's result.

## Review scoring

The review judge is a neutral LLM call. It receives:

- Each branch's label and output
- Your criteria list

It scores each branch 1-10 on each criterion, then picks a winner. The judge model defaults to your local model (`qwen2.5:7b`) — you can override it with `:model` on the `(review ...)` block.

Criteria are strings you define. Pick names that make sense for your task:

- Model comparison: `"accuracy"`, `"clarity"`, `"conciseness"`
- Strategy comparison: `"insight_depth"`, `"actionability"`, `"coverage"`
- Code review: `"correctness"`, `"completeness"`, `"style"`

The judge doesn't know which model or strategy produced each output — it sees branch labels only.

## CLI flags

Three flags on `glitch workflow run` give you compare capabilities without editing workflow files.

### --variant

Inject ad-hoc compare blocks around every LLM step in your workflow. Each `--variant` specifies a `provider:model` pair:

```bash
glitch workflow run code-review \
  --variant ollama:qwen2.5:7b \
  --variant ollama:llama3.2
```

This wraps every LLM step in an implicit `(compare ...)` with one branch per variant. You need at least two variants — with fewer, gl1tch prints a warning and runs normally.

### --compare

Discover sibling variant workflows and cross-review them:

```bash
glitch workflow run my-workflow --compare
```

This looks for variant workflows alongside your main workflow, runs them all, and produces a cross-review of the results.

### --review-criteria

Set review criteria from the command line instead of in the workflow file:

```bash
glitch workflow run code-review \
  --variant ollama:qwen2.5:7b \
  --variant ollama:llama3.2 \
  --review-criteria "accuracy,clarity,conciseness"
```

Comma-separated list. Overrides any `:criteria` defined in the workflow's `(review ...)` block.

## Saving and reading results

Compare results save like any other step output. Use `(save ...)` with `:from` pointing to the compare block's ID:

```glitch
(step "winner-report"
  (save "results/~param.repo/analysis.md" :from "analysis"))
```

The saved file contains the winning branch's output. For the full scoring breakdown, the review scores are written to your results directory automatically.

Read saved results back into a later workflow:

```glitch
(step "previous"
  (read "results/~param.repo/analysis.md"))
```

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the full form reference that compare builds on
- [Plugins](/docs/plugins) — package reusable subcommands and compose them into workflows
- [Local Models](/docs/local-models) — set up the local models your compare judges use
