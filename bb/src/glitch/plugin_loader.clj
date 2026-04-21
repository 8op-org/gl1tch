(ns glitch.plugin-loader
  "Discovers and loads plugins from disk into the plugin registry.

   Two plugin types:
     1. .glitch file plugins — each .glitch file becomes a command, evaluated
        via glitch.runner/run at invocation time.
     2. .clj namespace plugins — self-register via glitch.plugin/defcommand,
        loaded via load-file.

   Discovery paths (checked in order):
     - .glitch/plugins/   (project-local, relative to cwd)
     - ~/.config/glitch/plugins/  (global)
     - any extra dirs passed to load-plugins"
  (:require [glitch.plugin :as plugin]
            [glitch.runner :as runner]
            [clojure.java.io :as io]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; .glitch file parsing helpers
;; ---------------------------------------------------------------------------

(defn- parse-description
  "Extract the first line comment as a description, or empty string."
  [source]
  (let [first-line (first (str/split-lines source))]
    (if (and first-line (str/starts-with? (str/trim first-line) ";;"))
      (str/trim (subs (str/trim first-line) 2))
      "")))

(defn- parse-args
  "Extract (arg ...) forms from source via regex. Returns a vec of arg maps."
  [source]
  (let [pattern #"\(arg\s+\"([^\"]+)\"(?:\s+\"([^\"]*)\")?\s*(?::required\s+(true|false))?\)"
        matches (re-seq pattern source)]
    (mapv (fn [[_ name desc required]]
            (cond-> {:name name}
              desc     (assoc :description desc)
              required (assoc :required (= required "true"))))
          (or matches []))))

;; ---------------------------------------------------------------------------
;; .glitch file plugin loader (Path 1)
;; ---------------------------------------------------------------------------

(defn- load-glitch-plugin
  "Load a directory of .glitch files as a plugin. Each .glitch file (except
   plugin.glitch) becomes a command that delegates to glitch.runner/run."
  [plugin-name dir]
  (let [files  (->> (.listFiles dir)
                    (filter #(str/ends-with? (.getName %) ".glitch"))
                    (remove #(= (.getName %) "plugin.glitch"))
                    sort)
        cmds   (reduce
                 (fn [acc f]
                   (try
                     (let [fname   (.getName f)
                           cmd-name (subs fname 0 (- (count fname) 7)) ;; strip .glitch
                           source   (slurp f)
                           path     (.getAbsolutePath f)]
                       (assoc acc cmd-name
                              {:fn          (fn [opts]
                                              (runner/run path
                                                :params (or opts {})))
                               :args        (parse-args source)
                               :description (parse-description source)}))
                     (catch Exception e
                       (binding [*out* *err*]
                         (println (str "warn: failed to load plugin command "
                                       plugin-name "/" (.getName f)
                                       ": " (.getMessage e))))
                       acc)))
                 {}
                 files)]
    (when (seq cmds)
      (plugin/register plugin-name
                       {:name        plugin-name
                        :description ""
                        :commands    cmds}))))

;; ---------------------------------------------------------------------------
;; .clj namespace plugin loader (Path 2)
;; ---------------------------------------------------------------------------

(defn- load-clj-plugin
  "Load .clj files from a plugin directory. Each file self-registers via
   glitch.plugin/defcommand."
  [plugin-name dir]
  (doseq [f (sort (.list dir))]
    (when (str/ends-with? f ".clj")
      (try
        (load-file (str (.getPath dir) "/" f))
        (catch Exception e
          (binding [*out* *err*]
            (println (str "warn: failed to load plugin "
                          plugin-name "/" (subs f 0 (- (count f) 4))
                          ": " (.getMessage e)))))))))

;; ---------------------------------------------------------------------------
;; Main entry point
;; ---------------------------------------------------------------------------

(def ^:private loaded? (atom false))

(defn- do-load-plugins
  "Internal: scan plugin directories and load all discovered plugins."
  [dirs]
  (let [home        (System/getProperty "user.home")
        search-dirs (concat
                      [".glitch/plugins"
                       (str home "/.config/glitch/plugins")]
                      dirs)]
    (doseq [dir search-dirs]
      (let [d (io/file dir)]
        (when (.isDirectory d)
          (doseq [subdir-name (sort (.list d))]
            (let [subdir (io/file d subdir-name)]
              (when (.isDirectory subdir)
                (let [plugin-name subdir-name
                      files       (.list subdir)
                      has-clj?    (some #(str/ends-with? % ".clj") files)
                      has-glitch? (some #(and (str/ends-with? % ".glitch")
                                              (not= % "plugin.glitch")) files)]
                  ;; Path 2: .clj files self-register
                  (when has-clj?
                    (load-clj-plugin plugin-name subdir))
                  ;; Path 1: .glitch files get wrapped and registered
                  (when has-glitch?
                    (load-glitch-plugin plugin-name subdir)))))))))))

(defn reload-plugins!
  "Force reload all plugins, clearing the registry first."
  [& dirs]
  (plugin/reset!)
  (clojure.core/reset! loaded? false)
  (do-load-plugins dirs)
  (clojure.core/reset! loaded? true))

(defn load-plugins
  "Scan plugin directories and load all discovered plugins.
   No-ops if plugins have already been loaded.

   Default search dirs:
     - .glitch/plugins/  (project-local)
     - ~/.config/glitch/plugins/  (global)
   Plus any extra dirs passed as arguments."
  [& dirs]
  (when-not @loaded?
    (do-load-plugins dirs)
    (clojure.core/reset! loaded? true)))

(defn plugin-names
  "Return a sorted seq of registered plugin names. Loads plugins first."
  []
  (load-plugins)
  (sort (keys (plugin/all-plugins))))
