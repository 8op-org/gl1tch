# Confidence Graph — Live Investigation Mesh with Pathfinding

**Date:** 2026-04-21
**Status:** Design
**Scope:** New `graph.clj` namespace + new `investigate` SCI form + store tables
**Companion to:** `2026-04-21-confidence-math-design.md` (scoring foundation)

## Problem

The confidence-math spec gives us per-step scores — Bayesian accumulation, composite harmonic means, domain relevance. But scores in isolation don't answer: **is there a gapless chain of confidence from what I know to what I need to prove?**

Current glitch workflows are linear step sequences. Each step produces output, optionally scored. But there's no structure connecting facts, no way to detect when step 7 contradicts step 3, and no way to measure whether the investigation has converged or still has holes.

Max Chaban's approach: treat every discovered fact as a node in a live graph, every derivation as a weighted edge, and use pathfinding to measure whether you can reach the goal from your axioms without gaps. Contradictions are first-class events that lower confidence on both sides and trigger re-investigation.

## Core Concepts

### Fact

A discrete claim discovered during investigation. Has a confidence score, a source (which step/provider produced it), and a timestamp.

```clj
{:id          "f-001"
 :claim       "PR #3405 added null guards for hot/warm/cold/frozen tiers"
 :confidence  0.65
 :source      {:step "git-log" :provider "shell" :tokens 0}
 :created-at  1713700000
 :status      :unapproved}   ;; :unapproved | :approved | :contradicted
```

### Edge

A derivation link between two facts. "Fact B was derived from Fact A" or "Fact B corroborates Fact A". Weight = confidence of the derivation.

```clj
{:from   "f-001"
 :to     "f-005"
 :rel    :derives      ;; :derives | :corroborates | :contradicts
 :weight 0.60          ;; confidence of this derivation
 :source "analysis-step"}
```

### Contradiction

When a new fact contradicts an existing one. Both facts get their confidence reduced. The contradiction itself is stored as a `:contradicts` edge.

### Goal

The target state — what the investigation needs to prove. Could be "this PR is safe to merge", "the migration will succeed", "the bug is in module X".

```clj
{:id    "goal"
 :claim "PR is safe to approve — all changes verified"
 :confidence 0.0   ;; starts at 0, climbs as paths connect
 :status :unapproved}
```

## Architecture

New file: `bb/src/glitch/graph.clj`

```
graph.clj
  ├── make-graph          — create empty graph atom
  ├── add-fact!           — insert fact node, check contradictions
  ├── add-edge!           — link two facts
  ├── contradict!         — mark contradiction, reduce confidence
  ├── approve!            — mark fact as approved (axiom)
  ├── shortest-path       — Dijkstra from approved facts to goal
  ├── confidence-gap      — find weakest edge in best path
  ├── reachable?          — can goal be reached at all?
  ├── suggest-next        — what investigation would close the biggest gap?
  ├── graph-stats         — node count, edge count, approved count, min-path confidence
  └── efficiency-metrics  — confidence-per-token, confidence-per-time
```

Integration points:

```
runner.clj    — new `*graph*` dynamic var, wired during run
core.clj      — `llm` and `step` optionally emit facts to graph
store.clj     — new `facts` and `fact_edges` tables for persistence
confidence.clj — scoring functions feed fact confidence values
```

## 1. Live Fact Graph

The graph is an atom holding `{:facts {} :edges [] :goal nil}`, created at investigation start and updated with every step.

```clj
(defn make-graph
  "Create an empty investigation graph."
  []
  (atom {:facts {}    ;; id -> fact map
         :edges []    ;; vector of edge maps
         :goal  nil   ;; goal fact (set by investigate form)
         :stats {:tokens-spent 0
                 :time-spent-ms 0
                 :facts-added 0}}))
```

### Adding facts

Every `llm` call during an investigation can emit facts. The workflow author decides what constitutes a fact — the engine provides the mechanism:

```clj
(defn add-fact!
  "Add a fact to the graph. Checks for contradictions against existing facts.
   Returns the fact (possibly with reduced confidence if contradiction found)."
  [graph fact]
  (let [existing   (vals (:facts @graph))
        conflicts  (detect-contradictions fact existing)
        conf-after (if (seq conflicts)
                     ;; Reduce both sides — geometric decay
                     (* (:confidence fact) (Math/pow 0.7 (count conflicts)))
                     (:confidence fact))
        fact'      (assoc fact :confidence conf-after)]
    ;; Mark contradictions
    (doseq [c conflicts]
      (contradict! graph (:id c) (:id fact')))
    ;; Insert
    (swap! graph assoc-in [:facts (:id fact')] fact')
    (swap! graph update-in [:stats :facts-added] inc)
    fact'))
```

### Contradiction detection

Two approaches, tried in order:

1. **Negation overlap** — if fact A says "X is true" and fact B says "X is false" or "not X", flag it. Simple keyword heuristic on the claim text.

2. **Semantic contradiction** — if TEI is available (from companion spec), embed both claims and check if they're similar (cosine > 0.8) but contain negation markers. This catches "safe to deploy" vs "deployment is risky".

```clj
(defn detect-contradictions
  "Find existing facts that contradict a new fact.
   Returns seq of conflicting fact maps."
  [new-fact existing-facts]
  (let [claim (str/lower-case (:claim new-fact))
        neg-markers #{"not" "never" "no" "false" "unsafe" "fail" "wrong"
                      "incorrect" "missing" "absent" "broken"}]
    (->> existing-facts
         (filter (fn [ef]
                   (let [ec (str/lower-case (:claim ef))
                         ;; Share significant keywords but differ on polarity
                         new-kw (set (re-seq #"\w{4,}" claim))
                         old-kw (set (re-seq #"\w{4,}" ec))
                         overlap (count (clojure.set/intersection new-kw old-kw))
                         new-neg (some neg-markers (re-seq #"\w+" claim))
                         old-neg (some neg-markers (re-seq #"\w+" ec))]
                     ;; High overlap + polarity flip = contradiction
                     (and (>= overlap 3)
                          (not= (boolean new-neg) (boolean old-neg))))))
         vec)))

(defn contradict!
  "Mark two facts as contradicting. Reduce confidence on both.
   Add a :contradicts edge."
  [graph id-a id-b]
  (let [decay 0.7]
    (swap! graph update-in [:facts id-a :confidence] * decay)
    (swap! graph update-in [:facts id-a :status] (constantly :contradicted))
    (swap! graph update-in [:facts id-b :confidence] * decay)
    (swap! graph update-in [:edges] conj
           {:from id-a :to id-b :rel :contradicts :weight 0.0
            :source "contradiction-detector"})))
```

## 2. Confidence Ceiling — 70% Single-Source Cap

No single provider call can push a fact above 0.7. This is enforced at fact creation:

```clj
(def single-source-cap 0.70)

;; In add-fact!, before insertion:
(let [capped (update fact :confidence min single-source-cap)]
  ...)
```

To break through 0.7, the fact needs **corroboration** — a second source (different provider, shell command, file read) that supports the same claim. Corroboration uses Bayesian accumulation from the companion spec:

```clj
(defn corroborate!
  "A second source supports an existing fact. Apply Bayesian accumulation.
   This is the ONLY way to push confidence above the single-source cap."
  [graph fact-id new-source]
  (let [fact     (get-in @graph [:facts fact-id])
        old-conf (:confidence fact)
        src-auth (or (:authority new-source)
                     (confidence/provider-authority (:provider new-source)))
        ;; Bayesian: 1 - (1 - old) × (1 - new_authority × base)
        new-conf (- 1.0 (* (- 1.0 old-conf)
                           (- 1.0 (* src-auth 0.85))))]
    (swap! graph assoc-in [:facts fact-id :confidence] new-conf)
    (swap! graph update-in [:edges] conj
           {:from fact-id :to fact-id :rel :corroborates
            :weight new-conf :source (:step new-source)})))
```

Example: fact starts at 0.65 from lmstudio. Copilot (authority=0.90) corroborates:

```
1 - (1 - 0.65) × (1 - 0.90 × 0.85)
= 1 - 0.35 × 0.235
= 1 - 0.082
= 0.918
```

Now it's above 0.7 — triangulated.

## 3. Pathfinding — Dijkstra and Gap Analysis

### Shortest path from axioms to goal

Nodes: facts. Edge weight: `1 - confidence` (lower confidence = higher cost). Approved facts have weight 0 (free to start from).

```clj
(defn shortest-path
  "Dijkstra from any approved fact to the goal.
   Edge cost = 1 - edge_weight (confidence).
   Returns {:path [fact-ids] :cost float :min-edge float} or nil if unreachable."
  [graph goal-id]
  (let [g       @graph
        facts   (:facts g)
        edges   (:edges g)
        starts  (->> facts vals (filter #(= :approved (:status %))) (map :id) set)]
    (when (and (seq starts) (contains? facts goal-id))
      (dijkstra facts edges starts goal-id))))

(defn- dijkstra
  "Standard Dijkstra with multiple start nodes."
  [facts edges starts goal-id]
  (let [adj (build-adjacency edges)
        ;; Priority queue: [cost node-id path]
        init (map (fn [s] [0.0 s [s]]) starts)]
    (loop [pq   (into (sorted-set-by #(compare (first %1) (first %2))) init)
           seen #{}]
      (when-let [[cost node path] (first pq)]
        (let [pq' (disj pq (first pq))]
          (cond
            (= node goal-id)
            {:path path
             :cost cost
             :min-edge (min-edge-confidence edges path)}

            (seen node)
            (recur pq' seen)

            :else
            (let [neighbors (get adj node [])
                  new-entries (for [{:keys [to weight]} neighbors
                                   :when (not (seen to))
                                   :let [edge-cost (- 1.0 (max weight 0.01))]]
                               [(+ cost edge-cost) to (conj path to)])]
              (recur (into pq' new-entries)
                     (conj seen node)))))))))
```

### Gap analysis

The weakest edge in the best path is the investigation's bottleneck:

```clj
(defn confidence-gap
  "Find the weakest link in the path from axioms to goal.
   Returns {:fact-id id :confidence float :claim text} or nil if no path."
  [graph goal-id]
  (when-let [result (shortest-path graph goal-id)]
    (let [facts (:facts @graph)
          path-facts (map #(get facts %) (:path result))
          weakest (apply min-key :confidence path-facts)]
      {:fact-id    (:id weakest)
       :confidence (:confidence weakest)
       :claim      (:claim weakest)
       :path-cost  (:cost result)})))
```

### Suggest next investigation

What action would improve total confidence the most?

```clj
(defn suggest-next
  "Suggest which fact to investigate next for maximum confidence gain.
   Priority: contradicted facts first, then lowest-confidence facts on the critical path."
  [graph goal-id]
  (let [g     @graph
        facts (vals (:facts g))
        contradicted (filter #(= :contradicted (:status %)) facts)
        gap   (confidence-gap graph goal-id)]
    (cond
      ;; Contradictions take priority — must resolve before path is trustworthy
      (seq contradicted)
      {:action :resolve-contradiction
       :facts  (map #(select-keys % [:id :claim :confidence]) contradicted)}

      ;; Weakest link on critical path
      gap
      {:action :strengthen
       :fact   (select-keys gap [:fact-id :claim :confidence])
       :suggestion (str "Corroborate: \"" (:claim gap) "\" — current confidence "
                        (format "%.2f" (:confidence gap)))}

      ;; No path at all — need more facts
      :else
      {:action :explore
       :suggestion "No path to goal exists. Discover connecting facts."})))
```

## 4. Efficiency Metrics

Track confidence gained per resource spent:

```clj
(defn efficiency-metrics
  "Compute confidence-per-token and confidence-per-time for the investigation."
  [graph]
  (let [{:keys [tokens-spent time-spent-ms facts-added]} (:stats @graph)
        facts    (vals (:facts @graph))
        total-conf (reduce + (map :confidence facts))
        avg-conf   (if (pos? (count facts))
                     (/ total-conf (count facts))
                     0.0)]
    {:total-facts       (count facts)
     :avg-confidence    (double avg-conf)
     :tokens-spent      tokens-spent
     :time-spent-ms     time-spent-ms
     :conf-per-token    (if (pos? tokens-spent)
                          (/ total-conf tokens-spent)
                          0.0)
     :conf-per-time-ms  (if (pos? time-spent-ms)
                          (/ total-conf time-spent-ms)
                          0.0)
     :facts-per-1k-tok  (if (pos? tokens-spent)
                          (* 1000.0 (/ facts-added tokens-spent))
                          0.0)}))
```

Shell commands (`git log`, `cat`, `grep`) cost 0 tokens and are fast. The efficiency metrics naturally reward using read-only shell commands for fact discovery and reserving LLM calls for synthesis.

## 5. The `investigate` Form

New SCI macro that sets up the graph, runs investigation steps, and enforces the confidence ceiling:

```clj
(workflow "pr-review"
  :description "Investigate a PR for safety"

  ;; Set up investigation with a goal
  (investigate "PR #1234 is safe to merge"

    ;; Approved axioms — things we know for certain
    (approve! (fact "repo is elastic/observability-test-environments"
                    :confidence 1.0 :source "user-input"))

    ;; Discovery phase — collect facts
    (step "commits" (sh "git log --oneline HEAD~5..HEAD"))
    (fact-from "commits"
      :claim "PR has 3 commits touching main.tf"
      :confidence 0.65 :source {:step "commits" :provider "shell"})

    (step "diff" (sh "git diff HEAD~3"))
    (step "analysis"
      (llm :prompt (str "Analyze this diff for safety:\n" (ref "diff"))
           :schema {:required ["safe" "risks" "confidence"]}
           :min-confidence 0.5))

    ;; Extract facts from LLM analysis
    (fact-from "analysis"
      :extract-key "safe"     ;; pull claim from JSON field
      :derives-from "commits" ;; link to parent fact
      :confidence-key "confidence")

    ;; Corroborate with second provider
    (step "second-opinion"
      (llm :prompt (str "Is this diff safe?\n" (ref "diff"))
           :provider "lmstudio"
           :schema {:required ["safe" "confidence"]}))

    (corroborate-from "second-opinion"
      :target "analysis"      ;; which fact to corroborate
      :confidence-key "confidence")

    ;; Check: do we have a path?
    (when-not (reachable?)
      (step "explore"
        (llm :prompt "What else should we check?")))

    ;; Final: report path confidence
    (step "verdict"
      (str (graph-stats) "\n"
           "Gap: " (confidence-gap) "\n"
           "Suggestion: " (suggest-next)))))
```

### Helper forms

```clj
;; fact — create and add a fact to the graph
(fact claim & {:keys [confidence source derives-from]})

;; fact-from — extract a fact from a step's output
(fact-from step-id & {:keys [claim extract-key confidence-key derives-from]})

;; corroborate-from — use a step's output to corroborate an existing fact
(corroborate-from step-id & {:keys [target confidence-key]})

;; approve! — mark a fact as an axiom (confidence = 1.0)
(approve! fact)

;; reachable? — can we reach the goal from approved facts?
(reachable?)

;; graph-stats — summary of current graph state
(graph-stats)

;; confidence-gap — weakest link in best path
(confidence-gap)

;; suggest-next — what to investigate next
(suggest-next)
```

## 6. Store Schema

Two new tables for persistence across runs:

```sql
CREATE TABLE IF NOT EXISTS facts (
  id          TEXT PRIMARY KEY,
  run_id      INTEGER NOT NULL REFERENCES runs(id),
  claim       TEXT NOT NULL,
  confidence  REAL NOT NULL,
  status      TEXT NOT NULL DEFAULT 'unapproved',
  source_step TEXT,
  source_prov TEXT,
  tokens_cost INTEGER DEFAULT 0,
  created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS fact_edges (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id   INTEGER NOT NULL REFERENCES runs(id),
  from_id  TEXT NOT NULL,
  to_id    TEXT NOT NULL,
  rel      TEXT NOT NULL,
  weight   REAL NOT NULL,
  source   TEXT
);

CREATE INDEX IF NOT EXISTS idx_facts_run ON facts(run_id);
CREATE INDEX IF NOT EXISTS idx_fact_edges_run ON fact_edges(run_id);
```

## 7. SCI Bindings

New bindings in `runner.clj`:

```clj
;; investigation graph
'fact            graph/sci-fact
'fact-from       graph/sci-fact-from
'approve!        graph/sci-approve!
'corroborate-from graph/sci-corroborate-from
'reachable?      graph/sci-reachable?
'graph-stats     graph/sci-graph-stats
'confidence-gap  graph/sci-confidence-gap
'suggest-next    graph/sci-suggest-next
```

The `investigate` form is a SCI macro (defined in `sci-macros` string) that:
1. Creates the graph atom
2. Sets the goal fact
3. Evaluates the body
4. Returns the graph state + path analysis

## 8. Runner Integration

The runner gains a `*graph*` dynamic var:

```clj
(def ^:dynamic *graph* (atom nil))
```

During an `investigate` block, the graph atom is bound. All `fact`, `fact-from`, `corroborate-from` calls operate on it. Outside `investigate`, these forms are no-ops (return nil).

The `llm` function is extended: when `*graph*` is active, it automatically tracks tokens and time in the graph stats:

```clj
;; In core/llm, after getting result:
(when-let [g @*graph*]
  (swap! g update-in [:stats :tokens-spent] + (or (:tokens-in result) 0)
                                               (or (:tokens-out result) 0))
  (swap! g update-in [:stats :time-spent-ms] + (Math/round elapsed)))
```

## Testing Strategy

### Unit tests (`graph_test.clj`)

- `add-fact!`: basic insertion, confidence capping at 0.70
- `detect-contradictions`: polarity flip detection, no false positives on unrelated facts
- `contradict!`: both facts get confidence reduced, edge added
- `corroborate!`: Bayesian accumulation, breaks through 0.70 cap
- `dijkstra`: known graph with 5 nodes, verify shortest path and cost
- `confidence-gap`: verify it finds the weakest node on the critical path
- `suggest-next`: contradicted facts prioritized, then weakest link, then explore
- `efficiency-metrics`: known token/time values, verify ratios

### Integration tests (`runner_test.clj`)

- Full `investigate` workflow with `fact`, `corroborate-from`, `reachable?`
- Contradiction detection mid-workflow triggers re-investigation
- Graph persisted to store, recoverable across runs
- Efficiency metrics recorded

## Backwards Compatibility

Fully additive:
- `investigate` is a new form — existing workflows don't use it
- `*graph*` is nil by default — `llm`/`step` behavior unchanged
- New store tables created on open (no migration)
- No changes to existing functions' signatures

## Relationship to Confidence Math Spec

```
confidence-math-design.md          confidence-graph-design.md
┌─────────────────────────┐        ┌──────────────────────────────┐
│ bayesian-combine        │◄───────│ corroborate! uses it         │
│ provider-authority      │◄───────│ fact sources carry authority  │
│ quality-score           │        │                              │
│ harmonic-mean           │◄───────│ graph-stats could use it     │
│ cosine-similarity       │◄───────│ detect-contradictions uses it│
│ domain-relevance        │        │                              │
│ embed (TEI)             │◄───────│ semantic contradiction check │
└─────────────────────────┘        └──────────────────────────────┘

scoring is the foundation          investigation consumes it
per-node confidence                graph-level confidence
```
