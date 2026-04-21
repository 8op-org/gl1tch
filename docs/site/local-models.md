# Local Models

gl1tch runs your LLM locally by default. Nothing leaves your machine unless you tell it to. This page covers how to set up and tune local inference for the best experience with LM Studio — the default local provider.

This page builds on [Getting Started](/docs/getting-started). If you haven't installed glitch yet, start there.

## LM Studio setup

### Install and configure

1. Download [LM Studio](https://lmstudio.ai) and launch it
2. Go to **Settings → Server** and enable the local server (port 1234)
3. Pull a model — we recommend **qwen3-8b** for general use

gl1tch auto-detects LM Studio at `localhost:1234`. If a model isn't loaded, gl1tch logs a warning and waits. If a model isn't downloaded at all, gl1tch triggers the download via LM Studio's API and polls until it's ready.

### Recommended models for Apple Silicon

These are tested on an M2 Pro with 32GB. If you have 16GB, drop down one size in each category.

| Task | Model | Size | Why |
|------|-------|------|-----|
| **General / tool use** | `qwen3-8b` | ~5GB | Best tool-use quality at this size. gl1tch default. |
| **Classification / JSON** | `qwen3-8b` | ~5GB | Reliable structured output with `:format "json"` |
| **Code review** | `qwen3-coder-30b` | ~18GB | Understands code deeply. Use `:tier 0` to pin. |
| **Fast triage** | `qwen3.5-a3b` | ~2GB | Mixture-of-experts, only 3B active. Very fast. |
| **Heavy reasoning** | `qwen3-30b` | ~18GB | When 8b isn't enough but you don't want to escalate to cloud |

### GPU allocation and context length

LM Studio lets you control how much of the model stays in GPU memory and how large the context window is. Both affect speed and quality.

**In LM Studio → Model Settings:**

- **GPU Offload Layers**: Set to maximum (`-1` or "all"). On an M2 Pro with 32GB, you can fully offload an 8B model with room to spare. For 30B models, you'll partially offload — LM Studio handles the split automatically.
- **Context Length**: Default is usually 4096. For gl1tch workflows that feed multiple step outputs into a single LLM prompt, set this to **8192** or **16384**. The issue-to-pr workflow can produce 6-10k tokens of context easily.
- **Batch Size**: Leave at default (512) unless you're running batch comparison workflows, then bump to 1024.

**Quick reference for 32GB M2 Pro:**

| Model | GPU Layers | Context | Tokens/sec (approx) |
|-------|-----------|---------|---------------------|
| qwen3-8b | all | 8192 | 35-45 |
| qwen3-8b | all | 16384 | 25-35 |
| qwen3.5-a3b | all | 8192 | 60-80 |
| qwen3-coder-30b | ~60% | 8192 | 10-15 |

**For 16GB machines:** Stick with 8B or smaller models. Set context to 4096-8192. The 30B models won't fit comfortably.

### Keep one model loaded

LM Studio loads models on first request and keeps them warm. The first call is slow (10-30s for loading). After that, it's fast.

Tip: before starting a glitch session, load your default model in LM Studio manually. Or just run any workflow once — gl1tch triggers the load and waits.

If you switch between models frequently, LM Studio can keep multiple models loaded if you have the RAM. On 32GB, you can comfortably keep qwen3-8b loaded while also loading a specialized model for specific steps.

### Using LM Studio in workflows

Explicit provider:

```glitch
(step "classify"
  (llm
    :provider "lmstudio"
    :model "qwen3-8b"
    :format "json"
    :prompt "Classify this issue..."))
```

Or use it as your default tier 0 in `~/.config/glitch/config.glitch`:

```glitch
(tiers
  (tier :providers ("lmstudio") :model "qwen3-8b")
  (tier :providers ("copilot"))
  (tier :providers ("claude")))
```

With tiers configured, omit `:provider` and gl1tch auto-routes. Tier 0 (LM Studio) runs first. If quality is too low, it escalates.

### Pin tiers by step complexity

In production workflows, pin cheap steps to local and reserve cloud for what needs it:

````glitch
;; Classification: fast, low stakes — local
(step "classify"
  (llm :tier 0 :format "json"
    :prompt "Classify this issue..."))

;; Research: needs depth — let it escalate
(step "research"
  (llm
    :prompt "Produce an implementation plan..."))

;; Review: fast pass/fail — local is fine
(step "review"
  (llm :tier 0
    :prompt "Review against acceptance criteria..."))
````

## Performance checklist

Before running heavy workflows:

- [ ] One model is loaded and warm (run a test prompt first)
- [ ] GPU offload is set to max for your primary model
- [ ] Context length is at least 8192 for multi-step workflows
- [ ] No other GPU-heavy apps competing for memory
- [ ] For batch runs: close browser tabs, Xcode, anything that eats RAM

## Troubleshooting

**First request is very slow (10-30s)**
Normal — the model is loading into GPU memory. Subsequent requests are fast. Pre-warm by running a simple workflow first.

**"Connection refused" on port 1234 or 11434**
The provider isn't running. Start LM Studio's server or run `ollama serve`.

**Output quality is poor on local models**
Try a larger model or add `:format "json"` for structured output. If that doesn't help, let the tiered router escalate — omit `:tier` and `:provider` and gl1tch will try local first, then cloud.

**Out of memory / system swap**
Your model is too large for available RAM. Drop to a smaller model or reduce context length. On 32GB, stay at 8B for full speed. 30B models will work but slower due to partial CPU offload.

**gl1tch says "lm-studio: loading model, expect delay"**
The model exists but isn't in GPU memory yet. gl1tch waits automatically. To avoid this delay, keep your primary model loaded in LM Studio.
