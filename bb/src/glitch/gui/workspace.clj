(ns glitch.gui.workspace
  (:require [glitch.gui.telemetry :as tel]
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

(defn get-workspace [{:keys [workspace]}]
  (json-response 200 {:path workspace
                       :name (when workspace (str (fs/file-name workspace)))}))

(defn put-workspace [_ctx]
  (json-response 200 {:status "updated"}))

(defn list-workspaces [_ctx]
  (json-response 200 []))

(defn use-workspace [{:keys [req]}]
  (let [body (parse-json-body req)
        name (get body "name")]
    (json-response 200 {:status "switched" :name name})))

(defn handle-kibana [{:keys [req tel]}]
  (let [uri (:uri req)]
    (cond
      (str/starts-with? uri "/api/kibana/workflow/")
      (let [name (subs uri (count "/api/kibana/workflow/"))
            results (when tel
                      (tel/search tel ["glitch-workflow-runs"]
                        {:query {:match {:workflow_name name}}
                         :size 20
                         :sort [{:timestamp {:order "desc"}}]}))]
        (json-response 200 (or (get-in results [:hits :hits]) [])))

      (str/starts-with? uri "/api/kibana/run/")
      (let [id (subs uri (count "/api/kibana/run/"))
            results (when tel
                      (tel/search tel ["glitch-llm-calls" "glitch-tool-calls"]
                        {:query {:match {:run_id id}}
                         :size 100}))]
        (json-response 200 (or (get-in results [:hits :hits]) [])))

      :else
      (json-response 404 {:error "not found"}))))
