// Package config provides utilities for loading and validating lowkey daemon
// configurations. It defines the structure of manifest files, handles `.lowkey`
// ignore patterns, and includes helpers for parsing configurations from both
// disk and command-line arguments.
//
// This package ensures that all configuration data, such as directory paths,
// is normalized into a consistent, absolute format for reliable use by other
// parts of the application, such as the watcher and daemon services.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest represents the persisted daemon configuration. It specifies which
// directories to watch, where to write logs, and which ignore file to use.
// The fields are used by the daemon to configure its file system monitoring
// and logging behavior.
type Manifest struct {
	Directories []string `json:"directories"`
	LogPath     string   `json:"log_path,omitempty"`
	IgnoreFile  string   `json:"ignore_file,omitempty"`
}

// LoadManifest parses a manifest file from disk. It performs validation and
// normalization, ensuring that all paths are absolute and ready for use.
// This function is the primary entry point for loading a daemon's
// configuration from a file.
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
// semantics understood by the watcher layer. This allows for flexible and
// powerful ignore patterns.
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
// basePath parameter is typically the current working directory, used to resolve
// relative directory paths into absolute ones.
func BuildManifestFromArgs(basePath string, dirs []string) (*Manifest, error) {
	normalized, err := normalizeDirectories(basePath, dirs)
	if err != nil {
		return nil, err
	}
	return &Manifest{Directories: normalized}, nil
}
