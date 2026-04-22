(ns gl1tch.site.components.nav
  (:require [reitit.frontend.easy :as rfe]
            [gl1tch.site.styles.theme :as t]))

(defn nav-link [route-name label]
  [:a {:href (rfe/href route-name)
       :style {:font-size (t/sizes :xs)
               :color (t/colors :fg-dim)
               :text-decoration :none
               :letter-spacing "0.04em"}}
   label])

(defn nav []
  [:nav.site-nav
   [:div {:style {:max-width "1120px"
                  :margin "0 auto"
                  :padding "0 40px"
                  :height "52px"
                  :display :flex
                  :align-items :center
                  :justify-content :space-between}}
    [:a {:href (rfe/href :home)
         :style {:font-size "0.9rem"
                 :font-weight 700
                 :color (t/colors :fg)
                 :text-decoration :none
                 :letter-spacing "-0.02em"}}
     "gl1tch"]
    [:div {:style {:display :flex :gap "24px"}}
     [nav-link :docs/index "Docs"]
     [nav-link :labs/index "Labs"]
     [nav-link :changelog "Changelog"]
     [:a {:href "https://github.com/8op-org/gl1tch"
          :style {:font-size (t/sizes :xs)
                  :color (t/colors :fg-dim)
                  :text-decoration :none
                  :letter-spacing "0.04em"}}
      "GitHub"]]]])
