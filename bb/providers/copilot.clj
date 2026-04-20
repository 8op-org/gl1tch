(require '[babashka.process :as bp])
(require '[clojure.java.io :as io])
(require '[clojure.string :as str])
(require '[glitch.provider :as provider])

(provider/register "copilot"
  (fn [{:keys [prompt]}]
    (let [tmp (str (System/getProperty "java.io.tmpdir")
                    "/glitch-copilot-" (System/currentTimeMillis) "-" (rand-int 100000) ".txt")]
      (spit tmp prompt)
      (.setReadable (java.io.File. tmp) false true)
      (try
        (let [result (bp/shell {:out :string :err :string :continue true}
                       "bash" "-c" (str "cat '" tmp "' | gh copilot 2>&1"))
              raw    (str/trim (:out result))]
          (when (not= 0 (:exit result))
            (throw (ex-info (str "copilot: failed (exit " (:exit result) "): " raw) {})))
          ;; Strip trailing stats block (Changes/Requests/Tokens lines)
          (let [clean (if-let [idx (str/index-of raw "\nChanges")]
                        (str/trim (subs raw 0 idx))
                        raw)
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
             :tokens-out 0}))
        (finally
          (io/delete-file tmp true))))))
