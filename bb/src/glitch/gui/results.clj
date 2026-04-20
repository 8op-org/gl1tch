(ns glitch.gui.results
  (:require [cheshire.core :as json]
            [clojure.string :as str]
            [babashka.fs :as fs]))

(defn- json-response [status body]
  {:status status
   :headers {"Content-Type" "application/json"}
   :body (json/generate-string body)})

(defn- results-dir [ctx]
  (str (:workspace ctx) "/results"))

(defn get-result [{:keys [req] :as ctx}]
  (let [uri (:uri req)
        path (subs uri (count "/api/results/"))
        full-path (str (results-dir ctx) "/" path)]
    (if (str/includes? path "..")
      (json-response 400 {:error "invalid path"})
      (if-not (fs/exists? full-path)
        (json-response 404 {:error "not found"})
        {:status 200
         :headers {"Content-Type" "text/plain"}
         :body (slurp full-path)}))))

(defn put-result [{:keys [req] :as ctx}]
  (let [uri (:uri req)
        path (subs uri (count "/api/results/"))
        full-path (str (results-dir ctx) "/" path)
        body (slurp (:body req))]
    (if (str/includes? path "..")
      (json-response 400 {:error "invalid path"})
      (do
        (fs/create-dirs (str (fs/parent full-path)))
        (spit full-path body)
        (json-response 200 {:status "saved"})))))
