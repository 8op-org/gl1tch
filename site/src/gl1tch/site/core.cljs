(ns gl1tch.site.core
  (:require [reagent.dom :as rdom]
            [gl1tch.site.router :as router]
            [gl1tch.site.state :as state]
            [gl1tch.site.styles.global :as styles]
            [gl1tch.site.components.nav :as nav]))

(defn current-page []
  (let [match @state/current-match
        name (get-in match [:data :name])]
    [:div.app
     [nav/nav]
     [:main {:style {:padding-top "52px"}}
      (case name
        :home [:div.wrap [:h1 "gl1tch"]]
        :docs/index [:div.wrap [:h1 "Docs"]]
        :docs/page [:div.wrap [:h1 (str "Doc: " (get-in match [:path-params :slug]))]]
        :labs/index [:div.wrap [:h1 "Labs"]]
        :labs/page [:div.wrap [:h1 (str "Lab: " (get-in match [:path-params :slug]))]]
        :changelog [:div.wrap [:h1 "Changelog"]]
        [:div.wrap [:h1 "404"]])]]))

(defn ^:export init []
  (styles/inject!)
  (router/start!)
  (rdom/render [current-page] (.getElementById js/document "app")))

(defn ^:dev/after-load reload []
  (styles/inject!)
  (rdom/render [current-page] (.getElementById js/document "app")))
