---
title: "Providers"
order: 5
description: "Configure Copilot, Claude, OpenRouter, and LM Studio — with MCP tools and tiered escalation."
---

## Providers

gl1tch routes LLM steps through providers. Each one has different authentication, tool-use behavior, and cost. Pick explicitly with `:provider`, or configure tiered escalation and let gl1tch route automatically.

## Comparison

| Provider | Type | Tool calling | Notes |
|----------|------|-------------|-------|
| **lmstudio** | Local HTTP | `tool_loop` | Free, private. Falls back to no-tools mode if the model doesn't support it. |
| **copilot** | CLI | Via MCP | Attach tools with `--additional-mcp-config`; restrict with `--available-tools`. |
| **claude** | CLI | Via MCP | Attach tools with `--mcp-config`. |
| **openrouter** | HTTP API | `tool_loop` | Broad model selection including free-tier models. |

## Copilot

Copilot runs via the GitHub CLI. Use `--additional-mcp-config` to attach an MCP server and `--available-tools` to restrict which tools the agent can call:

```glitch
(provider "copilot"
  :type "cli"
  :additional-mcp-config "~/.config/glitch/mcp.json"
  :available-tools ("search" "index" "workflow_run"))
```

Both flags compose — attach your MCP config and then restrict to a safe subset. In a workflow:

```glitch
(step "plan"
  (llm
    :provider "copilot"
    :prompt ```
      Produce an implementation plan for this issue:
      ~(step fetch-issue)
      ```))
```

## Claude

Claude runs via the Anthropic CLI. Use `--mcp-config` to point it at an MCP server:

```glitch
(provider "claude"
  :type "cli"
  :mcp-config "~/.config/glitch/mcp.json")
```

With a config in place, Claude can call gl1tch tools natively during a step. In a workflow:

```glitch
(step "review"
  (llm
    :provider "claude"
    :skill "reviewer-verify"
    :prompt "Review these staged changes for correctness, security, and style:\n\n~(step diff)"))
```

## OpenRouter

OpenRouter communicates over HTTP and gives you access to hundreds of models — including free-tier options — through a single API key. Configure it in `~/.config/glitch/config.glitch`:

```glitch
(provider "openrouter"
  :type "openai-compatible"
  :base-url "https://openrouter.ai/api/v1"
  :api-key-env "OPENROUTER_API_KEY"
  :default-model "google/gemma-3-12b-it:free")
```

Set `OPENROUTER_API_KEY` in your environment before running workflows that use this provider.

Tool use with OpenRouter runs a `tool_loop` — gl1tch keeps sending requests until the model stops requesting tools. Each round dispatches any tool calls in the response before continuing. In a workflow:

```glitch
(step "research"
  (llm
    :provider "openrouter"
    :model "google/gemma-3-12b-it:free"
    :prompt ```
      Find and summarize recent activity in ~param.repo.
      ```))
```

## LM Studio

LM Studio communicates over HTTP at `localhost:1234`. It's the recommended tier 0 provider — free, local, and private.

Tool use runs the same `tool_loop` as OpenRouter. If the loaded model doesn't support tool calling, gl1tch automatically falls back to a no-tools prompt — your workflow gets a response either way.

Explicit use in a workflow:

```glitch
(step "classify"
  (llm
    :provider "lm-studio"
    :model "qwen3-8b"
    :format "json"
    :prompt ```
      Classify this issue. Respond with ONLY valid JSON:
      {
        "type": "bug|feature|refactor|documentation",
        "complexity": "small|medium|large",
        "summary": "one line summary"
      }

      ISSUE:
      ~(step fetch-issue)
      ```))
```

See [Local Models](/docs/local-models) for model recommendations, GPU tuning, and context length guidance.

## Tiered escalation

The recommended escalation order is: **lmstudio → copilot → claude → openrouter**.

Start free and local. Escalate to cloud only when local quality isn't sufficient:

```glitch
(tiers
  (tier :providers ("lm-studio") :model "qwen3-8b")
  (tier :providers ("copilot"))
  (tier :providers ("claude"))
  (tier :providers ("openrouter") :model "google/gemma-3-12b-it:free"))
```

With this config in `~/.config/glitch/config.glitch`, any `(llm ...)` step without an explicit `:provider` or `:tier` routes automatically. Tier 0 runs first. If quality clears the eval threshold, routing stops. If not, gl1tch escalates. You pay for cloud only when local can't handle it.

Pin a step to skip escalation when you know what you need:

```glitch
;; Fast and low-stakes — keep it local
(step "classify"
  (llm :tier 0 :format "json"
    :prompt "Classify this issue. Respond with ONLY valid JSON."))

;; High rigor required — go straight to premium
(step "final-review"
  (llm :tier 2
    :prompt "Review this PR plan against acceptance criteria with high rigor."))
```

The eval threshold is configurable:

```glitch
(config
  :eval-threshold 4
  ...)
```

Responses score 1–10. Steps scoring below the threshold escalate to the next tier.

## MCP server

```bash
glitch mcp
```

`glitch mcp` starts an MCP server over stdio, exposing gl1tch's tools — search, index, workflow execution, and more — to any MCP-compatible agent. Once connected, the agent can call those tools natively during a step, regardless of which provider is handling the LLM call.

Create `~/.config/glitch/mcp.json`:

```json
{
  "mcpServers": {
    "glitch": {
      "command": "glitch",
      "args": ["mcp"]
    }
  }
}
```

Then reference it from your provider config:

```glitch
;; Copilot with glitch MCP tools
(provider "copilot"
  :type "cli"
  :additional-mcp-config "~/.config/glitch/mcp.json")

;; Claude with glitch MCP tools
(provider "claude"
  :type "cli"
  :mcp-config "~/.config/glitch/mcp.json")
```

Use `--available-tools` on the Copilot provider to limit which MCP tools are exposed:

```glitch
(provider "copilot"
  :type "cli"
  :additional-mcp-config "~/.config/glitch/mcp.json"
  :available-tools ("search" "index"))
```

## Next steps

- [Local Models](/docs/local-models) — LM Studio model selection, GPU tuning, and context length
- [Workflow Syntax](/docs/workflow-syntax) — `:provider`, `:tier`, and all LLM step options
- [Projects](/docs/workspaces) — scope workflows and run history to your codebase