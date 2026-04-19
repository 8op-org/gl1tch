(defn kebab-case [s]
  "Convert string to kebab-case."
  (->> s
       string/ascii-lower
       (string/replace-all " " "-")
       (string/replace-all "_" "-")))

(defn snake-case [s]
  "Convert string to snake_case."
  (->> s
       string/ascii-lower
       (string/replace-all " " "_")
       (string/replace-all "-" "_")))

(defn blank? [s]
  "True if string is empty or whitespace only."
  (= "" (string/trim (or s ""))))

(defn present? [s]
  "True if string is non-empty and not just whitespace."
  (not (blank? s)))

(defn words [s]
  "Split string on whitespace."
  (filter |(not= $ "")
    (string/split " " (string/trim s))))

(defn unwords [lst]
  "Join array with spaces."
  (string/join lst " "))
