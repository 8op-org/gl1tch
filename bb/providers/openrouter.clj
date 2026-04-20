(require '[babashka.http-client :as http])
(require '[cheshire.core :as json])
(require '[clojure.string :as str])
(require '[glitch.provider :as provider])

(provider/register "openrouter"
  (fn [{:keys [prompt model]}]
    (let [api-key (System/getenv "OPENROUTER_API_KEY")]
      (when-not api-key
        (throw (ex-info "openrouter: OPENROUTER_API_KEY not set" {})))
      (let [base-url (or (System/getenv "OPENROUTER_BASE_URL") "https://openrouter.ai/api/v1")
            model    (or model "google/gemma-3-4b-it:free")
            body     (json/generate-string
                       {:model    model
                        :messages [{:role "user" :content prompt}]
                        :stream   false})
            resp     (http/post (str base-url "/chat/completions")
                       {:headers {"Content-Type"  "application/json"
                                  "Authorization" (str "Bearer " api-key)
                                  "HTTP-Referer"  "https://gl1tch.dev"
                                  "X-Title"       "glitch"}
                        :body    body})
            parsed   (json/parse-string (:body resp) true)]
        (when (:error parsed)
          (throw (ex-info (str "openrouter: " (get-in parsed [:error :message] "unknown error")) {})))
        (let [choice (get-in parsed [:choices 0 :message :content] "")
              usage  (get parsed :usage {})]
          {:response   choice
           :tokens-in  (or (:prompt_tokens usage) 0)
           :tokens-out (or (:completion_tokens usage) 0)})))))
