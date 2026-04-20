(require '[babashka.process :as bp])
(require '[clojure.string :as str])
(require '[glitch.provider :as provider])

(provider/register "claude"
  (fn [{:keys [prompt model]}]
    (let [args   (cond-> ["claude" "--print" prompt]
                   model (into ["--model" model]))
          result (apply bp/shell {:out :string :err :inherit :continue true} args)]
      (when (not= 0 (:exit result))
        (throw (ex-info (str "claude: failed (exit " (:exit result) ")") {})))
      {:response   (str/trim (:out result))
       :tokens-in  0
       :tokens-out 0})))
