(ns glitch.gui.telemetry-test
  (:require [clojure.test :refer [deftest is testing]]
            [clojure.string :as str]
            [glitch.gui.telemetry :as tel]))

(deftest new-run-id-test
  (let [id (tel/new-run-id)]
    (is (str/starts-with? id "run-"))
    (is (> (count id) 10))))

(deftest nil-safe-test
  (testing "nil telemetry no-ops"
    (is (nil? (tel/index-run nil {})))
    (is (nil? (tel/index-workflow-run nil {})))
    (is (nil? (tel/ensure-indices nil)))))
