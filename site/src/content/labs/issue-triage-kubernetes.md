---
title: "Triaging a Kubernetes Issue with Tier Routing"
slug: "issue-triage-kubernetes"
description: "Free model attempts it, paid model nails it — real input, real output from kubernetes/kubernetes#138431"
date: "2026-04-17"
---

## The Scenario

[kubernetes/kubernetes#138431](https://github.com/kubernetes/kubernetes/issues/138431) -- kubelet resets `cpu.max` and `memory.*` values under `/sys/fs/cgroup/kubepods.slice/` every minute. If you manage cgroups externally on your nodes, your settings get silently clobbered on every reconcile loop. No flag to disable it. SIG Node tagged it `priority/critical-urgent` in under fifteen minutes.

The chain: PR [#137719](https://github.com/kubernetes/kubernetes/pull/137719) removed the alpha `MemoryQoS` feature gate from `setMemoryQoS()` inside `UpdateCgroups`, making kubelet unconditionally overwrite cgroup values. PR [#137889](https://github.com/kubernetes/kubernetes/pull/137889) then removed downstream reset logic as a side-effect fix but left the ungated call in place. Two PRs, one regression, zero opt-out.

---

## The Workflow

Same prompt, three providers. You compare the output yourself.

````glitch
(def triage-prompt ```
  You are a senior SRE triaging a Kubernetes bug report.
  Analyze this issue: severity, root cause, affected components, next steps.
  Be specific — reference exact PRs, code paths, or config if mentioned.

  Issue data:
  ~(step fetch)
  ```)

(workflow "triage"
  :description "Triage a GitHub issue across three model tiers"

  (step "fetch"
    (run "gh issue view ~input -R kubernetes/kubernetes --json title,body,comments,labels"))

  (step "triage-local"
    (llm :provider "lm-studio" :model "qwen3-8b"
      :prompt ~triage-prompt))

  (step "triage-paid"
    (llm :provider "openrouter" :model "qwen/qwen3.5-flash-02-23"
      :prompt ~triage-prompt))

  (step "triage-copilot"
    (llm :provider "copilot" :model "sonnet"
      :prompt ~triage-prompt)))
````

```bash
glitch run triage 138431
```

---

## Local: qwen3-8b (LM Studio) -- $0.00

<div class="model-output">

**Severity: Critical.** Kubelet's cgroup reconciliation is overwriting externally-set values, breaking resource isolation in production. The `priority/critical-urgent` label is justified.

**Root cause:** Removal of the `MemoryQoS` feature gate in PR #137719 made kubelet unconditionally reconcile cgroup values. PR #137889 removed downstream reset logic but left the ungated `setMemoryQoS()` call in place.

**Affected component:** `kubelet` cgroup reconciliation -- cited as `pkg/kubelet/cgroup.go`. Versions guessed at "v1.26+".

**Fix direction:** Reintroduce the feature gate for memory QoS. Audit other cgroup properties for similar unintended resets. Add safety guards for external cgroup modifications.

</div>

**The verdict.** Gets severity right and identifies the two-PR chain as the trigger. Recommends re-gating behind a feature gate, which is exactly what PR #138430 did. But it mis-attributes the gate removal to #137889 instead of #137719, cites the wrong source file (`pkg/kubelet/cgroup.go` -- actual path is `pkg/kubelet/cm/qos_container_manager_linux.go`), speculates about a race condition that does not exist, and guesses the version range as "v1.26+" when this is a v1.36 regression. Usable as a first pass. Not something you paste into a PR comment.

---

## Paid: qwen/qwen3.5-flash-02-23 (OpenRouter) -- ~$0.001

<div class="model-output">

**Severity: Critical (Urgent).** Confirmed by @SergeyKanzhelev within minutes. No feature gate means no opt-out -- affects all nodes running the affected kubelet version.

**Root cause:** PR **#137719** removed the alpha feature gate, making kubelet unconditionally reconcile memory cgroup values. PR **#137889** attempted mitigation but left the core reconciliation firing. The memory QoS loop overwrites `cpu.max` under `/sys/fs/cgroup/kubepods.slice/` every minute.

**Evidence from thread:** @SergeyKanzhelev confirmed `cpu.max` values reset to `max` even when set by external processes. @liggitt noted moving back to a gate-protected off-by-default block is the right call.

**Fix direction:** Revert the gate removal from #137719. If memory QoS stays enabled, scope reconciliation strictly to pod-specific containers. Assess backport to stable release branches.

</div>

**The verdict.** A real step up. Correctly pins #137719 as the triggering change, quotes @SergeyKanzhelev and @liggitt verbatim from the thread, and captures the "no feature gate = no opt-out" framing. The symptom path is accurate. Version guesses (v1.31/v1.32) are still off -- this is v1.36 -- but it hedges with "Main/Development," which is correct. Still missing: the secondary regression from the fix itself, exact source file, and cgroupv2-only scoping. Actionable enough to hand to a maintainer.

---

## Copilot: Claude Sonnet -- Copilot seat

<div class="model-output">

**Severity: Critical.** Classified `priority/critical-urgent` by @SergeyKanzhelev within ~14 minutes. No opt-out. Affects `cpu.max` beyond the memory-only scope of the original change.

**Root cause chain:**

- **PR #137719** (line R393-R395): removed `MemoryQoS` alpha feature gate from `setMemoryQoS()` call path inside `UpdateCgroups`
- **PR #137889**: removed cgroup reset logic but left the unconditional `setMemoryQoS()` call in place
- **Result:** `UpdateCgroups` in `pkg/kubelet/cm/qos_container_manager_linux.go` clobbers both `memory.*` and `cpu.max` on cgroupv2 hosts

**Secondary regression:** PR **#138430** gated the entire `setMemoryQoS` call, but made the zero-out branch at `qos_container_manager_linux.go:L303-L308` unreachable. Nodes with the alpha feature previously enabled now retain stale `memory.min`/`memory.low` values, skewing OOM killer behavior. PR **#138439** proposes the correct fix: move the gate check *inside* `setMemoryQoS`.

**Status:** PR #138430 merged 2026-04-17. PR #138439 open, marked `do-not-merge/hold`. Milestone: **v1.36**.

</div>

**The verdict.** The only output that names the exact source file, the exact PR line reference, the cgroupv2-only scope, the correct milestone, and the secondary regression introduced by the fix itself. It identified that #138430 was already merged and shifted focus to what remains open. No version speculation, no hallucinated code paths. You could hand this to a maintainer and they would not need to do their own research.

---

## Comparison

| Model | Severity | Root Cause | Actionable? | Cost |
|---|---|---|---|---|
| qwen3-8b (LM Studio) | Correct | Partial -- wrong PR attribution, wrong file | Directional, has hallucinations | $0.00 |
| qwen3.5-flash (OpenRouter) | Correct | Correct chain, missing secondary regression | Quotes thread accurately, version off | ~$0.001 |
| Claude Sonnet (Copilot) | Correct | Full chain + secondary regression in #138439 | Exact file, line, PR, milestone | Copilot seat |

---

## Takeaway

The free model found the right severity and the general fix direction at zero cost -- enough to route the issue to the right SIG. The paid model added thread context and correct PR attribution. Copilot identified the secondary regression introduced by the fix itself, the kind of analysis that normally requires reading three PRs and two code diffs. Tier routing lets you run cheap first and escalate when the stakes justify it.
