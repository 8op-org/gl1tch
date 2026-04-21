# Confidence Math — Bayesian Accumulation, Composite Scoring, Semantic Consensus

**Date:** 2026-04-21
**Status:** Design
**Scope:** New `confidence.clj` namespace + upgrades to `consensus` and new `composite-score` function
**Builds on:** `2026-04-20-workflow-confidence-design.md` (implemented)

## Problem

The current confidence framework uses simple threshold comparisons (`conf >= 0.7`) and string-equality consensus (`str/lower-case str/trim`). This works but leaves accuracy on the table:

- All providers are weighted equally — a local 7B model's vote counts the same as Sonnet's
- Consensus uses exact string comparison — "deploy immediately" vs "we should deploy" disagree
- No way to combine schema + confidence + grounding + consensus into a single quality score
- No mechanism to penalize off-topic LLM responses
- Disagreement has no tiebreaker beyond "failed"

Six mathematical techniques from prior invariant-derivation and KB-audit work address these gaps.

## Architecture

New file: `bb/src/glitch/confidence.clj` — pure math functions, no state, no side effects.

```
confidence.clj  (pure math + embedding client)
  ├── bayesian-combine       — independent witness accumulation
  ├── provider-authority      — trust weights per provider
  ├── quality-score           — authority × log(1 + length)
  ├── harmonic-mean           — weighted composite scoring
  ├── keyword-overlap         — domain relevance ratio
  ├── cosine-similarity       — vector similarity (pure math)
  └── embed                   — HTTP client for TEI container (side effect)

core.clj  (orchestration — modified)
  ├── consensus  — uses bayesian-combine, cosine-similarity, quality-score
  └── composite-score  — NEW, uses harmonic-mean over gate results

runner.clj  (SCI bindings — adds composite-score)

store.clj  (no schema changes — composite score goes in artifacts JSON)
```

## 1. Bayesian Witness Accumulation

**Replaces:** `count(matching) / total` in `consensus`

**Math:** Given k independent witnesses each with authority `a_i`:

```
P(true | witnesses) = 1 - ∏(1 - a_i × base_conf)
```

If copilot (a=0.90) and lmstudio (a=0.65) agree:

```
1 - (1 - 0.90) × (1 - 0.65) = 1 - 0.035 = 0.965
```

If they disagree, only the agreeing subset contributes.

**Reference:** Ernst et al. (2001) Daikon; Bayesian independent evidence model.

```clj
(defn bayesian-combine
  "Accumulate confidence from k independent witnesses.
   Each witness contributes (authority × base-conf).
   Returns: 1 - product(1 - contribution_i)"
  [witnesses base-conf]
  (if (empty? witnesses)
    0.0
    (- 1.0 (reduce * (map #(- 1.0 (* (:authority %) base-conf)) witnesses)))))
```

## 2. Provider Authority Weighting

**New:** Each provider gets a trust weight reflecting model capability.

```clj
(def default-authority
  {"copilot"    0.90   ;; Sonnet-class via GH
   "claude"     0.95   ;; Direct Anthropic API
   "openrouter" 0.85   ;; Model-dependent, conservative default
   "lmstudio"   0.65}) ;; Local model, lower capability

(defn provider-authority
  "Look up authority for a provider name.
   Falls back to 0.50 for unknown providers."
  [provider-name]
  (get default-authority provider-name 0.50))
```

Workflow authors can override per-call:

```clj
(consensus ["copilot" "lmstudio"]
  :prompt "..."
  :authority {"copilot" 0.95 "lmstudio" 0.80}  ;; override defaults
  :compare-key "decision")
```

## 3. Quality-Based Tiebreaker

**Replaces:** Consensus disagreement returning `nil` value — now picks the best response.

**Math:**

```
quality = authority × log(1 + response_length)
```

The `log` prevents pure length gaming. A 500-char response from copilot (a=0.90) scores:

```
0.90 × log(501) ≈ 0.90 × 6.22 = 5.59
```

A 200-char response from lmstudio (a=0.65):

```
0.65 × log(201) ≈ 0.65 × 5.30 = 3.45
```

Copilot wins the tiebreak.

**Reference:** González (1985) greedy k-center competitive analysis.

```clj
(defn quality-score
  "Score a response by authority × log(1 + length).
   Favors authoritative, detailed responses without rewarding pure verbosity."
  [authority response-text]
  (* authority (Math/log (inc (count response-text)))))
```

## 4. Weighted Harmonic Mean Composite Score

**New function:** Combines all gate signals into one number. Harmonic mean penalizes weak dimensions — a response that passes schema but fails grounding gets hammered.

**Math:**

```
composite = 1 / Σ(weight_i / max(score_i, ε))
```

**Weights:**

| Gate | Weight | Score source |
|------|--------|-------------|
| confidence | 0.40 | LLM-reported confidence value (0.0-1.0) |
| grounding | 0.30 | 1.0 if grounded, 0.0 if not, partial via `1 - unsupported/total_claims` |
| schema | 0.20 | 1.0 if valid, 0.0 if not |
| contract | 0.10 | 1.0 if passed, 0.0 if not |

```clj
(defn harmonic-mean
  "Weighted harmonic mean of scores. Penalizes weak dimensions.
   scores: seq of [weight value] pairs. Values clamped to [epsilon, 1.0]."
  [scores]
  (let [epsilon 1e-6
        denom (reduce + (map (fn [[w v]] (/ w (max v epsilon))) scores))]
    (/ 1.0 denom)))
```

Exposed to workflows as `composite-score`:

```clj
(step "analysis"
  (llm :prompt "..." :schema {:required ["decision" "confidence"]}
       :min-confidence 0.7))

(grounded? "analysis" (param "context") :strict false)

;; Compute composite — reads recorded gate results from the run
(step "quality"
  (composite-score "analysis"
    :weights {:confidence 0.4 :grounding 0.3 :schema 0.2 :contract 0.1}))
```

`composite-score` reads gate results from the step store. It queries the `steps` table by `(run_id, step_id)` to pull `confidence`, `gate_passed`, and `artifacts` JSON (which contains `schema_valid`, `grounded`, `contract` fields). This requires the store connection to be available — same as `get-steps` already works today. Returns a float 0.0-1.0 as a string (step output).

## 5. Domain Relevance Filtering

**New:** Penalize LLM responses whose vocabulary is off-topic relative to the input context.

**Math:** Keyword overlap ratio — no embeddings needed:

```
stopwords = {the, a, an, is, are, in, of, to, and, or, for, not, ...}
prompt_kw = {w ∈ words(prompt) | len(w) ≥ 4, w ∉ stopwords}
response_kw = {w ∈ words(response) | len(w) ≥ 4, w ∉ stopwords}
relevance = |prompt_kw ∩ response_kw| / max(|prompt_kw|, 1)
adjusted_conf = raw_conf × min(1.0, relevance × 3)
```

The `× 3` scaling means 33% keyword overlap = full credit. Below that, confidence gets penalized.

**Reference:** Robertson & Sparck Jones (1976) TF-IDF; simplified to keyword overlap for workflow use.

```clj
(defn keyword-overlap
  "Fraction of prompt keywords present in response.
   Keywords: words ≥ 4 chars, not in stopword set."
  [prompt-text response-text]
  (let [stopwords #{"the" "that" "this" "with" "from" "have" "been"
                     "will" "would" "could" "should" "their" "there"
                     "which" "when" "what" "were" "they" "also" "than"
                     "into" "your" "does" "more" "only" "just" "some"}
        extract   (fn [text]
                    (->> (re-seq #"\w{4,}" (str/lower-case text))
                         (remove stopwords)
                         set))
        p-kw      (extract prompt-text)
        r-kw      (extract response-text)]
    (if (empty? p-kw)
      1.0
      (/ (double (count (clojure.set/intersection p-kw r-kw)))
         (count p-kw)))))

(defn domain-relevance
  "Adjust a confidence score by domain relevance.
   Returns adjusted confidence: conf × min(1.0, overlap × 3)"
  [confidence prompt-text response-text]
  (* confidence (min 1.0 (* 3.0 (keyword-overlap prompt-text response-text)))))
```

Integrated into `llm` as an optional post-check — only when `:domain-check true` is passed:

```clj
(step "research"
  (llm :prompt (str "Analyze: " (ref "context"))
       :schema {:required ["findings" "confidence"]}
       :min-confidence 0.7
       :domain-check true))  ;; penalize off-topic responses
```

## 6. Semantic Consensus via Embeddings

**Replaces:** String-equality consensus comparison.

**Infrastructure:** HuggingFace Text Embeddings Inference (TEI) container.

```bash
docker run -p 8090:80 \
  ghcr.io/huggingface/text-embeddings-inference:cpu-1.7 \
  --model-id sentence-transformers/all-MiniLM-L6-v2
```

~500MB image, CPU-only, sub-millisecond per embedding. Exposes `/embed` endpoint.

**Clojure integration:**

```clj
(def ^:private tei-url
  (or (System/getenv "GLITCH_TEI_URL") "http://localhost:8090"))

(defn embed
  "Get embedding vector from TEI service.
   Returns a double array or nil on failure."
  [text]
  (try
    (let [resp (http/post (str tei-url "/embed")
                 {:headers {"Content-Type" "application/json"}
                  :body (json/generate-string {:inputs text})
                  :timeout 5000})
          body (json/parse-string (:body resp))]
      (first body))  ;; TEI returns [[float...]]
    (catch Exception _ nil)))

(defn cosine-similarity
  "Cosine similarity between two vectors. Returns 0.0-1.0."
  [a b]
  (let [dot   (reduce + (map * a b))
        mag-a (Math/sqrt (reduce + (map #(* % %) a)))
        mag-b (Math/sqrt (reduce + (map #(* % %) b)))]
    (if (or (zero? mag-a) (zero? mag-b))
      0.0
      (/ dot (* mag-a mag-b)))))
```

**Consensus upgrade:** When TEI is available, consensus uses cosine similarity (threshold 0.85) instead of string equality. Falls back to string comparison when TEI is unreachable.

```clj
;; In consensus, the comparison becomes:
(let [semantic? (some? (embed "test"))  ;; probe TEI availability
      agree-fn  (if semantic?
                  (fn [a b]
                    (let [ea (embed a) eb (embed b)]
                      (and ea eb (>= (cosine-similarity ea eb) 0.85))))
                  (fn [a b]
                    (= (str/lower-case (str/trim a))
                       (str/lower-case (str/trim b)))))]
  ;; use agree-fn to compare votes
  ...)
```

Workflow authors can force a mode:

```clj
(consensus ["copilot" "lmstudio"]
  :prompt "..."
  :compare-key "decision"
  :compare-mode :semantic)  ;; or :exact (default when no TEI)
```

## Updated Consensus Return Shape

The consensus result gains new fields:

```clj
{"agreed"      true
 "value"       {"decision" "deploy" "reason" "..."}
 "confidence"  0.965            ;; Bayesian combined (was: count/total)
 "similarity"  0.94             ;; cosine sim between votes (nil if exact mode)
 "winner"      "copilot"        ;; tiebreaker winner (nil if agreed)
 "votes"       [{"provider" "copilot"
                 "response" {...}
                 "authority" 0.90
                 "quality"   5.59}
                {"provider" "lmstudio"
                 "response" {...}
                 "authority" 0.65
                 "quality"   3.45}]}
```

## New SCI Bindings

Added to `runner.clj` `make-sci-ctx`:

```clj
'composite-score  composite-score   ;; new
```

`bayesian-combine`, `cosine-similarity`, etc. are internal to `confidence.clj` — not exposed as SCI bindings. They're implementation details of `consensus` and `composite-score`.

## Store Changes

No schema changes. The new fields (`similarity`, `winner`, `authority`, `quality`) are stored in the existing `artifacts` JSON column. The `confidence` REAL column now stores Bayesian-combined confidence instead of simple ratio.

## Testing Strategy

### Unit tests (`confidence_test.clj`)

Pure math — no mocking needed:

- `bayesian-combine`: 0 witnesses = 0.0, 1 witness, 2 agreeing, mixed authorities
- `quality-score`: short vs long, high vs low authority
- `harmonic-mean`: all 1.0 = 1.0, one 0.0 ≈ 0.0, weighted correctly
- `keyword-overlap`: full overlap = 1.0, no overlap = 0.0, stopwords excluded
- `cosine-similarity`: identical vectors = 1.0, orthogonal = 0.0, known angle
- `domain-relevance`: on-topic boosts, off-topic penalizes

### Integration tests (`core_test.clj`)

- Consensus with authority: verify Bayesian confidence > simple ratio
- Consensus with quality tiebreaker: disagreeing providers, verify winner selection
- Consensus with semantic comparison: mock TEI responses, verify cosine-based agreement
- Composite score: workflow with all gates, verify harmonic mean output
- Domain relevance: on-topic vs off-topic LLM responses, verify confidence adjustment

### Docker test (`bb test` with TEI running)

- Start TEI container, run consensus with `:compare-mode :semantic`
- Verify "deploy now" and "we should deploy" agree semantically
- Verify "deploy" and "rollback" disagree
- Graceful fallback when TEI is down

## Backwards Compatibility

All changes are additive:

- `consensus` without new options behaves identically (string comparison, equal weights)
- `composite-score` is a new function — no existing workflows affected
- `:domain-check` defaults to false on `llm`
- `:authority` defaults to `default-authority` map
- `:compare-mode` defaults to `:exact` unless TEI is detected and `:compare-key` value is a string

No migrations needed. No breaking changes.

## Workflow Example — All 6 Techniques

```clj
(workflow "high-confidence-deploy"
  :description "Deploy decision with full mathematical confidence"

  (step "diff" (sh "git diff HEAD~1"))

  ;; Consensus with Bayesian accumulation + authority weights + semantic comparison
  (step "decision"
    (consensus ["copilot" "lmstudio"]
      :prompt (str "Analyze this diff. Decide: deploy, hold, or rollback.\n"
                   (ref "diff"))
      :schema {:required ["decision" "confidence" "reason"]
               :types {"decision" :string "confidence" :number "reason" :string}
               :enum {"decision" ["deploy" "hold" "rollback"]}}
      :compare-key "decision"
      :compare-mode :semantic
      :min-confidence 0.7))

  ;; Domain relevance check on the reasoning
  (step "notes"
    (llm :prompt (str "Write deploy notes based on:\n" (ref "diff"))
         :schema {:required ["summary" "risks"]}
         :domain-check true)
    :expects {:non-empty true :min-length 50})

  ;; Ground against the actual diff
  (grounded? "notes" (ref "diff") :provider "lmstudio" :strict false)

  ;; Composite quality score across all gates
  (step "quality"
    (composite-score "notes"
      :weights {:confidence 0.4 :grounding 0.3 :schema 0.2 :contract 0.1}))

  ;; Gate on composite
  (when (< (json/decode (ref "quality")) 0.6)
    (throw (ex-info "quality below threshold" {}))))
```
