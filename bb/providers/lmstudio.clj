(require '[babashka.http-client :as http])
(require '[cheshire.core :as json])
(require '[glitch.provider :as provider])
(require '[glitch.tool_loop :as tool-loop])

(defn- mcp->openai-tool
  "Convert an MCP tool definition to OpenAI function-calling format."
  [mcp-def]
  {"type"     "function"
   "function" {"name"        (get mcp-def "name")
               "description" (get mcp-def "description")
               "parameters"  (get mcp-def "inputSchema")}})

(provider/register "lmstudio"
  (fn [{:keys [prompt model tool-defs tool-handler agentic max-rounds]}]
    (let [base-url  (or (System/getenv "LMSTUDIO_BASE_URL") "http://localhost:1234")
          model     (or model "default")
          messages  [{:role "user" :content prompt}]
          call-fn   (fn [msgs tools]
                      (let [body-map (cond-> {:model   model
                                              :messages msgs
                                              :stream  false}
                                       (seq tools) (assoc :tools tools))
                            resp     (http/post (str base-url "/v1/chat/completions")
                                       {:headers {"Content-Type" "application/json"}
                                        :body    (json/generate-string body-map)})
                            parsed   (json/parse-string (:body resp) true)]
                        (when (:error parsed)
                          (throw (ex-info (str "lmstudio: " (get-in parsed [:error :message] "unknown error")) {})))
                        parsed))
          tools     (when (seq tool-defs)
                      (mapv mcp->openai-tool tool-defs))]
      (if (seq tools)
        ;; Tool path: try with tools, fall back to plain call on failure or empty result
        (let [result (try
                       (tool-loop/run-loop call-fn messages tools tool-handler
                                           {:agentic   (boolean agentic)
                                            :max-rounds (or max-rounds 5)})
                       (catch Exception e
                         (binding [*out* *err*]
                           (println (str "lmstudio: tool-loop failed, retrying without tools: " (.getMessage e))))
                         nil))
              ;; Fall back when result is nil, empty string, or blank
              response (if (and (string? result) (seq (.trim result)))
                         result
                         ;; Retry without tools
                         (let [parsed  (call-fn messages nil)
                               usage   (get parsed :usage {})]
                           (get-in parsed [:choices 0 :message :content] "")))]
          ;; Usage is not available after run-loop, so return zeros for tokens
          {:response   response
           :tokens-in  0
           :tokens-out 0})

        ;; No tools — plain single call, preserve token counts
        (let [parsed  (call-fn messages nil)
              usage   (get parsed :usage {})]
          {:response   (get-in parsed [:choices 0 :message :content] "")
           :tokens-in  (or (:prompt_tokens usage) 0)
           :tokens-out (or (:completion_tokens usage) 0)})))))
