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

# --- Workflow macro ---

(defmacro workflow [name & body]
  "Define and execute a workflow. Returns {:name :output :steps}."
  ~(do
     (var wf-last nil)
     ,;(seq [form :in body
             :when (not (keyword? form))]
         ~(set wf-last ,form))
     {:name ,name
      :output (string wf-last)
      :steps (,get-steps)}))

# --- Parallel execution ---

(defmacro par [& forms]
  "Execute forms concurrently via ev/gather. Returns array of results."
  ~(ev/gather ,;forms))

# --- Retry ---

(defmacro retry [n & body]
  "Retry body up to n times."
  ~(do
     (var last-err nil)
     (var result nil)
     (var succeeded false)
     (for i 0 ,n
       (try
         (do (set result (do ,;body))
             (set succeeded true)
             (break))
         ([err] (set last-err err))))
     (if succeeded
       result
       (error (string "retry exhausted after " ,n " attempts: " last-err)))))

# --- Timeout ---

(defmacro with-timeout [seconds & body]
  "Execute body with a timeout. Raises on expiry."
  ~(ev/with-deadline ,seconds (do ,;body)))

# --- Gate ---

(defn gate [id predicate]
  "Assert a gate condition. Records result."
  (when *step-recorder*
    (*step-recorder* {:step-id id :output (string predicate)
                      :kind "gate" :gate-passed (if predicate 1 0)}))
  (put *steps* id (string predicate))
  predicate)

# --- Phase ---

(defmacro phase [name & body]
  "Group steps under a named phase. Returns last value."
  ~(do
     (when ,*step-recorder*
       (,*step-recorder* {:step-id ,name :kind "phase" :output "started"}))
     (var phase-result nil)
     ,;(map (fn [f] ~(set phase-result ,f)) body)
     (when ,*step-recorder*
       (,*step-recorder* {:step-id ,name :kind "phase"
                         :output (string phase-result)}))
     phase-result))

# --- Call-workflow ---

(var- *call-stack* @[])
(var- *workflows-dir* ".")

(defn set-workflows-dir! [d]
  (set *workflows-dir* d))

(defn call-workflow [name &named input set]
  "Execute another workflow file by name. Detects cycles."
  (when (find |(= $ name) *call-stack*)
    (errorf "call-workflow cycle: %s already on stack %s"
            name (string/join *call-stack* " -> ")))
  (array/push *call-stack* name)
  (defer (array/pop *call-stack*)
    (def path (string *workflows-dir* "/" name ".janet"))
    (unless (os/stat path)
      (errorf "call-workflow: %s not found" path))
    (def saved-input *input*)
    (def saved-steps *steps*)
    (set *input* (or input ""))
    (set *steps* @{})
    (when set
      (eachp [k v] set
        (put *params* k v)))
    (dofile path)
    (def child-steps (table/clone *steps*))
    (var last-output "")
    (eachp [_ v] child-steps (set last-output v))
    (def result {:output last-output :steps child-steps})
    (set *input* saved-input)
    (set *steps* saved-steps)
    result))
