(require '[babashka.http-client :as http])
(require '[cheshire.core :as json])
(require '[clojure.string :as str])
(require '[glitch.provider :as provider])
(require '[glitch.tool_loop :as tool-loop])

(provider/register "openrouter"
  (fn [{:keys [prompt model tool-defs tool-handler agentic max-rounds]}]
    (let [api-key (System/getenv "OPENROUTER_API_KEY")]
      (when-not api-key
        (throw (ex-info "openrouter: OPENROUTER_API_KEY not set" {})))
      (let [base-url   (or (System/getenv "OPENROUTER_BASE_URL") "https://openrouter.ai/api/v1")
            model      (or model (System/getenv "OPENROUTER_MODEL") "google/gemma-3-12b-it")
            headers    {"Content-Type"  "application/json"
                        "Authorization" (str "Bearer " api-key)
                        "HTTP-Referer"  "https://gl1tch.dev"
                        "X-Title"       "glitch"}
            tools      (when (seq tool-defs)
                         (mapv (fn [d]
                                 {"type"     "function"
                                  "function" {"name"        (get d "name")
                                              "description" (get d "description")
                                              "parameters"  (get d "inputSchema")}})
                               tool-defs))
            last-usage (atom {})
            call-fn    (fn [messages active-tools]
                         (let [payload (cond-> {:model    model
                                                :messages messages
                                                :stream   false}
                                         (seq active-tools) (assoc :tools active-tools))
                               resp    (http/post (str base-url "/chat/completions")
                                         {:headers headers
                                          :body    (json/generate-string payload)})
                               parsed  (json/parse-string (:body resp) true)]
                           (when (:error parsed)
                             (throw (ex-info (str "openrouter: "
                                                  (get-in parsed [:error :message] "unknown error"))
                                             {})))
                           (reset! last-usage (get parsed :usage {}))
                           parsed))
            messages   [{:role "user" :content prompt}]
            response   (tool-loop/run-loop call-fn messages tools tool-handler
                                           {:agentic    (boolean agentic)
                                            :max-rounds (or max-rounds 5)})
            usage      @last-usage]
        {:response   response
         :tokens-in  (or (:prompt_tokens usage) 0)
         :tokens-out (or (:completion_tokens usage) 0)}))))
