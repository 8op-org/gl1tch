# Investigation Graphs (Bayesian Reasoning)

Structured investigation with fact tracking and confidence scoring.

## Usage

```clojure
(investigate "The deploy failure was caused by a config change"

  ;; Create facts with confidence scores
  (step "check-logs"
    (sh "kubectl logs deploy/app --tail 100"))
  (fact "logs-show-config-error"
    :claim "Logs contain ConfigMap parse failure"
    :confidence 0.6
    :source-step "check-logs")

  ;; Corroborate from another source
  (step "check-git"
    (sh "git log --oneline -5 -- k8s/"))
  (corroborate-from "config-changed-recently"
    :claim "ConfigMap was modified in last commit"
    :step "check-git"
    :corroborates "logs-show-config-error")

  ;; Query graph state
  (approve! "logs-show-config-error")   ;; mark as trusted (confidence 1.0)
  (reachable? "goal")                    ;; can we reach the goal from approved facts?
  (confidence-gap "goal")                ;; find weakest link
  (suggest-next)                         ;; what to investigate next
  (graph-stats))                         ;; introspection
```

## Confidence Rules

- **Single-source cap:** 0.70 max from one source
- **Bayesian combination** breaks the cap when multiple sources corroborate
- **Contradiction detection:** keyword overlap + negation polarity flip
- **Geometric decay** (0.7^n) when contradictions are detected

## Graph Primitives

| Form | Description |
|------|-------------|
| `(investigate "hypothesis" body...)` | Start an investigation |
| `(fact "id" :claim "..." :confidence N :source-step "id")` | Create a fact node |
| `(corroborate-from "id" :claim "..." :step "id" :corroborates "id")` | Add corroborating evidence |
| `(approve! "id")` | Mark fact as trusted (confidence 1.0) |
| `(reachable? "goal")` | Can the goal be reached from approved facts? |
| `(confidence-gap "goal")` | Find the weakest link in the chain |
| `(suggest-next)` | Suggest what to investigate next |
| `(graph-stats)` | Introspection — counts, scores, gaps |
