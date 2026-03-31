package welcome

import (
	"os/exec"
	"strings"
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

// RenderTitle tries to render "ORCAI" via tdfiglet (amnesiax font), then plain
// figlet, and finally falls back to the embedded block art.
func RenderTitle() []string {
	if lines := runFiglet("tdfiglet", "-f", "amnesiax", "ORCAI"); lines != nil {
		return lines
	}
	if lines := runFiglet("figlet", "-f", "standard", "ORCAI"); lines != nil {
		return lines
	}
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
