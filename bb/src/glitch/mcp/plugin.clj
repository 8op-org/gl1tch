(ns glitch.mcp.plugin
  (:require [babashka.fs :as fs]
            [clojure.string :as str]
            [glitch.mcp.sci-bindings :as sci-bind]
            [sci.core :as sci]))

(def registry (atom {}))

(defn props->input-schema
  "Convert a properties map to MCP JSON Schema inputSchema."
  [properties]
  (let [schema {"type"       "object"
                "properties" (into {}
                               (map (fn [[k v]]
                                      [k {"type"        (:type v)
                                          "description" (:description v)}])
                                    properties))}
        required (->> properties
                      (filter (fn [[_ v]] (:required v)))
                      (mapv first))]
    (if (seq required)
      (assoc schema "required" required)
      schema)))

(defn register-tool!
  "Register an MCP tool with name, description, properties, and handler fn."
  [name description properties handler-fn]
  (let [input-schema (props->input-schema properties)
        definition   {"name"        name
                      "description" description
                      "inputSchema" input-schema}]
    (swap! registry assoc name {:definition definition
                                :handler-fn handler-fn})
    nil))

(defn defmcp-tool
  "Register an MCP tool. Intended for use inside SCI plugin files.
   Takes name, description, properties map, and a handler-fn."
  [name description properties handler-fn]
  (register-tool! name description properties handler-fn))

(defn- make-plugin-sci-ctx
  "Build a SCI context with the glitch DSL plus defmcp-tool."
  []
  (sci/init {:namespaces
             {'user           (assoc (sci-bind/user-bindings)
                                     'defmcp-tool defmcp-tool)
              'clojure.string sci-bind/string-bindings
              'str            sci-bind/string-bindings}}))

(defn load-tools!
  "Scan .glitch/mcp-tools/*.clj and evaluate each in a SCI context.
   Resets the registry before loading. Returns nil."
  []
  (reset! registry {})
  (let [dir (fs/path ".glitch" "mcp-tools")]
    (when (fs/directory? dir)
      (let [ctx   (make-plugin-sci-ctx)
            files (->> (fs/list-dir dir)
                       (filter #(str/ends-with? (str %) ".clj"))
                       sort)]
        (doseq [f files]
          (sci/eval-string* ctx (slurp (str f)))))))
  nil)

(defn tool-definitions
  "Return a vector of all registered MCP tool definition maps."
  []
  (mapv :definition (vals @registry)))

(defn handle-tool
  "Dispatch a tool call. Returns result string or nil if not found."
  [tool-name args]
  (when-let [entry (get @registry tool-name)]
    ((:handler-fn entry) args)))
