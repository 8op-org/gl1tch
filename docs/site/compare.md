# Compare Runs

Compare runs let you evaluate the same task across different models, providers, or prompt strategies, then score the results with a neutral judge.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it, start there.

> **Note:** The `(compare ...)` form described here is planned but not yet implemented in the workflow engine. Today, you build comparison workflows manually using standard steps. This page documents the current manual approach and the planned syntax.

## Manual comparison (current approach)

Use parallel steps with explicit providers, then a grading step:

````glitch
(def prompt ```
  Review this PR for correctness and security:
  ~(step diff)
  ```)

(workflow "compare-models"
  :description "Compare PR review quality across providers"

  (step "diff"
    (run "git diff main...HEAD"))

  ;; Run the same prompt through different providers
  (step "review-local"
    (llm :provider "lmstudio" :model "qwen3-8b"
      :prompt prompt))

  (step "review-copilot"
    (llm :provider "copilot"
      :prompt prompt))

  (step "review-claude"
    (llm :provider "claude"
      :prompt prompt))

  ;; Neutral grader scores each variant
  (step "grade"
    (llm :provider "lmstudio" :model "qwen3-8b"
      :prompt ```
        Score each review variant 1-10 on completeness, specificity, and accuracy.

        --- LOCAL ---
        ~(step review-local)

        --- COPILOT ---
        ~(step review-copilot)

        --- CLAUDE ---
        ~(step review-claude)

        Output format:
        WINNER: <variant>
        SCORES: local=N copilot=N claude=N
        REASON: <one sentence>
        ```))

  (step "save-comparison"
    (save "results/comparison.md" :from "grade")))
````

### With timeout for slow providers

Wrap individual review steps with `timeout` to cap slow providers:

```glitch
(timeout "60s"
  (step "review-claude"
    (llm :provider "claude" :prompt prompt)))
```

### Batch comparison across issues

Combine `each` with comparison to evaluate across multiple inputs:

````glitch
(step "issues"
  (run "gh issue list --repo acme/backend --json number --limit 5 | jq -r '.[].number'"))

(each "issues"
  (step "compare-issue"
    (call-workflow "compare-models"
      :set (issue "~param.item"))))
````

## Planned syntax: `(compare ...)`

The planned `(compare ...)` form will simplify this pattern:

````glitch
(compare
  (branch "local"
    :provider "lmstudio" :model "qwen3-8b")
  (branch "copilot"
    :provider "copilot")
  (branch "claude"
    :provider "claude")

  (review :criteria ("completeness" "specificity" "accuracy")
    :judge "lmstudio"
    :model "qwen3-8b"))
````

Each `(branch ...)` runs the same workflow steps with its provider override. The `(review ...)` block scores branches against your criteria using a neutral judge.

### Planned CLI flags

- `--variant <provider>` — inject an ad-hoc comparison branch without editing the workflow
- `--compare` — discover sibling variant workflows and cross-review them
- `--review-criteria "criteria"` — override the review criteria from the command line

## Tips

- **Use a local model as judge.** The grader should be cheap and fast — it's scoring, not generating. `qwen3-8b` handles scoring well.
- **Save results.** Use `(save ...)` so you can compare across runs over time.
- **Pin the grader.** Always use an explicit `:provider` for the grading step so it doesn't escalate — you want the same judge every time.

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) �� the full form reference
- [Workspaces](/docs/workspaces) — batch fan-out for comparing across multiple inputs
