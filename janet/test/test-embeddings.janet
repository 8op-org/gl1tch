(use spork/test)

# Add src to module path so we can import glitch/mcp/embeddings
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/mcp/embeddings :as emb)
(import spork/json)

(start-suite "embeddings")

# --- pack-f32 / unpack-f32 roundtrip ---

(assert (deep= @[1.0 2.0 3.0]
               (emb/unpack-f32 (emb/pack-f32 [1.0 2.0 3.0])))
        "pack/unpack roundtrip preserves values")

# empty vector
(assert (deep= @[]
               (emb/unpack-f32 (emb/pack-f32 [])))
        "pack/unpack roundtrip with empty vector")

# single element
(assert (deep= @[42.5]
               (emb/unpack-f32 (emb/pack-f32 [42.5])))
        "pack/unpack roundtrip single element")

# pack returns a buffer
(assert (buffer? (emb/pack-f32 [1 2 3]))
        "pack-f32 returns a buffer")

# --- parse-embedding-response ---

(def sample-response
  (json/encode
    {"object" "list"
     "data" [{"object" "embedding"
              "index" 0
              "embedding" [0.1 0.2 0.3]}
             {"object" "embedding"
              "index" 1
              "embedding" [0.4 0.5 0.6]}]
     "model" "text-embedding-nomic-embed-text-v1.5"
     "usage" {"prompt_tokens" 10 "total_tokens" 10}}))

(def parsed (emb/parse-embedding-response sample-response))
(assert (= 2 (length parsed))
        "parse-embedding-response returns correct count")
(assert (deep= @[0.1 0.2 0.3] (in parsed 0))
        "parse-embedding-response first vector correct")
(assert (deep= @[0.4 0.5 0.6] (in parsed 1))
        "parse-embedding-response second vector correct")

# single embedding
(def single-response
  (json/encode
    {"object" "list"
     "data" [{"object" "embedding"
              "index" 0
              "embedding" [1.0 2.0]}]
     "model" "test"
     "usage" {"prompt_tokens" 1 "total_tokens" 1}}))

(def single-parsed (emb/parse-embedding-response single-response))
(assert (= 1 (length single-parsed))
        "parse-embedding-response handles single embedding")

# --- embed with mock http-fn ---

(defn mock-http-post [url &named headers body]
  "Mock HTTP POST that returns a fake LM Studio embedding response."
  (def req (json/decode body))
  (def texts (get req "input"))
  # Return a fake embedding for each input text
  (def data @[])
  (each i (range (length texts))
    (array/push data
      {"object" "embedding"
       "index" i
       "embedding" [0.1 0.2 0.3]}))
  (json/encode
    {"object" "list"
     "data" data
     "model" (get req "model")
     "usage" {"prompt_tokens" 5 "total_tokens" 5}}))

(def result
  (emb/embed ["hello" "world"]
             :model "test-model"
             :base-url "http://localhost:1234"
             :http-fn mock-http-post))

(assert (= 2 (length result))
        "embed returns one vector per input text")
(assert (deep= @[0.1 0.2 0.3] (in result 0))
        "embed returns correct vector values")

# single text input
(def single-result
  (emb/embed ["just one"]
             :model "test-model"
             :base-url "http://localhost:1234"
             :http-fn mock-http-post))
(assert (= 1 (length single-result))
        "embed handles single text input")

(end-suite)
