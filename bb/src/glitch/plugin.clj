(ns glitch.plugin
  (:require [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Plugin registry — each entry is
;; {plugin-name -> {:name :description :commands {cmd-name -> {:fn :args :description}}}}
;; ---------------------------------------------------------------------------

(def registry (atom {}))

(defn reset!
  "Clear all registered plugins."
  []
  (clojure.core/reset! registry {}))

(defn register
  "Register a plugin map under `name`.
   plugin-map should be {:name :description :commands {cmd-name -> {:fn :args :description}}}."
  [name plugin-map]
  (swap! registry assoc name plugin-map))

(defmacro defcommand
  "Define a plugin command and register it in the plugin registry.

   Usage inside a glitch.plugin.<name> namespace:

     (defcommand health
       \"Check cluster health\"
       {:args [{:name \"cluster\" :description \"Cluster URL\" :required true}]}
       [opts]
       (let [resp (http-get (str (:cluster opts) \"/_cluster/health\"))]
         (json-parse (:body resp))))

   Expands to a defn plus a registry swap that files the command under
   the plugin name derived from the current namespace."
  [cmd-name docstring meta-map args & body]
  (let [cmd-str (str cmd-name)]
    `(do
       (defn ~cmd-name ~docstring ~args ~@body)
       (let [ns-name# (str (ns-name *ns*))
             plugin#  (if (str/starts-with? ns-name# "glitch.plugin.")
                        (subs ns-name# (count "glitch.plugin."))
                        ns-name#)]
         (swap! registry update plugin#
                (fn [p#]
                  (-> (or p# {:name plugin# :description "" :commands {}})
                      (assoc-in [:commands ~cmd-str]
                                {:fn         (var ~cmd-name)
                                 :args        (:args ~meta-map)
                                 :description ~docstring}))))))))

(defn lookup [plugin-name cmd-name]
  "Return the command map for `cmd-name` inside `plugin-name`, or nil."
  (get-in @registry [plugin-name :commands cmd-name]))

(defn commands [plugin-name]
  "Return all commands for `plugin-name` as {cmd-name -> cmd-map}."
  (get-in @registry [plugin-name :commands]))

(defn all-plugins []
  "Return the full registry."
  @registry)
