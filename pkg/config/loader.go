package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loader.go initializes Viper, loads `.lowkey` files, and reads daemon manifests.
// Export helpers consumed by cmd/lowkey and internal/daemon.

// Manifest represents the persisted daemon configuration.
type Manifest struct {
	Directories []string `json:"directories"`
	LogPath     string   `json:"log_path,omitempty"`
	IgnoreFile  string   `json:"ignore_file,omitempty"`
}

// LoadManifest parses a manifest file from disk, applying validation and
// normalisation rules so downstream callers can rely on absolute paths.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read manifest %q: %w", path, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("config: decode manifest %q: %w", path, err)
	}

	dir := filepath.Dir(path)
	manifest.Directories, err = normalizeDirectories(dir, manifest.Directories)
	if err != nil {
		return nil, err
	}
	manifest.LogPath, err = normalizeLogPath(dir, manifest.LogPath)
	if err != nil {
		return nil, err
	}

	if manifest.IgnoreFile != "" && !filepath.IsAbs(manifest.IgnoreFile) {
		manifest.IgnoreFile = filepath.Clean(filepath.Join(dir, manifest.IgnoreFile))
	}

	return &manifest, nil
}

// LoadIgnorePatterns reads a `.lowkey` ignore file. Lines beginning with `#`
// or blank lines are ignored. Paths are returned as provided to match glob
// semantics understood by the watcher layer.
func LoadIgnorePatterns(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read ignore file %q: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	patterns := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		patterns = append(patterns, trimmed)
	}
	return patterns, nil
}

// BuildManifestFromArgs creates a manifest from CLI-supplied directories. The
// basePath parameter is typically the current working directory.
func BuildManifestFromArgs(basePath string, dirs []string) (*Manifest, error) {
	normalized, err := normalizeDirectories(basePath, dirs)
	if err != nil {
		return nil, err
	}
	return &Manifest{Directories: normalized}, nil
}
