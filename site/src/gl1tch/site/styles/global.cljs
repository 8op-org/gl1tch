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
    {:font-family (t/fonts :sans)
     :background (t/colors :bg)
     :color (t/colors :fg)
     :line-height 1.6
     :overflow-x :hidden
     :-webkit-font-smoothing :antialiased
     :font-size "15px"}]
   [:a {:color (t/colors :accent)
        :text-decoration :none}]
   [:a:hover {:opacity 0.8}]])

;; ── Code ────────────────────────────────────────
(def code-block
  [[:pre.code
    {:background "#11121b"
     :border [["1px" :solid "rgba(255,255,255,0.06)"]]
     :border-radius "6px"
     :padding "14px 18px"
     :font-family (t/fonts :mono)
     :font-size "13px"
     :line-height 1.7
     :overflow-x :auto
     :margin "10px 0"
     :white-space :pre}]
   [:code
    {:font-family (t/fonts :mono)
     :font-size "0.87em"
     :color "#9dd6fc"
     :background "rgba(157, 214, 252, 0.07)"
     :padding "2px 6px"
     :border-radius "4px"
     :letter-spacing "-0.01em"}]
   ["pre.code > code"
    {:background :none :padding 0 :color :inherit :font-size "inherit"
     :letter-spacing 0}]])

;; ── Highlight.js ────────────────────────────────
(def hljs-theme
  [[:.hljs-keyword {:color "#7dcfff"}]
   [:.hljs-built_in {:color "#7dcfff"}]
   [:.hljs-type {:color "#7dcfff"}]
   [:.hljs-string {:color "#e0af68"}]
   [:.hljs-number {:color "#ff9e64"}]
   [:.hljs-literal {:color "#ff9e64"}]
   [:.hljs-comment {:color "#4a4e69" :font-style :italic}]
   [:.hljs-title {:color "#7aa2f7"}]
   [:.hljs-function {:color "#7aa2f7"}]
   [:.hljs-name {:color "#7dcfff"}]
   [:.hljs-variable {:color "#bcc4e0"}]
   [:.hljs-attr {:color "#bb9af7"}]
   [:.hljs-params {:color "#bcc4e0"}]
   [:.hljs-meta {:color "#4a4e69"}]
   [:.hljs-symbol {:color "#00ff9f"}]
   [:.hljs-selector-tag {:color "#7dcfff"}]
   [:.hljs-selector-class {:color "#bb9af7"}]])

;; ── Animations ──────────────────────────────────
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

;; ── Nav ─────────────────────────────────────────
(def nav-styles
  [[:.site-nav
    {:position :fixed
     :top 0 :left 0 :right 0
     :z-index 100
     :background "rgba(10, 10, 15, 0.92)"
     :backdrop-filter "blur(16px)"
     :-webkit-backdrop-filter "blur(16px)"
     :border-bottom [["1px" :solid "rgba(255,255,255,0.05)"]]}]])

;; ── Layout ──────────────────────────────────────
(def layout
  [[:.wrap {:max-width "900px" :margin "0 auto" :padding "0 40px"
            :position :relative :z-index 1}]
   [:.wrap-wide {:max-width "1120px" :margin "0 auto" :padding "0 40px"
                 :position :relative :z-index 1}]
   [:section {:padding "80px 0" :position :relative}]])

;; ── Doc typography ──────────────────────────────
(def doc-styles
  [;; 3-column layout
   [:.doc-layout
    {:max-width "1100px"
     :margin "0 auto"
     :padding "56px 28px 80px"
     :display :grid
     :grid-template-columns "180px minmax(0, 1fr) 160px"
     :gap "0 36px"
     :position :relative
     :z-index 1}]
   [:.doc-sidebar-col {:position :relative}]
   [:.doc-toc-col {:position :relative}]

   ;; Content area
   [:.doc-content
    {:min-width 0
     :max-width "600px"
     :color "#d0d0d8"
     :font-size "14.5px"
     :line-height 1.7
     :letter-spacing "0.01em"}]

   ;; Page title
   [:.doc-title
    {:font-family (t/fonts :sans)
     :font-size "1.6rem"
     :font-weight 700
     :letter-spacing "-0.025em"
     :margin-bottom "6px"
     :color "#e8e8f0"}]
   [:.doc-desc
    {:color "#808098"
     :font-size "14px"
     :line-height 1.5
     :margin-bottom "24px"
     :padding-bottom "16px"
     :border-bottom [["1px" :solid "rgba(255,255,255,0.06)"]]}]

   ;; Section spacing
   [:.doc-section {:margin-top "28px"}]
   [:.doc-section:first-child {:margin-top 0}]

   ;; h2 — primary section header
   [:.doc-h2
    {:font-family (t/fonts :sans)
     :font-size "1.1rem"
     :font-weight 600
     :color (t/colors :accent)
     :margin-bottom "10px"
     :padding-bottom "8px"
     :border-bottom [["1px" :solid "rgba(255,255,255,0.06)"]]}]

   ;; h3 — subsection
   [:.doc-h3
    {:font-family (t/fonts :sans)
     :font-size "0.95rem"
     :font-weight 600
     :color "#b4b4cc"
     :margin-bottom "6px"
     :margin-top "22px"}]

   ;; h4
   [:.doc-h4
    {:font-family (t/fonts :sans)
     :font-size "0.88rem"
     :font-weight 600
     :color "#9898b0"
     :margin-bottom "4px"
     :margin-top "16px"}]

   ;; Paragraphs — the core reading experience
   [:.doc-p
    {:margin-bottom "12px"
     :color "#b8b8c8"
     :line-height 1.75}]

   ;; Links
   [:.doc-link
    {:color (t/colors :accent)
     :text-decoration :none
     :font-weight 500}]
   [:.doc-link:hover {:opacity 0.8}]

   ;; Lists — tighter than paragraphs
   [:.doc-list
    {:padding-left "18px"
     :margin "6px 0 12px"}]
   [:.doc-list>li
    {:margin-bottom "6px"
     :color "#b8b8c8"
     :line-height 1.65}]
   ["li > code"
    {:font-size "0.85em"}]])

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
