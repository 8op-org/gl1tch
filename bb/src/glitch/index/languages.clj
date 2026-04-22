(ns glitch.index.languages
  (:require [clojure.string :as str]))

;; Extension -> language name mapping
(def ext->language
  {".go"   "go",   ".py"   "python",  ".js"  "javascript", ".ts"  "javascript",
   ".jsx"  "javascript", ".tsx" "javascript", ".rs" "rust", ".java" "java",
   ".c"    "c",    ".h"    "c",
   ".clj"  "clojure", ".cljs" "clojure", ".cljc" "clojure", ".bb" "clojure"})

;; Language -> ast-grep language name (used in rule files)
;; Clojure is absent — it uses bb-native extraction, not ast-grep
(def lang->sg-lang
  {"go" "Go", "python" "Python", "javascript" "JavaScript",
   "rust" "Rust", "java" "Java", "c" "C"})

;; Languages that use bb-native extraction instead of ast-grep
(def native-languages #{"clojure"})

(defn detect-language
  "Detect language from file path extension. Returns nil if unknown."
  [path]
  (let [ext (re-find #"\.[^.]+$" path)]
    (get ext->language ext)))

(defn supported-languages []
  (set (vals ext->language)))

(defn native-language?
  "True if this language uses bb-native extraction instead of ast-grep."
  [language]
  (contains? native-languages language))

(defn- homebrew-prefix
  "Return the Homebrew prefix, or nil."
  []
  (let [arm "/opt/homebrew"
        intel "/usr/local"]
    (cond
      (.isDirectory (java.io.File. (str arm "/share/glitch"))) arm
      (.isDirectory (java.io.File. (str intel "/share/glitch"))) intel
      :else nil)))

(defn rules-dir
  "Return path to ast-grep rules for a language category.
   category is one of: symbols, imports, calls, references.
   Resolution order:
   1. GLITCH_RULES_DIR env var
   2. Relative to cwd: resources/ast-grep-rules/
   3. Relative to cwd: bb/resources/ast-grep-rules/
   4. Homebrew share: <prefix>/share/glitch/ast-grep-rules/"
  [language category]
  (let [suffix (str language "/" category ".yml")
        env-dir (System/getenv "GLITCH_RULES_DIR")
        cwd    (System/getProperty "user.dir")
        candidates (cond-> []
                     env-dir (conj (str env-dir "/" suffix))
                     true    (conj (str cwd "/resources/ast-grep-rules/" suffix))
                     true    (conj (str cwd "/bb/resources/ast-grep-rules/" suffix)))]
    (or (some (fn [c]
                (let [f (java.io.File. c)]
                  (when (.exists f) (.getAbsolutePath f))))
              candidates)
        ;; Try Homebrew share path
        (when-let [prefix (homebrew-prefix)]
          (let [f (java.io.File. (str prefix "/share/glitch/ast-grep-rules/" suffix))]
            (when (.exists f) (.getAbsolutePath f))))
        ;; Fallback
        (str "resources/ast-grep-rules/" suffix))))
