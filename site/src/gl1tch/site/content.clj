(ns gl1tch.site.content
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]
            [clojure.string :as str]))

(defmacro load-content [dir]
  (let [content-dir (io/file "content" dir)
        files (when (.exists content-dir)
                (->> (.listFiles content-dir)
                     (filter #(str/ends-with? (.getName %) ".edn"))
                     (mapv (fn [f]
                             (let [slug (str/replace (.getName f) #"\.edn$" "")
                                   data (edn/read-string (slurp f))]
                               (assoc data :slug slug))))))]
    (vec (or files []))))
