package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
)

// schema.go defines validation rules for manifests and CLI configuration. Keep
// this aligned with PRD expectations for watch directories and logging.

// Validation errors returned by manifest helpers.
var (
	ErrNoDirectories = errors.New("config: manifest must specify at least one directory")
)

// normalizeDirectories ensures every watch directory is absolute, deduplicated,
// and sorted for deterministic persistence.
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

// normalizeLogPath cleans and absolutizes the log path when supplied.
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
