---
title: "Triaging a Kubernetes Issue with Tier Routing"
slug: "issue-triage-kubernetes"
description: "Free model attempts it, paid model nails it — real input, real output from kubernetes/kubernetes#138431"
date: "2026-04-17"
---

## The Scenario

[kubernetes/kubernetes#138431](https://github.com/kubernetes/kubernetes/issues/138431) — **"kubepods.slice cgroups (memory and cpu) are reset every minute by kubelet"** — is a `priority/critical-urgent` regression filed on 2026-04-16 against the v1.36 development branch.

The chain: PR [#137719](https://github.com/kubernetes/kubernetes/pull/137719) removed the alpha `MemoryQoS` feature gate from `setMemoryQoS()` inside `UpdateCgroups`. That made kubelet unconditionally overwrite cgroup values under `/sys/fs/cgroup/kubepods.slice/` on every ~1-minute reconcile loop. PR [#137889](https://github.com/kubernetes/kubernetes/pull/137889) then removed the downstream reset logic as a side-effect fix — but left the ungated `setMemoryQoS()` call in place. Net effect: `cpu.max` and `memory.*` values set externally (by operators or node-level tuning tools) are silently clobbered on every reconcile, with no flag to disable it. SIG Node triaged it as critical in under 15 minutes.

Why it matters: production clusters with any external cgroup management on `kubepods.slice` are affected. There is no opt-out.

---

## The Workflow

A two-branch triage workflow. One branch runs the issue body through a local free model; the other sends it to a paid remote model. The `compare` step picks a winner; results are saved per model.

````lisp
(workflow "triage-k8s-issue"
  :description "Triage a GitHub issue with tier-routed LLMs"

  (step "fetch"
    (run "gh issue view ~param.issue --repo ~param.repo --json title,body,labels,comments"))

  (step "triage"
    (compare
      :objective "identify severity, root cause, affected components, and next steps"
      (branch "free"
        (llm :model "lmstudio:qwen3-8b"
          :prompt ```
          You are a senior Kubernetes maintainer. Triage this issue.

          ISSUE:
          ~(step fetch)

          Provide:
          1. Severity assessment with reasoning
          2. Root cause hypothesis
          3. Affected components and versions
          4. Recommended next steps for the maintainer team
          ```))
      (branch "paid"
        (llm :provider "openrouter" :model "qwen/qwen3.5-flash-02-23"
          :prompt ```
          You are a senior Kubernetes maintainer. Triage this issue.

          ISSUE:
          ~(step fetch)

          Provide:
          1. Severity assessment with reasoning
          2. Root cause hypothesis
          3. Affected components and versions
          4. Recommended next steps for the maintainer team
          ```))
      (review :criteria ("accuracy" "root-cause-depth" "actionability" "precision"))))

  (step "save"
    (save "results/triage-~param.issue.md" :from "triage")))
````

Run with:
<div class="model-output">

glitch workflow run triage-k8s-issue \
  --set issue=138431 \
  --set repo=kubernetes/kubernetes

</div>

---

## Free Tier: qwen3-8b (LM Studio)

<div class="model-output">

### 1. **Severity Assessment: Critical**
   - **Reasoning**: The issue is causing previously working environments to break by resetting cgroup values
     (e.g., `cpu.max`) that were explicitly set by external processes. This directly impacts critical
     Kubernetes workloads, especially in environments where custom cgroup configurations are used for
     resource management or isolation. The severity is amplified because the problem affects **kubelet's
     core cgroup reconciliation logic**, which is fundamental to container runtime operations. The label
     `priority/critical-urgent` and the comment about "breaking previously working environments" justify
     this classification.

---

### 2. **Root Cause Hypothesis**
   - **Unconditional Reconciliation of Memory Cgroups**: The removal of the feature gate for memory QoS
     in PR [#137889](https://github.com/kubernetes/kubernetes/pull/137889) caused kubelet to
     unconditionally reconcile memory cgroup values. This logic conflicts with external processes that
     manually set cgroup parameters (e.g., `cpu.max`) under `/sys/fs/cgroup/kubepods.slice/`, leading
     to unintended resets.
   - **Race Condition or Overriding Logic**: The prior gating in PR
     [#137719](https://github.com/kubernetes/kubernetes/pull/137719) (which was removed later) likely
     prevented this override by making memory QoS optional. The absence of this gate allows kubelet to
     overwrite external cgroup settings, causing instability.

---

### 3. **Affected Components and Versions**
   - **Component**: `kubelet` (specifically its cgroup reconciliation logic for memory and CPU).
   - **Versions**: Likely affects Kubernetes versions where PRs
     [#137719](https://github.com/kubernetes/kubernetes/pull/137719) and
     [#137889](https://github.com/kubernetes/kubernetes/pull/137889) were merged (e.g., v1.26+).
   - **Configurations**: Environments where external processes manage cgroup properties (e.g., custom
     kubelet configurations, third-party resource managers).

---

### 4. **Recommended Next Steps for the Maintainer Team**
   1. **Reintroduce Feature Gate for Memory QoS**:
      - Restore the alpha feature gate mechanism for memory QoS as in PR
        [#137719](https://github.com/kubernetes/kubernetes/pull/137719) to make this logic optional
        and avoid breaking existing workflows.
      - Reference: [Kubernetes cgroup reconciliation code](https://github.com/kubernetes/kubernetes/blob/main/pkg/kubelet/cgroup.go).

   2. **Audit Cgroup Reconciliation Logic**:
      - Investigate if other cgroup properties (e.g., `cpu.max`, `memory.soft_limit_in_bytes`) are
        also being reset unintentionally due to the same logic.
      - Ensure reconciliation only applies to memory-related settings and avoids overriding external
        configurations.

   3. **Add Safety Guards for External Cgroup Modifications**:
      - Implement checks in kubelet to detect conflicts between internal cgroup management and external
        modifications (e.g., log warnings or block writes if a cgroup parameter is set externally).

   4. **Rollout with Caution**:
      - If the fix requires rolling out changes, ensure it is done via a feature gate to avoid
        disrupting existing deployments.
      - Coordinate with SIG Node and release teams for a controlled rollout.

   5. **Documentation and Communication**:
      - Update documentation to clarify that memory QoS settings are gated by an alpha feature and
        may conflict with external cgroup configurations.
      - Add warnings in relevant code paths (e.g., when `cpu.max` is modified).

---

### Summary of Key References
- **PRs Involved**: [#137719](https://github.com/kubernetes/kubernetes/pull/137719) (original gating),
  [#137889](https://github.com/kubernetes/kubernetes/pull/137889) (removed gating).
- **Code Paths**: `pkg/kubelet/cgroup.go` (cgroup reconciliation logic).
- **Labels**: `priority/critical-urgent`, `sig/node`, `kind/bug`.

This issue requires immediate attention to prevent outages in environments relying on custom cgroup
configurations.

</div>

**Verdict.** The 8b model gets the severity right and correctly identifies the two-PR chain as the trigger. It recommends re-gating behind a feature gate — which is exactly what PR #138430 did. But the analysis has meaningful gaps: it mis-attributes the gate removal to #137889 rather than #137719, cites `pkg/kubelet/cgroup.go` (wrong file — the actual path is `pkg/kubelet/cm/qos_container_manager_linux.go`), speculates about a race condition that isn't present, and suggests adding "external cgroup modification guards" — a plausible-sounding idea that isn't grounded in the issue thread. Version range guess of "v1.26+" is wrong; this is a v1.36 development-branch regression. Usable as a first pass, not as a PR comment.

---

## Paid Tier: qwen/qwen3.5-flash-02-23

<div class="model-output">

### Triage Report: Kubepods.slice Cgroup Reset Issue

#### 1. Severity Assessment
**Severity: Critical (Urgent)**
*   **Level:** `priority/critical-urgent` (Confirmed by comment from `@SergeyKanzhelev`)
*   **Reasoning:**
    *   **Service Disruption:** The issue involves the Kubelet actively overwriting system-level cgroup
        configurations (`/sys/fs/cgroup/kubepods.slice/`) every minute.
    *   **Configuration Persistence Failure:** External tools or operators managing `cpu.max` or other
        QoS parameters at the `kubepods` level are being reverted to defaults (reset to `max`). This
        breaks resource isolation, performance tuning, and potentially SLA guarantees in production
        clusters where manual tuning or external management tools are used.
    *   **Scope:** Affects all nodes running the affected Kubelet version without requiring the user
        to enable any specific feature gates (was previously gated).

#### 2. Root Cause Hypothesis
The root cause is a regression in Kubelet's **Node QoS/Cgroup Reconciliation logic** following the
modification of the Memory QoS feature.

*   **Triggering Change:** **Pull Request #137719** removed the alpha feature gate behind setting
    Memory QoS. This allowed Kubelet to unconditionally reconcile memory cgroup values.
*   **Side Effect:** The reconciliation logic in `#137719` inadvertently touches broader cgroup paths
    (`kubepods.slice`) when attempting to manage memory constraints.
*   **Incomplete Fix:** **Pull Request #137889** attempted to mitigate the cgroup reset issue by
    removing specific logic, but as noted in the issue body ("It seems like this is causing unintended
    side effects with the same behavior here"), the core reconciliation still fires.
*   **Symptom Path:**
    1.  Kubelet starts Memory QoS loop.
    2.  Code attempts to set memory properties.
    3.  Due to shared cgroup handling or logic scope leakage, `cpu.max` under
        `/sys/fs/cgroup/kubepods.slice/` is overwritten with `max` (default/hog).
    4.  This cycle repeats every minute (Kubelet reconciliation interval).

**Evidence:** `@SergeyKanzhelev` confirmed: *"specifically, we see that some values like `cpu.max`
under `/sys/fs/cgroup/kubepods.slice/` may got reset back to max even if previously set to specific
value by outside process."`

#### 3. Affected Components and Versions
*   **Component:** `kubelet` (specifically the cgroup manager and QoS estimator).
*   **Affected Paths:**
    *   `/sys/fs/cgroup/kubepods.slice/` (Node-level controller cgroups).
    *   Memory QoS implementation details (referenced in PR `#137719`).
*   **Versions:**
    *   Affected branch: **Main/Development** (implied by the PR numbers and `2026` timestamp).
    *   **Backport Status:** Likely affects `v1.31` (if PRs were merged there) or `v1.32`. Immediate
        release team involvement is required.
*   **Feature Gates:** Previously gated by **MemoryQoS** (or similar internal memory QoS gate),
    which was removed.

#### 4. Recommended Next Steps for Maintainer Team

**Immediate Actions:**
1.  **Patch/Revert (Hotfix Priority):**
    *   The team led by `@liggitt` and `@SergeyKanzhelev` needs to review PR `#137719` again.
    *   **Action:** Revert the removal of the gating for Memory QoS settings. As noted by `@liggitt`,
        *"moving back into a gate-protected off-by-default block is a good idea."*
    *   **Alternative:** If Memory QoS is to remain enabled, the cgroup reconciliation code must be
        scoped strictly to pod-specific containers to avoid touching `kubepods.slice`.

2.  **Verification:**
    *   Verify the fix against a node where an operator script sets
        `echo 100m /sys/fs/cgroup/kubopods.slice/cpu.max`.
    *   Monitor that Kubelet does not overwrite this value over a 10-minute reconciliation interval.

3.  **Release Management:**
    *   `dims` has already CC'd the **release-team-leads**. Ensure a backport patch is prioritized
        for the current stable release branch if the bug is introduced in v1.31 or v1.32.
    *   Tag the issue with `area/bugfix` and update the changelog warning regarding cgroup
        modifications.

4.  **Communication:**
    *   Notify users of the affected version via the mailing list to mitigate risk (e.g., "Disable
        external cgroup tuning on kubepods.slice if you must upgrade").

</div>

**Verdict.** A meaningful step up. The paid model correctly pins #137719 as the triggering change, quotes @SergeyKanzhelev and @liggitt verbatim from the thread, and captures the "no feature gate = no opt-out" framing that makes this critical. The symptom path (memory QoS loop → `cpu.max` clobber → 1-minute repeat) is accurate. Version guesses (v1.31/v1.32) are still off — this is v1.36 — but it hedges with "Main/Development," which is correct. Still missing: the secondary regression introduced by the fix itself (PR #138439, stale `memory.min`/`memory.low` after the gate is re-enabled), exact source file, and the cgroupv2-only scoping. Actionable enough to hand to a maintainer; not complete enough to file a PR.

---

## Copilot (Sonnet)

<div class="model-output">

## Triage Report: kubernetes/kubernetes#138431

### 1. 🔴 Severity: **CRITICAL**

Classified `priority/critical-urgent` by SIG Node member @SergeyKanzhelev within ~14 minutes of
filing. Justified because:
- Affects **actively running production clusters** — kubelet's ~1-minute `UpdateCgroups` loop
  continuously overwrites cgroup values externally set on `kubepods.slice`
- Impacts `cpu.max` (CPU throttling limits) in addition to memory cgroup values — a regression
  beyond the memory-only scope of the original change
- **No opt-out** — behavior is unconditional; there is no flag or config to disable it

---

### 2. 🔍 Root Cause Hypothesis

A chain of two PRs introduced a logic regression:

| PR | Change | Effect |
|----|--------|--------|
| **#137719** (line R393–R395) | Removed the `MemoryQoS` alpha feature gate from the `setMemoryQoS()` call path inside `UpdateCgroups` | Memory QoS values are now unconditionally written on every kubelet reconcile loop (~1 min) |
| **#137889** | Removed cgroup reset logic that was resetting *other* pod cgroup properties as a side-effect of #137719 | Left the unconditional `setMemoryQoS()` call in place, but stripped the safety valves |

**Result:** `UpdateCgroups` in `pkg/kubelet/cm/qos_container_manager_linux.go` now calls
`setMemoryQoS()` unconditionally on cgroupv2 hosts. This clobbers **both** `memory.*` and `cpu.max`
values under `/sys/fs/cgroup/kubepods.slice/` — even values managed externally (e.g., by node-level
tuning tools or cluster operators).

A **secondary regression** was then introduced by the fix itself: PR **#138430** gated the entire
`setMemoryQoS` call behind `MemoryQoS` feature gate, but this made the zero-out branch at
`qos_container_manager_linux.go:L303–L308` unreachable. Nodes that previously had the alpha feature
enabled now **retain stale `memory.min`/`memory.low` values** after the gate is disabled, skewing
OOM killer behavior indefinitely. PR **#138439** (`kubelet: zero QoS-level memory.min / memory.low
when MemoryQoS is disabled`) proposes the correct fix: move the gate check *inside* `setMemoryQoS`
so only the populating branch is gated, keeping the zero-out branch always reachable.

---

### 3. 📦 Affected Components & Versions

| Component | Detail |
|-----------|--------|
| **Kubelet** | `pkg/kubelet/cm/qos_container_manager_linux.go` — `UpdateCgroups` + `setMemoryQoS` |
| **Cgroup subsystem** | cgroupv2 unified mode only (`/sys/fs/cgroup/kubepods.slice/`) |
| **Feature gate** | `MemoryQoS` (alpha) — incorrectly ungated by #137719 |
| **Milestone** | **v1.36** (actively targeted) |
| **Prior fix reference** | Related to #137886, #137726, #137674 |

Nodes running cgroup v1 are not affected.

---

### 4. ✅ Recommended Next Steps

**Immediate (done):**
- ✅ PR #138430 merged 2026-04-17 — re-gates `setMemoryQoS` behind `MemoryQoS` feature gate;
  issue #138431 closed

**Outstanding (action required):**

1. **Review & merge PR #138439** — Fixes the secondary regression introduced by #138430: stale
   `memory.min`/`memory.low` values persist when the feature gate is disabled. The fix moves the
   gate check inside `setMemoryQoS` rather than at the call site. Currently marked
   `do-not-merge/hold` and `needs-priority` — SIG Node should review and set priority.

2. **Assess backport candidates** — Determine if #138430 needs cherry-picking to active release
   branches (v1.33, v1.34, v1.35). The `cpu.max` reset behavior is user-visible and breaking.

3. **Validate no other `UpdateCgroups` call sites** have similar ungated cgroup writes introduced
   by #137719 or adjacent PRs in the same batch.

4. **Notify cluster operators** managing `kubepods.slice` externally — any node running the affected
   kubelet version between #137719 landing and #138430 merging will have had cgroup values silently
   reset. A `CHANGELOG` or advisory note for v1.36 release notes is warranted even though
   `release-note-none` was set on #138430.

5. **Consider a test** that verifies `cpu.max` is not clobbered by `UpdateCgroups` when `MemoryQoS`
   is disabled — closes the gap that allowed #137719 to regress undetected.

</div>

**Verdict.** The only output that names the exact source file (`pkg/kubelet/cm/qos_container_manager_linux.go`), the exact PR line reference (`R393–R395`), the cgroupv2-only scope, the correct milestone (v1.36), and — critically — the secondary regression introduced by the fix (PR #138439, stale `memory.min`/`memory.low` skewing OOM killer). It also correctly identifies that PR #138430 was already merged by the time the workflow ran and shifts focus to what remains open. No version speculation, no hallucinated code paths. The next-steps list is directly actionable without further research.

---

## Comparison

| Model | Severity Correct? | Root Cause Found? | Actionable Steps? | Cost |
|---|---|---|---|---|
| qwen3-8b (LM Studio, free) | ✅ Critical | ⚠️ Partial — wrong PR attribution, wrong file path | ⚠️ Directionally correct, contains hallucinations | $0.00 (local) |
| qwen/qwen3.5-flash-02-23 (OpenRouter) | ✅ Critical | ✅ Correct chain, missing secondary regression | ✅ Quotes thread accurately, version guess off | ~$0.001 |
| Copilot / Claude Sonnet 4.6 | ✅ Critical | ✅ Full chain + secondary regression in PR #138439 | ✅ Exact file, line, PR, milestone, open items | ~$0.01 |

---

## Takeaway

The free local model is not useless — it identified the correct severity and general fix direction at zero cost, which is enough to route the issue to the right SIG. The paid flash model added the thread context and correct PR chain, making its output paste-able into a reply. Copilot's output went further: it identified the secondary regression introduced by the fix itself — the kind of analysis that normally requires a maintainer to read three PRs and two code diffs. Tier routing in gl1tch lets you run the cheap model first and escalate only when the stakes justify it.