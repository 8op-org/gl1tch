(use spork/test)

# Add src to module path so we can import glitch/mcp/vecmath
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/mcp/vecmath :as vm)

(start-suite "vecmath")

# --- dot-product ---

(assert (= 0 (vm/dot-product @[1 0] @[0 1]))
        "orthogonal dot product is 0")

(assert (= 1 (vm/dot-product @[1 0] @[1 0]))
        "identical unit vectors dot product is 1")

(assert (= 32 (vm/dot-product @[1 2 3] @[4 5 6]))
        "basic dot product computation")

(assert (= 0 (vm/dot-product @[0 0 0] @[1 2 3]))
        "dot product with zero vector is 0")

# --- magnitude ---

(assert (< (math/abs (- (vm/magnitude @[3 4]) 5)) 1e-9)
        "magnitude of [3,4] is 5")

(assert (= 0 (vm/magnitude @[0 0 0]))
        "magnitude of zero vector is 0")

(assert (< (math/abs (- (vm/magnitude @[1 1 1]) (math/sqrt 3))) 1e-9)
        "magnitude of [1,1,1] is sqrt(3)")

# --- cosine-similarity ---

(assert (< (math/abs (- (vm/cosine-similarity @[1 0] @[1 0]) 1)) 1e-9)
        "identical vectors have cosine similarity 1")

(assert (< (math/abs (vm/cosine-similarity @[1 0] @[0 1])) 1e-9)
        "orthogonal vectors have cosine similarity 0")

(assert (= 0 (vm/cosine-similarity @[0 0] @[1 2]))
        "zero vector returns cosine similarity 0")

(assert (= 0 (vm/cosine-similarity @[1 2] @[0 0]))
        "cosine similarity with second zero vector returns 0")

(assert (< (math/abs (- (vm/cosine-similarity @[1 2 3] @[4 5 6])
                        (vm/cosine-similarity @[4 5 6] @[1 2 3]))) 1e-9)
        "cosine similarity is symmetric")

(assert (< (math/abs (+ (vm/cosine-similarity @[1 0] @[-1 0]) 1)) 1e-9)
        "opposite vectors have cosine similarity -1")

(end-suite)
