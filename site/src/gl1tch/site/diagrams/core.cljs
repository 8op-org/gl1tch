(ns gl1tch.site.diagrams.core)

(def ^:private node-w 120)
(def ^:private node-h 40)
(def ^:private gap-x 60)
(def ^:private gap-y 30)
(def ^:private padding 20)

(defn assign-layers
  "Topological sort into layers. Nodes with no incoming edges go first."
  [nodes edges]
  (let [ids (set (map :id nodes))
        in-map (reduce (fn [m [from to]] (update m to (fnil conj #{}) from))
                       {} edges)]
    (loop [remaining ids
           layers []]
      (if (empty? remaining)
        layers
        (let [ready (filterv #(every? (fn [dep] (not (remaining dep)))
                                       (get in-map % #{}))
                             remaining)]
          (if (empty? ready)
            (conj layers (vec remaining))
            (recur (apply disj remaining ready)
                   (conj layers ready))))))))

(defn- node-center [node]
  [(+ (:x node) (/ (:w node) 2))
   (+ (:y node) (/ (:h node) 2))])

(defn- compute-edge-path [from-node to-node direction]
  (let [[fx fy] (node-center from-node)
        [tx ty] (node-center to-node)]
    (if (= direction :lr)
      [[(+ (:x from-node) (:w from-node)) fy]
       [(:x to-node) ty]]
      [[fx (+ (:y from-node) (:h from-node))]
       [tx (:y to-node)]])))

(defn compute
  "Takes a flowchart spec {:direction :lr/:tb :nodes [...] :edges [...]}
   Returns {:nodes [...with :x :y :w :h...] :edges [...with :path...]
            :width :height}"
  [{:keys [direction nodes edges]
    :or {direction :lr}}]
  (let [layers (assign-layers nodes edges)
        node-map (into {} (map (juxt :id identity) nodes))
        positioned
        (reduce
          (fn [acc [layer-idx layer]]
            (reduce
              (fn [acc2 [pos-in-layer node-id]]
                (let [node (get node-map node-id)
                      [x y] (if (= direction :lr)
                              [(+ padding (* layer-idx (+ node-w gap-x)))
                               (+ padding (* pos-in-layer (+ node-h gap-y)))]
                              [(+ padding (* pos-in-layer (+ node-w gap-x)))
                               (+ padding (* layer-idx (+ node-h gap-y)))])]
                  (assoc acc2 node-id
                         (merge node {:x x :y y :w node-w :h node-h}))))
              acc
              (map-indexed vector layer)))
          {}
          (map-indexed vector layers))
        edge-results
        (mapv (fn [[from to]]
                {:from from :to to
                 :path (compute-edge-path (positioned from) (positioned to) direction)})
              edges)
        all-nodes (mapv positioned (mapcat identity layers))
        max-x (apply max (map #(+ (:x %) (:w %)) all-nodes))
        max-y (apply max (map #(+ (:y %) (:h %)) all-nodes))]
    {:nodes all-nodes
     :edges edge-results
     :width (+ max-x padding)
     :height (+ max-y padding)}))
