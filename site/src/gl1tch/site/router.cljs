(ns gl1tch.site.router
  (:require [reitit.frontend :as rf]
            [reitit.frontend.easy :as rfe]
            [gl1tch.site.state :as state]))

(def routes
  [["/" {:name :home}]
   ["/docs" {:name :docs/index}]
   ["/docs/:slug" {:name :docs/page}]
   ["/labs" {:name :labs/index}]
   ["/labs/:slug" {:name :labs/page}]
   ["/changelog" {:name :changelog}]])

(def router (rf/router routes))

(defn match [path]
  (let [m (rf/match-by-path router path)]
    (when m
      {:name (get-in m [:data :name])
       :params (:path-params m)})))

(defn start! []
  (rfe/start!
    router
    (fn [match _]
      (reset! state/current-match match))
    {:use-fragment true}))
