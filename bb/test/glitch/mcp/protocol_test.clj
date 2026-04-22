(ns glitch.mcp.protocol-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.mcp.protocol :as proto]
            [cheshire.core :as json]))

(deftest parse-message-test
  (testing "valid JSON"
    (let [msg (proto/parse-message "{\"method\":\"initialize\",\"id\":1}")]
      (is (= "initialize" (get msg "method")))
      (is (= 1 (get msg "id")))))
  (testing "invalid JSON"
    (let [msg (proto/parse-message "not json")]
      (is (:error msg)))))

(deftest format-result-test
  (let [resp (json/parse-string (proto/format-result 1 {"ok" true}))]
    (is (= "2.0" (get resp "jsonrpc")))
    (is (= 1 (get resp "id")))
    (is (= {"ok" true} (get resp "result")))))

(deftest format-error-test
  (let [resp (json/parse-string (proto/format-error 1 -32601 "Method not found"))]
    (is (= "2.0" (get resp "jsonrpc")))
    (is (= -32601 (get-in resp ["error" "code"])))
    (is (= "Method not found" (get-in resp ["error" "message"])))))

(deftest format-tool-result-test
  (let [resp (json/parse-string (proto/format-tool-result 1 "hello"))]
    (is (= [{"type" "text" "text" "hello"}]
           (get-in resp ["result" "content"])))))

(deftest format-tool-error-test
  (let [resp (json/parse-string (proto/format-tool-error 1 "oops"))]
    (is (true? (get-in resp ["result" "isError"])))
    (is (= "oops" (get-in resp ["result" "content" 0 "text"])))))

(deftest dispatch-test
  (testing "initialize"
    (let [resp (proto/dispatch {"id" 1 "method" "initialize"} {})
          parsed (json/parse-string resp)]
      (is (= "glitch" (get-in parsed ["result" "serverInfo" "name"])))
      (is (= "2024-11-05" (get-in parsed ["result" "protocolVersion"])))))
  (testing "tools/list"
    (let [tools [{"name" "test"}]
          resp (proto/dispatch {"id" 2 "method" "tools/list"} {:tools tools})
          parsed (json/parse-string resp)]
      (is (= tools (get-in parsed ["result" "tools"])))))
  (testing "tools/call"
    (let [handler (fn [_ _] "result")
          ctx {:tool-handler handler}
          resp (proto/dispatch {"id" 3 "method" "tools/call"
                                "params" {"name" "t" "arguments" {}}} ctx)
          parsed (json/parse-string resp)]
      (is (= "result" (get-in parsed ["result" "content" 0 "text"])))))
  (testing "unknown method"
    (let [resp (proto/dispatch {"id" 4 "method" "bogus"} {})
          parsed (json/parse-string resp)]
      (is (= -32601 (get-in parsed ["error" "code"])))))
  (testing "notification (no id) returns nil"
    (is (nil? (proto/dispatch {"method" "notifications/initialized"} {})))))
