(ns glitch.mcp.vecmath)

(defn dot-product
  "Compute dot product of two numeric sequences."
  [a b]
  (reduce + 0.0 (map * a b)))

(defn magnitude
  "Compute Euclidean norm (L2) of a vector."
  [v]
  (Math/sqrt (dot-product v v)))

(defn cosine-similarity
  "Compute cosine similarity. Returns 0 if either vector is zero."
  [a b]
  (let [mag-a (magnitude a)
        mag-b (magnitude b)]
    (if (or (zero? mag-a) (zero? mag-b))
      0.0
      (/ (dot-product a b) (* mag-a mag-b)))))
