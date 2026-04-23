(ns glitch.index.sources.github
  "GitHub issues and pull requests indexer.
   Uses the `gh` CLI to fetch data and indexes into Elasticsearch."
  (:require [glitch.index.sources :refer [defindexer]]
            [babashka.process :as proc]
            [cheshire.core :as json]
            [clojure.string :as str]))

;; --- Helpers ----------------------------------------------------------------

(defn- detect-github-remote
  "Parse owner/repo from a git remote URL.
   Handles HTTPS (github.com/owner/repo) and SSH (github.com:owner/repo)."
  [repo-path]
  (let [result (proc/shell {:out :string :err :string :continue true}
                           "git" "-C" repo-path "remote" "get-url" "origin")]
    (when (zero? (:exit result))
      (let [url (str/trim (:out result))]
        (when (str/includes? url "github.com")
          (second (re-find #"github\.com[:/]([^/]+/[^/.\s]+?)(?:\.git)?$" url)))))))

(defn- gh-json
  "Shell out to `gh` with --json fields, parse result. Returns parsed vec or nil."
  [repo-path args]
  (let [result (apply proc/shell
                      {:out :string :err :string :continue true :dir repo-path}
                      args)]
    (when (zero? (:exit result))
      (let [out (str/trim (:out result))]
        (when-not (str/blank? out)
          (json/parse-string out true))))))

(defn- now-iso [] (str (java.time.Instant/now)))

(defn- transform-issue [owner-repo item]
  {:id         (str "github:" owner-repo ":issue:" (:number item))
   :source     "github"
   :type       "issue"
   :number     (:number item)
   :title      (:title item)
   :body       (:body item)
   :state      (str/lower-case (or (:state item) ""))
   :labels     (mapv :name (:labels item))
   :author     (get-in item [:author :login])
   :assignees  (mapv :login (:assignees item))
   :url        (:url item)
   :created_at (:createdAt item)
   :updated_at (:updatedAt item)
   :closed_at  (:closedAt item)
   :repo       owner-repo
   :indexed_at (now-iso)})

(defn- transform-pr [owner-repo item]
  {:id          (str "github:" owner-repo ":pr:" (:number item))
   :source      "github"
   :type        "pull_request"
   :number      (:number item)
   :title       (:title item)
   :body        (:body item)
   :state       (str/lower-case (or (:state item) ""))
   :labels      (mapv :name (:labels item))
   :author      (get-in item [:author :login])
   :assignees   (mapv :login (:assignees item))
   :url         (:url item)
   :created_at  (:createdAt item)
   :updated_at  (:updatedAt item)
   :closed_at   (:closedAt item)
   :repo        owner-repo
   :indexed_at  (now-iso)
   :base_branch (:baseRefName item)
   :head_branch (:headRefName item)
   :merged      (some? (:mergedAt item))
   :merged_at   (:mergedAt item)})

(defn- since-filter [since items]
  (if (str/blank? since)
    items
    (filter #(>= (compare (or (:updatedAt %) "") since) 0) items)))

;; --- Indexer registration ---------------------------------------------------

(defindexer github
  {:doc    "Index GitHub issues and pull requests via the gh CLI"
   :detect (fn [repo-path] (some? (detect-github-remote repo-path)))
   :index-name (fn [repo-name] (str "glitch-github-" repo-name))

   :mapping
   {:properties
    {:id          {:type "keyword"}
     :source      {:type "keyword"}
     :type        {:type "keyword"}
     :number      {:type "integer"}
     :title       {:type "text"}
     :body        {:type "text"}
     :state       {:type "keyword"}
     :labels      {:type "keyword"}
     :author      {:type "keyword"}
     :assignees   {:type "keyword"}
     :url         {:type "keyword"}
     :created_at  {:type "date"}
     :updated_at  {:type "date"}
     :closed_at   {:type "date"}
     :repo        {:type "keyword"}
     :indexed_at  {:type "date"}
     :base_branch {:type "keyword"}
     :head_branch {:type "keyword"}
     :merged      {:type "boolean"}
     :merged_at   {:type "date"}}}

   :cli-opts
   [{:flag "--since" :doc "Only items updated after this date (YYYY-MM-DD)"}
    {:flag "--limit" :doc "Max items to fetch (default: 100)"}
    {:flag "--state" :doc "Filter by state: open, closed, all (default: all)"}]

   :fetch
   (fn [{:keys [repo-path since limit state]}]
     (let [owner-repo   (detect-github-remote repo-path)
           lim          (str (or limit 100))
           st           (or state "all")
           issue-fields "number,title,body,state,labels,author,assignees,url,createdAt,updatedAt,closedAt"
           pr-fields    "number,title,body,state,labels,author,assignees,url,createdAt,updatedAt,closedAt,baseRefName,headRefName,mergedAt"
           issues (gh-json repo-path
                            ["gh" "issue" "list"
                             "--json" issue-fields
                             "--state" st "--limit" lim])
           prs    (gh-json repo-path
                            ["gh" "pr" "list"
                             "--json" pr-fields
                             "--state" st "--limit" lim])]
       (concat
        (->> (or issues []) (since-filter since) (mapv #(transform-issue owner-repo %)))
        (->> (or prs [])    (since-filter since) (mapv #(transform-pr owner-repo %))))))})
