(use spork/test)

(array/push module/paths ["src/:all:.janet" :source])

(import glitch/batch :as b)

(start-suite "batch")

# parse-cross-review — numeric format
(def review-output `
VARIANT: local
plan_completeness: 9/10
code_quality: 8/10

VARIANT: claude
plan_completeness: 7/10
code_quality: 6/10

WINNER: local
`)

(def scores (b/parse-cross-review review-output))
(assert (= 2 (length scores)) "two variants parsed")
(def local (find |(= ($ :variant) "local") scores))
(assert (= 2 (local :passed)) "local passed 2")
(assert (= 2 (local :total)) "local total 2")
(assert (local :winner) "local is winner")

(def claude-score (find |(= ($ :variant) "claude") scores))
(assert (= 1 (claude-score :passed)) "claude passed 1 (7 passes, 6 fails)")
(assert (not (claude-score :winner)) "claude is not winner")

# parse-cross-review — pass/fail format
(def pf-output `
--- LOCAL ---
1. Specificity — PASS
2. Coverage — FAIL
SCORE: 1/2

--- CLAUDE ---
1. Specificity — PASS
2. Coverage — PASS
SCORE: 2/2

WINNER: CLAUDE
`)

(def pf-scores (b/parse-cross-review pf-output))
(assert (= 2 (length pf-scores)) "two variants in pass/fail")
(def pf-claude (find |(= ($ :variant) "claude") pf-scores))
(assert (pf-claude :winner) "claude wins in pass/fail format")
(assert (= 2 (pf-claude :passed)) "claude passed 2")

(def pf-local (find |(= ($ :variant) "local") pf-scores))
(assert (= 1 (pf-local :passed)) "local passed 1")
(assert (not (pf-local :winner)) "local is not winner")

(end-suite)
