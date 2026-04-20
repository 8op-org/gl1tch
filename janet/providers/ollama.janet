(import spork/json)

(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "qwen2.5:7b")
  (def body (json/encode {:model model :prompt prompt :stream false}))
  (def proc (os/spawn
    ["curl" "-sS" "-X" "POST"
     "http://localhost:11434/api/generate"
     "-H" "Content-Type: application/json"
     "-d" (string body)]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "ollama: request failed (exit %d): %s" exit (string err-out)))
  (def parsed (json/decode (string out)))
  {:response (or (get parsed "response") "")
   :tokens-in (or (get parsed "prompt_eval_count") 0)
   :tokens-out (or (get parsed "eval_count") 0)
   :latency (or (get parsed "total_duration") 0)
   :cost 0})
