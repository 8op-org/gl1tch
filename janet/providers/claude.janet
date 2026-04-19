(defn call [opts]
  (def {:model model :prompt prompt} opts)
  (default model "sonnet")
  (def proc (os/spawn
    ["claude" "--print" "--model" model "-" prompt]
    :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (os/proc-wait proc)
  {:response (string/trim (string out))
   :tokens-in 0 :tokens-out 0 :latency 0 :cost 0})
