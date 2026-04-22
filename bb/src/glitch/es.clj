(ns glitch.es
  "Thin Elasticsearch REST client over babashka.http-client."
  (:require [babashka.http-client :as http]
            [cheshire.core :as json]
            [clojure.string :as str]))

(defn- parse-body [resp]
  (some-> resp :body (json/parse-string true)))

(defn- req [method url & {:keys [body]}]
  (let [opts (cond-> {:method method :uri url :throw false
                       :headers {"Content-Type" "application/json"}}
               body (assoc :body (if (string? body) body (json/generate-string body))))]
    (http/request opts)))

(defn index-exists?
  "HEAD /{index}. Returns boolean."
  [url index]
  (let [resp (req :head (str url "/" index))]
    (= 200 (:status resp))))

(defn create-index
  "PUT /{index} with mappings body. Returns true; ignores 400 (already exists)."
  [url index mappings]
  (let [resp (req :put (str url "/" index) :body mappings)]
    (or (= 200 (:status resp))
        (= 400 (:status resp)))))

(defn delete-index
  "DELETE /{index}. Returns true; ignores 404."
  [url index]
  (let [resp (req :delete (str url "/" index))]
    (or (= 200 (:status resp))
        (= 404 (:status resp)))))

(defn doc-count
  "GET /{index}/_count. Returns count as integer."
  [url index]
  (let [resp (req :get (str url "/" index "/_count"))]
    (or (-> resp parse-body :count) 0)))

(defn bulk-index
  "POST /_bulk with NDJSON body. Returns parsed bulk response."
  [url index docs]
  (let [lines (mapcat (fn [doc]
                         [(json/generate-string {:index {:_index index}})
                          (json/generate-string doc)])
                       docs)
        ndjson (str (str/join "\n" lines) "\n")
        resp   (http/request {:method :post
                              :uri    (str url "/_bulk")
                              :headers {"Content-Type" "application/x-ndjson"}
                              :body   ndjson
                              :throw  false})]
    (parse-body resp)))

(defn search
  "POST /{index}/_search. Returns vector of _source maps."
  [url index query]
  (let [resp (-> (req :post (str url "/" index "/_search") :body query)
                 parse-body)]
    (mapv :_source (get-in resp [:hits :hits]))))

(defn delete-by-query
  "POST /{index}/_delete_by_query. Returns deleted count."
  [url index query]
  (let [resp (-> (req :post (str url "/" index "/_delete_by_query") :body query)
                 parse-body)]
    (or (:deleted resp) 0)))

(defn terms-agg
  "POST /{index}/_search with terms filter + agg. Returns aggregation buckets."
  [url index field values]
  (let [query {:size 0
               :query {:terms {(keyword field) values}}
               :aggs {:by_field {:terms {:field field :size (count values)}}}}
        resp (-> (req :post (str url "/" index "/_search") :body query)
                 parse-body)]
    (get-in resp [:aggregations :by_field :buckets])))
