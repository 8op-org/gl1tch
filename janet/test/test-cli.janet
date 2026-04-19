(use spork/test)

(array/push module/paths ["src/:all:.janet" :source])

(import glitch/main :as m)

(start-suite "cli")

(assert (= "version" (m/resolve-command @["version"]))
        "resolve version command")
(assert (= "run" (m/resolve-command @["run" "my-wf"]))
        "resolve run command")
(assert (= "workspace" (m/resolve-command @["workspace" "list"]))
        "resolve workspace command")
(assert (nil? (m/resolve-command @[]))
        "empty args returns nil")

(end-suite)
