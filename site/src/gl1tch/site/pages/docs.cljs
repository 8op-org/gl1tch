(ns gl1tch.site.pages.docs
  (:require [gl1tch.site.content :as content]
            [gl1tch.site.components.sidebar :as sidebar]
            [gl1tch.site.components.toc :as toc]
            [gl1tch.site.components.code :as code]
            [gl1tch.site.diagrams.flowchart :as flowchart]
            [gl1tch.site.styles.theme :as t]
            [clojure.string :as str]))

(defn- heading->id [heading]
  (-> heading str/lower-case (str/replace #"[^a-z0-9]+" "-") (str/replace #"^-|-$" "")))

(defn render-hiccup
  "Renders a hiccup body vector, resolving :code and :diagram tags."
  [element]
  (cond
    (string? element) element
    (nil? element) nil
    (vector? element)
    (let [[tag & rest] element
          [attrs children] (if (map? (first rest))
                             [(first rest) (vec (next rest))]
                             [nil (vec rest)])]
      (case tag
        :p (into [:p] (map render-hiccup children))
        :code (if attrs
                [code/code-block attrs (first children)]
                [code/inline-code (first children)])
        :diagram (case (:type attrs)
                   :flowchart [flowchart/render
                               (flowchart/parse-inline attrs children)]
                   [:div "Unknown diagram type"])
        :a (into [:a attrs] (map render-hiccup children))
        :strong (into [:strong] (map render-hiccup children))
        :em (into [:em] (map render-hiccup children))
        :ul (into [:ul] (map render-hiccup children))
        :ol (into [:ol] (map render-hiccup children))
        :li (into [:li] (map render-hiccup children))
        ;; Default: pass through
        (if attrs
          (into [tag attrs] (map render-hiccup children))
          (into [tag] (map render-hiccup children)))))
    :else (str element)))

(defn- render-section [{:keys [heading level body]}]
  [:section {:id (heading->id heading)
             :style {:margin-top (if (= level 2) "32px" "20px")}}
   [(keyword (str "h" level))
    {:style {:color (if (= level 2) (t/colors :accent) (t/colors :accent-2))
             :margin-bottom "12px"}}
    heading]
   (into [:div] (map render-hiccup body))])

(defn page [slug]
  (if-let [doc (content/doc-by-slug slug)]
      [:div {:style {:max-width "1280px"
                     :margin "0 auto"
                     :padding "80px 40px 0"
                     :display :grid
                     :grid-template-columns "220px 1fr 200px"
                     :gap "0 48px"}}
       [:div [sidebar/sidebar]]
       [:article {:style {:min-width 0 :max-width "720px"}}
        [:h1 {:style {:font-size (t/sizes :h2)
                      :margin-bottom "12px"}}
         (:title doc)]
        (when (:description doc)
          [:p {:style {:color (t/colors :fg-dim)
                       :margin-bottom "40px"}}
           (:description doc)])
        (for [section (:sections doc)]
          ^{:key (:heading section)}
          [render-section section])]
       [:div [toc/toc (content/extract-toc doc)]]]
      [:div.wrap [:h1 "Doc not found"]]))

(defn index []
  [:div.wrap {:style {:padding-top "80px"}}
   [:h1 "Documentation"]
   [:div {:style {:margin-top "40px"}}
    (for [doc (content/docs-index)]
      ^{:key (:slug doc)}
      [:div {:style {:margin-bottom "24px"}}
       [:a {:href (str "#/docs/" (:slug doc))
            :style {:color (t/colors :accent)
                    :font-size "1.05rem"
                    :font-weight 700
                    :text-decoration :none}}
        (:title doc)]
       (when (:description doc)
         [:p {:style {:color (t/colors :fg-dim)
                      :font-size (t/sizes :sm)}}
          (:description doc)])])]])
