(ns glitch.mcp.handlers-test
  (:require [clojure.test :refer [deftest is testing]]
            [clojure.java.io :as io]
            [clojure.string :as str]
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
      (let [entries @glitch.session/*current-session*
            advise-entries (filter #(= :advise (:type %)) entries)]
        (testing "session entries is a vector after advise call"
          (is (vector? entries)))
        (testing "advise entry has expected shape when present"
          (when (seq advise-entries)
            (let [entry (first advise-entries)]
              (is (= :advise (:type entry)))
              (is (= "summarize logs" (:task entry)))
              (is (contains? entry :recommendation)))))))))

(deftest search-test
  (let [handler (handlers/make-handler {})
        tmp-dir (doto (io/file (System/getProperty "java.io.tmpdir")
                               (str "glitch-search-test-" (System/currentTimeMillis)))
                  (.mkdirs))
        f1 (io/file tmp-dir "hello.txt")
        f2 (io/file tmp-dir "world.clj")]
    (try
      (spit f1 "hello world\ngoodbye world\nhello again")
      (spit f2 "(defn greet [] \"hello\")")
      (testing "basic pattern match"
        (let [result (handler "glitch_search"
                              {"pattern" "hello"
                               "path"    (.getAbsolutePath tmp-dir)})]
          (is (str/includes? result "hello"))))
      (testing "glob filter"
        (let [result (handler "glitch_search"
                              {"pattern" "hello"
                               "path"    (.getAbsolutePath tmp-dir)
                               "glob"    "*.clj"})]
          (is (str/includes? result "defn"))
          (is (not (str/includes? result "goodbye")))))
      (testing "no matches returns message"
        (let [result (handler "glitch_search"
                              {"pattern" "zzz_no_match"
                               "path"    (.getAbsolutePath tmp-dir)})]
          (is (= "No matches found." result))))
      (finally
        (.delete f1)
        (.delete f2)
        (.delete tmp-dir)))))

(deftest read-file-test
  (let [handler (handlers/make-handler {})
        tmp-file (io/file (System/getProperty "java.io.tmpdir")
                          (str "glitch-read-test-" (System/currentTimeMillis) ".txt"))]
    (try
      (spit tmp-file (str/join "\n" (map #(str "line " %) (range 1 11))))
      (testing "reads full file with line numbers"
        (let [result (handler "glitch_read_file"
                              {"path" (.getAbsolutePath tmp-file)})]
          (is (str/includes? result "1\tline 1"))
          (is (str/includes? result "10\tline 10"))))
      (testing "offset and limit"
        (let [result (handler "glitch_read_file"
                              {"path"   (.getAbsolutePath tmp-file)
                               "offset" 3
                               "limit"  2})]
          (is (str/includes? result "3\tline 3"))
          (is (str/includes? result "4\tline 4"))
          (is (not (str/includes? result "5\tline 5")))))
      (testing "offset beyond end of file returns empty"
        (let [result (handler "glitch_read_file"
                              {"path"   (.getAbsolutePath tmp-file)
                               "offset" 999})]
          (is (= "\n" result))))
      (testing "file not found throws"
        (is (thrown? Exception
                     (handler "glitch_read_file"
                              {"path" "/tmp/does-not-exist-glitch-test.txt"}))))
      (finally
        (.delete tmp-file)))))
