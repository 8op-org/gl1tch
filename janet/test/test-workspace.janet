(use spork/test)

# Add src to module path so we can import glitch/workspace
(array/push module/paths ["src/:all:.janet" :source])

(import glitch/workspace :as ws)

(start-suite "workspace")

# Create a test workspace file
(def tmp-dir (string "/tmp/glitch-ws-test-" (os/time)))
(os/shell (string "mkdir -p " tmp-dir "/.glitch"))

# Write workspace file — imports glitch/workspace via module/paths
(spit (string tmp-dir "/.glitch/workspace.janet")
  (string
    `(import glitch/workspace :as ws)
     (ws/workspace "test-ws"
       :description "Test workspace"
       :owner "tester")
     (ws/defaults
       :model "qwen2.5:7b"
       :provider "ollama")
     (ws/resource "myrepo"
       :type "git"
       :url "https://github.com/example/repo"
       :ref "main")
     (ws/resource "local-docs"
       :type "local"
       :path "/tmp/docs")`))

(def w (ws/load (string tmp-dir "/.glitch/workspace.janet")))
(assert (= "test-ws" (w :name)) "workspace name")
(assert (= "Test workspace" (w :description)) "workspace description")
(assert (= "qwen2.5:7b" (get-in w [:defaults :model])) "default model")
(assert (= "ollama" (get-in w [:defaults :provider])) "default provider")
(assert (= 2 (length (w :resources))) "two resources")

(def repo (ws/get-resource w "myrepo"))
(assert (= "git" (repo :type)) "resource type")
(assert (= "main" (repo :ref)) "resource ref")

(def docs (ws/get-resource w "local-docs"))
(assert (= "local" (docs :type)) "local resource type")
(assert (= "/tmp/docs" (docs :path)) "local resource path")

# find-workspace-file walks up
(def found (ws/find-workspace-file tmp-dir))
(assert found "find-workspace-file finds workspace")

# find-workspace-file returns nil when nothing found
(def not-found (ws/find-workspace-file "/tmp"))
(assert (nil? not-found) "find-workspace-file returns nil when not found")

# resources-by-type
(def git-res (ws/resources-by-type w "git"))
(assert (= 1 (length git-res)) "one git resource")
(assert (= "myrepo" ((first git-res) :name)) "git resource name")

# get-resource returns nil for missing
(assert (nil? (ws/get-resource w "nonexistent")) "missing resource is nil")

# cleanup
(os/shell (string "rm -rf " tmp-dir))

(end-suite)
