(ns gl1tch.site.pages.labs
  (:require [gl1tch.site.content :as content]
            [gl1tch.site.pages.docs :as docs]
            [gl1tch.site.styles.theme :as t]))

(defn page [slug]
  (if-let [lab (content/lab-by-slug slug)]
    [:div {:style {:max-width "640px"
                   :margin "0 auto"
                   :padding "56px 28px 80px"
                   :position :relative
                   :z-index 1}}
     [:h1 {:style {:font-family (get-in t/fonts [:sans])
                   :font-size "1.6rem"
                   :font-weight 700
                   :letter-spacing "-0.025em"
                   :margin-bottom "8px"
                   :color "#e8e8f0"}}
      (:title lab)]
     [:div {:style {:display :flex :gap "12px" :align-items :center
                    :margin-bottom "28px"
                    :padding-bottom "16px"
                    :border-bottom "1px solid rgba(255,255,255,0.06)"}}
      [:span {:style {:font-family (t/fonts :sans)
                      :font-size "13px"
                      :color "#808098"}}
       (:date lab)]
      (when (:duration lab)
        [:span {:style {:font-family (t/fonts :sans)
                        :font-size "12px"
                        :font-weight 500
                        :color (t/colors :accent)
                        :background "rgba(0, 255, 159, 0.06)"
                        :padding "2px 8px"
                        :border-radius "4px"}}
         (:duration lab)])]
     (for [[i step] (map-indexed vector (:steps lab))]
       ^{:key i}
       [:div {:style {:margin-bottom "28px"}}
        [:div {:style {:display :flex :align-items :baseline :gap "10px"
                       :margin-bottom "8px"}}
         [:span {:style {:font-family (t/fonts :sans)
                         :font-size "11px"
                         :font-weight 600
                         :color (t/colors :accent)
                         :opacity 0.5
                         :letter-spacing "0.08em"
                         :text-transform :uppercase}}
          (str "step " (inc i))]
         [:h2 {:style {:font-family (t/fonts :sans)
                       :font-size "1.05rem"
                       :font-weight 600
                       :color "#b4b4cc"
                       :margin 0}}
          (:heading step)]]
        [:div {:style {:font-size "14.5px"
                       :line-height 1.7
                       :color "#b8b8c8"}}
         (into [:div] (map docs/render-hiccup (:body step)))]])]
    [:div.wrap [:h1 "Lab not found"]]))

(defn index []
  [:div.wrap {:style {:padding-top "56px"}}
   [:h1 {:style {:font-family (t/fonts :sans)
                 :font-size "1.6rem"
                 :font-weight 700
                 :letter-spacing "-0.025em"
                 :margin-bottom "6px"
                 :color "#e8e8f0"}}
    "Labs"]
   [:p {:style {:font-family (t/fonts :sans)
                :font-size "14px"
                :color "#808098"
                :margin-bottom "24px"}}
    "Focused 15-minute walkthroughs."]
   [:div {:style {:display :grid :gap "1px"
                  :background "rgba(255,255,255,0.04)"
                  :border "1px solid rgba(255,255,255,0.06)"
                  :border-radius "8px"
                  :overflow :hidden}}
    (for [lab (content/labs-index)]
      ^{:key (:slug lab)}
      [:a {:href (str "#/labs/" (:slug lab))
           :style {:display :block
                   :background (t/colors :bg-card)
                   :padding "16px 20px"
                   :text-decoration :none
                   :transition "background 0.15s"}}
       [:div {:style {:display :flex :justify-content :space-between
                      :align-items :center}}
        [:div
         [:div {:style {:font-family (t/fonts :sans)
                        :font-weight 600
                        :font-size "14px"
                        :color "#d0d0d8"
                        :margin-bottom "3px"}}
          (:title lab)]
         [:div {:style {:font-family (t/fonts :sans)
                        :font-size "13px"
                        :color "#808098"}}
          (:description lab)]]
        (when (:duration lab)
          [:span {:style {:font-family (t/fonts :sans)
                          :font-size "12px"
                          :color (t/colors :accent)
                          :white-space :nowrap}}
           (:duration lab)])]])]])
