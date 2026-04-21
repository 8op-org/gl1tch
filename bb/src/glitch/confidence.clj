(ns glitch.confidence
  "Pure math functions for confidence scoring — no state, no side effects
   (except `embed` which calls TEI)."
  (:require [babashka.http-client :as http]
            [cheshire.core :as json]
            [clojure.string :as str]))

;; --- Provider authority ---

(def default-authority
  "Map of provider name to trust weight (0.0-1.0)."
  {"copilot"   0.90
   "openrouter" 0.85
   "lmstudio"  0.60
   "ollama"    0.55})

(defn provider-authority
  "Look up the trust weight for a provider. Unknown providers get 0.50."
  [provider-name]
  (get default-authority (str provider-name) 0.50))

;; --- Bayesian combine ---

(defn bayesian-combine
  "Combine independent witness confidences using: 1 - product(1 - authority_i * base_conf).
   `witnesses` is a seq of authority weights (0.0-1.0), `base-conf` is the
   base confidence from each witness.
   Returns 0.0 for empty witnesses."
  [witnesses base-conf]
  (if (empty? witnesses)
    0.0
    (- 1.0 (reduce * 1.0 (map #(- 1.0 (* % base-conf)) witnesses)))))

;; --- Quality score ---

(defn quality-score
  "Score a response's quality: authority * log(1 + char-count)."
  [authority response-text]
  (* authority (Math/log (+ 1.0 (count (str response-text))))))

;; --- Harmonic mean ---

(defn harmonic-mean
  "Weighted harmonic mean of [[weight value] ...] pairs.
   Filters out pairs where value <= 0. Returns 0.0 if no valid pairs."
  [scores]
  (let [valid (filter (fn [[_ v]] (and (number? v) (pos? v))) scores)]
    (if (empty? valid)
      0.0
      (let [total-weight (reduce + 0.0 (map first valid))
            weighted-inv (reduce + 0.0 (map (fn [[w v]] (/ w v)) valid))]
        (if (zero? weighted-inv)
          0.0
          (/ total-weight weighted-inv))))))

;; --- Keyword overlap ---

(def ^:private stopwords
  #{"the" "a" "an" "is" "are" "was" "were" "be" "been" "being"
    "have" "has" "had" "do" "does" "did" "will" "would" "could"
    "should" "may" "might" "shall" "can" "need" "dare" "ought"
    "used" "to" "of" "in" "for" "on" "with" "at" "by" "from"
    "as" "into" "through" "during" "before" "after" "above"
    "below" "between" "out" "off" "over" "under" "again"
    "further" "then" "once" "here" "there" "when" "where"
    "why" "how" "all" "each" "every" "both" "few" "more"
    "most" "other" "some" "such" "no" "nor" "not" "only"
    "own" "same" "so" "than" "too" "very" "just" "because"
    "but" "and" "or" "if" "while" "although" "though" "that"
    "this" "these" "those" "it" "its" "i" "me" "my" "we"
    "our" "you" "your" "he" "him" "his" "she" "her" "they"
    "them" "their" "what" "which" "who" "whom"})

(defn- tokenize [text]
  (when (and text (not (str/blank? text)))
    (->> (str/split (str/lower-case (str text)) #"[^a-z0-9]+")
         (remove str/blank?)
         (remove stopwords))))

(defn keyword-overlap
  "Fraction of prompt keywords that appear in the response.
   Returns 1.0 when prompt is empty (nothing to miss)."
  [prompt-text response-text]
  (let [prompt-kws (set (tokenize prompt-text))]
    (if (empty? prompt-kws)
      1.0
      (let [response-kws (set (tokenize response-text))
            hits (count (clojure.set/intersection prompt-kws response-kws))]
        (/ (double hits) (count prompt-kws))))))

;; --- Domain relevance ---

(defn domain-relevance
  "Adjust confidence by keyword overlap: conf * min(1.0, overlap * 3)."
  [confidence prompt-text response-text]
  (let [overlap (keyword-overlap prompt-text response-text)
        factor  (min 1.0 (* overlap 3.0))]
    (* confidence factor)))

;; --- Cosine similarity ---

(defn cosine-similarity
  "Cosine similarity between two numeric vectors. Returns 0.0 if either is empty
   or magnitudes are zero."
  [a b]
  (if (or (empty? a) (empty? b) (not= (count a) (count b)))
    0.0
    (let [dot   (reduce + 0.0 (map * a b))
          mag-a (Math/sqrt (reduce + 0.0 (map #(* % %) a)))
          mag-b (Math/sqrt (reduce + 0.0 (map #(* % %) b)))]
      (if (or (zero? mag-a) (zero? mag-b))
        0.0
        (/ dot (* mag-a mag-b))))))

;; --- TEI embedding ---

(defn embed
  "Call a TEI (Text Embeddings Inference) endpoint to get a vector embedding.
   Uses GLITCH_TEI_URL env var (default http://localhost:8090).
   Returns a vector of doubles, or nil on failure."
  [text]
  (try
    (let [base-url (or (System/getenv "GLITCH_TEI_URL") "http://localhost:8090")
          url      (str base-url "/embed")
          resp     (http/post url
                     {:headers {"content-type" "application/json"}
                      :body    (json/generate-string {:inputs text})
                      :throw   false})]
      (when (= 200 (:status resp))
        (let [parsed (json/parse-string (:body resp))]
          ;; TEI returns [[...]] for single input
          (if (and (vector? parsed) (vector? (first parsed)))
            (vec (first parsed))
            (when (vector? parsed)
              (vec parsed))))))
    (catch Exception _
      nil)))
