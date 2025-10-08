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
	"errors"
	"fmt"
	"path/filepath"
	"sort"
)

// ErrNoDirectories is returned when a manifest or configuration is invalid
// because it fails to specify any directories to watch.
var ErrNoDirectories = errors.New("config: manifest must specify at least one directory")

// normalizeDirectories ensures every watch directory is absolute, deduplicated,
// and sorted. This guarantees a deterministic and reliable list of directories
// for the file system watcher.
func normalizeDirectories(base string, dirs []string) ([]string, error) {
	if len(dirs) == 0 {
		return nil, ErrNoDirectories
	}

	seen := make(map[string]struct{}, len(dirs))
	result := make([]string, 0, len(dirs))

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		abs := dir
		if !filepath.IsAbs(abs) {
			if base == "" {
				return nil, fmt.Errorf("config: relative path %q requires a base directory", dir)
			}
			abs = filepath.Join(base, dir)
		}
		abs = filepath.Clean(abs)
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		result = append(result, abs)
	}

	if len(result) == 0 {
		return nil, ErrNoDirectories
	}

	sort.Strings(result)
	return result, nil
}

// normalizeLogPath cleans and absolutizes the log path when supplied. If the
// path is relative, it is resolved against the provided base directory.
func normalizeLogPath(base, logPath string) (string, error) {
	if logPath == "" {
		return "", nil
	}
	if !filepath.IsAbs(logPath) {
		if base == "" {
			return "", fmt.Errorf("config: relative log path %q requires a base directory", logPath)
		}
		logPath = filepath.Join(base, logPath)
	}
	return filepath.Clean(logPath), nil
}
