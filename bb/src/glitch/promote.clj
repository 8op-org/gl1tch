(ns glitch.promote
  "Convert a recorded REPL session into a clean .glitch workflow file
   using an LLM to distil the session into reusable DSL steps."
  (:require [clojure.string :as str]
            [glitch.core :as g]))

;; --- Helpers ---

(defn slugify
  "Convert a description string into a filename-safe slug.
   Lowercase, spaces to hyphens, strip non-alphanumeric chars."
  [s]
  (-> (str s)
      str/lower-case
      str/trim
      (str/replace #"\s+" "-")
      (str/replace #"[^a-z0-9\-]" "")
      (str/replace #"-{2,}" "-")
      (str/replace #"^-|-$" "")))

;; --- Session formatting ---

(defn- format-entry
  "Format a single session entry map into a readable text block."
  [{:keys [type id args output duration]}]
  (str (when id (str "[" id "] "))
       (or type "unknown")
       (when args (str " | " (pr-str args)))
       (when duration (str " (" duration "ms)"))
       (when output (str "\n  => " (str/replace (str output) #"\n" "\n     ")))))

(defn- format-session
  "Format a vector of session entry maps into a text summary for the LLM."
  [entries]
  (str/join "\n\n" (map-indexed
                     (fn [i entry]
                       (str (inc i) ". " (format-entry entry)))
                     entries)))

;; --- Prompt ---

(def ^:private promote-prompt
  "Turn this REPL session into a clean .glitch workflow file.

Rules:
- Use the glitch DSL: step, sh, llm, search, ref, input, params, param
- Parameterize obvious inputs (file paths, URLs, search terms) using (input) or (param :key default)
- Keep the step IDs from the original session where they exist
- Remove false starts, errors, and exploratory dead ends
- The workflow should be clean, readable, and reusable
- Output ONLY the workflow source code, no markdown fences or commentary

Session entries:
%s

Description of what this workflow does: %s")

;; --- Public API ---

(defn promote!
  "Promote a REPL session into a .glitch workflow file via LLM.

   Arguments:
     description — human description of what the workflow does

   Options (keyword args):
     :session — vector of session entry maps (keys: :type :id :args :output :duration)
     :from    — start step-id for slicing (currently unused, reserved)
     :to      — end step-id for slicing (currently unused, reserved)
     :model   — LLM model override
     :provider — LLM provider override

   Returns a map:
     {:workflow-source <string>  :path <suggested-path>}"
  [description & {:keys [session from to model provider]}]
  (when (or (nil? session) (empty? session))
    (throw (ex-info "promote!: :session is required (vector of entry maps)" {})))
  (let [entries  (cond
                   ;; Slice by :from / :to when both are provided
                   (and from to)
                   (let [ids   (mapv :id session)
                         raw-s (.indexOf ids from)
                         start (if (neg? raw-s) 0 raw-s)
                         raw-e (.indexOf ids to)
                         end   (inc (if (neg? raw-e) (dec (count session)) raw-e))]
                     (subvec (vec session) (max start 0) (min end (count session))))

                   :else session)
        summary  (format-session entries)
        prompt   (format promote-prompt summary description)
        response (g/llm :prompt prompt
                        :model model
                        :provider provider
                        :step-id "promote")
        ;; Strip markdown fences if the LLM wraps the output anyway
        cleaned  (-> (str response)
                     (str/replace #"^```[a-z]*\n?" "")
                     (str/replace #"\n?```\s*$" "")
                     str/trim)
        slug     (slugify description)
        path     (str ".glitch/workflows/" slug ".glitch")
        tags     (try
                   (let [tag-str (g/llm :prompt (str "Extract 3-6 short keyword tags from this description. "
                                                     "Return only a comma-separated list, nothing else.\n\n"
                                                     description)
                                        :model model
                                        :provider provider
                                        :step-id "promote-tags")]
                     (mapv (comp str/trim str/lower-case) (str/split (str tag-str) #",")))
                   (catch Exception _ []))
        task-shape (let [advise-entries (filter #(= :advise (:type %)) (or session []))]
                     (when (seq advise-entries)
                       (:task (last advise-entries))))]
    {:workflow-source cleaned
     :path           path
     :tags           tags
     :task-shape     task-shape}))
