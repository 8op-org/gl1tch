package cron

import (
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

const (
	// cronLogFile is the path relative to the data directory.
	cronLogFile = "cron.log"
	// maxLogSize is 10 MB; the file is truncated if it exceeds this.
	maxLogSize = 10 * 1024 * 1024
)

// NewLogger creates a charmbracelet/log Logger that writes to both stderr and
// the cron log file at ~/.local/share/glitch/cron.log. If the log file exceeds
// 10 MB it is truncated before opening. Returns the logger and any error
// encountered while setting up the log file (stderr-only logging is still
// returned on file errors).
func NewLogger() (*log.Logger, error) {
	logPath, err := cronLogPath()
	if err != nil {
		// Fall back to stderr-only.
		return log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: true}), err
	}

	// Truncate oversized log file.
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > maxLogSize {
		if terr := os.Truncate(logPath, 0); terr != nil {
			// Non-fatal: log to stderr about it but continue.
			_ = terr
		}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: true}), err
	}

	w := io.MultiWriter(os.Stderr, f)
	logger := log.NewWithOptions(w, log.Options{
		ReportTimestamp: true,
		Prefix:          "glitch-cron",
	})
	return logger, nil
}

// cronLogPath resolves the path to the cron log file, creating the data
// directory if necessary.
func cronLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(home, ".local", "share", "glitch")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dataDir, cronLogFile), nil
}
