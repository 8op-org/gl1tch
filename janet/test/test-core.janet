(use spork/test)

# Add src to module path so we can import glitch/core
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/core :as g)

(start-suite "core")

# step stores output and returns it
(g/reset-steps!)
(assert (= "hello" (g/step "greet" "hello"))
        "step returns its body value")
(assert (= "hello" (g/ref "greet"))
        "ref retrieves step output")

# ref for missing step returns nil
(g/reset-steps!)
(assert (nil? (g/ref "missing"))
        "ref returns nil for unknown step")

# sh runs a command and returns stdout
(def result (g/sh "echo" "hi"))
(assert (= "hi\n" result)
        "sh captures stdout")

# sh with failing command raises error
(assert-error "sh fails on bad command"
  (g/sh "false"))

# save writes file, read-file reads it
(def tmp (string "/tmp/glitch-test-" (os/time)))
(g/save tmp "test content")
(assert (= "test content" (g/read-file tmp))
        "save + read-file roundtrip")
(os/rm tmp)

# input returns the current workflow input
(g/set-input! "my input")
(assert (= "my input" (g/input))
        "input returns current input")

# params
(g/set-params! @{"key" "val"})
(assert (= "val" (g/param "key"))
        "param returns value")
(assert (nil? (g/param "missing"))
        "param returns nil for missing key")
(assert (= "default" (g/param "missing" "default"))
        "param returns default")

# step recorder callback
(var recorded nil)
(g/set-step-recorder! (fn [rec] (set recorded rec)))
(g/reset-steps!)
(g/step "test-step" "output")
(assert (= "test-step" (recorded :step-id))
        "step recorder called with step-id")
(assert (= "output" (recorded :output))
        "step recorder called with output")
(g/set-step-recorder! nil)

# get-steps returns snapshot
(g/reset-steps!)
(g/step "a" "1")
(g/step "b" "2")
(def snap (g/get-steps))
(assert (= "1" (snap "a")) "get-steps includes a")
(assert (= "2" (snap "b")) "get-steps includes b")

(end-suite)
