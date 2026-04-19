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

(defn websearch [query &named max-results]
  "Search the web via DuckDuckGo. Returns array of {:title :url :snippet}."
  (default max-results 5)
  (def encoded (string/replace-all " " "+" query))
  (def url (string "https://html.duckduckgo.com/html/?q=" encoded))
  (def proc (os/spawn
    ["curl" "-sS" "-L"
     "-H" "User-Agent: glitch/0.1"
     url]
    :p {:out :pipe :err :pipe}))
  (def out (string (ev/read (proc :out) :all)))
  (def err-out (ev/read (proc :err) :all))
  (os/proc-wait proc)
  # Parse result links from DuckDuckGo HTML
  (def results @[])
  (def link-peg
    (peg/compile
      ~{:main (any (+ :result 1))
        :result (* "class=\"result__a\"" (thru "href=\"") (capture (to "\"")) "\"" (thru ">")
                   (capture (to "</a>"))
                   (thru "class=\"result__snippet")
                   (thru ">") (capture (to "</")))}))
  (def matches (peg/match link-peg out))
  (when matches
    (var i 0)
    (while (and (< i (length matches)) (< (length results) max-results))
      (when (>= (+ i 2) (length matches)) (break))
      (def url-raw (in matches i))
      (def title (in matches (+ i 1)))
      (def snippet (in matches (+ i 2)))
      (array/push results
        {:title (string/trim title)
         :url (string/trim url-raw)
         :snippet (string/trim snippet)})
      (set i (+ i 3))))
  results)

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
