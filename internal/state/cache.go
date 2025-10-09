// Package state provides data structures and utilities for managing persistent
// and in-memory state for the lowkey daemon. This includes caching file
// signatures for incremental scanning and persisting daemon manifests.
//
// The components in this package are designed to be thread-safe and provide
// atomic operations for file-based persistence, ensuring data consistency even
// in the case of unexpected termination.
package state

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const smallFileThreshold = 4096 // 4KB threshold for hashing small files

// FileSignature captures the metadata of a file at a specific point in time.
// It is used to detect changes to files without having to re-hash their
// contents on every scan.
type FileSignature struct {
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Hash    string    `json:"hash,omitempty"`
}

// Equal reports whether two file signatures are identical. This is the core
// logic for determining if a file has been modified.
func (s FileSignature) Equal(other FileSignature) bool {
	return s.Size == other.Size && s.ModTime.Equal(other.ModTime) && s.Hash == other.Hash
}

// Cache stores file signatures in memory, keyed by their absolute paths. It
// provides thread-safe access to the signatures and is used by the watcher to
// maintain a consistent view of the file system state.
type Cache struct {
	mu    sync.RWMutex
	files map[string]FileSignature
}

// NewCache constructs an empty, ready-to-use Cache.
func NewCache() *Cache {
	return &Cache{files: make(map[string]FileSignature)}
}

// NewCacheFromSnapshot creates a new cache pre-populated with a given set of
// file signatures. The provided map is copied to prevent shared ownership.
func NewCacheFromSnapshot(entries map[string]FileSignature) *Cache {
	cache := NewCache()
	cache.ReplaceAll(entries)
	return cache
}

// Get retrieves the signature for a given path from the cache. It returns the
// signature and a boolean indicating whether the path was found.
func (c *Cache) Get(path string) (FileSignature, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sig, ok := c.files[path]
	return sig, ok
}

// Set adds or updates a file signature in the cache.
func (c *Cache) Set(path string, sig FileSignature) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[path] = sig
}

// Delete removes a file signature from the cache.
func (c *Cache) Delete(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.files, path)
}

// Snapshot returns a deep copy of all file signatures currently in the cache.
// This provides a thread-safe way to access a consistent view of the cache's
// contents.
func (c *Cache) Snapshot() map[string]FileSignature {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot := make(map[string]FileSignature, len(c.files))
	for path, sig := range c.files {
		snapshot[path] = sig
	}
	return snapshot
}

// ReplaceAll atomically replaces the entire contents of the cache with a new
// set of file signatures.
func (c *Cache) ReplaceAll(entries map[string]FileSignature) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.files = make(map[string]FileSignature, len(entries))
	for path, sig := range entries {
		c.files[path] = sig
	}
}

// Len returns the number of file signatures currently stored in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.files)
}

// FilesUnder returns a copy of all cache entries whose paths are within the
// given directory.
func (c *Cache) FilesUnder(dir string) map[string]FileSignature {
	cleanDir := filepath.Clean(dir)
	prefix := cleanDir
	if prefix != string(os.PathSeparator) {
		prefix += string(os.PathSeparator)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]FileSignature)
	for path, sig := range c.files {
		if path == cleanDir || strings.HasPrefix(path, prefix) {
			result[path] = sig
		}
	}
	return result
}

// ComputeSignature calculates the signature for a file based on its size,
// modification time, and, for small files, its content hash. It returns an
// error if the path is a directory.
func ComputeSignature(path string, info fs.FileInfo) (FileSignature, error) {
	if info.IsDir() {
		return FileSignature{}, errors.New("state: compute signature called for directory")
	}

	sig := FileSignature{Size: info.Size(), ModTime: info.ModTime().UTC()}
	if info.Size() > 0 && info.Size() <= smallFileThreshold {
		file, err := os.Open(path)
		if err != nil {
			return FileSignature{}, err
		}
		defer file.Close()

		digest := sha256.New()
		if _, err := io.Copy(digest, io.LimitReader(file, smallFileThreshold)); err != nil {
			return FileSignature{}, err
		}
		sig.Hash = hex.EncodeToString(digest.Sum(nil))
	}

	return sig, nil
}

// DetectChange compares a cached file signature with the current state of the
// file on disk. It returns the new signature and a boolean indicating whether a
// change was detected.
func DetectChange(cached FileSignature, hasCached bool, info fs.FileInfo, path string) (FileSignature, bool, error) {
	sig, err := ComputeSignature(path, info)
	if err != nil {
		return FileSignature{}, false, err
	}
	if !hasCached || !cached.Equal(sig) {
		return sig, true, nil
	}
	return sig, false, nil
}

// NormalizePath cleans and absolutizes a file path. If the path is relative,
// it is resolved against the current working directory.
func NormalizePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("state: empty path")
	}
	abs := path
	if !filepath.IsAbs(abs) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		abs = filepath.Join(cwd, abs)
	}
	return filepath.Clean(abs), nil
}
