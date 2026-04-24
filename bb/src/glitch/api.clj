(ns glitch.api
  "REPL-first agent primitives: deftool, agent, use-provider!, use-model!
   Intentionally separate from glitch.mcp.plugin/registry — this registry
   is for ad-hoc REPL tool definitions, not MCP server tools.
   Injected into the user namespace by glitch.repl/start."
  (:require [glitch.core :as g]))

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

;; ---------------------------------------------------------------------------
;; Agent — agentic tool loop
;; ---------------------------------------------------------------------------

(defn- tool->mcp-def
  "Convert a registry tool map to MCP-format for the provider."
  [{:keys [name description parameters]}]
  {"name"        name
   "description" description
   "inputSchema" parameters})

(defn- make-tool-handler
  "Build a dispatch fn that calls the registered tool fn by name.
   Traces each invocation to stderr."
  []
  (fn [tool-name args]
    (let [tool (get @tool-registry tool-name)]
      (when-not tool
        (throw (ex-info (str "agent: unknown tool: " tool-name) {:tool tool-name})))
      (g/trace "  tool:" tool-name)
      (str ((:fn tool) args)))))

(defn agent
  "Run an agentic LLM loop with the given tools.

   Usage:
     (agent [\"search\" \"run-tests\"] \"why is ci failing?\")
     (agent {:system \"you are a senior engineer\"} [\"search\"] \"what broke?\")

   tool-names — vector of string tool names registered via deftool/register-tool!
   prompt     — user prompt string
   opts       — optional map: :system :model :provider :max-rounds (default 10)

   Streams tool call names to stderr. Returns final text string."
  ([tool-names prompt]
   (agent {} tool-names prompt))
  ([opts tool-names prompt]
   (let [tool-defs    (mapv (fn [n]
                              (let [t (get @tool-registry n)]
                                (when-not t
                                  (throw (ex-info (str "agent: tool not registered: " n)
                                                  {:tool n})))
                                (tool->mcp-def t)))
                            tool-names)
         tool-handler (make-tool-handler)
         call-opts    (merge {:agentic    true
                              :max-rounds 10}
                             opts
                             {:prompt       prompt
                              :tool-defs    tool-defs
                              :tool-handler tool-handler})]
     (apply g/llm (apply concat call-opts)))))
