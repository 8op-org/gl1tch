(ns glitch.project
  "Project root resolution — walks up from cwd looking for .glitch/ directory."
  (:require [clojure.java.io :as io]))

(defn resolve
  "Find the project root by looking for a .glitch/ directory.
   Checks: explicit :flag value, then current dir, then parent dirs up to root.
   Returns the absolute path string or nil."
  [& {:keys [flag]}]
  (or flag
      (loop [dir (io/file (System/getProperty "user.dir"))]
        (when dir
          (if (.exists (io/file dir ".glitch"))
            (.getPath dir)
            (recur (.getParentFile dir)))))))
