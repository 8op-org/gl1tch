(ns glitch.gui.telemetry
  "Elasticsearch telemetry client — thin HTTP wrapper for bulk indexing."
  (:require [babashka.http-client :as http]
            [cheshire.core :as json]
            [clojure.string :as str]))

(def ^:private index-names
  {:runs "glitch-runs"
   :workflow-runs "glitch-workflow-runs"
   :llm-calls "glitch-llm-calls"
   :tool-calls "glitch-tool-calls"
   :research-runs "glitch-research-runs"
   :cross-reviews "glitch-cross-reviews"
   :learnings "glitch-learnings"})

(defn new-run-id []
  (format "run-%d-%s"
    (System/currentTimeMillis)
    (subs (str (java.util.UUID/randomUUID)) 0 8)))

(defn- strip-trailing-slash [s]
  (if (str/ends-with? s "/")
    (subs s 0 (dec (count s)))
    s))

(defn- ping [base-url]
  (try
    (let [resp (http/get base-url {:timeout 2000})]
      (< (:status resp) 400))
    (catch Exception _ false)))

(defn connect
  "Create a telemetry client. Returns nil if ES unreachable."
  [& {:keys [base-url] :or {base-url "http://localhost:9200"}}]
  (when (ping base-url)
    {:base-url (strip-trailing-slash base-url)}))

(defn ensure-indices [client]
  (when client
    (doseq [[_ idx-name] index-names]
      (try
        (http/put (str (:base-url client) "/" idx-name)
          {:headers {"Content-Type" "application/json"}
           :body "{\"mappings\":{\"dynamic\":true}}"})
        (catch Exception _)))))

(defn- index-doc [client index doc]
  (when client
    (let [id (or (:run_id doc) (:run-id doc) (new-run-id))
          url (str (:base-url client) "/" index "/_doc/" id)]
      (try
        (http/put url
          {:headers {"Content-Type" "application/json"}
           :body (json/generate-string doc)})
        (catch Exception _)))))

(defn index-run [client doc]
  (index-doc client (:runs index-names) doc))

(defn index-workflow-run [client doc]
  (index-doc client (:workflow-runs index-names) doc))

(defn index-llm-call [client doc]
  (index-doc client (:llm-calls index-names) doc))

(defn index-tool-call [client doc]
  (index-doc client (:tool-calls index-names) doc))

(defn search [client indices query-body]
  (when client
    (try
      (let [url (str (:base-url client) "/" (str/join "," indices) "/_search")
            resp (http/post url
                   {:headers {"Content-Type" "application/json"}
                    :body (json/generate-string query-body)})]
        (when (< (:status resp) 400)
          (json/parse-string (:body resp) true)))
      (catch Exception _ nil))))
