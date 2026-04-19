(use spork/test)

(array/push module/paths ["src/:all:.janet" :source])

(import glitch/core :as g)
(import glitch/runner :as runner)
(import glitch/store :as s)
(import glitch/provider :as prov)

(start-suite "integration")

# Setup
(def test-dir (string "/tmp/glitch-integration-" (os/time)))
(os/shell (string "mkdir -p " test-dir))
(def db-path (string test-dir "/test.db"))
(def db (s/open db-path))

# Register a mock provider
(prov/reset!)
(prov/register "mock"
  (fn [opts]
    {:response (string "analyzed: " (opts :prompt))
     :tokens-in 5 :tokens-out 10 :latency 0 :cost 0}))
(prov/register "ollama"
  (fn [opts]
    {:response (string "ollama: " (opts :prompt))
     :tokens-in 5 :tokens-out 10 :latency 0 :cost 0}))

# --- Test 1: Simple workflow with shell steps ---
(spit (string test-dir "/simple.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"simple\"\n"
    "  (g/step \"data\" (g/sh \"echo\" \"hello world\"))\n"
    "  (g/step \"result\" (string \"got: \" (string/trim (g/ref \"data\")))))\n"))

(def r1 (runner/run (string test-dir "/simple.janet") "test-input" :db db))
(assert (string/find "got: hello world" (r1 :output))
        "simple workflow runs shell and chains steps")
(assert (> (r1 :run-id) 0) "run recorded in store")

(def stored (s/get-run db (r1 :run-id)))
(assert (= 0 (stored :exit_status)) "run completed successfully")

# --- Test 2: Workflow with LLM call ---
(spit (string test-dir "/llm-test.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"llm-test\"\n"
    "  (g/step \"analysis\"\n"
    "    (g/llm :prompt \"analyze this\" :provider \"mock\"))\n"
    "  (g/step \"summary\" (string \"Summary: \" (g/ref \"analysis\"))))\n"))

(def r2 (runner/run (string test-dir "/llm-test.janet") "" :db db))
(assert (string/find "analyzed:" (r2 :output))
        "LLM workflow dispatches to mock provider")

# --- Test 3: Par execution ---
(spit (string test-dir "/par-test.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"par-test\"\n"
    "  (g/par\n"
    "    (g/step \"a\" (g/sh \"echo\" \"alpha\"))\n"
    "    (g/step \"b\" (g/sh \"echo\" \"beta\")))\n"
    "  (g/step \"combined\"\n"
    "    (string (string/trim (g/ref \"a\")) \"+\" (string/trim (g/ref \"b\")))))\n"))

(def r3 (runner/run (string test-dir "/par-test.janet") ""))
(assert (string/find "alpha" (r3 :output))
        "par step a ran")
(assert (string/find "beta" (get (r3 :steps) "b"))
        "par step b accessible")

# --- Test 4: Error handling ---
(spit (string test-dir "/error-test.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"error-test\"\n"
    "  (g/step \"fail\" (error \"intentional error\")))\n"))

(var caught false)
(try
  (runner/run (string test-dir "/error-test.janet") "" :db db)
  ([err] (set caught true)))
(assert caught "runner propagates errors")

# --- Test 5: Params ---
(spit (string test-dir "/params-test.janet")
  (string
    "(array/push module/paths [\"src/:all:.janet\" :source])\n"
    "(import glitch/core :as g)\n"
    "(g/workflow \"params-test\"\n"
    "  (g/step \"out\" (string (g/param \"name\") \":\" (g/input))))\n"))

(def r5 (runner/run (string test-dir "/params-test.janet") "world"
          :params @{"name" "hello"}))
(assert (= "hello:world" (r5 :output))
        "params and input passed through")

# --- Test 6: Store records steps ---
(def steps (s/get-steps db (r1 :run-id)))
(assert (>= (length steps) 1) "steps recorded in store")

# Cleanup
(s/close db)
(os/shell (string "rm -rf " test-dir))

(end-suite)
