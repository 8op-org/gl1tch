(ns glitch.core
  (:require [babashka.process :as bp]
            [clojure.string :as str]
            [clojure.java.io :as io]))

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

(defn step [id body]
  (let [val (str body)]
    (swap! *steps* assoc id val)
    (swap! *step-order* conj id)
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id id :output val :kind "step"}))
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

;; --- LLM invocation ---

(defn llm [& {:keys [prompt model provider skill step-id tools agentic max-rounds] :as opts}]
  (when-not @*provider-fn*
    (throw (ex-info "llm: no provider function set — call set-provider-fn! first" {})))
  (let [full-prompt (if skill
                      (str (slurp skill) "\n\n" prompt)
                      prompt)
        start (System/nanoTime)
        result (@*provider-fn* (assoc opts :prompt full-prompt))
        elapsed-ms (/ (- (System/nanoTime) start) 1e6)]
    (when-let [recorder @*step-recorder*]
      (recorder {:step-id (or step-id "llm")
                 :prompt full-prompt
                 :output (:response result)
                 :model (or model "")
                 :duration (Math/round elapsed-ms)
                 :kind "llm"
                 :tokens-in (or (:tokens-in result) 0)
                 :tokens-out (or (:tokens-out result) 0)}))
    (:response result)))

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
