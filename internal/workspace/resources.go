package workspace

// Resource is a declared workspace resource (git repo, local folder, or tracker alias).
type Resource struct {
	Name string
	Type string // "git" | "local" | "tracker"
	URL  string // git only
	Ref  string // git only
	Pin  string // git only (tool-managed)
	Path string // local only
	Repo string // tracker only; "org/name"
}
