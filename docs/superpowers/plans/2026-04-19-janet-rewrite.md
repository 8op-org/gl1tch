# glitch-on-Janet Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite glitch from Go to Janet — a single-language Lisp runtime where workflows, providers, plugins, workspace config, CLI, and GUI are all Janet programs.

**Architecture:** Janet application compiled to a standalone binary via `jpm`. The `glitch` module provides macros (`workflow`, `step`, `par`) and functions (`sh`, `llm`, `ref`, `save`) that wrap Janet primitives. Providers are Janet functions. The store uses `janet-lang/sqlite3`. The GUI server uses `spork/httpf`. No custom evaluator — Janet IS the evaluator.

**Tech Stack:** Janet 1.41+, spork (HTTP server/client, JSON, argparse), janet-lang/sqlite3, jpm (build/package)

**Distribution:** Users install `glitch` via Homebrew or download a binary. They never install Janet. The `jpm build` / `declare-executable` process bakes the Janet runtime into a standalone binary (~1MB). Janet is an implementation detail — invisible to users. Workflow files use `.janet` extension but users don't need to know or care that it's Janet under the hood.

**Spec:** `docs/superpowers/specs/2026-04-19-janet-rewrite-design.md`

**Go reference codebase:** The existing Go code under `internal/` and `cmd/` — read for porting accuracy, do not modify.

---

### Task 1: Project Scaffold + Build

**Files:**
- Create: `janet/project.janet`
- Create: `janet/src/glitch/main.janet`
- Create: `janet/test/test-smoke.janet`

This task creates the Janet project skeleton, installs dependencies, and verifies the build produces a working binary.

- [ ] **Step 1: Create project directory**

```bash
mkdir -p janet/src/glitch janet/test janet/providers janet/stdlib
```

- [ ] **Step 2: Write project.janet**

Create `janet/project.janet`:

```janet
(declare-project
  :name "glitch"
  :description "Workflow engine"
  :dependencies
    ["https://github.com/janet-lang/spork.git"
     "https://github.com/janet-lang/sqlite3.git"])

(declare-executable
  :name "glitch"
  :entry "src/glitch/main.janet"
  :install true)
```

- [ ] **Step 3: Write minimal main.janet**

Create `janet/src/glitch/main.janet`:

```janet
(defn main [& argv]
  (if (= (length argv) 0)
    (do
      (print "glitch - workflow engine")
      (print "usage: glitch <command> [args...]")
      (print "commands: version")
      (os/exit 1))
    (match (first argv)
      "version" (print "glitch 0.1.0-janet")
      (do (eprintf "unknown command: %s" (first argv))
          (os/exit 1)))))
```

- [ ] **Step 4: Write smoke test**

Create `janet/test/test-smoke.janet`:

```janet
(use spork/test)

(start-suite "smoke")

(assert (= 1 1) "janet runs")

(end-suite)
```

- [ ] **Step 5: Install dependencies and build**

```bash
cd janet && jpm deps && jpm build
```

Expected: binary at `janet/build/glitch`

- [ ] **Step 6: Verify binary runs**

```bash
./janet/build/glitch version
```

Expected output: `glitch 0.1.0-janet`

- [ ] **Step 7: Run tests**

```bash
cd janet && jpm test
```

Expected: `smoke` suite passes.

- [ ] **Step 8: Commit**

```bash
git add janet/
git commit -m "feat(janet): project scaffold with build and smoke test"
```

---

### Task 2: Core Module — step, ref, sh, save

**Files:**
- Create: `janet/src/glitch/core.janet`
- Create: `janet/test/test-core.janet`

The core module provides the fundamental workflow primitives. This is the heart of the rewrite — these macros/functions replace the entire Go evaluator.

**Reference:** `internal/pipeline/eval.go` (special forms), `internal/pipeline/eval_builtins.go` (builtins `sh`, `ref`, `save`, `read-file`, `write-file`)

- [ ] **Step 1: Write failing tests for core primitives**

Create `janet/test/test-core.janet`:

```janet
(use spork/test)
(import glitch/core :as g)

(start-suite "core")

# step stores output and returns it
(g/reset-steps!)
(assert (= "hello" (g/step "greet" "hello"))
        "step returns its body value")
(assert (= "hello" (g/ref "greet"))
        "ref retrieves step output")

# ref for missing step returns nil
(g/reset-steps!)
(assert (nil? (g/ref "missing"))
        "ref returns nil for unknown step")

# sh runs a command and returns stdout
(def result (g/sh "echo" "hi"))
(assert (= "hi" (string/trim result))
        "sh captures stdout")

# sh with failing command raises error
(assert-error "sh fails on bad command"
  (g/sh "false"))

# save writes file, read-file reads it
(def tmp (string "/tmp/glitch-test-" (os/time)))
(g/save tmp "test content")
(assert (= "test content" (g/read-file tmp))
        "save + read-file roundtrip")
(os/rm tmp)

# input returns the current workflow input
(g/set-input! "my input")
(assert (= "my input" (g/input))
        "input returns current input")

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

Expected: FAIL — `glitch/core` module not found.

- [ ] **Step 3: Implement core.janet**

Create `janet/src/glitch/core.janet`:

```janet
# Core workflow primitives.
# Replaces: internal/pipeline/eval.go + eval_builtins.go

# --- Workflow state (fiber-local via dynamic bindings) ---

(var- *steps* @{})
(var- *input* "")
(var- *params* @{})
(var- *step-recorder* nil)

(defn reset-steps! []
  (set *steps* @{}))

(defn set-input! [s]
  (set *input* s))

(defn set-params! [p]
  (set *params* p))

(defn set-step-recorder! [f]
  (set *step-recorder* f))

# --- Primitives ---

(defn input []
  "Return the current workflow input string."
  *input*)

(defn params []
  "Return the current workflow params table."
  *params*)

(defn param [key &opt default]
  "Return a single param by keyword or string key."
  (or (get *params* key)
      (get *params* (string key))
      default))

(defn ref [step-id]
  "Look up a step's output by ID. Returns nil if not found."
  (get *steps* step-id))

(defn step [id body]
  "Record a step output. Returns the body value (converted to string)."
  (def val (if (string? body) body (string body)))
  (put *steps* id val)
  (when *step-recorder*
    (*step-recorder* {:step-id id :output val :kind "step"}))
  val)

(defn sh [& args]
  "Execute a shell command. Returns stdout as string. Raises on non-zero exit."
  (def proc (os/spawn args :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "sh: command %s failed (exit %d): %s"
            (string/join args " ") exit (string err-out)))
  (string out))

(defn save [path content]
  "Write content to a file. Creates parent directories."
  (def dir (string/slice path 0
             (or (string/find-last "/" path) 0)))
  (when (and (not= dir "") (not (os/stat dir)))
    (os/shell (string "mkdir -p " dir)))
  (spit path (if (string? content) content (string content)))
  content)

(defn read-file [path]
  "Read a file and return its contents as a string."
  (slurp path))

(defn write-file [path content]
  "Alias for save."
  (save path content))

(defn get-steps []
  "Return a snapshot of all step outputs."
  (table/clone *steps*))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `core` suite passes.

- [ ] **Step 5: Commit**

```bash
git add janet/src/glitch/core.janet janet/test/test-core.janet
git commit -m "feat(janet): core module — step, ref, sh, save, read-file"
```

---

### Task 3: Core Module — workflow, par, threading, control flow

**Files:**
- Modify: `janet/src/glitch/core.janet`
- Create: `janet/test/test-workflow.janet`

Adds the `workflow` macro, `par` macro (wrapping `ev/gather`), thread-first `->` (already in Janet), `retry`, `timeout`, `gate`, `phase`, and `call-workflow`.

**Reference:** `internal/pipeline/eval.go` lines for `par`, `retry`, `timeout`, `gate`, `phase`, `call-workflow` special forms.

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-workflow.janet`:

```janet
(use spork/test)
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

# timeout raises on slow operation
(assert-error "timeout fires"
  (g/with-timeout 0.01
    (ev/sleep 1)
    "too slow"))

# gate checks a predicate
(g/reset-steps!)
(g/step "data" "good")
(assert (g/gate "quality" (not (nil? (g/ref "data"))))
        "gate passes when predicate is true")

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

Expected: FAIL — `workflow`, `par`, `retry`, `with-timeout`, `gate` not defined.

- [ ] **Step 3: Add workflow, par, retry, timeout, gate to core.janet**

Append to `janet/src/glitch/core.janet`:

```janet
# --- Workflow macro ---

(defmacro workflow [name & body]
  "Define and execute a workflow. Returns {:name :output :steps}."
  ~(do
     (var- wf-last nil)
     # Extract keyword args before body forms
     ,;(seq [form :in body
             :when (not (keyword? form))]
         ~(set wf-last ,form))
     {:name ,name
      :output (string wf-last)
      :steps (,get-steps)}))

# --- Parallel execution ---

(defmacro par [& forms]
  "Execute forms concurrently via ev/gather. Returns array of results."
  ~(ev/gather ,;forms))

# --- Retry ---

(defn retry [max-attempts f]
  "Retry a thunk up to max-attempts times. Returns first success."
  (var last-err nil)
  (for i 0 max-attempts
    (try
      (do (def result (f))
          (return result))
      ([err] (set last-err err))))
  (error (string "retry exhausted after " max-attempts " attempts: " last-err)))

(defmacro retry [n & body]
  "Retry body up to n times."
  ~(,retry ,n (fn [] ,;body)))

# --- Timeout ---

(defmacro with-timeout [seconds & body]
  "Execute body with a timeout. Raises on expiry."
  ~(do
     (def deadline (ev/deadline ,seconds))
     (defer (ev/cancel deadline nil)
       ,;body)))

# --- Gate ---

(defn gate [id predicate]
  "Assert a gate condition. Records result and returns predicate value."
  (when *step-recorder*
    (*step-recorder* {:step-id id :output (string predicate)
                      :kind "gate" :gate-passed (if predicate 1 0)}))
  (put *steps* id (string predicate))
  predicate)

# --- Phase ---

(defmacro phase [name & body]
  "Group steps under a named phase. Returns last value."
  ~(do
     (when *step-recorder*
       (*step-recorder* {:step-id ,name :kind "phase" :output "started"}))
     (var- phase-result nil)
     ,;(map (fn [f] ~(set phase-result ,f)) body)
     (when *step-recorder*
       (*step-recorder* {:step-id ,name :kind "phase"
                         :output (string phase-result)}))
     phase-result))

# --- Call-workflow ---

(var- *call-stack* @[])
(var- *workflows-dir* ".")

(defn set-workflows-dir! [d]
  (set *workflows-dir* d))

(defn call-workflow [name &named input set]
  "Execute another workflow file by name. Detects cycles."
  (when (find |(= $ name) *call-stack*)
    (errorf "call-workflow cycle: %s already on stack %s"
            name (string/join *call-stack* " -> ")))
  (array/push *call-stack* name)
  (defer (array/pop *call-stack*)
    (def path (string *workflows-dir* "/" name ".janet"))
    (unless (os/stat path)
      (errorf "call-workflow: %s not found" path))
    (def saved-input *input*)
    (def saved-steps *steps*)
    (set *input* (or input ""))
    (set *steps* @{})
    (when set
      (eachp [k v] set
        (put *params* k v)))
    (def mod (dofile path))
    (def result {:output (string (last (values *steps*)))
                 :steps (table/clone *steps*)})
    (set *input* saved-input)
    (set *steps* saved-steps)
    result))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `workflow` suite passes.

- [ ] **Step 5: Commit**

```bash
git add janet/src/glitch/core.janet janet/test/test-workflow.janet
git commit -m "feat(janet): workflow, par, retry, timeout, gate, phase, call-workflow"
```

---

### Task 4: LLM + HTTP Builtins

**Files:**
- Modify: `janet/src/glitch/core.janet`
- Create: `janet/src/glitch/http.janet`
- Create: `janet/test/test-http.janet`

Adds `llm` (dispatches to provider), `http-get`/`fetch`, `http-post`/`send`, `websearch`, and `json-pick`/`pick`.

**Reference:** `internal/pipeline/eval_builtins.go` — `builtinLLM`, `builtinHttpGet`, `builtinHttpPost`, `builtinWebsearch`, `builtinPick`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-http.janet`:

```janet
(use spork/test)
(import glitch/http :as h)

(start-suite "http")

# json-pick extracts from JSON string
(def data `{"name":"alice","items":[1,2,3]}`)
(assert (= "alice" (h/json-pick data "name"))
        "json-pick extracts top-level key")
(assert (deep= @[1 2 3] (h/json-pick data "items"))
        "json-pick extracts array")

# nested pick
(def nested `{"a":{"b":{"c":"deep"}}}`)
(assert (= "deep" (h/json-pick nested "a.b.c"))
        "json-pick handles dotted paths")

# pick with $ returns whole object
(def obj `{"x":1}`)
(assert (deep= {"x" 1} (h/json-pick obj "$"))
        "json-pick $ returns whole object")

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement http.janet**

Create `janet/src/glitch/http.janet`:

```janet
# HTTP client helpers and JSON path extraction.
# Replaces: builtinHttpGet, builtinHttpPost, builtinWebsearch, builtinPick

(import spork/json)

(defn http-get [url &named headers]
  "HTTP GET request. Returns response body as string."
  (def args @["curl" "-sS" "-L"])
  (when headers
    (eachp [k v] headers
      (array/push args "-H" (string k ": " v))))
  (array/push args url)
  (def proc (os/spawn args :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "http-get %s failed (exit %d): %s" url exit (string err-out)))
  (string out))

(defn http-post [url &named headers body]
  "HTTP POST request. Returns response body as string."
  (def args @["curl" "-sS" "-L" "-X" "POST"])
  (when headers
    (eachp [k v] headers
      (array/push args "-H" (string k ": " v))))
  (when body
    (array/push args "-d" body))
  (array/push args url)
  (def proc (os/spawn args :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "http-post %s failed (exit %d): %s" url exit (string err-out)))
  (string out))

(defn json-pick [json-str path]
  "Extract a value from a JSON string using a dotted path.
   '$' returns the whole parsed object."
  (def parsed (json/decode json-str))
  (if (= path "$")
    parsed
    (do
      (var current parsed)
      (each seg (string/split "." path)
        (when (nil? current) (break))
        (set current (get current seg)))
      current)))

(defn websearch [query &named url]
  "Search via SearXNG. Returns JSON results."
  (default url "http://localhost:8888")
  (def search-url (string url "/search?q="
                          (uri-encode query)
                          "&format=json"))
  (http-get search-url))

(defn- uri-encode [s]
  "Minimal URI percent-encoding."
  (def buf @"")
  (each byte s
    (if (or (<= (chr "a") byte (chr "z"))
            (<= (chr "A") byte (chr "Z"))
            (<= (chr "0") byte (chr "9"))
            (find |(= $ byte) [(chr "-") (chr "_") (chr ".") (chr "~")]))
      (buffer/push buf byte)
      (buffer/format buf "%%%02X" byte)))
  (string buf))
```

- [ ] **Step 4: Add llm function to core.janet**

Append to `janet/src/glitch/core.janet`:

```janet
# --- LLM invocation ---
# Provider dispatch is pluggable. Set via set-provider-fn!

(var- *provider-fn* nil)

(defn set-provider-fn! [f]
  "Set the LLM provider dispatch function.
   f should be (fn [opts] {:response ... :tokens-in ... :tokens-out ...})"
  (set *provider-fn* f))

(defn llm [& kvs]
  "Call an LLM provider. Keyword args: :prompt :model :provider :skill.
   Returns the response string."
  (def opts (table ;kvs))
  (unless *provider-fn*
    (error "llm: no provider function set — call set-provider-fn! first"))
  (when (opts :skill)
    (def skill-content (slurp (opts :skill)))
    (put opts :prompt (string skill-content "\n\n" (opts :prompt))))
  (def start (os/clock))
  (def result (*provider-fn* opts))
  (def elapsed (- (os/clock) start))
  (when *step-recorder*
    (*step-recorder* {:step-id (or (opts :step-id) "llm")
                      :prompt (opts :prompt)
                      :output (result :response)
                      :model (or (opts :model) "")
                      :duration (math/round (* elapsed 1000))
                      :kind "llm"
                      :tokens-in (or (result :tokens-in) 0)
                      :tokens-out (or (result :tokens-out) 0)}))
  (result :response))
```

- [ ] **Step 5: Run tests**

```bash
cd janet && jpm test
```

Expected: `http` suite passes.

- [ ] **Step 6: Commit**

```bash
git add janet/src/glitch/http.janet janet/src/glitch/core.janet janet/test/test-http.janet
git commit -m "feat(janet): HTTP client, json-pick, websearch, llm dispatch"
```

---

### Task 5: Provider System

**Files:**
- Create: `janet/src/glitch/provider.janet`
- Create: `janet/providers/ollama.janet`
- Create: `janet/providers/claude.janet`
- Create: `janet/providers/copilot.janet`
- Create: `janet/providers/lmstudio.janet`
- Create: `janet/test/test-provider.janet`

Providers are Janet functions. The registry scans directories for `.janet` files. Tier escalation tries providers in order.

**Reference:** `internal/provider/provider.go`, `internal/provider/tiers.go`, `internal/provider/openai.go`, `internal/provider/lmstudio.go`, `internal/provider/agent.go`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-provider.janet`:

```janet
(use spork/test)
(import glitch/provider :as p)

(start-suite "provider")

# register and call a mock provider
(p/reset!)
(p/register "mock"
  (fn [opts]
    {:response (string "echo:" (opts :prompt))
     :tokens-in 10
     :tokens-out 20
     :latency 0
     :cost 0}))

(assert (find |(= $ "mock") (p/names))
        "registered provider appears in names")

(def result (p/call-provider "mock" {:prompt "hello" :model "test"}))
(assert (= "echo:hello" (result :response))
        "call-provider dispatches correctly")

# unknown provider raises
(assert-error "unknown provider"
  (p/call-provider "nonexistent" {:prompt "x"}))

# tier escalation
(p/register "fail-provider"
  (fn [opts] (error "provider down")))

(def tiers [{:providers ["fail-provider"]}
            {:providers ["mock"]}])

(def tiered (p/call-tiered {:prompt "test"} tiers))
(assert (= "echo:test" (tiered :response))
        "tier escalation falls through to working provider")

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement provider.janet**

Create `janet/src/glitch/provider.janet`:

```janet
# Provider registry and tier escalation.
# Replaces: internal/provider/provider.go, tiers.go

(var- registry @{})

(defn reset! []
  (set registry @{}))

(defn register [name provider-fn]
  "Register a provider function. provider-fn takes opts table,
   returns {:response :tokens-in :tokens-out :latency :cost}."
  (put registry name provider-fn))

(defn names []
  "Return sorted list of registered provider names."
  (sorted (keys registry)))

(defn call-provider [name opts]
  "Call a registered provider by name."
  (def provider (get registry name))
  (unless provider
    (errorf "unknown provider: %s" name))
  (provider opts))

(defn call-tiered [opts tiers]
  "Try providers in tier order. Returns first success."
  (var last-err nil)
  (each tier tiers
    (def model-override (get tier :model))
    (each pname (tier :providers)
      (def merged (merge opts
                    (if model-override {:model model-override} {})))
      (try
        (do
          (def result (call-provider pname merged))
          (when (or (nil? (result :response))
                    (= "" (string/trim (result :response))))
            (error "empty response"))
          (return result))
        ([err] (set last-err err)
               (eprintf "tier: %s failed: %s" pname (string err))))))
  (errorf "all tiers exhausted: %s" (string last-err)))

(def default-tiers
  [{:providers ["ollama"] :model "qwen2.5:7b"}
   {:providers ["codex" "gemini"]}
   {:providers ["copilot" "claude"]}])

(defn load-providers [& dirs]
  "Load .janet provider files from directories.
   Each file must export a :call function."
  (def search-dirs
    (if (> (length dirs) 0)
      dirs
      [(string (os/getenv "HOME") "/.config/glitch/providers")
       "providers"]))
  (each dir search-dirs
    (when (os/stat dir)
      (each f (os/dir dir)
        (when (string/has-suffix? ".janet" f)
          (def name (string/slice f 0 (- (length f) 6)))
          (try
            (do
              (def mod (dofile (string dir "/" f)))
              (when (mod :call)
                (register name (mod :call))))
            ([err]
              (eprintf "warn: failed to load provider %s: %s" name (string err)))))))))
```

- [ ] **Step 4: Write ollama provider**

Create `janet/providers/ollama.janet`:

```janet
# Ollama provider — local LLM via HTTP API.
# Reference: internal/provider/openai.go (Ollama uses its own API)

(import spork/json)

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "qwen2.5:7b")
  (def body (json/encode {:model model :prompt prompt :stream false}))
  (def proc (os/spawn
    ["curl" "-sS" "-X" "POST"
     "http://localhost:11434/api/generate"
     "-H" "Content-Type: application/json"
     "-d" body]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "ollama: request failed (exit %d): %s" exit (string err-out)))
  (def parsed (json/decode (string out)))
  {:response (or (get parsed "response") "")
   :tokens-in (or (get parsed "prompt_eval_count") 0)
   :tokens-out (or (get parsed "eval_count") 0)
   :latency (or (get parsed "total_duration") 0)
   :cost 0})
```

- [ ] **Step 5: Write claude provider**

Create `janet/providers/claude.janet`:

```janet
# Claude provider — headless CLI invocation.
# Reference: internal/provider/agent.go

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "sonnet")
  (def proc (os/spawn
    ["claude" "--print" "--model" model "-" prompt]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (os/proc-wait proc)
  {:response (string/trim (string out))
   :tokens-in 0
   :tokens-out 0
   :latency 0
   :cost 0})
```

- [ ] **Step 6: Write copilot provider**

Create `janet/providers/copilot.janet`:

```janet
# Copilot provider — GitHub Copilot CLI.
# Reference: internal/provider/agent.go

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "gpt-4o")
  # Copilot needs prompt in a temp file
  (def tmp (string "/tmp/glitch-copilot-" (os/time) ".txt"))
  (spit tmp prompt)
  (defer (os/rm tmp)
    (def proc (os/spawn
      ["gh" "copilot" "suggest" "-t" "shell" (slurp tmp)]
      :p {:out :pipe :err :pipe}))
    (def out (ev/read (proc :out) :all))
    (os/proc-wait proc)
    {:response (string/trim (string out))
     :tokens-in 0
     :tokens-out 0
     :latency 0
     :cost 0}))
```

- [ ] **Step 7: Write lmstudio provider**

Create `janet/providers/lmstudio.janet`:

```janet
# LM Studio provider — OpenAI-compatible local API.
# Reference: internal/provider/lmstudio.go

(import spork/json)

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "default")
  (def body (json/encode
    {:model model
     :messages [{:role "user" :content prompt}]
     :stream false}))
  (def proc (os/spawn
    ["curl" "-sS" "-X" "POST"
     "http://localhost:1234/v1/chat/completions"
     "-H" "Content-Type: application/json"
     "-d" body]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "lmstudio: request failed (exit %d): %s" exit (string err-out)))
  (def parsed (json/decode (string out)))
  (def choice (get-in parsed ["choices" 0 "message" "content"] ""))
  (def usage (get parsed "usage" {}))
  {:response choice
   :tokens-in (or (get usage "prompt_tokens") 0)
   :tokens-out (or (get usage "completion_tokens") 0)
   :latency 0
   :cost 0})
```

- [ ] **Step 8: Run tests**

```bash
cd janet && jpm test
```

Expected: `provider` suite passes.

- [ ] **Step 9: Commit**

```bash
git add janet/src/glitch/provider.janet janet/providers/ janet/test/test-provider.janet
git commit -m "feat(janet): provider registry, tier escalation, ollama/claude/copilot/lmstudio"
```

---

### Task 6: Store (SQLite)

**Files:**
- Create: `janet/src/glitch/store.janet`
- Create: `janet/test/test-store.janet`

SQLite-backed run/step recording. Same schema as Go — existing `.db` files are compatible.

**Reference:** `internal/store/store.go`, `internal/store/schema.go`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-store.janet`:

```janet
(use spork/test)
(import glitch/store :as s)

(start-suite "store")

(def db-path (string "/tmp/glitch-test-" (os/time) ".db"))
(def db (s/open db-path))

# record a run
(def run-id (s/record-run db
  {:name "test-wf" :input "hello" :workflow-file "test.janet"
   :model "qwen2.5:7b" :workspace "default"}))
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement store.janet**

Create `janet/src/glitch/store.janet`:

```janet
# SQLite run/step recording.
# Replaces: internal/store/store.go + schema.go
# Schema is identical to Go — existing .db files are compatible.

(import sqlite3 :as sql)

(def- schema `
CREATE TABLE IF NOT EXISTS runs (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  kind            TEXT NOT NULL DEFAULT 'workflow',
  name            TEXT NOT NULL,
  input           TEXT,
  output          TEXT,
  exit_status     INTEGER,
  started_at      INTEGER NOT NULL,
  finished_at     INTEGER,
  metadata        TEXT,
  workflow_file   TEXT,
  repo            TEXT,
  model           TEXT,
  tokens_in       INTEGER,
  tokens_out      INTEGER,
  cost_usd        REAL,
  variant         TEXT,
  workspace       TEXT NOT NULL DEFAULT '',
  parent_run_id   INTEGER REFERENCES runs(id),
  workflow_name   TEXT
);
CREATE INDEX IF NOT EXISTS idx_runs_parent ON runs(parent_run_id);

CREATE TABLE IF NOT EXISTS steps (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id      INTEGER NOT NULL,
  step_id     TEXT NOT NULL,
  prompt      TEXT,
  output      TEXT,
  model       TEXT,
  duration_ms INTEGER,
  kind        TEXT,
  exit_status INTEGER,
  tokens_in   INTEGER,
  tokens_out  INTEGER,
  gate_passed INTEGER,
  artifacts   TEXT,
  UNIQUE(run_id, step_id)
);

CREATE TABLE IF NOT EXISTS research_events (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  query_id        TEXT NOT NULL,
  question        TEXT NOT NULL,
  researchers     TEXT NOT NULL,
  composite_score REAL,
  reason          TEXT,
  created_at      INTEGER NOT NULL
);
`)

(defn open [&opt path]
  "Open or create a glitch database. Returns the db handle."
  (default path (string (os/getenv "HOME")
                        "/.local/share/glitch/glitch.db"))
  # Ensure parent directory exists
  (def dir (string/slice path 0
             (or (string/find-last "/" path) 0)))
  (when (and (not= dir "") (not (os/stat dir)))
    (os/shell (string "mkdir -p " dir)))
  (def db (sql/open path))
  (sql/eval db schema)
  db)

(defn open-for-workspace [ws-path]
  "Open a workspace-scoped database."
  (open (string ws-path "/.glitch/glitch.db")))

(defn close [db]
  (sql/close db))

(defn record-run [db rec]
  "Insert a run record. Returns the row ID."
  (sql/eval db
    `INSERT INTO runs (name, input, workflow_file, model,
                       variant, workspace, parent_run_id,
                       workflow_name, started_at, kind)
     VALUES (:name, :input, :workflow-file, :model,
             :variant, :workspace, :parent-run-id,
             :workflow-name, :now, 'workflow')`
    (merge rec {:now (os/time)}))
  (first (first (sql/eval db `SELECT last_insert_rowid()`))))

(defn finish-run [db id output exit-status totals]
  "Update a run with final output, status, and token totals."
  (sql/eval db
    `UPDATE runs SET output=:output, exit_status=:exit,
       tokens_in=:tokens-in, tokens_out=:tokens-out,
       cost_usd=:cost, finished_at=:now
     WHERE id=:id`
    (merge (or totals {})
      {:id id :output output :exit exit-status
       :now (os/time)})))

(defn record-step [db rec]
  "Insert or replace a step record."
  (sql/eval db
    `INSERT OR REPLACE INTO steps
       (run_id, step_id, prompt, output, model,
        duration_ms, kind, exit_status,
        tokens_in, tokens_out, gate_passed, artifacts)
     VALUES (:run-id, :step-id, :prompt, :output, :model,
             :duration, :kind, :exit,
             :tokens-in, :tokens-out, :gate, :artifacts)`
    rec))

(defn get-run [db id]
  "Fetch a single run by ID."
  (def rows (sql/eval db
    `SELECT * FROM runs WHERE id = :id` {:id id}))
  (when (> (length rows) 0) (first rows)))

(defn get-steps [db run-id]
  "Fetch all steps for a run."
  (sql/eval db
    `SELECT * FROM steps WHERE run_id = :rid ORDER BY id`
    {:rid run-id}))

(defn list-runs [db &named parent-id workflow limit]
  "List runs with optional filters."
  (default limit 50)
  (var query `SELECT * FROM runs WHERE 1=1`)
  (def params @{:limit limit})
  (when parent-id
    (set query (string query ` AND parent_run_id = :pid`))
    (put params :pid parent-id))
  (when workflow
    (set query (string query ` AND workflow_file = :wf`))
    (put params :wf workflow))
  (set query (string query ` ORDER BY id DESC LIMIT :limit`))
  (sql/eval db query params))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `store` suite passes.

- [ ] **Step 5: Commit**

```bash
git add janet/src/glitch/store.janet janet/test/test-store.janet
git commit -m "feat(janet): SQLite store — run/step recording, compatible schema"
```

---

### Task 7: Workspace System

**Files:**
- Create: `janet/src/glitch/workspace.janet`
- Create: `janet/test/test-workspace.janet`

Workspace files are Janet programs. The macros `workspace`, `defaults`, `resource` build a table.

**Reference:** `internal/workspace/workspace.go`, `internal/workspace/resources.go`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-workspace.janet`:

```janet
(use spork/test)
(import glitch/workspace :as ws)

(start-suite "workspace")

# parse a workspace definition
(def tmp-dir (string "/tmp/glitch-ws-test-" (os/time)))
(os/shell (string "mkdir -p " tmp-dir "/.glitch"))

(spit (string tmp-dir "/.glitch/workspace.janet")
  `(import glitch/workspace)
   (glitch/workspace/workspace "test-ws"
     :description "Test workspace"
     :owner "tester"
     (glitch/workspace/defaults
       :model "qwen2.5:7b"
       :provider "ollama")
     (glitch/workspace/resource "myrepo"
       :type "git"
       :url "https://github.com/example/repo"
       :ref "main")
     (glitch/workspace/resource "local-docs"
       :type "local"
       :path "/tmp/docs"))`)

(def w (ws/load (string tmp-dir "/.glitch/workspace.janet")))
(assert (= "test-ws" (w :name)) "workspace name")
(assert (= "Test workspace" (w :description)) "workspace description")
(assert (= "qwen2.5:7b" (get-in w [:defaults :model])) "default model")
(assert (= "ollama" (get-in w [:defaults :provider])) "default provider")
(assert (= 2 (length (w :resources))) "two resources")

(def repo (ws/get-resource w "myrepo"))
(assert (= "git" (repo :type)) "resource type")
(assert (= "main" (repo :ref)) "resource ref")

(def docs (ws/get-resource w "local-docs"))
(assert (= "local" (docs :type)) "local resource type")
(assert (= "/tmp/docs" (docs :path)) "local resource path")

# find-workspace-file walks up
(def found (ws/find-workspace-file tmp-dir))
(assert found "find-workspace-file finds workspace")

# cleanup
(os/shell (string "rm -rf " tmp-dir))

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement workspace.janet**

Create `janet/src/glitch/workspace.janet`:

```janet
# Workspace parsing, resource binding, and discovery.
# Replaces: internal/workspace/workspace.go + resources.go

(var- *current-ws* nil)

(defn workspace [name & body]
  "Define a workspace. Called from workspace.janet files."
  (set *current-ws*
    @{:name name :resources @[] :defaults @{} :path ""})
  # Process keyword args
  (var i 0)
  (while (< i (length body))
    (def item (get body i))
    (cond
      (= item :description) (do (put *current-ws* :description (get body (+ i 1)))
                                (+= i 2))
      (= item :owner) (do (put *current-ws* :owner (get body (+ i 1)))
                          (+= i 2))
      # body forms (defaults, resource) are already evaluated
      (+= i 1)))
  *current-ws*)

(defn defaults [& kvs]
  "Set workspace defaults. Called inside workspace form."
  (def tbl (table ;kvs))
  (when *current-ws*
    (merge-into (*current-ws* :defaults) tbl))
  tbl)

(defn resource [name & kvs]
  "Add a resource to the current workspace."
  (def res (merge @{:name name} (table ;kvs)))
  (when *current-ws*
    (array/push (*current-ws* :resources) res))
  res)

(defn load [path]
  "Load a workspace.janet file and return the workspace table."
  (set *current-ws* nil)
  (dofile path)
  (when *current-ws*
    (put *current-ws* :path
      (string/slice path 0
        (or (string/find-last "/" path) 0))))
  *current-ws*)

(defn get-resource [ws name]
  "Find a resource by name."
  (find |(= ($ :name) name) (ws :resources)))

(defn resources-by-type [ws type-str]
  "Filter resources by type."
  (filter |(= ($ :type) type-str) (ws :resources)))

(defn resource-fields [ws name]
  "Return a resource's fields as a flat table for template binding."
  (def res (get-resource ws name))
  (or res @{}))

(defn find-workspace-file [dir]
  "Walk up from dir to find .glitch/workspace.janet."
  (var current dir)
  (while true
    (def path (string current "/.glitch/workspace.janet"))
    (when (os/stat path)
      (return path))
    (def parent (string/slice current 0
                  (or (string/find-last "/" current) 0)))
    (when (or (= parent current) (= parent ""))
      (break))
    (set current parent))
  nil)

(defn resolve [&named workspace-flag]
  "Resolve the active workspace.
   Priority: explicit flag > env var > walk cwd > nil."
  (def path
    (or workspace-flag
        (os/getenv "GLITCH_WORKSPACE")
        (find-workspace-file (os/cwd))))
  (when path
    (load path)))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `workspace` suite passes.

- [ ] **Step 5: Commit**

```bash
git add janet/src/glitch/workspace.janet janet/test/test-workspace.janet
git commit -m "feat(janet): workspace system — parsing, resources, discovery"
```

---

### Task 8: Runner — Full Execution Lifecycle

**Files:**
- Create: `janet/src/glitch/runner.janet`
- Create: `janet/test/test-runner.janet`

Wires core + provider + store together. Loads a `.janet` workflow file, sets up context, executes, records results.

**Reference:** `internal/pipeline/runner.go` — `Run()`, `RunOpts`, `Result`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-runner.janet`:

```janet
(use spork/test)
(import glitch/runner :as r)
(import glitch/store :as s)

(start-suite "runner")

# Create a test workflow file
(def wf-dir (string "/tmp/glitch-runner-test-" (os/time)))
(os/shell (string "mkdir -p " wf-dir))
(spit (string wf-dir "/hello.janet")
  `(import glitch/core :as g)
   (g/workflow "hello"
     :description "test"
     (g/step "greet" (string "hello " (g/input))))`)

# Run without store
(def result (r/run (string wf-dir "/hello.janet") "world"))
(assert (= "hello world" (result :output)) "runner captures output")
(assert (= "hello" (result :name)) "runner captures workflow name")
(assert (= "hello world" (get-in result [:steps "greet"]))
        "runner captures step outputs")

# Run with store
(def db-path (string "/tmp/glitch-runner-store-" (os/time) ".db"))
(def db (s/open db-path))
(def result2 (r/run (string wf-dir "/hello.janet") "db-test"
               :db db))
(assert (> (result2 :run-id) 0) "runner records run id")
(def stored-run (s/get-run db (result2 :run-id)))
(assert (= "hello db-test" (stored-run :output)) "store has correct output")

# cleanup
(s/close db)
(os/rm db-path)
(os/shell (string "rm -rf " wf-dir))

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement runner.janet**

Create `janet/src/glitch/runner.janet`:

```janet
# Workflow execution lifecycle.
# Replaces: internal/pipeline/runner.go

(import glitch/core :as g)
(import glitch/store :as s)
(import glitch/provider :as p)

(defn run [workflow-path input &named db workspace model
           params seed-steps workflows-dir parent-run-id
           tiers]
  "Execute a workflow file. Returns {:name :output :steps :run-id}."
  (default input "")
  (default model "qwen2.5:7b")
  (default params @{})
  (default workflows-dir
    (string/slice workflow-path 0
      (or (string/find-last "/" workflow-path) 0)))

  # Reset core state
  (g/reset-steps!)
  (g/set-input! input)
  (g/set-params! params)
  (g/set-workflows-dir! workflows-dir)

  # Wire provider dispatch
  (g/set-provider-fn!
    (fn [opts]
      (def pname (or (opts :provider) "ollama"))
      (def merged (merge opts {:model (or (opts :model) model)}))
      (if tiers
        (p/call-tiered merged tiers)
        (try
          (p/call-provider pname merged)
          ([err]
            (p/call-tiered merged p/default-tiers))))))

  # Pre-seed steps
  (when seed-steps
    (eachp [k v] seed-steps
      (g/step k v)))

  # Record run in store
  (var run-id 0)
  (when db
    (set run-id (s/record-run db
      {:name "" :input input
       :workflow-file workflow-path
       :model model
       :workspace (or workspace "")
       :parent-run-id parent-run-id}))
    # Wire step recorder to store
    (g/set-step-recorder!
      (fn [rec]
        (s/record-step db
          (merge rec {:run-id run-id})))))

  # Execute the workflow
  (var result nil)
  (try
    (do
      (def mod (dofile workflow-path))
      (set result (or (mod :result) {:output "" :steps @{}}))
      # If the workflow used the workflow macro, result has :name :output :steps
      # Otherwise gather from core state
      (when (nil? (result :name))
        (def steps (g/get-steps))
        (var last-output "")
        (eachp [_ v] steps (set last-output v))
        (set result {:name (or (result :name) "")
                     :output last-output
                     :steps steps})))
    ([err]
      (when db
        (s/finish-run db run-id (string "ERROR: " err) 1 {}))
      (error err)))

  # Finalize
  (when db
    (s/finish-run db run-id (result :output) 0 {}))

  (merge result {:run-id run-id}))

(defn list-workflows [dir]
  "List all .janet workflow files in a directory."
  (def results @[])
  (when (os/stat dir)
    (each f (os/dir dir)
      (when (string/has-suffix? ".janet" f)
        (array/push results
          {:name (string/slice f 0 (- (length f) 6))
           :path (string dir "/" f)}))))
  (sort-by |($ :name) results))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `runner` suite passes.

- [ ] **Step 5: Commit**

```bash
git add janet/src/glitch/runner.janet janet/test/test-runner.janet
git commit -m "feat(janet): runner — full execution lifecycle with store recording"
```

---

### Task 9: Stdlib

**Files:**
- Create: `janet/stdlib/collections.janet`
- Create: `janet/stdlib/io.janet`
- Create: `janet/stdlib/strings.janet`
- Create: `janet/test/test-stdlib.janet`

Port the three stdlib files from the Go version.

**Reference:** `internal/pipeline/stdlib/collections.glitch`, `io.glitch`, `strings.glitch`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-stdlib.janet`:

```janet
(use spork/test)
(import glitch/stdlib/collections :as c)
(import glitch/stdlib/strings :as str)
(import glitch/stdlib/io :as gio)

(start-suite "stdlib")

# collections
(assert (deep= @["a" "b"] (c/compact @["a" "" "b" ""]))
        "compact removes empties")
(assert (= "first" (c/first @["first" "second"]))
        "first returns index 0")
(assert (deep= @["x" "y"]
          (c/pluck :name @[{:name "x"} {:name "y"}]))
        "pluck extracts field")
(assert (deep= @[1 2] (c/take 2 @[1 2 3 4]))
        "take returns first n")
(assert (deep= @[1 2 3] (c/unique @[1 2 2 3 1]))
        "unique deduplicates")
(assert (deep= @[1 3] (c/without @[1 2 3] @[2]))
        "without excludes items")

# strings
(assert (= "hello-world" (str/kebab-case "hello world"))
        "kebab-case")
(assert (= "hello_world" (str/snake-case "hello world"))
        "snake-case")
(assert (str/blank? "") "blank? for empty")
(assert (str/blank? "  ") "blank? for whitespace")
(assert (not (str/blank? "hi")) "blank? for non-empty")
(assert (str/present? "hi") "present? for non-empty")
(assert (not (str/present? "")) "present? for empty")
(assert (deep= @["hello" "world"] (str/words "hello world"))
        "words splits on whitespace")
(assert (= "hello world" (str/unwords @["hello" "world"]))
        "unwords joins with space")

# io
(def tmp (string "/tmp/glitch-stdlib-test-" (os/time)))
(gio/write-lines tmp @["one" "two" "three"])
(assert (deep= @["one" "two" "three"] (gio/read-lines tmp))
        "write-lines + read-lines roundtrip")
(os/rm tmp)

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement collections.janet**

Create `janet/stdlib/collections.janet`:

```janet
# Collection utilities.
# Port of: internal/pipeline/stdlib/collections.glitch

(defn compact [source]
  "Remove empty strings from an array."
  (filter |(not= $ "") source))

(defn first [source]
  "Return the first element."
  (get source 0))

(defn pluck [key source]
  "Extract a field from each element."
  (map |(get $ key) source))

(defn take [n source]
  "Return the first n elements."
  (array/slice source 0 (min n (length source))))

(defn unique [source]
  "Remove duplicates, preserving order."
  (def seen @{})
  (def result @[])
  (each item source
    (unless (get seen item)
      (put seen item true)
      (array/push result item)))
  result)

(defn without [source exclude]
  "Return source with exclude items removed."
  (def ex-set (from-pairs (map |[$ true] exclude)))
  (filter |(not (get ex-set $)) source))
```

- [ ] **Step 4: Implement strings.janet**

Create `janet/stdlib/strings.janet`:

```janet
# String utilities.
# Port of: internal/pipeline/stdlib/strings.glitch

(defn kebab-case [s]
  "Convert string to kebab-case."
  (-> s
      string/ascii-lower
      (string/replace-all " " "-")
      (string/replace-all "_" "-")))

(defn snake-case [s]
  "Convert string to snake_case."
  (-> s
      string/ascii-lower
      (string/replace-all " " "_")
      (string/replace-all "-" "_")))

(defn blank? [s]
  "True if string is empty or whitespace only."
  (= "" (string/trim (or s ""))))

(defn present? [s]
  "True if string is non-empty and not just whitespace."
  (not (blank? s)))

(defn words [s]
  "Split string on whitespace."
  (filter |(not= $ "")
    (string/split " " (string/trim s))))

(defn unwords [lst]
  "Join array with spaces."
  (string/join lst " "))
```

- [ ] **Step 5: Implement io.janet**

Create `janet/stdlib/io.janet`:

```janet
# File I/O utilities.
# Port of: internal/pipeline/stdlib/io.glitch

(import spork/json)
(import glitch/http)

(defn fetch-json [url]
  "HTTP GET a URL and parse the response as JSON."
  (def body (http/http-get url))
  (json/decode body))

(defn read-lines [path]
  "Read a file and split into lines."
  (def content (string/trim (slurp path)))
  (if (= content "")
    @[]
    (string/split "\n" content)))

(defn write-lines [path lines]
  "Join lines and write to file."
  (spit path (string/join lines "\n")))
```

- [ ] **Step 6: Run tests**

```bash
cd janet && jpm test
```

Expected: `stdlib` suite passes.

- [ ] **Step 7: Commit**

```bash
git add janet/stdlib/ janet/test/test-stdlib.janet
git commit -m "feat(janet): stdlib — collections, strings, io"
```

---

### Task 10: CLI Dispatch

**Files:**
- Modify: `janet/src/glitch/main.janet`
- Create: `janet/test/test-cli.janet`

Full CLI with `spork/argparse`. Commands: `run`, `check`, `eval`, `workspace`, `config`, `plugin`, `up`, `version`.

**Reference:** `cmd/root.go`, `cmd/run.go`, `cmd/workspace.go`, `cmd/check.go`, `cmd/eval.go`

- [ ] **Step 1: Write failing tests for arg parsing**

Create `janet/test/test-cli.janet`:

```janet
(use spork/test)

# Test the parse-run-args function directly
(import glitch/main :as m)

(start-suite "cli")

# Basic command dispatch
(assert (= "version" (m/resolve-command @["version"]))
        "resolve version command")
(assert (= "run" (m/resolve-command @["run" "my-wf"]))
        "resolve run command")
(assert (= "workspace" (m/resolve-command @["workspace" "list"]))
        "resolve workspace command")
(assert (nil? (m/resolve-command @[]))
        "empty args returns nil")

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Rewrite main.janet with full CLI**

Overwrite `janet/src/glitch/main.janet`:

```janet
# CLI entry point.
# Replaces: cmd/root.go + all cmd/*.go files

(import spork/argparse :prefix "")
(import glitch/core :as g)
(import glitch/runner :as runner)
(import glitch/store :as store)
(import glitch/workspace :as ws)
(import glitch/provider :as prov)

(defn resolve-command [argv]
  "Return the command name from argv, or nil."
  (when (> (length argv) 0)
    (first argv)))

# --- Commands ---

(defn cmd-version []
  (print "glitch 0.1.0-janet"))

(defn cmd-check [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch check <file.janet>")
    (os/exit 1))
  (def path (first argv))
  (try
    (do
      (parse (slurp path))
      (print "ok"))
    ([err]
      (eprintf "check failed: %s" (string err))
      (os/exit 1))))

(defn cmd-eval [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch eval <file.janet>")
    (os/exit 1))
  (def path (first argv))
  (def result (dofile path))
  (when result (pp result)))

(defn cmd-run [argv]
  (def res
    (argparse
      "Execute a workflow"
      "workspace" {:kind :option :short "w"
                   :help "Workspace path"}
      "path"      {:kind :option :short "p"
                   :help "Explicit workflow file path"}
      "set"       {:kind :accumulate :short "s"
                   :help "Set param key=value"}
      "variant"   {:kind :accumulate
                   :help "Add variant provider:model"}
      "compare"   {:kind :flag
                   :help "Cross-compare all variants"}
      "model"     {:kind :option :short "m"
                   :help "Default model"}
      :default    {:kind :accumulate}))
  (unless res (os/exit 1))

  (def positional (or (res :default) @[]))
  (def wf-name (get positional 0))
  (def input (get positional 1 ""))

  (unless wf-name
    (eprint "usage: glitch run <workflow> [input]")
    (os/exit 1))

  # Resolve workspace
  (def workspace (ws/resolve
    :workspace-flag (res "workspace")))

  # Load providers
  (prov/load-providers)

  # Resolve workflow path
  (def wf-path
    (or (res "path")
        (let [dirs @[".glitch/workflows" "workflows"]]
          (when workspace
            (array/insert dirs 0
              (string (workspace :path) "/workflows")))
          (var found nil)
          (each dir dirs
            (def p (string dir "/" wf-name ".janet"))
            (when (os/stat p) (set found p) (break)))
          found)
        (do (eprintf "workflow not found: %s" wf-name)
            (os/exit 1))))

  # Parse --set params
  (def params @{})
  (when (res "set")
    (each kv (res "set")
      (def eq (string/find "=" kv))
      (when eq
        (put params
          (string/slice kv 0 eq)
          (string/slice kv (+ eq 1))))))

  # Model
  (def model (or (res "model")
                 (when workspace (get-in workspace [:defaults :model]))
                 "qwen2.5:7b"))

  # Open store
  (def db
    (if workspace
      (store/open-for-workspace (workspace :path))
      (store/open)))

  # Run
  (def result
    (runner/run wf-path input
      :db db
      :workspace (when workspace (workspace :name))
      :model model
      :params params))

  (print (result :output))
  (store/close db))

(defn cmd-workspace [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch workspace <init|use|list|status|add|rm|sync|pin|resources|gui>")
    (os/exit 1))
  (match (first argv)
    "init"      (ws-init (slice argv 1))
    "use"       (ws-use (slice argv 1))
    "list"      (ws-list)
    "status"    (ws-status)
    "add"       (ws-add (slice argv 1))
    "rm"        (ws-rm (slice argv 1))
    "sync"      (ws-sync (slice argv 1))
    "pin"       (ws-pin (slice argv 1))
    "resources" (ws-resources)
    "gui"       (ws-gui (slice argv 1))
    sub         (do (eprintf "unknown workspace command: %s" sub)
                    (os/exit 1))))

# Workspace subcommands (stubs — each is small, expand as needed)
(defn- ws-init [argv]
  (def name (or (first argv) (last (string/split "/" (os/cwd)))))
  (def dir (string (os/cwd) "/.glitch"))
  (os/shell (string "mkdir -p " dir))
  (spit (string dir "/workspace.janet")
    (string "(import glitch/workspace)\n"
            "(glitch/workspace/workspace " (string/format "%q" name) "\n"
            "  :description \"\")\n"))
  (printf "workspace %s initialized at %s" name dir))

(defn- ws-list []
  (def config-dir (string (os/getenv "HOME") "/.config/glitch"))
  (if (os/stat (string config-dir "/workspaces.janet"))
    (do
      (def ws-list (dofile (string config-dir "/workspaces.janet")))
      (each w (or ws-list @[])
        (printf "%s — %s" (w :name) (or (w :path) ""))))
    (print "no workspaces registered")))

(defn- ws-status []
  (def w (ws/resolve))
  (if w
    (do
      (printf "workspace: %s" (w :name))
      (printf "path: %s" (or (w :path) ""))
      (printf "model: %s" (get-in w [:defaults :model] ""))
      (printf "resources: %d" (length (or (w :resources) @[]))))
    (print "no active workspace")))

(defn- ws-use [argv]
  (printf "workspace use: %s (not yet implemented)" (first argv)))
(defn- ws-add [argv]
  (printf "workspace add: %s (not yet implemented)" (first argv)))
(defn- ws-rm [argv]
  (printf "workspace rm: %s (not yet implemented)" (first argv)))
(defn- ws-sync [argv]
  (printf "workspace sync (not yet implemented)"))
(defn- ws-pin [argv]
  (printf "workspace pin (not yet implemented)"))
(defn- ws-resources []
  (def w (ws/resolve))
  (if w
    (each r (or (w :resources) @[])
      (printf "%s (%s) — %s" (r :name) (r :type)
              (or (r :url) (r :path) "")))
    (print "no active workspace")))
(defn- ws-gui [argv]
  (printf "workspace gui (not yet implemented)"))

(defn cmd-plugin [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch plugin <name> [args...]")
    (os/exit 1))
  (def name (first argv))
  (def plugin-dir (string (os/getenv "HOME")
                          "/.config/glitch/plugins/" name))
  (def entry (string plugin-dir "/main.janet"))
  (unless (os/stat entry)
    (eprintf "plugin not found: %s" name)
    (os/exit 1))
  (def mod (dofile entry))
  (when (mod :main)
    ((mod :main) (slice argv 1))))

(defn cmd-up []
  # Check that required tools are available
  (each tool ["janet" "curl"]
    (def proc (os/spawn ["which" tool] :p {:out :pipe}))
    (def out (ev/read (proc :out) :all))
    (os/proc-wait proc)
    (if (> (length (string/trim (string out))) 0)
      (printf "  %s ✓" tool)
      (printf "  %s ✗ (not found)" tool))))

# --- Main ---

(def commands
  {"run"       cmd-run
   "check"     cmd-check
   "eval"      cmd-eval
   "workspace" cmd-workspace
   "config"    (fn [_] (print "config (not yet implemented)"))
   "plugin"    cmd-plugin
   "up"        (fn [_] (cmd-up))
   "version"   (fn [_] (cmd-version))})

(defn main [& argv]
  (def cmd (resolve-command argv))
  (if-let [handler (get commands cmd)]
    (handler (slice argv 1))
    (do
      (print "glitch - workflow engine (janet)")
      (print "")
      (print "commands:")
      (each name (sorted (keys commands))
        (printf "  %s" name))
      (when (nil? cmd) (os/exit 1)))))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `cli` suite passes.

- [ ] **Step 5: Build and test binary**

```bash
cd janet && jpm build && ./build/glitch version
```

Expected: `glitch 0.1.0-janet`

```bash
./build/glitch up
```

Expected: shows janet and curl status.

- [ ] **Step 6: Commit**

```bash
git add janet/src/glitch/main.janet janet/test/test-cli.janet
git commit -m "feat(janet): full CLI dispatch with argparse — run, check, eval, workspace, plugin"
```

---

### Task 11: GUI Server

**Files:**
- Create: `janet/src/glitch/gui.janet`
- Create: `janet/test/test-gui.janet`

HTTP server via `spork/httpf`. Serves the REST API that the Svelte frontend consumes, plus static file serving for the SPA.

**Reference:** `internal/gui/server.go`, `internal/gui/api_runs.go`, `internal/gui/api_workflows.go`, `internal/gui/api_resources.go`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-gui.janet`:

```janet
(use spork/test)
(import glitch/gui :as gui)
(import glitch/store :as s)
(import spork/json)

(start-suite "gui")

# Test JSON response helper
(def resp (gui/json-response {:hello "world"}))
(assert (= 200 (resp :status)) "json-response status 200")
(assert (= "application/json"
           (get-in resp [:headers "Content-Type"]))
        "json-response content type")
(def body (json/decode (resp :body)))
(assert (= "world" (get body "hello"))
        "json-response encodes body")

# Test route builder builds a handler table
(def db-path (string "/tmp/glitch-gui-test-" (os/time) ".db"))
(def db (s/open db-path))
(def routes (gui/build-routes db nil))
(assert (> (length routes) 0) "build-routes returns routes")

# cleanup
(s/close db)
(os/rm db-path)

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement gui.janet**

Create `janet/src/glitch/gui.janet`:

```janet
# HTTP server for the glitch GUI.
# Replaces: internal/gui/server.go + api_*.go
# Serves same REST API contract — Svelte frontend unchanged.

(import spork/http)
(import spork/json)
(import glitch/store :as s)
(import glitch/runner :as runner)
(import glitch/workspace :as ws)
(import glitch/provider :as prov)

# --- Helpers ---

(defn json-response [data &opt status]
  "Build a JSON HTTP response."
  (default status 200)
  {:status status
   :headers {"Content-Type" "application/json"
             "Access-Control-Allow-Origin" "*"}
   :body (json/encode data)})

(defn- text-response [text &opt status]
  (default status 200)
  {:status status
   :headers {"Content-Type" "text/plain"}
   :body text})

(defn- parse-json-body [req]
  (when (req :body)
    (json/decode (req :body))))

(defn- match-route [method path routes]
  "Match a request against route patterns. Returns [handler params]."
  (each [m pattern handler] routes
    (when (= m method)
      (def params (route-match pattern path))
      (when params
        (return [handler params]))))
  nil)

(defn- route-match [pattern path]
  "Match a URL pattern with :param placeholders. Returns params table or nil."
  (def pat-segs (string/split "/" pattern))
  (def path-segs (string/split "/" path))
  # Handle wildcard at end
  (def has-wildcard (and (> (length pat-segs) 0)
                        (string/has-prefix? "*" (last pat-segs))))
  (when (and (not has-wildcard)
             (not= (length pat-segs) (length path-segs)))
    (return nil))
  (when (and (not has-wildcard)
             (< (length path-segs) (- (length pat-segs) 1)))
    (return nil))
  (def params @{})
  (for i 0 (min (length pat-segs) (length path-segs))
    (def pat (get pat-segs i))
    (def seg (get path-segs i))
    (cond
      (string/has-prefix? ":" pat)
        (put params (keyword (string/slice pat 1)) seg)
      (string/has-prefix? "*" pat)
        (do
          (put params (keyword (string/slice pat 1))
            (string/join (slice path-segs i) "/"))
          (break))
      (not= pat seg) (return nil)))
  params)

# --- Route builders ---

(defn build-routes [db workspace]
  "Build the route table. Returns array of [method pattern handler]."

  @[# --- Workflows ---
    [:get "/api/workflows"
     (fn [req params]
       (def wf-dirs @[".glitch/workflows" "workflows"])
       (def wfs @[])
       (each dir wf-dirs
         (when (os/stat dir)
           (each f (os/dir dir)
             (when (string/has-suffix? ".janet" f)
               (array/push wfs
                 {:name (string/slice f 0 (- (length f) 6))
                  :path (string dir "/" f)})))))
       (json-response wfs))]

    [:get "/api/workflows/:name"
     (fn [req params]
       (def name (params :name))
       (var found nil)
       (each dir [".glitch/workflows" "workflows"]
         (def p (string dir "/" name ".janet"))
         (when (os/stat p)
           (set found p) (break)))
       (if found
         (json-response {:name name :source (slurp found)})
         (json-response {:error "not found"} 404)))]

    [:post "/api/workflows/:name/run"
     (fn [req params]
       (def body (parse-json-body req))
       (def name (params :name))
       (def input (or (get body "input") ""))
       (var wf-path nil)
       (each dir [".glitch/workflows" "workflows"]
         (def p (string dir "/" name ".janet"))
         (when (os/stat p) (set wf-path p) (break)))
       (unless wf-path
         (return (json-response {:error "workflow not found"} 404)))
       (def run-id (s/record-run db
         {:name name :input input :workflow-file wf-path
          :model "qwen2.5:7b" :workspace ""}))
       # Run async
       (ev/spawn
         (try
           (do
             (def result (runner/run wf-path input :db db))
             (s/finish-run db run-id (result :output) 0 {}))
           ([err]
             (s/finish-run db run-id (string "ERROR: " err) 1 {}))))
       (json-response {:run-id run-id}))]

    # --- Runs ---
    [:get "/api/runs"
     (fn [req params]
       (def q (or (req :query) {}))
       (def runs (s/list-runs db
                   :parent-id (get q "parent_id")
                   :workflow (get q "workflow")))
       (json-response runs))]

    [:get "/api/runs/:id"
     (fn [req params]
       (def id (scan-number (params :id)))
       (def run (s/get-run db id))
       (if run
         (do
           (def steps (s/get-steps db id))
           (json-response (merge run {:steps steps})))
         (json-response {:error "not found"} 404)))]

    [:get "/api/runs/:id/tree"
     (fn [req params]
       (def id (scan-number (params :id)))
       (def run (s/get-run db id))
       (def children (s/list-runs db :parent-id id))
       (json-response (merge (or run {}) {:children children})))]

    # --- Workspace ---
    [:get "/api/workspace"
     (fn [req params]
       (json-response (or workspace {})))]

    [:get "/api/workspace/resources"
     (fn [req params]
       (json-response (get workspace :resources @[])))]

    [:post "/api/workspace/resources"
     (fn [req params]
       (def body (parse-json-body req))
       (when workspace
         (ws/resource (get body "name")
           :type (get body "type")
           :url (get body "url")
           :ref (get body "ref")
           :path (get body "path")))
       (json-response {:ok true}))]

    [:delete "/api/workspace/resources/:name"
     (fn [req params]
       (when workspace
         (def res (ws/get-resource workspace (params :name)))
         (when res
           (def idx (find-index
             |(= ($ :name) (params :name))
             (workspace :resources)))
           (when idx
             (array/remove (workspace :resources) idx))))
       (json-response {:ok true}))]

    [:post "/api/workspace/sync/:name"
     (fn [req params]
       # Sync is a git pull on the resource
       (json-response {:ok true :msg "sync not yet implemented"}))]

    [:post "/api/workspace/pin"
     (fn [req params]
       (json-response {:ok true :msg "pin not yet implemented"}))]

    # --- Providers ---
    [:get "/api/providers"
     (fn [req params]
       (json-response (prov/names)))]])

(defn start [opts]
  "Start the GUI HTTP server."
  (def {:addr addr :workspace ws-path :static-dir static-dir} opts)
  (default addr "localhost:3000")
  (default static-dir "gui/dist")

  (def db (if ws-path
            (s/open-for-workspace ws-path)
            (s/open)))
  (def workspace (when ws-path (ws/load ws-path)))
  (prov/load-providers)

  (def routes (build-routes db workspace))

  (defn handler [req]
    (def method (keyword (string/ascii-lower (req :method))))
    (def path (req :path))

    # Try API routes
    (each [m pattern handler-fn] routes
      (def params (route-match pattern path))
      (when (and (= m method) params)
        (return (handler-fn req params))))

    # Static file serving for SPA
    (def file-path (string static-dir path))
    (if (and (os/stat file-path)
             (= :file ((os/stat file-path) :mode)))
      {:status 200
       :headers {"Content-Type" (mime-type file-path)}
       :body (slurp file-path)}
      # SPA fallback
      (if (os/stat (string static-dir "/index.html"))
        {:status 200
         :headers {"Content-Type" "text/html"}
         :body (slurp (string static-dir "/index.html"))}
        (text-response "not found" 404))))

  (printf "glitch gui → http://%s" addr)
  (def [host port] (string/split ":" addr))
  (http/server handler host (scan-number port)))

(defn- mime-type [path]
  (cond
    (string/has-suffix? ".html" path) "text/html"
    (string/has-suffix? ".js" path) "application/javascript"
    (string/has-suffix? ".css" path) "text/css"
    (string/has-suffix? ".json" path) "application/json"
    (string/has-suffix? ".svg" path) "image/svg+xml"
    (string/has-suffix? ".png" path) "image/png"
    (string/has-suffix? ".ico" path) "image/x-icon"
    "application/octet-stream"))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `gui` suite passes.

- [ ] **Step 5: Wire gui command into CLI**

Add to `janet/src/glitch/main.janet` — update the `ws-gui` function:

```janet
(defn- ws-gui [argv]
  (import glitch/gui)
  (def res
    (argparse
      "Start the workspace GUI"
      "addr" {:kind :option :short "a"
              :help "Listen address" :default "localhost:3000"}
      "static" {:kind :option :short "s"
                :help "Static files directory" :default "gui/dist"}))
  (unless res (os/exit 1))
  (def w (ws/resolve))
  (gui/start {:addr (res "addr")
              :workspace (when w (w :path))
              :static-dir (res "static")}))
```

- [ ] **Step 6: Commit**

```bash
git add janet/src/glitch/gui.janet janet/test/test-gui.janet janet/src/glitch/main.janet
git commit -m "feat(janet): GUI server — REST API + SPA static serving"
```

---

### Task 12: Batch/Compare

**Files:**
- Create: `janet/src/glitch/batch.janet`
- Create: `janet/test/test-batch.janet`

Multi-variant runner with cross-review. Runs workflow with different provider:model combos, then LLM-judges the results.

**Reference:** `internal/pipeline/runner.go` — `ParseCrossReview`, `CrossReviewScore`

- [ ] **Step 1: Write failing tests**

Create `janet/test/test-batch.janet`:

```janet
(use spork/test)
(import glitch/batch :as b)

(start-suite "batch")

# parse-cross-review — numeric format
(def review-output `
VARIANT: local
plan_completeness: 9/10
code_quality: 8/10
total: 17/20

VARIANT: claude
plan_completeness: 7/10
code_quality: 6/10
total: 13/20

WINNER: local
`)

(def scores (b/parse-cross-review review-output))
(assert (= 2 (length scores)) "two variants parsed")
(def local (find |(= ($ :variant) "local") scores))
(assert (= 2 (local :passed)) "local passed 2")
(assert (= 2 (local :total)) "local total 2")
(assert (local :winner) "local is winner")

(def claude (find |(= ($ :variant) "claude") scores))
(assert (= 1 (claude :passed)) "claude passed 1 (7/10 passes, 6/10 fails)")
(assert (not (claude :winner)) "claude is not winner")

# parse-cross-review — pass/fail format
(def pf-output `
--- LOCAL ---
1. Specificity — PASS — good
2. Coverage — FAIL — missing tests
SCORE: 1/2

--- CLAUDE ---
1. Specificity — PASS — ok
2. Coverage — PASS — comprehensive
SCORE: 2/2

WINNER: CLAUDE
`)

(def pf-scores (b/parse-cross-review pf-output))
(assert (= 2 (length pf-scores)) "two variants in pass/fail")
(def pf-claude (find |(= ($ :variant) "claude") pf-scores))
(assert (pf-claude :winner) "claude wins in pass/fail format")
(assert (= 2 (pf-claude :passed)) "claude passed 2")

(end-suite)
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd janet && jpm test
```

- [ ] **Step 3: Implement batch.janet**

Create `janet/src/glitch/batch.janet`:

```janet
# Batch/compare: multi-variant runner + cross-review.
# Replaces: internal/batch/ + ParseCrossReview from runner.go

(import glitch/runner :as runner)
(import glitch/provider :as prov)

(defn parse-cross-review [output]
  "Parse cross-review LLM output into per-variant scores.
   Supports both numeric (VARIANT: / N/M) and pass/fail (--- NAME --- / PASS/FAIL) formats."
  (def upper (string/replace-all "*" "" (string/ascii-upper output)))
  (if (or (string/find "\nVARIANT:" upper)
          (string/has-prefix? "VARIANT:" upper))
    (parse-numeric output)
    (parse-pass-fail output)))

(defn- find-winner [text]
  (def upper (string/ascii-upper text))
  (var winner nil)
  (each line (string/split "\n" upper)
    (def trimmed (string/trim line))
    (when (string/has-prefix? "WINNER:" trimmed)
      (set winner (string/trim (string/slice trimmed 7)))
      (break)))
  winner)

(defn- parse-numeric [output]
  (def lines (string/split "\n" output))
  (def winner (find-winner output))
  (def results @[])
  (var current nil)
  (var passed 0)
  (var total 0)

  (each line lines
    (def trimmed (string/trim line))
    (def upper (string/ascii-upper trimmed))

    (when (string/has-prefix? "VARIANT:" upper)
      # Save previous
      (when (and current (> total 0))
        (array/push results
          {:variant (string/ascii-lower current)
           :passed passed :total total
           :winner (and winner
                     (= (string/ascii-lower winner)
                        (string/ascii-lower current)))}))
      (set current (string/trim (string/slice trimmed 8)))
      (set passed 0)
      (set total 0)
      (continue))

    # Skip meta lines
    (when (or (string/has-prefix? "WINNER:" upper)
              (string/has-prefix? "REASON:" upper)
              (string/has-prefix? "NOTES:" upper)
              (string/has-prefix? "TOTAL:" upper))
      (continue))

    # Parse score lines: "label: N/M"
    (when (and current
               (string/find "/" trimmed)
               (string/find ":" trimmed))
      (def parts (string/split ":" trimmed 0 2))
      (when (= 2 (length parts))
        (def score-part (string/trim (get parts 1)))
        (def num-denom (string/split "/" score-part 0 2))
        (when (= 2 (length num-denom))
          (try
            (do
              (def num (scan-number (string/trim (get num-denom 0))))
              (scan-number (string/trim (get num-denom 1)))
              (++ total)
              (when (>= num 7) (++ passed)))
            ([_] nil))))))

  # Save last
  (when (and current (> total 0))
    (array/push results
      {:variant (string/ascii-lower current)
       :passed passed :total total
       :winner (and winner
                 (= (string/ascii-lower winner)
                    (string/ascii-lower current)))}))
  results)

(defn- parse-pass-fail [output]
  (def upper (string/replace-all "*" "" (string/ascii-upper output)))
  (def lines (string/split "\n" upper))
  (def winner (find-winner output))
  (def results @[])
  (var current nil)
  (var passed 0)
  (var total 0)

  (each line lines
    (def trimmed (string/trim line))

    # Detect --- VARIANT ---
    (when (and (string/has-prefix? "---" trimmed)
               (string/has-suffix? "---" trimmed))
      (when (and current (> total 0))
        (array/push results
          {:variant (string/ascii-lower current)
           :passed passed :total total
           :winner (and winner
                     (string/find (string/ascii-upper current)
                                  (string/ascii-upper winner)))}))
      (def name (string/trim (string/replace-all "-" "" trimmed)))
      (when (not= name "")
        (set current (string/trim name))
        (set passed 0)
        (set total 0))
      (continue))

    # Skip meta lines
    (when (or (string/has-prefix? "SCORE:" trimmed)
              (string/has-prefix? "OVERALL" trimmed)
              (string/has-prefix? "WINNER" trimmed))
      (continue))

    # Count PASS/FAIL
    (when current
      (def has-pass (string/find "PASS" trimmed))
      (def has-fail (string/find "FAIL" trimmed))
      (cond
        (and has-pass (not has-fail)) (do (++ passed) (++ total))
        has-fail (++ total))))

  # Save last
  (when (and current (> total 0))
    (array/push results
      {:variant (string/ascii-lower current)
       :passed passed :total total
       :winner (and winner
                 (string/find (string/ascii-upper current)
                              (string/ascii-upper winner)))}))
  results)

(defn run-compare [wf-path input variants &named db model]
  "Run a workflow with multiple provider:model variants and cross-review.
   variants is an array of strings like 'ollama:qwen2.5:7b'."
  (default model "qwen2.5:7b")
  (def variant-results @{})

  # Run all variants in parallel
  (def fibers
    (seq [v :in variants]
      (def [pname mname] (string/split ":" v 0 2))
      (def vmodel (or mname model))
      [v (ev/spawn
           (runner/run wf-path input
             :db db :model vmodel
             :params @{"variant" v}))]))

  # Gather results (ev/gather was used inline above via spawn)
  # Actually let's use ev/gather properly
  (def results
    (ev/gather
      ;(seq [v :in variants]
         (do
           (def [pname mname] (string/split ":" v 0 2))
           (def vmodel (or mname model))
           (runner/run wf-path input
             :db db :model vmodel
             :params @{"variant" v})))))

  (map (fn [v r] [v r]) variants results))
```

- [ ] **Step 4: Run tests**

```bash
cd janet && jpm test
```

Expected: `batch` suite passes.

- [ ] **Step 5: Commit**

```bash
git add janet/src/glitch/batch.janet janet/test/test-batch.janet
git commit -m "feat(janet): batch/compare — cross-review parsing, multi-variant runner"
```

---

### Task 13: Integration Test — End to End

**Files:**
- Create: `janet/test/test-integration.janet`

End-to-end test: create a workflow file, run it through the runner with store, verify results.

- [ ] **Step 1: Write integration test**

Create `janet/test/test-integration.janet`:

```janet
(use spork/test)
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

# Write a test workflow
(spit (string test-dir "/analyze.janet")
  `(import glitch/core :as g)

   (g/workflow "analyze"
     :description "Test workflow"

     (g/step "data"
       (g/sh "echo" "file1.go file2.go"))

     (g/step "review"
       (g/llm :prompt (string "Review: " (g/ref "data"))
              :provider "mock"))

     (g/step "result"
       (string "Done: " (g/ref "review"))))`)

# Run it
(def result (runner/run
  (string test-dir "/analyze.janet") "test-input"
  :db db))

# Verify result
(assert (= "analyze" (result :name))
        "workflow name captured")
(assert (string/has-prefix? "Done: analyzed:" (result :output))
        "output flows through steps")
(assert (> (result :run-id) 0)
        "run recorded in store")

# Verify store
(def run (s/get-run db (result :run-id)))
(assert (= 0 (run :exit_status))
        "run completed successfully")
(def steps (s/get-steps db (result :run-id)))
(assert (>= (length steps) 1)
        "steps recorded")

# Test par execution
(spit (string test-dir "/par-test.janet")
  `(import glitch/core :as g)

   (g/workflow "par-test"
     (g/par
       (g/step "a" (g/sh "echo" "alpha"))
       (g/step "b" (g/sh "echo" "beta")))
     (g/step "combined"
       (string (g/ref "a") "+" (g/ref "b"))))`)

(def par-result (runner/run
  (string test-dir "/par-test.janet") ""))
(assert (string/find "alpha" (par-result :output))
        "par step a ran")
(assert (not (nil? (g/ref "a"))) "par step a accessible via ref")
(assert (not (nil? (g/ref "b"))) "par step b accessible via ref")

# Cleanup
(s/close db)
(os/shell (string "rm -rf " test-dir))

(end-suite)
```

- [ ] **Step 2: Run all tests**

```bash
cd janet && jpm test
```

Expected: all suites pass.

- [ ] **Step 3: Build final binary and smoke test**

```bash
cd janet && jpm build
./build/glitch version
./build/glitch up
./build/glitch check nonexistent.janet 2>&1 || true
```

- [ ] **Step 4: Commit**

```bash
git add janet/test/test-integration.janet
git commit -m "test(janet): end-to-end integration tests — runner, store, par, providers"
```

---

## Task Dependency Graph

```
Task 1 (scaffold)
  └─ Task 2 (core: step/ref/sh/save)
       └─ Task 3 (core: workflow/par/retry/timeout)
       │    └─ Task 8 (runner)
       │         └─ Task 11 (gui)
       │         └─ Task 12 (batch)
       │         └─ Task 13 (integration)
       └─ Task 4 (http + llm)
       │    └─ Task 5 (providers)
       │         └─ Task 8 (runner)
       └─ Task 9 (stdlib)
  └─ Task 6 (store) ──── Task 8 (runner)
  └─ Task 7 (workspace) ── Task 8 (runner)
  └─ Task 10 (cli) ────── Task 8 (runner)
```

**Parallelizable groups after Task 2:**
- Group A: Task 3 (workflow/par) + Task 4 (http/llm) — independent
- Group B: Task 5 (providers) + Task 6 (store) + Task 7 (workspace) + Task 9 (stdlib) — all independent of each other, depend on Tasks 2-4
- Sequential: Task 8 → Task 10 → Task 11, Task 12, Task 13
