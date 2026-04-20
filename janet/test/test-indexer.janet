(use spork/test)

# Add src to module path so we can import glitch/mcp/indexer
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/mcp/indexer :as idx)
(import sqlite3 :as sql)

(start-suite "indexer")

# ============================================================
# detect-language
# ============================================================

(assert (= "go" (idx/detect-language "main.go"))
        "detect-language: .go -> go")
(assert (= "janet" (idx/detect-language "foo.janet"))
        "detect-language: .janet -> janet")
(assert (= "python" (idx/detect-language "script.py"))
        "detect-language: .py -> python")
(assert (= "javascript" (idx/detect-language "app.js"))
        "detect-language: .js -> javascript")
(assert (= "markdown" (idx/detect-language "README.md"))
        "detect-language: .md -> markdown")
(assert (= "typescript" (idx/detect-language "index.ts"))
        "detect-language: .ts -> typescript")
(assert (= "typescript" (idx/detect-language "comp.tsx"))
        "detect-language: .tsx -> typescript")
(assert (= "javascript" (idx/detect-language "comp.jsx"))
        "detect-language: .jsx -> javascript")
(assert (= "rust" (idx/detect-language "lib.rs"))
        "detect-language: .rs -> rust")
(assert (= "yaml" (idx/detect-language "config.yaml"))
        "detect-language: .yaml -> yaml")
(assert (= "yaml" (idx/detect-language "config.yml"))
        "detect-language: .yml -> yaml")
(assert (= "shell" (idx/detect-language "run.sh"))
        "detect-language: .sh -> shell")
(assert (nil? (idx/detect-language "Makefile"))
        "detect-language: unknown -> nil")
(assert (nil? (idx/detect-language "image.png"))
        "detect-language: .png -> nil")

# ============================================================
# extract-symbols
# ============================================================

# Go
(let [go-src `
func Hello() {}
func (r *Receiver) Method() {}
type MyStruct struct {}
type MyInterface interface {}
func standalone(a int) string {}
`
      syms (idx/extract-symbols go-src "go")]
  (assert (> (length syms) 0) "extract-symbols: Go finds symbols")
  (assert (find |(= $ "Hello") syms)
          "extract-symbols: Go finds func Hello")
  (assert (find |(= $ "Method") syms)
          "extract-symbols: Go finds method receiver")
  (assert (find |(= $ "MyStruct") syms)
          "extract-symbols: Go finds type MyStruct")
  (assert (find |(= $ "MyInterface") syms)
          "extract-symbols: Go finds type MyInterface")
  (assert (find |(= $ "standalone") syms)
          "extract-symbols: Go finds func standalone"))

# Janet
(let [janet-src `
(defn my-func [x] x)
(defn- private-func [] nil)
(def my-var 42)
(defmacro my-macro [x] x)
`
      syms (idx/extract-symbols janet-src "janet")]
  (assert (find |(= $ "my-func") syms)
          "extract-symbols: Janet finds defn")
  (assert (find |(= $ "private-func") syms)
          "extract-symbols: Janet finds defn-")
  (assert (find |(= $ "my-var") syms)
          "extract-symbols: Janet finds def")
  (assert (find |(= $ "my-macro") syms)
          "extract-symbols: Janet finds defmacro"))

# Python
(let [py-src `
def hello_world():
    pass

class MyClass:
    def method(self):
        pass
`
      syms (idx/extract-symbols py-src "python")]
  (assert (find |(= $ "hello_world") syms)
          "extract-symbols: Python finds def")
  (assert (find |(= $ "MyClass") syms)
          "extract-symbols: Python finds class")
  (assert (find |(= $ "method") syms)
          "extract-symbols: Python finds method"))

# Unknown language
(assert (deep= @[] (idx/extract-symbols "some content" nil))
        "extract-symbols: nil language -> empty")
(assert (deep= @[] (idx/extract-symbols "some content" "unknown"))
        "extract-symbols: unknown language -> empty")

# JavaScript/TypeScript
(let [js-src `
function greet(name) {}
const API_KEY = "abc";
class Widget {}
export function render() {}
`
      syms (idx/extract-symbols js-src "javascript")]
  (assert (find |(= $ "greet") syms)
          "extract-symbols: JS finds function")
  (assert (find |(= $ "API_KEY") syms)
          "extract-symbols: JS finds const")
  (assert (find |(= $ "Widget") syms)
          "extract-symbols: JS finds class")
  (assert (find |(= $ "render") syms)
          "extract-symbols: JS finds export function"))

# Rust
(let [rs-src `
fn process() {}
struct Config {}
enum Color {}
trait Drawable {}
`
      syms (idx/extract-symbols rs-src "rust")]
  (assert (find |(= $ "process") syms)
          "extract-symbols: Rust finds fn")
  (assert (find |(= $ "Config") syms)
          "extract-symbols: Rust finds struct")
  (assert (find |(= $ "Color") syms)
          "extract-symbols: Rust finds enum")
  (assert (find |(= $ "Drawable") syms)
          "extract-symbols: Rust finds trait"))

# ============================================================
# chunk-text
# ============================================================

(assert (deep= @[] (idx/chunk-text ""))
        "chunk-text: empty -> empty array")
(assert (deep= @[] (idx/chunk-text nil))
        "chunk-text: nil -> empty array")

(let [small "hello world"]
  (assert (deep= @[small] (idx/chunk-text small))
          "chunk-text: small text -> single chunk"))

# Large text: generate text bigger than default chunk-size
(let [line "abcdefghij "  # 11 chars
      text (string/repeat line 200) # 2200 chars > 1500
      chunks (idx/chunk-text text)]
  (assert (> (length chunks) 1)
          "chunk-text: large text produces multiple chunks")
  (assert (<= (length (in chunks 0)) 1500)
          "chunk-text: first chunk <= chunk-size")
  # Verify overlap: end of chunk 0 should overlap start of chunk 1
  (def c0 (in chunks 0))
  (def c1 (in chunks 1))
  (def overlap-region (string/slice c0 (- (length c0) 150)))
  (assert (string/has-prefix? overlap-region c1)
          "chunk-text: chunks overlap correctly"))

# Custom chunk-size and overlap
(let [text (string/repeat "x" 100)
      chunks (idx/chunk-text text :chunk-size 40 :overlap 10)]
  (assert (> (length chunks) 1)
          "chunk-text: custom chunk-size works"))

# ============================================================
# content-hash
# ============================================================

(let [h1 (idx/content-hash "hello")
      h2 (idx/content-hash "hello")
      h3 (idx/content-hash "world")]
  (assert (= h1 h2) "content-hash: deterministic")
  (assert (not= h1 h3) "content-hash: differs for different content")
  (assert (> (length h1) 10) "content-hash: non-trivial length"))

# ============================================================
# skip?
# ============================================================

(assert (idx/skip? ".git") "skip?: .git -> true")
(assert (idx/skip? "node_modules") "skip?: node_modules -> true")
(assert (idx/skip? "vendor") "skip?: vendor -> true")
(assert (idx/skip? "__pycache__") "skip?: __pycache__ -> true")
(assert (idx/skip? ".DS_Store") "skip?: .DS_Store -> true")
(assert (idx/skip? ".venv") "skip?: .venv -> true")
(assert (idx/skip? "build") "skip?: build -> true")
(assert (idx/skip? "target") "skip?: target -> true")
(assert (not (idx/skip? "src")) "skip?: src -> false")
(assert (not (idx/skip? "main.go")) "skip?: main.go -> false")

# ============================================================
# SQLite integration: open-search-db
# ============================================================

(def tmp-dir (string "/tmp/test-indexer-" (os/time)))
(os/mkdir tmp-dir)
(os/mkdir (string tmp-dir "/.glitch"))

(var test-db nil)

(defer (do
         (when test-db (sql/close test-db))
         (os/execute ["rm" "-rf" tmp-dir] :p))

  (set test-db (idx/open-search-db tmp-dir))
  (assert test-db "open-search-db: returns db handle")

  # Verify chunks table exists
  (def tables (sql/eval test-db
    `SELECT name FROM sqlite_master WHERE type='table' AND name='chunks'`))
  (assert (= 1 (length tables))
          "open-search-db: chunks table exists")

  # Verify FTS5 table exists
  (def fts-tables (sql/eval test-db
    `SELECT name FROM sqlite_master WHERE type='table' AND name='chunks_fts'`))
  (assert (= 1 (length fts-tables))
          "open-search-db: chunks_fts FTS5 table exists")

  # Verify index_meta table exists
  (def meta-tables (sql/eval test-db
    `SELECT name FROM sqlite_master WHERE type='table' AND name='index_meta'`))
  (assert (= 1 (length meta-tables))
          "open-search-db: index_meta table exists")

  # Insert a row and query FTS
  (sql/eval test-db
    `INSERT INTO chunks (repo, path, content, language, symbols, hash, indexed_at)
     VALUES ('test-repo', 'main.go', 'func Hello() {}', 'go', 'Hello', 'abc123', 0)`)

  (def fts-results (sql/eval test-db
    `SELECT * FROM chunks_fts WHERE chunks_fts MATCH 'Hello'`))
  (assert (> (length fts-results) 0)
          "open-search-db: FTS5 query works after insert")

  (sql/close test-db)
  (set test-db nil))

# ============================================================
# walk-repo
# ============================================================

(def walk-dir (string "/tmp/test-walk-" (os/time)))
(os/mkdir walk-dir)
(defer (os/execute ["rm" "-rf" walk-dir] :p)

  # Create test file structure
  (spit (string walk-dir "/main.go") "package main")
  (spit (string walk-dir "/readme.md") "# Hello")
  (os/mkdir (string walk-dir "/src"))
  (spit (string walk-dir "/src/lib.py") "def foo(): pass")
  (os/mkdir (string walk-dir "/.git"))
  (spit (string walk-dir "/.git/config") "gitconfig")
  (os/mkdir (string walk-dir "/node_modules"))
  (spit (string walk-dir "/node_modules/pkg.js") "module.exports = {}")

  (def files (idx/walk-repo walk-dir))
  (assert (> (length files) 0) "walk-repo: finds files")

  # Should find main.go, readme.md, src/lib.py
  (def paths (map |($ :rel-path) files))
  (assert (find |(= $ "main.go") paths)
          "walk-repo: finds main.go")
  (assert (find |(= $ "readme.md") paths)
          "walk-repo: finds readme.md")
  (assert (find |(= $ "src/lib.py") paths)
          "walk-repo: finds src/lib.py")

  # Should NOT find .git or node_modules content
  (assert (not (find |(string/has-prefix? ".git/" $) paths))
          "walk-repo: skips .git")
  (assert (not (find |(string/has-prefix? "node_modules/" $) paths))
          "walk-repo: skips node_modules"))

# ============================================================
# index-repo integration test
# ============================================================

(def idx-dir (string "/tmp/test-index-repo-" (os/time)))
(os/mkdir idx-dir)
(def repo-dir (string idx-dir "/repo"))
(os/mkdir repo-dir)
(os/mkdir (string idx-dir "/.glitch"))

(var idx-db nil)

(defer (do
         (when idx-db (sql/close idx-db))
         (os/execute ["rm" "-rf" idx-dir] :p))

  # Create test repo files
  (spit (string repo-dir "/main.go")
    "package main\n\nfunc Hello() {\n\tprintln(\"hello\")\n}\n\ntype Config struct {\n\tName string\n}\n")
  (spit (string repo-dir "/lib.janet")
    "(defn greet [name]\n  (print name))\n\n(def version 1)\n")
  (os/mkdir (string repo-dir "/sub"))
  (spit (string repo-dir "/sub/util.py")
    "def helper():\n    pass\n\nclass Tool:\n    pass\n")

  (set idx-db (idx/open-search-db idx-dir))

  # First index
  (def result (idx/index-repo idx-db repo-dir))
  (assert (> (result :files-indexed) 0)
          "index-repo: indexes files")
  (assert (> (result :chunks-created) 0)
          "index-repo: creates chunks")
  (assert (= (result :total-files) (result :files-indexed))
          "index-repo: total-files = files-indexed on first run")

  # Verify chunks in DB
  (def chunks (sql/eval idx-db
    `SELECT * FROM chunks WHERE repo = :repo`
    {:repo repo-dir}))
  (assert (> (length chunks) 0)
          "index-repo: chunks exist in DB")

  # Verify FTS works
  (def fts-hits (sql/eval idx-db
    `SELECT * FROM chunks_fts WHERE chunks_fts MATCH 'Hello'`))
  (assert (> (length fts-hits) 0)
          "index-repo: FTS finds Hello")

  # Re-index: should skip unchanged
  (def result2 (idx/index-repo idx-db repo-dir))
  (assert (= 0 (result2 :chunks-created))
          "index-repo: re-index skips unchanged (chunks-created = 0)")

  # Force reindex
  (def result3 (idx/index-repo idx-db repo-dir :reindex true))
  (assert (> (result3 :chunks-created) 0)
          "index-repo: reindex true forces re-creation")

  (sql/close idx-db)
  (set idx-db nil))

(end-suite)
