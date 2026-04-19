# Provider registry and tier escalation.

(var- registry @{})

(defn reset! []
  (set registry @{}))

(defn register [name provider-fn]
  "Register a provider function."
  (put registry name provider-fn))

(defn names []
  "Return sorted list of registered provider names."
  (sorted (keys registry)))

(defn call-provider [name opts]
  "Call a registered provider by name."
  (def provider (get registry name))
  (unless provider
    (errorf "unknown provider: %s" name))
  (provider opts))

(defn call-tiered [opts tiers]
  "Try providers in tier order. Returns first success."
  (var last-err nil)
  (var success nil)
  (each tier tiers
    (when (not success)
      (def model-override (get tier :model))
      (each pname (tier :providers)
        (when (not success)
          (def merged (merge opts
                        (if model-override {:model model-override} {})))
          (try
            (do
              (def result (call-provider pname merged))
              (when (or (nil? (result :response))
                        (= "" (string/trim (or (result :response) ""))))
                (error "empty response"))
              (set success result))
            ([err] (set last-err err)
                   (eprintf "tier: %s failed: %s" pname (string err))))))))
  (if success
    success
    (errorf "all tiers exhausted: %s" (string last-err))))

(def default-tiers
  [{:providers ["ollama"] :model "qwen2.5:7b"}
   {:providers ["codex" "gemini"]}
   {:providers ["copilot" "claude"]}])

(defn load-providers [& dirs]
  "Load .janet provider files from directories."
  (def search-dirs
    (if (> (length dirs) 0)
      dirs
      [(string (os/getenv "HOME") "/.config/glitch/providers")
       "providers"]))
  (each dir search-dirs
    (when (os/stat dir)
      (each f (os/dir dir)
        (when (string/has-suffix? ".janet" f)
          (def name (string/slice f 0 (- (length f) 6)))
          (try
            (do
              (def mod (dofile (string dir "/" f)))
              (when (mod :call)
                (register name (mod :call))))
            ([err]
              (eprintf "warn: failed to load provider %s: %s" name (string err)))))))))
