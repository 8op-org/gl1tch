(ns gl1tch.site.styles.global
  (:require [garden.core :as garden]
            [garden.stylesheet :refer [at-keyframes]]
            [gl1tch.site.styles.theme :as t]))

(def reset
  [[:* :*::before :*::after
    {:margin 0 :padding 0 :box-sizing :border-box}]])

(def base
  [[:html {:scroll-behavior :smooth}]
   [:body
    {:font-family (t/fonts :mono)
     :background (t/colors :bg)
     :color (t/colors :fg)
     :line-height 1.7
     :overflow-x :hidden
     :-webkit-font-smoothing :antialiased
     :font-size (t/sizes :base)}]
   [:a {:color (t/colors :accent)
        :text-decoration :none}]
   [:a:hover {:text-decoration :underline}]])

(def code-block
  [[:pre.code
    {:background "#1a1b2e"
     :border [["1px" :solid "rgba(0, 255, 159, 0.1)"]]
     :border-radius "6px"
     :padding "16px 20px"
     :font-family (t/fonts :mono)
     :font-size (t/sizes :sm)
     :line-height 1.9
     :overflow-x :auto
     :margin-bottom "16px"
     :white-space :pre}]
   [:code
    {:font-family (t/fonts :mono)
     :font-size "0.88em"
     :color (t/colors :accent)
     :background "rgba(0, 255, 159, 0.06)"
     :padding "2px 6px"
     :border-radius t/radius}]
   ["pre.code > code"
    {:background :none :padding 0 :color :inherit}]])

(def glitch-animation
  [(at-keyframes :glitch-1
     [:0% {:clip-path "inset(40% 0 61% 0)" :transform "translate(-2px, -1px)"}]
     [:20% {:clip-path "inset(92% 0 1% 0)" :transform "translate(1px, 2px)"}]
     [:40% {:clip-path "inset(43% 0 1% 0)" :transform "translate(-1px, 1px)"}]
     [:60% {:clip-path "inset(25% 0 58% 0)" :transform "translate(2px, -1px)"}]
     [:80% {:clip-path "inset(54% 0 7% 0)" :transform "translate(-1px, 2px)"}]
     [:100% {:clip-path "inset(58% 0 43% 0)" :transform "translate(0px, -2px)"}])
   (at-keyframes :glitch-2
     [:0% {:clip-path "inset(65% 0 13% 0)" :transform "translate(2px, 1px)"}]
     [:20% {:clip-path "inset(15% 0 62% 0)" :transform "translate(-2px, -1px)"}]
     [:40% {:clip-path "inset(79% 0 2% 0)" :transform "translate(1px, 1px)"}]
     [:60% {:clip-path "inset(2% 0 78% 0)" :transform "translate(-1px, 2px)"}]
     [:80% {:clip-path "inset(40% 0 34% 0)" :transform "translate(2px, -1px)"}]
     [:100% {:clip-path "inset(67% 0 5% 0)" :transform "translate(-2px, 0px)"}])])

;; Highlight.js syntax colors — matches our terminal palette
(def hljs-theme
  [[:.hljs-keyword {:color "#7dcfff"}]
   [:.hljs-built_in {:color "#7dcfff"}]
   [:.hljs-type {:color "#7dcfff"}]
   [:.hljs-string {:color "#e0af68"}]
   [:.hljs-number {:color "#ff9e64"}]
   [:.hljs-literal {:color "#ff9e64"}]
   [:.hljs-comment {:color "#565f89"}]
   [:.hljs-title {:color "#7aa2f7"}]
   [:.hljs-function {:color "#7aa2f7"}]
   [:.hljs-name {:color "#7dcfff"}]
   [:.hljs-variable {:color "#c0caf5"}]
   [:.hljs-attr {:color "#bb9af7"}]
   [:.hljs-params {:color "#c0caf5"}]
   [:.hljs-meta {:color "#565f89"}]
   [:.hljs-symbol {:color "#00ff9f"}]
   [:.hljs-selector-tag {:color "#7dcfff"}]
   [:.hljs-selector-class {:color "#bb9af7"}]])

(def nav-styles
  [[:.site-nav
    {:position :fixed
     :top 0 :left 0 :right 0
     :z-index 100
     :background "rgba(10, 10, 15, 0.85)"
     :backdrop-filter "blur(12px)"
     :-webkit-backdrop-filter "blur(12px)"
     :border-bottom [["1px" :solid (t/colors :border)]]}]])

(def layout
  [[:.wrap {:max-width "900px" :margin "0 auto" :padding "0 40px"
            :position :relative :z-index 1}]
   [:.wrap-wide {:max-width "1120px" :margin "0 auto" :padding "0 40px"
                 :position :relative :z-index 1}]
   [:section {:padding "120px 0" :position :relative}]])

(defn stylesheet []
  (garden/css
    (concat reset base code-block hljs-theme glitch-animation nav-styles layout)))

(defn inject! []
  (let [style-el (or (.getElementById js/document "gl1tch-styles")
                     (let [el (.createElement js/document "style")]
                       (set! (.-id el) "gl1tch-styles")
                       (.appendChild (.-head js/document) el)
                       el))]
    (set! (.-textContent style-el) (stylesheet))))
