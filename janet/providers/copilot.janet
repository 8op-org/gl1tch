(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "gpt-4o")
  (def tmp (string "/tmp/glitch-copilot-" (os/time) ".txt"))
  (spit tmp prompt)
  (defer (os/rm tmp)
    (def proc (os/spawn
      ["gh" "copilot" "suggest" "-t" "shell" (string (slurp tmp))]
      :p {:out :pipe :err :pipe}))
    (def out (ev/read (proc :out) :all))
    (os/proc-wait proc)
    {:response (string/trim (string out))
     :tokens-in 0 :tokens-out 0 :latency 0 :cost 0}))
