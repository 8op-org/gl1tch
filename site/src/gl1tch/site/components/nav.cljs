(ns gl1tch.site.components.nav
  (:require [reitit.frontend.easy :as rfe]
            [gl1tch.site.styles.theme :as t]))

(defn- nav-link [route-name label]
  [:a {:href (rfe/href route-name)
       :style {:font-family (t/fonts :sans)
               :font-size "13px"
               :font-weight 500
               :color "#808098"
               :text-decoration :none
               :transition "color 0.15s"}}
   label])

(defn nav []
  [:nav.site-nav
   [:div {:style {:max-width "1100px"
                  :margin "0 auto"
                  :padding "0 28px"
                  :height "48px"
                  :display :flex
                  :align-items :center
                  :justify-content :space-between}}
    [:a {:href (rfe/href :home)
         :style {:font-family (t/fonts :mono)
                 :font-size "14px"
                 :font-weight 700
                 :color "#e8e8f0"
                 :text-decoration :none
                 :letter-spacing "-0.02em"}}
     "gl1tch"]
    [:div {:style {:display :flex :gap "20px"}}
     [nav-link :docs/index "Docs"]
     [nav-link :labs/index "Labs"]
     [nav-link :changelog "Changelog"]
     [:a {:href "https://github.com/8op-org/gl1tch"
          :style {:font-family (t/fonts :sans)
                  :font-size "13px"
                  :font-weight 500
                  :color "#808098"
                  :text-decoration :none}}
      "GitHub"]]]])
