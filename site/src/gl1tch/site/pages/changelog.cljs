(ns gl1tch.site.pages.changelog
  (:require [gl1tch.site.content :as content]
            [gl1tch.site.styles.theme :as t]))

(def type-colors
  {:feat (t/colors :accent)
   :fix  (t/colors :danger)
   :docs (t/colors :accent-2)
   :refactor (t/colors :fg-dim)})

(defn page []
  [:div.wrap {:style {:padding-top "80px"}}
   [:h1 "Changelog"]
   [:div {:style {:margin-top "40px"}}
    (for [entry (content/changelog-entries)]
      ^{:key (:date entry)}
      [:div {:style {:margin-bottom "48px"}}
       [:h2 {:style {:font-size "1.1rem"
                     :color (t/colors :accent)
                     :margin-bottom "16px"}}
        (:date entry)]
       [:div {:style {:display :flex :flex-direction :column :gap "8px"}}
        (for [[i item] (map-indexed vector (:entries entry))]
          ^{:key i}
          [:div {:style {:display :flex :gap "12px" :align-items :baseline}}
           [:span {:style {:font-size (t/sizes :xxs)
                           :font-weight 700
                           :text-transform :uppercase
                           :letter-spacing "0.1em"
                           :color (get type-colors (:type item) (t/colors :fg-dim))
                           :min-width "60px"}}
            (name (:type item))]
           [:span {:style {:font-size (t/sizes :sm)
                           :color (t/colors :fg)}}
            (:summary item)]])]])]])
