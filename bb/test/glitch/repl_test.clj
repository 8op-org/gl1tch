(ns glitch.repl-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.repl :as repl]))

(deftest port-file-path-test
  (testing "port-file-path returns .nrepl-port in given dir"
    (is (= "/tmp/.nrepl-port" (repl/port-file-path "/tmp")))))
