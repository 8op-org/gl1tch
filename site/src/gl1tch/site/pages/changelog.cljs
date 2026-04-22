(ns gl1tch.site.pages.changelog
  (:require [gl1tch.site.content :as content]
            [gl1tch.site.styles.theme :as t]))

(def type-colors
  {:feat (t/colors :accent)
   :fix  "#ff6b6b"
   :docs (t/colors :accent-2)
   :refactor "#808098"})

(defn page []
  [:div {:style {:max-width "640px"
                 :margin "0 auto"
                 :padding "56px 28px 80px"
                 :position :relative
                 :z-index 1}}
   [:h1 {:style {:font-family (t/fonts :sans)
                 :font-size "1.6rem"
                 :font-weight 700
                 :letter-spacing "-0.025em"
                 :margin-bottom "28px"
                 :color "#e8e8f0"}}
    "Changelog"]
   (for [entry (content/changelog-entries)]
     ^{:key (:date entry)}
     [:div {:style {:margin-bottom "32px"}}
      [:h2 {:style {:font-family (t/fonts :mono)
                    :font-size "14px"
                    :font-weight 600
                    :color (t/colors :accent)
                    :margin-bottom "12px"
                    :padding-bottom "8px"
                    :border-bottom "1px solid rgba(255,255,255,0.06)"}}
       (:date entry)]
      [:div {:style {:display :flex :flex-direction :column :gap "6px"}}
       (for [[i item] (map-indexed vector (:entries entry))]
         ^{:key i}
         [:div {:style {:display :flex :gap "10px" :align-items :baseline}}
          [:span {:style {:font-family (t/fonts :mono)
                          :font-size "11px"
                          :font-weight 600
                          :text-transform :uppercase
                          :letter-spacing "0.05em"
                          :color (get type-colors (:type item) "#808098")
                          :min-width "52px"}}
           (name (:type item))]
          [:span {:style {:font-family (t/fonts :sans)
                          :font-size "13.5px"
                          :color "#b8b8c8"
                          :line-height 1.5}}
           (:summary item)]])]])])
