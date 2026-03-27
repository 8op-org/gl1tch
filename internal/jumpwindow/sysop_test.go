package jumpwindow

import "testing"

func TestListSysopWindows_NoSession(t *testing.T) {
	// listSysopWindows queries "orcai-cron" session which won't exist in test env.
	// Expect nil result without panicking.
	wins := listSysopWindows()
	// In a test environment without tmux / without orcai-cron session, this should be nil.
	// We can't assert nil definitively (CI might have tmux), just ensure no panic.
	_ = wins
}
