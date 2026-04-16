package resource

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateName rejects names that would escape the workspace resources/ directory
// or collide with other special files. Returns an error on empty, containing "/"
// or the OS path separator, "..", or starting with ".".
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name required")
	}
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, filepath.Separator) {
		return fmt.Errorf("resource name %q must not contain path separators", name)
	}
	if name == "." || name == ".." || strings.HasPrefix(name, ".") {
		return fmt.Errorf("resource name %q must not start with '.'", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("resource name %q must not contain '..'", name)
	}
	return nil
}

// Kind is the resource type enum.
type Kind string

const (
	KindGit     Kind = "git"
	KindLocal   Kind = "local"
	KindTracker Kind = "tracker"
)

// Resource is the input to Sync.
type Resource struct {
	Name string
	Kind Kind
	URL  string // git
	Ref  string // git
	Pin  string // git; ignored on input — written back on output
	Path string // local
	Repo string // tracker
}

// Result is Sync's output. Mirrors Resource with pin + timestamp filled.
type Result struct {
	Name      string
	Kind      Kind
	Pin       string
	Path      string // materialized path (clone root / symlink)
	Repo      string
	FetchedAt int64 // unix seconds
}
