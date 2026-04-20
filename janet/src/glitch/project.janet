# Project root discovery.

(defn find-root [dir]
  "Walk up from dir looking for a .glitch/ directory.
   Returns the parent directory containing .glitch/, or nil."
  (var current dir)
  (var result nil)
  (while true
    (def path (string current "/.glitch"))
    (when (os/stat path)
      (set result current)
      (break))
    (def parts (string/find-all "/" current))
    (if (empty? parts)
      (break)
      (do
        (def parent (string/slice current 0 (last parts)))
        (when (or (= parent current) (= parent ""))
          (break))
        (set current parent))))
  result)

(defn resolve [&named flag]
  "Resolve the project root.
   Priority: explicit flag > GLITCH_PROJECT env var > walk from cwd.
   Returns a directory path string or nil."
  (or flag
      (os/getenv "GLITCH_PROJECT")
      (find-root (os/cwd))))
