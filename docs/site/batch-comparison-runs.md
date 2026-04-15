# Batch Comparison Runs

# Batch Comparison Runs

## Overview

gl1tch allows you to run multiple workflows in parallel and compare their results. This page explains how to set up such batch comparisons using parameterized workflows and parallel execution.

## Examples

### 1. Parameterized Workflows for Batch Comparison

Use parameters to run the same workflow with different inputs and collect outputs.

**Example:** The `parameterized.glitch` workflow can be run with varying parameters to compare results across different configurations.

```glitch
(def model "qwen2.5:7b")
(workflow "parameterized"
  :description "Show how to pass runtime parameters into a workflow"

  (step "info"
    (run "echo 'Analyzing repo: {{.param.repo}}'"))

  (step "structure"
    (run "find {{.param.repo}} -maxdepth 2 -type f | head -30"))

  (step "summary"
    (llm
      :model model
      :prompt ```
        Here is the file tree for {{.param.repo}}:

        {{step "structure"}}

        Describe the project structure in 3-4 sentences.
        What kind of project is this?
        ```))

  (step "save-it"
    (save "results/{{.param.repo}}/summary.md" :from "summary")))
```

**Run with different parameters to compare outputs:**
```bash
glitch workflow run parameterized --set repo=gl1tch
glitch workflow run parameterized --set repo=some-other-repo
```

### 2. Parallel Execution for Simultaneous Comparisons

Use (par ...) to execute steps concurrently, which can be used to compare outputs from different models or processes.

**Example:** A workflow that runs multiple LLM steps in parallel and compares their responses.

```glitch
(workflow "parallel-comparison"
  :description "Run multiple LLM steps in parallel and compare results"

  (par
    (step "model-a"
      (llm
        :model "qwen2.5:7b"
        :prompt "Analyze this data:\n{{step \"input\"}}"))
    (step "model-b"
      (llm
        :model "another-model"
        :prompt "Analyze this data:\n{{step \"input\"}}"))
    (step "input"
      (run "echo 'Sample input data'")))
)
```

This example uses the `par` form to run both LLM steps simultaneously, allowing you to compare their outputs.

## Summary

Use parameterized workflows and parallel execution to perform batch comparisons in gl1tch. This approach enables efficient analysis of different configurations or models.
