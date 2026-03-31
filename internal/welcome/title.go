package welcome

import (
	"os/exec"
	"strings"

	"github.com/adam-stokes/orcai/internal/tdf"
)

// fallbackTitle is a pre-rendered block-art "ORCAI" approximating the amnesiax figlet font style.
var fallbackTitle = []string{
	` ▄██████▄  ██████████   ████████  ██████████  ████`,
	`███    ███ ░░███░░░░███ ███░░░░███░░███░░░░░█ ░░███`,
	`███    ███  ░███   ░░███░███        ░███  █ ░   ░███`,
	`███    ███  ░███    ░███░░█████████ ░██████     ░███`,
	`███    ███  ░███    ░███ ░░░░░░░░███░███░░█     ░███`,
	`░███  ████  ░███    ███  ███    ░███░███ ░   █  ░███      █`,
	` ░░████████ ██████████  ░░████████ ██████████ ███████████`,
	`  ░░░░░░░░ ░░░░░░░░░░    ░░░░░░░░ ░░░░░░░░░░░░░░░░░░░░░`,
}

// RenderTitle tries to render "ORCAI" using the native Go TDF renderer first,
// then tdfiglet binary, then plain figlet, and finally falls back to the embedded block art.
func RenderTitle() []string {
	// 1. Try native Go TDF renderer (no external binary needed)
	if f, err := tdf.LoadEmbedded("amnesiax"); err == nil {
		lines := tdf.RenderString("ORCAI", f)
		if len(lines) > 0 {
			return lines
		}
	}
	// 2. tdfiglet binary
	if lines := runFiglet("tdfiglet", "-f", "amnesiax", "ORCAI"); lines != nil {
		return lines
	}
	// 3. figlet binary
	if lines := runFiglet("figlet", "-f", "standard", "ORCAI"); lines != nil {
		return lines
	}
	// 4. Embedded block art
	return fallbackTitle
}

// runFiglet executes the given command and returns its output split into lines,
// or nil if the command fails or produces empty output.
func runFiglet(name string, args ...string) []string {
	out, err := exec.Command(name, args...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 0 {
		return nil
	}
	return lines
}
