(ns glitch.gui.runs
  (:require [glitch.store :as store]
            [cheshire.core :as json]
            [clojure.string :as str]))

(defn- json-response [status body]
  {:status status
   :headers {"Content-Type" "application/json"}
   :body (json/generate-string body)})

(defn- parse-query-params [req]
  (when-let [qs (:query-string req)]
    (into {} (map #(let [[k v] (str/split % #"=" 2)] [k v]) (str/split qs #"&")))))

(defn list-runs [{:keys [req db]}]
  (if-not db
    (json-response 200 [])
    (let [params (parse-query-params req)
          workflow (get params "workflow")
          parent-id (when-let [p (get params "parent_id")] (parse-long p))
          runs (cond
                 parent-id (store/list-runs db :parent-id parent-id)
                 workflow (store/list-runs db :workflow workflow)
                 :else (store/list-runs db :limit 100))]
      (json-response 200 (or runs [])))))

(defn get-run [{:keys [req db]}]
  (if-not db
    (json-response 500 {:error "store not available"})
    (let [uri (:uri req)
          id (parse-long (re-find #"\d+" (subs uri (count "/api/runs/"))))
          run (store/get-run db id)
          steps (store/get-steps db id)]
      (if run
        (json-response 200 {:run run :steps (or steps [])})
        (json-response 404 {:error "not found"})))))

(defn get-run-tree [{:keys [req db]}]
  (if-not db
    (json-response 200 [])
    (let [uri (:uri req)
          id (parse-long (second (re-matches #"/api/runs/(\d+)/tree" uri)))
          children (store/list-runs db :parent-id id)]
      (json-response 200 (or children [])))))
