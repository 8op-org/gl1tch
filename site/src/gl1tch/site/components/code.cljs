(ns gl1tch.site.components.code
  (:require ["highlight.js/lib/core" :as hljs]
            ["highlight.js/lib/languages/bash" :as bash]
            ["highlight.js/lib/languages/clojure" :as clojure-lang]
            ["highlight.js/lib/languages/json" :as json-lang]
            ["highlight.js/lib/languages/yaml" :as yaml-lang]))

;; Register languages once
(defonce _register
  (do
    (.registerLanguage hljs "bash" bash)
    (.registerLanguage hljs "clojure" clojure-lang)
    (.registerLanguage hljs "json" json-lang)
    (.registerLanguage hljs "yaml" yaml-lang)
    true))

(defn highlight [code lang]
  (if (and lang (.getLanguage hljs lang))
    (.-value (.highlight hljs code #js {:language lang}))
    code))

(defn code-block
  "Renders a syntax-highlighted code block.
   Usage in hiccup EDN: [:code {:lang \"bash\"} \"echo hello\"]
   Note: dangerouslySetInnerHTML is safe here — content comes from
   Highlight.js processing our own EDN, never user input."
  [{:keys [lang]} code]
  (let [highlighted (highlight code (or lang "plaintext"))]
    [:pre.code
     [:code {:dangerouslySetInnerHTML {:__html highlighted}}]]))

(defn inline-code
  "Renders inline code. Usage: [:code \"some-thing\"]"
  [text]
  [:code text])
