(use spork/test)

# Add stdlib to module path
(array/push module/paths ["./:all:.janet" :source])

(import stdlib/collections :as c)
(import stdlib/strings :as str)
(import stdlib/io :as gio)

(start-suite "stdlib")

# collections
(assert (deep= @["a" "b"] (c/compact @["a" "" "b" ""]))
        "compact removes empties")
(assert (= "first" (c/first @["first" "second"]))
        "first returns index 0")
(assert (deep= @["x" "y"]
          (c/pluck :name @[{:name "x"} {:name "y"}]))
        "pluck extracts field")
(assert (deep= @[1 2] (c/take 2 @[1 2 3 4]))
        "take returns first n")
(assert (deep= @[1 2 3] (c/unique @[1 2 2 3 1]))
        "unique deduplicates")
(assert (deep= @[1 3] (c/without @[1 2 3] @[2]))
        "without excludes items")

# strings
(assert (= "hello-world" (str/kebab-case "hello world"))
        "kebab-case")
(assert (= "hello_world" (str/snake-case "hello world"))
        "snake-case")
(assert (str/blank? "") "blank? for empty")
(assert (str/blank? "  ") "blank? for whitespace")
(assert (not (str/blank? "hi")) "blank? for non-empty")
(assert (str/present? "hi") "present? for non-empty")
(assert (not (str/present? "")) "present? for empty")
(assert (deep= @["hello" "world"] (str/words "hello world"))
        "words splits on whitespace")
(assert (= "hello world" (str/unwords @["hello" "world"]))
        "unwords joins with space")

# io
(def tmp (string "/tmp/glitch-stdlib-test-" (os/time)))
(gio/write-lines tmp @["one" "two" "three"])
(assert (deep= @["one" "two" "three"] (gio/read-lines tmp))
        "write-lines + read-lines roundtrip")
(os/rm tmp)

(end-suite)
