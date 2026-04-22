(ns lab-capture
  "REPL capture library for recording live lab sessions as site EDN."
  (:require [clojure.pprint :as pprint]
            [clojure.java.io :as io]))

;; ---------------------------------------------------------------------------
;; State
;; ---------------------------------------------------------------------------

(def ^:private *session
  (atom {:slug        nil
         :title       nil
         :description nil
         :duration    nil
         :date        nil
         :steps       []}))

;; ---------------------------------------------------------------------------
;; Helpers
;; ---------------------------------------------------------------------------

(defn- today []
  (str (java.time.LocalDate/now)))

(defn- form->str [form]
  (binding [*print-namespace-maps* false]
    (pr-str form)))

(defn- result->str [v]
  (binding [*print-namespace-maps* false]
    (if (string? v) v (pr-str v))))

;; ---------------------------------------------------------------------------
;; Public API
;; ---------------------------------------------------------------------------

(defn start-lab
  "Reset state and begin a new lab session.
   Required opts: :title, :description. Optional: :duration."
  [slug {:keys [title description duration]}]
  (assert (string? slug) "slug must be a string")
  (assert (string? title) ":title is required")
  (assert (string? description) ":description is required")
  (reset! *session
          {:slug        slug
           :title       title
           :description description
           :duration    (or duration "15min")
           :date        (today)
           :steps       []})
  (println (str "lab started: " slug))
  slug)

(defmacro record!
  "Capture an expression, eval it, and append a step to the current session.
   If a step with the same heading exists it is replaced (re-record)."
  [heading expr]
  `(let [src#    ~(form->str expr)
         result# ~expr
         step#   {:heading ~heading
                  :body    [[:code {:lang "clojure"} src#]
                            [:code {:lang "output"} (result->str result#)]]}]
     (swap! *session
            (fn [s#]
              (let [idx# (first (keep-indexed
                                  (fn [i# st#]
                                    (when (= (:heading st#) ~heading) i#))
                                  (:steps s#)))]
                (if idx#
                  (assoc-in s# [:steps idx#] step#)
                  (update s# :steps conj step#)))))
     result#))

(defn annotate!
  "Append a :p element to the most recent step's body."
  [prose]
  (swap! *session
         (fn [s]
           (let [steps (:steps s)
                 n     (count steps)]
             (assert (pos? n) "no steps recorded yet")
             (update-in s [:steps (dec n) :body] conj [:p prose]))))
  prose)

(defn save-lab!
  "Pretty-print the session EDN to site/content/labs/{slug}.edn.
   Returns the path string."
  []
  (let [{:keys [slug title description date duration steps]} @*session
        _    (assert (and slug (seq steps)) "no session or steps to save")
        edn  {:title       title
               :description description
               :date        date
               :duration    duration
               :steps       steps}
        dir  (io/file "site" "content" "labs")
        file (io/file dir (str slug ".edn"))]
    (.mkdirs dir)
    (with-open [w (io/writer file)]
      (pprint/pprint edn w))
    (let [path (str file)]
      (println (str "saved: " path))
      path)))
