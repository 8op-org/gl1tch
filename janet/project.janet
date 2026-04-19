(declare-project
  :name "glitch"
  :description "Workflow engine"
  :dependencies
    [{:url "https://github.com/janet-lang/spork.git" :tag "v1.0.1"}
     "https://github.com/janet-lang/sqlite3.git"])

(array/push module/paths ["src/:all:.janet" :source])

(declare-source
  :source ["src/glitch"]
  :prefix "glitch")

(declare-executable
  :name "glitch"
  :entry "src/glitch/main.janet"
  :install true)
