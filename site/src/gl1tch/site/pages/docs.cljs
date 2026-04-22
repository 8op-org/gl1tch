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
        :p (into [:p.doc-p] (map render-hiccup children))
        :code (if attrs
                [code/code-block attrs (first children)]
                [code/inline-code (first children)])
        :diagram (case (:type attrs)
                   :flowchart [flowchart/render
                               (flowchart/parse-inline attrs children)]
                   [:div "Unknown diagram type"])
        :a (into [:a.doc-link attrs] (map render-hiccup children))
        :strong (into [:strong] (map render-hiccup children))
        :em (into [:em] (map render-hiccup children))
        :ul (into [:ul.doc-list] (map render-hiccup children))
        :ol (into [:ol.doc-list] (map render-hiccup children))
        :li (into [:li] (map render-hiccup children))
        ;; Default: pass through
        (if attrs
          (into [tag attrs] (map render-hiccup children))
          (into [tag] (map render-hiccup children)))))
    :else (str element)))

(defn- render-section [{:keys [heading level body]}]
  [:section.doc-section {:id (heading->id heading)}
   [(keyword (str "h" level))
    {:class (str "doc-h" level)}
    heading]
   (into [:div] (map render-hiccup body))])

(defn page [slug]
  (if-let [doc (content/doc-by-slug slug)]
    [:div.doc-layout
     [:div.doc-sidebar-col [sidebar/sidebar]]
     [:article.doc-content
      [:h1.doc-title (:title doc)]
      (when (:description doc)
        [:p.doc-desc (:description doc)])
      (for [section (:sections doc)]
        ^{:key (:heading section)}
        [render-section section])]
     [:div.doc-toc-col [toc/toc (content/extract-toc doc)]]]
    [:div.wrap [:h1 "Doc not found"]]))

(defn index []
  [:div.wrap {:style {:padding-top "80px"}}
   [:h1 "Documentation"]
   [:div {:style {:margin-top "32px"}}
    (for [doc (content/docs-index)]
      ^{:key (:slug doc)}
      [:div {:style {:margin-bottom "20px"}}
       [:a {:href (str "#/docs/" (:slug doc))
            :style {:color (t/colors :accent)
                    :font-size "0.95rem"
                    :font-weight 600
                    :text-decoration :none}}
        (:title doc)]
       (when (:description doc)
         [:p {:style {:color (t/colors :fg-dim)
                      :font-size (t/sizes :sm)
                      :margin-top "2px"}}
          (:description doc)])])]])
