---
title: Getting Started
order: 1
description: Install glitch and run your first workflow
---

## Install

- brew install 8op-org/tap/glitch
- requires: ollama running locally (brew install ollama && ollama pull qwen2.5:7b)
- verify: glitch --help

## Your first workflow

- show the hello.glitch example from examples/hello.glitch
- walk through what each step does: (def ...) binds constants, (step ...) runs shell or LLM
- show how to run it: glitch workflow run hello-sexpr

## Your first ask

- glitch ask "what time is it" routes your question to the best matching workflow
- routing works via local LLM — nothing leaves your machine
- show 2-3 examples of ask routing to different workflows

## Writing your own workflow

- create .glitch/workflows/my-workflow.glitch
- minimal example: one shell step, one LLM step
- mention step references: {{step "id"}} chains outputs
- glitch workflow list to see available workflows

## Tone

- "your" framing throughout
- examples before explanation
- no internal implementation details (no BubbleTea, no tmux, no SQLite)
