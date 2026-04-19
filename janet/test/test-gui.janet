(use spork/test)

(array/push module/paths ["src/:all:.janet" :source])

(import glitch/gui :as gui)
(import glitch/store :as s)
(import spork/json)

(start-suite "gui")

# Test JSON response helper
(def resp (gui/json-response {:hello "world"}))
(assert (= 200 (resp :status)) "json-response status 200")
(assert (= "application/json"
           (get-in resp [:headers "Content-Type"]))
        "json-response content type")
(def body (json/decode (resp :body)))
(assert (= "world" (get body "hello"))
        "json-response encodes body")

# Test route matching
(def params (gui/route-match "/api/runs/:id" "/api/runs/42"))
(assert params "route-match matches parameterized route")
(assert (= "42" (params :id)) "route-match extracts param")

(assert (nil? (gui/route-match "/api/runs/:id" "/api/workflows"))
        "route-match rejects non-matching route")

(assert (gui/route-match "/api/runs" "/api/runs")
        "route-match matches exact route")

# Test build-routes returns routes
(def db-path (string "/tmp/glitch-gui-test-" (os/time) ".db"))
(def db (s/open db-path))
(def routes (gui/build-routes db nil))
(assert (> (length routes) 0) "build-routes returns routes")

# cleanup
(s/close db)
(os/rm db-path)

(end-suite)
