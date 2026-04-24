(ns glitch.repl-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.api :as api]
            [glitch.repl :as repl]))

(deftest port-file-path-test
  (testing "port-file-path returns .nrepl-port in given dir"
    (is (= "/tmp/.nrepl-port" (repl/port-file-path "/tmp")))))

(deftest api-symbols-available-test
  (testing "glitch.api public functions exist and have correct types"
    (is (fn? api/agent))
    (is (fn? api/list-tools))
    (is (fn? api/remove-tool!))
    (is (fn? api/use-provider!))
    (is (fn? api/use-model!))
    (is (var? #'api/deftool))))
