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
                   :top "64px"
                   :max-height "calc(100vh - 80px)"
                   :overflow-y :auto
                   :scrollbar-width :none
                   :padding-right "8px"}}
     [:div {:style {:display :flex
                    :flex-direction :column
                    :gap "1px"}}
      (for [doc docs]
        (let [active? (= current-slug (:slug doc))]
          ^{:key (:slug doc)}
          [:a {:href (rfe/href :docs/page {:slug (:slug doc)})
               :style {:font-family (t/fonts :sans)
                       :font-size "13px"
                       :font-weight (if active? 500 400)
                       :color (if active? (t/colors :accent) "#808098")
                       :text-decoration :none
                       :padding "5px 10px"
                       :border-left (str "2px solid "
                                         (if active? (t/colors :accent) "transparent"))
                       :background (when active? "rgba(0, 255, 159, 0.04)")
                       :border-radius "0 4px 4px 0"
                       :line-height 1.4
                       :transition "color 0.15s, background 0.15s"}}
           (:title doc)]))]]))
