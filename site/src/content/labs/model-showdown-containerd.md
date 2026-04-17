---
title: "Model Showdown: Free vs Paid vs Copilot"
slug: "model-showdown-containerd"
description: "Same containerd feature request through three tiers — see where free falls short and paid shines"
date: "2026-04-17"
---

## The Scenario

[containerd/containerd#11339](https://github.com/containerd/containerd/issues/11339) proposes a new `UpdatePodSandboxResources` CRI method. With InPlacePodVerticalScaling (KEP-1287), pod sandbox resources are now mutable at runtime — but NRI plugins have no visibility into those changes. A pod gets resized; the plugins keep operating on the original numbers.

The design is narrow by intent: an informational-only CRI call. No cgroup mutations. Kubelet stays in charge of the cgroup writes; the runtime forwards the update envelope to NRI. Implementation PR [#11406](https://github.com/containerd/containerd/pull/11406) is code-complete. The only blocker is external — Kubernetes 1.33 (~April 23) must ship before the `cri-api` module can be tagged and the containerd dependency bump can land.

## The Workflow

One shell fetch step pulls the full issue JSON including comment thread. Three chained LLM steps each isolate a concern. Swapping `(def provider ...)` and `(def model ...)` is the only change across tier variants.

````clojure
(def model "qwen3:8b")
(def provider "ollama")

(workflow "analyze-cri-issue"
  :description "Fetch and analyze a CRI issue: problem, deps, complexity, risks, approach"

  (step "fetch-issue"
    (run "gh issue view 11339 --repo containerd/containerd --json title,body,comments,labels | jq '.'"))

  (step "analyze-problem"
    (llm :provider provider :model model
      :prompt ```
        You are a senior infrastructure engineer reviewing a container runtime proposal.

        Issue data:
        ~(step fetch-issue)

        Answer: (1) core problem and proposed solution, (2) cross-project dependencies.
        ```))

  (step "assess-complexity"
    (llm :provider provider :model model
      :prompt ```
        Prior analysis:
        ~(step analyze-problem)

        Full thread:
        ~(step fetch-issue)

        Answer: (3) implementation complexity with justification, (4) key risks from the thread.
        ```))

  (step "suggest-approach"
    (llm :provider provider :model model
      :prompt ```
        Analysis so far:
        ~(step analyze-problem)
        ~(step assess-complexity)

        Answer: (5) concrete sequenced implementation approach for containerd maintainers.
        ```)))
````

## Free Tier: qwen3-8b (LM Studio)

````
### 1. **Problem & Proposed Solution**
- **Problem**: Pod sandbox cgroups are managed by Kubelet, but mutable resources (via InPlacePodVerticalScaling) require NRI plugins to access them dynamically. Existing CRI calls don't pass this mutable resource info to NRI plugins.
- **Solution**: Introduce a new `UpdatePodSandboxResources` informational CRI call to propagate pod sandbox resource updates to NRI plugins, enabling features like in-place resizing and service mesh integration. This requires aligning containerd with Kubernetes (CRI API) and NRI plugin changes.

---

### 2. **Cross-Project Dependencies**
- **Kubernetes**: Must implement the `UpdatePodSandboxResources` CRI method via [KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources#cri-changes) and [PR](https://github.com/kubernetes/kubernetes/pull/128123).
- **NRI (Node Resource Interface)**: Requires updates to NRI plugins to consume the new CRI method. Related changes in [containerd/nri#141](https://github.com/containerd/nri/pull/141).
- **cri-o**: Mentioned as a potential runtime dependency (via [issue comment](https://github.com/containerd/containerd/issues/11339#issuecomment-2637908647)), though not directly required.
- **CRI API Version Skew**: Must coordinate with Kubernetes to align CRI API versions, as per [Kubernetes' policy](https://www.kubernetes.dev/docs/code/cri-api-version-skew-policy/#feature-development).

---

### 3. **Implementation Complexity**
- **Assessment**: **Medium**
  - Adding a new CRI method involves modifying containerd's codebase to handle the `UpdatePodSandboxResources` call and pass data to NRI plugins.
  - Integration with Kubernetes and NRI requires coordination, especially ensuring compatibility with CRI API versioning.
  - Existing workarounds (e.g., using annotations) may reduce complexity but could leave gaps for advanced use cases.

---

### 4. **Risks & Concerns**
- **Dependency Delays**: The containerd PR is blocked until Kubernetes cuts the CRI API version (April 23rd), which could delay implementation.
- **NRI Plugin Support**: If NRI plugins are not updated to handle this new method, the feature may lack utility.
- **Incomplete Annotations/Labels**: Comments suggest the current solution may not fully address pod label/annotation updates for service mesh or network plugins (e.g., [discussion](https://github.com/containerd/containerd/issues/11339#issuecomment-2645906997)).
- **Version Skew Risks**: Misalignment between CRI API versions across Kubernetes, containerd, and NRI could cause compatibility issues.

---

### 5. **Implementation Approach**
1. **Coordinate with Kubernetes**: Ensure the `UpdatePodSandboxResources` CRI method is implemented in [Kubernetes' CRI PR](https://github.com/kubernetes/kubernetes/pull/128123) and aligned with containerd.
2. **Extend containerd**: Add support for the new CRI call, ensuring it forwards pod sandbox resource data to NRI plugins via the `annotations` field in the request.
3. **Update NRI Plugins**: Collaborate with NRI maintainers (e.g., [containerd/nri#141](https://github.com/containerd/nri/pull/141)) to ensure plugins consume this data.
4. **Test & Validate**: Verify compatibility across Kubernetes, containerd, and NRI plugins. Address gaps in label/annotation handling if needed (e.g., by expanding the `annotations` field).
5. **Monitor CRI API Alignment**: Track Kubernetes' release schedule for CRI API version cuts to unblock containerd's implementation.

**Next Steps**: Prioritize resolving dependency blockers, validate NRI plugin support, and refine the annotations/labels handling if required.
````

**Verdict:** qwen3-8b produced a structurally complete response — all five sections, correct dependency graph, right complexity call. It misses what's inside the thread. The `klihub`/`chrishenzie` annotation semantics debate — the part that could cause a real design mistake — isn't mentioned. Risk enumeration is flat: four items, no severity weighting, no differentiation between hard blockers and long-tail concerns. Good enough to triage. Not enough to make an architectural call.

## Paid Tier: qwen/qwen3.5-flash-02-23

````
### 1. Problem and Proposed Solution

**Problem:**
Pod sandbox cgroups are currently managed exclusively by the Kubelet. However, with the introduction of **InPlacePodVerticalScaling (IPVS)** in Kubernetes, sandbox resources (CPU/memory) have become mutable. The current container runtime interface only allows resource definition during `RunPodSandbox`. NRI (Node Resource Interface) plugins (used for network, storage, and service mesh logic) lose visibility into sandbox resource changes, potentially causing them to operate on stale resource assumptions or fail to resize their own cgroups/containers accordingly.

**Proposed Solution:**
Implement a new CRI gRPC call: `UpdatePodSandboxResources`.
*   **Function:** Passively notify the runtime (containerd) when a sandbox's resource limits change.
*   **Primary Goal:** Bridge the information to NRI plugins so they can reconcile their own sandbox resources without restarting the pod.
*   **Mechanism:** Containerd will expose this CRI API and forward the resource update information to the NRI hook system.

### 2. Cross-Project Dependencies

This feature is highly dependent on synchronization across the Kubernetes ecosystem:

*   **Kubernetes (Core):**
    *   **CRI API Change:** Requires the `UpdatePodSandboxResourcesRequest` proto definition (referenced in K8s PR #128123).
    *   **Kubelet Logic:** Kubelet must be enabled to invoke this CRI call during IPVS operations.
    *   **Release Schedule:** The implementation is currently blocked pending a specific K8s release containing the CRI API version (identified as v1.33/April 23rd in comments).
*   **NRI (Node Resource Interface):**
    *   Containerd's `nri-framework` must have matching capability PR #141 to handle the new `UpdatePodSandboxResources` event.
*   **containerd:**
    *   Implementation in the CRI shim and integration with the runtime plugin system.
*   **Ecosystem Parity:**
    *   **CRI-O:** Identified in comments as having a corresponding dependency issue (#8990), implying ecosystem fragmentation risks if other runtimes do not adopt the interface immediately.

### 3. Implementation Complexity: **Medium**

While the containerd code change described as "complete" (PR #11406) is described as a pass-through wrapper, the overall complexity is Medium due to integration factors:

*   **Code Logic:** Low/Minimal. It involves mapping CRI requests to internal runtime structures and forwarding to NRI.
*   **Version Skew & Compatibility:** High. Must handle versions of K8s that do not support the new RPC gracefully (feature detection or blocking).
*   **API Semantics:** Medium complexity in ensuring definitions match exactly between K8s and containerd to avoid protocol errors during upgrade paths.
*   **Blocking Factors:** Dependency on upstream K8s release timelines creates a timeline constraint rather than a coding constraint.

### 4. Risks and Concerns

1.  **Annotation/Label Sync Gap (Critical):**
    *   **Issue:** User `MikeZappa87` raised a concern regarding syncing **Pod Labels** (not just resources/annotations in the request metadata).
    *   **Conflict:** User `klihub` warned that the CRI API `annotations` field is intended for *update metadata*, not for syncing the actual Pod Labels/Annotations.
    *   **Impact:** Service mesh and network plugins relying on dynamic pod labels for security policies may not be updated via this RPC. This represents a potential architectural shortfall requiring a separate KEP or API extension (e.g., a dedicated `SyncPodLabels` or inclusion of `PodSpec` in the update).
2.  **Release Coordination:**
    *   Containerd code is ready, but deployment is blocked waiting on K8s 1.33. This delays production value. If K8s ships without the CRI change (e.g., beta delays), containerd users get a version with the call but no K8s client usage.
3.  **Runtime Security Context:**
    *   Allowing the Kubelet to push state changes to the sandbox requires ensuring NRI plugins are sandbox-aware regarding resource changes (e.g., preventing privilege escalation during resize).
4.  **Backward Compatibility:**
    *   Older K8s versions will treat `UpdatePodSandboxResources` as unsupported. If a cluster mixes runtimes, some may support IPVS resources while others ignore the resize.

### 5. Suggested Implementation Approach

1.  **Merge Containerd Changes:**
    *   Proceed with merging PR #11406 immediately. Delaying the runtime code is unnecessary and blocks testing.
2.  **Track External Blockers:**
    *   Establish a tracking milestone for the CRI API cut in K8s (v1.33). Communicate the cut date to containerd maintainers to ensure feature gating works correctly upon merge.
3.  **Address Label Sync Separately:**
    *   Create an internal follow-up issue or KEP discussion regarding `MikeZappa87`'s label sync requirement.
    *   *Recommendation:* Do not repurpose the CRI `annotations` field for this, as `klihub` correctly identified the semantics are for "update options." Investigate if K8s can extend the `UpdatePodSandboxResources` request to include `PodSpec.Revision` or if a new `SyncPodResources` RPC is needed in a future CRI version.
4.  **NRI Integration Testing:**
    *   Before merge, ensure the NRI containerd integration actually *calls* the NRI plugins with the new update parameters. A PR exists (#141), but verify it handles the specific "informational" requirement (i.e., does not fail if the runtime cannot physically resize, but should notify the plugin).
5.  **Documentation Update:**
    *   Update the containerd CRI documentation to explicitly mark `UpdatePodSandboxResources` as "Feature Gate: in-place-pod-vertical-scaling". This prevents confusion when the call appears in logs but is not active.
````

**Verdict:** The paid tier earns its cost on one specific call: the annotation field semantics. It names the conflict — who said what, why it matters, what the wrong move looks like — with enough precision that a maintainer could act on it. It also surfaces a risk the free model skipped entirely: security context around privilege escalation during resize. The `kubernetes/cri-api` module appears as a distinct dependency, not conflated with `kubernetes/kubernetes`. Implementation steps include a concrete doc-comment recommendation. This is the output you put in front of a team deciding whether to merge.

## Copilot (Sonnet)

````
## Analysis: `UpdatePodSandboxResources` CRI Method (containerd#11339)

---

### 1. Problem & Proposed Solution

**Problem:** `InPlacePodVerticalScaling` (pod resize, KEP-1287) makes pod sandbox resources mutable at runtime. The Kubelet manages sandbox cgroups directly, but NRI plugins have no way to observe these resource changes — they're left in the dark when a pod is resized.

**Solution:** Add a new **informational-only** `UpdatePodSandboxResources` CRI method. It doesn't mutate state in the runtime; it's a pass-through to notify NRI plugins of the updated resource envelope so they can react accordingly.

---

### 2. Cross-Project Dependencies

| Project | Dependency | Status |
|---------|-----------|--------|
| **kubernetes/kubernetes** | CRI API PR [#128123](https://github.com/kubernetes/kubernetes/pull/128123) — defines the new RPC | Pending (blocked on k8s 1.33, ~April 23) |
| **kubernetes/cri-api** | New `UpdatePodSandboxResourcesRequest` proto type must be cut as a versioned release | Blocked on k8s 1.33 cut |
| **containerd/nri** | NRI API changes [nri#141](https://github.com/containerd/nri/pull/141) — NRI needs new hook/event type | In progress |
| **cri-o** | Parallel implementation tracked in [cri-o#8990](https://github.com/cri-o/cri-o/issues/8990) | Parallel effort |
| **containerd/containerd** | Implementation PR [#11406](https://github.com/containerd/containerd/pull/11406) | Code complete, blocked on deps |

Per the [CRI API version skew policy](https://www.kubernetes.dev/docs/code/cri-api-version-skew-policy/#feature-development), containerd and Kubernetes drafts must exist simultaneously before the CRI API merges — this constraint is being followed correctly.

---

### 3. Implementation Complexity: **Medium**

- The CRI handler itself is **low complexity** — it's a new RPC endpoint that does minimal work (no cgroup mutations)
- NRI event dispatch is **low-to-medium** — requires new hook type in the NRI protocol
- The **dependency chain** (k8s -> cri-api -> nri -> containerd) is the primary complexity driver, not the code itself
- PR #11406 is already code complete, confirming this assessment

---

### 4. Risks & Concerns

**[BLOCKER] Release coupling risk**
The feature is fully blocked until Kubernetes 1.33 ships (~April 23). A slip in the k8s release directly delays containerd.

**[WATCH] Annotations semantics ambiguity**
The `map<string,string> annotations` field in `UpdatePodSandboxResourcesRequest` is ambiguous. The thread shows disagreement:
- `chrishenzie` suggests it could carry pod labels/annotations for NRI
- `klihub` correctly points out these *annotate the update operation itself*, not the pod
- This scope confusion could lead to misuse or the field being overloaded

**[WATCH] Pod labels/annotations not included**
`MikeZappa87` raised that pod labels/annotations aren't propagated in the update, which limits NRI's ability to support service mesh and network plugins that depend on up-to-date pod metadata. This is a **gap in the KEP design** that can't be fixed in containerd alone — it would require upstream CRI API changes.

**[SOUND] Informational-only design is correct**
Making this a no-op at the runtime level (cgroups are Kubelet-managed) avoids double-management bugs and keeps the runtime's responsibility clear.

---

### 5. Suggested Implementation Approach

The implementation path is well-scoped. Here's the recommended sequencing:

<div class="model-output">

1. [DONE]    containerd PR #11406 — CRI handler skeleton + NRI dispatch
2. [DONE]    NRI PR #141          — NRI protocol UpdatePodSandbox hook
3. [WAITING] k8s 1.33 cut (~Apr 23) -> cri-api module tagged
4. [THEN]    Bump cri-api dependency in containerd, unblock merge of #11406
5. [THEN]    Integration test: resize a pod, verify NRI plugin receives the event

</div>

**For the annotations ambiguity**, I'd recommend adding a doc comment clarifying that the `annotations` field in `UpdatePodSandboxResourcesRequest` annotates the *update event* — and file a separate upstream issue/KEP addendum if propagating live pod labels to the runtime is a genuine requirement (it is for service mesh use cases, but that's a broader CRI design conversation).

---

**Bottom line:** Solid, well-scoped feature. The code is done; the only blocker is the dependency release train. Low execution risk once k8s 1.33 lands. The pod labels/annotations gap is worth tracking as a follow-on issue.
````

**Verdict:** The most immediately actionable output. The dependency table includes current status — you don't have to decode the thread to know what's blocked. Risk items carry severity markers, so triage is instant. The implementation sequence is a numbered dependency chain ready to paste into a PR description or team standup. It validates the informational-only architecture rather than just summarizing it — that's a different kind of analysis. The annotation ambiguity gets the clearest framing of any tier: it annotates the update event, not the pod.

## Comparison

| Metric | Free | Paid | Copilot |
|--------|------|------|---------|
| Problem summary | Correct, no thread nuance | Correct, thread-aware | Correct, architecture-validated |
| Cross-project deps identified | 4/5 — cri-api implicit | 5/5 — cri-api explicit, release schedule noted | 5/5 — dep table with current status |
| Complexity assessment | Medium, unqualified | Medium — code low, integration high | Medium — dep chain is the primary driver |
| Risk identification | 4 risks, no severity | 5 risks, adds security context | 4 risks, severity-ranked |
| Implementation approach | 5 steps, general | 5 steps + doc recommendation | Sequenced dep chain, paste-ready |
| Overall quality | 3/5 | 4/5 | 5/5 |

## Takeaway

The free tier handles triage — you get the right answer on what's blocked and why. The paid tier justifies its cost when architectural nuance is on the line; catching the annotation semantics trap before it becomes a design mistake is worth the token spend. Copilot (Sonnet) turns the analysis into something you can act on immediately — for issues where the output feeds a PR description or a team decision, the formatting and depth delta makes the tier difference obvious.