(ns glitch.main
  (:require [glitch.runner :as runner]
            [glitch.store :as store]
            [glitch.provider :as prov]
            [glitch.project :as project]
            [clojure.string :as str]
            [clojure.java.io :as io]))

;; ---------------------------------------------------------------------------
;; Arg parsing
;; ---------------------------------------------------------------------------

(defn parse-args
  "Parse CLI args. valid-opts is a map of flag-name -> {:short :kind}
   :kind can be :option (takes value), :flag (boolean), :accumulate (collects values).
   Returns {:opts {} :positional []}"
  [args valid-opts]
  (let [;; Build lookup: long flag -> key, short flag -> key
        lookup (reduce-kv
                 (fn [m k spec]
                   (let [long-flag  (str "--" (name k))
                         short-flag (when (:short spec) (str "-" (:short spec)))]
                     (cond-> (assoc m long-flag k)
                       short-flag (assoc short-flag k))))
                 {} valid-opts)]
    (loop [remaining (vec args)
           opts      {}
           positional []]
      (if (empty? remaining)
        {:opts opts :positional positional}
        (let [arg (first remaining)
              key (get lookup arg)]
          (if key
            (let [kind (get-in valid-opts [key :kind] :option)]
              (case kind
                :flag
                (recur (subvec remaining 1)
                       (assoc opts key true)
                       positional)

                :accumulate
                (if (< 1 (count remaining))
                  (recur (subvec remaining 2)
                         (update opts key (fnil conj []) (second remaining))
                         positional)
                  (recur (subvec remaining 1) opts positional))

                ;; :option (default)
                (if (< 1 (count remaining))
                  (recur (subvec remaining 2)
                         (assoc opts key (second remaining))
                         positional)
                  (recur (subvec remaining 1) opts positional))))
            ;; Not a known flag — positional arg
            (recur (subvec remaining 1) opts (conj positional arg))))))))

;; ---------------------------------------------------------------------------
;; Workflow file resolution
;; ---------------------------------------------------------------------------

(defn- resolve-workflow
  "Find a workflow file by name. Searches the given directories for
   <name>.glitch or <name>.clj. Returns the path string or nil."
  [name search-dirs]
  (some (fn [dir]
          (some (fn [ext]
                  (let [f (io/file dir (str name ext))]
                    (when (.exists f) (.getPath f))))
                [".glitch" ".clj"]))
        search-dirs))

;; ---------------------------------------------------------------------------
;; Commands
;; ---------------------------------------------------------------------------

(defn- cmd-run [args]
  (let [{:keys [opts positional]}
        (parse-args args {:project {:short "P" :kind :option}
                          :path    {:short "p" :kind :option}
                          :set     {:short "s" :kind :accumulate}
                          :model   {:short "m" :kind :option}
                          :help    {:short "h" :kind :flag}})
        wf-name (first positional)
        input   (str/join " " (rest positional))]
    (when (or (:help opts) (and (nil? wf-name) (nil? (:path opts))))
      (println "Usage: glitch run [options] <workflow> [input...]")
      (println)
      (println "Options:")
      (println "  -P, --project <path>   Project root path")
      (println "  -p, --path <file>      Explicit workflow file path")
      (println "  -s, --set <key=value>  Set parameter (repeatable)")
      (println "  -m, --model <model>    Default model name")
      (System/exit (if (:help opts) 0 1)))

    (let [project-root (project/resolve :flag (:project opts))
          _            (prov/load-providers)
          ;; Resolve workflow path
          wf-path      (or (:path opts)
                           (when wf-name
                             (let [search-dirs (cond-> []
                                                project-root
                                                (conj (str project-root "/.glitch/workflows"))
                                                true (conj ".glitch/workflows" "workflows"))]
                               (resolve-workflow wf-name search-dirs)))
                           (do (println (str "error: workflow not found: " wf-name))
                               (System/exit 1)))
          ;; Parse --set params into a map
          params       (reduce (fn [m pair]
                                 (let [idx (str/index-of pair "=")]
                                   (if idx
                                     (assoc m
                                       (subs pair 0 idx)
                                       (subs pair (inc idx)))
                                     (assoc m pair ""))))
                               {} (:set opts))
          ;; Open store
          db           (try
                         (if project-root
                           (store/open-for-project project-root)
                           (store/open))
                         (catch Exception e
                           (binding [*out* *err*]
                             (println (str "warn: store unavailable: " (.getMessage e))))
                           nil))
          ;; Run the workflow
          result       (runner/run wf-path
                         :input   (if (str/blank? input) nil input)
                         :db      db
                         :project (or project-root "")
                         :model   (:model opts)
                         :params  params
                         :workflows-dir
                         (when wf-path
                           (.getParent (io/file wf-path))))]
      ;; Print output
      (when-let [output (:output result)]
        (when-not (str/blank? output)
          (println output)))
      ;; Close store
      (when db (store/close db)))))

(defn- cmd-check [args]
  (let [path (first args)]
    (when-not path
      (println "Usage: glitch check <file>")
      (System/exit 1))
    (try
      (read-string (slurp path))
      (println "ok")
      (catch Exception e
        (println (str "error: " (.getMessage e)))
        (System/exit 1)))))

(defn- cmd-eval [args]
  (let [path (first args)]
    (when-not path
      (println "Usage: glitch eval <file>")
      (System/exit 1))
    (load-file path)))

(defn- cmd-init [_args]
  (let [d (io/file ".glitch/workflows")]
    (if (.exists d)
      (println ".glitch directory already exists")
      (do (.mkdirs d)
          (println "created .glitch/workflows/")))))

(defn- cmd-up []
  (println "checking required tools...")
  (let [tools ["bb" "curl" "gh"]
        results (map (fn [t]
                       (try
                         (let [proc (-> (ProcessBuilder. ["which" t])
                                        (.redirectErrorStream true)
                                        .start)]
                           (.waitFor proc)
                           (if (zero? (.exitValue proc))
                             {:tool t :ok true}
                             {:tool t :ok false}))
                         (catch Exception _
                           {:tool t :ok false})))
                     tools)]
    (doseq [{:keys [tool ok]} results]
      (println (str "  " (if ok "[ok]" "[MISSING]") " " tool)))
    (when (some #(not (:ok %)) results)
      (System/exit 1))
    (println "all tools available")))

(defn- cmd-plugin [args]
  (let [name (first args)]
    (when-not name
      (println "Usage: glitch plugin <name> [args...]")
      (System/exit 1))
    (let [home   (System/getProperty "user.home")
          path   (str home "/.config/glitch/plugins/" name "/main.clj")
          f      (io/file path)]
      (if (.exists f)
        (load-file path)
        (do (println (str "error: plugin not found: " path))
            (System/exit 1))))))

;; ---------------------------------------------------------------------------
;; Entry point
;; ---------------------------------------------------------------------------

(defn -main [& args]
  (let [cmd       (first args)
        rest-args (rest args)]
    (case cmd
      "run"     (cmd-run rest-args)
      "check"   (cmd-check rest-args)
      "eval"    (cmd-eval rest-args)
      "init"    (cmd-init rest-args)
      "up"      (cmd-up)
      "version" (println "glitch 0.2.0-bb")
      "plugin"  (cmd-plugin rest-args)
      "mcp"     (do
                  (require '[glitch.mcp :as mcp])
                  ((resolve 'glitch.mcp/start)
                   {:workspace-path (System/getProperty "user.dir")}))
      "gui"     (let [{:keys [opts]} (parse-args rest-args
                                        {:port      {:short "p" :kind :option}
                                         :workspace {:short "w" :kind :option}
                                         :dist      {:short "d" :kind :option}})
                      port (or (some-> (:port opts) parse-long) 3000)
                      workspace (or (:workspace opts)
                                    (project/resolve))
                      dist-dir (or (:dist opts) "gui/dist")]
                  (require '[glitch.gui :as gui])
                  ((resolve 'glitch.gui/start)
                   {:port port :workspace workspace :dist-dir dist-dir})
                  @(promise))
      (do (println "glitch - workflow engine (babashka)")
          (println)
          (doseq [c (sort ["check" "eval" "gui" "init" "mcp" "plugin" "run" "up" "version"])]
            (println (str "  " c)))
          (when-not cmd (System/exit 1))))))

