(ns glitch.mcp.sci-bindings
  "Shared SCI namespace bindings for the glitch DSL."
  (:require [clojure.string :as str]
            [glitch.core :as g]))

(defn user-bindings
  "Base DSL bindings for the 'user namespace in SCI contexts."
  []
  {'trace            g/trace
   'input            g/input
   'params           g/params
   'param            g/param
   'ref              g/ref
   'sh               g/sh
   'search           g/search
   'save             g/save
   'read-file        g/read-file
   'write-file       g/write-file
   'get-steps        g/get-steps
   'last-output      g/last-output
   'gate             g/gate
   'call-workflow    g/call-workflow
   'json-extract     g/json-extract
   'validate-schema  g/validate-schema
   'validate         g/validate
   'llm              g/llm
   'grounded?        g/grounded?
   'consensus        g/consensus
   'composite-score  g/composite-score
   'search-symbols   g/search-symbols
   'search-edges     g/search-edges
   'symbol-context   g/symbol-context})

(def string-bindings
  "clojure.string bindings for SCI contexts."
  {'upper-case   str/upper-case
   'lower-case   str/lower-case
   'trim         str/trim
   'split        str/split
   'join         str/join
   'replace      str/replace
   'starts-with? str/starts-with?
   'ends-with?   str/ends-with?
   'includes?    str/includes?
   'blank?       str/blank?})
