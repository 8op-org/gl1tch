(ns glitch.main
  (:require [glitch.runner :as runner]
            [glitch.store :as store]
            [glitch.provider :as prov]
            [glitch.plugin :as plugin]
            [glitch.plugin-loader :as plugin-loader]
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
;; Commands
;; ---------------------------------------------------------------------------

(defn- cmd-run [args]
  (let [{:keys [opts positional]}
        (parse-args args {:provider {:short "p" :kind :option}
                          :set      {:short "s" :kind :accumulate}
                          :model    {:short "m" :kind :option}
                          :help     {:short "h" :kind :flag}})
        wf-path (first positional)
        input   (str/join " " (rest positional))]
    (when (or (:help opts) (nil? wf-path))
      (println "Usage: glitch run [options] <file> [input...]")
      (println)
      (println "Options:")
      (println "  -p, --provider <name>   Default provider for LLM calls")
      (println "  -s, --set <key=value>   Set parameter (repeatable)")
      (println "  -m, --model <model>     Default model name")
      (System/exit (if (:help opts) 0 1)))

    (let [f (io/file wf-path)]
      (when-not (.exists f)
        (println (str "error: file not found: " wf-path))
        (System/exit 1)))

    (let [_      (prov/load-providers)
          ;; Parse --set params into a map
          params (reduce (fn [m pair]
                           (let [idx (str/index-of pair "=")]
                             (if idx
                               (assoc m
                                 (subs pair 0 idx)
                                 (subs pair (inc idx)))
                               (assoc m pair ""))))
                         {} (:set opts))
          ;; Open store
          db     (try
                   (store/open)
                   (catch Exception e
                     (binding [*out* *err*]
                       (println (str "warn: store unavailable: " (.getMessage e))))
                     nil))
          ;; Run the workflow
          result (runner/run wf-path
                   :input    (if (str/blank? input) nil input)
                   :db       db
                   :provider (:provider opts)
                   :model    (:model opts)
                   :params   params
                   :workflows-dir (.getParent (io/file wf-path)))]
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

(defn- cmd-up []
  (println "checking required tools...")
  (let [tools   ["bb" "curl" "gh" "rg"]
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

(defn- args-meta->valid-opts
  "Convert a command's :args metadata into the valid-opts format for parse-args.
   {:name \"repo\" :required true} -> {:repo {:kind :option}}
   {:name \"verbose\" :type :flag} -> {:verbose {:kind :flag}}"
  [args-meta]
  (reduce (fn [m arg]
            (let [k    (keyword (:name arg))
                  kind (if (= :flag (:type arg)) :flag :option)]
              (assoc m k {:kind kind})))
          {} (or args-meta [])))

(defn- run-plugin-command
  "Look up and run a plugin command. plugin-name is the plugin, cmd-name is the
   command within it, and cmd-args are the remaining CLI args to parse."
  [plugin-name cmd-name cmd-args]
  (let [cmd-map (plugin/lookup plugin-name cmd-name)]
    (when-not cmd-map
      (let [cmds (plugin/commands plugin-name)]
        (if cmds
          (do (println (str "error: unknown command '" cmd-name
                            "' for plugin '" plugin-name "'"))
              (println "available commands:")
              (doseq [c (sort (keys cmds))]
                (let [desc (get-in cmds [c :description] "")]
                  (println (str "  " c (when-not (str/blank? desc)
                                         (str "  - " desc))))))
              (System/exit 1))
          (do (println (str "error: plugin '" plugin-name "' not found"))
              (System/exit 1)))))
    (let [valid-opts  (args-meta->valid-opts (:args cmd-map))
          {:keys [opts]} (parse-args cmd-args valid-opts)
          result      ((:fn cmd-map) opts)]
      (when result
        (if (string? result)
          (println result)
          (prn result))))))

(defn- cmd-plugin [args]
  (plugin-loader/load-plugins)
  (let [sub (first args)]
    (cond
      ;; glitch plugin list
      (= sub "list")
      (let [plugins (plugin/all-plugins)]
        (if (empty? plugins)
          (println "no plugins registered")
          (doseq [[pname pmap] (sort-by key plugins)]
            (println (str "  " pname))
            (doseq [[cname cmap] (sort-by key (:commands pmap))]
              (let [desc (:description cmap "")]
                (println (str "    " cname
                              (when-not (str/blank? desc)
                                (str "  - " desc)))))))))

      ;; glitch plugin <name> <cmd> [args...]
      sub
      (let [plugin-name sub
            cmd-name    (second args)
            cmd-args    (drop 2 args)]
        (when-not cmd-name
          (let [cmds (plugin/commands plugin-name)]
            (if cmds
              (do (println (str "plugin: " plugin-name))
                  (println "commands:")
                  (doseq [[cname cmap] (sort-by key cmds)]
                    (let [desc (:description cmap "")]
                      (println (str "  " cname
                                    (when-not (str/blank? desc)
                                      (str "  - " desc))))))
                  (System/exit 0))
              (do (println (str "error: plugin '" plugin-name "' not found"))
                  (System/exit 1)))))
        (run-plugin-command plugin-name cmd-name (vec cmd-args)))

      ;; glitch plugin (no args)
      :else
      (do (println "Usage: glitch plugin <name> <command> [args...]")
          (println "       glitch plugin list")
          (System/exit 1)))))

(defn- cmd-repl [args]
  (let [{:keys [opts]} (parse-args args {:port {:short "p" :kind :option}})
        port (or (some-> (:port opts) parse-long) 1667)]
    (require '[glitch.repl :as repl])
    ((resolve 'glitch.repl/start) {:port port})))

;; ---------------------------------------------------------------------------
;; Entry point
;; ---------------------------------------------------------------------------

(defn -main [& args]
  ;; Load plugins once before dispatch so plugin names are available
  (plugin-loader/load-plugins)
  (let [cmd          (first args)
        rest-args    (rest args)
        plugin-names (set (plugin-loader/plugin-names))]
    (cond
      ;; Built-in commands
      (= cmd "run")     (cmd-run rest-args)
      (= cmd "check")   (cmd-check rest-args)
      (= cmd "eval")    (cmd-eval rest-args)
      (= cmd "up")      (cmd-up)
      (= cmd "version") (println "glitch 0.3.0")
      (= cmd "plugin")  (cmd-plugin rest-args)
      (= cmd "repl")    (cmd-repl rest-args)
      (= cmd "mcp")     (do
                           (require '[glitch.mcp :as mcp])
                           ((resolve 'glitch.mcp/start) {}))

      ;; Plugin name as top-level command: glitch github fetch-issue ...
      (and cmd (contains? plugin-names cmd))
      (let [plugin-name cmd
            cmd-name    (first rest-args)
            cmd-args    (vec (rest rest-args))]
        (if cmd-name
          (run-plugin-command plugin-name cmd-name cmd-args)
          ;; Just the plugin name — show its commands
          (let [cmds (plugin/commands plugin-name)]
            (println (str "plugin: " plugin-name))
            (println "commands:")
            (doseq [[cname cmap] (sort-by key cmds)]
              (let [desc (:description cmap "")]
                (println (str "  " cname
                              (when-not (str/blank? desc)
                                (str "  - " desc))))))
            (System/exit 0))))

      ;; Help / unknown command
      :else
      (do (println "glitch - workflow engine (babashka)")
          (println)
          (println "built-in commands:")
          (doseq [c (sort ["check" "eval" "mcp" "plugin" "repl" "run" "up" "version"])]
            (println (str "  " c)))
          (when (seq plugin-names)
            (println)
            (println "plugins:")
            (doseq [p (sort plugin-names)]
              (println (str "  " p))))
          (when-not cmd (System/exit 1))))))
