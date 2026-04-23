(ns glitch.session
  "Session recording, slice extraction, and workflow index for
   REPL-to-workflow promotion. Sessions are persisted as EDN files
   under ~/.local/share/glitch/sessions/. The workflow index lives
   at ~/.local/share/glitch/index.edn."
  (:require [clojure.java.io :as io]
            [clojure.edn :as edn]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Paths
;; ---------------------------------------------------------------------------

(def sessions-dir
  (str (System/getProperty "user.home") "/.local/share/glitch/sessions/"))

(def index-path
  (str (System/getProperty "user.home") "/.local/share/glitch/index.edn"))

(defn- ensure-parent! [path]
  (let [parent (.getParentFile (io/file path))]
    (when (and parent (not (.exists parent)))
      (.mkdirs parent))))

(defn- ensure-dir! [dir]
  (let [f (io/file dir)]
    (when-not (.exists f)
      (.mkdirs f))))

;; ---------------------------------------------------------------------------
;; Session state (atoms)
;; ---------------------------------------------------------------------------

(def ^:dynamic *current-session*
  "Atom holding a vector of recorded session entries."
  (atom []))

(def ^:dynamic *session-id*
  "Atom holding the current session ID (timestamp-based string)."
  (atom nil))

(def ^:dynamic *slice*
  "Atom holding {:from step-id :to step-id} or nil."
  (atom nil))

;; ---------------------------------------------------------------------------
;; Session helpers
;; ---------------------------------------------------------------------------

(defn- session-path [sid]
  (str sessions-dir sid ".edn"))

(defn- gen-session-id []
  (str (System/currentTimeMillis)))

(defn- flush-session!
  "Write the current session entries to disk."
  []
  (let [sid @*session-id*]
    (when sid
      (let [path (session-path sid)]
        (ensure-dir! sessions-dir)
        (spit path (pr-str @*current-session*))))))

;; ---------------------------------------------------------------------------
;; Session functions
;; ---------------------------------------------------------------------------

(defn init-session!
  "Reset atoms and generate a new timestamp-based session-id.
   Returns the new session-id."
  []
  (let [sid (gen-session-id)]
    (reset! *current-session* [])
    (reset! *session-id* sid)
    (reset! *slice* nil)
    sid))

(defn record!
  "Append an entry map to the current session and flush to disk.
   Entry shape: {:type :step/:sh/:llm/:search/:ref,
                 :id, :args, :output, :duration, :tokens, :timestamp}"
  [entry]
  (let [entry (cond-> entry
                (nil? (:timestamp entry))
                (assoc :timestamp (System/currentTimeMillis)))]
    (swap! *current-session* conj entry)
    (flush-session!)
    entry))

(defn current-session
  "Return the current session entries."
  []
  @*current-session*)

(defn clear-session!
  "Reset the session to an empty vector and generate a new session-id."
  []
  (init-session!))

(defn set-slice!
  "Set the slice range. `opts` is a map with optional :from and :to step-ids."
  [opts]
  (reset! *slice* (select-keys opts [:from :to])))

(defn get-slice
  "Return entries within the slice range, or all entries if no slice is set.
   Matches on the :id field of each entry."
  []
  (let [entries @*current-session*
        sl      @*slice*]
    (if (or (nil? sl) (empty? sl))
      entries
      (let [from-id (:from sl)
            to-id   (:to sl)
            ;; Find the index boundaries
            idx-of  (fn [id]
                      (when id
                        (first (keep-indexed
                                (fn [i e] (when (= id (:id e)) i))
                                entries))))
            from-ix (or (idx-of from-id) 0)
            to-ix   (or (idx-of to-id) (dec (count entries)))]
        (subvec entries from-ix (min (count entries) (inc to-ix)))))))

(defn save-session!
  "Write the current session to sessions-dir/session-id.edn."
  []
  (flush-session!))

(defn load-session
  "Read a session file back by session-id. Returns a vector of entries or nil."
  [session-id]
  (let [path (session-path session-id)
        f    (io/file path)]
    (when (.exists f)
      (let [raw (slurp path)]
        (when (seq (str/trim raw))
          (edn/read-string raw))))))

;; ---------------------------------------------------------------------------
;; Workflow index
;; ---------------------------------------------------------------------------

(defn- read-index []
  (let [f (io/file index-path)]
    (if (.exists f)
      (let [raw (slurp f)]
        (if (seq (str/trim raw))
          (edn/read-string raw)
          []))
      [])))

(defn- write-index! [entries]
  (ensure-parent! index-path)
  (spit index-path (pr-str entries)))

(defn index-workflow!
  "Append a workflow entry to index.edn.
   Entry shape: {:path, :description, :promoted-from, :created, :tags}"
  [entry]
  (let [entry (cond-> entry
                (nil? (:created entry))
                (assoc :created (System/currentTimeMillis)))
        entries (conj (read-index) entry)]
    (write-index! entries)
    entry))

(defn recall-search
  "Fuzzy search the workflow index. Tokenizes the query and matches
   against :description, :tags, and :task-shape of each entry. Returns
   results sorted by match score (descending)."
  [query]
  (let [tokens  (-> (str/lower-case (str query))
                    (str/split #"\s+")
                    (->> (remove str/blank?)))
        entries (read-index)
        score   (fn [entry]
                  (let [desc   (str/lower-case (or (:description entry) ""))
                        tag-s  (str/lower-case
                                (str/join " "
                                          (map str (or (:tags entry) []))))
                        task-s (str/lower-case (or (:task-shape entry) ""))]
                    (count (filter (fn [tok]
                                    (or (str/includes? desc tok)
                                        (str/includes? tag-s tok)
                                        (str/includes? task-s tok)))
                                  tokens))))]
    (->> entries
         (map (fn [e] (assoc e ::score (score e))))
         (filter #(pos? (::score %)))
         (sort-by ::score #(compare %2 %1))
         (mapv #(dissoc % ::score)))))

(defn list-indexed
  "Return all entries from index.edn."
  []
  (read-index))

(defn record-advise!
  "Record an advisory recommendation in the current session.
   Entry shape: {:type :advise, :task, :recommendation, :followed?, :timestamp}"
  [{:keys [task recommendation]}]
  (let [entry {:type :advise
               :task task
               :recommendation recommendation
               :followed? nil
               :timestamp (System/currentTimeMillis)}]
    (swap! *current-session* conj entry)
    (flush-session!)
    entry))

(defn mark-advise-followed!
  "Update the most recent :advise entry's :followed? field."
  [followed?]
  (swap! *current-session*
         (fn [entries]
           (let [idx (some (fn [[i e]] (when (= :advise (:type e)) i))
                           (map-indexed vector (reverse entries)))]
             (if idx
               (let [real-idx (- (dec (count entries)) idx)]
                 (assoc-in entries [real-idx :followed?] followed?))
               entries))))
  (flush-session!))
