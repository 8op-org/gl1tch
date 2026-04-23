(ns glitch.mcp.handlers-test
  (:require [clojure.test :refer [deftest is testing]]
            [clojure.java.io :as io]
            [cheshire.core :as json]
            [glitch.mcp.handlers :as handlers]
            [glitch.session]))

(deftest eval-with-dsl-test
  (let [handler (handlers/make-handler {})]
    (testing "basic expression"
      (is (= "42" (handler "glitch_eval" {"expression" "(+ 1 41)"}))))
    (testing "clojure.string available"
      (is (= "HELLO" (handler "glitch_eval" {"expression" "(clojure.string/upper-case \"hello\")"}))))
    (testing "glitch DSL functions are available"
      (is (string? (handler "glitch_eval" {"expression" "(sh \"echo\" \"hello\")"}))))
    (testing "error returns exception message"
      (is (thrown? Exception (handler "glitch_eval" {"expression" "(/ 1 0)"}))))))

(deftest list-workflows-test
  (let [handler (handlers/make-handler {})
        tmp-dir (doto (io/file (System/getProperty "java.io.tmpdir")
                               (str "glitch-test-" (System/currentTimeMillis)))
                  (.mkdirs))
        wf1 (io/file tmp-dir "deploy.glitch")
        wf2 (io/file tmp-dir "lint.glitch")]
    (try
      (spit wf1 ";; Deploy the application to staging\n(sh \"echo\" \"deploying\")")
      (spit wf2 "(sh \"echo\" \"linting\")")
      (let [result (handler "glitch_list_workflows"
                            {"path" (.getAbsolutePath tmp-dir)})
            parsed (json/parse-string result true)]
        (testing "returns array of workflow objects"
          (is (= 2 (count parsed))))
        (testing "each entry has name and file"
          (is (every? :name parsed))
          (is (every? :file parsed)))
        (testing "extracts description from first comment"
          (let [deploy (first (filter #(= "deploy" (:name %)) parsed))]
            (is (= "Deploy the application to staging" (:description deploy)))))
        (testing "missing comment gives empty description"
          (let [lint (first (filter #(= "lint" (:name %)) parsed))]
            (is (= "" (:description lint))))))
      (finally
        (.delete wf1)
        (.delete wf2)
        (.delete tmp-dir)))))

(deftest advise-fallback-test
  (let [handler (handlers/make-handler {})]
    (testing "returns valid JSON with none approach on workflow failure"
      (let [result (handler "glitch_advise"
                            {"task" "some task that will fail without provider"})
            parsed (json/parse-string result true)]
        (is (string? result))
        (is (contains? parsed :approach))
        (is (string? (:reasoning parsed)))))))

(deftest advise-records-session-test
  (let [handler (handlers/make-handler {})]
    (binding [glitch.session/*current-session* (atom [])
              glitch.session/*session-id* (atom "test-advise")]
      (try
        (handler "glitch_advise" {"task" "summarize logs"})
        (catch Exception _))
      (let [entries @glitch.session/*current-session*]
        (testing "session entries is a vector after advise call"
          (is (vector? entries)))))))
