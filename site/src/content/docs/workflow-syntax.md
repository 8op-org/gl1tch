---
title: Workflow Syntax
order: 2
description: S-expression workflow reference for glitch
---

## Overview

- glitch workflows use s-expressions (.glitch files)
- Lisp-like syntax: (form arg1 arg2 :keyword value)
- lives in .glitch/workflows/ for auto-discovery

## Workflow structure

- (workflow "name" :description "..." ...steps...)
- show a complete small example from examples/code-review.glitch

## Definitions

- (def name value) binds constants
- used for DRY: define model/provider once, reference everywhere
- show example from examples/hello.glitch

## Steps

- (step "id" (run "shell command"))
- (step "id" (llm :prompt "..."))
- (step "id" (save "path" :from "step-id"))
- each step produces a named output

## Step references

- {{step "id"}} inserts a prior step's output into prompts or commands
- {{.input}} for workflow input
- {{.param.key}} for --set key=value parameters
- show parameterized.glitch example

## LLM options

- :provider — "ollama", "claude", "copilot", "gemini", or custom
- :model — model identifier
- :skill — prepend skill context to prompt
- :format — "json" or "yaml" for structural validation
- :tier — pin to specific tier (0, 1, 2)

## Tiered cost routing

- no :provider and no :tier → auto-route through tiers
- tier 0: local (ollama, free), tier 1: cheap cloud, tier 2: premium
- self-eval at each non-final tier, escalates if quality too low
- :format triggers structural validation (must parse as JSON/YAML)

## Comments and discard

- ; line comments
- #_ discard next form (useful for debugging)

## Multiline strings

- triple backticks for readable multi-line prompts
- auto-dedented

## Tone

- practical, show-don't-tell
- every concept gets a real example from examples/
- no internal Go types or parser details
