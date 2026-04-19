# Batch/compare: cross-review parsing.

(defn- find-winner [text]
  (def upper (string/ascii-upper text))
  (var winner nil)
  (each line (string/split "\n" upper)
    (def trimmed (string/trim line))
    (when (string/has-prefix? "WINNER:" trimmed)
      (set winner (string/trim (string/slice trimmed 7)))
      (break)))
  winner)

(defn- parse-numeric [output]
  (def lines (string/split "\n" output))
  (def winner (find-winner output))
  (def results @[])
  (var current nil)
  (var passed 0)
  (var total 0)

  (each line lines
    (def trimmed (string/trim line))
    (def upper (string/ascii-upper trimmed))

    (label skip
      (when (string/has-prefix? "VARIANT:" upper)
        # Save previous
        (when (and current (> total 0))
          (array/push results
            {:variant (string/ascii-lower current)
             :passed passed :total total
             :winner (and winner
                       (= (string/ascii-lower winner)
                          (string/ascii-lower current)))}))
        (set current (string/trim (string/slice trimmed 8)))
        (set passed 0)
        (set total 0)
        (return skip))

      # Skip meta lines
      (when (or (string/has-prefix? "WINNER:" upper)
                (string/has-prefix? "REASON:" upper)
                (string/has-prefix? "NOTES:" upper)
                (string/has-prefix? "TOTAL:" upper))
        (return skip))

      # Parse score lines: "label: N/M"
      (when (and current
                 (string/find "/" trimmed)
                 (string/find ":" trimmed))
        (def parts (string/split ":" trimmed 0 2))
        (when (= 2 (length parts))
          (def score-part (string/trim (get parts 1)))
          (def num-denom (string/split "/" score-part 0 2))
          (when (= 2 (length num-denom))
            (def num (scan-number (string/trim (get num-denom 0))))
            (when (scan-number (string/trim (get num-denom 1)))
              (when num
                (++ total)
                (when (>= num 7) (++ passed)))))))))

  # Save last
  (when (and current (> total 0))
    (array/push results
      {:variant (string/ascii-lower current)
       :passed passed :total total
       :winner (and winner
                 (= (string/ascii-lower winner)
                    (string/ascii-lower current)))}))
  results)

(defn- parse-pass-fail [output]
  (def upper (string/replace-all "*" "" (string/ascii-upper output)))
  (def lines (string/split "\n" upper))
  (def winner (find-winner output))
  (def results @[])
  (var current nil)
  (var passed 0)
  (var total 0)

  (each line lines
    (def trimmed (string/trim line))

    (label skip
      # Detect --- VARIANT ---
      (when (and (string/has-prefix? "---" trimmed)
                 (string/has-suffix? "---" trimmed))
        (when (and current (> total 0))
          (array/push results
            {:variant (string/ascii-lower current)
             :passed passed :total total
             :winner (and winner
                       (not (nil? (string/find (string/ascii-upper current)
                                              (string/ascii-upper winner)))))}))
        (def name (string/trim (string/replace-all "-" "" trimmed)))
        (when (not= name "")
          (set current (string/trim name))
          (set passed 0)
          (set total 0))
        (return skip))

      # Skip meta lines
      (when (or (string/has-prefix? "SCORE:" trimmed)
                (string/has-prefix? "OVERALL" trimmed)
                (string/has-prefix? "WINNER" trimmed))
        (return skip))

      # Count PASS/FAIL
      (when current
        (def has-pass (string/find "PASS" trimmed))
        (def has-fail (string/find "FAIL" trimmed))
        (cond
          (and has-pass (not has-fail)) (do (++ passed) (++ total))
          has-fail (++ total)))))

  # Save last
  (when (and current (> total 0))
    (array/push results
      {:variant (string/ascii-lower current)
       :passed passed :total total
       :winner (and winner
                 (not (nil? (string/find (string/ascii-upper current)
                                        (string/ascii-upper winner)))))}))
  results)

(defn parse-cross-review [output]
  "Parse cross-review LLM output into per-variant scores."
  (def upper (string/replace-all "*" "" (string/ascii-upper output)))
  (if (or (string/find "VARIANT:" upper)
          (string/has-prefix? "VARIANT:" upper))
    (parse-numeric output)
    (parse-pass-fail output)))
