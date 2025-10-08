// Package state provides data structures and utilities for managing persistent
// and in-memory state for the lowkey daemon. This includes caching file
// signatures for incremental scanning and persisting daemon manifests.
//
// The components in this package are designed to be thread-safe and provide
// atomic operations for file-based persistence, ensuring data consistency even
// in the case of unexpected termination.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"lowkey/pkg/config"
)

// ManifestStore provides a thread-safe way to read and write the daemon's
// manifest file. It handles the persistence of the daemon's configuration,
// ensuring that it can be reliably loaded across restarts.
type ManifestStore struct {
	dir  string
	path string
}

// NewManifestStore creates a new ManifestStore for the given directory. The
// manifest file will be stored as `daemon.json` within this directory.
func NewManifestStore(dir string) (*ManifestStore, error) {
	if dir == "" {
		return nil, errors.New("state: empty state directory")
	}
	cleanDir := filepath.Clean(dir)
	path := filepath.Join(cleanDir, manifestFilename)
	return &ManifestStore{dir: cleanDir, path: path}, nil
}

// DefaultStateDir determines the appropriate platform-specific directory for
// storing the daemon's state, following the XDG Base Directory Specification.
func DefaultStateDir() (string, error) {
	if custom := os.Getenv("XDG_STATE_HOME"); custom != "" {
		return filepath.Join(custom, "lowkey"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("state: resolve home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "lowkey"), nil
	case "windows":
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "lowkey"), nil
		}
		return filepath.Join(home, "AppData", "Local", "lowkey"), nil
	default:
		return filepath.Join(home, ".local", "state", "lowkey"), nil
	}
}

// Path returns the full path to the manifest file.
func (s *ManifestStore) Path() string {
	return s.path
}

// Save atomically writes the given manifest to disk. It uses a temporary file
// and an atomic rename to prevent data corruption.
func (s *ManifestStore) Save(man *config.Manifest) error {
	if man == nil {
		return errors.New("state: manifest is nil")
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("state: create directory %q: %w", s.dir, err)
	}

	file, err := os.CreateTemp(s.dir, "manifest-*.json")
	if err != nil {
		return fmt.Errorf("state: create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(man); err != nil {
		file.Close()
		return fmt.Errorf("state: encode manifest: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("state: close temp manifest: %w", err)
	}

	if err := os.Rename(file.Name(), s.path); err != nil {
		return fmt.Errorf("state: atomically replace %q: %w", s.path, err)
	}
	return nil
}

// Load reads and decodes the manifest from disk. If the file does not exist,
// it returns a nil manifest without an error, indicating that no manifest has
// been saved yet.
func (s *ManifestStore) Load() (*config.Manifest, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("state: read manifest %q: %w", s.path, err)
	}

	var manifest config.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("state: decode manifest %q: %w", s.path, err)
	}
	return &manifest, nil
}

// Clear removes the manifest file from disk.
func (s *ManifestStore) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("state: remove manifest %q: %w", s.path, err)
	}
	return nil
}
