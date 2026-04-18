---
title: "Learning to Review: Three Passes on the Same PR"
slug: "review-learns-prometheus"
description: "glitch reviews prometheus/prometheus#18531 three times — cold, with feedback, then with full context. Watch the review quality improve."
date: "2026-04-18"
---

## The PR

[prometheus/prometheus#18531](https://github.com/prometheus/prometheus/pull/18531) fixes a timestamp mismatch bug in `smoothSeries()`. When the `@` modifier is used, the function was stamping output points with offset-adjusted timestamps (`evalTS - offMS`) instead of evaluator timestamps (`evalTS`). Since `gatherVector()` matches by exact timestamp equality against `evalTS`, those points were silently dropped — no error, no data, just binary operations returning nothing. The fix iterates over evaluator timestamps and derives data timestamps by subtracting the offset, so output aligns with what `gatherVector()` expects. Two files changed: 13 lines in `engine.go`, 23 new test lines in `extended_vectors.test`.

---

## Pass 1: Cold Review

Cold, no context, the review found the right root cause immediately. It correctly identified all three `FPoint`-appending branches, traced the `ts` vs `evalTS` mismatch, and called out that the data-window computation correctly continues using `dataTS`. It also flagged missing test cases for the interpolation and combined `@+offset` paths.

What it got right:

> "Data-window computation and interpolation continue to use `dataTS`, which is correct — the offset must apply to *where you look*, not *when you label the result*."

What it got wrong:

> "Combined `@` + explicit offset not tested... If `offMS` changes per evaluation pass, the loop `evalTS - offMS` may not stay pinned to the expected data timestamp."

The framing was off. The `offset` path wasn't an untested edge case — it was equally broken. The review also spent significant space on a pre-existing zero-step infinite loop and staleness window issues the maintainers would never touch.

---

## What the Maintainers Actually Said

Two LGTM approvals, one concrete suggestion. The key reviewer noted: "the offset calculation was also based on the wrong value, so suggest a couple of more tests." They proposed a `plain → smoothed → smoothed + 0` pattern for both the `@` and `offset` modifier paths, with plain reference lines before each smoothed variant as diagnostic anchors. A second reviewer confirmed the fix aligns with how `evalSeries` handles the same pattern — a consistency check that validates the approach without re-deriving it.

The maintainers didn't engage with the infinite loop, the staleness window, or the release note scope. Not because they missed them — because a PR review is about the PR.

---

## Pass 2: Informed Review

After seeing the maintainer feedback, the review corrected course. It recognized that `offset` was not undertested — it was broken in the same way as `@`, fixed by the same change, and the test gap was a real omission rather than a symmetry nicety. It adopted the `plain → smoothed → smoothed + 0` structure and explained *why* the offset block provides more branch coverage: `@ 100` always lands on exact 10s boundaries, so only the exact-match branch fires; `offset -100` with 15s steps hits interpolation.

> "The `offset -100` smoothed case (`10 11.5 13 14.5 16`) exercises the interpolation path, which the `@ 100` case does not... making it the more valuable block for branch coverage despite being the less obvious one."

It also explicitly dropped the noise: zero-step loop, staleness window, release note wording — flagged as pre-existing, not the PR's problem.

---

## Pass 3: Full Context

The final pass tightened further. Correctness analysis was the same, but the review added one observation neither the maintainers nor prior passes surfaced: that the branch-coverage asymmetry between the two test blocks means the `offset` block is doing more work than it looks like. It acknowledged the `evalTS`/`dataTS` rename as documentation of the invariant, not just a cosmetic change.

> "The rename from `ts` → `evalTS`/`dataTS` makes the invariant explicit: offset controls *where you look*, not *when you label the result*. This matches how `evalSeries` handles the same pattern, which is a strong consistency signal."

Verdict was LGTM. No noise, no pre-existing issue archaeology.

---

## Score Progression

| Criterion | Pass 1 | Pass 2 | Pass 3 |
|---|:---:|:---:|:---:|
| Correctness of analysis | 8 | 9 | 9 |
| Edge cases identified | 7 | 6 | 8 |
| Actionable feedback quality | 6 | 8 | 9 |
| Alignment with maintainer perspective | 4 | 7 | 9 |

Pass 1 scores high on correctness because the root cause analysis was right. It scores low on alignment because the priorities were wrong — it raised four issues the maintainers ignored and framed the most important gap (offset path equally broken) as a minor test hole. Pass 2 drops slightly on edge cases because it deliberately deprioritized valid observations. That's the correct call given context, but it reflects a real tradeoff: signal vs. completeness. Pass 3 recovers by adding a genuine new observation — the branch-coverage asymmetry — without reopening closed questions.

---

## Takeaway

A review's value is not proportional to the number of issues raised. The cold review was technically accurate and still misaligned — it understood the code and misread the context. Iterative review with accumulated feedback doesn't just improve coverage; it changes what the reviewer thinks is worth saying. That's the harder skill, and it's the one that feedback loops develop. glitch shortens the loop.