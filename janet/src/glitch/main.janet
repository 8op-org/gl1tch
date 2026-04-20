# CLI entry point.

(import spork/argparse :prefix "")
(import glitch/core :as g)
(import glitch/runner :as runner)
(import glitch/store :as store)
(import glitch/project :as project)
(import glitch/provider :as prov)
(import glitch/gui :as gui)
(import glitch/mcp :as mcp)
(import glitch/mcp/indexer :as idx)
(import glitch/mcp/embeddings :as emb)

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
      "project"   {:kind :option :short "P"
                   :help "Project root path"}
      "path"      {:kind :option :short "p"
                   :help "Explicit workflow file path"}
      "set"       {:kind :accumulate :short "s"
                   :help "Set param key=value"}
      "model"     {:kind :option :short "m"
                   :help "Default model"}
      :default    {:kind :accumulate}
      :args (array/concat @["run"] argv)))
  (unless res (os/exit 1))

  (def positional (or (res :default) @[]))
  (def wf-name (get positional 0))
  (def input (get positional 1 ""))

  (unless wf-name
    (eprint "usage: glitch run <workflow> [input]")
    (os/exit 1))

  # Resolve project root
  (def project-root (project/resolve
    :flag (res "project")))

  # Load providers
  (prov/load-providers)

  # Resolve workflow path
  (def wf-path
    (or (res "path")
        (do
          (def dirs @[".glitch/workflows" "workflows"])
          (when project-root
            (array/insert dirs 0
              (string project-root "/.glitch/workflows")))
          (var found nil)
          (each dir dirs
            (each ext [".glitch" ".janet"]
              (def p (string dir "/" wf-name ext))
              (when (and (os/stat p) (nil? found))
                (set found p))))
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
  (def model (or (res "model") "gemma4"))

  # Open store
  (def db
    (if project-root
      (store/open-for-project project-root)
      (store/open)))

  # Run
  (def result
    (runner/run wf-path input
      :db db
      :project (or project-root "")
      :model model
      :params params))

  (print (result :output))
  (store/close db))

(defn cmd-init [argv]
  (def dir (string (os/cwd) "/.glitch"))
  (g/mkdir-p dir)
  (printf "project initialized at %s" dir))

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
      "addr"    {:kind :option :short "a"
                 :help "Listen address (default localhost:3000)"}
      "project" {:kind :option :short "P"
                 :help "Project root path"}))
  (unless res (os/exit 1))
  (def project-root (project/resolve :flag (res "project")))
  (gui/start {:addr (or (res "addr") "localhost:3000")
              :project-root project-root}))

(defn cmd-up []
  (each tool ["janet" "curl"]
    (def proc (os/spawn ["which" tool] :p {:out :pipe}))
    (def out (ev/read (proc :out) :all))
    (os/proc-wait proc)
    (if (> (length (string/trim (string out))) 0)
      (printf "  %s ok" tool)
      (printf "  %s missing" tool))))

(defn cmd-mcp [argv]
  (def res
    (argparse
      "Start the MCP stdio server"
      "workspace" {:kind :option :short "w"
                   :help "Workspace path"}
      "model"     {:kind :option :short "m"
                   :help "Embedding model name"}
      "base-url"  {:kind :option
                   :help "LM Studio base URL"}))
  (unless res (os/exit 1))
  (def workspace-path
    (or (res "workspace")
        (project/resolve)
        (os/cwd)))
  (mcp/start @{:workspace-path workspace-path
               :model (res "model")
               :base-url (res "base-url")}))

(defn cmd-index [argv]
  (def res
    (argparse
      "Index a repository for semantic search"
      "workspace" {:kind :option :short "w"
                   :help "Workspace path"}
      "repo"      {:kind :option :short "r"
                   :help "Repository path to index"}
      "reindex"   {:kind :flag
                   :help "Force full reindex"}
      "model"     {:kind :option :short "m"
                   :help "Embedding model name"}
      "base-url"  {:kind :option
                   :help "LM Studio base URL"}))
  (unless res (os/exit 1))
  (def workspace-path
    (or (res "workspace")
        (project/resolve)
        (os/cwd)))
  (def repo-path (or (res "repo") (os/cwd)))
  (def model (res "model"))
  (def base-url (res "base-url"))
  (def db (idx/open-search-db workspace-path))
  (def embed-fn
    (when model
      (fn [texts]
        (emb/embed texts :model model :base-url base-url))))
  (def result
    (idx/index-repo db repo-path
      :embed-fn embed-fn
      :model model
      :reindex (truthy? (res "reindex"))))
  (printf "indexed %d files, %d new chunks (of %d total files)"
    (result :files-indexed)
    (result :chunks-created)
    (result :total-files))
  (idx/close-search-db db))

# --- Main ---

(def commands
  {"run"     cmd-run
   "init"    cmd-init
   "check"   cmd-check
   "eval"    cmd-eval
   "gui"     cmd-gui
   "mcp"     cmd-mcp
   "index"   cmd-index
   "plugin"  cmd-plugin
   "up"      (fn [_] (cmd-up))
   "version" (fn [_] (cmd-version))})

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
