(ns glitch.index-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.index :as idx]
            [glitch.es :as es]
            [babashka.process :as proc]
            [cheshire.core :as json]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; file-hash
;; ---------------------------------------------------------------------------

(deftest file-hash-consistent-for-same-content
  (testing "file-hash produces consistent SHA256 for same content"
    (let [path "/tmp/glitch-test-hash.txt"]
      (spit path "hello world\n")
      (let [h1 (idx/file-hash path)
            h2 (idx/file-hash path)]
        (is (= h1 h2))
        (is (= 64 (count h1)))  ;; SHA256 hex is 64 chars
        (is (re-matches #"[0-9a-f]{64}" h1)))
      (clojure.java.io/delete-file path true))))

(deftest file-hash-differs-for-different-content
  (testing "file-hash produces different hash for different content"
    (let [p1 "/tmp/glitch-test-hash-a.txt"
          p2 "/tmp/glitch-test-hash-b.txt"]
      (spit p1 "aaa")
      (spit p2 "bbb")
      (is (not= (idx/file-hash p1) (idx/file-hash p2)))
      (clojure.java.io/delete-file p1 true)
      (clojure.java.io/delete-file p2 true))))

;; ---------------------------------------------------------------------------
;; detect-kind (private)
;; ---------------------------------------------------------------------------

(deftest detect-kind-function
  (testing "detect-kind identifies functions"
    (is (= "function" (#'glitch.index/detect-kind "func main() {")))
    (is (= "function" (#'glitch.index/detect-kind "fn process() {")))
    (is (= "function" (#'glitch.index/detect-kind "def handle_request(self):")))
    (is (= "function" (#'glitch.index/detect-kind "function render() {")))))

(deftest detect-kind-method
  (testing "detect-kind: Go method receiver syntax detected as method"
    (is (= "method" (#'glitch.index/detect-kind "func (s *Server) Start() {")))))

(deftest detect-kind-class
  (testing "detect-kind identifies classes"
    (is (= "class" (#'glitch.index/detect-kind "class MyService {")))))

(deftest detect-kind-struct
  (testing "detect-kind identifies structs"
    (is (= "struct" (#'glitch.index/detect-kind "type Config struct {")))
    (is (= "struct" (#'glitch.index/detect-kind "struct Point {")))))

(deftest detect-kind-interface
  (testing "detect-kind identifies interfaces"
    (is (= "interface" (#'glitch.index/detect-kind "type Handler interface {")))))

(deftest detect-kind-import
  (testing "detect-kind identifies imports"
    (is (= "import" (#'glitch.index/detect-kind "import \"fmt\"")))
    (is (= "import" (#'glitch.index/detect-kind "from os import path")))
    (is (= "import" (#'glitch.index/detect-kind "use std::io;")))))

(deftest detect-kind-unknown
  (testing "detect-kind returns unknown for unrecognized patterns"
    (is (= "unknown" (#'glitch.index/detect-kind "// just a comment")))))

;; ---------------------------------------------------------------------------
;; extract-name (private)
;; ---------------------------------------------------------------------------

(deftest extract-name-go-func
  (testing "extract-name from Go function"
    (is (= "main" (#'glitch.index/extract-name "func main() {")))
    (is (= "Start" (#'glitch.index/extract-name "func (s *Server) Start() {")))))

(deftest extract-name-go-type
  (testing "extract-name from Go type declarations"
    (is (= "Config" (#'glitch.index/extract-name "type Config struct {")))
    (is (= "Handler" (#'glitch.index/extract-name "type Handler interface {")))))

(deftest extract-name-rust
  (testing "extract-name from Rust declarations"
    (is (= "process" (#'glitch.index/extract-name "fn process() {")))
    (is (= "handle" (#'glitch.index/extract-name "pub fn handle() {")))))

(deftest extract-name-python
  (testing "extract-name from Python declarations"
    (is (= "handle_request" (#'glitch.index/extract-name "def handle_request(self):")))
    (is (= "MyService" (#'glitch.index/extract-name "class MyService:")))))

(deftest extract-name-javascript
  (testing "extract-name from JavaScript declarations"
    (is (= "render" (#'glitch.index/extract-name "function render() {")))
    (is (= "App" (#'glitch.index/extract-name "class App {")))))

;; ---------------------------------------------------------------------------
;; find-source-files (mock git ls-files)
;; ---------------------------------------------------------------------------

(deftest find-source-files-filters-by-language
  (testing "find-source-files filters to supported language extensions"
    (with-redefs [proc/shell (fn [_opts & _args]
                               {:exit 0
                                :out "main.go\nREADME.md\nlib.py\ndata.csv\napp.js\n"
                                :err ""})]
      (let [files (idx/find-source-files "/repo")]
        (is (= 3 (count files)))
        (is (every? #(or (str/ends-with? % ".go")
                         (str/ends-with? % ".py")
                         (str/ends-with? % ".js"))
                    files))))))

(deftest find-source-files-filters-by-language-option
  (testing "find-source-files respects :languages filter"
    (with-redefs [proc/shell (fn [_opts & _args]
                               {:exit 0
                                :out "main.go\nlib.py\napp.js\n"
                                :err ""})]
      (let [files (idx/find-source-files "/repo" :languages #{"go"})]
        (is (= 1 (count files)))
        (is (str/ends-with? (first files) ".go"))))))

;; ---------------------------------------------------------------------------
;; resolve-imports
;; ---------------------------------------------------------------------------

(deftest resolve-imports-matches-known-symbols
  (testing "resolve-imports creates edges for imports that match symbol table"
    (let [raw-imports [{:name "Config" :file "main.go"}
                       {:name "unknown_mod" :file "main.go"}]
          symbol-table {"Config" ["myrepo:config.go:Config:5"]}
          edges (idx/resolve-imports raw-imports symbol-table "myrepo")]
      (is (= 1 (count edges)))
      (is (= "imports" (:kind (first edges))))
      (is (= "myrepo:config.go:Config:5" (:target (first edges)))))))

;; ---------------------------------------------------------------------------
;; resolve-calls
;; ---------------------------------------------------------------------------

(deftest resolve-calls-matches-known-symbols
  (testing "resolve-calls creates edges for calls that match symbol table"
    (let [raw-calls [{:name "process" :file "main.go" :start_line 10}
                     {:name "missing_fn" :file "main.go" :start_line 15}]
          symbol-table {"process" ["myrepo:handler.go:process:3"]}
          edges (idx/resolve-calls raw-calls symbol-table "myrepo")]
      (is (= 1 (count edges)))
      (is (= "calls" (:kind (first edges))))
      (is (= "myrepo:handler.go:process:3" (:target (first edges))))
      (is (str/includes? (:source (first edges)) "call:process:10")))))

;; ---------------------------------------------------------------------------
;; resolve-references
;; ---------------------------------------------------------------------------

(deftest resolve-references-matches-known-symbols
  (testing "resolve-references creates edges for refs that match symbol table"
    (let [raw-refs [{:name "Config" :file "app.go" :start_line 20}
                    {:name "nope" :file "app.go" :start_line 25}]
          symbol-table {"Config" ["myrepo:config.go:Config:1"]}
          edges (idx/resolve-references raw-refs symbol-table "myrepo")]
      (is (= 1 (count edges)))
      (is (= "references" (:kind (first edges))))
      (is (= "myrepo:config.go:Config:1" (:target (first edges)))))))

;; ---------------------------------------------------------------------------
;; query-symbols (constructs correct ES query)
;; ---------------------------------------------------------------------------

(deftest query-symbols-constructs-correct-query
  (testing "query-symbols sends correct ES query"
    (let [captured (atom nil)]
      (with-redefs [es/search (fn [_url _idx query]
                                (reset! captured query)
                                [])]
        (idx/query-symbols "http://localhost:9200" "myrepo"
                           {:name "Config" :kind "struct" :language "go" :limit 50})
        (let [q @captured
              clauses (get-in q [:query :bool :must])]
          (is (= 50 (:size q)))
          ;; Should have 4 clauses: repo + kind + language + name
          (is (= 4 (count clauses)))
          (is (some #(= {:term {:repo "myrepo"}} %) clauses))
          (is (some #(= {:term {:kind "struct"}} %) clauses))
          (is (some #(= {:term {:language "go"}} %) clauses))
          (is (some #(= {:term {:name "Config"}} %) clauses)))))))

(deftest query-symbols-wildcard-name
  (testing "query-symbols uses wildcard for name with *"
    (let [captured (atom nil)]
      (with-redefs [es/search (fn [_url _idx query]
                                (reset! captured query)
                                [])]
        (idx/query-symbols "http://localhost:9200" "myrepo"
                           {:name "Config*"})
        (let [clauses (get-in @captured [:query :bool :must])]
          (is (some #(= {:wildcard {:name "Config*"}} %) clauses)))))))

;; ---------------------------------------------------------------------------
;; query-edges (single-hop and BFS)
;; ---------------------------------------------------------------------------

(deftest query-edges-single-hop
  (testing "query-edges with depth 1 constructs single-hop query"
    (let [captured (atom nil)]
      (with-redefs [es/search (fn [_url _idx query]
                                (reset! captured query)
                                [])]
        (idx/query-edges "http://localhost:9200" "myrepo"
                         {:source "myrepo:main.go:main:1" :kind "calls" :limit 50})
        (let [q @captured
              clauses (get-in q [:query :bool :must])]
          (is (= 50 (:size q)))
          (is (= 3 (count clauses)))
          (is (some #(= {:term {:repo "myrepo"}} %) clauses))
          (is (some #(= {:term {:source "myrepo:main.go:main:1"}} %) clauses))
          (is (some #(= {:term {:kind "calls"}} %) clauses)))))))

(deftest query-edges-bfs-depth-2
  (testing "query-edges with depth > 1 performs BFS"
    (let [call-count (atom 0)]
      (with-redefs [es/search (fn [_url _idx _query]
                                (swap! call-count inc)
                                (if (= 1 @call-count)
                                  ;; First hop returns one edge
                                  [{:source "a" :target "b" :kind "calls"}]
                                  ;; Second hop returns no new edges
                                  []))]
        (let [results (idx/query-edges "http://localhost:9200" "myrepo"
                                       {:source "a" :depth 2})]
          ;; BFS should have called search at least twice
          (is (>= @call-count 2))
          ;; Should contain the first hop results
          (is (= 1 (count results))))))))

;; ---------------------------------------------------------------------------
;; query-context
;; ---------------------------------------------------------------------------

(deftest query-context-returns-structured-result
  (testing "query-context returns symbol with callers, callees, etc."
    (let [call-count (atom 0)]
      (with-redefs [es/search (fn [_url idx query]
                                (swap! call-count inc)
                                (cond
                                  ;; First call: symbol lookup
                                  (= 1 @call-count)
                                  [{:id "myrepo:main.go:Start:10"
                                    :name "Start"
                                    :kind "function"
                                    :parent_id nil}]
                                  ;; Second call: outgoing edges
                                  (= 2 @call-count)
                                  [{:source "myrepo:main.go:Start:10"
                                    :target "myrepo:handler.go:process:5"
                                    :kind "calls"}
                                   {:source "myrepo:main.go:Start:10"
                                    :target "myrepo:child.go:helper:20"
                                    :kind "contains"}]
                                  ;; Third call: incoming edges
                                  (= 3 @call-count)
                                  [{:source "myrepo:app.go:Run:1"
                                    :target "myrepo:main.go:Start:10"
                                    :kind "calls"}
                                   {:source "myrepo:ref.go:ref:50"
                                    :target "myrepo:main.go:Start:10"
                                    :kind "references"}]
                                  :else []))]
        (let [ctx (idx/query-context "http://localhost:9200" "myrepo" "Start")]
          (is (some? ctx))
          (is (= "Start" (get-in ctx [:symbol :name])))
          ;; callers = incoming calls edges
          (is (= 1 (count (:callers ctx))))
          ;; callees = outgoing calls edges
          (is (= 1 (count (:callees ctx))))
          ;; children = outgoing contains edges
          (is (= 1 (count (:children ctx))))
          ;; references = incoming references edges
          (is (= 1 (count (:references ctx))))
          ;; parent is nil since parent_id is nil
          (is (nil? (:parent ctx))))))))

(deftest query-context-returns-nil-when-not-found
  (testing "query-context returns nil when symbol not found"
    (with-redefs [es/search (fn [_url _idx _query] [])]
      (is (nil? (idx/query-context "http://localhost:9200" "myrepo" "NonExistent"))))))
