---
title: "Getting Started"
order: 1
description: "Install glitch and run your first workflow"
---

## Install

```bash
brew install 8op-org/tap/glitch
```

gl1tch routes LLM steps through Ollama by default. Install and start it before running any workflow:

```bash
brew install ollama && ollama pull qwen2.5:7b
```

Verify the install:

```bash
glitch --help
```

## Your first workflow

```bash
glitch workflow run hello-sexpr
```

That runs this workflow:

````glitch
;; hello.glitch — example gl1tch s-expression workflows
;;
;; Run with: glitch workflow run hello-sexpr

(def model "qwen2.5:7b")
(def provider "ollama")

(workflow "hello-sexpr"
  :description "Demo s-expression workflow format"

  (step "gather"
    (run "echo 'hello from a .glitch workflow'"))

  (step "respond"
    (llm
      :provider provider
      :model model
      :prompt ```
        You received this message from a shell command:
        {{step "gather"}}

        Respond with a short, enthusiastic acknowledgment.
        ```)))
````

- `(def model "qwen2.5:7b")` binds a constant — define your model and provider once, reference them anywhere.
- `(step "gather" (run "..."))` runs a shell command and captures its output.
- `(step "respond" (llm ...))` sends a prompt to your local model. `{{step "gather"}}` injects the previous step's output directly into the prompt.

## Your first ask

```bash
glitch ask "review this code"
```

`glitch ask` reads your question, picks the best matching workflow using your local LLM, and runs it. Routing happens entirely on your machine — nothing leaves it. The question above routes to the `code-review` workflow.

## Writing your own workflow

Create `.glitch/workflows/my-workflow.glitch`. The structure below — one shell step feeding one LLM step — is the minimal pattern:

````glitch
(def model "qwen2.5:7b")
(def provider "ollama")

(workflow "hello-sexpr"
  :description "Demo s-expression workflow format"

  (step "gather"
    (run "echo 'hello from a .glitch workflow'"))

  (step "respond"
    (llm
      :provider provider
      :model model
      :prompt ```
        You received this message from a shell command:
        {{step "gather"}}

        Respond with a short, enthusiastic acknowledgment.
        ```)))
````

Rename the workflow, swap in your own shell command and prompt, and you have a working workflow. Use `{{step "id"}}` to pass any step's output into a later step's prompt or shell command. Workflows placed in `.glitch/workflows/` are discovered automatically.

List everything available:

```bash
glitch workflow list
```