# File walking, chunking, symbol extraction, and SQLite+FTS5 storage
# for the glitch MCP semantic search indexer.

(import sqlite3 :as sql)
(import glitch/mcp/embeddings :as emb)

# ============================================================
# Extension -> language map
# ============================================================

(def- ext-map
  {".go" "go"
   ".janet" "janet"
   ".py" "python"
   ".js" "javascript"
   ".ts" "typescript"
   ".tsx" "typescript"
   ".jsx" "javascript"
   ".rs" "rust"
   ".rb" "ruby"
   ".java" "java"
   ".c" "c"
   ".cpp" "cpp"
   ".h" "c"
   ".md" "markdown"
   ".yaml" "yaml"
   ".yml" "yaml"
   ".json" "json"
   ".sh" "shell"
   ".bash" "shell"
   ".zsh" "shell"
   ".sql" "sql"
   ".html" "html"
   ".css" "css"})

(defn detect-language [path]
  "Detect programming language from file extension. Returns string or nil."
  (when path
    (if-let [dot-idx (last (string/find-all "." path))]
      (get ext-map (string/slice path dot-idx))
      nil)))

# ============================================================
# Symbol extraction via PEG
# ============================================================

(def- ident-char
  ~(+ :w (set "-_")))

(def- go-peg
  (peg/compile
    ~{:main (any (+ :func-method :func :type-decl :skip))
      :ws (some :s)
      :ident (<- (some ,ident-char))
      :func (* "func" :ws :ident)
      :func-method (* "func" :ws "(" (any (if-not ")" 1)) ")" :ws :ident)
      :type-decl (* "type" :ws :ident)
      :skip 1}))

(def- janet-peg
  (peg/compile
    ~{:main (any (+ :form :skip))
      :ws (some :s)
      :ident (<- (some ,ident-char))
      :form (* "(" (+ "defmacro" "defn-" "defn" "def") :ws :ident)
      :skip 1}))

(def- python-peg
  (peg/compile
    ~{:main (any (+ :def :class :skip))
      :ws (some :s)
      :ident (<- (some ,ident-char))
      :def (* "def" :ws :ident)
      :class (* "class" :ws :ident)
      :skip 1}))

(def- js-peg
  (peg/compile
    ~{:main (any (+ :export-func :func :const :class :skip))
      :ws (some :s)
      :ident (<- (some ,ident-char))
      :func (* "function" :ws :ident)
      :export-func (* "export" :ws "function" :ws :ident)
      :const (* "const" :ws :ident)
      :class (* "class" :ws :ident)
      :skip 1}))

(def- rust-peg
  (peg/compile
    ~{:main (any (+ :fn-decl :struct :enum :trait :skip))
      :ws (some :s)
      :ident (<- (some ,ident-char))
      :fn-decl (* "fn" :ws :ident)
      :struct (* "struct" :ws :ident)
      :enum (* "enum" :ws :ident)
      :trait (* "trait" :ws :ident)
      :skip 1}))

(defn extract-symbols [content language]
  "Extract symbol names from source code. Returns array of strings."
  (if (or (nil? language) (nil? content))
    @[]
    (let [peg (case language
                "go" go-peg
                "janet" janet-peg
                "python" python-peg
                "javascript" js-peg
                "typescript" js-peg
                "rust" rust-peg
                nil)]
      (if peg
        (or (peg/match peg content) @[])
        @[]))))

# ============================================================
# Text chunking
# ============================================================

(defn chunk-text [text &named chunk-size overlap]
  "Split text into overlapping chunks. Returns array of strings."
  (default chunk-size 1500)
  (default overlap 150)
  (cond
    (or (nil? text) (= text "")) @[]
    (<= (length text) chunk-size) @[text]
    (do
      (def stride (- chunk-size overlap))
      (def chunks @[])
      (var pos 0)
      (while (< pos (length text))
        (def end (min (+ pos chunk-size) (length text)))
        (array/push chunks (string/slice text pos end))
        (set pos (+ pos stride)))
      chunks)))

# ============================================================
# Content hashing
# ============================================================

(defn content-hash [content]
  "SHA-256 hash of content string via shasum subprocess."
  (def proc (os/spawn ["shasum" "-a" "256"] :p
              {:in :pipe :out :pipe}))
  (:write (proc :in) content)
  (:close (proc :in))
  (def out (:read (proc :out) :all))
  (os/proc-wait proc)
  # shasum output: "<hash>  -\n"
  (string/trim (first (string/split " " (string/trim out)))))

# ============================================================
# Skip list
# ============================================================

(def- skip-set
  {".git" true
   "node_modules" true
   "vendor" true
   "__pycache__" true
   ".venv" true
   "venv" true
   "dist" true
   "build" true
   "target" true
   ".next" true
   ".idea" true
   ".vscode" true
   ".DS_Store" true})

(defn skip? [name]
  "Return true if directory/file should be skipped during indexing."
  (truthy? (get skip-set name)))

# ============================================================
# SQLite schema
# ============================================================

(def- schema-stmts
  [`CREATE TABLE IF NOT EXISTS chunks (
     id         INTEGER PRIMARY KEY AUTOINCREMENT,
     repo       TEXT NOT NULL,
     path       TEXT NOT NULL,
     content    TEXT NOT NULL,
     language   TEXT,
     symbols    TEXT,
     hash       TEXT NOT NULL,
     embedding  BLOB,
     indexed_at INTEGER NOT NULL
   )`
   `CREATE INDEX IF NOT EXISTS idx_chunks_repo ON chunks(repo)`
   `CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path)`
   `CREATE INDEX IF NOT EXISTS idx_chunks_hash ON chunks(hash)`
   `CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
     path, content, symbols, content=chunks, content_rowid=id
   )`
   `CREATE TRIGGER IF NOT EXISTS chunks_ai AFTER INSERT ON chunks BEGIN
     INSERT INTO chunks_fts(rowid, path, content, symbols)
     VALUES (new.id, new.path, new.content, new.symbols);
   END`
   `CREATE TRIGGER IF NOT EXISTS chunks_ad AFTER DELETE ON chunks BEGIN
     INSERT INTO chunks_fts(chunks_fts, rowid, path, content, symbols)
     VALUES ('delete', old.id, old.path, old.content, old.symbols);
   END`
   `CREATE TABLE IF NOT EXISTS index_meta (
     repo       TEXT PRIMARY KEY,
     model      TEXT,
     dimensions INTEGER,
     indexed_at INTEGER NOT NULL
   )`])

(defn- mkdir-p [dir]
  "Create directory and all parents."
  (when (and dir (not= dir "") (not (os/stat dir)))
    (def parts (string/find-all "/" dir))
    (when (not (empty? parts))
      (mkdir-p (string/slice dir 0 (last parts))))
    (os/mkdir dir)))

(defn close-search-db [db]
  "Close a search database."
  (sql/close db))

(defn open-search-db [workspace-path]
  "Open SQLite search database at <workspace>/.glitch/search.db.
   Creates schema if needed."
  (def db-dir (string workspace-path "/.glitch"))
  (mkdir-p db-dir)
  (def db-path (string db-dir "/search.db"))
  (def db (sql/open db-path))
  (each stmt schema-stmts
    (sql/eval db stmt))
  db)

# ============================================================
# Directory walking
# ============================================================

(defn walk-repo [repo-path]
  "Walk directory tree, skip dirs from skip?, return array of
   {:path :rel-path :language} for files with detected language."
  (def results @[])
  (defn walk [dir rel-prefix]
    (def entries (os/dir dir))
    (each name entries
      (unless (skip? name)
        (def full (string dir "/" name))
        (def rel (if (= rel-prefix "") name (string rel-prefix "/" name)))
        (def stat (os/stat full))
        (when stat
          (case (stat :mode)
            :directory (walk full rel)
            :file
            (when-let [lang (detect-language name)]
              (array/push results {:path full :rel-path rel :language lang})))))))
  (walk repo-path "")
  results)

# ============================================================
# Main indexing
# ============================================================

(defn index-repo [db repo-path &named embed-fn model reindex]
  "Index a repository into the search database.
   Returns {:files-indexed N :chunks-created N :total-files N}."
  (default model nil)
  (default reindex false)
  (var do-reindex reindex)

  # Check if model changed — force reindex
  (when model
    (def meta-rows (sql/eval db
      `SELECT model FROM index_meta WHERE repo = :repo`
      {:repo repo-path}))
    (when (and (> (length meta-rows) 0)
               (not= ((first meta-rows) :model) model))
      (set do-reindex true)))

  # If reindex, wipe existing chunks
  (when do-reindex
    (sql/eval db `DELETE FROM chunks WHERE repo = :repo` {:repo repo-path}))

  # Collect existing hashes for this repo
  (def existing-hashes @{})
  (unless do-reindex
    (def rows (sql/eval db
      `SELECT hash FROM chunks WHERE repo = :repo`
      {:repo repo-path}))
    (each row rows
      (put existing-hashes (row :hash) true)))

  # Walk repo
  (def files (walk-repo repo-path))
  (def total-files (length files))
  (var files-indexed 0)
  (var chunks-created 0)

  # Track paths we see for stale cleanup
  (def seen-paths @{})

  # Chunks to embed (when embed-fn is provided)
  (def pending-embeds @[])

  (each file files
    (def path (file :path))
    (def rel (file :rel-path))
    (def lang (file :language))
    (put seen-paths rel true)

    (def content (slurp path))
    (def chunks (chunk-text content))
    (var file-had-new false)

    (each chunk chunks
      (def hash (content-hash chunk))
      (unless (get existing-hashes hash)
        (def syms (extract-symbols chunk lang))
        (def sym-str (string/join syms " "))
        (def now (os/time))

        (if embed-fn
          (array/push pending-embeds
            {:repo repo-path :path rel :content chunk
             :language lang :symbols sym-str :hash hash
             :indexed-at now})
          (do
            (sql/eval db
              `INSERT INTO chunks (repo, path, content, language, symbols, hash, indexed_at)
               VALUES (:repo, :path, :content, :lang, :syms, :hash, :at)`
              {:repo repo-path :path rel :content chunk
               :lang lang :syms sym-str :hash hash :at now})
            (++ chunks-created)))
        (set file-had-new true)))

    (when file-had-new (++ files-indexed)))

  # Batch embed pending chunks in groups of 32
  (when (and embed-fn (> (length pending-embeds) 0))
    (var i 0)
    (while (< i (length pending-embeds))
      (def batch-end (min (+ i 32) (length pending-embeds)))
      (def batch (array/slice pending-embeds i batch-end))
      (def texts (map |($ :content) batch))
      (def embeddings (embed-fn texts))

      (each j (range (length batch))
        (def rec (in batch j))
        (def emb-vec (in embeddings j))
        (def packed (emb/pack-f32 emb-vec))
        (sql/eval db
          `INSERT INTO chunks (repo, path, content, language, symbols, hash, embedding, indexed_at)
           VALUES (:repo, :path, :content, :lang, :syms, :hash, :emb, :at)`
          {:repo (rec :repo) :path (rec :path) :content (rec :content)
           :lang (rec :language) :syms (rec :symbols) :hash (rec :hash)
           :emb packed :at (rec :indexed-at)})
        (++ chunks-created))

      (set i batch-end))
    # Count files that had new chunks
    (def embed-files @{})
    (each rec pending-embeds
      (put embed-files (rec :path) true))
    (set files-indexed (length embed-files)))

  # Delete chunks for files that no longer exist
  (def db-paths (sql/eval db
    `SELECT DISTINCT path FROM chunks WHERE repo = :repo`
    {:repo repo-path}))
  (each row db-paths
    (unless (get seen-paths (row :path))
      (sql/eval db
        `DELETE FROM chunks WHERE repo = :repo AND path = :path`
        {:repo repo-path :path (row :path)})))

  # Determine embedding dimensions from first embedded chunk
  (def dims
    (if (and embed-fn (> (length pending-embeds) 0))
      (do
        (def first-emb (sql/eval db
          `SELECT embedding FROM chunks WHERE repo = :repo AND embedding IS NOT NULL LIMIT 1`
          {:repo repo-path}))
        (if (> (length first-emb) 0)
          (length (emb/unpack-f32 ((first first-emb) :embedding)))
          0))
      0))

  # Update index_meta
  (sql/eval db
    `INSERT OR REPLACE INTO index_meta (repo, model, dimensions, indexed_at)
     VALUES (:repo, :model, :dims, :at)`
    {:repo repo-path :model model :dims dims :at (os/time)})

  {:files-indexed files-indexed
   :chunks-created chunks-created
   :total-files total-files})
