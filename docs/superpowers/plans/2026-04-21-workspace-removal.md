# Workspace Removal & Shebang/REPL Model — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the workspace/project concept from glitch. Workflows become standalone files runnable via shebang or REPL. Replace SQLite with Datascript, FTS5 search with ripgrep.

**Architecture:** Surgical removal — delete workspace-related code, rewrite `glitch.store` against Datascript, simplify the runner and CLI, add `glitch repl`, and replace MCP search tools with ripgrep wrappers.

**Tech Stack:** Babashka, Datascript (`:deps` in bb.edn), ripgrep (`rg --json`), babashka.nrepl

**Spec:** `docs/superpowers/specs/2026-04-21-workspace-removal-design.md`

---

## File Map

**Delete entirely:**
- `bb/src/glitch/project.clj`
- `bb/src/glitch/gui/workspace.clj`
- `bb/src/glitch/mcp/indexer.clj`
- `bb/src/glitch/mcp/search.clj`
- `bb/src/glitch/mcp/embeddings.clj`
- `bb/src/glitch/mcp/vecmath.clj`
- `bb/test/glitch/project_test.clj`
- `bb/test/glitch/mcp/indexer_test.clj`
- `bb/test/glitch/mcp/search_test.clj`
- `bb/test/glitch/mcp/vecmath_test.clj`

**Rewrite:**
- `bb/src/glitch/store.clj` — SQLite → Datascript
- `bb/test/glitch/store_test.clj` — new tests for Datascript store
- `bb/src/glitch/mcp/handlers.clj` — remove workspace tools, add ripgrep tools
- `bb/src/glitch/mcp/tools.clj` — updated tool definitions

**Modify:**
- `bb/bb.edn` — swap pod for dep, add `rg` to required tools
- `bb/src/glitch/main.clj` — simplified CLI, new repl command
- `bb/src/glitch/runner.clj` — remove workspace threading, add `:provider` kwarg
- `bb/src/glitch/mcp.clj` — remove workspace init
- `bb/src/glitch/test_runner.clj` — remove project-test, add new test ns

**Create:**
- `bb/src/glitch/repl.clj` — nREPL server with DSL preloaded
- `bb/test/glitch/repl_test.clj` — REPL startup tests

---

### Task 1: Update bb.edn — swap SQLite pod for Datascript dep

**Files:**
- Modify: `bb/bb.edn`

- [ ] **Step 1: Update bb.edn**

Replace the SQLite pod with Datascript dependency:

```edn
{:paths ["src" "providers"]
 :deps {datascript/datascript {:mvn/version "1.7.3"}}
 :tasks
 {build {:doc "Build the glitch uberscript"
         :task (do (babashka.fs/create-dirs "build")
                   (shell "bb uberscript build/glitch --main glitch.main")
                   (spit "build/glitch.tmp"
                     (str "#!/usr/bin/env bb\n" (slurp "build/glitch")))
                   (babashka.fs/move "build/glitch.tmp" "build/glitch" {:replace-existing true})
                   (shell "chmod +x build/glitch"))}
  install {:doc "Install glitch to /usr/local/bin"
           :depends [build]
           :task (do (shell "cp build/glitch /usr/local/bin/glitch")
                     (babashka.fs/create-dirs
                       (str (System/getProperty "user.home") "/.config/glitch/providers"))
                     (doseq [f (babashka.fs/glob "providers" "*.clj")]
                       (babashka.fs/copy f
                         (str (System/getProperty "user.home") "/.config/glitch/providers/"
                              (babashka.fs/file-name f))
                         {:replace-existing true}))
                     (println "installed providers to ~/.config/glitch/providers/"))}
  test {:doc "Run tests"
        :task (shell "bb -cp src:test:providers -m glitch.test-runner")}
  clean {:doc "Clean build artifacts"
         :task (babashka.fs/delete-tree "build")}
  site:dev {:doc "Run Astro dev server"
            :task (shell {:dir "../site"} "npm run dev")}
  site:build {:doc "Build the 8op.org site"
              :task (shell {:dir "../site"} "npm run build")}}}
```

- [ ] **Step 2: Verify datascript loads**

Run: `cd bb && bb -e '(require (quote datascript.core)) (println "ok")'`
Expected: `ok`

- [ ] **Step 3: Commit**

```bash
git add bb/bb.edn
git commit -m "chore: replace SQLite pod with Datascript dependency"
```

---

### Task 2: Rewrite glitch.store — Datascript backend

**Files:**
- Rewrite: `bb/src/glitch/store.clj`
- Rewrite: `bb/test/glitch/store_test.clj`

- [ ] **Step 1: Write store tests**

```clojure
(ns glitch.store-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.store :as store]))

(def ^:dynamic *db* nil)

(defn temp-db-fixture [f]
  (let [path (str "/tmp/glitch-test-" (System/currentTimeMillis) ".edn")
        db   (store/open path)]
    (binding [*db* db]
      (try
        (f)
        (finally
          (store/close db)
          (let [file (java.io.File. path)]
            (when (.exists file) (.delete file))))))))

(use-fixtures :each temp-db-fixture)

(deftest test-open-creates-db
  (testing "open returns a store map with :conn and :path"
    (is (some? (:conn *db*)))
    (is (some? (:path *db*)))))

(deftest test-record-run-and-get-run
  (testing "insert and retrieve a run"
    (let [id (store/record-run *db*
               {:name          "test-workflow"
                :input         "hello world"
                :workflow-file "test.glitch"
                :model         "gemma4"})
          run (store/get-run *db* id)]
      (is (integer? id))
      (is (= "test-workflow" (:run/name run)))
      (is (= "hello world" (:run/input run)))
      (is (= "test.glitch" (:run/workflow run)))
      (is (= "gemma4" (:run/model run)))
      (is (integer? (:run/started-at run)))
      (is (nil? (:run/output run))))))

(deftest test-record-step-and-get-steps
  (testing "insert steps and retrieve by run-id"
    (let [run-id (store/record-run *db* {:name "step-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id    run-id
                    :step-id   "s1"
                    :prompt    "do the thing"
                    :output    "done"
                    :model     "gemma4"
                    :duration  150
                    :kind      "llm"})
          _      (store/record-step *db*
                   {:run-id    run-id
                    :step-id   "s2"
                    :prompt    "next thing"
                    :output    "ok"
                    :model     "gemma4"
                    :duration  80
                    :kind      "shell"})
          steps  (store/get-steps *db* run-id)]
      (is (= 2 (count steps)))
      (is (= #{"s1" "s2"} (set (map :step/id steps))))
      (is (= "do the thing" (:step/prompt (first (filter #(= "s1" (:step/id %)) steps)))))
      (is (= 150 (:step/duration (first (filter #(= "s1" (:step/id %)) steps))))))))

(deftest test-finish-run
  (testing "finish-run updates output and exit status"
    (let [id  (store/record-run *db* {:name "finish-test" :input "x"})
          _   (store/finish-run *db* id "all done" 0
                {:tokens-in 500 :tokens-out 200 :cost 0.02})
          run (store/get-run *db* id)]
      (is (= "all done" (:run/output run)))
      (is (= 0 (:run/exit run)))
      (is (integer? (:run/finished-at run))))))

(deftest test-list-runs
  (testing "list-runs returns runs in descending order with limit"
    (dotimes [i 5]
      (store/record-run *db* {:name (str "run-" i) :input (str i)}))
    (let [runs (store/list-runs *db* :limit 3)]
      (is (= 3 (count runs))))))

(deftest test-list-runs-workflow-filter
  (testing "list-runs with workflow filter"
    (store/record-run *db* {:name "a" :input "x" :workflow-file "deploy.glitch"})
    (store/record-run *db* {:name "b" :input "y" :workflow-file "deploy.glitch"})
    (store/record-run *db* {:name "c" :input "z" :workflow-file "other.glitch"})
    (let [runs (store/list-runs *db* :workflow "deploy.glitch")]
      (is (= 2 (count runs)))
      (is (every? #(= "deploy.glitch" (:run/workflow %)) runs)))))

(deftest test-step-confidence-column
  (testing "record-step persists confidence value"
    (let [run-id (store/record-run *db* {:name "conf-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "done"
                    :kind "llm" :confidence 0.85})
          steps  (store/get-steps *db* run-id)]
      (is (= 0.85 (:step/confidence (first steps)))))))

(deftest test-step-upsert
  (testing "recording a step with same run-id+step-id replaces the old row"
    (let [run-id (store/record-run *db* {:name "upsert-test" :input "x"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "first"})
          _      (store/record-step *db*
                   {:run-id run-id :step-id "s1" :output "second"})
          steps  (store/get-steps *db* run-id)]
      (is (= 1 (count steps)))
      (is (= "second" (:step/output (first steps)))))))

(deftest test-record-fact-and-edges
  (testing "facts and edges are recorded and retrieved"
    (let [run-id (store/record-run *db* {:name "fact-test" :input "x"})
          _      (store/record-fact *db*
                   {:id "f1" :run-id run-id :claim "the sky is blue"
                    :confidence 0.9 :status :approved})
          _      (store/record-fact *db*
                   {:id "f2" :run-id run-id :claim "grass is green"
                    :confidence 0.8 :status :unapproved})
          _      (store/record-fact-edge *db*
                   {:run-id run-id :from-id "f1" :to-id "f2"
                    :rel :supports :weight 0.7})
          facts  (store/get-facts *db* run-id)
          edges  (store/get-fact-edges *db* run-id)]
      (is (= 2 (count facts)))
      (is (= 1 (count edges))))))

(deftest test-persistence
  (testing "data survives close and reopen"
    (let [path (str "/tmp/glitch-persist-test-" (System/currentTimeMillis) ".edn")
          db1  (store/open path)
          id   (store/record-run db1 {:name "persist" :input "x"})]
      (store/finish-run db1 id "done" 0)
      (store/close db1)
      (let [db2 (store/open path)
            run (store/get-run db2 id)]
        (try
          (is (= "done" (:run/output run)))
          (finally
            (store/close db2)
            (.delete (java.io.File. path))))))))
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd bb && bb -cp src:test:providers -e '(require (quote glitch.store-test)) (clojure.test/run-tests (quote glitch.store-test))'`
Expected: Compilation failure — store API doesn't match yet.

- [ ] **Step 3: Write the Datascript store implementation**

```clojure
(ns glitch.store
  "Datascript-backed run/step/fact recording.
   Persists to a single EDN file. No pods, no SQLite."
  (:require [datascript.core :as d]
            [clojure.java.io :as io]
            [clojure.edn :as edn]))

;; ---------------------------------------------------------------------------
;; Schema
;; ---------------------------------------------------------------------------

(def schema
  {:run/id          {:db/unique :db.unique/identity}
   :run/workflow    {}
   :run/name        {}
   :run/input       {}
   :run/output      {}
   :run/exit        {}
   :run/model       {}
   :run/started-at  {}
   :run/finished-at {}
   :run/parent-id   {}
   :run/tokens-in   {}
   :run/tokens-out  {}
   :run/cost        {}
   :step/run-id     {}
   :step/id         {}
   :step/prompt     {}
   :step/output     {}
   :step/model      {}
   :step/duration   {}
   :step/kind       {}
   :step/exit       {}
   :step/tokens-in  {}
   :step/tokens-out {}
   :step/gate       {}
   :step/artifacts  {}
   :step/confidence {}
   :fact/id         {:db/unique :db.unique/identity}
   :fact/run-id     {}
   :fact/claim      {}
   :fact/confidence {}
   :fact/status     {}
   :fact/source-step {}
   :fact/source-prov {}
   :fact/tokens-cost {}
   :fact/created-at  {}
   :edge/run-id     {}
   :edge/from-id    {}
   :edge/to-id      {}
   :edge/rel        {}
   :edge/weight     {}
   :edge/source     {}})

;; ---------------------------------------------------------------------------
;; ID generation
;; ---------------------------------------------------------------------------

(def ^:private id-counter (atom 0))

(defn- next-id []
  (swap! id-counter inc))

;; ---------------------------------------------------------------------------
;; Persistence
;; ---------------------------------------------------------------------------

(defn- flush-db!
  "Serialize the current DB state to disk."
  [{:keys [conn path]}]
  (let [parent (.getParentFile (io/file path))]
    (when (and parent (not (.exists parent)))
      (.mkdirs parent)))
  (spit path (pr-str (d/datoms @conn :eavt))))

(defn- load-datoms
  "Read persisted datoms from EDN file."
  [path]
  (when (.exists (io/file path))
    (let [raw (edn/read-string (slurp path))]
      (when (seq raw) raw))))

;; ---------------------------------------------------------------------------
;; Open / Close
;; ---------------------------------------------------------------------------

(defn open
  "Open or create a glitch database at `path`
   (default ~/.local/share/glitch/glitch.edn).
   Returns a store map {:conn <datascript-conn> :path <string>}."
  [& [path]]
  (let [path (or path
                 (str (System/getProperty "user.home")
                      "/.local/share/glitch/glitch.edn"))
        parent (.getParentFile (io/file path))]
    (when (and parent (not (.exists parent)))
      (.mkdirs parent))
    (let [conn (if-let [datoms (load-datoms path)]
                 (let [db (d/init-db datoms schema)]
                   ;; Set id-counter to max existing run/id
                   (let [max-id (or (d/q '[:find (max ?id) .
                                           :where [_ :run/id ?id]]
                                         db)
                                    0)]
                     (reset! id-counter max-id))
                   (d/conn-from-db db))
                 (do
                   (reset! id-counter 0)
                   (d/create-conn schema)))]
      {:conn conn :path path})))

(defn close
  "Flush and release. After close the store should not be used."
  [store]
  (when store
    (flush-db! store)))

;; ---------------------------------------------------------------------------
;; Runs
;; ---------------------------------------------------------------------------

(defn record-run
  "Insert a run record. Returns the integer run ID."
  [store rec]
  (let [id  (next-id)
        now (quot (System/currentTimeMillis) 1000)
        tx  [{:run/id          id
               :run/name        (or (:name rec) "")
               :run/input       (:input rec)
               :run/workflow    (:workflow-file rec)
               :run/model       (:model rec)
               :run/started-at  now
               :run/parent-id   (:parent-run-id rec)}]]
    (d/transact! (:conn store) tx)
    id))

(defn finish-run
  "Update a run with final output, exit status, and optional totals."
  [store id output exit-status & [totals]]
  (let [now (quot (System/currentTimeMillis) 1000)
        eid (d/q '[:find ?e .
                   :in $ ?id
                   :where [?e :run/id ?id]]
                 @(:conn store) id)]
    (when eid
      (d/transact! (:conn store)
        [(cond-> {:db/id eid
                  :run/output output
                  :run/exit exit-status
                  :run/finished-at now}
           (:tokens-in totals)  (assoc :run/tokens-in (:tokens-in totals))
           (:tokens-out totals) (assoc :run/tokens-out (:tokens-out totals))
           (:cost totals)       (assoc :run/cost (:cost totals)))])
      (flush-db! store))))

(defn get-run
  "Fetch a single run by ID. Returns a map or nil."
  [store id]
  (when-let [eid (d/q '[:find ?e .
                         :in $ ?id
                         :where [?e :run/id ?id]]
                       @(:conn store) id)]
    (d/pull @(:conn store) '[*] eid)))

(defn list-runs
  "List runs with optional filters. Options: :parent-id, :workflow, :limit (default 50)."
  [store & {:keys [parent-id workflow limit] :or {limit 50}}]
  (let [db @(:conn store)
        eids (cond
               parent-id
               (d/q '[:find [?e ...]
                       :in $ ?pid
                       :where [?e :run/parent-id ?pid]]
                     db parent-id)

               workflow
               (d/q '[:find [?e ...]
                       :in $ ?wf
                       :where [?e :run/workflow ?wf]]
                     db workflow)

               :else
               (d/q '[:find [?e ...]
                       :where [?e :run/id _]]
                     db))
        runs (->> eids
                  (map #(d/pull db '[*] %))
                  (sort-by :run/id #(compare %2 %1))
                  (take limit)
                  vec)]
    runs))

;; ---------------------------------------------------------------------------
;; Steps
;; ---------------------------------------------------------------------------

(defn record-step
  "Insert or replace a step record."
  [store rec]
  (let [db @(:conn store)
        ;; Check for existing step with same run-id + step-id (upsert)
        existing (d/q '[:find ?e .
                         :in $ ?rid ?sid
                         :where
                         [?e :step/run-id ?rid]
                         [?e :step/id ?sid]]
                       db (:run-id rec) (:step-id rec))
        tx (cond-> {:step/run-id    (:run-id rec)
                    :step/id        (:step-id rec)
                    :step/prompt    (:prompt rec)
                    :step/output    (:output rec)
                    :step/model     (:model rec)
                    :step/duration  (:duration rec)
                    :step/kind      (:kind rec)
                    :step/exit      (:exit rec)
                    :step/tokens-in  (:tokens-in rec)
                    :step/tokens-out (:tokens-out rec)
                    :step/gate      (:gate rec)
                    :step/artifacts (:artifacts rec)
                    :step/confidence (:confidence rec)}
             existing (assoc :db/id existing))]
    (d/transact! (:conn store) [tx])))

(defn get-steps
  "Fetch all steps for a run."
  [store run-id]
  (let [db @(:conn store)
        eids (d/q '[:find [?e ...]
                     :in $ ?rid
                     :where [?e :step/run-id ?rid]]
                   db run-id)]
    (->> eids
         (map #(d/pull db '[*] %))
         (sort-by :db/id)
         vec)))

;; ---------------------------------------------------------------------------
;; Facts
;; ---------------------------------------------------------------------------

(defn record-fact
  "Insert a fact record."
  [store rec]
  (let [now (quot (System/currentTimeMillis) 1000)]
    (d/transact! (:conn store)
      [{:fact/id         (:id rec)
        :fact/run-id     (:run-id rec)
        :fact/claim      (:claim rec)
        :fact/confidence (:confidence rec)
        :fact/status     (or (some-> (:status rec) name) "unapproved")
        :fact/source-step (:source-step rec)
        :fact/source-prov (:source-prov rec)
        :fact/tokens-cost (or (:tokens-cost rec) 0)
        :fact/created-at  now}])))

(defn record-fact-edge
  "Insert a fact edge record."
  [store rec]
  (d/transact! (:conn store)
    [{:edge/run-id  (:run-id rec)
      :edge/from-id (:from-id rec)
      :edge/to-id   (:to-id rec)
      :edge/rel     (or (some-> (:rel rec) name) (str (:rel rec)))
      :edge/weight  (or (:weight rec) 1.0)
      :edge/source  (:source rec)}]))

(defn get-facts
  "Fetch all facts for a run."
  [store run-id]
  (let [db @(:conn store)
        eids (d/q '[:find [?e ...]
                     :in $ ?rid
                     :where [?e :fact/run-id ?rid]]
                   db run-id)]
    (->> eids
         (map #(d/pull db '[*] %))
         (sort-by :fact/created-at)
         vec)))

(defn get-fact-edges
  "Fetch all fact edges for a run."
  [store run-id]
  (let [db @(:conn store)
        eids (d/q '[:find [?e ...]
                     :in $ ?rid
                     :where [?e :edge/run-id ?rid]]
                   db run-id)]
    (->> eids
         (map #(d/pull db '[*] %))
         (sort-by :db/id)
         vec)))
```

- [ ] **Step 4: Run store tests**

Run: `cd bb && bb -cp src:test:providers -e '(require (quote glitch.store-test)) (clojure.test/run-tests (quote glitch.store-test))'`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add bb/src/glitch/store.clj bb/test/glitch/store_test.clj
git commit -m "feat: rewrite store with Datascript backend

Replaces SQLite pod with pure Clojure Datascript.
Single EDN file at ~/.local/share/glitch/glitch.edn.
Same public API, no workspace column."
```

---

### Task 3: Delete workspace files and update test runner

**Files:**
- Delete: `bb/src/glitch/project.clj`
- Delete: `bb/src/glitch/gui/workspace.clj`
- Delete: `bb/src/glitch/mcp/indexer.clj`
- Delete: `bb/src/glitch/mcp/search.clj`
- Delete: `bb/src/glitch/mcp/embeddings.clj`
- Delete: `bb/src/glitch/mcp/vecmath.clj`
- Delete: `bb/test/glitch/project_test.clj`
- Delete: `bb/test/glitch/mcp/indexer_test.clj`
- Delete: `bb/test/glitch/mcp/search_test.clj`
- Delete: `bb/test/glitch/mcp/vecmath_test.clj`
- Modify: `bb/src/glitch/test_runner.clj`

- [ ] **Step 1: Delete workspace-related files**

```bash
cd bb
rm -f src/glitch/project.clj
rm -f src/glitch/gui/workspace.clj
rm -f src/glitch/mcp/indexer.clj
rm -f src/glitch/mcp/search.clj
rm -f src/glitch/mcp/embeddings.clj
rm -f src/glitch/mcp/vecmath.clj
rm -f test/glitch/project_test.clj
rm -f test/glitch/mcp/indexer_test.clj
rm -f test/glitch/mcp/search_test.clj
rm -f test/glitch/mcp/vecmath_test.clj
```

- [ ] **Step 2: Update test runner — remove project-test**

```clojure
(ns glitch.test-runner
  "Entry point for `bb test` — discovers and runs all test namespaces."
  (:require [clojure.test :as t]
            [glitch.core-test]
            [glitch.confidence-test]
            [glitch.provider-test]
            [glitch.store-test]
            [glitch.runner-test]
            [glitch.graph-test]))

(defn -main [& _]
  (let [results (mapv #(t/run-tests %)
                  ['glitch.core-test
                   'glitch.confidence-test
                   'glitch.provider-test
                   'glitch.store-test
                   'glitch.runner-test
                   'glitch.graph-test])
        total-fail (apply + (map :fail results))
        total-err  (apply + (map :error results))]
    (println (str "\n=== " (apply + (map :test results)) " tests, "
                  (apply + (map :pass results)) " passed, "
                  total-fail " failures, " total-err " errors ==="))
    (System/exit (if (and (zero? total-fail) (zero? total-err)) 0 1))))
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: delete workspace/project files and related tests

Removes project.clj, gui/workspace.clj, mcp/indexer.clj,
mcp/search.clj, mcp/embeddings.clj, mcp/vecmath.clj and
their tests. Updates test runner."
```

---

### Task 4: Simplify main.clj — remove workspace, add provider flag

**Files:**
- Modify: `bb/src/glitch/main.clj`

- [ ] **Step 1: Rewrite main.clj**

```clojure
(ns glitch.main
  (:require [glitch.runner :as runner]
            [glitch.store :as store]
            [glitch.provider :as prov]
            [clojure.string :as str]
            [clojure.java.io :as io]))

;; ---------------------------------------------------------------------------
;; Arg parsing
;; ---------------------------------------------------------------------------

(defn parse-args
  "Parse CLI args. valid-opts is a map of flag-name -> {:short :kind}
   :kind can be :option (takes value), :flag (boolean), :accumulate (collects values).
   Returns {:opts {} :positional []}"
  [args valid-opts]
  (let [lookup (reduce-kv
                 (fn [m k spec]
                   (let [long-flag  (str "--" (name k))
                         short-flag (when (:short spec) (str "-" (:short spec)))]
                     (cond-> (assoc m long-flag k)
                       short-flag (assoc short-flag k))))
                 {} valid-opts)]
    (loop [remaining (vec args)
           opts      {}
           positional []]
      (if (empty? remaining)
        {:opts opts :positional positional}
        (let [arg (first remaining)
              key (get lookup arg)]
          (if key
            (let [kind (get-in valid-opts [key :kind] :option)]
              (case kind
                :flag
                (recur (subvec remaining 1)
                       (assoc opts key true)
                       positional)

                :accumulate
                (if (< 1 (count remaining))
                  (recur (subvec remaining 2)
                         (update opts key (fnil conj []) (second remaining))
                         positional)
                  (recur (subvec remaining 1) opts positional))

                ;; :option (default)
                (if (< 1 (count remaining))
                  (recur (subvec remaining 2)
                         (assoc opts key (second remaining))
                         positional)
                  (recur (subvec remaining 1) opts positional))))
            (recur (subvec remaining 1) opts (conj positional arg))))))))

;; ---------------------------------------------------------------------------
;; Commands
;; ---------------------------------------------------------------------------

(defn- cmd-run [args]
  (let [{:keys [opts positional]}
        (parse-args args {:provider {:short "p" :kind :option}
                          :set      {:short "s" :kind :accumulate}
                          :model    {:short "m" :kind :option}
                          :help     {:short "h" :kind :flag}})
        wf-path (first positional)
        input   (str/join " " (rest positional))]
    (when (or (:help opts) (nil? wf-path))
      (println "Usage: glitch run [options] <file> [input...]")
      (println)
      (println "Options:")
      (println "  -p, --provider <name>   Default provider for LLM calls")
      (println "  -s, --set <key=value>   Set parameter (repeatable)")
      (println "  -m, --model <model>     Default model name")
      (System/exit (if (:help opts) 0 1)))

    (when-not (.exists (io/file wf-path))
      (println (str "error: file not found: " wf-path))
      (System/exit 1))

    (let [_      (prov/load-providers)
          params (reduce (fn [m pair]
                           (let [idx (str/index-of pair "=")]
                             (if idx
                               (assoc m
                                 (subs pair 0 idx)
                                 (subs pair (inc idx)))
                               (assoc m pair ""))))
                         {} (:set opts))
          db     (try
                   (store/open)
                   (catch Exception e
                     (binding [*out* *err*]
                       (println (str "warn: store unavailable: " (.getMessage e))))
                     nil))
          result (runner/run wf-path
                   :input    (if (str/blank? input) nil input)
                   :db       db
                   :provider (:provider opts)
                   :model    (:model opts)
                   :params   params
                   :workflows-dir
                   (when wf-path
                     (.getParent (io/file wf-path))))]
      (when-let [output (:output result)]
        (when-not (str/blank? output)
          (println output)))
      (when db (store/close db)))))

(defn- cmd-check [args]
  (let [path (first args)]
    (when-not path
      (println "Usage: glitch check <file>")
      (System/exit 1))
    (try
      (read-string (slurp path))
      (println "ok")
      (catch Exception e
        (println (str "error: " (.getMessage e)))
        (System/exit 1)))))

(defn- cmd-eval [args]
  (let [path (first args)]
    (when-not path
      (println "Usage: glitch eval <file>")
      (System/exit 1))
    (load-file path)))

(defn- cmd-up []
  (println "checking required tools...")
  (let [tools ["bb" "curl" "gh" "rg"]
        results (map (fn [t]
                       (try
                         (let [proc (-> (ProcessBuilder. ["which" t])
                                        (.redirectErrorStream true)
                                        .start)]
                           (.waitFor proc)
                           (if (zero? (.exitValue proc))
                             {:tool t :ok true}
                             {:tool t :ok false}))
                         (catch Exception _
                           {:tool t :ok false})))
                     tools)]
    (doseq [{:keys [tool ok]} results]
      (println (str "  " (if ok "[ok]" "[MISSING]") " " tool)))
    (when (some #(not (:ok %)) results)
      (System/exit 1))
    (println "all tools available")))

(defn- cmd-plugin [args]
  (let [name (first args)]
    (when-not name
      (println "Usage: glitch plugin <name> [args...]")
      (System/exit 1))
    (let [home   (System/getProperty "user.home")
          path   (str home "/.config/glitch/plugins/" name "/main.clj")
          f      (io/file path)]
      (if (.exists f)
        (load-file path)
        (do (println (str "error: plugin not found: " path))
            (System/exit 1))))))

;; ---------------------------------------------------------------------------
;; Entry point
;; ---------------------------------------------------------------------------

(defn -main [& args]
  (let [cmd       (first args)
        rest-args (rest args)]
    (case cmd
      "run"     (cmd-run rest-args)
      "check"   (cmd-check rest-args)
      "eval"    (cmd-eval rest-args)
      "up"      (cmd-up)
      "version" (println "glitch 0.3.0-bb")
      "plugin"  (cmd-plugin rest-args)
      "mcp"     (do
                  (require '[glitch.mcp :as mcp])
                  ((resolve 'glitch.mcp/start) {}))
      "repl"    (let [{:keys [opts]} (parse-args rest-args
                                        {:port {:short "p" :kind :option}})
                      port (or (some-> (:port opts) parse-long) 1667)]
                  (require '[glitch.repl :as repl])
                  ((resolve 'glitch.repl/start) {:port port}))
      (do (println "glitch - workflow engine (babashka)")
          (println)
          (doseq [c (sort ["check" "eval" "mcp" "plugin" "repl" "run" "up" "version"])]
            (println (str "  " c)))
          (when-not cmd (System/exit 1))))))
```

- [ ] **Step 2: Verify compilation**

Run: `cd bb && bb -cp src:providers -e '(require (quote glitch.main))'`
Expected: No errors (repl.clj doesn't exist yet, but it's lazily required so this is fine).

- [ ] **Step 3: Commit**

```bash
git add bb/src/glitch/main.clj
git commit -m "feat: simplify CLI — remove workspace, add -p provider flag

glitch run now takes explicit file paths only.
Adds -p/--provider flag, removes -P/--project.
Adds rg to glitch up checks. Wires up repl and mcp commands."
```

---

### Task 5: Simplify runner — remove workspace threading

**Files:**
- Modify: `bb/src/glitch/runner.clj`
- Modify: `bb/test/glitch/runner_test.clj`

- [ ] **Step 1: Update runner_test.clj — remove list-workflows test**

Remove the `list-workflows-test` deftest from `bb/test/glitch/runner_test.clj` (lines 274-285). It tests `runner/list-workflows` which is being deleted.

- [ ] **Step 2: Rewrite runner.clj**

Key changes to `bb/src/glitch/runner.clj`:
- Remove requires: `glitch.mcp.tools`, `glitch.mcp.handlers`, `glitch.mcp.indexer`, `glitch.mcp.search`
- Replace `*search-fn*` implementation — backed by ripgrep instead of SQLite FTS5
- Keep `search` in SCI user-ns bindings (same API: `(search "query")` or `(search "query" :path "/repo" :limit 10)`)
- Remove `list-workflows` function
- In `run` function:
  - Remove `:project` kwarg, add `:provider` kwarg
  - Remove the entire workspace-path / search-db / tool-context / tool-handler block (lines 297-319)
  - Remove `available-defs` and tool injection into provider dispatch
  - Bind `*search-fn*` to a ripgrep-backed function using cwd as default path
  - Simplify provider-fn to just use the `:provider` kwarg as default
  - Remove `:workspace` from `store/record-run` call

Replace the `run` function with:

```clojure
(defn run
  "Evaluate a .glitch workflow file.

   Required:
     workflow-path — path to the .glitch or .clj file

   Options (keyword args):
     :input          — string input available via (input)
     :db             — store map (from store/open); nil to skip recording
     :provider       — default provider name for LLM calls
     :model          — default model name for LLM calls
     :params         — map of parameters available via (params)/(param k)
     :seed-steps     — map of step-id->value to pre-populate before eval
     :workflows-dir  — directory for call-workflow resolution
     :parent-run-id  — parent run ID for sub-workflow recording
     :tiers          — provider tier list (overrides default-tiers)"
  [workflow-path & {:keys [input db provider model params seed-steps
                           workflows-dir parent-run-id tiers]}]
  (let [input    (or input "")
        model    (or model "default")
        provider (or provider "lmstudio")
        params   (or params {})
        wf-dir   (or workflows-dir
                     (let [f (io/file workflow-path)]
                       (if-let [parent (.getParent f)]
                         parent
                         ".")))]
    ;; Reset core state
    (g/reset!)
    (g/set-input! input)
    (g/set-params! params)
    (g/set-workflows-dir! wf-dir)

    ;; Bind ripgrep-backed search function
    (reset! *search-fn*
      (fn [query limit path]
        (let [search-path (or path (System/getProperty "user.dir"))
              cmd ["rg" "-n" "-S" "-C" "2" "-m" (str limit) "--" query search-path]
              result (bp/shell {:out :string :err :string :continue true} cmd)]
          (if (<= (:exit result) 1)
            (or (:out result) "")
            (throw (ex-info (str "search failed: " (:err result)) {}))))))

    ;; Wire provider dispatch
    (g/set-provider-fn!
      (fn [opts]
        (let [pname (or (:provider opts) provider)
              merged (assoc opts
                       :model (let [m (or (:model opts) model)]
                                (when-not (= m "default") m)))]
          (if tiers
            (prov/call-tiered merged tiers)
            (if (:provider opts)
              (prov/call-provider pname merged)
              (try
                (prov/call-provider pname merged)
                (catch Exception _
                  (prov/call-tiered merged prov/default-tiers))))))))

    ;; Pre-seed steps
    (when seed-steps
      (doseq [[k v] seed-steps]
        (g/step k v)))

    ;; Record run in store (when db is provided)
    (let [run-id (when db
                   (let [rid (store/record-run db
                               {:name          ""
                                :input         input
                                :workflow-file workflow-path
                                :model         model
                                :parent-run-id parent-run-id})]
                     (g/set-step-recorder!
                       (fn [rec]
                         (store/record-step db (assoc rec :run-id rid))))
                     rid))
          ctx (make-sci-ctx)]
      (binding [*sci-ctx* ctx]
        (try
          (sci/eval-string* ctx (slurp workflow-path))
          (let [result {:name    ""
                        :output  (g/last-output)
                        :steps   (g/get-steps)
                        :run-id  (or run-id 0)}]
            (when db
              (store/finish-run db run-id (:output result) 0))
            result)
          (catch Exception e
            (when db
              (store/finish-run db run-id (str "ERROR: " (.getMessage e)) 1))
            (throw e)))))))
```

Also update the ns requires at the top of runner.clj — remove:
```clojure
[glitch.mcp.tools :as mcp-tools]
[glitch.mcp.handlers :as mcp-handlers]
[glitch.mcp.indexer :as indexer]
[glitch.mcp.search :as search]
```

Add `babashka.process` if not already required:
```clojure
[babashka.process :as bp]
```

Keep `*search-fn*` atom. Update the `search` binding in `make-sci-ctx` user-ns — same API but now accepts optional `:path` kwarg:
```clojure
'search (fn [query & {:keys [limit path] :or {limit 10}}]
          (if-let [f @*search-fn*]
            (f query limit path)
            (throw (ex-info "search not available" {}))))
```

Keep `*search-fn*` atom. Update the `search` binding in `make-sci-ctx` to accept optional `:path` kwarg (see step 2 above).

- [ ] **Step 3: Run runner tests**

Run: `cd bb && bb -cp src:test:providers -e '(require (quote glitch.runner-test)) (clojure.test/run-tests (quote glitch.runner-test))'`
Expected: All tests pass.

- [ ] **Step 4: Commit**

```bash
git add bb/src/glitch/runner.clj bb/test/glitch/runner_test.clj
git commit -m "feat: simplify runner — remove workspace threading

Drops workspace-path, search indexing, and MCP tool wiring.
Adds :provider kwarg for CLI-level provider override.
Removes list-workflows function."
```

---

### Task 6: Simplify MCP — remove workspace, add ripgrep tools

**Files:**
- Modify: `bb/src/glitch/mcp.clj`
- Rewrite: `bb/src/glitch/mcp/handlers.clj`
- Rewrite: `bb/src/glitch/mcp/tools.clj`

- [ ] **Step 1: Rewrite mcp/tools.clj**

```clojure
(ns glitch.mcp.tools)

(def tool-definitions
  [{"name" "glitch_search"
    "description" "Search code using ripgrep. Returns structured results with file paths, line numbers, and matched content."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query"      {"type" "string" "description" "Search pattern (regex by default)"}
      "path"       {"type" "string" "description" "Directory to search in"}
      "glob"       {"type" "string" "description" "File filter glob (e.g. *.clj, *.{ts,tsx})"}
      "fixed"      {"type" "boolean" "description" "Literal string mode, no regex"}
      "multiline"  {"type" "boolean" "description" "Match across newlines"}
      "pcre2"      {"type" "boolean" "description" "Enable PCRE2 (lookaround, backreferences)"}
      "context"    {"type" "integer" "description" "Lines of context around matches"}
      "limit"      {"type" "integer" "description" "Max matches per file"}
      "smart_case" {"type" "boolean" "description" "Case-insensitive unless uppercase present (default true)"}}
     "required" ["query" "path"]}}

   {"name" "glitch_symbols"
    "description" "Search for symbol definitions (functions, types, classes) using language-aware ripgrep patterns."
    "inputSchema"
    {"type" "object"
     "properties"
     {"query"    {"type" "string" "description" "Symbol name to search for"}
      "path"     {"type" "string" "description" "Directory to search in"}
      "language" {"type" "string" "description" "Language hint: clojure, go, python, javascript, typescript, rust"}}
     "required" ["query" "path"]}}

   {"name" "glitch_run"
    "description" "Execute a glitch workflow file."
    "inputSchema"
    {"type" "object"
     "properties"
     {"file"  {"type" "string" "description" "Path to the workflow file"}
      "input" {"type" "string" "description" "Input text to pass to the workflow"}
      "set"   {"type" "object" "description" "Key-value pairs to set as workflow parameters"}}
     "required" ["file"]}}

   {"name" "glitch_eval"
    "description" "Evaluate a Clojure expression via SCI and return the result."
    "inputSchema"
    {"type" "object"
     "properties"
     {"expression" {"type" "string" "description" "Clojure expression to evaluate"}}
     "required" ["expression"]}}

   {"name" "glitch_check"
    "description" "Check a workflow file for syntax errors."
    "inputSchema"
    {"type" "object"
     "properties"
     {"file" {"type" "string" "description" "Path to the workflow file to check"}}
     "required" ["file"]}}

   {"name" "glitch_read_file"
    "description" "Read a file and return its first 200 lines."
    "inputSchema"
    {"type" "object"
     "properties"
     {"path" {"type" "string" "description" "Path to the file to read"}}
     "required" ["path"]}}])
```

- [ ] **Step 2: Rewrite mcp/handlers.clj**

```clojure
(ns glitch.mcp.handlers
  (:require [babashka.process :as bp]
            [cheshire.core :as json]
            [clojure.string :as str]
            [sci.core :as sci]))

;; ---------------------------------------------------------------------------
;; Ripgrep search
;; ---------------------------------------------------------------------------

(defn- handle-search [arguments]
  (let [query      (get arguments "query")
        path       (get arguments "path")
        glob       (get arguments "glob")
        fixed      (get arguments "fixed" false)
        multiline  (get arguments "multiline" false)
        pcre2      (get arguments "pcre2" false)
        context    (get arguments "context")
        limit      (get arguments "limit")
        smart-case (get arguments "smart_case" true)
        cmd (cond-> ["rg" "--json"]
              smart-case         (conj "-S")
              (not smart-case)   (conj)
              fixed              (conj "-F")
              multiline          (conj "-U")
              pcre2              (conj "-P")
              context            (conj "-C" (str context))
              limit              (conj "-m" (str limit))
              glob               (conj "-g" glob)
              true               (conj "--" query path))
        result (bp/shell {:out :string :err :string :continue true} cmd)]
    (if (zero? (:exit result))
      (:out result)
      (if (= 1 (:exit result))
        "[]"  ;; no matches
        (throw (ex-info (str "rg failed: " (:err result)) {}))))))

;; ---------------------------------------------------------------------------
;; Symbol search
;; ---------------------------------------------------------------------------

(def ^:private symbol-patterns
  {"clojure"    "^\\(def[n\\-]?\\s+%s"
   "go"         "^(func|type|var|const)\\s+.*%s"
   "python"     "^(def|class)\\s+%s"
   "javascript" "(function|const|let|var|class|export)\\s+%s"
   "typescript" "(function|const|let|var|class|export|interface|type)\\s+%s"
   "rust"       "^(fn|struct|enum|trait|type|const|static)\\s+%s"})

(defn- detect-language
  "Guess language from directory contents."
  [path]
  (let [result (bp/shell {:out :string :err :string :continue true}
                 "rg" "--files" "--max-depth" "2" path)
        files (str/split-lines (:out result))]
    (cond
      (some #(str/ends-with? % ".clj") files)  "clojure"
      (some #(str/ends-with? % ".go") files)   "go"
      (some #(str/ends-with? % ".py") files)   "python"
      (some #(str/ends-with? % ".ts") files)   "typescript"
      (some #(str/ends-with? % ".js") files)   "javascript"
      (some #(str/ends-with? % ".rs") files)   "rust"
      :else nil)))

(defn- handle-symbols [arguments]
  (let [query    (get arguments "query")
        path     (get arguments "path")
        language (or (get arguments "language")
                     (detect-language path))
        pattern  (if language
                   (format (get symbol-patterns language
                             "(def|fn|func|class|type|struct|const|let|var)\\s+%s")
                           query)
                   query)
        cmd      ["rg" "-n" "-S" "--" pattern path]
        result   (bp/shell {:out :string :err :string :continue true} cmd)]
    (if (<= (:exit result) 1)
      (or (:out result) "")
      (throw (ex-info (str "rg failed: " (:err result)) {})))))

;; ---------------------------------------------------------------------------
;; Workflow execution
;; ---------------------------------------------------------------------------

(defn- handle-run [arguments]
  (let [file       (get arguments "file")
        input      (get arguments "input")
        set-params (get arguments "set")
        cmd (cond-> ["glitch" "run"]
              set-params (into (mapcat (fn [[k v]] ["-s" (str k "=" v)]) set-params))
              true       (conj file)
              input      (conj input))
        result (apply bp/shell {:out :string :err :string :continue true} cmd)]
    (if (zero? (:exit result))
      (:out result)
      (throw (ex-info (str "workflow failed (exit " (:exit result) "): " (:err result)) {})))))

;; ---------------------------------------------------------------------------
;; Eval / Check / Read
;; ---------------------------------------------------------------------------

(defn- handle-eval [arguments]
  (let [expression (get arguments "expression")
        ctx (sci/init {:namespaces {'user {}}})
        result (sci/eval-string* ctx expression)]
    (str result)))

(defn- handle-check [arguments]
  (let [file (get arguments "file")
        content (slurp file)]
    (try
      (read-string (str "[" content "]"))
      "ok"
      (catch Exception e
        (str "error: " (.getMessage e))))))

(defn- handle-read-file [arguments]
  (let [path (get arguments "path")
        f    (java.io.File. path)]
    (when-not (.exists f)
      (throw (ex-info (str "file not found: " path) {})))
    (let [lines (str/split-lines (slurp f))
          selected (take 200 lines)]
      (str/join "\n" selected))))

;; ---------------------------------------------------------------------------
;; Handler dispatch
;; ---------------------------------------------------------------------------

(defn make-handler
  "Create a tool handler function. Context is unused but kept for API compat."
  [_context]
  (fn [tool-name arguments]
    (case tool-name
      "glitch_search"    (handle-search arguments)
      "glitch_symbols"   (handle-symbols arguments)
      "glitch_run"       (handle-run arguments)
      "glitch_eval"      (handle-eval arguments)
      "glitch_check"     (handle-check arguments)
      "glitch_read_file" (handle-read-file arguments)
      (throw (ex-info (str "unknown tool: " tool-name) {})))))
```

- [ ] **Step 3: Simplify mcp.clj**

```clojure
(ns glitch.mcp
  "MCP stdio server entry point.
   Reads JSON-RPC messages from stdin, dispatches via protocol, writes to stdout."
  (:require [glitch.mcp.protocol :as proto]
            [glitch.mcp.tools :as tools]
            [glitch.mcp.handlers :as handlers]
            [clojure.string :as str]))

(defn start [_opts]
  (let [handler (handlers/make-handler {})
        dispatch-ctx {:tools tools/tool-definitions
                      :tool-handler handler}]
    (binding [*out* *err*]
      (println "[glitch-mcp] server started"))
    (try
      (loop []
        (when-let [line (read-line)]
          (let [trimmed (str/trim line)]
            (when (seq trimmed)
              (let [msg (proto/parse-message trimmed)]
                (if (:error msg)
                  (do
                    (println (proto/format-error nil -32700 "Parse error"))
                    (flush))
                  (when-let [resp (proto/dispatch msg dispatch-ctx)]
                    (println resp)
                    (flush))))))
          (recur)))
      (finally
        (binding [*out* *err*]
          (println "[glitch-mcp] server stopped"))))))

(defn -main [& _args]
  (start {}))
```

- [ ] **Step 4: Commit**

```bash
git add bb/src/glitch/mcp.clj bb/src/glitch/mcp/handlers.clj bb/src/glitch/mcp/tools.clj
git commit -m "feat: MCP tools powered by ripgrep, remove workspace deps

Replaces glitch_search with rg --json wrapper.
Replaces glitch_symbols with language-aware rg patterns.
Removes glitch_index, glitch_grep (subsumed by glitch_search).
Drops confined-path, workspace-path, search-db, embeddings."
```

---

### Task 7: Add glitch repl

**Files:**
- Create: `bb/src/glitch/repl.clj`
- Create: `bb/test/glitch/repl_test.clj`
- Modify: `bb/src/glitch/test_runner.clj`

- [ ] **Step 1: Write repl test**

```clojure
(ns glitch.repl-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.repl :as repl]))

(deftest port-file-path-test
  (testing "port-file-path returns .nrepl-port in given dir"
    (is (= "/tmp/.nrepl-port" (repl/port-file-path "/tmp")))))
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd bb && bb -cp src:test:providers -e '(require (quote glitch.repl-test)) (clojure.test/run-tests (quote glitch.repl-test))'`
Expected: Failure — `glitch.repl` doesn't exist.

- [ ] **Step 3: Write glitch.repl**

```clojure
(ns glitch.repl
  "nREPL server with glitch DSL preloaded.
   Starts babashka.nrepl and writes .nrepl-port for CIDER."
  (:require [babashka.nrepl.server :as nrepl]
            [clojure.java.io :as io]
            [glitch.core :as g]
            [glitch.provider :as prov]))

(defn port-file-path
  "Return the .nrepl-port file path for a directory."
  [dir]
  (str dir "/.nrepl-port"))

(defn start
  "Start an nREPL server with glitch DSL available in the user namespace.
   Options:
     :port — port number (default 1667)
     :dir  — directory for .nrepl-port file (default cwd)"
  [{:keys [port dir] :or {port 1667
                           dir (System/getProperty "user.dir")}}]
  ;; Load providers so (llm ...) works out of the box
  (prov/load-providers)

  ;; Wire up default provider
  (g/set-provider-fn!
    (fn [opts]
      (let [pname (or (:provider opts) "lmstudio")]
        (prov/call-provider pname opts))))

  ;; Write .nrepl-port file (CIDER convention)
  (let [port-file (io/file (port-file-path dir))]
    (spit port-file (str port))
    (.deleteOnExit port-file))

  (binding [*out* *err*]
    (println (str "glitch repl on port " port))
    (println (str "connect: cider-connect localhost " port)))

  ;; Start the nREPL server — blocks until killed
  (nrepl/start-server! {:host "localhost" :port port})
  @(promise))
```

- [ ] **Step 4: Run test**

Run: `cd bb && bb -cp src:test:providers -e '(require (quote glitch.repl-test)) (clojure.test/run-tests (quote glitch.repl-test))'`
Expected: PASS.

- [ ] **Step 5: Add repl-test to test runner**

Update `bb/src/glitch/test_runner.clj` — add `[glitch.repl-test]` to requires and `'glitch.repl-test` to the run list.

- [ ] **Step 6: Commit**

```bash
git add bb/src/glitch/repl.clj bb/test/glitch/repl_test.clj bb/src/glitch/test_runner.clj
git commit -m "feat: add glitch repl — nREPL server with DSL preloaded

Starts babashka.nrepl on port 1667 (configurable with -p).
Writes .nrepl-port for CIDER auto-detection.
Pre-loads providers so (llm ...) works immediately."
```

---

### Task 8: Run full test suite and verify build

**Files:** None (verification only)

- [ ] **Step 1: Run full test suite**

Run: `cd bb && bb test`
Expected: All tests pass, no compilation errors.

- [ ] **Step 2: Build uberscript**

Run: `cd bb && bb build`
Expected: `build/glitch` created successfully.

- [ ] **Step 3: Verify shebang works**

```bash
cd /tmp
cat > hello.glitch << 'EOF'
#!/usr/bin/env glitch run
(step "greeting" "hello from shebang")
EOF
chmod +x hello.glitch
./hello.glitch
```
Expected: `hello from shebang`

- [ ] **Step 4: Verify CLI help**

Run: `cd bb && bb -cp src:providers -m glitch.main run --help`
Expected: Shows usage with `-p/--provider`, `-m/--model`, `-s/--set` flags. No mention of `-P/--project`.

- [ ] **Step 5: Commit if any fixups needed, then final commit**

```bash
git add -A
git commit -m "chore: verify full test suite and build after workspace removal"
```

---

Plan complete and saved to `docs/superpowers/plans/2026-04-21-workspace-removal.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?