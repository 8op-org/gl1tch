(ns glitch.api
  "REPL-first agent primitives: deftool, agent, use-provider!, use-model!
   These are injected into the user namespace by glitch.repl/start."
  (:require [glitch.core :as g]
            [clojure.pprint :as pp]))

;; ---------------------------------------------------------------------------
;; Tool registry — atom of {name-string -> tool-map}
;; tool-map: {:name :description :parameters :fn}
;; ---------------------------------------------------------------------------

(def tool-registry (atom {}))

(defn register-tool!
  "Register a tool map in the registry. Keys: :name :description :parameters :fn"
  [{:keys [name] :as tool}]
  (swap! tool-registry assoc name tool))

(defn list-tools
  "Return a sorted list of registered tool names."
  []
  (sort (keys @tool-registry)))

(defn remove-tool!
  "Remove a tool from the registry by name string. No-ops if not found."
  [name]
  (swap! tool-registry dissoc name)
  nil)
