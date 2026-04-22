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

(defn- feature-code [& children]
  (into [:pre.code {:style {:margin-bottom 0
                            :font-size "0.75rem"
                            :line-height 1.8
                            :padding "16px 20px"
                            :background "rgba(10, 10, 15, 0.6)"
                            :border-color "rgba(0, 255, 159, 0.15)"}}]
        children))

(defn- feature-card [{:keys [num title desc hero?]} & code-children]
  [:div {:style (merge {:background (t/colors :bg-card)
                        :padding "40px 36px"
                        :border-left "2px solid transparent"}
                       (when hero? {:grid-column "span 2"
                                    :border-left-color (t/colors :accent)}))}
   [:div {:style {:font-size "0.6rem" :color (t/colors :accent)
                  :letter-spacing "0.15em" :margin-bottom "14px" :opacity 0.5}}
    num]
   [:h3 {:style {:font-size "1.05rem" :font-weight 700
                 :color (t/colors :fg) :margin-bottom "12px"}}
    title]
   [:p {:style {:font-size "0.85rem" :color (t/colors :fg-dim)
                :line-height 1.85 :margin-bottom "20px" :max-width "420px"}}
    desc]
   (into [feature-code] code-children)])

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

(defn features []
  [:section {:style {:background (t/colors :bg-alt)
                     :border-top (str "1px solid " (t/colors :border))
                     :border-bottom (str "1px solid " (t/colors :border))
                     :padding "100px 0"}}
   [:div.wrap-wide
    [:div {:style {:font-size "0.65rem" :text-transform :uppercase
                   :letter-spacing "0.2em" :color (t/colors :accent)
                   :margin-bottom "12px"}}
     "what you get"]
    [:div {:style {:display :grid
                   :grid-template-columns "repeat(2, 1fr)"
                   :gap "1px"
                   :background (t/colors :border)
                   :border (str "1px solid " (t/colors :border))
                   :border-radius "8px"
                   :overflow :hidden}}
     [feature-card {:num "01" :title "Agentic LLM steps" :hero? true
                    :desc "Set :agentic true and the LLM loops — calling tools, reading files, grepping code — until it has everything it needs. No more pre-stuffing prompts with context."}
      "(" [code-span "k" "llm"] " " [code-span "r" ":agentic"] " " [code-span "s" "true"] " "
      [code-span "r" ":max-rounds"] " " [code-span "s" "10"] "\n"
      "  " [code-span "r" ":prompt"] " " [code-span "s" "\"Rewrite the homepage\n  features section...\""] ")"]
     [feature-card {:num "02" :title "4 providers, one interface"
                    :desc "Switch between LM Studio, Copilot, OpenRouter, and Claude with a keyword. Each provider gets the same tool-use interface."}
      [code-span "r" ";; same tools, different provider"] "\n"
      "(" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"copilot\""] "\n"
      "  " [code-span "r" ":prompt"] " " [code-span "s" "\"Review this PR...\""] ")"]
     [feature-card {:num "03" :title "Tiered escalation"
                    :desc "Start on your local model for free. gl1tch self-evaluates the output. Escalates to cloud only when quality demands it."}
      [code-span "r" ";; tier 0: lm-studio (free)"] "\n"
      [code-span "r" ";; tier 1: copilot"] "\n"
      [code-span "r" ";; tier 2: claude"] "\n"
      "(" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"lmstudio\""] " "
      [code-span "r" ":prompt"] " " [code-span "s" "\"...\""] ")"]
     [feature-card {:num "04" :title "S-expressions, not YAML"
                    :desc "Workflows are parenthesized lists. Every construct composes — retry wraps timeout wraps step. No indentation wars."}
      "(" [code-span "k" "retry"] " " [code-span "s" "2"] "\n"
      "  (" [code-span "k" "with-timeout"] " " [code-span "s" "30"] "\n"
      "    (" [code-span "k" "step"] " " [code-span "s" "\"fetch\""] "\n"
      "      (" [code-span "k" "sh"] " " [code-span "s" "\"curl -sf ...\""] "))))"]
     [feature-card {:num "05" :title "Confidence framework"
                    :desc "Verify LLM output is grounded in your codebase. Schema validation, confidence scoring with retries, and multi-provider consensus."}
      [code-span "r" ";; verify output against source"] "\n"
      "(" [code-span "k" "grounded?"] " " [code-span "s" "\"rewrite\""] " (" [code-span "k" "ref"] " " [code-span "s" "\"context\""] ")\n"
      "  " [code-span "r" ":provider"] " " [code-span "s" "\"openrouter\""] ")"]
     [feature-card {:num "06" :title "Built-in A/B testing"
                    :desc "Run the same prompt through different models or strategies. A neutral judge scores each branch. Automatic winner selection."}
      "(" [code-span "k" "compare"] "\n"
      "  (" [code-span "k" "branch"] " " [code-span "s" "\"local\""] "\n"
      "    (" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"lmstudio\""] " " [code-span "r" ":prompt"] " " [code-span "s" "\"...\""] "))\n"
      "  (" [code-span "k" "branch"] " " [code-span "s" "\"cloud\""] "\n"
      "    (" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"claude\""] " " [code-span "r" ":prompt"] " " [code-span "s" "\"...\""] "))\n"
      "  (" [code-span "k" "review"] " " [code-span "r" ":criteria"] " (" [code-span "s" "\"accuracy\""] " " [code-span "s" "\"clarity\""] ")))"]
     [feature-card {:num "07" :title "Plugins are directories"
                    :desc "A plugin is a folder of .glitch files. Each file is a subcommand. Args become flags. No compilation. No release pipeline."}
      [code-span "p" "$ "] "glitch plugin github prs --since week\n"
      [code-span "r" "[\"Fix flaky test\", \"Add retry logic\"]"]]]]])

(defn how-it-works []
  [:section {:style {:border-top (str "1px solid " (t/colors :border))
                     :padding "120px 0"}}
   [:div.wrap-wide
    [:div {:style {:font-size "0.65rem" :text-transform :uppercase
                   :letter-spacing "0.2em" :color (t/colors :accent)
                   :margin-bottom "12px"}}
     "how it works"]
    [:h2 {:style {:font-size (t/sizes :h2) :margin-bottom "56px"}}
     "The LLM pulls its own context. You write the intent."]
    [:div {:style {:display :grid :grid-template-columns "repeat(2, 1fr)" :gap "32px"}}
     (for [{:keys [num title desc code]}
           [{:num "01" :title "shell gathers the seed"
             :desc "Shell steps fetch what you already know you need — a PR number, a commit SHA, a file list. Fast and deterministic."
             :code ["(" [code-span "k" "step"] " " [code-span "s" "\"pr\""] "\n"
                    "  (" [code-span "k" "sh"] " " [code-span "s" "\"gh pr view ~input --json title,body,files\""] "))"]}
            {:num "02" :title "LLM explores on demand"
             :desc "The LLM reads files, greps for usages, and checks symbols — looping until it has enough context."
             :code ["(" [code-span "k" "step"] " " [code-span "s" "\"review\""] "\n"
                    "  (" [code-span "k" "llm"] " " [code-span "r" ":agentic"] " " [code-span "s" "true"] " " [code-span "r" ":max-rounds"] " " [code-span "s" "10"] "\n"
                    "    " [code-span "r" ":prompt"] " " [code-span "s" "\"PR: ~(ref \\\"pr\\\")...\""] "))"]}
            {:num "03" :title "C.L.E.A.R. gates verify"
             :desc "LLM-powered gates verify output against the codebase. Clarity, Logic, Empathy, Actionability, Relevance."
             :code ["(" [code-span "k" "llm"] " " [code-span "r" ":agentic"] " " [code-span "s" "true"] "\n"
                    "  " [code-span "r" ":tools"] " [" [code-span "s" "\"glitch_read_file\""] " " [code-span "s" "\"glitch_grep\""] "]\n"
                    "  " [code-span "r" ":prompt"] " " [code-span "s" "\"Verify against codebase...\""] ")"]}
            {:num "04" :title "compare and pick the best"
             :desc "Run agentic branches in parallel, let a neutral judge score them. Pick the winner automatically."
             :code ["(" [code-span "k" "compare"] "\n"
                    "  (" [code-span "k" "branch"] " " [code-span "s" "\"local\""] " (" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"lmstudio\""] " ...))\n"
                    "  (" [code-span "k" "branch"] " " [code-span "s" "\"cloud\""] " (" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"claude\""] " ...))\n"
                    "  (" [code-span "k" "review"] " " [code-span "r" ":criteria"] " (" [code-span "s" "\"accuracy\""] " " [code-span "s" "\"depth\""] ")))"]}]]
       ^{:key num}
       [:div {:style {:display :flex :flex-direction :column}}
        [:div {:style {:font-size "0.65rem" :font-weight 700
                       :color (t/colors :accent) :letter-spacing "0.15em"
                       :opacity 0.5 :margin-bottom "14px"}}
         (str num " " title)]
        (into [:pre.code {:style {:font-size "0.75rem" :line-height 1.9 :margin-bottom "16px"}}]
              code)
        [:p {:style {:font-size "0.82rem" :color (t/colors :fg-dim) :line-height 1.8}}
         desc]])]]])

(defn meta-section []
  [:section {:style {:border-top (str "1px solid " (t/colors :border))
                     :padding "120px 0"}}
   [:div.wrap
    [:div {:style {:font-size "0.65rem" :text-transform :uppercase
                   :letter-spacing "0.2em" :color (t/colors :accent)
                   :margin-bottom "12px"}}
     "proof it works"]
    [:h2 {:style {:font-size (t/sizes :h2) :margin-bottom "20px"}}
     "This site builds itself."]
    [:p {:style {:color (t/colors :fg) :opacity 0.5 :font-size "0.88rem"
                 :max-width "560px" :margin-bottom "48px" :line-height 1.9}}
     "Every page on this site was written by a gl1tch workflow. The LLM reads existing pages, greps the codebase for examples, and writes copy that matches what's actually true. C.L.E.A.R. gates verify every page before it hits disk."]
    [:div {:style {:display :grid :grid-template-columns "repeat(2, 1fr)" :gap "1px"
                   :background (t/colors :border) :border (str "1px solid " (t/colors :border))
                   :border-radius "6px" :overflow :hidden :margin-top "32px"}}
     (for [[cmd desc] [["glitch run site \"add a page about agentic steps\""
                         "LLM reads the codebase, writes the page, C.L.E.A.R. gates verify, saves to disk"]
                        ["glitch run site \"update homepage\""
                         "Agentic rewrite — LLM reads current page + source files, updates with verified content"]
                        ["glitch run site \"refresh all doc pages\""
                         "Regenerate every page sequentially, each gated"]
                        ["glitch run site \"start dev server\""
                         "shadow-cljs dev server with hot reload on port 3000"]]]
       ^{:key cmd}
       [:div {:style {:background (t/colors :bg-card) :padding "20px 24px"
                      :display :flex :flex-direction :column :gap "6px"}}
        [:code {:style {:font-size "0.75rem" :color (t/colors :accent)
                        :background :none :padding 0}}
         cmd]
        [:span {:style {:font-size "0.75rem" :color (t/colors :fg-dim)}}
         desc]])]]])

(defn get-started []
  [:section {:style {:border-top (str "1px solid " (t/colors :border))
                     :padding "120px 0"}}
   [:div.wrap
    [:div {:style {:font-size "0.65rem" :text-transform :uppercase
                   :letter-spacing "0.2em" :color (t/colors :accent)
                   :margin-bottom "12px"}}
     "get started"]
    [:h2 {:style {:font-size (t/sizes :h2) :margin-bottom "40px"}}
     "Three steps. Your first workflow runs in under a minute."]
    [:div {:style {:display :flex :flex-direction :column :gap "2px"
                   :background (t/colors :border) :border (str "1px solid " (t/colors :border))
                   :border-radius "8px" :overflow :hidden}}
     [:div {:style {:background (t/colors :bg-card) :padding "24px"}}
      [:div {:style {:font-size "0.8rem" :font-weight 700 :color (t/colors :accent)
                     :margin-bottom "12px"}}
       "1. Install"]
      [:pre.code {:style {:margin 0 :border :none :border-radius "4px"
                          :font-size "0.78rem" :padding "12px 16px"
                          :background (t/colors :bg-alt)}}
       [code-span "p" "$ "] "brew install 8op-org/tap/glitch"]
      [:p {:style {:font-size "0.82rem" :color (t/colors :fg-dim) :line-height 1.6 :margin-top "8px"}}
       "Single binary via Homebrew. No runtime dependencies, no Docker, no config files."]]
     [:div {:style {:background (t/colors :bg-card) :padding "24px"}}
      [:div {:style {:font-size "0.8rem" :font-weight 700 :color (t/colors :accent)
                     :margin-bottom "12px"}}
       "2. Write your first workflow"]
      [:pre.code {:style {:margin 0 :border :none :border-radius "4px"
                          :font-size "0.78rem" :padding "12px 16px"
                          :background (t/colors :bg-alt)}}
       "(" [code-span "k" "workflow"] " " [code-span "s" "\"hello\""] "\n"
       "  (" [code-span "k" "step"] " " [code-span "s" "\"greet\""] "\n"
       "    (" [code-span "k" "llm"] " " [code-span "r" ":provider"] " " [code-span "s" "\"copilot\""] "\n"
       "      " [code-span "r" ":prompt"] " (" [code-span "k" "str"] " " [code-span "s" "\"Summarize this repo: \""] "\n"
       "        (" [code-span "k" "sh"] " " [code-span "s" "\"git log --oneline -10\""] ")))))"]
      [:p {:style {:font-size "0.82rem" :color (t/colors :fg-dim) :line-height 1.6 :margin-top "8px"}}
       "Create " [:code ".glitch/workflows/hello.glitch"] " with the above."]]
     [:div {:style {:background (t/colors :bg-card) :padding "24px"}}
      [:div {:style {:font-size "0.8rem" :font-weight 700 :color (t/colors :accent)
                     :margin-bottom "12px"}}
       "3. Run it"]
      [:pre.code {:style {:margin 0 :border :none :border-radius "4px"
                          :font-size "0.78rem" :padding "12px 16px"
                          :background (t/colors :bg-alt)}}
       [code-span "p" "$ "] "glitch run hello"]
      [:p {:style {:font-size "0.82rem" :color (t/colors :fg-dim) :line-height 1.6 :margin-top "8px"}}
       "That's it. Shell fetches the git log, your LLM summarizes it. Add more steps, swap providers, add gates — your workflow grows with you."]]]

    ;; MCP section
    [:div {:style {:font-size "0.65rem" :text-transform :uppercase :letter-spacing "0.2em"
                   :color (t/colors :accent) :margin-bottom "12px" :margin-top "3rem"}}
     "give your agents glitch"]
    [:h2 {:style {:font-size (t/sizes :h2) :margin-bottom "12px"}}
     "Connect glitch as an MCP server to any AI agent."]
    [:p {:style {:color (t/colors :fg) :opacity 0.5 :font-size "0.88rem"
                 :max-width "560px" :margin-bottom "40px"}}
     "Your agent gets 8 tools: read files, grep, search, run workflows, eval expressions, check syntax, search symbols, and index repos."]
    [:div {:style {:display :flex :flex-direction :column :gap "2px"
                   :background (t/colors :border) :border (str "1px solid " (t/colors :border))
                   :border-radius "8px" :overflow :hidden}}
     [:div {:style {:background (t/colors :bg-card) :padding "24px"}}
      [:div {:style {:font-size "0.8rem" :font-weight 700 :color (t/colors :accent)
                     :margin-bottom "12px"}}
       "Claude Code"]
      [:pre.code {:style {:margin 0 :border :none :border-radius "4px"
                          :font-size "0.78rem" :padding "12px 16px"
                          :background (t/colors :bg-alt)}}
       [code-span "p" "$ "] "claude mcp add glitch -- glitch mcp"]]
     [:div {:style {:background (t/colors :bg-card) :padding "24px"}}
      [:div {:style {:font-size "0.8rem" :font-weight 700 :color (t/colors :accent)
                     :margin-bottom "12px"}}
       "Any MCP client"]
      [:pre.code {:style {:margin 0 :border :none :border-radius "4px"
                          :font-size "0.78rem" :padding "12px 16px"
                          :background (t/colors :bg-alt)}}
       [code-span "p" "$ "] "glitch mcp"]
      [:p {:style {:font-size "0.82rem" :color (t/colors :fg-dim) :line-height 1.6 :margin-top "8px"}}
       "Stdio server. Works with Cursor, Windsurf, Copilot CLI, or any MCP-compatible agent."]]]]])

(defn footer []
  [:footer {:style {:padding "80px 0" :border-top (str "1px solid " (t/colors :border))}}
   [:div.wrap {:style {:text-align :center}}
    [:div {:style {:display :flex :justify-content :center :gap "32px" :margin-bottom "24px"}}
     (for [[label href] [["GitHub" "https://github.com/8op-org/gl1tch"]
                          ["Docs" "#/docs"]
                          ["Labs" "#/labs"]
                          ["Changelog" "#/changelog"]]]
       ^{:key label}
       [:a {:href href
            :style {:color (t/colors :fg-dim) :text-decoration :none
                    :font-size "0.75rem" :letter-spacing "0.04em"}}
        label])]
    [:div {:style {:font-size "0.7rem" :color (t/colors :fg-dim) :opacity 0.4}}
     "gl1tch — your AI, your terminal, your rules"]]])

(defn page []
  [:div
   [hex-rain/hex-rain]
   [hero]
   [features]
   [how-it-works]
   [meta-section]
   [get-started]
   [footer]])
