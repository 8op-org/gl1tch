(ns gl1tch.site.diagrams.flowchart
  (:require [gl1tch.site.diagrams.core :as layout]
            [gl1tch.site.diagrams.svg :as svg]))

(defn parse-inline
  "Parses inline diagram children from content EDN into a flowchart spec.
   Input attrs: {:type :flowchart :direction :lr}
   Input children: [[:node :a \"Start\"] [:node :b \"End\"] [:edge :a :b]]
   Returns: {:direction :lr :nodes [...] :edges [...]}"
  [attrs children]
  (let [nodes (filterv #(= :node (first %)) children)
        edges (filterv #(= :edge (first %)) children)]
    {:direction (or (:direction attrs) :lr)
     :nodes (mapv (fn [[_ id label]] {:id id :label label}) nodes)
     :edges (mapv (fn [[_ from to]] [from to]) edges)}))

(defn render
  "Renders a flowchart spec as hiccup SVG."
  [spec]
  (let [{:keys [nodes edges width height]} (layout/compute spec)]
    [:svg {:width width :height height
           :viewBox (str "0 0 " width " " height)
           :xmlns "http://www.w3.org/2000/svg"
           :style {:display :block :margin "16px 0"}}
     [svg/arrowhead-marker]
     (for [node nodes]
       ^{:key (:id node)} [svg/box node])
     (for [edge edges]
       ^{:key (str (:from edge) "-" (:to edge))} [svg/arrow edge])]))
