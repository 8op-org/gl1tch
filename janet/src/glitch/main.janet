# CLI entry point.

(import spork/argparse :prefix "")
(import glitch/core :as g)
(import glitch/runner :as runner)
(import glitch/store :as store)
(import glitch/workspace :as ws)
(import glitch/provider :as prov)
(import glitch/gui :as gui)

(defn resolve-command [argv]
  "Return the command name from argv, or nil."
  (when (> (length argv) 0)
    (first argv)))

# --- Commands ---

(defn cmd-version []
  (print "glitch 0.1.0-janet"))

(defn cmd-check [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch check <file.janet>")
    (os/exit 1))
  (def path (first argv))
  (try
    (do (parse (string (slurp path))) (print "ok"))
    ([err]
      (eprintf "check failed: %s" (string err))
      (os/exit 1))))

(defn cmd-eval [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch eval <file.janet>")
    (os/exit 1))
  (def path (first argv))
  (def result (dofile path))
  (when result (pp result)))

(defn cmd-run [argv]
  (def res
    (argparse
      "Execute a workflow"
      "workspace" {:kind :option :short "w"
                   :help "Workspace path"}
      "path"      {:kind :option :short "p"
                   :help "Explicit workflow file path"}
      "set"       {:kind :accumulate :short "s"
                   :help "Set param key=value"}
      "model"     {:kind :option :short "m"
                   :help "Default model"}
      :default    {:kind :accumulate}))
  (unless res (os/exit 1))

  (def positional (or (res :default) @[]))
  (def wf-name (get positional 0))
  (def input (get positional 1 ""))

  (unless wf-name
    (eprint "usage: glitch run <workflow> [input]")
    (os/exit 1))

  # Resolve workspace
  (def workspace (ws/resolve
    :workspace-flag (res "workspace")))

  # Load providers
  (prov/load-providers)

  # Resolve workflow path
  (def wf-path
    (or (res "path")
        (do
          (def dirs @[".glitch/workflows" "workflows"])
          (when workspace
            (array/insert dirs 0
              (string (workspace :path) "/workflows")))
          (var found nil)
          (each dir dirs
            (def p (string dir "/" wf-name ".janet"))
            (when (and (os/stat p) (nil? found))
              (set found p)))
          (or found
              (do (eprintf "workflow not found: %s" wf-name)
                  (os/exit 1))))))

  # Parse --set params
  (def params @{})
  (when (res "set")
    (each kv (res "set")
      (def eq (string/find "=" kv))
      (when eq
        (put params
          (string/slice kv 0 eq)
          (string/slice kv (+ eq 1))))))

  # Model
  (def model (or (res "model")
                 (when workspace (get-in workspace [:defaults :model]))
                 "qwen2.5:7b"))

  # Open store
  (def db
    (if workspace
      (store/open-for-workspace (workspace :path))
      (store/open)))

  # Run
  (def result
    (runner/run wf-path input
      :db db
      :workspace (when workspace (workspace :name))
      :model model
      :params params))

  (print (result :output))
  (store/close db))

(defn- ws-init [argv]
  (def name (or (first argv) (last (string/split "/" (os/cwd)))))
  (def dir (string (os/cwd) "/.glitch"))
  (g/mkdir-p dir)
  (spit (string dir "/workspace.janet")
    (string "(import glitch/workspace :as ws)\n"
            "(ws/workspace " (string/format "%q" name) "\n"
            "  :description \"\")\n"))
  (printf "workspace %s initialized at %s" name dir))

(defn- ws-status []
  (def w (ws/resolve))
  (if w
    (do
      (printf "workspace: %s" (w :name))
      (printf "path: %s" (or (w :path) ""))
      (printf "model: %s" (get-in w [:defaults :model] ""))
      (printf "resources: %d" (length (or (w :resources) @[]))))
    (print "no active workspace")))

(defn- ws-resources []
  (def w (ws/resolve))
  (if w
    (each r (or (w :resources) @[])
      (printf "%s (%s) — %s" (r :name) (r :type)
              (or (r :url) (r :path) "")))
    (print "no active workspace")))

(defn cmd-workspace [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch workspace <init|status|resources>")
    (os/exit 1))
  (match (first argv)
    "init"      (ws-init (tuple/slice argv 1))
    "status"    (ws-status)
    "resources" (ws-resources)
    sub         (do (eprintf "workspace subcommand %s: not yet implemented" sub)
                    (os/exit 1))))

(defn cmd-plugin [argv]
  (when (= (length argv) 0)
    (eprint "usage: glitch plugin <name> [args...]")
    (os/exit 1))
  (def name (first argv))
  (def plugin-dir (string (os/getenv "HOME")
                          "/.config/glitch/plugins/" name))
  (def entry (string plugin-dir "/main.janet"))
  (unless (os/stat entry)
    (eprintf "plugin not found: %s" name)
    (os/exit 1))
  (def mod (dofile entry))
  (when (mod :main)
    ((mod :main) (tuple/slice argv 1))))

(defn cmd-gui [argv]
  (def res
    (argparse
      "Start the GUI server"
      "addr"      {:kind :option :short "a"
                   :help "Listen address (default localhost:3000)"}
      "workspace" {:kind :option :short "w"
                   :help "Workspace path"}))
  (unless res (os/exit 1))
  (def workspace (ws/resolve :workspace-flag (res "workspace")))
  (gui/start {:addr (or (res "addr") "localhost:3000")
              :workspace (when workspace (workspace :path))}))

(defn cmd-up []
  (each tool ["janet" "curl"]
    (def proc (os/spawn ["which" tool] :p {:out :pipe}))
    (def out (ev/read (proc :out) :all))
    (os/proc-wait proc)
    (if (> (length (string/trim (string out))) 0)
      (printf "  %s ok" tool)
      (printf "  %s missing" tool))))

# --- Main ---

(def commands
  {"run"       cmd-run
   "check"     cmd-check
   "eval"      cmd-eval
   "gui"       cmd-gui
   "workspace" cmd-workspace
   "plugin"    cmd-plugin
   "up"        (fn [_] (cmd-up))
   "version"   (fn [_] (cmd-version))})

(defn main [& argv]
  (def args (tuple/slice argv 1))
  (def cmd (resolve-command args))
  (if-let [handler (get commands cmd)]
    (handler (tuple/slice args 1))
    (do
      (print "glitch - workflow engine (janet)")
      (print "")
      (print "commands:")
      (each name (sorted (keys commands))
        (printf "  %s" name))
      (when (nil? cmd) (os/exit 1)))))
