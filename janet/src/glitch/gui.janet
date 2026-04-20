# HTTP server for the glitch GUI.

(import spork/http)
(import spork/json)
(import glitch/store :as s)
(import glitch/runner :as runner)
(import glitch/provider :as prov)

# --- Helpers ---

(defn json-response [data &opt status]
  "Build a JSON HTTP response."
  (default status 200)
  {:status status
   :headers {"Content-Type" "application/json"
             "Access-Control-Allow-Origin" "*"}
   :body (string (json/encode data))})

(defn- text-response [text &opt status]
  (default status 200)
  {:status status
   :headers {"Content-Type" "text/plain"}
   :body text})

(defn- parse-json-body [req]
  (when (req :body)
    (json/decode (string (req :body)))))

(defn route-match [pattern path]
  "Match a URL pattern with :param placeholders. Returns params table or nil."
  (def pat-segs (string/split "/" pattern))
  (def path-segs (string/split "/" path))
  (def has-wildcard (and (> (length pat-segs) 0)
                        (string/has-prefix? "*" (last pat-segs))))
  (when (and (not has-wildcard)
             (not= (length pat-segs) (length path-segs)))
    (break nil))
  (def params @{})
  (for i 0 (min (length pat-segs) (length path-segs))
    (def pat (get pat-segs i))
    (def seg (get path-segs i))
    (cond
      (string/has-prefix? ":" pat)
        (put params (keyword (string/slice pat 1)) seg)
      (string/has-prefix? "*" pat)
        (do
          (put params (keyword (string/slice pat 1))
            (string/join (tuple/slice path-segs i) "/"))
          (break))
      (not= pat seg) (break nil)))
  params)

# --- Route builders ---

(defn build-routes [db project-root]
  "Build the route table."
  @[[:get "/api/workflows"
     (fn [req params]
       (def wfs @[])
       (each dir [".glitch/workflows" "workflows"]
         (when (os/stat dir)
           (each f (os/dir dir)
             (when (string/has-suffix? ".janet" f)
               (array/push wfs
                 {:name (string/slice f 0 (- (length f) 6))
                  :path (string dir "/" f)})))))
       (json-response wfs))]

    [:get "/api/runs"
     (fn [req params]
       (def runs (s/list-runs db))
       (json-response runs))]

    [:get "/api/runs/:id"
     (fn [req params]
       (def id (scan-number (params :id)))
       (def run (s/get-run db id))
       (if run
         (do
           (def steps (s/get-steps db id))
           (json-response (merge run {:steps steps})))
         (json-response {:error "not found"} 404)))]

    [:get "/api/project"
     (fn [req params]
       (json-response (if project-root {:root project-root} {})))]

    [:get "/api/providers"
     (fn [req params]
       (json-response (prov/names)))]])

(defn start [opts]
  "Start the GUI HTTP server."
  (def {:addr addr :project-root project-root :static-dir static-dir} opts)
  (default addr "localhost:3000")
  (default static-dir "gui/dist")

  (def db (if project-root
            (s/open-for-project project-root)
            (s/open)))
  (prov/load-providers)

  (def routes (build-routes db project-root))

  (defn handler [req]
    (def method (keyword (string/ascii-lower (or (req :method) "get"))))
    (def path (or (req :path) "/"))

    # Try API routes
    (var matched nil)
    (each [m pattern handler-fn] routes
      (when (and (= m method) (nil? matched))
        (def params (route-match pattern path))
        (when params
          (set matched (handler-fn req params)))))
    (when matched (break matched))

    # SPA fallback
    (text-response "not found" 404))

  (printf "glitch gui → http://%s" addr)
  (def [host port] (string/split ":" addr))
  (http/server handler host (scan-number port)))

(defn- mime-type [path]
  (cond
    (string/has-suffix? ".html" path) "text/html"
    (string/has-suffix? ".js" path) "application/javascript"
    (string/has-suffix? ".css" path) "text/css"
    (string/has-suffix? ".json" path) "application/json"
    (string/has-suffix? ".svg" path) "image/svg+xml"
    (string/has-suffix? ".png" path) "image/png"
    "application/octet-stream"))
