(require '[babashka.process :as bp])
(require '[clojure.string :as str])
(require '[cheshire.core :as json])
(require '[glitch.provider :as provider])

(def ^:private mcp-config
  (json/generate-string
    {"mcpServers"
     {"glitch" {"type" "stdio"
                "command" "glitch"
                "args" ["mcp"]}}}))

(defn- strip-agent-trace
  "Strip copilot agent-mode trace lines from output.
   Trace lines start with: ● ✗ │ └ or are indented continuations.
   Returns only the final response content after the last trace block."
  [s]
  (let [lines (str/split-lines s)
        ;; Find the last non-trace line index working backwards
        trace-prefix? (fn [line]
                        (let [trimmed (str/trim line)]
                          (or (str/starts-with? trimmed "●")
                              (str/starts-with? trimmed "✗")
                              (str/starts-with? trimmed "│")
                              (str/starts-with? trimmed "└")
                              ;; Indented tool call detail lines
                              (re-matches #"\s+[│└●✗].*" line))))]
    (if (some trace-prefix? lines)
      ;; Has agent trace — extract content after the last trace block
      (let [;; Split into segments: trace blocks and content blocks
            result (reduce
                     (fn [acc line]
                       (if (trace-prefix? line)
                         (assoc acc :in-trace true)
                         (if (:in-trace acc)
                           ;; Transitioning out of trace — start new content
                           (-> acc
                               (assoc :in-trace false)
                               (assoc :content [line]))
                           ;; Continuing content
                           (update acc :content (fnil conj []) line))))
                     {:in-trace false :content []}
                     lines)]
        (str/trim (str/join "\n" (:content result))))
      ;; No agent trace — return as-is
      s)))

(provider/register "copilot"
  (fn [{:keys [prompt tool-defs]}]
    (let [args   (if (seq tool-defs)
                   ["copilot" "-p" "-" "--disable-builtin-mcps" "--additional-mcp-config" mcp-config]
                   ["copilot" "-p" "-" "--available-tools="])
          result (apply bp/shell {:out :string :err :string :continue true
                                   :in prompt}
                        args)]
      (when (not= 0 (:exit result))
        (throw (ex-info (str "copilot: failed (exit " (:exit result) "): "
                             (str/trim (:err result)))
                        {})))
      (let [raw (str/trim (:out result))
            ;; Strip agent trace if copilot was in tool-use mode
            clean (if (seq tool-defs)
                    (strip-agent-trace raw)
                    raw)
            ;; Strip trailing stats block — lines starting with Changes/Requests/Tokens
            clean (reduce (fn [s marker]
                            (if-let [idx (str/index-of s (str "\n" marker))]
                              (str/trim (subs s 0 idx))
                              s))
                          clean
                          ["Changes" "Requests" "Tokens"])
            ;; Strip markdown fences if present
            clean (if (str/starts-with? clean "```")
                    (let [first-nl (str/index-of clean "\n")]
                      (if first-nl
                        (let [inner (subs clean (inc first-nl))
                              end   (str/index-of inner "```")]
                          (if end (subs inner 0 end) inner))
                        clean))
                    clean)]
        {:response   (str/trim clean)
         :tokens-in  0
         :tokens-out 0}))))
