(import spork/json)

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "default")
  (def body (json/encode
    {:model model
     :messages [{:role "user" :content prompt}]
     :stream false}))
  (def proc (os/spawn
    ["curl" "-sS" "-X" "POST"
     "http://localhost:1234/v1/chat/completions"
     "-H" "Content-Type: application/json"
     "-d" body]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "lmstudio: request failed (exit %d): %s" exit (string err-out)))
  (def parsed (json/decode (string out)))
  (def choice (get-in parsed ["choices" 0 "message" "content"] ""))
  (def usage (get parsed "usage" {}))
  {:response choice
   :tokens-in (or (get usage "prompt_tokens") 0)
   :tokens-out (or (get usage "completion_tokens") 0)
   :latency 0 :cost 0})
