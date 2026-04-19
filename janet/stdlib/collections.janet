(defn compact [source]
  "Remove empty strings from an array."
  (filter |(not= $ "") source))

(defn first [source]
  "Return the first element."
  (get source 0))

(defn pluck [key source]
  "Extract a field from each element."
  (map |(get $ key) source))

(defn take [n source]
  "Return the first n elements."
  (array/slice source 0 (min n (length source))))

(defn unique [source]
  "Remove duplicates, preserving order."
  (def seen @{})
  (def result @[])
  (each item source
    (unless (get seen item)
      (put seen item true)
      (array/push result item)))
  result)

(defn without [source exclude]
  "Return source with exclude items removed."
  (def ex-set (from-pairs (map |[$ true] exclude)))
  (filter |(not (get ex-set $)) source))
