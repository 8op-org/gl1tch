package resource

import (
	"fmt"
	"time"
)

// SyncOpts controls sync behaviour.
type SyncOpts struct {
	Force bool // re-clone git resources from scratch
}

// Sync materializes (or refreshes) a resource into ws/resources/<name>
// and returns a Result with pin + timestamp populated.
func Sync(ws string, r Resource, opts ...SyncOpts) (Result, error) {
	if err := ValidateName(r.Name); err != nil {
		return Result{}, err
	}
	var opt SyncOpts
	if len(opts) > 0 {
		opt = opts[0]
	}
	res := Result{Name: r.Name, Kind: r.Kind, FetchedAt: time.Now().Unix()}
	switch r.Kind {
	case KindGit:
		sha, path, err := materializeGit(ws, r, opt.Force)
		if err != nil {
			return Result{}, err
		}
		res.Pin = sha
		res.Path = path
	case KindLocal:
		path, err := materializeLocal(ws, r)
		if err != nil {
			return Result{}, err
		}
		res.Path = path
	case KindTracker:
		res.Repo = materializeTracker(r)
	default:
		return Result{}, fmt.Errorf("unknown resource kind %q", r.Kind)
	}
	return res, nil
}
