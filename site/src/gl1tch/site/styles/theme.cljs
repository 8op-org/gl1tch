(ns gl1tch.site.styles.theme)

(def colors
  {:bg       "#0a0a0f"
   :bg-alt   "#12121a"
   :bg-card  "#161722"
   :fg       "#c0c0c8"
   :fg-dim   "#606070"
   :accent   "#00ff9f"
   :accent-2 "#7b61ff"
   :danger   "#ff4040"
   :border   "#1a1a2e"
   :glow     "rgba(0, 255, 159, 0.08)"})

(def fonts
  {:mono "'JetBrains Mono', 'Fira Code', monospace"
   :sans "Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"})

(def sizes
  {:base   "15px"
   :sm     "0.82rem"
   :xs     "0.72rem"
   :xxs    "0.65rem"
   :h1     "clamp(3rem, 8vw, 5rem)"
   :h2     "2rem"
   :h3     "1.05rem"})

(def spacing
  {:xs 4 :sm 8 :md 16 :lg 24 :xl 40})

(def radius "3px")
