---
title: "Model Showdown: Free vs Paid vs Copilot"
slug: "model-showdown-containerd"
description: "Same containerd feature request through three tiers — see where free falls short and paid shines"
date: "2026-04-17"
---

## The Scenario

[containerd/containerd#11339](https://github.com/containerd/containerd/issues/11339) wants a new CRI method called `UpdatePodSandboxResources`. The backstory: InPlacePodVerticalScaling (KEP-1287) made pod sandbox resources mutable at runtime, but NRI plugins never got the memo. A pod resizes. The plugins keep working with the old numbers. Nobody told them.

The proposed fix is deliberately narrow. An informational-only CRI call. No cgroup mutations. Kubelet still owns the writes; containerd just forwards the update envelope to NRI. Implementation PR [#11406](https://github.com/containerd/containerd/pull/11406) is code-complete and waiting on one external gate: Kubernetes 1.33 (~April 23) has to ship before `cri-api` can be tagged.

Good test case. Real cross-project dependency chain. A live thread with an architectural disagreement buried in the comments. Three models, same prompt, different price tags.

---

## The Workflow

Save as `analyze-issue.glitch` and swap the issue number to point at anything.

````glitch
(def analysis-prompt ```
  You are a container runtime architect analyzing a feature proposal.
  1. Summarize the problem and proposed solution
  2. Identify cross-project dependencies (CRI, NRI, Kubernetes)
  3. Assess implementation complexity (small/medium/large)
  4. Flag risks or concerns
  5. Suggest an implementation approach

  Issue data:
  ~(step fetch)
  ```)

(workflow "analyze-issue"
  :description "Analyze a feature proposal across three model tiers"

  (step "fetch"
    (run "gh issue view 11339 -R containerd/containerd --json title,body,comments,labels"))

  (step "analysis-local"
    (llm :provider "lm-studio" :model "qwen3-8b"
      :prompt ~analysis-prompt))

  (step "analysis-paid"
    (llm :provider "openrouter" :model "qwen/qwen3.5-flash-02-23"
      :prompt ~analysis-prompt))

  (step "analysis-copilot"
    (llm :provider "copilot" :model "sonnet"
      :prompt ~analysis-prompt)))
````

```bash
glitch run analyze-issue
```

---

## Local: qwen3-8b (LM Studio) — $0.00

<div class="model-output">

**Problem:** Pod sandbox cgroups are Kubelet-managed, but InPlacePodVerticalScaling makes resources mutable. NRI plugins have no visibility into changes after `RunPodSandbox`.

**Solution:** New `UpdatePodSandboxResources` informational CRI call. Propagates resource updates to NRI plugins without mutating cgroups.

**Dependencies identified:** Kubernetes (KEP + PR #128123), NRI (containerd/nri#141), CRI-O (parallel effort), CRI API version skew policy.

**Complexity:** Medium. New CRI method plus coordination across Kubernetes and NRI for compatibility.

**Risks:**
- Blocked until Kubernetes cuts the CRI API version (April 23)
- NRI plugins may not be updated to handle the new method
- Current solution may not address pod label/annotation updates for service mesh plugins
- CRI API version misalignment across runtimes

**Approach:** Coordinate with K8s on CRI PR, extend containerd with CRI call, update NRI plugins, validate cross-project compatibility.

</div>

Structurally complete. All five sections present, correct dependency graph, right complexity call. But it reads the thread like a summary, not like an architect. The `klihub`/`chrishenzie` annotation semantics debate — the part that could cause a real design mistake — never surfaces. Four risks listed, no severity weighting, no distinction between a hard blocker and a long-tail concern. Good enough for triage. Not enough to make a call.

---

## Paid: qwen/qwen3.5-flash-02-23 (OpenRouter) — ~$0.001

<div class="model-output">

**Problem:** IPVS makes sandbox resources mutable, but the CRI only allows resource definition during `RunPodSandbox`. NRI plugins operate on stale assumptions after a resize.

**Solution:** New `UpdatePodSandboxResources` gRPC call. Passively notifies containerd, which forwards to NRI hooks. No cgroup mutations.

**Dependencies:**
- **Kubernetes core:** CRI API change (PR #128123), Kubelet must invoke new call during IPVS ops. Blocked on K8s 1.33.
- **kubernetes/cri-api:** Proto definition must be cut as a versioned release. Separate from kubernetes/kubernetes.
- **containerd/nri:** PR #141 for new UpdatePodSandbox hook.
- **CRI-O:** Parallel dependency (issue #8990). Ecosystem fragmentation risk if adoption lags.

**Complexity:** Medium. Code logic is low/minimal — mapping CRI requests and forwarding to NRI. Version skew handling is high. API semantics alignment is medium. Timeline constraints dominate over coding constraints.

**Risks:**
1. **Annotation/Label Sync Gap (Critical):** `MikeZappa87` raised pod label sync concerns. `klihub` warned the CRI `annotations` field annotates the *update operation*, not the pod itself. Service mesh plugins relying on dynamic pod labels will not be updated via this RPC. Requires a separate KEP or API extension.
2. **Release coupling:** Code is ready, deployment blocked on K8s 1.33. If K8s ships without the CRI change, containerd users get a call with no client.
3. **Security context:** Pushing state changes to the sandbox requires ensuring NRI plugins are sandbox-aware regarding privilege escalation during resize.
4. **Backward compatibility:** Older K8s versions treat the call as unsupported. Mixed-runtime clusters may have inconsistent IPVS behavior.

**Approach:** Merge PR #11406 now (delaying is unnecessary). Track CRI API cut as a milestone. Address label sync separately — do not repurpose the `annotations` field. Verify NRI integration handles the "informational" contract. Add doc comments marking the feature gate.

</div>

This earns its fraction of a cent on one specific call: the annotation field semantics. It names the conflict — who said what, why it matters, what the wrong move looks like — with enough precision that a maintainer could act on it. It also surfaces a risk the free model skipped: security context around privilege escalation during resize. The `kubernetes/cri-api` module appears as a distinct dependency, not conflated with `kubernetes/kubernetes`. This is the output you put in front of a team deciding whether to merge.

---

## Copilot: Claude Sonnet — Copilot seat

<div class="model-output">

**Problem:** InPlacePodVerticalScaling makes sandbox resources mutable. Kubelet manages the cgroups, but NRI plugins have no channel to observe changes. They stay in the dark after a resize.

**Solution:** Informational-only `UpdatePodSandboxResources` CRI method. No runtime state mutation. Pass-through to NRI plugin notification.

**Dependencies:**

| Project | What | Status |
|---------|------|--------|
| kubernetes/kubernetes | CRI API PR #128123 — defines the new RPC | Blocked on K8s 1.33 |
| kubernetes/cri-api | Versioned release of new proto type | Blocked on K8s 1.33 cut |
| containerd/nri | PR #141 — new hook/event type | In progress |
| cri-o | Parallel implementation (issue #8990) | Parallel effort |
| containerd/containerd | PR #11406 — implementation | Code complete, blocked on deps |

Follows the CRI API version skew policy correctly: containerd and Kubernetes drafts must exist simultaneously before CRI API merges.

**Complexity:** Medium. CRI handler is low complexity. NRI dispatch is low-to-medium. The dependency chain is the primary complexity driver, not the code.

**Risks:**
- **[BLOCKER]** Release coupling. Fully blocked on K8s 1.33 (~April 23). Slip in K8s directly delays containerd.
- **[WATCH]** Annotation semantics ambiguity. Thread shows disagreement: `chrishenzie` suggests carrying pod labels/annotations for NRI, `klihub` correctly identifies these as annotating the update operation itself. Scope confusion invites misuse.
- **[WATCH]** Pod labels/annotations not included. Limits NRI's ability to support service mesh and network plugins. Gap in KEP design, not fixable in containerd alone.
- **[SOUND]** Informational-only design is correct. No double-management bugs. Clear responsibility boundary.

**Approach (sequenced):**
1. [DONE] containerd PR #11406 — CRI handler + NRI dispatch
2. [DONE] NRI PR #141 — protocol hook
3. [WAITING] K8s 1.33 cut → cri-api tagged
4. [THEN] Bump cri-api dep, unblock #11406 merge
5. [THEN] Integration test: resize pod, verify NRI plugin receives event

Recommends clarifying doc comment on `annotations` field semantics. Separate upstream issue for pod label propagation if service mesh use case is real.

</div>

The most immediately actionable output. The dependency table includes current status — you do not have to decode the thread to know what is blocked. Risk items carry severity markers, so triage is instant. The implementation sequence is a numbered dependency chain ready to paste into a PR description. It validates the informational-only architecture rather than just summarizing it. The annotation ambiguity gets the clearest framing of any tier: it annotates the update event, not the pod.

---

## Comparison

| Metric | Free | Paid | Copilot |
|--------|------|------|---------|
| Problem summary | Correct, no thread nuance | Correct, thread-aware | Correct, architecture-validated |
| Cross-project deps | 4/5 — cri-api implicit | 5/5 — cri-api explicit, release schedule noted | 5/5 — dep table with current status |
| Complexity assessment | Medium, unqualified | Medium — code low, integration high | Medium — dep chain is the driver |
| Risk identification | 4 risks, no severity | 5 risks, adds security context | 4 risks, severity-ranked |
| Implementation approach | 5 steps, general | 5 steps + doc recommendation | Sequenced dep chain, paste-ready |
| Overall quality | 3/5 | 4/5 | 5/5 |

---

## Takeaway

The free tier handles triage. You get the right answer on what is blocked and why. The paid tier justifies its cost when architectural nuance matters — catching the annotation semantics trap before it becomes a design mistake is worth the token spend. Copilot turns analysis into something you can act on immediately; when the output feeds a PR description or a team decision, the formatting and depth delta makes the tier difference obvious.
