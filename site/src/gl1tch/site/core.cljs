(ns gl1tch.site.core
  (:require [reagent.dom :as rdom]
            [gl1tch.site.router :as router]
            [gl1tch.site.state :as state]
            [gl1tch.site.styles.global :as styles]
            [gl1tch.site.components.nav :as nav]
            [gl1tch.site.pages.home :as home]
            [gl1tch.site.pages.docs :as docs]
            [gl1tch.site.pages.labs :as labs]
            [gl1tch.site.pages.changelog :as changelog]))

(defn current-page []
  (let [match @state/current-match
        name (get-in match [:data :name])
        slug (get-in match [:path-params :slug])]
    [:div.app
     [nav/nav]
     [:main {:style {:padding-top "52px"}}
      (case name
        :home [home/page]
        :docs/index [docs/index]
        :docs/page [docs/page slug]
        :labs/index [labs/index]
        :labs/page [labs/page slug]
        :changelog [changelog/page]
        [:div.wrap [:h1 "404"]])]]))

(defn ^:export init []
  (styles/inject!)
  (router/start!)
  (rdom/render [current-page] (.getElementById js/document "app")))

(defn ^:dev/after-load reload []
  (styles/inject!)
  (rdom/render [current-page] (.getElementById js/document "app")))
