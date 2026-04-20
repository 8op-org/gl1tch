# Hybrid FTS5 + semantic search for the glitch MCP server.
# Combines keyword (BM25) and cosine-similarity scoring with
# configurable weights.

(import sqlite3 :as sql)
(import glitch/mcp/vecmath :as vm)
(import glitch/mcp/embeddings :as emb)

# ============================================================
# Score normalization
# ============================================================

(defn normalize-scores [results]
  "Normalize :score values in an array of tables to 0-1 range.
   Empty -> empty. Single item -> 1. All same -> all 1.
   Otherwise linear scale min->0 max->1."
  (case (length results)
    0 @[]
    1 @[(merge (first results) {:score 1})]
    (do
      (var mn math/inf)
      (var mx (- math/inf))
      (each r results
        (def s (r :score))
        (when (< s mn) (set mn s))
        (when (> s mx) (set mx s)))
      (def range (- mx mn))
      (if (= range 0)
        (map |(merge $ {:score 1}) results)
        (map |(merge $ {:score (/ (- ($ :score) mn) range)}) results)))))

# ============================================================
# Score merging
# ============================================================

(defn merge-scores [keyword-results semantic-results
                    &named keyword-weight semantic-weight]
  "Combine keyword and semantic results by :id.
   Default weights: keyword=0.4 semantic=0.6.
   Returns array sorted descending by merged score."
  (default keyword-weight 0.4)
  (default semantic-weight 0.6)
  (def combined @{})
  (each r keyword-results
    (def id (r :id))
    (put combined id (+ (or (get combined id) 0)
                        (* keyword-weight (r :score)))))
  (each r semantic-results
    (def id (r :id))
    (put combined id (+ (or (get combined id) 0)
                        (* semantic-weight (r :score)))))
  (def out @[])
  (eachp [id score] combined
    (array/push out {:id id :score score}))
  (sort out |(compare> ($0 :score) ($1 :score))))

# ============================================================
# Keyword search (FTS5)
# ============================================================

(defn keyword-search [db query &named repo limit]
  "FTS5 MATCH search against chunks_fts.
   FTS5 rank is negative — we negate it so higher = better.
   Returns array of {:id :path :content :symbols :score}."
  (default limit 20)
  (def sql-query
    (if repo
      `SELECT c.id, c.path, c.content, c.symbols, f.rank
       FROM chunks_fts f
       JOIN chunks c ON c.id = f.rowid
       WHERE chunks_fts MATCH :query AND c.repo = :repo
       ORDER BY f.rank
       LIMIT :limit`
      `SELECT c.id, c.path, c.content, c.symbols, f.rank
       FROM chunks_fts f
       JOIN chunks c ON c.id = f.rowid
       WHERE chunks_fts MATCH :query
       ORDER BY f.rank
       LIMIT :limit`))
  (def params
    (if repo
      {:query query :repo repo :limit limit}
      {:query query :limit limit}))
  (def rows (sql/eval db sql-query params))
  (def out @[])
  (each row rows
    (array/push out
      {:id (row :id)
       :path (row :path)
       :content (row :content)
       :symbols (or (row :symbols) "")
       :score (- (row :rank))}))
  out)

# ============================================================
# Semantic search (cosine similarity)
# ============================================================

(defn semantic-search [db query repo &named embed-fn limit]
  "Vector similarity search. Embeds query via embed-fn, computes
   cosine similarity against all repo embeddings.
   Returns array of {:id :score} sorted descending, top N."
  (default limit 20)
  (when (nil? embed-fn) (break @[]))
  # Embed the query text
  (def query-vecs (embed-fn [query]))
  (def query-vec (first query-vecs))
  # Load all chunks with embeddings for this repo
  (def rows (sql/eval db
    `SELECT id, embedding FROM chunks
     WHERE repo = :repo AND embedding IS NOT NULL`
    {:repo repo}))
  (def scored @[])
  (each row rows
    (def stored-vec (emb/unpack-f32 (row :embedding)))
    (def sim (vm/cosine-similarity query-vec stored-vec))
    (array/push scored {:id (row :id) :score sim}))
  (sort scored |(compare> ($0 :score) ($1 :score)))
  (array/slice scored 0 (min limit (length scored))))

# ============================================================
# Hybrid search
# ============================================================

(defn hybrid-search [db query repo &named embed-fn limit]
  "Combined keyword + semantic search.
   Runs both, normalizes scores, merges with 0.4/0.6 weighting,
   enriches top results with content from chunks table."
  (default limit 10)
  # Run both searches
  (def kw-raw (keyword-search db query :repo repo :limit 20))
  (def sem-raw (semantic-search db query repo :embed-fn embed-fn :limit 20))
  # Normalize
  (def kw-norm (normalize-scores kw-raw))
  (def sem-norm (normalize-scores sem-raw))
  # Merge
  (def merged (merge-scores kw-norm sem-norm
                :keyword-weight 0.4 :semantic-weight 0.6))
  # Take top N
  (def top (array/slice merged 0 (min limit (length merged))))
  # Enrich with content from chunks table
  (def out @[])
  (each item top
    (def rows (sql/eval db
      `SELECT path, content, symbols, language FROM chunks WHERE id = :id`
      {:id (item :id)}))
    (if (> (length rows) 0)
      (let [row (first rows)]
        (array/push out
          {:path (row :path)
           :content (row :content)
           :symbols (or (row :symbols) "")
           :score (item :score)
           :language (row :language)}))
      # ID came from merge but chunk may have been deleted; skip
      nil))
  out)
