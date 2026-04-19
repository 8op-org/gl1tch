(use spork/test)

# Add src to module path so we can import glitch/provider
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/provider :as p)

(start-suite "provider")

# register and call a mock provider
(p/reset!)
(p/register "mock"
  (fn [opts]
    {:response (string "echo:" (opts :prompt))
     :tokens-in 10 :tokens-out 20 :latency 0 :cost 0}))

(assert (find |(= $ "mock") (p/names))
        "registered provider appears in names")

(def result (p/call-provider "mock" {:prompt "hello" :model "test"}))
(assert (= "echo:hello" (result :response))
        "call-provider dispatches correctly")

# unknown provider raises
(assert-error "unknown provider"
  (p/call-provider "nonexistent" {:prompt "x"}))

# tier escalation
(p/register "fail-provider"
  (fn [opts] (error "provider down")))

(def tiers [{:providers ["fail-provider"]}
            {:providers ["mock"]}])

(def tiered (p/call-tiered {:prompt "test"} tiers))
(assert (= "echo:test" (tiered :response))
        "tier escalation falls through to working provider")

(end-suite)
