package research

import "context"

type Researcher interface {
	Name() string
	Describe() string
	Gather(ctx context.Context, q ResearchQuery, prior EvidenceBundle) (Evidence, error)
}
