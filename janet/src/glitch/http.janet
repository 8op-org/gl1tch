# HTTP client helpers and JSON path extraction.

(import spork/json)

(defn http-get [url &named headers]
  "HTTP GET request via curl. Returns response body as string."
  (def args @["curl" "-sS" "-L"])
  (when headers
    (eachp [k v] headers
      (array/push args "-H" (string k ": " v))))
  (array/push args url)
  (def proc (os/spawn args :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "http-get %s failed (exit %d): %s" url exit (string err-out)))
  (string out))

(defn http-post [url &named headers body]
  "HTTP POST request via curl. Returns response body as string."
  (def args @["curl" "-sS" "-L" "-X" "POST"])
  (when headers
    (eachp [k v] headers
      (array/push args "-H" (string k ": " v))))
  (when body
    (array/push args "-d" body))
  (array/push args url)
  (def proc (os/spawn args :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (def err-out (ev/read (proc :err) :all))
  (def exit (os/proc-wait proc))
  (unless (= exit 0)
    (errorf "http-post %s failed (exit %d): %s" url exit (string err-out)))
  (string out))

(defn json-pick [json-str path]
  "Extract a value from a JSON string using a dotted path.
   '$' returns the whole parsed object."
  (def parsed (json/decode json-str))
  (if (= path "$")
    parsed
    (do
      (var current parsed)
      (each seg (string/split "." path)
        (when (nil? current) (break))
        (set current (get current seg)))
      current)))
