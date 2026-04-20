(ns glitch.provider
  (:require [clojure.java.io :as io]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Provider registry — each entry is {name -> (fn [opts] => {:response :tokens-in :tokens-out})}
;; ---------------------------------------------------------------------------

(def registry (atom {}))

(defn reset! []
  "Clear all registered providers."
  (clojure.core/reset! registry {}))

(defn register [name provider-fn]
  "Register a provider function under `name`."
  (swap! registry assoc name provider-fn))

(defn names []
  "Return a sorted seq of registered provider names."
  (sort (keys @registry)))

(defn call-provider [name opts]
  "Look up provider `name` in the registry and call it with `opts`.
   Throws if the provider is not registered."
  (let [provider-fn (get @registry name)]
    (when-not provider-fn
      (throw (ex-info (str "unknown provider: " name) {:provider name})))
    (provider-fn opts)))

(defn call-tiered
  "Try each tier in order. A tier is {:providers [\"name\" ...] :model \"optional\"}.
   For every provider in the tier, attempt call-provider with merged opts.
   Returns the first non-nil, non-empty response.
   Throws when all tiers are exhausted."
  [opts tiers]
  (loop [remaining tiers
         last-err nil]
    (if (empty? remaining)
      (throw (ex-info (str "all tiers exhausted: " (or last-err "no providers tried"))
                      {:last-error last-err}))
      (let [tier          (first remaining)
            model-override (:model tier)
            providers      (:providers tier)]
        (let [result (reduce
                       (fn [_ pname]
                         (try
                           (let [merged (cond-> opts
                                          model-override (assoc :model model-override))
                                 res    (call-provider pname merged)]
                             (if (or (nil? (:response res))
                                     (str/blank? (:response res)))
                               nil                        ;; keep trying
                               (reduced res)))            ;; success — short-circuit
                           (catch Exception e
                             (binding [*out* *err*]
                               (println (str "tier: " pname " failed: " (.getMessage e))))
                             nil)))
                       nil
                       providers)]
          (if result
            result
            (recur (rest remaining) (str "tier " (:providers (first remaining)) " failed"))))))))

(def default-tiers
  [{:providers ["copilot"]}
   {:providers ["claude"]}
   {:providers ["openrouter"]}
   {:providers ["lmstudio"]}])

(defn load-providers
  "Load .clj provider files from directories. Each file should call
   (glitch.provider/register name fn) at the top level.
   Default search dirs: ~/.config/glitch/providers/, providers/ (relative),
   plus any explicitly passed dirs."
  [& dirs]
  (let [home        (System/getProperty "user.home")
        search-dirs (concat
                      [(str home "/.config/glitch/providers")
                       "providers"]
                      dirs)]
    (doseq [dir search-dirs]
      (let [d (io/file dir)]
        (when (.isDirectory d)
          (doseq [f (sort (.list d))]
            (when (str/ends-with? f ".clj")
              (try
                (load-file (str (.getPath d) "/" f))
                (catch Exception e
                  (binding [*out* *err*]
                    (println (str "warn: failed to load provider "
                                  (subs f 0 (- (count f) 4))
                                  ": " (.getMessage e)))))))))))))
