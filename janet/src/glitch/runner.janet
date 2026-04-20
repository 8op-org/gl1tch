# Workflow execution lifecycle.
# Wires core + provider + store together.

(import glitch/core :as g)
(import glitch/store :as s)
(import glitch/provider :as p)

(defn run [workflow-path input &named db project model
           params seed-steps workflows-dir parent-run-id tiers]
  "Execute a workflow file. Returns {:name :output :steps :run-id}."
  (default input "")
  (default model "qwen2.5:7b")
  (default params @{})
  (var wf-dir workflows-dir)
  (when (nil? wf-dir)
    (def parts (string/find-all "/" workflow-path))
    (set wf-dir
      (if (empty? parts)
        "."
        (string/slice workflow-path 0 (last parts)))))

  # Reset core state
  (g/reset-steps!)
  (g/set-input! input)
  (g/set-params! params)
  (g/set-workflows-dir! wf-dir)

  # Wire provider dispatch
  (g/set-provider-fn!
    (fn [opts]
      (def pname (or (opts :provider) "ollama"))
      (def merged (merge opts {:model (or (opts :model) model)}))
      (if tiers
        (p/call-tiered merged tiers)
        (try
          (p/call-provider pname merged)
          ([err]
            (p/call-tiered merged p/default-tiers))))))

  # Pre-seed steps
  (when seed-steps
    (eachp [k v] seed-steps
      (g/step k v)))

  # Record run in store
  (var run-id 0)
  (when db
    (set run-id (s/record-run db
      {:name "" :input input
       :workflow-file workflow-path
       :model model
       :workspace (or project "")
       :parent-run-id parent-run-id}))
    # Wire step recorder to store
    (g/set-step-recorder!
      (fn [rec]
        (s/record-step db
          (merge rec {:run-id run-id})))))

  # Ensure module paths include src directory for workflow imports
  (def src-path "src/:all:.janet")
  (var has-src false)
  (each p module/paths
    (when (and (indexed? p) (> (length p) 0) (= (first p) src-path))
      (set has-src true)))
  (unless has-src
    (array/push module/paths [src-path :source]))

  # Execute the workflow
  (var result nil)
  (try
    (do
      (dofile workflow-path)
      # Gather results from core state
      (def steps (g/get-steps))
      (set result {:name ""
                   :output (g/last-output)
                   :steps steps}))
    ([err]
      (when db
        (s/finish-run db run-id (string "ERROR: " err) 1))
      (error err)))

  # Finalize
  (when db
    (s/finish-run db run-id (result :output) 0))

  (merge result {:run-id run-id}))

(defn list-workflows [dir]
  "List all .glitch and .janet workflow files in a directory."
  (def results @[])
  (when (os/stat dir)
    (each f (os/dir dir)
      (var ext-len nil)
      (cond
        (string/has-suffix? ".glitch" f) (set ext-len 7)
        (string/has-suffix? ".janet" f) (set ext-len 6))
      (when ext-len
        (array/push results
          {:name (string/slice f 0 (- (length f) ext-len))
           :path (string dir "/" f)}))))
  (sort-by |($ :name) results))
