(ns glitch.gui
  "HTTP server for the glitch GUI — serves Svelte SPA + REST API."
  (:require [org.httpkit.server :as http]
            [glitch.store :as store]
            [glitch.provider :as prov]
            [glitch.gui.telemetry :as tel]
            [glitch.gui.workflows :as wf]
            [glitch.gui.runs :as runs]
            [glitch.gui.resources :as res]
            [glitch.gui.results :as results]
            [glitch.gui.workspace :as ws]
            [cheshire.core :as json]
            [clojure.string :as str]
            [clojure.java.io :as io]
            [babashka.fs :as fs]))

(defn- json-response [status body]
  {:status status
   :headers {"Content-Type" "application/json"}
   :body (json/generate-string body)})

(defn- content-type-for [path]
  (condp #(str/ends-with? %2 %1) path
    ".html" "text/html"
    ".js" "application/javascript"
    ".css" "text/css"
    ".json" "application/json"
    ".svg" "image/svg+xml"
    ".png" "image/png"
    ".woff2" "font/woff2"
    "application/octet-stream"))

(defn- serve-static [dist-dir uri]
  (let [path (if (= uri "/") "index.html" (subs uri 1))
        file (io/file dist-dir path)
        base-canonical (.getCanonicalPath (io/file dist-dir))
        file-canonical (.getCanonicalPath file)]
    (if (and (.exists file) (.isFile file)
             (str/starts-with? file-canonical base-canonical))
      {:status 200
       :headers {"Content-Type" (content-type-for path)}
       :body file}
      ;; SPA fallback
      (let [index (io/file dist-dir "index.html")]
        (if (.exists index)
          {:status 200
           :headers {"Content-Type" "text/html"}
           :body index}
          {:status 404 :body "not found"})))))

(defn- api-handler [req ctx]
  (let [uri (:uri req)
        method (:request-method req)
        ctx (assoc ctx :req req)]
    (cond
      (and (= method :get) (= uri "/api/workflows"))
      (wf/list-workflows ctx)

      (and (= method :get) (str/starts-with? uri "/api/workflows/actions/"))
      (wf/get-actions ctx)

      (and (= method :get) (str/starts-with? uri "/api/workflows/")
           (not (str/includes? (subs uri (count "/api/workflows/")) "/")))
      (wf/get-workflow ctx)

      (and (= method :put) (str/starts-with? uri "/api/workflows/")
           (not (str/includes? (subs uri (count "/api/workflows/")) "/")))
      (wf/put-workflow ctx)

      (and (= method :post) (str/ends-with? uri "/run"))
      (wf/run-workflow ctx)

      (and (= method :get) (= uri "/api/runs"))
      (runs/list-runs ctx)

      (and (= method :get) (re-matches #"/api/runs/\d+/tree" uri))
      (runs/get-run-tree ctx)

      (and (= method :get) (re-matches #"/api/runs/\d+" uri))
      (runs/get-run ctx)

      (and (= method :get) (str/starts-with? uri "/api/results/"))
      (results/get-result ctx)

      (and (= method :put) (str/starts-with? uri "/api/results/"))
      (results/put-result ctx)

      (and (= method :get) (str/starts-with? uri "/api/kibana/"))
      (ws/handle-kibana ctx)

      (and (= method :get) (= uri "/api/workspace"))
      (ws/get-workspace ctx)

      (and (= method :put) (= uri "/api/workspace"))
      (ws/put-workspace ctx)

      (and (= method :get) (= uri "/api/workspaces"))
      (ws/list-workspaces ctx)

      (and (= method :post) (= uri "/api/workspaces/use"))
      (ws/use-workspace ctx)

      (and (= method :get) (= uri "/api/workspace/resources"))
      (res/list-resources ctx)

      (and (= method :post) (= uri "/api/workspace/resources"))
      (res/add-resource ctx)

      (and (= method :delete) (str/starts-with? uri "/api/workspace/resources/"))
      (res/remove-resource ctx)

      (and (= method :post) (str/starts-with? uri "/api/workspace/sync"))
      (res/sync-resources ctx)

      (and (= method :post) (= uri "/api/workspace/pin"))
      (res/pin-resource ctx)

      (and (= method :get) (= uri "/api/providers"))
      (json-response 200 (mapv (fn [[k _]] {:name (name k)}) (or (:providers ctx) {})))

      :else
      (json-response 404 {:error "not found"}))))

(defn start [{:keys [port workspace dist-dir]
              :or {port 3000}}]
  (let [db (if workspace
             (store/open-for-project workspace)
             (store/open))
        tel-client (tel/connect)
        providers (prov/load-providers)]
    (when tel-client (tel/ensure-indices tel-client))

    (let [ctx {:db db :tel tel-client :providers providers :workspace workspace :dist-dir dist-dir}
          handler (fn [req]
                    (let [uri (:uri req)]
                      (if (str/starts-with? uri "/api/")
                        (api-handler req ctx)
                        (if (and (= (:request-method req) :get) dist-dir)
                          (serve-static dist-dir uri)
                          (json-response 404 {:error "not found"})))))]
      (println (str "[glitch-gui] serving on http://localhost:" port))
      (http/run-server handler {:port port}))))
