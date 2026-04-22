(ns gl1tch.site.diagrams.svg
  (:require [gl1tch.site.styles.theme :as t]))

(defn box [{:keys [x y w h label style]}]
  (let [fill (or (:fill style) (t/colors :bg-alt))
        stroke (or (:stroke style) (t/colors :accent))
        text-fill (or (:color style) (t/colors :fg))]
    [:g
     [:rect {:x x :y y :width w :height h
             :rx 4 :ry 4
             :fill fill
             :stroke stroke
             :stroke-width 1.5}]
     [:text {:x (+ x (/ w 2))
             :y (+ y (/ h 2) 5)
             :text-anchor "middle"
             :fill text-fill
             :font-family (t/fonts :mono)
             :font-size 12}
      label]]))

(defn arrow [{:keys [path style]}]
  (let [stroke (or (:stroke style) (t/colors :fg-dim))
        points (apply str (interpose " " (map (fn [[px py]] (str px "," py)) path)))]
    [:polyline {:points points
                :fill "none"
                :stroke stroke
                :stroke-width 1.5
                :marker-end "url(#arrowhead)"}]))

(defn arrowhead-marker []
  [:defs
   [:marker {:id "arrowhead"
             :markerWidth 10 :markerHeight 7
             :refX 10 :refY 3.5
             :orient "auto"}
    [:polygon {:points "0 0, 10 3.5, 0 7"
               :fill (t/colors :fg-dim)}]]])

(defn label [{:keys [x y text style]}]
  [:text {:x x :y y
          :fill (or (:color style) (t/colors :fg-dim))
          :font-family (t/fonts :mono)
          :font-size 11}
   text])
