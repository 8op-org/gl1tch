(ns glitch.core
  (:require [babashka.process :as bp]
            [clojure.string :as str]
            [clojure.java.io :as io]
            [cheshire.core :as json]))

;; --- Workflow state — dynamic vars holding atoms ---
;; Atoms inside dynamic vars give us both rebindable scope (for call-workflow)
;; and safe concurrent mutation (for par).

(def ^:dynamic *steps* (atom {}))
(def ^:dynamic *step-order* (atom []))
(def ^:dynamic *input* (atom ""))
(def ^:dynamic *params* (atom {}))
(def ^:dynamic *step-recorder* (atom nil))
(def ^:dynamic *provider-fn* (atom nil))
(def ^:dynamic *call-stack* (atom []))
(def ^:dynamic *workflows-dir* (atom "."))
(def ^:dynamic *graph* (atom nil))

;; --- State management ---

(defn reset! []
  (clojure.core/reset! *steps* {})
  (clojure.core/reset! *step-order* []))

(defn set-input! [s]
  (clojure.core/reset! *input* s))

(defn set-params! [p]
  (clojure.core/reset! *params* p))

(defn set-step-recorder! [f]
  (clojure.core/reset! *step-recorder* f))

(defn set-provider-fn! [f]
  (clojure.core/reset! *provider-fn* f))

(defn set-workflows-dir! [d]
  (clojure.core/reset! *workflows-dir* d))

;; --- Primitives ---

(defn input []
  @*input*)

(defn params []
  @*params*)

(defn param [key & [default]]
  (let [p @*params*]
    (or (get p key)
        (get p (keyword key))
        (get p (name key))
        (get p (str key))
        default)))

(defn ref [step-id]
  (get @*steps* step-id))

(declare json-extract)

(defn check-contract
  "Check a step output against a contract map.
   Returns nil on success, or a vector of violation strings on failure."
  [output expects]
  (let [violations (atom [])]
    (when (and (:non-empty expects) (str/blank? output))
      (swap! violations conj "output is empty"))
    (when-let [min-len (:min-length expects)]
      (when (< (count output) min-len)
        (swap! violations conj (str "output length " (count output) " < min " min-len))))
    (when-let [max-len (:max-length expects)]
      (when (> (count output) max-len)
        (swap! violations conj (str "output length " (count output) " > max " max-len))))
    (when (:json expects)
      (let [extracted (json-extract output)]
        (when-not (try (json/parse-string extracted) true (catch Exception _ false))
          (swap! violations conj "output is not valid JSON"))))
    (when-let [required-keys (:keys expects)]
      (let [extracted (json-extract output)
            parsed (try (json/parse-string extracted) (catch Exception _ nil))]
        (if (nil? parsed)
          (swap! violations conj "output is not valid JSON (keys check)")
          (doseq [k required-keys]
            (when-not (contains? parsed k)
              (swap! violations conj (str "missing key: " k)))))))
    (when-let [pattern (:matches expects)]
      (when-not (re-find pattern output)
        (swap! violations conj (str "output does not match pattern: " pattern))))
    (when-let [pred-fn (:pred expects)]
      (when-not (pred-fn output)
        (swap! violations conj "custom predicate failed")))
    (let [v @violations]
      (when (seq v) v))))

(defn step [id body & {:keys [expects]}]
  (let [val (str body)]
    (when expects
      (let [violations (check-contract val expects)]
        (when violations
          (throw (ex-info (str "contract-violation: " (str/join "; " violations))
                          {:kind :contract-violation
                           :step-id id
                           :expected expects
                           :got val
                           :violations violations})))))
    (swap! *steps* assoc id val)
    (swap! *step-order* conj id)
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id id :output val :kind "step"
                 :artifacts (when expects
                              (json/generate-string {:contract "pass"}))}))
    val))

(defn sh [& args]
  (let [result (apply bp/shell {:out :string :err :string :continue true} args)]
    (when (not= 0 (:exit result))
      (throw (ex-info (str "sh: command failed (exit " (:exit result) "): " (:err result))
                      {:cmd (str/join " " args) :exit (:exit result) :err (:err result)})))
    (str/trim (:out result))))

(defn save [path content]
  (io/make-parents path)
  (spit path content)
  content)

(defn read-file [path]
  (slurp path))

(defn write-file [path content]
  (save path content))

(defn get-steps []
  @*steps*)

(defn last-output []
  (let [order @*step-order*]
    (if (empty? order)
      ""
      (or (get @*steps* (last order)) ""))))

;; --- Macros ---

(defmacro workflow [name & body]
  (let [forms (loop [i 0 acc []]
                (if (>= i (count body))
                  acc
                  (if (keyword? (nth body i))
                    (recur (+ i 2) acc)
                    (recur (inc i) (conj acc (nth body i))))))]
    `(let [~'wf-last# (do ~@forms)]
       {:name ~name
        :output (str ~'wf-last#)
        :steps (get-steps)})))

(defmacro par [& forms]
  `(let [futures# (mapv (fn [f#] (future (f#)))
                        [~@(map (fn [form] `(fn [] ~form)) forms)])]
     (mapv deref futures#)))

(defmacro retry [n & body]
  `(loop [attempts# ~n
          last-err# nil]
     (if (<= attempts# 0)
       (throw (ex-info (str "retry exhausted after " ~n " attempts: " last-err#)
                       {:attempts ~n :last-error last-err#}))
       (let [result# (try
                       {:ok (do ~@body)}
                       (catch Exception e#
                         {:err e#}))]
         (if (:ok result#)
           (:ok result#)
           (recur (dec attempts#) (:err result#)))))))

(defmacro with-timeout [seconds & body]
  `(let [f# (future (do ~@body))
         result# (deref f# (* ~seconds 1000) ::timeout)]
     (when (= result# ::timeout)
       (future-cancel f#)
       (throw (ex-info (str "timeout after " ~seconds " seconds")
                       {:timeout ~seconds})))
     result#))

;; --- Gate ---

(defn gate [id predicate]
  (when-let [recorder @*step-recorder*]
    (recorder {:step-id id
               :output (str predicate)
               :kind "gate"
               :gate-passed (if predicate 1 0)}))
  (swap! *steps* assoc id (str predicate))
  predicate)

;; --- Phase ---

(defmacro phase [name & body]
  `(do
     (when-let [recorder# @*step-recorder*]
       (recorder# {:step-id ~name :kind "phase" :output "started"}))
     (let [result# (do ~@body)]
       (when-let [recorder# @*step-recorder*]
         (recorder# {:step-id ~name :kind "phase" :output (str result#)}))
       result#)))

;; --- Call-workflow ---

(defn call-workflow
  "Execute a child workflow. NOTE: This default implementation uses load-file
   (unsandboxed). The runner overrides this in the SCI context with
   sci-call-workflow which uses sci/eval-string* (sandboxed). Do not call
   this directly outside of tests."
  [name & {:keys [input set]}]
  (when (some #{name} @*call-stack*)
    (throw (ex-info (str "call-workflow cycle: " name " already on stack "
                         (str/join " -> " @*call-stack*))
                    {:name name :stack @*call-stack*})))
  (swap! *call-stack* conj name)
  (try
    (let [wf-dir @*workflows-dir*
          path (or (some #(when (.exists (io/file %)) %)
                         [(str wf-dir "/" name ".glitch")
                          (str wf-dir "/" name ".clj")])
                   (throw (ex-info (str "call-workflow: " name " not found in " wf-dir)
                                   {:name name :dir wf-dir})))
          saved-recorder @*step-recorder*]
      (binding [*steps* (atom {})
                *step-order* (atom [])
                *input* (atom (or input ""))
                *params* (atom (merge @*params* (or set {})))
                *step-recorder* (atom saved-recorder)]
        (load-file path)
        {:output (last-output) :steps @*steps*}))
    (finally
      (swap! *call-stack* (comp vec pop)))))

;; --- JSON extraction helper ---

(defn json-extract
  "Extract the first JSON object or array from a string.
   Handles markdown fences, thinking indicators, and other LLM noise."
  [s]
  (let [trimmed (str/trim (str s))
        ;; Find first { or [
        start (some #(str/index-of trimmed (str %)) [\{ \[])]
    (if-not start
      trimmed
      (let [open-char (nth trimmed start)
            close-char (if (= open-char \{) \} \])
            ;; Walk forward tracking nesting depth
            end (loop [i (inc start) depth 1]
                  (cond
                    (>= i (count trimmed)) i
                    (zero? depth) i
                    :else (let [c (nth trimmed i)]
                            (cond
                              (= c open-char) (recur (inc i) (inc depth))
                              (= c close-char) (recur (inc i) (dec depth))
                              (= c \") ;; skip string contents
                              (recur (loop [j (inc i)]
                                       (cond
                                         (>= j (count trimmed)) j
                                         (= (nth trimmed j) \\) (recur (+ j 2))
                                         (= (nth trimmed j) \") (inc j)
                                         :else (recur (inc j))))
                                     depth)
                              :else (recur (inc i) depth)))))]
        (subs trimmed start end)))))

(defn validate-schema
  "Validate a parsed JSON map against a schema.
   Returns nil on success, or a vector of violation strings on failure."
  [parsed schema]
  (let [violations (atom [])]
    (doseq [k (:required schema)]
      (when-not (contains? parsed k)
        (swap! violations conj (str "missing required key: " k))))
    (doseq [[k expected-type] (:types schema)]
      (when (contains? parsed k)
        (let [v (get parsed k)
              ok (case expected-type
                   :string  (string? v)
                   :number  (number? v)
                   :bool    (boolean? v)
                   :array   (vector? v)
                   :object  (map? v)
                   true)]
          (when-not ok
            (swap! violations conj (str "key '" k "' expected " (name expected-type)
                                        ", got " (type v)))))))
    (doseq [[k allowed] (:enum schema)]
      (when (contains? parsed k)
        (let [v (get parsed k)]
          (when-not (some #{v} allowed)
            (swap! violations conj (str "key '" k "' value '" v
                                        "' not in " (pr-str allowed)))))))
    (let [v @violations]
      (when (seq v) v))))

(defn validate
  "Validate a step's output against a schema. Returns true on success, throws on failure."
  [step-id schema]
  (let [output (ref step-id)
        extracted (json-extract (str output))
        parsed (try
                 (json/parse-string extracted)
                 (catch Exception _
                   nil))
        violations (if (nil? parsed)
                     [(str "failed to parse JSON from step '" step-id "'")]
                     (validate-schema parsed schema))
        passed (nil? violations)]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id (str "validate:" step-id)
                 :kind "validate"
                 :output (str passed)
                 :gate-passed (if passed 1 0)
                 :artifacts (when violations
                              (json/generate-string {:violations violations}))}))
    (if passed
      true
      (throw (ex-info (str "schema-violation: " (str/join "; " violations))
                      {:kind :schema-violation
                       :step-id step-id
                       :expected schema
                       :got extracted
                       :violations violations})))))

;; --- LLM invocation ---

(defn llm [& {:keys [prompt model provider skill step-id schema retries
                      min-confidence tools agentic max-rounds] :as opts}]
  (when-not @*provider-fn*
    (throw (ex-info "llm: no provider function set — call set-provider-fn! first" {})))
  (let [full-prompt (if skill
                      (str (slurp skill) "\n\n" prompt)
                      prompt)
        max-retries (or retries 0)
        sid         (or step-id "llm")]
    (loop [attempt 0
           current-prompt full-prompt]
      (let [start    (System/nanoTime)
            result   (@*provider-fn* (assoc opts :prompt current-prompt))
            elapsed  (/ (- (System/nanoTime) start) 1e6)
            response (:response result)
            _        (when-let [g @*graph*]
                       (swap! g update-in [:stats :tokens-spent]
                              + (or (:tokens-in result) 0) (or (:tokens-out result) 0))
                       (swap! g update-in [:stats :time-spent-ms]
                              + (Math/round elapsed)))]
        (if schema
          (let [extracted  (json-extract (str response))
                parsed     (try (json/parse-string extracted) (catch Exception _ nil))
                violations (if (nil? parsed)
                             ["response is not valid JSON"]
                             (validate-schema parsed schema))
                valid?     (nil? violations)]
            (if valid?
              ;; Schema passed — check confidence
              (let [conf-val    (when (contains? parsed "confidence")
                                  (get parsed "confidence"))
                    threshold   (cond
                                  min-confidence min-confidence
                                  (and (some #{"confidence"} (:required schema))
                                       (nil? min-confidence))
                                  0.7
                                  :else nil)
                    conf-ok?    (or (nil? threshold)
                                   (nil? conf-val)
                                   (>= conf-val threshold))]
                (if conf-ok?
                  ;; All good — record and return
                  (do
                    (when-let [recorder @*step-recorder*]
                      (recorder {:step-id sid :prompt current-prompt
                                 :output response :model (or model "")
                                 :duration (Math/round elapsed) :kind "llm"
                                 :tokens-in (or (:tokens-in result) 0)
                                 :tokens-out (or (:tokens-out result) 0)
                                 :artifacts (json/generate-string {:schema_valid true})
                                 :confidence conf-val}))
                    response)
                  ;; Confidence too low — retry or throw
                  (if (< attempt max-retries)
                    (recur (inc attempt)
                           (str current-prompt "\n\nYour confidence was " conf-val
                                ", which is below the required " threshold
                                ". Think harder and be more specific."))
                    (do
                      (when-let [recorder @*step-recorder*]
                        (recorder {:step-id sid :prompt current-prompt
                                   :output response :model (or model "")
                                   :duration (Math/round elapsed) :kind "llm"
                                   :tokens-in (or (:tokens-in result) 0)
                                   :tokens-out (or (:tokens-out result) 0)
                                   :artifacts (json/generate-string
                                                {:schema_valid true :low_confidence true})
                                   :confidence conf-val}))
                      (throw (ex-info (str "low-confidence: " conf-val " < " threshold)
                                      {:kind :low-confidence
                                       :confidence conf-val
                                       :min threshold}))))))
              ;; Schema invalid — retry or throw
              (if (< attempt max-retries)
                (recur (inc attempt)
                       (str current-prompt "\n\nYour response didn't match the required schema: "
                            (str/join "; " violations) ". Please try again."))
                (do
                  (when-let [recorder @*step-recorder*]
                    (recorder {:step-id sid :prompt current-prompt
                               :output response :model (or model "")
                               :duration (Math/round elapsed) :kind "llm"
                               :tokens-in (or (:tokens-in result) 0)
                               :tokens-out (or (:tokens-out result) 0)
                               :artifacts (json/generate-string
                                            {:schema_valid false :violations violations})}))
                  (throw (ex-info (str "schema-violation: " (str/join "; " violations))
                                  {:kind :schema-violation
                                   :expected schema
                                   :got extracted
                                   :violations violations}))))))
          ;; No schema — pass through (original behavior)
          (do
            (when-let [recorder @*step-recorder*]
              (recorder {:step-id sid :prompt full-prompt
                         :output response :model (or model "")
                         :duration (Math/round elapsed) :kind "llm"
                         :tokens-in (or (:tokens-in result) 0)
                         :tokens-out (or (:tokens-out result) 0)}))
            response))))))

;; --- Grounding ---

(def ^:private grounding-prompt
  "You are a factual verification system. Compare the OUTPUT against the CONTEXT (ground truth).

Identify any claims, commands, features, or examples in the OUTPUT that are NOT directly supported by the CONTEXT.

Do not flag style issues, opinions, or reasonable inferences. Only flag factual claims that contradict or have no basis in the context.

CONTEXT:
%s

OUTPUT:
%s

Return JSON: {\"grounded\": true/false, \"unsupported\": [{\"claim\": \"exact text\", \"reason\": \"why unsupported\"}]}")

(defn grounded?
  "Verify that a step's output is factually supported by provided context.
   Options: :provider, :strict (default true), :max-unsupported (default 0)"
  [step-id context & {:keys [provider strict max-unsupported]
                       :or {strict true max-unsupported 0}}]
  (let [output   (ref step-id)
        prompt   (format grounding-prompt context output)
        response (@*provider-fn* {:prompt prompt :provider provider})
        extracted (json-extract (:response response))
        parsed   (try (json/parse-string extracted) (catch Exception _ nil))
        grounded (get parsed "grounded")
        unsupported (or (get parsed "unsupported") [])
        num-unsupported (count unsupported)
        passed   (cond
                   ;; Unparseable response — fail
                   (nil? grounded) false
                   ;; LLM says grounded — pass
                   grounded true
                   ;; LLM says not grounded — check tolerance
                   :else (<= num-unsupported max-unsupported))]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id (str "grounded:" step-id)
                 :kind "grounded"
                 :output (str passed)
                 :gate-passed (if passed 1 0)
                 :artifacts (json/generate-string
                              {:grounded grounded :unsupported unsupported})}))
    (if (or passed (not strict))
      passed
      (throw (ex-info (str "grounding-failure: " num-unsupported " unsupported claims")
                      {:kind :grounding-failure
                       :step-id step-id
                       :unsupported unsupported})))))

;; --- Consensus ---

(defn consensus
  "Run the same prompt through multiple providers and compare responses.
   Returns a JSON string with {agreed, value, votes}.
   Options: :prompt (required), :schema, :compare-key, :model, :min-confidence"
  [providers & {:keys [prompt schema compare-key model min-confidence]}]
  (when (< (count providers) 2)
    (throw (ex-info "consensus requires minimum 2 providers" {:providers providers})))
  (let [cmp-key  (or compare-key (first (:required schema)))
        _        (when-not cmp-key
                   (throw (ex-info "consensus requires :compare-key or :schema with :required"
                                   {:providers providers})))
        votes    (reduce
                   (fn [acc pname]
                     (try
                       (let [response (@*provider-fn*
                                        (cond-> {:prompt prompt :provider pname}
                                          model (assoc :model model)))
                             extracted (json-extract (:response response))
                             parsed    (try (json/parse-string extracted) (catch Exception _ nil))
                             violations (when (and schema parsed)
                                          (validate-schema parsed schema))]
                         (if (and parsed (nil? violations))
                           (conj acc {:provider pname :response parsed})
                           acc))
                       (catch Exception _
                         acc)))
                   []
                   providers)
        values   (when cmp-key
                   (map #(-> % :response (get cmp-key)
                             str str/lower-case str/trim)
                        votes))
        agreed   (and (>= (count votes) 2)
                      (apply = values))
        result   {"agreed" agreed
                  "value"  (when agreed (:response (first votes)))
                  "votes"  (mapv (fn [v] {"provider" (:provider v)
                                          "response" (:response v)})
                                 votes)}
        confidence (if (empty? votes) 0.0
                     (/ (count (filter #(= (first values)
                                           (-> % :response (get cmp-key)
                                               str str/lower-case str/trim))
                                       votes))
                        (double (count votes))))]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id "consensus"
                 :kind "consensus"
                 :output (json/generate-string result)
                 :gate-passed (if agreed 1 0)
                 :confidence confidence
                 :artifacts (json/generate-string {:votes (get result "votes")})}))
    (json/generate-string result)))
