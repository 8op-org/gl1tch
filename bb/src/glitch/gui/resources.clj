(ns glitch.gui.resources
  (:require [cheshire.core :as json]
            [babashka.fs :as fs]))

(defn- json-response [status body]
  {:status status
   :headers {"Content-Type" "application/json"}
   :body (json/generate-string body)})

(defn list-resources [{:keys [workspace]}]
  (let [ws-file (str workspace "/workspace.glitch")]
    (if-not (and ws-file (fs/exists? ws-file))
      (json-response 200 [])
      (json-response 200 []))))

(defn add-resource [_ctx]
  (json-response 200 {:status "added"}))

(defn remove-resource [_ctx]
  (json-response 200 {:status "removed"}))

(defn sync-resources [_ctx]
  (json-response 200 {:status "synced"}))

(defn pin-resource [_ctx]
  (json-response 200 {:status "pinned"}))
