(ns gl1tch.site.pages.labs
  (:require [gl1tch.site.content :as content]
            [gl1tch.site.pages.docs :as docs]
            [gl1tch.site.styles.theme :as t]))

(defn page [slug]
  (if-let [lab (content/lab-by-slug slug)]
      [:div.wrap {:style {:padding-top "80px" :max-width "720px"}}
       [:h1 {:style {:margin-bottom "8px"}} (:title lab)]
       [:div {:style {:display :flex :gap "16px" :margin-bottom "40px"
                      :color (t/colors :fg-dim) :font-size (t/sizes :sm)}}
        [:span (:date lab)]
        (when (:duration lab)
          [:span {:style {:color (t/colors :accent)
                          :background "rgba(0, 255, 159, 0.06)"
                          :padding "2px 8px"
                          :border-radius t/radius}}
           (:duration lab)])]
       (for [[i step] (map-indexed vector (:steps lab))]
         ^{:key i}
         [:section {:style {:margin-bottom "48px"}}
          [:div {:style {:display :flex :align-items :baseline :gap "12px"
                         :margin-bottom "16px"}}
           [:span {:style {:font-size (t/sizes :xxs)
                           :color (t/colors :accent)
                           :opacity 0.5
                           :letter-spacing "0.15em"}}
            (str "step " (inc i))]
           [:h2 {:style {:font-size (t/sizes :h3)
                         :color (t/colors :accent-2)
                         :margin 0}}
            (:heading step)]]
          (into [:div] (map docs/render-hiccup (:body step)))])]
      [:div.wrap [:h1 "Lab not found"]]))

(defn index []
  [:div.wrap {:style {:padding-top "80px"}}
   [:h1 "Labs"]
   [:p {:style {:color (t/colors :fg-dim)
                :margin-bottom "40px"}}
    "Focused 15-minute walkthroughs."]
   [:div {:style {:display :grid :gap "1px"
                  :background (t/colors :border)
                  :border (str "1px solid " (t/colors :border))
                  :border-radius "8px"
                  :overflow :hidden}}
    (for [lab (content/labs-index)]
      ^{:key (:slug lab)}
      [:a {:href (str "#/labs/" (:slug lab))
           :style {:display :block
                   :background (t/colors :bg-card)
                   :padding "24px"
                   :text-decoration :none}}
       [:div {:style {:display :flex :justify-content :space-between
                      :align-items :center}}
        [:div
         [:div {:style {:font-weight 700
                        :color (t/colors :fg)
                        :margin-bottom "6px"}}
          (:title lab)]
         [:div {:style {:font-size (t/sizes :sm)
                        :color (t/colors :fg-dim)}}
          (:description lab)]]
        (when (:duration lab)
          [:span {:style {:font-size (t/sizes :xs)
                          :color (t/colors :accent)
                          :white-space :nowrap}}
           (:duration lab)])]])]])
