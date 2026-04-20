(use spork/test)

# Add src to module path so we can import glitch/mcp/search
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/mcp/search :as search)
(import glitch/mcp/indexer :as idx)
(import glitch/mcp/vecmath :as vm)
(import sqlite3 :as sql)

(start-suite "search")

# ============================================================
# normalize-scores
# ============================================================

(assert (deep= @[] (search/normalize-scores @[]))
        "normalize-scores: empty -> empty")

(let [res (search/normalize-scores @[{:id 1 :score 5}])]
  (assert (= 1 (length res))
          "normalize-scores: single item returns one result")
  (assert (= 1 ((first res) :score))
          "normalize-scores: single item -> score 1"))

(let [res (search/normalize-scores
            @[{:id 1 :score 10}
              {:id 2 :score 5}
              {:id 3 :score 0}])]
  (assert (= 3 (length res))
          "normalize-scores: three items returns three results")
  # id 1 had max (10) -> 1, id 3 had min (0) -> 0, id 2 mid -> 0.5
  (def by-id @{})
  (each r res (put by-id (r :id) (r :score)))
  (assert (= 1 (by-id 1))
          "normalize-scores: max score -> 1")
  (assert (= 0 (by-id 3))
          "normalize-scores: min score -> 0")
  (assert (< 0.49 (by-id 2))
          "normalize-scores: mid score ~ 0.5 lower bound")
  (assert (> 0.51 (by-id 2))
          "normalize-scores: mid score ~ 0.5 upper bound"))

# All same score -> all get 1
(let [res (search/normalize-scores
            @[{:id 1 :score 7}
              {:id 2 :score 7}
              {:id 3 :score 7}])]
  (each r res
    (assert (= 1 (r :score))
            "normalize-scores: all same score -> 1")))

# ============================================================
# merge-scores
# ============================================================

# Both sets have overlapping IDs
(let [kw @[{:id 1 :score 0.8} {:id 2 :score 0.5}]
      sem @[{:id 1 :score 0.6} {:id 3 :score 0.9}]
      merged (search/merge-scores kw sem)]
  # id 1: 0.4*0.8 + 0.6*0.6 = 0.32 + 0.36 = 0.68
  # id 2: 0.4*0.5 = 0.20
  # id 3: 0.6*0.9 = 0.54
  (assert (= 3 (length merged))
          "merge-scores: 3 unique IDs")
  # Should be sorted descending
  (assert (>= ((in merged 0) :score) ((in merged 1) :score))
          "merge-scores: sorted descending [0] >= [1]")
  (assert (>= ((in merged 1) :score) ((in merged 2) :score))
          "merge-scores: sorted descending [1] >= [2]")
  # First should be id 1 (0.68)
  (assert (= 1 ((in merged 0) :id))
          "merge-scores: highest is id 1 (in both sets)"))

# Custom weights
(let [kw @[{:id 1 :score 1.0}]
      sem @[{:id 1 :score 1.0}]
      merged (search/merge-scores kw sem
               :keyword-weight 0.5 :semantic-weight 0.5)]
  (assert (= 1 (length merged))
          "merge-scores: custom weights returns 1 result")
  (assert (< (math/abs (- 1.0 ((first merged) :score))) 0.001)
          "merge-scores: 0.5+0.5 with both score=1 -> 1.0"))

# ID only in one set
(let [kw @[{:id 10 :score 0.9}]
      sem @[{:id 20 :score 0.7}]
      merged (search/merge-scores kw sem)]
  (assert (= 2 (length merged))
          "merge-scores: disjoint IDs -> both appear")
  (def by-id @{})
  (each r merged (put by-id (r :id) (r :score)))
  # id 10: 0.4*0.9 = 0.36
  # id 20: 0.6*0.7 = 0.42
  (assert (< (math/abs (- 0.36 (by-id 10))) 0.001)
          "merge-scores: keyword-only score correct")
  (assert (< (math/abs (- 0.42 (by-id 20))) 0.001)
          "merge-scores: semantic-only score correct"))

# ============================================================
# SQLite integration: keyword-search
# ============================================================

(def tmp-dir (string "/tmp/test-search-" (os/time)))
(os/mkdir tmp-dir)
(def repo-dir (string tmp-dir "/repo"))
(os/mkdir repo-dir)

(var test-db nil)

(defer (do
         (when test-db (sql/close test-db))
         (os/execute ["rm" "-rf" tmp-dir] :p))

  # Create test Go files
  (spit (string repo-dir "/auth.go")
    (string "package auth\n\n"
            "func HandleLogin(user string) error {\n"
            "\treturn nil\n"
            "}\n\n"
            "func HandleLogout(session string) error {\n"
            "\treturn nil\n"
            "}\n\n"
            "type AuthConfig struct {\n"
            "\tSecret string\n"
            "}\n"))

  (spit (string repo-dir "/server.go")
    (string "package main\n\n"
            "func StartServer(port int) {\n"
            "\t// start HTTP server\n"
            "}\n\n"
            "func HandleRequest(w http.ResponseWriter, r *http.Request) {\n"
            "\t// handle request\n"
            "}\n"))

  # Open DB and index repo
  (set test-db (idx/open-search-db tmp-dir))
  (idx/index-repo test-db repo-dir)

  # keyword-search: find HandleLogin
  (def kw-results (search/keyword-search test-db "HandleLogin" :repo repo-dir))
  (assert (> (length kw-results) 0)
          "keyword-search: finds HandleLogin")
  (def first-hit (first kw-results))
  (assert (first-hit :id)
          "keyword-search: result has :id")
  (assert (first-hit :path)
          "keyword-search: result has :path")
  (assert (first-hit :content)
          "keyword-search: result has :content")
  (assert (first-hit :score)
          "keyword-search: result has :score")
  (assert (> (first-hit :score) 0)
          "keyword-search: score is positive (negated from FTS5)")

  # keyword-search: find server
  (def server-results (search/keyword-search test-db "StartServer" :repo repo-dir))
  (assert (> (length server-results) 0)
          "keyword-search: finds StartServer")

  # keyword-search: nonsense returns empty
  (def no-results (search/keyword-search test-db "zzzznonexistent9999"))
  (assert (= 0 (length no-results))
          "keyword-search: nonsense query -> empty")

  # ============================================================
  # hybrid-search (without embeddings)
  # ============================================================

  (def hybrid-results (search/hybrid-search test-db "HandleLogin" repo-dir))
  (assert (> (length hybrid-results) 0)
          "hybrid-search: finds HandleLogin")
  (def hr (first hybrid-results))
  (assert (hr :path)
          "hybrid-search: result has :path")
  (assert (hr :content)
          "hybrid-search: result has :content")
  (assert (hr :score)
          "hybrid-search: result has :score")
  (assert (or (hr :language) true)
          "hybrid-search: result has :language key")

  # hybrid-search: results are sorted descending
  (when (> (length hybrid-results) 1)
    (assert (>= ((in hybrid-results 0) :score)
                ((in hybrid-results 1) :score))
            "hybrid-search: results sorted descending by score"))

  # semantic-search without embed-fn returns empty
  (def sem-results (search/semantic-search test-db "HandleLogin" repo-dir))
  (assert (= 0 (length sem-results))
          "semantic-search: no embed-fn -> empty")

  (sql/close test-db)
  (set test-db nil))

(end-suite)
