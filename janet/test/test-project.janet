(use spork/test)

# Add src to module path so we can import glitch/project
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/project :as proj)

(start-suite "project")

(def tmp-dir (string "/tmp/glitch-project-test-" (os/time)))

# --- find-root discovers .glitch/ directory ---
(os/mkdir tmp-dir)
(os/mkdir (string tmp-dir "/.glitch"))
(def root (proj/find-root tmp-dir))
(assert (= root tmp-dir) "find-root discovers .glitch/ directory")

# --- find-root walks up directories ---
(os/mkdir (string tmp-dir "/sub"))
(os/mkdir (string tmp-dir "/sub/deep"))
(def walked (proj/find-root (string tmp-dir "/sub/deep")))
(assert (= walked tmp-dir) "find-root walks up directories")

# --- find-root returns nil when no .glitch/ exists ---
(def missing (proj/find-root "/tmp"))
(assert (nil? missing) "find-root returns nil when no .glitch/ exists")

# --- resolve with explicit flag ---
(def resolved (proj/resolve :flag tmp-dir))
(assert (= resolved tmp-dir) "resolve with explicit :flag returns that path")

# cleanup
(os/shell (string "rm -rf " tmp-dir))

(end-suite)
