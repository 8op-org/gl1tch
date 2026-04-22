(ns gl1tch.site.pages.home
  (:require [gl1tch.site.components.hex-rain :as hex-rain]
            [gl1tch.site.styles.theme :as t]))

(defn- code-span [cls text]
  [:span {:style {:color (case cls
                           "k" "#7dcfff"
                           "s" "#e0af68"
                           "r" "#7a88a8"
                           "p" "#7dcfff"
                           (t/colors :fg))}}
   text])

(defn hero []
  [:section {:style {:min-height "100vh"
                     :display :flex
                     :align-items :center
                     :padding "120px 0 80px"}}
   [:div {:style {:max-width "1120px"
                  :margin "0 auto"
                  :padding "0 40px"
                  :display :grid
                  :grid-template-columns "1fr 1.2fr"
                  :gap "60px"
                  :align-items :center
                  :position :relative
                  :z-index 1}}
    [:div
     [:h1 {:style {:font-size (t/sizes :h1)
                   :font-weight 700
                   :letter-spacing "-0.04em"
                   :line-height 1.05
                   :margin-bottom "32px"}}
      "gl1tch"]
     [:p {:style {:font-size "1.15rem"
                  :opacity 0.7
                  :line-height 2
                  :margin-bottom "40px"}}
      "Shell does the work." [:br]
      "LLM does the thinking." [:br]
      "You own the workflow."]
     [:pre.code {:style {:margin-bottom "16px" :max-width "460px"}}
      [code-span "p" "$ "] "brew install 8op-org/tap/glitch"]
     [:div {:style {:display :flex :gap "12px"}}
      [:a {:href "https://github.com/8op-org/gl1tch"
           :style {:display :inline-block
                   :padding "12px 28px"
                   :background (t/colors :accent)
                   :color (t/colors :bg)
                   :border-radius "4px"
                   :font-weight 600
                   :font-size "0.8rem"
                   :text-decoration :none}}
       "GitHub"]
      [:a {:href "#/docs/getting-started"
           :style {:display :inline-block
                   :padding "12px 28px"
                   :border (str "1px solid " (t/colors :border))
                   :color (t/colors :fg-dim)
                   :border-radius "4px"
                   :font-size "0.8rem"
                   :text-decoration :none}}
       "Docs"]]]
    [:pre.code {:style {:font-size "0.78rem"
                        :line-height 2
                        :margin-bottom 0}}
     [code-span "r" ";; agentic mode -- LLM pulls its own context"] "\n\n"
     "(" [code-span "k" "workflow"] " " [code-span "s" "\"pr-review\""] "\n\n"
     "  (" [code-span "k" "step"] " " [code-span "s" "\"fetch\""] "\n"
     "    (" [code-span "k" "sh"] " " [code-span "s" "\"gh pr view ~input --json title,body,files\""] "))\n\n"
     "  (" [code-span "k" "step"] " " [code-span "s" "\"review\""] "\n"
     "    (" [code-span "k" "llm"] " " [code-span "r" ":agentic"] " " [code-span "s" "true"] " "
     [code-span "r" ":max-rounds"] " " [code-span "s" "10"] "\n"
     "      " [code-span "r" ":prompt"] " " [code-span "s" "```\n        PR: ~(ref \"fetch\")\n        Read the changed files.\n        Review as a senior engineer.\n        ```"] ")))"]]])

(defn page []
  [:div
   [hex-rain/hex-rain]
   [hero]])
