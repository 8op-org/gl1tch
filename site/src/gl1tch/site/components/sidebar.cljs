(ns gl1tch.site.components.sidebar
  (:require [reitit.frontend.easy :as rfe]
            [gl1tch.site.content :as content]
            [gl1tch.site.state :as state]
            [gl1tch.site.styles.theme :as t]))

(defn sidebar []
  (let [match @state/current-match
        current-slug (get-in match [:path-params :slug])
        docs (content/docs-index)]
    [:nav {:style {:position :sticky
                   :top "80px"
                   :max-height "calc(100vh - 100px)"
                   :overflow-y :auto}}
     [:div {:style {:display :flex
                    :flex-direction :column
                    :gap "2px"}}
      (for [doc docs]
        ^{:key (:slug doc)}
        [:a {:href (rfe/href :docs/page {:slug (:slug doc)})
             :style {:font-size (t/sizes :xs)
                     :color (if (= current-slug (:slug doc))
                              (t/colors :accent)
                              (t/colors :fg-dim))
                     :text-decoration :none
                     :padding "5px 12px"
                     :border-radius t/radius
                     :border-left (str "2px solid "
                                       (if (= current-slug (:slug doc))
                                         (t/colors :accent)
                                         "transparent"))}}
         (:title doc)])]]))
