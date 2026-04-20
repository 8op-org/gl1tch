(ns glitch.mcp.indexer
  (:require [babashka.pods :as pods]
            [babashka.process :as bp]
            [babashka.fs :as fs]
            [glitch.mcp.embeddings :as emb]
            [clojure.string :as str]))

(pods/load-pod 'org.babashka/go-sqlite3 "0.2.8")
(require '[pod.babashka.go-sqlite3 :as sql])

;; --- Extension -> language map ---

(def ^:private ext-map
  {".go" "go" ".clj" "clojure" ".cljs" "clojure" ".cljc" "clojure"
   ".janet" "janet" ".py" "python" ".js" "javascript" ".ts" "typescript"
   ".tsx" "typescript" ".jsx" "javascript" ".rs" "rust" ".rb" "ruby"
   ".java" "java" ".c" "c" ".cpp" "cpp" ".h" "c"
   ".md" "markdown" ".yaml" "yaml" ".yml" "yaml" ".json" "json"
   ".sh" "shell" ".bash" "shell" ".zsh" "shell"
   ".sql" "sql" ".html" "html" ".css" "css"})

(defn detect-language [path]
  (when path
    (let [dot-idx (str/last-index-of path ".")]
      (when dot-idx
        (get ext-map (subs path dot-idx))))))

;; --- Symbol extraction via regex ---

(def ^:private symbol-patterns
  {"go"         [#"(?m)^func\s+(?:\([^)]+\)\s+)?(\w+)"
                 #"(?m)^type\s+(\w+)"]
   "clojure"    [#"\((?:defn-?|defmacro|def)\s+([\w\-\?\!]+)"]
   "janet"      [#"\((?:defn-?|defmacro|def)\s+([\w\-\?\!]+)"]
   "python"     [#"(?m)^(?:def|class)\s+(\w+)"]
   "javascript" [#"(?m)(?:export\s+)?(?:function|class|const)\s+(\w+)"]
   "typescript" [#"(?m)(?:export\s+)?(?:function|class|const|interface|type)\s+(\w+)"]
   "rust"       [#"(?m)^(?:pub\s+)?(?:fn|struct|enum|trait)\s+(\w+)"]})

(defn extract-symbols [content language]
  (if (or (nil? content) (nil? language))
    []
    (let [patterns (get symbol-patterns language)]
      (if patterns
        (vec (distinct (mapcat #(map second (re-seq % content)) patterns)))
        []))))

;; --- Text chunking ---

(defn chunk-text [text & {:keys [chunk-size overlap] :or {chunk-size 1500 overlap 150}}]
  (cond
    (or (nil? text) (= text "")) []
    (<= (count text) chunk-size) [text]
    :else
    (let [stride (- chunk-size overlap)]
      (loop [pos 0 chunks []]
        (if (>= pos (count text))
          chunks
          (let [end (min (+ pos chunk-size) (count text))]
            (recur (+ pos stride) (conj chunks (subs text pos end)))))))))

;; --- Content hashing ---

(defn content-hash [content]
  (let [result (bp/shell {:in content :out :string} "shasum" "-a" "256")]
    (first (str/split (str/trim (:out result)) #"\s+"))))

;; --- Skip list ---

(def ^:private skip-set
  #{".git" "node_modules" "vendor" "__pycache__" ".venv" "venv"
    "dist" "build" "target" ".next" ".idea" ".vscode" ".DS_Store"})

(defn skip? [name]
  (contains? skip-set name))

;; --- SQLite schema ---

(def ^:private schema-stmts
  ["CREATE TABLE IF NOT EXISTS chunks (
     id         INTEGER PRIMARY KEY AUTOINCREMENT,
     repo       TEXT NOT NULL,
     path       TEXT NOT NULL,
     content    TEXT NOT NULL,
     language   TEXT,
     symbols    TEXT,
     hash       TEXT NOT NULL,
     embedding  BLOB,
     indexed_at INTEGER NOT NULL
   )"
   "CREATE INDEX IF NOT EXISTS idx_chunks_repo ON chunks(repo)"
   "CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path)"
   "CREATE INDEX IF NOT EXISTS idx_chunks_hash ON chunks(hash)"
   "CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
     path, content, symbols, content=chunks, content_rowid=id
   )"
   "CREATE TRIGGER IF NOT EXISTS chunks_ai AFTER INSERT ON chunks BEGIN
     INSERT INTO chunks_fts(rowid, path, content, symbols)
     VALUES (new.id, new.path, new.content, new.symbols);
   END"
   "CREATE TRIGGER IF NOT EXISTS chunks_ad AFTER DELETE ON chunks BEGIN
     INSERT INTO chunks_fts(chunks_fts, rowid, path, content, symbols)
     VALUES ('delete', old.id, old.path, old.content, old.symbols);
   END"
   "CREATE TABLE IF NOT EXISTS index_meta (
     repo       TEXT PRIMARY KEY,
     model      TEXT,
     dimensions INTEGER,
     indexed_at INTEGER NOT NULL
   )"])

(defn open-search-db [workspace-path]
  (let [db-dir (str workspace-path "/.glitch")
        db-path (str db-dir "/search.db")]
    (fs/create-dirs db-dir)
    (doseq [stmt schema-stmts]
      (sql/execute! db-path [stmt]))
    db-path))

(defn close-search-db [_db]
  nil)

;; --- Directory walking ---

(defn walk-repo [repo-path]
  (let [results (atom [])]
    (letfn [(walk [dir rel-prefix]
              (doseq [f (sort (fs/list-dir dir))]
                (let [fname (str (fs/file-name f))
                      full (str f)
                      rel (if (= rel-prefix "") fname (str rel-prefix "/" fname))]
                  (when-not (skip? fname)
                    (cond
                      (fs/directory? f)
                      (walk full rel)

                      (fs/regular-file? f)
                      (when-let [lang (detect-language fname)]
                        (swap! results conj {:path full :rel-path rel :language lang})))))))]
      (walk repo-path ""))
    @results))

;; --- Main indexing ---

(defn index-repo [db repo-path & {:keys [embed-fn model reindex]}]
  (when reindex
    (sql/execute! db ["DELETE FROM chunks WHERE repo = ?" repo-path]))

  (let [existing-hashes (if reindex
                          #{}
                          (set (map :hash (sql/query db ["SELECT hash FROM chunks WHERE repo = ?" repo-path]))))
        files (walk-repo repo-path)
        pending-embeds (atom [])
        files-indexed (atom 0)
        chunks-created (atom 0)]

    (doseq [{:keys [path rel-path language]} files]
      (let [content (slurp path)
            chunks (chunk-text content)
            file-had-new (atom false)]
        (doseq [chunk chunks]
          (let [hash (content-hash chunk)]
            (when-not (contains? existing-hashes hash)
              (let [syms (extract-symbols chunk language)
                    sym-str (str/join " " syms)
                    now (quot (System/currentTimeMillis) 1000)]
                (if embed-fn
                  (swap! pending-embeds conj
                    {:repo repo-path :path rel-path :content chunk
                     :language language :symbols sym-str :hash hash
                     :indexed-at now})
                  (do
                    (sql/execute! db
                      ["INSERT INTO chunks (repo, path, content, language, symbols, hash, indexed_at)
                        VALUES (?, ?, ?, ?, ?, ?, ?)"
                       repo-path rel-path chunk language sym-str hash now])
                    (swap! chunks-created inc))))
              (reset! file-had-new true))))
        (when @file-had-new
          (swap! files-indexed inc))))

    ;; Batch embed pending chunks in groups of 32
    (when (and embed-fn (pos? (count @pending-embeds)))
      (doseq [batch (partition-all 32 @pending-embeds)]
        (let [texts (mapv :content batch)
              embeddings (embed-fn texts)]
          (doseq [[rec emb-vec] (map vector batch embeddings)]
            (let [packed (emb/pack-f32 emb-vec)]
              (sql/execute! db
                ["INSERT INTO chunks (repo, path, content, language, symbols, hash, embedding, indexed_at)
                  VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
                 (:repo rec) (:path rec) (:content rec)
                 (:language rec) (:symbols rec) (:hash rec)
                 packed (:indexed-at rec)])
              (swap! chunks-created inc)))))
      (let [embed-files (set (map :path @pending-embeds))]
        (reset! files-indexed (count embed-files))))

    ;; Delete stale chunks
    (let [seen-paths (set (map :rel-path files))
          db-paths (sql/query db ["SELECT DISTINCT path FROM chunks WHERE repo = ?" repo-path])]
      (doseq [{p :path} db-paths]
        (when-not (contains? seen-paths p)
          (sql/execute! db ["DELETE FROM chunks WHERE repo = ? AND path = ?" repo-path p]))))

    ;; Update index_meta
    (let [now (quot (System/currentTimeMillis) 1000)]
      (sql/execute! db
        ["INSERT OR REPLACE INTO index_meta (repo, model, dimensions, indexed_at) VALUES (?, ?, ?, ?)"
         repo-path (or model "") 0 now]))

    {:files-indexed @files-indexed
     :chunks-created @chunks-created
     :total-files (count files)}))
