package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// persistence.go handles durable storage for the cache (e.g., boltDB or JSON).
// Ensure writes are atomic so crash recovery honors the PRD.

type persistedCache struct {
	Version int                      `json:"version"`
	Files   map[string]FileSignature `json:"files"`
}

// Save writes the cache contents to path atomically.
func Save(cache *Cache, path string) error {
	if cache == nil {
		return errors.New("state: cache is nil")
	}
	if path == "" {
		return errors.New("state: save path is empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("state: create cache directory %q: %w", dir, err)
	}

	snapshot := cache.Snapshot()
	payload := persistedCache{Version: 1, Files: snapshot}

	tempFile, err := os.CreateTemp(dir, "cache-*.json")
	if err != nil {
		return fmt.Errorf("state: create temp cache file: %w", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		tempFile.Close()
		return fmt.Errorf("state: encode cache: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("state: close temp cache: %w", err)
	}

	if err := os.Rename(tempFile.Name(), path); err != nil {
		return fmt.Errorf("state: atomically replace cache %q: %w", path, err)
	}

	return nil
}

// Load reads a cache from path, returning an empty cache when the file is absent.
func Load(path string) (*Cache, error) {
	if path == "" {
		return nil, errors.New("state: load path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewCache(), nil
		}
		return nil, fmt.Errorf("state: read cache %q: %w", path, err)
	}

	var payload persistedCache
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("state: decode cache %q: %w", path, err)
	}

	if payload.Files == nil {
		payload.Files = make(map[string]FileSignature)
	}

	return NewCacheFromSnapshot(payload.Files), nil
}
