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

// manifest_store.go persists daemon manifests in `$XDG_STATE_HOME/lowkey`. It is
// invoked by start/stop commands and the daemon manager.

const manifestFilename = "daemon.json"

// ManifestStore wraps disk persistence for the daemon manifest.
type ManifestStore struct {
	dir  string
	path string
}

// NewManifestStore returns a store whose manifest file lives at `dir/daemon.json`.
func NewManifestStore(dir string) (*ManifestStore, error) {
	if dir == "" {
		return nil, errors.New("state: empty state directory")
	}
	cleanDir := filepath.Clean(dir)
	path := filepath.Join(cleanDir, manifestFilename)
	return &ManifestStore{dir: cleanDir, path: path}, nil
}

// DefaultStateDir resolves the platform-specific directory for daemon metadata.
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

// Path returns the manifest file location.
func (s *ManifestStore) Path() string {
	return s.path
}

// Save marshals the manifest to disk atomically, creating parent directories as needed.
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

// Load reads the persisted manifest if one exists.
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

// Clear removes the stored manifest.
func (s *ManifestStore) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("state: remove manifest %q: %w", s.path, err)
	}
	return nil
}
