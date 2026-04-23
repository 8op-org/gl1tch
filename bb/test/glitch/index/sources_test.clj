(ns glitch.index.sources-test
  (:require [clojure.test :refer [deftest testing is]]
            [glitch.index.sources :as sources]
            [glitch.index.sources.github :as github]
            [glitch.index.sources.gitlab :as gitlab]))

;; --- Registry ---------------------------------------------------------------

(deftest defindexer-registers-source
  (testing "defindexer adds source to registry atom"
    (is (contains? @sources/registry "github"))
    (is (contains? @sources/registry "gitlab"))))

(deftest get-source-returns-registered
  (testing "get-source returns source map with required keys"
    (let [src (sources/get-source "github")]
      (is (some? src))
      (is (= "github" (:name src)))
      (is (fn? (:detect src)))
      (is (fn? (:fetch src)))
      (is (fn? (:index-name src)))
      (is (map? (:mapping src)))))
  (testing "get-source returns nil for unknown name"
    (is (nil? (sources/get-source "nonexistent-source")))))

(deftest list-sources-returns-all-with-detection
  (testing "list-sources returns vec with :name :doc :detected"
    (let [results (sources/list-sources "/tmp/fake-repo")]
      (is (vector? results))
      (is (pos? (count results)))
      (doseq [r results]
        (is (contains? r :name))
        (is (contains? r :doc))
        (is (boolean? (:detected r)))))))

;; --- GitHub detection -------------------------------------------------------

(deftest github-detect-remote-urls
  (let [detect (:detect (sources/get-source "github"))]
    (testing "HTTPS remote"
      (with-redefs [babashka.process/shell
                    (fn [_opts & _args]
                      {:exit 0 :out "https://github.com/owner/repo.git\n" :err ""})]
        (is (detect "/tmp/repo"))))
    (testing "SSH remote"
      (with-redefs [babashka.process/shell
                    (fn [_opts & _args]
                      {:exit 0 :out "git@github.com:owner/repo.git\n" :err ""})]
        (is (detect "/tmp/repo"))))))

;; --- GitLab detection -------------------------------------------------------

(deftest gitlab-detect-remote-urls
  (let [detect (:detect (sources/get-source "gitlab"))]
    (testing "HTTPS remote"
      (with-redefs [babashka.process/shell
                    (fn [_opts & _args]
                      {:exit 0 :out "https://gitlab.com/group/project.git\n" :err ""})]
        (is (detect "/tmp/repo"))))
    (testing "SSH remote with subgroups"
      (with-redefs [babashka.process/shell
                    (fn [_opts & _args]
                      {:exit 0 :out "git@gitlab.com:group/subgroup/project.git\n" :err ""})]
        (is (detect "/tmp/repo"))))))

;; --- GitHub document shape --------------------------------------------------

(deftest github-transform-issue-shape
  (testing "GitHub transform produces correct document fields"
    (let [raw {:number 42
               :title "Fix the widget"
               :state "OPEN"
               :body "The widget is broken"
               :author {:login "alice"}
               :labels [{:name "bug"} {:name "P1"}]
               :assignees [{:login "alice"}]
               :url "https://github.com/owner/repo/issues/42"
               :createdAt "2026-01-15T10:00:00Z"
               :updatedAt "2026-04-20T12:00:00Z"
               :closedAt nil}
          doc (#'github/transform-issue "owner/repo" raw)]
      (is (= "github:owner/repo:issue:42" (:id doc)))
      (is (= "github" (:source doc)))
      (is (= "issue" (:type doc)))
      (is (= 42 (:number doc)))
      (is (= "Fix the widget" (:title doc)))
      (is (= "open" (:state doc)))
      (is (= "The widget is broken" (:body doc)))
      (is (= "alice" (:author doc)))
      (is (= ["bug" "P1"] (:labels doc)))
      (is (= ["alice"] (:assignees doc)))
      (is (= "2026-01-15T10:00:00Z" (:created_at doc)))
      (is (= "2026-04-20T12:00:00Z" (:updated_at doc)))
      (is (string? (:indexed_at doc))))))

;; --- GitLab document shape --------------------------------------------------

(deftest gitlab-transform-issue-shape
  (testing "GitLab transform produces correct document fields"
    (let [raw {:iid 99
               :title "Add dark mode"
               :state "opened"
               :description "Users want dark mode"
               :author {:username "bob"}
               :assignees [{:username "bob"}]
               :labels ["enhancement" "frontend"]
               :web_url "https://gitlab.com/group/project/-/issues/99"
               :created_at "2026-02-01T08:00:00Z"
               :updated_at "2026-04-22T14:00:00Z"
               :closed_at nil}
          now "2026-04-23T00:00:00Z"
          doc (#'gitlab/transform-issue "group/project" now raw)]
      (is (= "gitlab:group/project:issue:99" (:id doc)))
      (is (= "gitlab" (:source doc)))
      (is (= "issue" (:type doc)))
      (is (= 99 (:number doc)))
      (is (= "Add dark mode" (:title doc)))
      (is (= "opened" (:state doc)))
      (is (= "Users want dark mode" (:body doc)))
      (is (= "bob" (:author doc)))
      (is (= ["enhancement" "frontend"] (:labels doc)))
      (is (= now (:indexed_at doc))))))
