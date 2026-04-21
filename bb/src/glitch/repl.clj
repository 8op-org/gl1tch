(ns glitch.repl
  "nREPL server with glitch DSL preloaded.
   Starts babashka.nrepl and writes .nrepl-port for CIDER."
  (:require [babashka.nrepl.server :as nrepl]
            [clojure.java.io :as io]
            [glitch.core :as g]
            [glitch.provider :as prov]
            [glitch.plugin :as plugin]
            [glitch.plugin-loader :as plugin-loader]))

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

  (let [port-file (io/file (port-file-path dir))]
    (spit port-file (str port))
    (.deleteOnExit port-file))

  (binding [*out* *err*]
    (println (str "glitch repl on port " port))
    (println (str "connect: cider-connect localhost " port)))

  (nrepl/start-server! {:host "localhost" :port port})
  @(promise))
