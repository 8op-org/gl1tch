# Workspace parsing, resource binding, and discovery.

(var- *current-ws* nil)

(defn workspace [name & body]
  "Define a workspace. Called from workspace.janet files."
  (set *current-ws*
    @{:name name :resources @[] :defaults @{} :path ""})
  (var i 0)
  (while (< i (length body))
    (def item (get body i))
    (cond
      (= item :description) (do (put *current-ws* :description (get body (+ i 1)))
                                (+= i 2))
      (= item :owner) (do (put *current-ws* :owner (get body (+ i 1)))
                          (+= i 2))
      (+= i 1)))
  *current-ws*)

(defn defaults [& kvs]
  "Set workspace defaults."
  (def tbl (table ;kvs))
  (when *current-ws*
    (merge-into (*current-ws* :defaults) tbl))
  tbl)

(defn resource [name & kvs]
  "Add a resource to the current workspace."
  (def res (merge @{:name name} (table ;kvs)))
  (when *current-ws*
    (array/push (*current-ws* :resources) res))
  res)

(defn load [path]
  "Load a workspace.janet file and return the workspace table."
  (set *current-ws* nil)
  (dofile path)
  (when *current-ws*
    (def parts (string/find-all "/" path))
    (when (not (empty? parts))
      (put *current-ws* :path
        (string/slice path 0 (last parts)))))
  *current-ws*)

(defn get-resource [ws name]
  "Find a resource by name."
  (find |(= ($ :name) name) (ws :resources)))

(defn resources-by-type [ws type-str]
  "Filter resources by type."
  (filter |(= ($ :type) type-str) (ws :resources)))

(defn find-workspace-file [dir]
  "Walk up from dir to find .glitch/workspace.janet."
  (var current dir)
  (var result nil)
  (while true
    (def path (string current "/.glitch/workspace.janet"))
    (when (os/stat path)
      (set result path)
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

(defn resolve [&named workspace-flag]
  "Resolve the active workspace.
   Priority: explicit flag > env var > walk cwd > nil."
  (def path
    (or workspace-flag
        (os/getenv "GLITCH_WORKSPACE")
        (find-workspace-file (os/cwd))))
  (when path
    (load path)))
