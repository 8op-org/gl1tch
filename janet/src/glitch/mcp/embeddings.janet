# LM Studio embedding client — /v1/embeddings + vector packing for SQLite.

(import spork/json)
(import glitch/http :as http)

(defn pack-f32 [vec]
  "Serialize a vector as JSON bytes for SQLite BLOB storage."
  (buffer (json/encode vec)))

(defn unpack-f32 [buf]
  "Deserialize a JSON-encoded vector."
  (def parsed (json/decode (string buf)))
  (array ;parsed))

(defn parse-embedding-response [body]
  "Parse an LM Studio /v1/embeddings JSON response.
   Returns an array of embedding vectors (each an array of numbers)."
  (def parsed (json/decode body))
  (def data (get parsed "data"))
  (def result @[])
  (each item data
    (array/push result (array ;(get item "embedding"))))
  result)

(defn embed [texts &named model base-url http-fn]
  "Embed texts via LM Studio /v1/embeddings endpoint.
   :model     — embedding model name (required)
   :base-url  — LM Studio base URL, e.g. http://localhost:1234
   :http-fn   — injectable HTTP POST function for testing (defaults to http/http-post)"
  (default model "text-embedding-nomic-embed-text-v1.5")
  (default base-url "http://localhost:1234")
  (default http-fn http/http-post)
  (def url (string base-url "/v1/embeddings"))
  (def payload (json/encode {"model" model "input" texts}))
  (def response
    (http-fn url
             :headers {"Content-Type" "application/json"}
             :body payload))
  (parse-embedding-response response))
