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

// cache.go tracks file signatures for incremental scanning. Implement thread-safe
// read/write helpers as described in algorithm_design.md.

const smallFileThreshold = 64 * 1024 // 64 KiB keeps hashing cheap for tiny files.

// FileSignature captures basic metadata for a file at a point in time.
type FileSignature struct {
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Hash    string    `json:"hash,omitempty"`
}

// Equal reports whether two signatures represent the same file contents.
func (s FileSignature) Equal(other FileSignature) bool {
	return s.Size == other.Size && s.ModTime.Equal(other.ModTime) && s.Hash == other.Hash
}

// Cache stores file signatures keyed by absolute path. Callers must treat the
// stored paths as immutable: the cache always returns copies to maintain privacy.
type Cache struct {
	mu    sync.RWMutex
	files map[string]FileSignature
}

// NewCache constructs an empty cache instance.
func NewCache() *Cache {
	return &Cache{files: make(map[string]FileSignature)}
}

// NewCacheFromSnapshot constructs a cache pre-populated with the supplied
// entries. The snapshot map is copied so the caller retains ownership.
func NewCacheFromSnapshot(entries map[string]FileSignature) *Cache {
	cache := NewCache()
	cache.ReplaceAll(entries)
	return cache
}

// Get returns the signature for path if present.
func (c *Cache) Get(path string) (FileSignature, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sig, ok := c.files[path]
	return sig, ok
}

// Set stores or updates the signature for the supplied path.
func (c *Cache) Set(path string, sig FileSignature) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[path] = sig
}

// Delete removes a path entry if it exists.
func (c *Cache) Delete(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.files, path)
}

// Snapshot returns a deep copy of the current cache entries.
func (c *Cache) Snapshot() map[string]FileSignature {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot := make(map[string]FileSignature, len(c.files))
	for path, sig := range c.files {
		snapshot[path] = sig
	}
	return snapshot
}

// ReplaceAll swaps the cache contents with the provided snapshot, copying the
// values to avoid aliasing.
func (c *Cache) ReplaceAll(entries map[string]FileSignature) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.files = make(map[string]FileSignature, len(entries))
	for path, sig := range entries {
		c.files[path] = sig
	}
}

// Len returns the current number of cached file entries.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.files)
}

// FilesUnder returns a copy of entries whose path is contained within the
// supplied directory.
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

// ComputeSignature builds a FileSignature for the provided file information.
// Directories are ignored and return an error.
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

// DetectChange compares the cached signature with the supplied file info and
// returns whether the file should be treated as modified. When the file does not
// exist in the cache it is reported as a change.
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

// NormalizePath cleans and absolutises the supplied path. Callers may pass
// relative paths; they are resolved against the current working directory.
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
