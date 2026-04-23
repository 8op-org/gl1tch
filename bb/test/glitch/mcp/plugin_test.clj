(ns glitch.mcp.plugin-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.mcp.plugin :as plugin]
            [sci.core :as sci]))

;; ---------------------------------------------------------------------------
;; props->input-schema
;; ---------------------------------------------------------------------------

(deftest props->input-schema-mixed-required-test
  (testing "properties with some required, some optional"
    (let [schema (plugin/props->input-schema
                   {"name"  {:type "string" :description "User name" :required true}
                    "age"   {:type "number" :description "User age"  :required false}
                    "email" {:type "string" :description "Email"     :required true}})]
      (is (= "object" (get schema "type")))
      (is (= {"type" "string" "description" "User name"}
             (get-in schema ["properties" "name"])))
      (is (= {"type" "number" "description" "User age"}
             (get-in schema ["properties" "age"])))
      (is (= {"type" "string" "description" "Email"}
             (get-in schema ["properties" "email"])))
      (let [required (set (get schema "required"))]
        (is (contains? required "name"))
        (is (contains? required "email"))
        (is (not (contains? required "age")))))))

(deftest props->input-schema-all-optional-test
  (testing "all optional — no required key in output"
    (let [schema (plugin/props->input-schema
                   {"x" {:type "string" :description "optional field" :required false}})]
      (is (= "object" (get schema "type")))
      (is (not (contains? schema "required"))))))

(deftest props->input-schema-empty-test
  (testing "empty properties map"
    (let [schema (plugin/props->input-schema {})]
      (is (= "object" (get schema "type")))
      (is (= {} (get schema "properties")))
      (is (not (contains? schema "required"))))))

;; ---------------------------------------------------------------------------
;; register-tool! / defmcp-tool
;; ---------------------------------------------------------------------------

(deftest register-tool-test
  (reset! plugin/registry {})
  (testing "register-tool! adds entry with correct definition and handler"
    (let [handler-fn (fn [args] (str "hello " (get args "who")))]
      (plugin/register-tool! "greet" "Say hello" {"who" {:type "string" :description "name" :required true}} handler-fn)
      (let [entry (get @plugin/registry "greet")]
        (is (some? entry))
        (is (= "greet"     (get-in entry [:definition "name"])))
        (is (= "Say hello" (get-in entry [:definition "description"])))
        (is (= "hello world" ((:handler-fn entry) {"who" "world"})))))))

(deftest defmcp-tool-test
  (reset! plugin/registry {})
  (testing "defmcp-tool registers tool the same as register-tool!"
    (plugin/defmcp-tool "echo" "Echo input"
      {"input" {:type "string" :description "text" :required true}}
      (fn [args] (str "echo: " (get args "input"))))
    (let [entry (get @plugin/registry "echo")]
      (is (some? entry))
      (is (= "echo: hi" ((:handler-fn entry) {"input" "hi"}))))))

;; ---------------------------------------------------------------------------
;; tool-definitions
;; ---------------------------------------------------------------------------

(deftest tool-definitions-test
  (reset! plugin/registry {})
  (testing "returns definition maps for all registered tools"
    (plugin/register-tool! "tool_a" "Tool A" {} (fn [_] "a"))
    (plugin/register-tool! "tool_b" "Tool B"
      {"x" {:type "string" :description "x" :required true}}
      (fn [_] "b"))
    (let [defs  (plugin/tool-definitions)
          names (set (map #(get % "name") defs))]
      (is (= 2 (count defs)))
      (is (contains? names "tool_a"))
      (is (contains? names "tool_b")))))

;; ---------------------------------------------------------------------------
;; handle-tool
;; ---------------------------------------------------------------------------

(deftest handle-tool-registered-test
  (reset! plugin/registry {})
  (testing "dispatches to registered handler"
    (plugin/register-tool! "upper" "Uppercase"
      {"text" {:type "string" :description "input" :required true}}
      (fn [args] (clojure.string/upper-case (get args "text"))))
    (is (= "FOO" (plugin/handle-tool "upper" {"text" "foo"})))))

(deftest handle-tool-unknown-test
  (reset! plugin/registry {})
  (testing "returns nil for unregistered tool"
    (is (nil? (plugin/handle-tool "no_such_tool" {})))))

;; ---------------------------------------------------------------------------
;; SCI plugin evaluation (simulates load-tools!)
;; ---------------------------------------------------------------------------

(deftest sci-plugin-eval-test
  (reset! plugin/registry {})
  (testing "evaluating plugin code in SCI context registers the tool"
    (let [ctx (#'plugin/make-plugin-sci-ctx)
          plugin-code "(defmcp-tool \"glitch_test_tool\"
                         \"A test tool\"
                         {\"input\" {:type \"string\" :description \"test input\" :required true}}
                         (fn [args] (str \"got: \" (get args \"input\"))))"]
      (sci/eval-string* ctx plugin-code)
      (let [entry (get @plugin/registry "glitch_test_tool")]
        (is (some? entry) "tool should be registered after SCI eval")
        (is (= "glitch_test_tool" (get-in entry [:definition "name"])))
        (is (= "A test tool"      (get-in entry [:definition "description"])))
        (is (= "got: hello"       ((:handler-fn entry) {"input" "hello"})))))))

;; ---------------------------------------------------------------------------
;; load-tools! filesystem integration
;; ---------------------------------------------------------------------------

(deftest load-tools-filesystem-test
  (reset! plugin/registry {})
  (let [mcp-dir ".glitch/mcp-tools"]
    (try
      (.mkdirs (java.io.File. mcp-dir))
      (spit (str mcp-dir "/test-tool.clj")
            "(defmcp-tool \"glitch_fs_test\"
               \"Filesystem test tool\"
               {\"x\" {:type \"string\" :description \"input\" :required true}}
               (fn [args] (str \"fs: \" (get args \"x\"))))")
      (plugin/load-tools!)
      (testing "tool registered from filesystem"
        (let [entry (get @plugin/registry "glitch_fs_test")]
          (is (some? entry))
          (is (= "glitch_fs_test" (get-in entry [:definition "name"])))
          (is (= "fs: hello" ((:handler-fn entry) {"x" "hello"})))))
      (testing "tool appears in tool-definitions"
        (let [names (set (map #(get % "name") (plugin/tool-definitions)))]
          (is (contains? names "glitch_fs_test"))))
      (finally
        ;; cleanup
        (doseq [f (.listFiles (java.io.File. mcp-dir))]
          (.delete f))
        (.delete (java.io.File. mcp-dir))
        (.delete (java.io.File. ".glitch"))))))
