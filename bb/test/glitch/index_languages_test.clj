(ns glitch.index-languages-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.index.languages :as lang]))

;; ---------------------------------------------------------------------------
;; detect-language
;; ---------------------------------------------------------------------------

(deftest detect-language-go
  (testing ".go files detected as go"
    (is (= "go" (lang/detect-language "main.go")))
    (is (= "go" (lang/detect-language "internal/server/handler.go")))))

(deftest detect-language-python
  (testing ".py files detected as python"
    (is (= "python" (lang/detect-language "app.py")))
    (is (= "python" (lang/detect-language "src/models/user.py")))))

(deftest detect-language-javascript-variants
  (testing ".js, .ts, .jsx, .tsx all detected as javascript"
    (is (= "javascript" (lang/detect-language "index.js")))
    (is (= "javascript" (lang/detect-language "app.ts")))
    (is (= "javascript" (lang/detect-language "Component.jsx")))
    (is (= "javascript" (lang/detect-language "Component.tsx")))))

(deftest detect-language-rust
  (testing ".rs files detected as rust"
    (is (= "rust" (lang/detect-language "lib.rs")))
    (is (= "rust" (lang/detect-language "src/main.rs")))))

(deftest detect-language-java
  (testing ".java files detected as java"
    (is (= "java" (lang/detect-language "Main.java")))
    (is (= "java" (lang/detect-language "com/example/Service.java")))))

(deftest detect-language-c
  (testing ".c and .h files detected as c"
    (is (= "c" (lang/detect-language "main.c")))
    (is (= "c" (lang/detect-language "include/header.h")))))

(deftest detect-language-unsupported
  (testing "unsupported extensions return nil"
    (is (nil? (lang/detect-language "readme.txt")))
    (is (nil? (lang/detect-language "data.csv")))
    (is (nil? (lang/detect-language "Makefile")))))

;; ---------------------------------------------------------------------------
;; rules-dir
;; ---------------------------------------------------------------------------

(deftest rules-dir-returns-correct-path-format
  (testing "rules-dir returns a path containing the language and category"
    (let [path (lang/rules-dir "go" "symbols")]
      (is (string? path))
      (is (clojure.string/includes? path "go"))
      (is (clojure.string/includes? path "symbols")))))

;; ---------------------------------------------------------------------------
;; supported-languages
;; ---------------------------------------------------------------------------

(deftest supported-languages-returns-set
  (testing "supported-languages returns set of all language names"
    (let [langs (lang/supported-languages)]
      (is (set? langs))
      (is (contains? langs "go"))
      (is (contains? langs "python"))
      (is (contains? langs "javascript"))
      (is (contains? langs "rust"))
      (is (contains? langs "java"))
      (is (contains? langs "c")))))
