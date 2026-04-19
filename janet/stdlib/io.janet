(import spork/json)

(defn fetch-json [url]
  "HTTP GET a URL and parse the response as JSON."
  (def proc (os/spawn ["curl" "-sS" "-L" url] :p {:out :pipe :err :pipe}))
  (def out (ev/read (proc :out) :all))
  (os/proc-wait proc)
  (json/decode (string out)))

(defn read-lines [path]
  "Read a file and split into lines."
  (def content (string/trim (string (slurp path))))
  (if (= content "")
    @[]
    (string/split "\n" content)))

(defn write-lines [path lines]
  "Join lines and write to file."
  (spit path (string/join lines "\n")))
