(import spork/json)

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "google/gemma-3-4b-it:free")
  (def api-key (os/getenv "OPENROUTER_API_KEY"))
  (unless api-key
    (error "openrouter: OPENROUTER_API_KEY not set"))
  (def base-url (or (os/getenv "OPENROUTER_BASE_URL")
                    "https://openrouter.ai/api/v1"))
  (def body (json/encode
    {:model model
     :messages [{:role "user" :content prompt}]
     :stream false}))
  (def proc (os/spawn
    ["curl" "-sS" "-X" "POST"
     (string base-url "/chat/completions")
     "-H" "Content-Type: application/json"
     "-H" (string "Authorization: Bearer " api-key)
     "-H" "HTTP-Referer: https://gl1tch.dev"
     "-H" "X-Title: glitch"
     "-D" "/dev/stderr"
     "-d" (string body)]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "openrouter: request failed (exit %d): %s" exit (string err-out)))
  (def parsed (json/decode (string out)))
  (when (get parsed "error")
    (errorf "openrouter: %s" (get-in parsed ["error" "message"] "unknown error")))
  (def choice (get-in parsed ["choices" 0 "message" "content"] ""))
  (def usage (get parsed "usage" {}))
  # Extract cost from response headers dumped to stderr
  (var cost 0)
  (when-let [idx (string/find "x-openrouter-cost:" (string/ascii-lower (string err-out)))]
    (def after (string/slice (string err-out) (+ idx 19)))
    (when-let [nl (string/find "\r" after)]
      (def cost-str (string/trim (string/slice after 0 nl)))
      (try (set cost (scan-number cost-str)) ([_]))))
  {:response choice
   :tokens-in (or (get usage "prompt_tokens") 0)
   :tokens-out (or (get usage "completion_tokens") 0)
   :latency 0 :cost cost})
