package resource

// trackers have no filesystem side effects today; the repo alias is all that's
// recorded in workspace.glitch. Kept as its own file for future 404 probes.
func materializeTracker(r Resource) string {
	return r.Repo
}
