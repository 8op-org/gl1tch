(ns glitch.gui.workflows
  (:require [glitch.store :as store]
            [cheshire.core :as json]
            [clojure.string :as str]
            [babashka.fs :as fs]))

(defn- json-response [status body]
  {:status status
   :headers {"Content-Type" "application/json"}
   :body (json/generate-string body)})

(defn- parse-json-body [req]
  (when-let [body (:body req)]
    (json/parse-string (slurp body))))

(defn- workflows-dir [ctx]
  (str (:workspace ctx) "/workflows"))

(defn- extract-params [source]
  (vec (distinct (map second (re-seq #"\{\{\.param\.(\w+)\}\}" source)))))

(defn list-workflows [ctx]
  (let [dir (workflows-dir ctx)]
    (if-not (fs/exists? dir)
      (json-response 200 [])
      (let [entries (fs/list-dir dir)
            workflows (keep (fn [f]
                              (let [name (str (fs/file-name f))]
                                (when (re-matches #".*\.(glitch|yaml|yml)$" name)
                                  {:name name :file name})))
                            entries)]
        (json-response 200 (vec workflows))))))

(defn get-workflow [{:keys [req] :as ctx}]
  (let [uri (:uri req)
        name (subs uri (count "/api/workflows/"))
        path (str (workflows-dir ctx) "/" name)]
    (if (str/includes? name "..")
      (json-response 400 {:error "invalid name"})
      (if-not (fs/exists? path)
        (json-response 404 {:error "not found"})
        (let [source (slurp path)]
          (json-response 200 {:name name :source source :params (extract-params source)}))))))

(defn put-workflow [{:keys [req] :as ctx}]
  (let [uri (:uri req)
        name (subs uri (count "/api/workflows/"))
        body (parse-json-body req)
        source (get body "source")
        path (str (workflows-dir ctx) "/" name)]
    (if (str/includes? name "..")
      (json-response 400 {:error "invalid name"})
      (do
        (spit path source)
        (json-response 200 {:status "saved"})))))

(defn run-workflow [{:keys [req db] :as ctx}]
  (let [uri (:uri req)
        name (second (re-find #"/api/workflows/([^/]+)/run" uri))
        body (parse-json-body req)
        params (get body "params" {})
        run-id (when db (store/record-run db {:name name :workflow-file name}))]
    (future
      (try
        (let [path (str (workflows-dir ctx) "/" name)
              content (slurp path)]
          (when (and db run-id)
            (store/finish-run db run-id content 0)))
        (catch Exception e
          (when (and db run-id)
            (store/finish-run db run-id (.getMessage e) 1)))))
    (json-response 200 {:status "started" :run_id run-id})))

(defn get-actions [{:keys [req] :as ctx}]
  (json-response 200 []))
