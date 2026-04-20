(ns glitch.mcp.vecmath-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.mcp.vecmath :as vm]))

(deftest dot-product-test
  (is (= 32.0 (vm/dot-product [1 2 3] [4 5 6])))
  (is (= 0.0 (vm/dot-product [1 0 0] [0 1 0])))
  (is (= 0.0 (vm/dot-product [] []))))

(deftest magnitude-test
  (is (= 5.0 (vm/magnitude [3 4])))
  (is (= 1.0 (vm/magnitude [1 0 0])))
  (is (= 0.0 (vm/magnitude []))))

(deftest cosine-similarity-test
  (testing "identical vectors"
    (is (= 1.0 (vm/cosine-similarity [1 2 3] [1 2 3]))))
  (testing "orthogonal vectors"
    (is (= 0.0 (vm/cosine-similarity [1 0] [0 1]))))
  (testing "zero vector returns 0"
    (is (= 0.0 (vm/cosine-similarity [0 0] [1 2])))
    (is (= 0.0 (vm/cosine-similarity [1 2] [0 0])))))
