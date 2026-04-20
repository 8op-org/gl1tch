(ns glitch.mcp.indexer-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.mcp.indexer :as idx]
            [babashka.fs :as fs]))

(deftest detect-language-test
  (is (= "go" (idx/detect-language "main.go")))
  (is (= "python" (idx/detect-language "script.py")))
  (is (= "typescript" (idx/detect-language "app.tsx")))
  (is (= "clojure" (idx/detect-language "core.clj")))
  (is (nil? (idx/detect-language "README")))
  (is (nil? (idx/detect-language nil))))

(deftest extract-symbols-test
  (testing "go functions"
    (let [syms (idx/extract-symbols "func main() {}\nfunc (s *Server) Start() {}" "go")]
      (is (some #{"main"} syms))
      (is (some #{"Start"} syms))))
  (testing "python"
    (let [syms (idx/extract-symbols "def hello():\n  pass\nclass Foo:" "python")]
      (is (some #{"hello"} syms))
      (is (some #{"Foo"} syms))))
  (testing "clojure"
    (let [syms (idx/extract-symbols "(defn greet [x] x)\n(def pi 3.14)" "clojure")]
      (is (some #{"greet"} syms))
      (is (some #{"pi"} syms))))
  (testing "unknown language returns empty"
    (is (empty? (idx/extract-symbols "foo" "unknown"))))
  (testing "nil returns empty"
    (is (empty? (idx/extract-symbols nil nil)))))

(deftest chunk-text-test
  (testing "short text returns single chunk"
    (is (= ["hello"] (idx/chunk-text "hello"))))
  (testing "nil/empty returns empty"
    (is (empty? (idx/chunk-text nil)))
    (is (empty? (idx/chunk-text ""))))
  (testing "long text splits with overlap"
    (let [text (apply str (repeat 200 "abcdefghij"))
          chunks (idx/chunk-text text :chunk-size 1500 :overlap 150)]
      (is (> (count chunks) 1))
      (let [end-of-first (subs (first chunks) (- (count (first chunks)) 150))
            start-of-second (subs (second chunks) 0 150)]
        (is (= end-of-first start-of-second))))))

(deftest skip?-test
  (is (true? (idx/skip? ".git")))
  (is (true? (idx/skip? "node_modules")))
  (is (false? (idx/skip? "src"))))

(deftest open-and-index-test
  (let [tmp-dir (str (fs/create-temp-dir {:prefix "idx-test"}))]
    (try
      (fs/create-dirs (str tmp-dir "/src"))
      (spit (str tmp-dir "/src/main.go") "package main\n\nfunc hello() {}\n")
      (let [db (idx/open-search-db tmp-dir)
            result (idx/index-repo db tmp-dir)]
        (is (= 1 (:files-indexed result)))
        (is (pos? (:chunks-created result)))
        (idx/close-search-db db))
      (finally
        (fs/delete-tree tmp-dir)))))
