(ns glitch.test-runner
  "Entry point for `bb test` — discovers and runs all test namespaces."
  (:require [clojure.test :as t]
            [glitch.core-test]
            [glitch.confidence-test]
            [glitch.provider-test]
            [glitch.store-test]
            [glitch.runner-test]
            [glitch.graph-test]
            [glitch.repl-test]
            [glitch.es-test]
            [glitch.index-languages-test]
            [glitch.index-test]
            [glitch.promote-test]))

(defn -main [& _]
  (let [results (mapv #(t/run-tests %)
                  ['glitch.core-test
                   'glitch.confidence-test
                   'glitch.provider-test
                   'glitch.store-test
                   'glitch.runner-test
                   'glitch.graph-test
                   'glitch.repl-test
                   'glitch.es-test
                   'glitch.index-languages-test
                   'glitch.index-test
                   'glitch.promote-test])
        total-fail (apply + (map :fail results))
        total-err  (apply + (map :error results))]
    (println (str "\n=== " (apply + (map :test results)) " tests, "
                  (apply + (map :pass results)) " passed, "
                  total-fail " failures, " total-err " errors ==="))
    (System/exit (if (and (zero? total-fail) (zero? total-err)) 0 1))))

