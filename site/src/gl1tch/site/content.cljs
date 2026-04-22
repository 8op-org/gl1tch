(ns gl1tch.site.content
  (:require-macros [gl1tch.site.content :refer [load-content]]))

(defn sort-docs [docs]
  (sort-by :order docs))

(defn extract-toc [doc]
  (mapv (fn [s] {:heading (:heading s) :level (:level s)})
        (:sections doc)))

(defn find-by-slug [items slug]
  (first (filter #(= slug (:slug %)) items)))

;; Content index — populated at compile time via macro
(def docs (load-content "docs"))
(def labs (load-content "labs"))
(def changelogs (load-content "changelog"))

(defn docs-index []
  (sort-docs docs))

(defn doc-by-slug [slug]
  (find-by-slug docs slug))

(defn labs-index []
  (sort-by :date #(compare %2 %1) labs))

(defn lab-by-slug [slug]
  (find-by-slug labs slug))

(defn changelog-entries []
  (sort-by :date #(compare %2 %1) changelogs))
