# Core workflow primitives.
# Replaces: internal/pipeline/eval.go + eval_builtins.go

# --- Workflow state ---

(var- *steps* @{})
(var- *input* "")
(var- *params* @{})
(var- *step-recorder* nil)

(defn reset-steps! []
  (set *steps* @{}))

(defn set-input! [s]
  (set *input* s))

(defn set-params! [p]
  (set *params* p))

(defn set-step-recorder! [f]
  (set *step-recorder* f))

# --- Primitives ---

(defn input []
  "Return the current workflow input string."
  *input*)

(defn params []
  "Return the current workflow params table."
  *params*)

(defn param [key &opt default]
  "Return a single param by keyword or string key."
  (or (get *params* key)
      (get *params* (string key))
      default))

(defn ref [step-id]
  "Look up a step's output by ID. Returns nil if not found."
  (get *steps* step-id))

(defn step [id body]
  "Record a step output. Returns the body value (converted to string)."
  (def val (if (string? body) body (string body)))
  (put *steps* id val)
  (when *step-recorder*
    (*step-recorder* {:step-id id :output val :kind "step"}))
  val)

(defn sh [& args]
  "Execute a shell command. Returns stdout as string. Raises on non-zero exit."
  (def proc (os/spawn args :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "sh: command %s failed (exit %d): %s"
            (string/join args " ") exit (string err-out)))
  (string out))

(defn save [path content]
  "Write content to a file. Creates parent directories."
  (def parts (string/find-all "/" path))
  (when (not (empty? parts))
    (def idx (last parts))
    (def dir (string/slice path 0 idx))
    (when (and (not= dir "") (not (os/stat dir)))
      (os/shell (string "mkdir -p " dir))))
  (spit path (if (string? content) content (string content)))
  content)

(defn read-file [path]
  "Read a file and return its contents as a string."
  (string (slurp path)))

(defn write-file [path content]
  "Alias for save."
  (save path content))

(defn get-steps []
  "Return a snapshot of all step outputs."
  (table/clone *steps*))
