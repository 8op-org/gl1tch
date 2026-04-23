(ns glitch.mcp-test
  (:require [clojure.test :refer [deftest is testing]]
            [babashka.process :as bp]
            [cheshire.core :as json]))

(deftest mcp-initialize-test
  (testing "server responds to initialize"
    (let [proc (bp/process {:dir "/Users/stokes/Projects/gl1tch/bb"
                            :in :pipe :out :pipe :err :inherit}
                 "bb" "-cp" "src:providers" "-m" "glitch.mcp")
          w (java.io.BufferedWriter. (java.io.OutputStreamWriter. (:in proc)))
          r (java.io.BufferedReader. (java.io.InputStreamReader. (:out proc)))]
      (try
        (Thread/sleep 500)
        (.write w (str (json/generate-string {"jsonrpc" "2.0" "id" 1 "method" "initialize" "params" {}}) "\n"))
        (.flush w)
        (let [line (.readLine r)
              resp (json/parse-string line)]
          (is (= "glitch" (get-in resp ["result" "serverInfo" "name"]))))
        (finally
          (.close w)
          @proc)))))

(deftest mcp-tools-list-test
  (testing "server lists 9 tools"
    (let [proc (bp/process {:dir "/Users/stokes/Projects/gl1tch/bb"
                            :in :pipe :out :pipe :err :inherit}
                 "bb" "-cp" "src:providers" "-m" "glitch.mcp")
          w (java.io.BufferedWriter. (java.io.OutputStreamWriter. (:in proc)))
          r (java.io.BufferedReader. (java.io.InputStreamReader. (:out proc)))]
      (try
        (Thread/sleep 500)
        ;; initialize
        (.write w (str (json/generate-string {"jsonrpc" "2.0" "id" 1 "method" "initialize" "params" {}}) "\n"))
        (.flush w)
        (.readLine r)
        ;; tools/list
        (.write w (str (json/generate-string {"jsonrpc" "2.0" "id" 2 "method" "tools/list"}) "\n"))
        (.flush w)
        (let [line (.readLine r)
              resp (json/parse-string line)]
          (is (= 9 (count (get-in resp ["result" "tools"])))))
        (finally
          (.close w)
          @proc)))))
