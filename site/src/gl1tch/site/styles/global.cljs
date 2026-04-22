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
     :line-height 1.6
     :overflow-x :hidden
     :-webkit-font-smoothing :antialiased
     :font-size "14px"}]
   [:a {:color (t/colors :accent)
        :text-decoration :none}]
   [:a:hover {:text-decoration :underline}]])

(def code-block
  [[:pre.code
    {:background "#13141f"
     :border [["1px" :solid "rgba(0, 255, 159, 0.08)"]]
     :border-radius "4px"
     :padding "12px 16px"
     :font-family (t/fonts :mono)
     :font-size "0.8rem"
     :line-height 1.7
     :overflow-x :auto
     :margin "12px 0"
     :white-space :pre}]
   [:code
    {:font-family (t/fonts :mono)
     :font-size "0.85em"
     :color (t/colors :accent)
     :background "rgba(0, 255, 159, 0.06)"
     :padding "1px 5px"
     :border-radius "3px"}]
   ["pre.code > code"
    {:background :none :padding 0 :color :inherit :font-size "inherit"}]])

;; Highlight.js syntax colors — matches terminal palette
(def hljs-theme
  [[:.hljs-keyword {:color "#7dcfff"}]
   [:.hljs-built_in {:color "#7dcfff"}]
   [:.hljs-type {:color "#7dcfff"}]
   [:.hljs-string {:color "#e0af68"}]
   [:.hljs-number {:color "#ff9e64"}]
   [:.hljs-literal {:color "#ff9e64"}]
   [:.hljs-comment {:color "#565f89" :font-style :italic}]
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

(def nav-styles
  [[:.site-nav
    {:position :fixed
     :top 0 :left 0 :right 0
     :z-index 100
     :background "rgba(10, 10, 15, 0.9)"
     :backdrop-filter "blur(12px)"
     :-webkit-backdrop-filter "blur(12px)"
     :border-bottom [["1px" :solid (t/colors :border)]]}]])

(def layout
  [[:.wrap {:max-width "900px" :margin "0 auto" :padding "0 40px"
            :position :relative :z-index 1}]
   [:.wrap-wide {:max-width "1120px" :margin "0 auto" :padding "0 40px"
                 :position :relative :z-index 1}]
   [:section {:padding "100px 0" :position :relative}]])

;; ── Doc page typography ──────────────────────────
(def doc-styles
  [;; Layout
   [:.doc-layout
    {:max-width "1200px"
     :margin "0 auto"
     :padding "64px 32px 0"
     :display :grid
     :grid-template-columns "200px 1fr 180px"
     :gap "0 40px"
     :position :relative
     :z-index 1}]
   [:.doc-sidebar-col {:position :relative}]
   [:.doc-toc-col {:position :relative}]

   ;; Article content
   [:.doc-content
    {:min-width 0
     :max-width "640px"
     :color (t/colors :fg)
     :font-size "0.88rem"
     :line-height 1.75}]

   ;; Title
   [:.doc-title
    {:font-size "1.5rem"
     :font-weight 700
     :letter-spacing "-0.02em"
     :margin-bottom "4px"
     :color (t/colors :fg)}]
   [:.doc-desc
    {:color (t/colors :fg-dim)
     :font-size "0.82rem"
     :margin-bottom "28px"
     :padding-bottom "20px"
     :border-bottom [["1px" :solid (t/colors :border)]]}]

   ;; Section headings
   [:.doc-section {:margin-top "28px"}]
   [:.doc-h2
    {:font-size "1.15rem"
     :font-weight 700
     :color (t/colors :accent)
     :margin-bottom "8px"
     :padding-top "4px"}]
   [:.doc-h3
    {:font-size "0.95rem"
     :font-weight 600
     :color (t/colors :accent-2)
     :margin-bottom "6px"
     :margin-top "20px"}]
   [:.doc-h4
    {:font-size "0.88rem"
     :font-weight 600
     :color (t/colors :fg)
     :margin-bottom "4px"
     :margin-top "16px"}]

   ;; Paragraphs
   [:.doc-p
    {:margin-bottom "10px"
     :color (t/colors :fg)
     :opacity 0.85}]

   ;; Links
   [:.doc-link
    {:color (t/colors :accent)
     :text-decoration :none
     :border-bottom [["1px" :solid "transparent"]]}]
   [:.doc-link:hover
    {:border-bottom-color (t/colors :accent)}]

   ;; Lists
   [:.doc-list
    {:padding-left "20px"
     :margin "8px 0"}]
   [:.doc-list>li
    {:margin-bottom "4px"
     :color (t/colors :fg)
     :opacity 0.85}]])

(defn stylesheet []
  (garden/css
    (concat reset base code-block hljs-theme glitch-animation
            nav-styles layout doc-styles)))

(defn inject! []
  (let [style-el (or (.getElementById js/document "gl1tch-styles")
                     (let [el (.createElement js/document "style")]
                       (set! (.-id el) "gl1tch-styles")
                       (.appendChild (.-head js/document) el)
                       el))]
    (set! (.-textContent style-el) (stylesheet))))
