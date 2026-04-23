# MCP Agent Experience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Trim the glitch MCP tool surface to 7 focused tools that give agents capabilities they can't get natively, with improved descriptions and a rebuilt eval handler.

**Architecture:** Remove 3 redundant tools (read_file, search, symbols), rebuild glitch_eval with full glitch.core DSL loaded into SCI, add glitch_list_workflows for workflow discovery, and improve all tool descriptions. Protocol fix (protocolVersion) is already done.

**Tech Stack:** Babashka, SCI, nREPL, Elasticsearch, Cheshire, babashka.process

---

### Task 1: Remove redundant tool definitions and handlers

**Files:**
- Modify: `bb/src/glitch/mcp/tools.clj` (remove 3 tool defs)
- Modify: `bb/src/glitch/mcp/handlers.clj` (remove 3 handlers + dead code)

- [ ] **Step 1: Remove tool definitions from tools.clj**

Remove the `glitch_search`, `glitch_symbols`, and `glitch_read_file` entries from `tool-definitions`. The resulting vector should have 6 entries (the existing `glitch_run`, `glitch_eval`, `glitch_check`, `glitch_search_symbols`, `glitch_search_edges`, `glitch_symbol_context`).

```clojure
;; tools.clj — remove these three entire map entries from tool-definitions:
;; 1. {"name" "glitch_search" ...}      (lines 4-18)
;; 2. {"name" "glitch_symbols" ...}     (lines 20-28)
;; 3. {"name" "glitch_read_file" ...}   (lines 56-62)
```

- [ ] **Step 2: Remove handlers and dead code from handlers.clj**

Delete `handle-search`, `symbol-patterns`, `detect-language`, `handle-symbols`, and `handle-read-file` functions. Remove their entries from `make-handler`.

```clojure
;; handlers.clj — delete these functions entirely:
;; - handle-search         (lines 8-32)
;; - symbol-patterns        (lines 34-40)
;; - detect-language        (lines 42-53)
;; - handle-symbols         (lines 55-69)
;; - handle-read-file       (lines 99-106)

;; In make-handler, remove these three case branches:
;;   "glitch_search"         (handle-search arguments)
;;   "glitch_symbols"        (handle-symbols arguments)
;;   "glitch_read_file"      (handle-read-file arguments)
```

- [ ] **Step 3: Run tests to verify nothing broke**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp.protocol-test]) (let [r (run-tests 'glitch.mcp.protocol-test)] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: 6 tests, 18 assertions, 0 failures

- [ ] **Step 4: Commit**

```bash
git add bb/src/glitch/mcp/tools.clj bb/src/glitch/mcp/handlers.clj
git commit -m "refactor: remove redundant MCP tools (search, symbols, read_file)"
```

---

### Task 2: Update tool descriptions for kept tools

**Files:**
- Modify: `bb/src/glitch/mcp/tools.clj`

- [ ] **Step 1: Update descriptions for all 6 remaining tools**

Replace the description strings for each tool with the agent-optimized versions:

```clojure
;; glitch_run — replace description:
"description" "Execute a glitch workflow file and return its output. Use this to run automation pipelines defined in .glitch/workflows/. Pass input text and/or key-value parameters. Returns the workflow's stdout on success, or the error message on failure."

;; glitch_eval — replace description:
"description" "Evaluate a Clojure expression with the full glitch DSL loaded. Use this to programmatically compose and execute workflow steps, query state, or build pipelines dynamically. Available functions: llm, sh, ref, input, params, param, search, save, read-file, call-workflow, json-extract, validate, validate-schema, gate, consensus, composite-score, search-symbols, search-edges, symbol-context, trace, grounded?"

;; glitch_check — replace description:
"description" "Validate a glitch workflow file for syntax errors without executing it. Returns 'ok' if valid, or a description of the syntax error found."

;; glitch_search_symbols — replace description:
"description" "Search the code intelligence index for symbol definitions (functions, methods, classes, types, structs, interfaces, traits, enums). Supports wildcard matching with *. Use this instead of grep when you need structured symbol metadata across a repository."

;; glitch_search_edges — replace description:
"description" "Query code relationships in the intelligence index: calls, imports, contains, extends, implements, references. Supports depth traversal for multi-hop queries (e.g. what calls the functions that call X). Use this to understand how code connects."

;; glitch_symbol_context — replace description:
"description" "Get a complete picture of a symbol: its definition plus all relationships (callers, callees, parent, children, implementors). Use this when you need to understand a symbol's role in the codebase in one call rather than multiple search_edges queries."
```

- [ ] **Step 2: Commit**

```bash
git add bb/src/glitch/mcp/tools.clj
git commit -m "docs: improve MCP tool descriptions for agent clarity"
```

---

### Task 3: Rebuild glitch_eval with full DSL context

**Files:**
- Modify: `bb/src/glitch/mcp/handlers.clj`
- Create: `bb/test/glitch/mcp/handlers_test.clj`

- [ ] **Step 1: Write failing test for eval with glitch.core DSL**

Create `bb/test/glitch/mcp/handlers_test.clj`:

```clojure
(ns glitch.mcp.handlers-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.mcp.handlers :as handlers]))

(deftest eval-with-dsl-test
  (let [handler (handlers/make-handler {})]
    (testing "basic expression"
      (is (= "42" (handler "glitch_eval" {"expression" "(+ 1 41)"}))))
    (testing "clojure.string available"
      (is (= "HELLO" (handler "glitch_eval" {"expression" "(clojure.string/upper-case \"hello\")"}))))
    (testing "glitch DSL functions are available"
      ;; sh is a glitch.core function that runs a shell command
      (is (string? (handler "glitch_eval" {"expression" "(sh \"echo\" \"hello\")"}))))
    (testing "error returns exception message"
      (is (thrown? Exception (handler "glitch_eval" {"expression" "(/ 1 0)"}))))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp.handlers-test]) (let [r (run-tests 'glitch.mcp.handlers-test)] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: FAIL — `sh` not available in empty SCI sandbox

- [ ] **Step 3: Rebuild handle-eval with glitch.core vars in SCI context**

Replace `handle-eval` in `bb/src/glitch/mcp/handlers.clj`:

```clojure
(defn- handle-eval [arguments]
  (let [expression (get arguments "expression")
        ctx (sci/init {:namespaces
                       {'user
                        {'trace            glitch.core/trace
                         'input            glitch.core/input
                         'params           glitch.core/params
                         'param            glitch.core/param
                         'ref              glitch.core/ref
                         'sh              glitch.core/sh
                         'search           glitch.core/search
                         'save             glitch.core/save
                         'read-file        glitch.core/read-file
                         'write-file       glitch.core/write-file
                         'get-steps        glitch.core/get-steps
                         'last-output      glitch.core/last-output
                         'gate             glitch.core/gate
                         'call-workflow    glitch.core/call-workflow
                         'json-extract     glitch.core/json-extract
                         'validate-schema  glitch.core/validate-schema
                         'validate         glitch.core/validate
                         'llm              glitch.core/llm
                         'grounded?        glitch.core/grounded?
                         'consensus        glitch.core/consensus
                         'composite-score  glitch.core/composite-score
                         'search-symbols   glitch.core/search-symbols
                         'search-edges     glitch.core/search-edges
                         'symbol-context   glitch.core/symbol-context}
                        'clojure.string
                        {'upper-case   clojure.string/upper-case
                         'lower-case   clojure.string/lower-case
                         'trim         clojure.string/trim
                         'split        clojure.string/split
                         'join         clojure.string/join
                         'replace      clojure.string/replace
                         'starts-with? clojure.string/starts-with?
                         'ends-with?   clojure.string/ends-with?
                         'includes?    clojure.string/includes?
                         'blank?       clojure.string/blank?}}})
        result (sci/eval-string* ctx expression)]
    (str result)))
```

Add the require for `glitch.core` at the top of `handlers.clj`:

```clojure
(ns glitch.mcp.handlers
  (:require [babashka.process :as bp]
            [cheshire.core :as json]
            [clojure.string :as str]
            [glitch.core :as g]
            [glitch.index :as index]
            [sci.core :as sci]))
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp.handlers-test]) (let [r (run-tests 'glitch.mcp.handlers-test)] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: 4 tests, 4 assertions, 0 failures

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/mcp/handlers.clj bb/test/glitch/mcp/handlers_test.clj
git commit -m "feat: rebuild glitch_eval with full DSL context in SCI"
```

---

### Task 4: Add glitch_list_workflows tool

**Files:**
- Modify: `bb/src/glitch/mcp/tools.clj` (add tool definition)
- Modify: `bb/src/glitch/mcp/handlers.clj` (add handler)
- Modify: `bb/test/glitch/mcp/handlers_test.clj` (add test)

- [ ] **Step 1: Write failing test for list_workflows**

Add to `bb/test/glitch/mcp/handlers_test.clj`:

```clojure
(ns glitch.mcp.handlers-test
  (:require [clojure.test :refer [deftest is testing]]
            [clojure.java.io :as io]
            [cheshire.core :as json]
            [glitch.mcp.handlers :as handlers]))

;; ... existing tests ...

(deftest list-workflows-test
  (let [handler (handlers/make-handler {})
        ;; Create a temp directory with test workflow files
        tmp-dir (doto (io/file (System/getProperty "java.io.tmpdir")
                               (str "glitch-test-" (System/currentTimeMillis)))
                  (.mkdirs))
        wf1 (io/file tmp-dir "deploy.glitch")
        wf2 (io/file tmp-dir "lint.glitch")]
    (try
      (spit wf1 ";; Deploy the application to staging\n(sh \"echo\" \"deploying\")")
      (spit wf2 "(sh \"echo\" \"linting\")")
      (let [result (handler "glitch_list_workflows"
                            {"path" (.getAbsolutePath tmp-dir)})
            parsed (json/parse-string result true)]
        (testing "returns array of workflow objects"
          (is (= 2 (count parsed))))
        (testing "each entry has name and file"
          (is (every? :name parsed))
          (is (every? :file parsed)))
        (testing "extracts description from first comment"
          (let [deploy (first (filter #(= "deploy" (:name %)) parsed))]
            (is (= "Deploy the application to staging" (:description deploy)))))
        (testing "missing comment gives empty description"
          (let [lint (first (filter #(= "lint" (:name %)) parsed))]
            (is (= "" (:description lint))))))
      (finally
        (.delete wf1)
        (.delete wf2)
        (.delete tmp-dir)))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp.handlers-test]) (let [r (run-tests 'glitch.mcp.handlers-test)] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: FAIL — `glitch_list_workflows` not in case dispatch

- [ ] **Step 3: Add tool definition to tools.clj**

Append to the `tool-definitions` vector in `bb/src/glitch/mcp/tools.clj`:

```clojure
{"name" "glitch_list_workflows"
 "description" "List available glitch workflows in .glitch/workflows/ with their filenames and descriptions. Use this to discover what automation is available before running a workflow."
 "inputSchema"
 {"type" "object"
  "properties"
  {"path" {"type" "string" "description" "Directory to scan (default: .glitch/workflows/)"}}}}
```

- [ ] **Step 4: Add handler to handlers.clj**

Add the handler function and wire it into `make-handler`:

```clojure
(defn- extract-description
  "Extract description from the first ;; comment line of a workflow file."
  [file]
  (try
    (let [lines (str/split-lines (slurp file))]
      (if-let [comment-line (first (filter #(str/starts-with? (str/trim %) ";;") lines))]
        (str/trim (subs (str/trim comment-line) 2))
        ""))
    (catch Exception _ "")))

(defn- handle-list-workflows [arguments]
  (let [path (or (get arguments "path") ".glitch/workflows")
        dir  (java.io.File. path)]
    (if (.isDirectory dir)
      (let [files (->> (.listFiles dir)
                       (filter #(or (str/ends-with? (.getName %) ".glitch")
                                    (str/ends-with? (.getName %) ".clj")))
                       sort)
            entries (mapv (fn [f]
                           {"name"        (str/replace (.getName f) #"\.(glitch|clj)$" "")
                            "file"        (.getAbsolutePath f)
                            "description" (extract-description f)})
                         files)]
        (json/generate-string entries {:pretty true}))
      (json/generate-string [] {:pretty true}))))
```

Add to the `case` in `make-handler`:

```clojure
"glitch_list_workflows" (handle-list-workflows arguments)
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp.handlers-test]) (let [r (run-tests 'glitch.mcp.handlers-test)] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: all tests pass

- [ ] **Step 6: Commit**

```bash
git add bb/src/glitch/mcp/tools.clj bb/src/glitch/mcp/handlers.clj bb/test/glitch/mcp/handlers_test.clj
git commit -m "feat: add glitch_list_workflows MCP tool"
```

---

### Task 5: Update integration test

**Files:**
- Modify: `bb/test/glitch/mcp_test.clj`

- [ ] **Step 1: Update tool count assertion**

Change the expected tool count from 8 to 7 in `mcp-tools-list-test`:

```clojure
;; In mcp-tools-list-test, change:
(is (= 8 (count (get-in resp ["result" "tools"]))))
;; To:
(is (= 7 (count (get-in resp ["result" "tools"]))))
```

- [ ] **Step 2: Run integration test**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp-test]) (let [r (run-tests 'glitch.mcp-test)] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: 2 tests, 2 assertions, 0 failures

- [ ] **Step 3: Commit**

```bash
git add bb/test/glitch/mcp_test.clj
git commit -m "test: update MCP integration test for 7-tool surface"
```

---

### Task 6: Update documentation

**Files:**
- Modify: `site/content/docs/mcp-server.edn`
- Modify: `skills/claude-code/SKILL.md`

- [ ] **Step 1: Update mcp-server.edn**

Replace the "Available tools" section body with the new 7-tool list:

```clojure
:body [[:p "Your editor sees these tools after connecting:"]
       [:ul
        [:li [:code "glitch_run"] " -- execute a workflow file with optional input and parameters."]
        [:li [:code "glitch_eval"] " -- evaluate Clojure with the full glitch DSL loaded (llm, sh, ref, call-workflow, and more)."]
        [:li [:code "glitch_check"] " -- validate a workflow file for syntax errors."]
        [:li [:code "glitch_list_workflows"] " -- list available workflows with descriptions."]
        [:li [:code "glitch_search_symbols"] " -- search indexed symbols by name, kind, or language."]
        [:li [:code "glitch_search_edges"] " -- query code relationships with depth traversal."]
        [:li [:code "glitch_symbol_context"] " -- get a symbol's definition plus all relationships."]]]
```

Also update the "What it does" section body — remove mention of "code search" and "Clojure eval sandbox":

```clojure
:body [[:p "Running " [:code "glitch mcp"] " starts a Model Context Protocol server over stdio. Your AI editor connects to it and gets access to workflow execution, the glitch DSL, code intelligence, and workflow discovery -- all through standard MCP tool calls."]]
```

- [ ] **Step 2: Update SKILL.md MCP section**

Find the MCP section in `skills/claude-code/SKILL.md` (around line 68-69) and update to mention the available tools:

```markdown
# MCP server (for IDE integration)
glitch mcp                                      # start JSON-RPC stdio server
# Tools: glitch_run, glitch_eval, glitch_check, glitch_list_workflows,
#         glitch_search_symbols, glitch_search_edges, glitch_symbol_context
```

- [ ] **Step 3: Commit**

```bash
git add site/content/docs/mcp-server.edn skills/claude-code/SKILL.md
git commit -m "docs: update MCP documentation for 7-tool surface"
```

---

### Task 7: Code review and push

- [ ] **Step 1: Run all MCP tests**

Run: `cd /Users/stokes/Projects/gl1tch/bb && bb -cp "src:test:providers" -e "(require '[clojure.test :refer [run-tests]] '[glitch.mcp.protocol-test] '[glitch.mcp.handlers-test] '[glitch.mcp-test]) (let [r (apply merge-with + (map run-tests ['glitch.mcp.protocol-test 'glitch.mcp.handlers-test 'glitch.mcp-test]))] (System/exit (if (zero? (+ (:fail r) (:error r))) 0 1)))"`

Expected: all tests pass, 0 failures

- [ ] **Step 2: Run code-reviewer agent**

Dispatch the `superpowers:code-reviewer` agent against the full changeset (diff from start of work to HEAD).

- [ ] **Step 3: Fix any review findings**

- [ ] **Step 4: Push to main**

```bash
git push origin main
```
