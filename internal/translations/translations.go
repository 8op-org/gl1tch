// Package translations provides a UI string overrides layer for GL1TCH.
// Operators can override any labeled string in the UI via a YAML file at
// ~/.config/glitch/translations.yaml. Values may contain raw ANSI escape
// sequences or the shorthand notations \e[, \033[, and \x1b[.
package translations

// Provider is implemented by any translations source.
type Provider interface {
	// T returns the translation for key, or fallback if key is not configured.
	T(key, fallback string) string
}
