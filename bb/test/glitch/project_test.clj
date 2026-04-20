(ns glitch.project-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.project :as project]
            [clojure.java.io :as io]))

(deftest resolve-with-flag-test
  (testing "resolve returns :flag when provided"
    (is (= "/some/path" (project/resolve :flag "/some/path")))))

(deftest resolve-finds-glitch-dir-test
  (testing "resolve finds .glitch directory in tmp hierarchy"
    (let [root "/tmp/glitch-project-test"
          dot  (io/file root ".glitch")]
      (try
        (.mkdirs dot)
        ;; Set user.dir to a subdirectory
        (let [sub (str root "/sub/deep")]
          (.mkdirs (io/file sub))
          (let [orig (System/getProperty "user.dir")]
            (System/setProperty "user.dir" sub)
            (try
              (is (= root (project/resolve)))
              (finally
                (System/setProperty "user.dir" orig)))))
        (finally
          (doseq [f (reverse (file-seq (io/file root)))]
            (.delete f)))))))

(deftest resolve-returns-nil-when-not-found-test
  (testing "resolve returns nil when no .glitch directory exists"
    (let [orig (System/getProperty "user.dir")]
      (System/setProperty "user.dir" "/tmp")
      (try
        ;; /tmp should not have a .glitch dir
        (is (nil? (project/resolve)))
        (finally
          (System/setProperty "user.dir" orig))))))
