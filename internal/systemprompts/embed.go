package systemprompts

import "embed"

//go:embed defaults/*.md
var defaultFS embed.FS
