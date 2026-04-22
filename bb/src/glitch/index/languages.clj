(ns glitch.index.languages)

;; Extension -> language name mapping
(def ext->language
  {".go"  "go", ".py" "python", ".js" "javascript", ".ts" "javascript",
   ".jsx" "javascript", ".tsx" "javascript", ".rs" "rust", ".java" "java",
   ".c" "c", ".h" "c"})

;; Language -> ast-grep language name (used in rule files)
(def lang->sg-lang
  {"go" "Go", "python" "Python", "javascript" "JavaScript",
   "rust" "Rust", "java" "Java", "c" "C"})

(defn detect-language
  "Detect language from file path extension. Returns nil if unknown."
  [path]
  (let [ext (re-find #"\.[^.]+$" path)]
    (get ext->language ext)))

(defn supported-languages []
  (set (vals ext->language)))

(defn rules-dir
  "Return path to ast-grep rules for a language category.
   category is one of: symbols, imports, calls, references.
   Tries multiple resolution strategies: absolute resource path,
   relative to cwd, and bb/resources/ prefix."
  [language category]
  (let [rel   (str "resources/ast-grep-rules/" language "/" category ".yml")
        bb    (str "bb/" rel)
        candidates [rel bb]
        cwd   (System/getProperty "user.dir")]
    (or (some (fn [c]
                (let [f (java.io.File. c)]
                  (when (.exists f) (.getAbsolutePath f))))
              candidates)
        (some (fn [c]
                (let [f (java.io.File. cwd c)]
                  (when (.exists f) (.getAbsolutePath f))))
              candidates)
        ;; Fallback: return the relative path and hope for the best
        rel)))
