(use spork/test)

# Add src to module path so we can import glitch/http
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/http :as h)

(start-suite "http")

# json-pick extracts from JSON string
(def data `{"name":"alice","items":[1,2,3]}`)
(assert (= "alice" (h/json-pick data "name"))
        "json-pick extracts top-level key")

# nested pick
(def nested `{"a":{"b":{"c":"deep"}}}`)
(assert (= "deep" (h/json-pick nested "a.b.c"))
        "json-pick handles dotted paths")

# pick with $ returns whole object
(def obj `{"x":1}`)
(def result (h/json-pick obj "$"))
(assert (= 1 (get result "x"))
        "json-pick $ returns whole object")

# json-pick with missing path returns nil
(assert (nil? (h/json-pick data "missing"))
        "json-pick returns nil for missing key")

(end-suite)
