//go:build windows

package cron

import "os"

// Windows does not support flock; these are no-ops.
// The cron scheduler is not supported on Windows but the package must compile.
func lockFile(_ *os.File) error  { return nil }
func unlockFile(_ *os.File) error { return nil }
