(ns glitch.session-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.session :as session]))

(deftest record-advise-test
  (session/init-session!)
  (testing "records advisory entry"
    (let [entry (session/record-advise!
                  {:task "check PR accuracy"
                   :recommendation {:approach "primitive"
                                    :primitives ["grounded?"]
                                    :reasoning "factual check"
                                    :example "(grounded? \"summary\" (ref \"diff\"))"
                                    :existing_workflows []}})]
      (is (= :advise (:type entry)))
      (is (= "check PR accuracy" (:task entry)))
      (is (= "primitive" (get-in entry [:recommendation :approach])))
      (is (nil? (:followed? entry)))))

  (testing "appears in session entries"
    (let [entries (session/current-session)]
      (is (= 1 (count entries)))
      (is (= :advise (:type (first entries)))))))

(deftest update-advise-followed-test
  (session/init-session!)
  (session/record-advise!
    {:task "test task"
     :recommendation {:approach "primitive" :primitives ["consensus"]}})
  (testing "updates followed? flag"
    (session/mark-advise-followed! true)
    (let [entries (session/current-session)
          advise-entry (first (filter #(= :advise (:type %)) entries))]
      (is (true? (:followed? advise-entry))))))
