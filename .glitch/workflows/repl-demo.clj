(require '[glitch.core :as g])

;; Put your cursor at the end of any form below and press C-c C-e to eval it.
;; The result appears inline next to the form.

;; 1. Shell command
(g/step "uptime" (g/sh "uptime"))

;; 2. Reference the result
(g/ref "uptime")

(g/sh "date")

;; 3. Feed it to the LLM
(g/step "commentary"
        (g/llm :provider "openrouter" :model "google/gemini-2.5-flash" :prompt (str "Give a one-sentence sarcastic take on this uptime:\n" (g/ref "uptime"))))

;; 4. See what you got
(g/ref "commentary")
(g/save "results/commentary.txt" (g/ref "commentary"))

;; 5. All steps so far
(g/get-steps)
