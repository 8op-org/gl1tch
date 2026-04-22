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
                   :top "68px"
                   :max-height "calc(100vh - 80px)"
                   :overflow-y :auto
                   :scrollbar-width :none}}
     [:div {:style {:display :flex
                    :flex-direction :column
                    :gap 0}}
      (for [doc docs]
        ^{:key (:slug doc)}
        [:a {:href (rfe/href :docs/page {:slug (:slug doc)})
             :style {:font-size "0.75rem"
                     :color (if (= current-slug (:slug doc))
                              (t/colors :accent)
                              (t/colors :fg-dim))
                     :text-decoration :none
                     :padding "4px 10px"
                     :border-left (str "2px solid "
                                       (if (= current-slug (:slug doc))
                                         (t/colors :accent)
                                         "transparent"))
                     :background (when (= current-slug (:slug doc))
                                   "rgba(0, 255, 159, 0.04)")
                     :line-height 1.5}}
         (:title doc)])]]))
