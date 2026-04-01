package cron

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// cronFilePath returns the path to ~/.config/glitch/cron.yaml.
func cronFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "glitch", "cron.yaml"), nil
}

// WriteEntry atomically adds or replaces a named entry in cron.yaml.
// If an entry with the same Name already exists it is replaced in-place.
// The file is created if it does not exist.
func WriteEntry(entry Entry) error {
	path, err := cronFilePath()
	if err != nil {
		return err
	}
	return writeEntry(path, entry)
}

// writeEntry is the testable core of WriteEntry that accepts an explicit path.
func writeEntry(path string, entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Acquire exclusive lock.
	if err := lockFile(f); err != nil {
		return err
	}
	defer unlockFile(f) //nolint:errcheck

	var cfg cronConfig
	dec := yaml.NewDecoder(f)
	_ = dec.Decode(&cfg) // ignore EOF / empty file

	// Upsert: replace existing entry with same name, else append.
	found := false
	for i, e := range cfg.Entries {
		if e.Name == entry.Name {
			cfg.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		cfg.Entries = append(cfg.Entries, entry)
	}

	return atomicWrite(path, cfg)
}

// RemoveEntry removes the named entry from cron.yaml. It is a no-op if the
// name is not found.
func RemoveEntry(name string) error {
	path, err := cronFilePath()
	if err != nil {
		return err
	}
	return removeEntry(path, name)
}

// removeEntry is the testable core of RemoveEntry.
func removeEntry(path string, name string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	if err := lockFile(f); err != nil {
		return err
	}
	defer unlockFile(f) //nolint:errcheck

	var cfg cronConfig
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil // empty or invalid file → nothing to remove
	}

	filtered := cfg.Entries[:0]
	for _, e := range cfg.Entries {
		if e.Name != name {
			filtered = append(filtered, e)
		}
	}
	cfg.Entries = filtered

	return atomicWrite(path, cfg)
}

// atomicWrite serialises cfg to a temp file next to path and renames into
// place. The file descriptor passed to the lock must stay open until after the
// rename — callers are responsible for that.
func atomicWrite(path string, cfg cronConfig) error {
	tmp := path + ".tmp"
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
