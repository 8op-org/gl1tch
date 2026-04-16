package resource

import (
	"os"
	"path/filepath"
)

// materializeLocal creates or verifies a symlink at ws/resources/<name>
// pointing to the expanded path.
func materializeLocal(ws string, r Resource) (string, error) {
	link := filepath.Join(ws, "resources", r.Name)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return "", err
	}
	target := expandHome(r.Path)
	if _, err := os.Lstat(link); err == nil {
		if cur, _ := os.Readlink(link); cur == target {
			return link, nil
		}
		_ = os.Remove(link)
	}
	if err := os.Symlink(target, link); err != nil {
		return "", err
	}
	return link, nil
}

func expandHome(p string) string {
	if len(p) > 1 && p[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
