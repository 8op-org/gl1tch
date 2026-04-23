(ns glitch.index.sources.gitlab
  "GitLab issues and merge-request indexer.
   Uses the `glab` CLI to fetch issues/MRs and indexes them into ES."
  (:require [glitch.index.sources :refer [defindexer]]
            [babashka.process :as proc]
            [cheshire.core :as json]
            [clojure.string :as str]))

;; ---------------------------------------------------------------------------
;; Helpers
;; ---------------------------------------------------------------------------

(defn- log [& args]
  (binding [*out* *err*]
    (apply println args)))

(defn- detect-gitlab-remote
  "Shell out to git and check whether origin points at a GitLab host."
  [repo-path]
  (let [res (proc/shell {:out :string :err :string :continue true}
                         "git" "-C" repo-path "remote" "get-url" "origin")]
    (when (zero? (:exit res))
      (let [url (str/trim (:out res))]
        (re-find #"(?i)gitlab" url)))))

(defn- parse-project-path
  "Extract the GitLab project path from a git remote URL.
   Handles HTTPS (gitlab.com/group/sub/project.git)
   and SSH   (git@gitlab.com:group/sub/project.git)."
  [repo-path]
  (let [res (proc/shell {:out :string :err :string :continue true}
                         "git" "-C" repo-path "remote" "get-url" "origin")
        url (str/trim (:out res))]
    (or
     ;; SSH:  git@gitlab.com:group/subgroup/project.git
     (when-let [[_ path] (re-find #"git@[^:]+:(.+?)(?:\.git)?$" url)]
       path)
     ;; HTTPS: https://gitlab.com/group/subgroup/project.git
     (when-let [[_ path] (re-find #"https?://[^/]+/(.+?)(?:\.git)?$" url)]
       path))))

(defn- glab-json
  "Run a glab sub-command from repo-path, return parsed JSON (vec of maps).
   `args` is a flat seq of command strings, e.g. [\"glab\" \"issue\" \"list\" \"-O\" \"json\"]."
  [repo-path args]
  (let [res (apply proc/shell
                    {:out :string :err :string :continue true :dir repo-path}
                    args)]
    (if (zero? (:exit res))
      (let [out (str/trim (:out res))]
        (when-not (str/blank? out)
          (json/parse-string out true)))
      (do
        (log "  WARN: glab command failed:" (pr-str args) (:err res))
        nil))))

(defn- after-date?
  "True when iso-date-str is on or after since-str (both YYYY-MM-DD or ISO)."
  [iso-date-str since-str]
  (when (and iso-date-str since-str)
    (>= (compare (subs iso-date-str 0 10) (subs since-str 0 10)) 0)))

(defn- transform-issue
  "Convert a glab issue JSON map to an ES document."
  [project-path now issue]
  {:id         (str "gitlab:" project-path ":issue:" (:iid issue))
   :source     "gitlab"
   :type       "issue"
   :number     (:iid issue)
   :title      (:title issue)
   :body       (:description issue)
   :state      (:state issue)
   :labels     (:labels issue)
   :author     (get-in issue [:author :username])
   :assignees  (mapv #(:username %) (:assignees issue))
   :url        (:web_url issue)
   :created_at (:created_at issue)
   :updated_at (:updated_at issue)
   :closed_at  (:closed_at issue)
   :repo       project-path
   :indexed_at now})

(defn- transform-mr
  "Convert a glab merge-request JSON map to an ES document."
  [project-path now mr]
  {:id            (str "gitlab:" project-path ":mr:" (:iid mr))
   :source        "gitlab"
   :type          "merge_request"
   :number        (:iid mr)
   :title         (:title mr)
   :body          (:description mr)
   :state         (:state mr)
   :labels        (:labels mr)
   :author        (get-in mr [:author :username])
   :assignees     (mapv #(:username %) (:assignees mr))
   :url           (:web_url mr)
   :created_at    (:created_at mr)
   :updated_at    (:updated_at mr)
   :closed_at     (:closed_at mr)
   :repo          project-path
   :indexed_at    now
   ;; MR-specific
   :target_branch (:target_branch mr)
   :source_branch (:source_branch mr)
   :merged        (= (:state mr) "merged")
   :merged_at     (:merged_at mr)})

;; ---------------------------------------------------------------------------
;; State-flag helpers
;; ---------------------------------------------------------------------------

(defn- issue-state-flags
  "Map canonical state string to glab issue list flags."
  [state]
  (case state
    "closed" ["--closed"]
    "all"    ["--all"]
    []))  ;; default: open issues

(defn- mr-state-flags
  "Map canonical state string to glab mr list flags."
  [state]
  (case state
    "closed" ["--closed"]
    "merged" ["--merged"]
    "all"    ["--all"]
    []))  ;; default: open MRs

;; ---------------------------------------------------------------------------
;; Indexer registration
;; ---------------------------------------------------------------------------

(defindexer gitlab
  {:doc        "Index GitLab issues and merge requests via glab CLI"
   :detect     (fn [repo-path] (detect-gitlab-remote repo-path))
   :index-name (fn [repo-name] (str "glitch-gitlab-" repo-name))
   :mapping    {:properties
                {:id            {:type "keyword"}
                 :source        {:type "keyword"}
                 :type          {:type "keyword"}
                 :number        {:type "integer"}
                 :title         {:type "text"}
                 :body          {:type "text"}
                 :state         {:type "keyword"}
                 :labels        {:type "keyword"}
                 :author        {:type "keyword"}
                 :assignees     {:type "keyword"}
                 :url           {:type "keyword"}
                 :created_at    {:type "date"}
                 :updated_at    {:type "date"}
                 :closed_at     {:type "date"}
                 :repo          {:type "keyword"}
                 :indexed_at    {:type "date"}
                 :target_branch {:type "keyword"}
                 :source_branch {:type "keyword"}
                 :merged        {:type "boolean"}
                 :merged_at     {:type "date"}}}
   :cli-opts   [{:flag "--since" :doc "Only items created on or after this date (YYYY-MM-DD)"}
                {:flag "--limit" :doc "Max items per category (default 100)"}
                {:flag "--state" :doc "Issue/MR state: open, closed, merged, all (default open)"}]
   :fetch      (fn [{:keys [repo-path since limit state]
                     :or   {limit "100" state "open"}}]
                 (let [project-path (parse-project-path repo-path)
                       per-page     (str limit)
                       now          (.toString (java.time.Instant/now))
                       ;; Fetch issues
                       issue-args   (into ["glab" "issue" "list"
                                           "-O" "json"
                                           "--per-page" per-page]
                                          (issue-state-flags state))
                       issues       (or (glab-json repo-path issue-args) [])
                       ;; Fetch MRs
                       mr-args      (into ["glab" "mr" "list"
                                           "-F" "json"
                                           "--per-page" per-page]
                                          (mr-state-flags state))
                       mrs          (or (glab-json repo-path mr-args) [])
                       ;; Transform
                       all-docs     (concat
                                    (map #(transform-issue project-path now %) issues)
                                    (map #(transform-mr project-path now %) mrs))]
                   (log "  GitLab:" (count issues) "issues," (count mrs) "MRs from" project-path)
                   (if since
                     (filter #(after-date? (:created_at %) since) all-docs)
                     all-docs)))})
