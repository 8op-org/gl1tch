package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "start Elasticsearch and Kibana via Docker Compose",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dockerCompose("up", "-d")
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "stop Elasticsearch and Kibana",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dockerCompose("down")
	},
}

// findComposeFile locates docker-compose.yml by checking three places:
//  1. Next to the running binary
//  2. Source tree relative to this file (for go run / dev mode)
//  3. Current working directory
func findComposeFile() (string, error) {
	const name = "docker-compose.yml"

	// 1. Next to the binary
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// 2. Source tree — works for `go run .` and tests
	_, thisFile, _, ok := runtime.Caller(0)
	if ok {
		// thisFile is cmd/up.go; repo root is one level up
		p := filepath.Join(filepath.Dir(thisFile), "..", name)
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}

	// 3. Current working directory
	if cwd, err := os.Getwd(); err == nil {
		p := filepath.Join(cwd, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("docker-compose.yml not found (checked next to binary, source tree, and cwd)")
}

func dockerCompose(args ...string) error {
	composePath, err := findComposeFile()
	if err != nil {
		return err
	}

	cmdArgs := append([]string{"compose", "-f", composePath}, args...)
	c := exec.Command("docker", cmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
