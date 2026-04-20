(ns glitch.mcp.search
  (:require [glitch.mcp.vecmath :as vm]
            [glitch.mcp.embeddings :as emb]
            [glitch.mcp.indexer]))

;; Pod already loaded by indexer
(require '[pod.babashka.go-sqlite3 :as sql])

(defn normalize-scores [results]
  (case (count results)
    0 []
    1 [(assoc (first results) :score 1.0)]
    (let [scores (map :score results)
          mn (apply min scores)
          mx (apply max scores)
          rng (- mx mn)]
      (if (zero? rng)
        (mapv #(assoc % :score 1.0) results)
        (mapv #(assoc % :score (double (/ (- (:score %) mn) rng))) results)))))

(defn merge-scores [keyword-results semantic-results
                    & {:keys [keyword-weight semantic-weight]
                       :or {keyword-weight 0.4 semantic-weight 0.6}}]
  (let [combined (atom {})]
    (doseq [r keyword-results]
      (swap! combined update (:id r) (fnil + 0.0) (* keyword-weight (:score r))))
    (doseq [r semantic-results]
      (swap! combined update (:id r) (fnil + 0.0) (* semantic-weight (:score r))))
    (sort-by :score > (map (fn [[id score]] {:id id :score score}) @combined))))

(defn keyword-search [db query & {:keys [repo limit] :or {limit 20}}]
  (let [params (if repo
                 ["SELECT c.id, c.path, c.content, c.symbols, f.rank
                   FROM chunks_fts f JOIN chunks c ON c.id = f.rowid
                   WHERE chunks_fts MATCH ? AND c.repo = ?
                   ORDER BY f.rank LIMIT ?"
                  query repo limit]
                 ["SELECT c.id, c.path, c.content, c.symbols, f.rank
                   FROM chunks_fts f JOIN chunks c ON c.id = f.rowid
                   WHERE chunks_fts MATCH ?
                   ORDER BY f.rank LIMIT ?"
                  query limit])
        rows (sql/query db params)]
    (mapv (fn [row]
            {:id (:id row)
             :path (:path row)
             :content (:content row)
             :symbols (or (:symbols row) "")
             :score (- (or (:rank row) 0))})
          rows)))

(defn semantic-search [db query repo & {:keys [embed-fn limit] :or {limit 20}}]
  (if-not embed-fn
    []
    (let [query-vec (first (embed-fn [query]))
          rows (sql/query db
                 ["SELECT id, embedding FROM chunks WHERE repo = ? AND embedding IS NOT NULL" repo])
          scored (map (fn [row]
                        {:id (:id row)
                         :score (vm/cosine-similarity query-vec (emb/unpack-f32 (:embedding row)))})
                      rows)]
      (->> scored (sort-by :score >) (take limit) vec))))

(defn hybrid-search [db query repo & {:keys [embed-fn limit] :or {limit 10}}]
  (let [kw-raw (keyword-search db query :repo repo :limit 20)
        sem-raw (semantic-search db query repo :embed-fn embed-fn :limit 20)
        kw-norm (normalize-scores kw-raw)
        sem-norm (normalize-scores sem-raw)
        merged (merge-scores kw-norm sem-norm)
        top (take limit merged)]
    (->> top
         (keep (fn [{:keys [id score]}]
                 (let [rows (sql/query db ["SELECT path, content, symbols, language FROM chunks WHERE id = ?" id])]
                   (when (seq rows)
                     (let [row (first rows)]
                       {:path (:path row)
                        :content (:content row)
                        :symbols (or (:symbols row) "")
                        :score score
                        :language (:language row)})))))
         vec)))
