(ns glitch.repl
  "nREPL server with glitch DSL preloaded.
   Starts babashka.nrepl and writes .nrepl-port for CIDER."
  (:require [babashka.classpath :as cp]
            [babashka.nrepl.server :as nrepl]
            [clojure.java.io :as io]
            [clojure.pprint :as pp]
            [clojure.string :as str]
            [glitch.api :as api]
            [glitch.core :as g]
            [glitch.provider :as prov]
            [glitch.plugin :as plugin]
            [glitch.plugin-loader :as plugin-loader]
            [glitch.session :as session]
            [glitch.promote :as promote]))

;; ---------------------------------------------------------------------------
;; REPL-injected session, promote, recall functions
;; ---------------------------------------------------------------------------

(defn- session-summary
  "Pretty-print a session summary: step count, types used, duration."
  [entries]
  (let [cnt    (count entries)
        types  (->> entries (map :type) (remove nil?) distinct sort vec)
        t0     (some :timestamp entries)
        t1     (some :timestamp (reverse entries))
        dur    (if (and t0 t1) (- t1 t0) 0)]
    (pp/pprint {:steps cnt
                :types types
                :duration-ms dur
                :session-id @session/*session-id*})))

(defn repl-session
  "Session inspector for the REPL.
   (session)                           — print summary
   (session :clear)                    — clear and restart session
   (session :slice :from \"id\")         — set slice start
   (session :slice :from \"a\" :to \"b\") — set slice range"
  [& args]
  (cond
    (empty? args)
    (session-summary (session/current-session))

    (= :clear (first args))
    (do (session/clear-session!)
        (println "session cleared"))

    (= :slice (first args))
    (let [opts (apply hash-map (rest args))]
      (session/set-slice! opts)
      (println (str "slice set: " (pr-str opts))))

    :else
    (println (str "unknown session command: " (first args)))))

(defn repl-promote
  "Promote the current session into a .glitch workflow.
   (promote \"description\")
   (promote \"description\" :from \"step-a\" :to \"step-b\")"
  [description & {:keys [from to] :as opts}]
  (let [entries (session/get-slice)
        result  (promote/promote! description
                                  :session entries
                                  :from from
                                  :to to)
        source  (:workflow-source result)
        path    (:path result)]
    (println "--- generated workflow ---")
    (println source)
    (println "---")
    ;; nREPL does not support interactive read-line, so always save
    (io/make-parents path)
    (spit path source)
    (session/index-workflow! {:path          path
                              :description   description
                              :promoted-from @session/*session-id*
                              :tags          (:tags result)})
    (println (str "saved to " path))
    path))

(defn repl-recall
  "Search the workflow index.
   (recall \"deploy\")
   (recall \"deploy\" :run true)
   (recall \"deploy\" :run true :input \"some input\")"
  [query & {:keys [run input] :as opts}]
  (let [results (session/recall-search query)]
    (cond
      (empty? results)
      (println "no matching workflows found")

      (= 1 (count results))
      (let [r (first results)]
        (pp/pprint r)
        (when (and (:promoted-from r) (not (:path r)))
          (println "hint: this is an unpromoted session — use (promote ...) to save as a workflow"))
        (when (and run (:path r))
          (println (str "running " (:path r) " ..."))
          (g/call-workflow (-> (:path r)
                               (str/replace #"\.glitch/workflows/" "")
                               (str/replace #"\.(glitch|clj)$" ""))
                           :input (or input ""))))

      :else
      (doseq [[i r] (map-indexed vector results)]
        (println (str (inc i) ". " (:description r) " — " (:path r)))
        (when (:promoted-from r)
          (println (str "   (promoted from session " (:promoted-from r) ")")))))))

(defn port-file-path
  "Return the .nrepl-port file path for a directory."
  [dir]
  (str dir "/.nrepl-port"))

(defn start
  "Start an nREPL server with glitch DSL available in the user namespace.
   Options:
     :port — port number (default 1667)
     :dir  — directory for .nrepl-port file (default cwd)"
  [{:keys [port dir] :or {port 1667
                           dir (System/getProperty "user.dir")}}]
  (prov/load-providers)
  (plugin-loader/load-plugins)

  ;; Start session recording
  (session/init-session!)
  (alter-var-root #'g/*recording* (constantly true))

  ;; Add project-local source to classpath for require support
  (let [local-src (io/file dir ".glitch" "src")]
    (when (.isDirectory local-src)
      (cp/add-classpath (.getAbsolutePath local-src))
      (binding [*out* *err*]
        (println (str "added " (.getAbsolutePath local-src) " to classpath")))))

  ;; Inject plugin commands as real namespaces for nREPL use
  (doseq [[pname pmap] (plugin/all-plugins)]
    (let [ns-sym (symbol pname)]
      (create-ns ns-sym)
      (doseq [[cmd-name cmd] (:commands pmap)]
        (intern ns-sym (symbol cmd-name)
                (fn [& {:as opts}] ((:fn cmd) opts))))))

  (g/set-provider-fn!
    (fn [opts]
      (let [pname (or (:provider opts) "lmstudio")]
        (prov/call-provider pname opts))))

  ;; Inject glitch DSL into user namespace so no require is needed
  (binding [*ns* (the-ns 'user)]
    (refer 'glitch.core))

  ;; Inject session, promote, recall into user namespace
  (intern 'user 'session  repl-session)
  (intern 'user 'promote  repl-promote)
  (intern 'user 'recall   repl-recall)

  ;; Inject glitch.api into user namespace
  (binding [*ns* (the-ns 'user)]
    (refer 'glitch.api :only '[agent list-tools use-provider! use-model!])
    ;; deftool is a macro — intern the var directly so nREPL resolves it as a macro
    (ns-unmap *ns* 'deftool)
    (intern *ns* 'deftool @#'api/deftool)
    ;; remove-tool exposed without bang per design spec
    (intern *ns* 'remove-tool api/remove-tool!))

  (let [port-file (io/file (port-file-path dir))]
    (spit port-file (str port))
    (.deleteOnExit port-file))

  (binding [*out* *err*]
    (println (str "glitch repl on port " port))
    (println (str "connect: cider-connect localhost " port)))

  (nrepl/start-server! {:host "localhost" :port port})
  @(promise))
