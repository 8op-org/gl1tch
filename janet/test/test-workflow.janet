(use spork/test)

# Add src to module path so we can import glitch/core
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/core :as g)

(start-suite "workflow")

# workflow macro captures name and runs body
(g/reset-steps!)
(g/set-input! "test-input")
(def result
  (g/workflow "test-wf"
    :description "A test workflow"
    (g/step "a" (string "got:" (g/input)))
    (g/step "b" (string (g/ref "a") ":done"))))
(assert (= "got:test-input:done" (result :output))
        "workflow returns last step output")
(assert (= "test-wf" (result :name))
        "workflow captures name")

# par runs steps concurrently
(g/reset-steps!)
(def results
  (g/par
    (g/step "p1" (do (ev/sleep 0.01) "one"))
    (g/step "p2" (do (ev/sleep 0.01) "two"))))
(assert (= "one" (g/ref "p1")) "par step p1")
(assert (= "two" (g/ref "p2")) "par step p2")

# retry retries on failure
(var attempts 0)
(def val (g/retry 3
  (do
    (++ attempts)
    (if (< attempts 3) (error "not yet") "ok"))))
(assert (= val "ok") "retry succeeds on 3rd attempt")
(assert (= attempts 3) "retry tried 3 times")

# gate checks a predicate
(g/reset-steps!)
(g/step "data" "good")
(assert (g/gate "quality" (not (nil? (g/ref "data"))))
        "gate passes when predicate is true")

(end-suite)
