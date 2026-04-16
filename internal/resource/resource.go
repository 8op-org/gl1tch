package resource

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
