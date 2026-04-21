---
title: "REPL-Driven Docs: Writing a Page Without Leaving Emacs"
slug: "repl-driven-docs"
description: "Walk through building the code intelligence doc from a live glitch REPL in Emacs — step by step, eval by eval, with grounding checks before save."
date: "2026-04-21"
---

## The Problem

You need a documentation page. You have a design spec, a codebase, and an LLM. The usual approach: copy-paste into a chat window, ask for a draft, paste the output into a file, squint at it, repeat. Four tabs. Three context switches. One hour of your life you won't get back.

Or: you open a workflow in Emacs, step through it with `C-c C-e`, inspect every intermediate result inline, and save when you're satisfied. One pane. One process. The machine does the fetching. You do the thinking.

This lab walks through exactly that — building the [Code Intelligence](/docs/code-intelligence) doc from scratch using `glitch repl` and CIDER.

---

## Setup

Two terminals. One Emacs.

Terminal 1 — start the REPL:

```bash
export OPENROUTER_API_KEY=$(grep OPENROUTER_API_KEY ~/.env | cut -d= -f2)
glitch repl
```

You'll see:

```
glitch repl on port 1667
connect: cider-connect localhost 1667
```

Terminal 2 — open Emacs and connect:

```
emacs
M-x glitch-repl-connect RET
```

CIDER connects. The `user>` prompt appears in the right pane. Every glitch DSL function — `step`, `llm`, `search`, `ref`, `save`, `grounded?` — is already available. No require needed.

---

## The Workflow

Open the workflow file:

```
C-x C-f .glitch/workflows/update-code-intel-doc.glitch RET
```

It looks like this:

```clojure
(workflow "update-code-intel-doc"
  :description "Update the code intelligence documentation page"

  (step "spec"
    (read-file "docs/superpowers/specs/2026-04-17-code-graph-design.md"))

  (step "voice-ref"
    (read-file "site/src/content/docs/tool-use.md"))

  (step "current"
    (sh "cat site/src/content/docs/code-intelligence.md 2>/dev/null || echo '(new file)'"))

  (step "impl-context"
    (search "tree-sitter|symbol.*graph|code.*index" :limit 20))

  (step "draft"
    (llm
      :provider "openrouter"
      :model "google/gemini-2.5-flash"
      :tools []
      :prompt (str
        "You are writing user-facing documentation for gl1tch.\n\n"
        "DESIGN SPEC (ground truth):\n" (ref "spec") "\n\n"
        "VOICE/FORMAT REFERENCE:\n" (ref "voice-ref") "\n\n"
        "CURRENT DOC:\n" (ref "current") "\n\n"
        "IMPLEMENTATION CONTEXT:\n" (ref "impl-context") "\n\n"
        "INSTRUCTIONS: " (input) "\n\n"
        "RULES:\n"
        "- Voice: user-first, 'your' framing, examples before explanation\n"
        "- Every code example must use real DSL syntax from the spec\n"
        "- Never mention internals: tree-sitter, Elasticsearch, SHA256\n"
        "- Focus on what the user can DO\n\n"
        "Return the complete markdown file.")))

  (step "verify"
    (grounded? "draft" (ref "spec")
      :provider "openrouter"
      :model "google/gemma-3-12b-it"
      :strict false))

  (step "save"
    (save "site/src/content/docs/code-intelligence.md" (ref "draft"))))
```

Seven steps. You don't run them all at once. That's the point.

---

## Step-Through

Put your cursor at the closing paren of each step. Hit `C-c C-e`. The result appears inline — right next to the code, no buffer switching.

### 1. Load the spec

```clojure
(step "spec"
  (read-file "docs/superpowers/specs/2026-04-17-code-graph-design.md"))
```

`C-c C-e`. The full design spec appears as an overlay. This is ground truth — every claim in the final doc must trace back here.

### 2. Load a voice reference

```clojure
(step "voice-ref"
  (read-file "site/src/content/docs/tool-use.md"))
```

The LLM needs to know what your existing docs sound like. Not a style guide — an actual page it can pattern-match against.

### 3. Check for an existing doc

```clojure
(step "current"
  (sh "cat site/src/content/docs/code-intelligence.md 2>/dev/null || echo '(new file)'"))
```

First run returns `(new file)`. Subsequent runs return the current version so the LLM can diff against it.

### 4. Search the codebase

```clojure
(step "impl-context"
  (search "tree-sitter|symbol.*graph|code.*index" :limit 20))
```

Ripgrep runs against your repo. The results give the LLM real implementation details — function names, file paths, actual code. Not hallucinated API surfaces.

### 5. Generate the draft

This is the expensive step. `C-c C-e` and wait. The LLM gets four context blocks — spec, voice reference, current doc, implementation details — plus your instructions and rules.

Inspect the result:

```clojure
(ref "draft")
```

Read it. Does it sound like your other docs? Does it claim features that don't exist? Does it mention Elasticsearch or tree-sitter? (It shouldn't. The rules said not to.)

If it's wrong, edit the prompt and re-eval. You don't restart anything. The spec, voice reference, and search results are still in memory from earlier steps. Change one thing, eval one step.

### 6. Grounding check

```clojure
(step "verify"
  (grounded? "draft" (ref "spec")
    :provider "openrouter"
    :model "google/gemma-3-12b-it"
    :strict false))
```

A second LLM compares the draft against the spec. Returns `true` or `false`. If false, the draft has claims that aren't supported by the design document.

`:strict false` means it won't throw — you decide whether to fix it or ship it. Check what failed:

```clojure
(ref "verify")
```

### 7. Save

```clojure
(save "site/src/content/docs/code-intelligence.md" (ref "draft"))
```

File written. `git diff` to review, commit when ready.

---

## Why This Works

The workflow is a script. The REPL makes it a conversation.

Running `glitch run update-code-intel-doc "add code graph section"` from the CLI executes all seven steps in sequence. That's fine for CI or batch runs. But when you're authoring — when you're figuring out the right prompt, the right context window, the right voice — you want to stop between steps. Look at what the machine produced. Adjust. Re-run one thing without re-running everything.

`C-c C-e` gives you that. Each eval is a single step. Each result is visible inline. The REPL holds state between evals — `(ref "spec")` still returns the spec you loaded three minutes ago. You're not copy-pasting between tabs. You're not re-running a pipeline from scratch because step 5 was wrong.

The workflow file is the plan. The REPL is the debugger.

---

## Useful Bindings

| Keys | What it does |
|------|-------------|
| `C-c C-e` | Eval form at point, show result inline |
| `C-c C-k` | Eval the entire buffer |
| `C-c C-c` | Interrupt a stuck eval |
| `C-c C-q` | Disconnect from REPL |
| `C-x b *cider-repl*` | Switch to REPL buffer |
| `M-x glitch-repl-connect` | Reconnect to glitch REPL |

---

## Takeaway

A doc pipeline is just a pipeline. The interesting part is what happens when you can pause it, look inside, and steer. The REPL doesn't make the LLM smarter. It makes *you* faster at deciding whether the LLM was right.

The workflow took seven steps. Getting it right took about twelve evals. That's the real number — the one that includes the adjustments, the re-prompts, the "no, don't mention Elasticsearch." Twelve evals, one Emacs session, zero tab switches. The machine fetched. You thought. The doc shipped.
