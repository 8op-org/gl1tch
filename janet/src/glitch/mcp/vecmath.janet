# Pure Janet vector math utilities for MCP semantic search.

(defn dot-product
  "Compute the dot product of two numeric arrays."
  [a b]
  (var sum 0)
  (for i 0 (length a)
    (set sum (+ sum (* (a i) (b i)))))
  sum)

(defn magnitude
  "Compute the euclidean magnitude (L2 norm) of a vector."
  [v]
  (math/sqrt (dot-product v v)))

(defn cosine-similarity
  "Compute cosine similarity of two vectors. Returns 0 if either vector is zero."
  [a b]
  (def mag-a (magnitude a))
  (def mag-b (magnitude b))
  (if (or (= mag-a 0) (= mag-b 0))
    0
    (/ (dot-product a b) (* mag-a mag-b))))
