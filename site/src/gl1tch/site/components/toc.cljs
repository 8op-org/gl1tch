(ns gl1tch.site.components.toc
  (:require [gl1tch.site.styles.theme :as t]
            [clojure.string :as str]))

(defn- heading->id [heading]
  (-> heading str/lower-case (str/replace #"[^a-z0-9]+" "-") (str/replace #"^-|-$" "")))

(defn toc [sections]
  (when (seq sections)
    [:nav {:style {:position :sticky
                   :top "80px"}}
     [:div {:style {:font-size (t/sizes :xxs)
                    :text-transform :uppercase
                    :letter-spacing "0.2em"
                    :color (t/colors :accent)
                    :margin-bottom "16px"
                    :opacity 0.6}}
      "On this page"]
     [:div {:style {:display :flex
                    :flex-direction :column
                    :gap "8px"}}
      (for [{:keys [heading level]} sections]
        ^{:key heading}
        [:a {:href (str "#" (heading->id heading))
             :style {:font-size (t/sizes :xs)
                     :color (t/colors :fg-dim)
                     :text-decoration :none
                     :padding-left (str (* (- level 2) 12) "px")
                     :border-left "1px solid transparent"}}
         heading])]]))
