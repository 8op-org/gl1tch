(use spork/test)

# Add src to module path so we can import glitch/store
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/store :as s)

(start-suite "store")

(def db-path (string "/tmp/glitch-test-" (os/time) ".db"))
(def db (s/open db-path))

# record a run
(def run-id (s/record-run db
  {:name "test-wf" :input "hello"
   :workflow-file "test.janet"
   :model "qwen2.5:7b"}))
(assert (> run-id 0) "record-run returns positive id")

# record steps
(s/record-step db
  {:run-id run-id :step-id "s1" :output "step one"
   :kind "step" :duration 100})
(s/record-step db
  {:run-id run-id :step-id "s2" :output "step two"
   :kind "llm" :duration 500 :tokens-in 10 :tokens-out 50})

# finish run
(s/finish-run db run-id "step two" 0
  {:tokens-in 10 :tokens-out 50 :cost 0.001})

# query run
(def run (s/get-run db run-id))
(assert (= "test-wf" (run :name)) "get-run returns name")
(assert (= "step two" (run :output)) "get-run returns output")

# query steps
(def steps (s/get-steps db run-id))
(assert (= 2 (length steps)) "get-steps returns 2 steps")

# list runs
(def runs (s/list-runs db))
(assert (>= (length runs) 1) "list-runs returns at least 1")

# cleanup
(s/close db)
(os/rm db-path)

(end-suite)
