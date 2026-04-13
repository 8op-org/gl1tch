package research

// ResearchDocument is the normalized input to the research loop.
type ResearchDocument struct {
	Source    string            `json:"source"`
	SourceURL string           `json:"source_url"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Repo      string            `json:"repo"`
	RepoPath  string            `json:"repo_path"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Links     []Link            `json:"links,omitempty"`
}

// Link is a URL reference extracted from a research document.
type Link struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}
