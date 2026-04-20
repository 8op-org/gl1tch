(require '[babashka.process :as bp])
(require '[clojure.string :as str])
(require '[cheshire.core :as json])
(require '[glitch.provider :as provider])

(def ^:private mcp-config
  (json/generate-string
    {"mcpServers"
     {"glitch" {"command" "glitch"
                "args" ["mcp"]}}}))

(provider/register "claude"
  (fn [{:keys [prompt model tool-defs]}]
    (let [use-tools? (seq tool-defs)
          args       (cond-> ["claude" "--print"]
                       use-tools? (into ["--mcp-config" mcp-config])
                       model      (into ["--model" model]))
          result     (apply bp/shell {:out :string :err :string :continue true
                                      :in prompt}
                           args)]
      (when (not= 0 (:exit result))
        (throw (ex-info (str "claude: failed (exit " (:exit result) "): "
                             (str/trim (:err result)))
                        {})))
      {:response   (str/trim (:out result))
       :tokens-in  0
       :tokens-out 0})))
