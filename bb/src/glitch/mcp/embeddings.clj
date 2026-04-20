(ns glitch.mcp.embeddings
  (:require [babashka.http-client :as http]
            [cheshire.core :as json]))

(defn pack-f32
  "Serialize a vector as JSON string for SQLite BLOB storage."
  [v]
  (json/generate-string v))

(defn unpack-f32
  "Deserialize a JSON-encoded vector from SQLite BLOB."
  [buf]
  (json/parse-string (if (bytes? buf) (String. ^bytes buf) (str buf))))

(defn parse-embedding-response
  "Parse an LM Studio /v1/embeddings response body.
   Returns a vector of embedding vectors."
  [body]
  (let [parsed (json/parse-string body true)]
    (mapv :embedding (:data parsed))))

(defn embed
  "Embed texts via LM Studio /v1/embeddings endpoint.
   Options: :model (default nomic-embed-text), :base-url (default localhost:1234)"
  [texts & {:keys [model base-url]
            :or {model "nomic-embed-text"
                 base-url "http://localhost:1234"}}]
  (let [url (str base-url "/v1/embeddings")
        payload (json/generate-string {"model" model "input" texts})
        response (http/post url {:headers {"Content-Type" "application/json"}
                                 :body payload})]
    (parse-embedding-response (:body response))))
