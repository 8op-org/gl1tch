(use spork/test)

(array/push module/paths ["src/:all:.janet" :source])

(import glitch/runner :as r)
(import glitch/store :as s)
(import glitch/provider :as prov)
(import glitch/core :as g)

(start-suite "runner")

# Register mock providers so LLM calls work
(prov/reset!)
(prov/register "mock"
  (fn [opts]
    {:response (string "analyzed: " (opts :prompt))
     :tokens-in 5 :tokens-out 10 :latency 0 :cost 0}))
(prov/register "ollama"
  (fn [opts]
    {:response (string "ollama: " (opts :prompt))
     :tokens-in 5 :tokens-out 10 :latency 0 :cost 0}))

# Create test workflow files
(def wf-dir (string "/tmp/glitch-runner-test-" (os/time)))
(os/shell (string "mkdir -p " wf-dir))

# Simple workflow — just steps, no LLM
(spit (string wf-dir "/hello.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"hello\"\n"
    "  :description \"test\"\n"
    "  (g/step \"greet\" (string \"hello \" (g/input))))\n"))

# Workflow with LLM call
(spit (string wf-dir "/llm-wf.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"llm-test\"\n"
    "  :description \"test llm\"\n"
    "  (g/step \"analyze\" (g/llm :prompt (string \"check: \" (g/input)))))\n"))

# --- Run without store ---
(def result (r/run (string wf-dir "/hello.janet") "world"))
(assert (= "hello world" (get (result :steps) "greet"))
        "runner captures step output")
(assert (= 0 (result :run-id))
        "run-id is 0 when no db provided")

# --- Run with store ---
(def db-path (string "/tmp/glitch-runner-store-" (os/time) ".db"))
(def db (s/open db-path))
(def result2 (r/run (string wf-dir "/hello.janet") "db-test"
               :db db))
(assert (> (result2 :run-id) 0) "runner records run id")
(def stored-run (s/get-run db (result2 :run-id)))
(assert (= "hello db-test" (stored-run :output)) "store has correct output")
(assert (= 0 (stored-run :exit_status)) "store has exit status 0")

# --- Run with LLM ---
(def result3 (r/run (string wf-dir "/llm-wf.janet") "my-code"
               :db db))
(assert (> (result3 :run-id) 0) "llm workflow records run")
(def stored3 (s/get-run db (result3 :run-id)))
(assert (string/has-prefix? "ollama:" (stored3 :output))
        "llm workflow output recorded")

# --- Seed steps ---
(def result4 (r/run (string wf-dir "/hello.janet") "seed-test"
               :seed-steps @{"prior" "seeded-value"}))
(assert (= "seeded-value" (get (result4 :steps) "prior"))
        "seed steps appear in output")

# --- Custom params ---
(spit (string wf-dir "/params-wf.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"params-test\"\n"
    "  (g/step \"p\" (string \"model=\" (g/param :model))))\n"))
(def result5 (r/run (string wf-dir "/params-wf.janet") ""
               :params @{:model "gpt-4"}))
(assert (= "model=gpt-4" (get (result5 :steps) "p"))
        "custom params passed through")

# --- Error handling ---
(spit (string wf-dir "/bad.janet")
  (string
    "(error \"workflow-broke\")\n"))
(var caught-err nil)
(try
  (r/run (string wf-dir "/bad.janet") "")
  ([err] (set caught-err err)))
(assert (not (nil? caught-err)) "runner propagates errors")

# Error with db records failure
(var caught-err2 nil)
(try
  (r/run (string wf-dir "/bad.janet") "" :db db)
  ([err] (set caught-err2 err)))
(assert (not (nil? caught-err2)) "runner propagates db errors")
# Check that the failed run was recorded
(def runs (s/list-runs db))
(def failed-run (find |(= 1 ($ :exit_status)) runs))
(assert (not (nil? failed-run)) "failed run recorded in store")
(assert (string/has-prefix? "ERROR:" (failed-run :output))
        "failed run has error output")

# --- list-workflows ---
(def wfs (r/list-workflows wf-dir))
(assert (> (length wfs) 0) "list-workflows finds files")
(assert (all |($ :name) wfs) "each workflow has a name")
(assert (all |($ :path) wfs) "each workflow has a path")
# Verify sorted by name
(var prev "")
(each wf wfs
  (assert (>= (compare (wf :name) prev) 0)
          (string "workflows sorted: " prev " <= " (wf :name)))
  (set prev (wf :name)))

# --- list-workflows on nonexistent dir ---
(def empty-wfs (r/list-workflows "/tmp/no-such-dir-glitch-test"))
(assert (= 0 (length empty-wfs)) "list-workflows returns empty for missing dir")

# Cleanup
(s/close db)
(os/rm db-path)
(os/shell (string "rm -rf " wf-dir))

(end-suite)
