(ns gl1tch.site.components.toc
  (:require [gl1tch.site.styles.theme :as t]
            [clojure.string :as str]))

(defn- heading->id [heading]
  (-> heading str/lower-case (str/replace #"[^a-z0-9]+" "-") (str/replace #"^-|-$" "")))

(defn- scroll-to [id]
  (fn [e]
    (.preventDefault e)
    (when-let [el (.getElementById js/document id)]
      (.scrollIntoView el #js {:behavior "smooth" :block "start"}))))

(defn toc [sections]
  (when (seq sections)
    [:nav {:style {:position :sticky
                   :top "64px"}}
     [:div {:style {:font-family (t/fonts :sans)
                    :font-size "11px"
                    :font-weight 600
                    :text-transform :uppercase
                    :letter-spacing "0.08em"
                    :color "#505068"
                    :margin-bottom "10px"}}
      "On this page"]
     [:div {:style {:display :flex
                    :flex-direction :column
                    :gap "2px"
                    :border-left "1px solid rgba(255,255,255,0.06)"}}
      (for [{:keys [heading level]} sections]
        ^{:key heading}
        [:a {:href "javascript:void(0)"
             :on-click (scroll-to (heading->id heading))
             :style {:font-family (t/fonts :sans)
                     :font-size "12px"
                     :color "#707088"
                     :text-decoration :none
                     :padding-left (str (+ 10 (* (- level 2) 10)) "px")
                     :line-height 1.5
                     :cursor :pointer
                     :transition "color 0.15s"}}
         heading])]]))
