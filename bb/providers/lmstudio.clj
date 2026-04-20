(require '[babashka.http-client :as http])
(require '[cheshire.core :as json])
(require '[glitch.provider :as provider])

(provider/register "lmstudio"
  (fn [{:keys [prompt model]}]
    (let [base-url (or (System/getenv "LMSTUDIO_BASE_URL") "http://localhost:1234")
          model    (or model "default")
          body     (json/generate-string
                     {:model    model
                      :messages [{:role "user" :content prompt}]
                      :stream   false})
          resp     (http/post (str base-url "/v1/chat/completions")
                     {:headers {"Content-Type" "application/json"}
                      :body    body})
          parsed   (json/parse-string (:body resp) true)]
      (when (:error parsed)
        (throw (ex-info (str "lmstudio: " (get-in parsed [:error :message] "unknown error")) {})))
      (let [choice (get-in parsed [:choices 0 :message :content] "")
            usage  (get parsed :usage {})]
        {:response   choice
         :tokens-in  (or (:prompt_tokens usage) 0)
         :tokens-out (or (:completion_tokens usage) 0)}))))
