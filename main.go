package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Forward ldflags to cmd package
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.SetVersionString()

	if home, err := os.UserHomeDir(); err == nil {
		loadDotenv(filepath.Join(home, ".config", "glitch", ".env"))
	}
	loadDotenv(".env")

	cmd.Execute()
}

func loadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if _, set := os.LookupEnv(k); !set {
			os.Setenv(k, v)
		}
	}
}
