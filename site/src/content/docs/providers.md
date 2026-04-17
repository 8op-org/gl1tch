---
title: "Providers & Config"
order: 7
description: "Provider config, Ollama, LM Studio, OpenAI-compat, tiered routing, config.glitch"
---

Config lives at `~/.config/glitch/config.glitch` — an s-expression file that controls which models and providers `glitch` uses by default, how the tier escalation ladder is built, and any custom OpenAI-compatible endpoints you want to add.

## Quick start

```bash
# Show your current config
glitch config show

# Change the default model
glitch config set default-model qwen3:8b

# Change the default provider
glitch config set default-provider ollama
```

## The config.glitch format

`config.glitch` uses the same s-expression syntax as workflows and workspace files. A full example:

```glitch
(config
  :default-model "qwen3:8b"
  :default-provider "ollama"
  :eval-threshold 4
  :workflows-dir "~/.config/glitch/workflows"

  (providers
    (provider "openrouter"
      :type "openai-compatible"
      :base-url "https://openrouter.ai/api/v1"
      :api-key-env "OPENROUTER_API_KEY"
      :default-model "google/gemma-3-12b-it:free")

    (provider "local-server"
      :type "openai-compatible"
      :base-url "http://localhost:8080/v1"
      :api-key-env "LOCAL_API_KEY"
      :default-model "my-fine-tune"))

  (tiers
    (tier :providers ("ollama") :model "qwen3:8b")
    (tier :providers ("openrouter") :model "google/gemma-3-12b-it:free")
    (tier :providers ("copilot" "claude"))))
```

### Top-level keys

| Key | Default | What it does |
|-----|---------|--------------|
| `:default-model` | `"qwen3:8b"` | Model used when no `:model` is specified in a step |
| `:default-provider` | `"ollama"` | Provider used when no `:provider` is specified |
| `:eval-threshold` | `4` | Minimum self-eval score (1–5) before escalating to the next tier |
| `:workflows-dir` | — | Override the global workflows search path |

## Provider types

### Ollama (default local)

Ollama is the default provider. `glitch` talks to it at `http://localhost:11434`. No config required — install Ollama, pull a model, and you're done.

```bash
ollama pull qwen3:8b
glitch run my-workflow
```

See [Local Models](/docs/local-models) for setup details.

### LM Studio

LM Studio exposes an OpenAI-compatible API at `http://localhost:1234`. `glitch` handles it natively — no extra config entry needed. Set your provider to `lm-studio` in a step or workflow:

```glitch
(step "generate"
  (llm :provider "lm-studio" :model "qwen3-8b"
    :prompt "Summarize this: ~(step context)"))
```

`glitch` checks whether the model is loaded and waits for it to be ready before sending the prompt.

### Claude, Copilot, Gemini (agent providers)

Claude (`claude`), GitHub Copilot (`copilot`), and Gemini (`gemini`) run as headless CLI agents. They need their respective CLIs installed and authenticated — no API key config in `config.glitch`.

```glitch
(step "review"
  (llm :provider "claude"
    :prompt "Review this PR: ~(step diff)"))
```

### OpenAI-compatible providers

Any API that implements the OpenAI chat completions spec works as an `openai-compatible` provider. Add it to the `(providers ...)` block:

```glitch
(providers
  (provider "my-endpoint"
    :type "openai-compatible"
    :base-url "https://api.example.com/v1"
    :api-key-env "EXAMPLE_API_KEY"
    :default-model "my-model-7b"))
```

| Key | Required | What it does |
|-----|----------|--------------|
| `:type` | yes | Must be `"openai-compatible"` |
| `:base-url` | yes | Base URL of the API (e.g., `https://openrouter.ai/api/v1`) |
| `:api-key-env` | recommended | Env var name holding the API key — never store keys directly |
| `:api-key` | — | Inline API key (not written to disk on `config show`) |
| `:default-model` | — | Model used when `:model` is not specified in a step |

Popular choices: [OpenRouter](https://openrouter.ai) for free-tier access to many models, [Together AI](https://together.ai), [Fireworks AI](https://fireworks.ai).

## Using providers in workflows

Override the provider or model for any individual step with `:provider` and `:model`:

```glitch
(step "local-draft"
  (llm :provider "ollama" :model "qwen3:8b"
    :prompt "Draft a summary of: ~(step input)"))

(step "polish"
  (llm :provider "claude"
    :prompt "Polish this draft: ~(step local-draft)"))
```

The `:tier` keyword pins a step to a specific escalation tier instead of letting `glitch` pick:

```glitch
(step "cheap-check"
  (llm :tier 0
    :prompt "Is this JSON valid? ~(step output)"))

(step "expensive-review"
  (llm :tier 2
    :prompt "Deep review: ~(step output)"))
```

Tier 0 is local (Ollama), tier 1 is the middle tier, tier 2 is the top tier — as configured in your `(tiers ...)` block.

## Tiered escalation

`glitch` can automatically escalate from a cheaper provider to a more capable one when the response quality isn't good enough. The default tier ladder is:

| Tier | Providers | Purpose |
|------|-----------|---------|
| 0 | `ollama` (qwen3:8b) | Free, local, fast — always tried first |
| 1 | `codex`, `gemini` | Free cloud tier |
| 2 | `copilot`, `claude` | Paid, highest quality |

How escalation works:

1. A step runs against the tier 0 provider
2. `glitch` evaluates the response with a local self-eval model
3. If the self-eval score is below `:eval-threshold` (default 4), the step is retried at the next tier
4. At the final tier, the response is accepted regardless of score

The self-eval uses your local Ollama model — escalation decisions are free.

### Customizing tiers

Override the default ladder in `config.glitch`:

```glitch
(tiers
  (tier :providers ("ollama") :model "qwen3:8b")
  (tier :providers ("openrouter") :model "google/gemma-3-12b-it:free")
  (tier :providers ("copilot" "claude")))
```

Each `(tier ...)` is tried in order. Within a tier, providers are tried left to right — if one errors, `glitch` moves to the next. If all providers in a tier fail structurally or score too low, `glitch` escalates to the next tier.

## Workspace-level defaults

If your workspace has a `(defaults ...)` block, it overrides global config for all runs in that workspace:

```glitch
(workspace "my-project"
  (defaults
    :model "llama3.2"
    :provider "ollama"))
```

Precedence (highest wins): CLI flags > workspace defaults > `config.glitch` global defaults.

## Next steps

- [Local Models](/docs/local-models) — Ollama setup, model selection, and embedding models
- [Workflow Syntax](/docs/workflow-syntax) — `:provider`, `:model`, and `:tier` keywords in steps
- [Workspaces](/docs/workspaces) — per-workspace model and provider defaults
