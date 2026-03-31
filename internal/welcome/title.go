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

// RenderTitle attempts to render "ORCAI" using tdfiglet with the amnesiax font.
// Falls back to the built-in block art if tdfiglet is unavailable or errors.
func RenderTitle() []string {
	out, err := exec.Command("tdfiglet", "-f", "amnesiax", "ORCAI").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return fallbackTitle
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 0 {
		return fallbackTitle
	}
	return lines
}
