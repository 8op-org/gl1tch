(ns gl1tch.site.components.hex-rain
  (:require [reagent.core :as r]))

(def ^:private col-w 28)
(def ^:private row-h 18)

(defn- hex-byte []
  (-> (js/Math.random) (* 256) js/Math.floor (.toString 16) (.padStart 2 "0") .toUpperCase))

(defn- setup [canvas]
  (let [w (.-innerWidth js/window)
        h (.-innerHeight js/window)
        cols (+ (js/Math.ceil (/ w col-w)) 2)
        rows (+ (js/Math.ceil (/ h row-h)) 3)
        grid-h (* rows 4)]
    (set! (.-width canvas) w)
    (set! (.-height canvas) h)
    {:cols cols :rows rows
     :offsets (vec (repeatedly cols #(* (js/Math.random) rows row-h)))
     :speeds (vec (repeatedly cols #(+ 0.25 (* (js/Math.random) 0.45))))
     :grid (vec (repeatedly grid-h
                  (fn [] (vec (repeatedly cols hex-byte)))))}))

(defn hex-rain []
  (let [state (r/atom nil)
        canvas-ref (r/atom nil)
        raf-id (r/atom nil)]
    (r/create-class
      {:component-did-mount
       (fn [_]
         (when-let [canvas @canvas-ref]
           (reset! state (setup canvas))
           (let [ctx (.getContext canvas "2d")]
             (letfn [(draw []
                       (let [{:keys [cols rows offsets speeds grid]} @state]
                         (.clearRect ctx 0 0 (.-width canvas) (.-height canvas))
                         (set! (.-font ctx) "12px 'JetBrains Mono', monospace")
                         (set! (.-fillStyle ctx) "rgba(135, 135, 175, 0.07)")
                         (let [new-offsets
                               (mapv (fn [col]
                                       (let [off (mod (+ (nth offsets col) (nth speeds col))
                                                      (* (count grid) row-h))
                                             start-row (js/Math.floor (/ off row-h))
                                             sub-px (mod off row-h)]
                                         (doseq [row (range (+ rows 2))]
                                           (let [y (- (* row row-h) sub-px)
                                                 grid-row (mod (+ start-row row) (count grid))]
                                             (.fillText ctx (get-in grid [grid-row col]) (* col col-w) y)))
                                         off))
                                     (range cols))]
                           (swap! state assoc :offsets new-offsets))
                         (reset! raf-id (js/requestAnimationFrame draw))))]
               (draw)))))
       :component-will-unmount
       (fn [_]
         (when @raf-id
           (js/cancelAnimationFrame @raf-id)))
       :reagent-render
       (fn []
         [:canvas {:ref #(reset! canvas-ref %)
                   :style {:position :fixed
                           :top 0 :left 0
                           :width "100%" :height "100%"
                           :z-index 0
                           :pointer-events :none
                           :opacity 0.04}}])})))
