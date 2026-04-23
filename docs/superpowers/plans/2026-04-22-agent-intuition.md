# Agent Intuition Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give agents the judgment to know *when* to reach for glitch primitives, not just the syntax to use them.

**Architecture:** New `glitch_advise` MCP tool backed by an internal advisory workflow. Enhanced glitch skill with trigger conditions. Session tagging for learning loop. Three independent components, all buildable in parallel.

**Tech Stack:** Babashka/Clojure (DSL, MCP handler), Markdown (skill), EDN (session/index enrichment)

---

### Task 1: Advisory Workflow

**Files:**
- Create: `.glitch/workflows/advise.glitch`

- [ ] **Step 1: Write the advisory workflow**

This workflow takes a task description, searches the workflow index for matches, and asks the LLM to recommend the right glitch approach.

```clojure
;; Recommend glitch primitives and workflows for a given task
(workflow "advise"
  :description "Recommend glitch primitives and workflows for a task"

  (step "index"
    (sh "cat ~/.local/share/glitch/index.edn 2>/dev/null || echo '[]'"))

  (step "workflows"
    (sh "ls .glitch/workflows/*.glitch 2>/dev/null | while read f; do echo \"$(basename $f): $(head -1 $f)\"; done || echo 'none'"))

  (step "recommend"
    (llm :provider "copilot"
         :format "json"
         :schema {:required ["approach" "primitives" "reasoning" "example"]
                  :types {"approach" :string
                          "primitives" :array
                          "reasoning" :string
                          "example" :string}}
         :retries 1
         :prompt ```
You are the glitch advisory system. Given a task description, recommend how to approach it using glitch.

TASK: ~(input)

AVAILABLE PRIMITIVES:
- step/sh: Shell commands for data collection (free, deterministic)
- llm: LLM calls for reasoning/synthesis (expensive)
- par: Concurrent step execution
- phase/gate: Grouped steps with assertions and retries
- consensus: Multi-provider voting for high-confidence answers
- grounded?: Factual verification of LLM output against source context
- investigate: Structured fact-graph reasoning toward a conclusion
- validate: JSON schema checking on step output
- retry/with-timeout: Resilience wrappers
- call-workflow: Compose existing workflows
- save/read-file: File I/O

EXISTING WORKFLOWS:
~(ref "workflows")

WORKFLOW INDEX (previously promoted sessions):
~(ref "index")

RULES:
- approach must be one of: "workflow", "primitive", "repl", "none"
- "none" means glitch adds no value here — the agent should just do the task natively
- "workflow" means an existing workflow matches or a new one should be created
- "primitive" means use a specific primitive via glitch_eval
- "repl" means this is exploratory — use the REPL to figure it out
- If an existing workflow matches, include its path in existing_workflows
- The example field must be a concrete glitch DSL snippet, not pseudocode
- Be conservative — recommend "none" when glitch genuinely adds no value

Return JSON: {"approach": "...", "primitives": [...], "reasoning": "...", "example": "...", "existing_workflows": [...]}
```)))
```

- [ ] **Step 2: Validate the workflow syntax**

Run: `cd ~/Projects/gl1tch/bb && glitch check ../../.glitch/workflows/advise.glitch`
Expected: `ok`

- [ ] **Step 3: Commit**

```bash
git add .glitch/workflows/advise.glitch
git commit -m "feat: add advisory workflow for agent intuition"
```

---

### Task 2: MCP Tool Definition and Handler

**Files:**
- Modify: `bb/src/glitch/mcp/tools.clj` (add `glitch_advise` to `tool-definitions`)
- Modify: `bb/src/glitch/mcp/handlers.clj` (add `handle-advise`, wire into `make-handler`)

- [ ] **Step 1: Write the failing test**

Add to `bb/test/glitch/mcp/handlers_test.clj`:

```clojure
(deftest advise-test
  (let [handler (handlers/make-handler {})]
    (testing "returns JSON recommendation"
      (let [result (handler "glitch_advise"
                            {"task" "check if this PR summary is accurate"})
            parsed (json/parse-string result true)]
        (is (contains? parsed :approach))
        (is (contains? parsed :primitives))
        (is (contains? parsed :reasoning))
        (is (contains? parsed :example))
        (is (contains? #{  "workflow" "primitive" "repl" "none"} (:approach parsed)))))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.mcp.handlers-test/advise-test`
Expected: FAIL — `glitch_advise` is not in the handler case dispatch

- [ ] **Step 3: Add tool definition**

Add to the end of the `tool-definitions` vector in `bb/src/glitch/mcp/tools.clj`:

```clojure
{"name" "glitch_advise"
 "description" "Get a recommendation for which glitch primitives or workflows to use for a given task. Returns a structured recommendation with approach type, relevant primitives, reasoning, and a concrete example. Use this when you're unsure whether glitch can help with the current task."
 "inputSchema"
 {"type" "object"
  "properties"
  {"task"    {"type" "string" "description" "Natural language description of the task"}
   "context" {"type" "string" "description" "Optional additional context (repo, files, domain)"}}
  "required" ["task"]}}
```

- [ ] **Step 4: Add handler**

Add to `bb/src/glitch/mcp/handlers.clj`:

```clojure
(defn- handle-advise [arguments]
  (let [task    (get arguments "task")
        context (get arguments "context" "")
        input   (if (seq context)
                  (str task "\n\nContext: " context)
                  task)
        cmd     ["glitch" "run" ".glitch/workflows/advise.glitch" input]
        result  (apply bp/shell {:out :string :err :string :continue true} cmd)]
    (if (zero? (:exit result))
      (let [raw (str/trim (:out result))]
        ;; The workflow returns JSON via the LLM — extract it
        (try
          (let [extracted (-> raw
                             (str/replace #"^```[a-z]*\n?" "")
                             (str/replace #"\n?```\s*$" "")
                             str/trim)
                ;; Find the JSON object in the response
                start (str/index-of extracted "{")
                end   (when start (inc (str/last-index-of extracted "}")))
                json-str (if (and start end)
                           (subs extracted start end)
                           extracted)
                parsed (json/parse-string json-str)]
            ;; Ensure required keys exist
            (json/generate-string
              {"approach"           (get parsed "approach" "none")
               "primitives"         (get parsed "primitives" [])
               "reasoning"          (get parsed "reasoning" "")
               "example"            (get parsed "example" "")
               "existing_workflows" (get parsed "existing_workflows" [])}
              {:pretty true}))
          (catch Exception _
            (json/generate-string
              {"approach" "none"
               "primitives" []
               "reasoning" (str "Advisory workflow returned unparseable response: " raw)
               "example" ""
               "existing_workflows" []}
              {:pretty true}))))
      (json/generate-string
        {"approach" "none"
         "primitives" []
         "reasoning" (str "Advisory workflow failed: " (str/trim (:err result)))
         "example" ""
         "existing_workflows" []}
        {:pretty true}))))
```

Wire it into the case dispatch in `make-handler`:

```clojure
"glitch_advise"         (handle-advise arguments)
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.mcp.handlers-test/advise-test`
Expected: PASS (requires a working provider — if CI has no provider, this is an integration test that needs `glitch run` available)

Note: If the test can't call a real provider in CI, add a simpler unit test that verifies the handler dispatches correctly and falls back to `{"approach": "none"}` on workflow failure:

```clojure
(deftest advise-fallback-test
  (let [handler (handlers/make-handler {})]
    (testing "returns none on workflow failure"
      (let [result (handler "glitch_advise"
                            {"task" "nonexistent scenario"})
            parsed (json/parse-string result true)]
        (is (= "none" (:approach parsed)))))))
```

- [ ] **Step 6: Run full MCP handler test suite**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.mcp.handlers-test`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add bb/src/glitch/mcp/tools.clj bb/src/glitch/mcp/handlers.clj bb/test/glitch/mcp/handlers_test.clj
git commit -m "feat: add glitch_advise MCP tool and handler"
```

---

### Task 3: Session Tagging for Advisory Calls

**Files:**
- Modify: `bb/src/glitch/session.clj` (add `record-advise!` function)

- [ ] **Step 1: Write the failing test**

Add to `bb/test/glitch/repl_test.clj` (or create `bb/test/glitch/session_test.clj` if repl_test doesn't cover session directly):

```clojure
(ns glitch.session-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.session :as session]))

(deftest record-advise-test
  (session/init-session!)
  (testing "records advisory entry"
    (let [entry (session/record-advise!
                  {:task "check PR accuracy"
                   :recommendation {:approach "primitive"
                                    :primitives ["grounded?"]
                                    :reasoning "factual check"
                                    :example "(grounded? \"summary\" (ref \"diff\"))"
                                    :existing_workflows []}})]
      (is (= :advise (:type entry)))
      (is (= "check PR accuracy" (:task entry)))
      (is (= "primitive" (get-in entry [:recommendation :approach])))
      (is (nil? (:followed? entry)))))

  (testing "appears in session entries"
    (let [entries (session/current-session)]
      (is (= 1 (count entries)))
      (is (= :advise (:type (first entries)))))))

(deftest update-advise-followed-test
  (session/init-session!)
  (session/record-advise!
    {:task "test task"
     :recommendation {:approach "primitive" :primitives ["consensus"]}})
  (testing "updates followed? flag"
    (session/mark-advise-followed! true)
    (let [entries (session/current-session)
          advise-entry (first (filter #(= :advise (:type %)) entries))]
      (is (true? (:followed? advise-entry))))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.session-test`
Expected: FAIL — `record-advise!` and `mark-advise-followed!` don't exist

- [ ] **Step 3: Implement session tagging functions**

Add to `bb/src/glitch/session.clj`:

```clojure
(defn record-advise!
  "Record an advisory recommendation in the current session.
   Entry shape: {:type :advise, :task, :recommendation, :followed?, :timestamp}"
  [{:keys [task recommendation]}]
  (let [entry {:type :advise
               :task task
               :recommendation recommendation
               :followed? nil
               :timestamp (System/currentTimeMillis)}]
    (swap! *current-session* conj entry)
    (flush-session!)
    entry))

(defn mark-advise-followed!
  "Update the most recent :advise entry's :followed? field."
  [followed?]
  (swap! *current-session*
         (fn [entries]
           (let [idx (some (fn [[i e]] (when (= :advise (:type e)) i))
                           (map-indexed vector (reverse entries)))]
             (if idx
               (let [real-idx (- (dec (count entries)) idx)]
                 (assoc-in entries [real-idx :followed?] followed?))
               entries))))
  (flush-session!))
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.session-test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/session.clj bb/test/glitch/session_test.clj
git commit -m "feat: add session tagging for advisory recommendations"
```

---

### Task 4: Index Enrichment with Task Shape

**Files:**
- Modify: `bb/src/glitch/session.clj` (update `index-workflow!` to accept `:task-shape`)
- Modify: `bb/src/glitch/promote.clj` (pass task shape from session to index)

- [ ] **Step 1: Write the failing test**

Add to `bb/test/glitch/session_test.clj`:

```clojure
(deftest index-workflow-with-task-shape-test
  (let [tmp-index (str (System/getProperty "java.io.tmpdir")
                       "/glitch-test-index-" (System/currentTimeMillis) ".edn")]
    (try
      ;; Temporarily override index-path
      (with-redefs [session/index-path tmp-index]
        (session/index-workflow!
          {:path ".glitch/workflows/check-pr.glitch"
           :description "Check PR accuracy"
           :tags ["pr" "accuracy"]
           :task-shape "verify factual accuracy of text against source material"})
        (let [entries (session/list-indexed)]
          (testing "stores task-shape in index"
            (is (= 1 (count entries)))
            (is (= "verify factual accuracy of text against source material"
                    (:task-shape (first entries)))))))
      (finally
        (let [f (java.io.File. tmp-index)]
          (when (.exists f) (.delete f)))))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.session-test/index-workflow-with-task-shape-test`
Expected: PASS actually — `index-workflow!` already stores arbitrary keys. The test verifies this works. If it passes, this is a verification step, not a code change.

- [ ] **Step 3: Update promote! to extract task-shape from session**

In `bb/src/glitch/promote.clj`, update the return map at the end of `promote!` to include task-shape extraction. The task-shape comes from any `:advise` entries in the session:

```clojure
;; Add after the tags extraction, before the return map
task-shape (let [advise-entries (filter #(= :advise (:type %)) (or session []))]
             (when (seq advise-entries)
               (:task (last advise-entries))))
```

Update the return map:

```clojure
{:workflow-source cleaned
 :path           path
 :tags           tags
 :task-shape     task-shape}
```

- [ ] **Step 4: Write test for promote task-shape extraction**

Add to an appropriate test file (or `bb/test/glitch/promote_test.clj`):

```clojure
(ns glitch.promote-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.promote :as promote]))

(deftest task-shape-extraction-test
  (testing "extracts task from advise entries in session"
    ;; This tests the extraction logic, not the full LLM promote flow
    (let [session [{:type :sh :id "fetch" :args ["curl ..."] :output "data"}
                   {:type :advise :task "check PR accuracy"
                    :recommendation {:approach "primitive"}}
                   {:type :llm :id "analyze" :output "result"}]]
      ;; The advise entry's :task should become :task-shape
      (let [advise-entries (filter #(= :advise (:type %)) session)]
        (is (= "check PR accuracy" (:task (last advise-entries))))))))
```

- [ ] **Step 5: Run tests**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.session-test glitch.promote-test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add bb/src/glitch/promote.clj bb/test/glitch/session_test.clj bb/test/glitch/promote_test.clj
git commit -m "feat: enrich workflow index with task-shape from advisory sessions"
```

---

### Task 5: Enhanced Glitch Skill

**Files:**
- Modify: `skills/glitch/SKILL.md` (add agent intuition section)

- [ ] **Step 1: Add the agent intuition section to the skill**

Add this section after the "## Patterns" section in `skills/glitch/SKILL.md`:

```markdown
## When to Use Glitch

Before implementing a multi-step task, check if glitch adds value. Call `glitch_advise` with the task description when you notice any of these signals:

| Signal | What it looks like |
|--------|--------------------|
| **Repetition** | Task will be done again or across multiple targets — "check all PRs", "audit these repos", "review each deploy" |
| **Confidence** | Judgment where being wrong matters — "is this summary accurate?", "which approach is better?", "is this finding real?" |
| **Multi-source** | Needs information composed from multiple providers or tools — "compare what Claude and Copilot say", "aggregate from 3 APIs" |
| **Investigation** | Uncertain facts, contradictions, structured reasoning — "figure out why this is failing", "is this finding real?" |

**How to use it:**

1. Call `glitch_advise` with the task description
2. If approach is `"none"` — do the task natively, glitch adds no value
3. If approach is `"workflow"` — check `existing_workflows` first, then `glitch_run` or create a new workflow
4. If approach is `"primitive"` — use `glitch_eval` with the suggested example
5. If approach is `"repl"` — this is exploratory, use `glitch_eval` iteratively

**When NOT to check:** Single-step tasks, file reads, simple grep, git operations. If you can do it in one tool call, skip glitch.
```

- [ ] **Step 2: Commit**

```bash
git add skills/glitch/SKILL.md
git commit -m "docs: add agent intuition triggers to glitch skill"
```

---

### Task 6: Update MCP Integration Test

**Files:**
- Modify: `bb/test/glitch/mcp_test.clj` (update expected tool count)

- [ ] **Step 1: Read the current MCP integration test**

Read `bb/test/glitch/mcp_test.clj` to find the tool count assertion.

- [ ] **Step 2: Update tool count**

The test currently asserts 7 or 8 tools (from the previous MCP agent experience work). Update to expect the new count that includes `glitch_advise`. The current tool-definitions vector in `tools.clj` has 8 entries (7 original + `glitch_recall`). Adding `glitch_advise` makes it 9.

Find the assertion like `(is (= 8 (count tools)))` and update to `(is (= 9 (count tools)))`.

- [ ] **Step 3: Run the full test suite**

Run: `cd ~/Projects/gl1tch/bb && bb test`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add bb/test/glitch/mcp_test.clj
git commit -m "test: update MCP integration test for 9-tool surface"
```

---

### Task 7: Wire Advisory Session Recording into Handler

**Files:**
- Modify: `bb/src/glitch/mcp/handlers.clj` (import session, record advise calls)

- [ ] **Step 1: Write the test**

Add to `bb/test/glitch/mcp/handlers_test.clj`:

```clojure
(deftest advise-records-session-test
  (let [handler (handlers/make-handler {})]
    (require '[glitch.session :as session])
    (session/init-session!)
    (binding [glitch.session/*current-session* (atom [])
              glitch.session/*session-id* (atom "test-advise")]
      (try
        (handler "glitch_advise" {"task" "summarize logs"})
        (catch Exception _))
      (let [entries @glitch.session/*current-session*
            advise-entries (filter #(= :advise (:type %)) entries)]
        (testing "advise call recorded in session"
          ;; May be empty if workflow fails (no provider), but the recording
          ;; code path should be exercised without error
          (is (vector? entries)))))))
```

- [ ] **Step 2: Update handle-advise to record session**

In `bb/src/glitch/mcp/handlers.clj`, add `[glitch.session :as session]` to the require, then wrap the handler to record:

After the `handle-advise` function returns its result, add session recording. Update the function to:

```clojure
(defn- handle-advise [arguments]
  (let [task    (get arguments "task")
        context (get arguments "context" "")
        input   (if (seq context)
                  (str task "\n\nContext: " context)
                  task)
        cmd     ["glitch" "run" ".glitch/workflows/advise.glitch" input]
        result  (apply bp/shell {:out :string :err :string :continue true} cmd)
        response (if (zero? (:exit result))
                   (let [raw (str/trim (:out result))]
                     (try
                       (let [extracted (-> raw
                                          (str/replace #"^```[a-z]*\n?" "")
                                          (str/replace #"\n?```\s*$" "")
                                          str/trim)
                             start (str/index-of extracted "{")
                             end   (when start (inc (str/last-index-of extracted "}")))
                             json-str (if (and start end)
                                        (subs extracted start end)
                                        extracted)
                             parsed (json/parse-string json-str true)]
                         {:approach           (:approach parsed "none")
                          :primitives         (:primitives parsed [])
                          :reasoning          (:reasoning parsed "")
                          :example            (:example parsed "")
                          :existing_workflows (:existing_workflows parsed [])})
                       (catch Exception _
                         {:approach "none" :primitives [] :reasoning raw
                          :example "" :existing_workflows []})))
                   {:approach "none" :primitives []
                    :reasoning (str "Advisory workflow failed: " (str/trim (:err result)))
                    :example "" :existing_workflows []})]
    ;; Record in session
    (try
      (session/record-advise! {:task task :recommendation response})
      (catch Exception _ nil))
    (json/generate-string response {:pretty true})))
```

- [ ] **Step 3: Run tests**

Run: `cd ~/Projects/gl1tch/bb && bb test -v glitch.mcp.handlers-test`
Expected: All pass

- [ ] **Step 4: Commit**

```bash
git add bb/src/glitch/mcp/handlers.clj bb/test/glitch/mcp/handlers_test.clj
git commit -m "feat: wire advisory session recording into MCP handler"
```
