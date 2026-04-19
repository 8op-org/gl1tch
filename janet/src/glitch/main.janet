(defn main [& argv]
  (def args (drop 1 argv))
  (if (= (length args) 0)
    (do
      (print "glitch - workflow engine")
      (print "usage: glitch <command> [args...]")
      (print "commands: version")
      (os/exit 1))
    (match (first args)
      "version" (print "glitch 0.1.0-janet")
      (do (eprintf "unknown command: %s" (first args))
          (os/exit 1)))))
