(require '[babashka.http-client :as http])
(require '[cheshire.core :as json])
(require '[glitch.provider :as provider])

(provider/register "ollama"
  (fn [{:keys [prompt model]}]
    (let [base-url (or (System/getenv "OLLAMA_BASE_URL") "http://localhost:11434")
          model    (or model "llama3")
          body     (json/generate-string
                     {:model    model
                      :messages [{:role "user" :content prompt}]
                      :stream   false})
          resp     (http/post (str base-url "/v1/chat/completions")
                     {:headers {"Content-Type" "application/json"}
                      :body    body})
          parsed   (json/parse-string (:body resp) true)]
      (when (:error parsed)
        (throw (ex-info (str "ollama: " (get-in parsed [:error :message] "unknown error")) {})))
      (let [choice (get-in parsed [:choices 0 :message :content] "")
            usage  (get parsed :usage {})]
        {:response   choice
         :tokens-in  (or (:prompt_tokens usage) 0)
         :tokens-out (or (:completion_tokens usage) 0)}))))
