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

// RenderTitle tries to render "ORCAI" via the native Go TDF renderer, then
// tdfiglet binary, then plain figlet, and finally falls back to embedded block
// art. The bool return is true when the lines already contain ANSI color codes
// (TDF or tdfiglet output) and should not be re-wrapped by lipgloss.
func RenderTitle() ([]string, bool) {
	// 1. Native Go TDF renderer (embedded font, no external binary needed).
	if f, err := tdf.LoadEmbedded("amnesiax"); err == nil {
		lines := tdf.RenderString("ORCAI", f)
		if len(lines) > 0 {
			return lines, true
		}
	}
	// 2. tdfiglet binary.
	if lines := runFiglet("tdfiglet", "-f", "amnesiax", "ORCAI"); lines != nil {
		return lines, true
	}
	// 3. Plain figlet binary (no color).
	if lines := runFiglet("figlet", "-f", "standard", "ORCAI"); lines != nil {
		return lines, false
	}
	// 4. Embedded block art.
	return fallbackTitle, false
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
