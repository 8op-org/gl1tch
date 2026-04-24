(ns glitch.api
  "REPL-first agent primitives: deftool, agent, use-provider!, use-model!
   Intentionally separate from glitch.mcp.plugin/registry — this registry
   is for ad-hoc REPL tool definitions, not MCP server tools.
   Injected into the user namespace by glitch.repl/start.")

;; ---------------------------------------------------------------------------
;; Tool registry — atom of {name-string -> tool-map}
;; tool-map: {:name :description :parameters :fn}
;; ---------------------------------------------------------------------------

(def tool-registry (atom {}))

(defn register-tool!
  "Register a tool map in the registry. Keys: :name :description :parameters :fn"
  [{:keys [name] :as tool}]
  (swap! tool-registry assoc name tool)
  nil)

(defn list-tools
  "Return a sorted list of registered tool names."
  []
  (sort (keys @tool-registry)))

(defn remove-tool!
  "Remove a tool from the registry by name string. No-ops if not found."
  [name]
  (swap! tool-registry dissoc name)
  nil)

(defmacro deftool
  "Define and register a tool.

   Usage:
     (deftool search [query]
       \"Search the codebase index\"
       (index-query query))

   Each symbol in the arg vector becomes a :type \"string\" parameter.
   The body is called with a string-keyed args map — args are bound by name.
   Returns the tool name as a string."
  [tool-name args docstring & body]
  (let [arg-strs  (mapv str args)
        props     (into {} (map (fn [a] [a {:type "string"}]) arg-strs))
        params    {:type "object" :properties props :required arg-strs}
        name-str  (str tool-name)
        args-sym  (gensym "args")]
    `(do
       (register-tool!
         {:name        ~name-str
          :description ~docstring
          :parameters  ~params
          :fn          (fn [~args-sym]
                         (let [~@(mapcat (fn [a] [(symbol a) `(get ~args-sym ~a)]) arg-strs)]
                           ~@body))})
       ~name-str)))
