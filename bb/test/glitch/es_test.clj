(ns glitch.es-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.es :as es]
            [babashka.http-client :as http]
            [cheshire.core :as json]))

;; ---------------------------------------------------------------------------
;; index-exists?
;; ---------------------------------------------------------------------------

(deftest index-exists?-returns-true-on-200
  (testing "index-exists? returns true when ES responds with 200"
    (with-redefs [http/request (fn [_] {:status 200})]
      (is (true? (es/index-exists? "http://localhost:9200" "test-idx"))))))

(deftest index-exists?-returns-false-on-404
  (testing "index-exists? returns false when ES responds with 404"
    (with-redefs [http/request (fn [_] {:status 404})]
      (is (false? (es/index-exists? "http://localhost:9200" "test-idx"))))))

;; ---------------------------------------------------------------------------
;; create-index
;; ---------------------------------------------------------------------------

(deftest create-index-returns-true-on-200
  (testing "create-index returns true when index is created successfully"
    (with-redefs [http/request (fn [_] {:status 200})]
      (is (true? (es/create-index "http://localhost:9200" "test-idx" {:mappings {}}))))))

(deftest create-index-returns-true-on-400-already-exists
  (testing "create-index returns true when index already exists (400)"
    (with-redefs [http/request (fn [_] {:status 400
                                         :body (json/generate-string
                                                 {:error {:type "resource_already_exists_exception"}})})]
      (is (true? (es/create-index "http://localhost:9200" "test-idx" {:mappings {}}))))))

;; ---------------------------------------------------------------------------
;; delete-index
;; ---------------------------------------------------------------------------

(deftest delete-index-returns-true-on-200
  (testing "delete-index returns true on successful deletion"
    (with-redefs [http/request (fn [_] {:status 200})]
      (is (true? (es/delete-index "http://localhost:9200" "test-idx"))))))

(deftest delete-index-returns-true-on-404
  (testing "delete-index returns true when index not found (404)"
    (with-redefs [http/request (fn [_] {:status 404})]
      (is (true? (es/delete-index "http://localhost:9200" "test-idx"))))))

;; ---------------------------------------------------------------------------
;; doc-count
;; ---------------------------------------------------------------------------

(deftest doc-count-parses-count-from-response
  (testing "doc-count extracts count from ES _count response body"
    (with-redefs [http/request (fn [_] {:status 200
                                         :body (json/generate-string {:count 42})})]
      (is (= 42 (es/doc-count "http://localhost:9200" "test-idx"))))))

;; ---------------------------------------------------------------------------
;; bulk-index
;; ---------------------------------------------------------------------------

(deftest bulk-index-constructs-correct-ndjson-body
  (testing "bulk-index sends correct NDJSON body to /_bulk"
    (let [captured (atom nil)]
      (with-redefs [http/request (fn [opts]
                                   (reset! captured opts)
                                   {:status 200
                                    :body (json/generate-string {:errors false :items []})})]
        (es/bulk-index "http://localhost:9200" "test-idx"
                       [{:name "foo"} {:name "bar"}])
        (let [body (:body @captured)
              lines (clojure.string/split-lines body)]
          ;; 2 docs => 4 NDJSON lines (action + doc for each) + trailing newline
          (is (= 4 (count (remove clojure.string/blank? lines))))
          ;; First line is an index action
          (let [action (json/parse-string (first lines) true)]
            (is (= "test-idx" (get-in action [:index :_index]))))
          ;; Second line is the first doc
          (let [doc (json/parse-string (second lines) true)]
            (is (= "foo" (:name doc))))
          ;; Verify content-type is ndjson
          (is (= "application/x-ndjson" (get-in @captured [:headers "Content-Type"]))))))))

;; ---------------------------------------------------------------------------
;; search
;; ---------------------------------------------------------------------------

(deftest search-parses-hits-from-response
  (testing "search returns vector of _source maps from hits"
    (with-redefs [http/request (fn [_] {:status 200
                                         :body (json/generate-string
                                                 {:hits {:total {:value 2}
                                                         :hits [{:_source {:name "foo" :kind "function"}}
                                                                {:_source {:name "bar" :kind "struct"}}]}})})]
      (let [results (es/search "http://localhost:9200" "test-idx" {:query {:match_all {}}})]
        (is (= 2 (count results)))
        (is (= "foo" (:name (first results))))
        (is (= "bar" (:name (second results))))))))

;; ---------------------------------------------------------------------------
;; terms-agg
;; ---------------------------------------------------------------------------

(deftest terms-agg-constructs-query-and-parses-buckets
  (testing "terms-agg sends correct query and returns aggregation buckets"
    (let [captured (atom nil)]
      (with-redefs [http/request (fn [opts]
                                   (reset! captured opts)
                                   {:status 200
                                    :body (json/generate-string
                                            {:aggregations
                                             {:by_field
                                              {:buckets [{:key "go" :doc_count 10}
                                                         {:key "python" :doc_count 5}]}}})})]
        (let [buckets (es/terms-agg "http://localhost:9200" "test-idx" "language" ["go" "python"])]
          ;; Verify the returned buckets
          (is (= 2 (count buckets)))
          (is (= "go" (:key (first buckets))))
          (is (= 10 (:doc_count (first buckets))))
          ;; Verify the query structure sent to ES
          (let [body (json/parse-string (:body @captured) true)]
            (is (= 0 (:size body)))
            (is (some? (get-in body [:aggs :by_field :terms])))))))))
