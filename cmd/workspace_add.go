package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/resource"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var (
	addAs  string
	addPin string
)

var workspaceAddCmd = &cobra.Command{
	Use:   "add <url|path>",
	Short: "add a resource (git url, local path, or org/name tracker)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		name := addAs
		if name == "" {
			name = inferResourceName(args[0])
		}
		return runWorkspaceAdd(ws, args[0], name, addPin, "")
	},
}

func init() {
	workspaceAddCmd.Flags().StringVar(&addAs, "as", "", "resource name (defaults to inferred)")
	workspaceAddCmd.Flags().StringVar(&addPin, "pin", "", "git ref to pin (git resources only)")
	workspaceCmd.AddCommand(workspaceAddCmd)
}

func runWorkspaceAdd(ws, input, name, pin, typeOverride string) error {
	kind := typeOverride
	if kind == "" {
		kind = inferKind(input)
	}
	r := resource.Resource{Name: name, Kind: resource.Kind(kind)}
	switch kind {
	case "git":
		r.URL = input
		if pin != "" {
			r.Ref = pin
		} else {
			r.Ref = "main"
		}
	case "local":
		r.Path = input
	case "tracker":
		r.Repo = input
	default:
		return fmt.Errorf("could not infer resource kind from %q", input)
	}
	res, err := resource.Sync(ws, r)
	if err != nil {
		return err
	}

	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	for _, existing := range w.Resources {
		if existing.Name == name {
			return fmt.Errorf("resource %q already exists", name)
		}
	}
	w.Resources = append(w.Resources, workspace.Resource{
		Name: name, Type: kind,
		URL: r.URL, Ref: r.Ref, Pin: res.Pin,
		Path: r.Path, Repo: r.Repo,
	})
	if err := os.WriteFile(wsFile, workspace.Serialize(w), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmdStderr(), "added resource %q (%s)\n", name, kind)
	return nil
}

func inferKind(input string) string {
	switch {
	case strings.HasPrefix(input, "http://"),
		strings.HasPrefix(input, "https://"),
		strings.HasPrefix(input, "git@"),
		strings.HasSuffix(input, ".git"):
		return "git"
	case strings.HasPrefix(input, "~"),
		strings.HasPrefix(input, "/"),
		strings.HasPrefix(input, "."):
		return "local"
	case strings.Contains(input, "/") && !strings.Contains(input, ":"):
		return "tracker"
	}
	return ""
}

func inferResourceName(input string) string {
	if u, err := url.Parse(input); err == nil && u.Host != "" {
		name := path.Base(u.Path)
		return strings.TrimSuffix(name, ".git")
	}
	if strings.HasPrefix(input, "~") || strings.HasPrefix(input, "/") || strings.HasPrefix(input, ".") {
		return filepath.Base(input)
	}
	if i := strings.Index(input, "/"); i >= 0 {
		return input[i+1:]
	}
	return input
}
