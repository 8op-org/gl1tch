(ns glitch.tool_loop
  "Tool-calling loop for HTTP providers (openrouter, lmstudio).
   Manages: tool call parsing, execution via handler, message assembly,
   max-rounds, and agentic mode."
  (:require [cheshire.core :as json]))

(defn- parse-tool-calls
  "Extract tool_calls from an OpenAI-compatible response choice."
  [choice]
  (get-in choice [:message :tool_calls]))

(defn- execute-tool-call
  "Execute a single tool call via the handler fn. Returns the tool result message."
  [tool-handler tool-call]
  (let [fn-name (get-in tool-call [:function :name])
        args-str (get-in tool-call [:function :arguments])
        args (try (json/parse-string args-str) (catch Exception _ {}))
        result (try
                 (tool-handler fn-name args)
                 (catch Exception e
                   (str "Error: " (.getMessage e))))]
    {:role "tool"
     :tool_call_id (:id tool-call)
     :content (str result)}))

(defn- has-text-response?
  "Check if the choice has a non-empty text content (not just tool calls)."
  [choice]
  (let [content (get-in choice [:message :content])]
    (and (string? content) (seq (.trim content)))))

(defn run-loop
  "Run the tool-calling loop for an HTTP provider.

   Arguments:
     call-fn      — (fn [messages tools] => parsed-response) makes the API call
     messages     — initial message list [{:role \"user\" :content \"...\"}]
     tools        — tool definitions in OpenAI format, or nil/[] to skip
     tool-handler — (fn [tool-name args-map] => string) executes a tool
     opts         — {:agentic bool, :max-rounds int}

   Returns the final text response string."
  [call-fn messages tools tool-handler {:keys [agentic max-rounds]
                                         :or {agentic false max-rounds 5}}]
  (if (or (nil? tools) (empty? tools))
    ;; No tools — single call, return text
    (let [resp (call-fn messages nil)
          choice (get-in resp [:choices 0])]
      (get-in choice [:message :content] ""))

    ;; Tool loop
    (loop [msgs messages
           round 0]
      (let [resp (call-fn msgs tools)
            choice (get-in resp [:choices 0])
            tool-calls (parse-tool-calls choice)]
        (cond
          ;; No tool calls — LLM produced text, we're done
          (or (nil? tool-calls) (empty? tool-calls))
          (get-in choice [:message :content] "")

          ;; Max rounds hit — force text response
          (>= round max-rounds)
          (let [;; Append assistant message + tool results, then call without tools
                assistant-msg {:role "assistant" :content (get-in choice [:message :content])
                               :tool_calls tool-calls}
                tool-results (mapv #(execute-tool-call tool-handler %) tool-calls)
                final-msgs (into (conj msgs assistant-msg) tool-results)
                final-resp (call-fn final-msgs nil)
                final-choice (get-in final-resp [:choices 0])]
            (get-in final-choice [:message :content] ""))

          ;; Tool calls present — execute and continue
          :else
          (let [assistant-msg {:role "assistant" :content (get-in choice [:message :content])
                               :tool_calls tool-calls}
                tool-results (mapv #(execute-tool-call tool-handler %) tool-calls)
                next-msgs (into (conj msgs assistant-msg) tool-results)]
            (if agentic
              ;; Agentic: re-send with tools for another round
              (recur next-msgs (inc round))
              ;; Single-round: one tool pass, then force text
              (let [final-resp (call-fn next-msgs nil)
                    final-choice (get-in final-resp [:choices 0])]
                (get-in final-choice [:message :content] "")))))))))
