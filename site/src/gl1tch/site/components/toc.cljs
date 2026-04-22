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
                   :top "68px"}}
     [:div {:style {:font-size "0.6rem"
                    :text-transform :uppercase
                    :letter-spacing "0.15em"
                    :color (t/colors :fg-dim)
                    :margin-bottom "10px"
                    :opacity 0.5}}
      "On this page"]
     [:div {:style {:display :flex
                    :flex-direction :column
                    :gap "3px"
                    :border-left (str "1px solid " (t/colors :border))}}
      (for [{:keys [heading level]} sections]
        ^{:key heading}
        [:a {:href "javascript:void(0)"
             :on-click (scroll-to (heading->id heading))
             :style {:font-size "0.7rem"
                     :color (t/colors :fg-dim)
                     :text-decoration :none
                     :padding-left (str (+ 8 (* (- level 2) 10)) "px")
                     :line-height 1.5
                     :cursor :pointer
                     :opacity 0.7}}
         heading])]]))
